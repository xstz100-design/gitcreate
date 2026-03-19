"""图片上传接口 - 流式写入 + 安全防御"""
import uuid
import shutil
from pathlib import Path
from fastapi import APIRouter, Depends, HTTPException, UploadFile, File, Request
from app.core.dependencies import get_current_user
from app.models import User
from app.services.image import image_service
from app.core.rate_limit import upload_rate_limiter

router = APIRouter()

ALLOWED_EXTENSIONS = {".jpg", ".jpeg", ".png", ".webp", ".gif"}
MAX_FILE_SIZE = 5 * 1024 * 1024       # 单文件上限 5MB
CHUNK_SIZE = 64 * 1024                 # 每次读取 64KB
UPLOAD_DIR = Path("uploads")
MIN_DISK_FREE_MB = 500                 # 磁盘至少保留 500MB 可用空间


def _check_disk_space() -> bool:
    """检查磁盘剩余空间是否充足"""
    try:
        usage = shutil.disk_usage(UPLOAD_DIR)
        free_mb = usage.free / (1024 * 1024)
        return free_mb >= MIN_DISK_FREE_MB
    except Exception:
        return True


@router.post("")
async def upload_image(
    request: Request,
    file: UploadFile = File(...),
    current_user: User = Depends(get_current_user),
):
    """上传图片 - 流式写入，自动压缩

    安全措施:
    - 流式分块读取，绝不将整个文件载入内存
    - 文件大小实时校验，超限立即中止并删除残留
    - 磁盘空间预检，防止写满硬盘
    - IP+用户双维度速率限制
    - 文件扩展名白名单
    """
    # --- 速率限制 ---
    client_ip = request.client.host if request.client else "unknown"
    upload_rate_limiter.check_upload(client_ip, current_user.id)

    # --- 磁盘空间预检 ---
    if not _check_disk_space():
        raise HTTPException(status_code=507, detail="服务器磁盘空间不足，请联系管理员")

    # --- 扩展名校验 ---
    ext = Path(file.filename or "upload.jpg").suffix.lower()
    if ext not in ALLOWED_EXTENSIONS:
        raise HTTPException(
            status_code=400,
            detail=f"不支持的文件格式，仅支持: {', '.join(ALLOWED_EXTENSIONS)}",
        )

    # --- 流式写入 ---
    filename = f"{uuid.uuid4().hex}{ext}"
    UPLOAD_DIR.mkdir(exist_ok=True)
    file_path = UPLOAD_DIR / filename
    total_size = 0

    try:
        with open(file_path, "wb") as f:
            while True:
                chunk = await file.read(CHUNK_SIZE)
                if not chunk:
                    break
                total_size += len(chunk)
                if total_size > MAX_FILE_SIZE:
                    f.close()
                    file_path.unlink(missing_ok=True)
                    raise HTTPException(
                        status_code=413,
                        detail=f"文件大小超过限制 ({MAX_FILE_SIZE // 1024 // 1024}MB)",
                    )
                f.write(chunk)
    except HTTPException:
        raise
    except Exception as e:
        file_path.unlink(missing_ok=True)
        raise HTTPException(status_code=500, detail=f"文件写入失败: {str(e)}")
    finally:
        await file.close()

    # --- 图片压缩优化 ---
    try:
        image_service.optimize_image(str(file_path), max_size=(1200, 1200), quality=85)
    except Exception:
        pass

    return {
        "url": f"/uploads/{filename}",
        "filename": filename,
    }
