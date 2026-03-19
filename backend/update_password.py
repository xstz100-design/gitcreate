"""
更新admin用户密码为123456
"""
from sqlmodel import Session, select, create_engine
from app.models.base import User
from app.core.config import settings
from app.core.security import get_password_hash

def update_admin_password():
    """更新管理员密码为123456"""
    engine = create_engine(settings.DATABASE_URL)
    
    with Session(engine) as session:
        statement = select(User).where(User.username == "admin")
        admin = session.exec(statement).first()
        
        if admin:
            admin.hashed_password = get_password_hash("123456")
            session.add(admin)
            session.commit()
            print("✅ 管理员密码已更新为: 123456")
        else:
            print("❌ 未找到admin用户")

if __name__ == "__main__":
    update_admin_password()
