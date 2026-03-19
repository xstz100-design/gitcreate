"""
初始化数据库并创建管理员账号
"""
from app.models import User, Product, Category
from app.core.database import engine, Session
from app.core.security import get_password_hash
from sqlmodel import SQLModel


def init_db():
    """初始化数据库表"""
    print("🔨 创建数据库表...")
    SQLModel.metadata.create_all(engine)
    print("✅ 数据库表创建完成")


def create_admin():
    """创建默认管理员账号"""
    with Session(engine) as session:
        # 检查是否已存在管理员
        existing_admin = session.query(User).filter(User.username == "admin").first()
        if existing_admin:
            print("⚠️ 管理员账号已存在")
            return
        
        print("👤 创建管理员账号...")
        admin = User(
            username="admin",
            hashed_password=get_password_hash("123456"),
            full_name="系统管理员",
            role="admin",
        )
        session.add(admin)
        session.commit()
        print("✅ 管理员账号创建成功")
        print("   用户名: admin")
        print("   密码: 123456")
        print("   ⚠️ 请及时修改默认密码!")


def create_demo_merchant():
    """创建演示商户账号"""
    with Session(engine) as session:
        existing_merchant = session.query(User).filter(User.username == "merchant1").first()
        if existing_merchant:
            print("⚠️ 演示商户已存在")
            return
        
        print("👤 创建演示商户...")
        merchant = User(
            username="merchant1",
            hashed_password=get_password_hash("123456"),
            full_name="张三批发店",
            role="merchant",
            phone="+855 12 345 678",
            address="金边市中心",
            credit_limit=1000.0,
        )
        session.add(merchant)
        session.commit()
        print("✅ 演示商户创建成功")
        print("   用户名: merchant1")
        print("   密码: 123456")


def create_demo_categories():
    """创建演示分类"""
    with Session(engine) as session:
        existing = session.query(Category).first()
        if existing:
            print("⚠️ 演示分类已存在")
            return
        
        print("📂 创建演示分类...")
        cats = [
            Category(name="饮料", sort_order=1),
            Category(name="食品", sort_order=2),
            Category(name="调料", sort_order=3),
            Category(name="日用品", sort_order=4),
            Category(name="零食", sort_order=5),
        ]
        for c in cats:
            session.add(c)
        session.commit()
        print(f"✅ 已创建 {len(cats)} 个演示分类")


def create_demo_products():
    """创建演示商品"""
    with Session(engine) as session:
        # 检查是否已有商品
        existing_product = session.query(Product).first()
        if existing_product:
            print("⚠️ 演示商品已存在")
            return
        
        print("📦 创建演示商品...")
        demo_products = [
            Product(
                name="可口可乐 (330ml)",
                name_kh="កូកាកូឡា",
                unit="箱",
                price_usd=12.0,
                stock=100,
                stock_warning=20,
                category="饮料",
                description="经典可口可乐,24瓶/箱",
                sort_order=1,
            ),
            Product(
                name="康师傅方便面",
                name_kh="មី​ញាំ​ភ្លាម",
                unit="箱",
                price_usd=8.5,
                stock=50,
                stock_warning=15,
                category="食品",
                description="多种口味,30包/箱",
                sort_order=2,
            ),
            Product(
                name="白糖",
                name_kh="ស្ករ​ស",
                unit="袋",
                price_usd=25.0,
                stock=30,
                stock_warning=10,
                category="调料",
                description="优质白砂糖,50kg/袋",
                sort_order=3,
            ),
            Product(
                name="食用油",
                name_kh="ប្រេង​ធម្មតា",
                unit="桶",
                price_usd=18.0,
                stock=5,
                stock_warning=10,
                category="调料",
                description="大豆油,5L/桶",
                sort_order=4,
            ),
        ]
        
        for product in demo_products:
            session.add(product)
        
        session.commit()
        print(f"✅ 已创建 {len(demo_products)} 个演示商品")


if __name__ == "__main__":
    print("=" * 50)
    print("🚀 东方优选 TONGFANG YOUXUAN - 数据库初始化")
    print("=" * 50)
    
    init_db()
    create_admin()
    create_demo_merchant()
    create_demo_categories()
    create_demo_products()
    
    print("\n" + "=" * 50)
    print("✨ 初始化完成! 可以启动应用了:")
    print("   python main.py")
    print("=" * 50)
