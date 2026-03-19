from datetime import datetime, timezone, timedelta
from typing import Optional
from enum import Enum
from sqlmodel import SQLModel, Field, Relationship

# 柬埔寨时区 UTC+7
CAMBODIA_TZ = timezone(timedelta(hours=7))

def _now_cambodia():
    return datetime.now(CAMBODIA_TZ)


class UserRole(str, Enum):
    """用户角色"""
    ADMIN = "admin"          # 管理员

    MERCHANT = "merchant"    # 商户


class User(SQLModel, table=True):
    """用户表"""
    __tablename__ = "users"
    
    id: Optional[int] = Field(default=None, primary_key=True)
    username: str = Field(index=True, unique=True, max_length=50)
    hashed_password: str
    full_name: str = Field(max_length=100)
    role: UserRole = Field(default=UserRole.MERCHANT)
    
    # 商户专属字段
    phone: Optional[str] = Field(default=None, max_length=20)
    address: Optional[str] = Field(default=None, max_length=200)
    credit_limit: float = Field(default=0.0)  # 月结金额(累计)
    billing_day: Optional[int] = Field(default=None)  # 月结日(1-31)
    allow_monthly_billing: bool = Field(default=False)  # 是否允许月结
    location_url: Optional[str] = Field(default=None, max_length=500)  # 谷歌地图链接
    store_photo: Optional[str] = Field(default=None, max_length=500)  # 门面照片URL
    
    # Telegram 通知设置 (管理员)
    telegram_bot_token: Optional[str] = Field(default=None, max_length=200)
    telegram_chat_id: Optional[str] = Field(default=None, max_length=100)
    
    must_change_password: bool = Field(default=True)  # 首次登录需修改密码
    is_active: bool = Field(default=True)
    created_at: datetime = Field(default_factory=_now_cambodia)
    
    # 关系 - 明确指定foreign_keys避免歧义
    orders: list["Order"] = Relationship(
        back_populates="merchant",
        sa_relationship_kwargs={"foreign_keys": "[Order.merchant_id]"}
    )
    transactions: list["Transaction"] = Relationship(
        back_populates="user",
        sa_relationship_kwargs={"foreign_keys": "[Transaction.user_id]"}
    )


class Product(SQLModel, table=True):
    """商品表"""
    __tablename__ = "products"
    
    id: Optional[int] = Field(default=None, primary_key=True)
    name: str = Field(index=True, max_length=100)
    name_kh: Optional[str] = Field(default=None, max_length=100)  # 高棉语名称
    
    # 规格和价格
    unit: str = Field(default="件", max_length=20)  # 件/箱/包
    specs: Optional[str] = Field(default=None, max_length=100)  # 规格：每件多少包
    barcode: Optional[str] = Field(default=None, max_length=100)  # 商品条码
    price_usd: float = Field(gt=0)  # 批发价(美金)
    retail_price_usd: Optional[float] = Field(default=None)  # 建议零售价(美金)
    
    # 库存管理
    stock: int = Field(default=0, ge=0)  # 当前库存
    stock_warning: int = Field(default=10)  # 库存预警值
    
    # 商品信息
    description: Optional[str] = Field(default=None, max_length=500)
    image_url: Optional[str] = Field(default=None, max_length=500)
    img1: Optional[str] = Field(default=None, max_length=500)  # 图片1路径
    img2: Optional[str] = Field(default=None, max_length=500)  # 图片2路径
    img3: Optional[str] = Field(default=None, max_length=500)  # 图片3路径
    img4: Optional[str] = Field(default=None, max_length=500)  # 图片4路径
    img5: Optional[str] = Field(default=None, max_length=500)  # 图片5路径
    category: Optional[str] = Field(default=None, max_length=50)
    sort_order: int = Field(default=0)  # 排序序号
    is_featured: bool = Field(default=False)  # 推荐置顶
    
    is_active: bool = Field(default=True)
    created_at: datetime = Field(default_factory=_now_cambodia)
    updated_at: datetime = Field(default_factory=_now_cambodia)
    
    # 关系
    order_items: list["OrderItem"] = Relationship(back_populates="product")
    
    @property
    def is_low_stock(self) -> bool:
        """是否库存不足"""
        return self.stock <= self.stock_warning
