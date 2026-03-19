from pydantic_settings import BaseSettings
from typing import Optional
from pathlib import Path

# 获取backend目录的绝对路径
BACKEND_DIR = Path(__file__).resolve().parent.parent.parent
ENV_FILE = BACKEND_DIR / ".env"


class Settings(BaseSettings):
    """应用配置"""

    # 应用信息
    APP_NAME: str = "柬埔寨批发管理系统"
    APP_VERSION: str = "1.0.0"
    DEBUG: bool = False

    # 数据库
    DATABASE_URL: str = "sqlite:///./cambodia_wholesale.db"

    # JWT 认证
    SECRET_KEY: str = "your-secret-key-please-change-in-production-123456789"
    ALGORITHM: str = "HS256"
    ACCESS_TOKEN_EXPIRE_MINUTES: int = 10080  # 7天

    # Telegram (可选)
    TELEGRAM_BOT_TOKEN: Optional[str] = None
    TELEGRAM_CHAT_ID: Optional[str] = None

    # 货币汇率
    USD_TO_KHR_RATE: float = 4000.0

    class Config:
        env_file = str(ENV_FILE)
        case_sensitive = True
        extra = "ignore"


settings = Settings()
