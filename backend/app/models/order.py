from datetime import datetime, timezone, timedelta
from typing import Optional
from enum import Enum
from sqlmodel import SQLModel, Field, Relationship

# 柬埔寨时区 UTC+7
CAMBODIA_TZ = timezone(timedelta(hours=7))

def _now_cambodia():
    return datetime.now(CAMBODIA_TZ)


class PaymentStatus(str, Enum):
    """支付状态"""
    UNPAID = "unpaid"        # 未支付
    CASH = "cash"            # 现结
    MONTHLY = "monthly"      # 月结


class DeliveryStatus(str, Enum):
    """配送状态"""
    PENDING = "pending"      # 待派送
    DELIVERING = "delivering"  # 送货中
    DELIVERED = "delivered"    # 已签收
    CANCELLED = "cancelled"    # 已取消


class Order(SQLModel, table=True):
    """订单表"""
    __tablename__ = "orders"
    
    id: Optional[int] = Field(default=None, primary_key=True)
    order_no: str = Field(unique=True, index=True, max_length=50)  # 订单号
    
    # 关联商户
    merchant_id: int = Field(foreign_key="users.id")
    merchant: "User" = Relationship(
        back_populates="orders",
        sa_relationship_kwargs={"foreign_keys": "[Order.merchant_id]"}
    )
    
    # 金额信息
    total_usd: float = Field(default=0.0, ge=0)
    
    # 状态
    payment_status: PaymentStatus = Field(default=PaymentStatus.UNPAID)
    delivery_status: DeliveryStatus = Field(default=DeliveryStatus.PENDING)
    
    # 配送信息
    delivery_address: Optional[str] = Field(default=None, max_length=200)
    delivery_phone: Optional[str] = Field(default=None, max_length=20)
    delivery_person_id: Optional[int] = Field(default=None, foreign_key="users.id")
    
    # 备注
    note: Optional[str] = Field(default=None, max_length=500)
    
    # 时间戳
    created_at: datetime = Field(default_factory=_now_cambodia)
    updated_at: datetime = Field(default_factory=_now_cambodia)
    delivered_at: Optional[datetime] = None
    
    # 关系
    items: list["OrderItem"] = Relationship(back_populates="order")
    
    @property
    def total_khr(self) -> float:
        """总价(瑞尔) - 需要从配置获取汇率"""
        from app.core.config import settings
        return self.total_usd * settings.USD_TO_KHR_RATE


class OrderItem(SQLModel, table=True):
    """订单商品明细"""
    __tablename__ = "order_items"
    
    id: Optional[int] = Field(default=None, primary_key=True)
    
    # 关联订单和商品
    order_id: int = Field(foreign_key="orders.id")
    order: Order = Relationship(back_populates="items")
    
    product_id: int = Field(foreign_key="products.id")
    product: "Product" = Relationship(back_populates="order_items")
    
    # 数量和价格(下单时记录,避免商品改价影响历史订单)
    quantity: int = Field(gt=0)
    unit_price_usd: float = Field(gt=0)
    subtotal_usd: float = Field(ge=0)
    
    created_at: datetime = Field(default_factory=_now_cambodia)
