from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles
from pathlib import Path
from app.core.config import settings
from app.core.database import create_db_and_tables
from app.core.rate_limit import RateLimitMiddleware

# 导入路由
from app.api import auth, products, orders, upload, categories, announcements, billing


@asynccontextmanager
async def lifespan(app: FastAPI):
    """应用生命周期: 启动时建表"""
    create_db_and_tables()
    yield


app = FastAPI(
    title=settings.APP_NAME,
    version=settings.APP_VERSION,
    description="柬埔寨批发管理系统 - 支持库存管理、商户下单、配送调度",
    lifespan=lifespan,
)

# ---------- 中间件（注册顺序: 后注册先执行）----------

# CORS - 必须最外层
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 全局 IP 速率限制 — 同 IP 1秒超过 30 次请求 → 429
app.add_middleware(RateLimitMiddleware, requests_per_second=30)

# 挂载静态文件目录 - 图片上传
uploads_dir = Path("uploads")
uploads_dir.mkdir(exist_ok=True)
app.mount("/uploads", StaticFiles(directory="uploads"), name="uploads")


@app.get("/")
def root():
    """健康检查接口"""
    return {
        "app": settings.APP_NAME,
        "version": settings.APP_VERSION,
        "status": "running",
    }


# 注册路由
app.include_router(auth.router, prefix="/api/auth", tags=["认证"])
app.include_router(products.router, prefix="/api/products", tags=["商品"])
app.include_router(orders.router, prefix="/api/orders", tags=["订单"])
app.include_router(upload.router, prefix="/api/upload", tags=["上传"])
app.include_router(categories.router, prefix="/api/categories", tags=["分类"])
app.include_router(announcements.router, prefix="/api/announcements", tags=["公告"])
app.include_router(billing.router, prefix="/api/billing", tags=["月结"])


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=8000, reload=False)
