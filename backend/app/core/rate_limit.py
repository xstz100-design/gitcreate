"""速率限制 + 防暴力破解模块

设计原则:
- 纯内存实现，无外部依赖
- 自动清理过期记录，防止内存泄漏
- 双维度限制: IP 全局 + 接口专项
"""
import time
import threading
from collections import defaultdict
from fastapi import HTTPException, Request, Response
from starlette.middleware.base import BaseHTTPMiddleware


class _SlidingWindow:
    """滑动窗口计数器"""

    def __init__(self):
        self._data: dict[str, list[float]] = defaultdict(list)
        self._lock = threading.Lock()

    def hit(self, key: str, now: float | None = None) -> int:
        """记录一次请求，返回窗口内当前计数"""
        now = now or time.time()
        with self._lock:
            self._data[key].append(now)
            return len(self._data[key])

    def count(self, key: str, window: float, now: float | None = None) -> int:
        """获取窗口内的请求数"""
        now = now or time.time()
        cutoff = now - window
        with self._lock:
            lst = self._data[key]
            # 清理过期
            self._data[key] = [t for t in lst if t > cutoff]
            return len(self._data[key])

    def cleanup(self, window: float = 300):
        """清理所有超过 window 秒的记录"""
        cutoff = time.time() - window
        with self._lock:
            empty_keys = []
            for key, lst in self._data.items():
                self._data[key] = [t for t in lst if t > cutoff]
                if not self._data[key]:
                    empty_keys.append(key)
            for k in empty_keys:
                del self._data[k]


# ========== 全局IP速率限制中间件 ==========

class RateLimitMiddleware(BaseHTTPMiddleware):
    """
    全局 IP 速率限制

    默认: 同一 IP 1 秒内最多 30 次请求 → 超过直接 429
    商户正常下单 2-3 次/秒完全够用，30次/秒必定是脚本
    """

    def __init__(self, app, requests_per_second: int = 30):
        super().__init__(app)
        self.limit = requests_per_second
        self.window = _SlidingWindow()
        self._last_cleanup = time.time()

    async def dispatch(self, request: Request, call_next):
        client_ip = request.client.host if request.client else "unknown"
        now = time.time()

        # 每 60 秒全局清理一次
        if now - self._last_cleanup > 60:
            self.window.cleanup(window=10)
            self._last_cleanup = now

        count = self.window.count(client_ip, window=1.0, now=now)
        if count >= self.limit:
            return Response(
                content='{"detail":"请求过于频繁，请稍后再试"}',
                status_code=429,
                media_type="application/json",
            )

        self.window.hit(client_ip, now)
        response = await call_next(request)
        return response


# ========== 登录防暴力破解 ==========

class LoginProtector:
    """
    登录接口防暴力破解

    规则:
    - 同一 IP 5 分钟内最多 10 次登录尝试
    - 同一用户名 5 分钟内最多 5 次失败尝试
    - 超限后锁定 5 分钟
    """

    IP_MAX_ATTEMPTS = 10
    USER_MAX_ATTEMPTS = 5
    WINDOW = 300  # 5 分钟

    def __init__(self):
        self._ip_window = _SlidingWindow()
        self._user_window = _SlidingWindow()

    def check_login(self, ip: str, username: str):
        """登录前检查，超限抛 429"""
        now = time.time()

        ip_count = self._ip_window.count(f"login:{ip}", self.WINDOW, now)
        if ip_count >= self.IP_MAX_ATTEMPTS:
            raise HTTPException(
                status_code=429,
                detail="登录尝试过于频繁，请 5 分钟后再试",
            )

        user_count = self._user_window.count(f"user:{username}", self.WINDOW, now)
        if user_count >= self.USER_MAX_ATTEMPTS:
            raise HTTPException(
                status_code=429,
                detail="该账号登录尝试过于频繁，请 5 分钟后再试",
            )

    def record_attempt(self, ip: str, username: str):
        """记录一次登录尝试 (不论成功失败都记录 IP, 失败才记录 username)"""
        now = time.time()
        self._ip_window.hit(f"login:{ip}", now)

    def record_failure(self, ip: str, username: str):
        """记录一次登录失败"""
        now = time.time()
        self._ip_window.hit(f"login:{ip}", now)
        self._user_window.hit(f"user:{username}", now)


# ========== 上传速率限制 ==========

class UploadRateLimiter:
    """
    上传接口速率限制

    规则:
    - 同一 IP 1 分钟内最多 20 次上传
    - 同一用户 1 分钟内最多 15 次上传
    """

    IP_MAX = 20
    USER_MAX = 15
    WINDOW = 60  # 1 分钟

    def __init__(self):
        self._ip_window = _SlidingWindow()
        self._user_window = _SlidingWindow()

    def check_upload(self, ip: str, user_id: int):
        """上传前检查，超限抛 429"""
        now = time.time()

        ip_count = self._ip_window.count(f"upload:{ip}", self.WINDOW, now)
        if ip_count >= self.IP_MAX:
            raise HTTPException(
                status_code=429,
                detail="上传过于频繁，请稍后再试",
            )

        user_count = self._user_window.count(f"upload:u{user_id}", self.WINDOW, now)
        if user_count >= self.USER_MAX:
            raise HTTPException(
                status_code=429,
                detail="上传过于频繁，请稍后再试",
            )

        self._ip_window.hit(f"upload:{ip}", now)
        self._user_window.hit(f"upload:u{user_id}", now)


# ========== 全局单例 ==========
login_protector = LoginProtector()
upload_rate_limiter = UploadRateLimiter()
