from sqlmodel import SQLModel, create_engine, Session
from sqlalchemy import event
from .config import settings

# ---------- 引擎配置（适配 2G 内存 + SQLite）----------
#
# SQLite 注意事项:
# - check_same_thread=False: 允许多线程复用同一连接（FastAPI 多线程需要）
# - pool_size=5:   最多保持 5 个常驻连接，节省内存
# - max_overflow=0: 不允许额外溢出连接，防止 OOM
# - pool_recycle=3600: 1小时回收空闲连接
# - pool_pre_ping=True: 使用前 ping 一次，防止用到已断开的连接
#
_connect_args = {}
if settings.DATABASE_URL.startswith("sqlite"):
    _connect_args = {"check_same_thread": False}

engine = create_engine(
    settings.DATABASE_URL,
    echo=False,                    # 生产环境关闭 SQL 日志
    pool_pre_ping=True,
    pool_size=5,
    max_overflow=0,
    pool_recycle=3600,
    connect_args=_connect_args,
)

# SQLite 性能优化: 每个连接启用 WAL 模式 + busy_timeout
if settings.DATABASE_URL.startswith("sqlite"):
    @event.listens_for(engine, "connect")
    def _set_sqlite_pragma(dbapi_conn, connection_record):
        cursor = dbapi_conn.cursor()
        cursor.execute("PRAGMA journal_mode=WAL")
        cursor.execute("PRAGMA busy_timeout=5000")
        cursor.execute("PRAGMA synchronous=NORMAL")
        cursor.execute("PRAGMA cache_size=-8000")  # 8MB cache
        cursor.close()


def create_db_and_tables():
    """创建所有表"""
    SQLModel.metadata.create_all(engine)


def get_session():
    """获取数据库会话 - 用于依赖注入"""
    with Session(engine) as session:
        yield session
