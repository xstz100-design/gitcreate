from typing import Optional
from pydantic import BaseModel, Field
from app.models import UserRole, PaymentStatus, DeliveryStatus, AnnouncementType, BillStatus


# ============= 通用响应 =============
class MessageResponse(BaseModel):
    """通用消息响应"""
    message: str
    

class TokenResponse(BaseModel):
    """Token 响应"""
    access_token: str
    token_type: str = "bearer"
    user: "UserResponse"


# ============= 用户相关 =============
class UserResponse(BaseModel):
    """用户响应"""
    id: int
    username: str
    full_name: str
    role: UserRole
    phone: Optional[str] = None
    address: Optional[str] = None
    credit_limit: float = 0.0
    billing_day: Optional[int] = None
    allow_monthly_billing: bool = False
    location_url: Optional[str] = None
    store_photo: Optional[str] = None
    telegram_bot_token: Optional[str] = None
    telegram_chat_id: Optional[str] = None
    must_change_password: bool = False
    is_active: bool


class UserCreate(BaseModel):
    """创建用户 - 账号自动生成"""
    full_name: str
    role: UserRole = UserRole.MERCHANT
    phone: Optional[str] = None
    address: Optional[str] = None
    location_url: Optional[str] = None
    credit_limit: float = 0.0
    billing_day: Optional[int] = None
    allow_monthly_billing: bool = False


class UserLogin(BaseModel):
    """用户登录"""
    username: str
    password: str


# ============= 商品相关 =============
class ProductResponse(BaseModel):
    """商品响应 - 精简版，只返回必要字段"""
    id: int
    name: str
    name_kh: Optional[str] = None
    unit: str
    specs: Optional[str] = None
    barcode: Optional[str] = None
    price_usd: float
    retail_price_usd: Optional[float] = None
    stock: int
    is_low_stock: bool
    description: Optional[str] = None
    image_url: Optional[str] = None
    img1: Optional[str] = None
    img2: Optional[str] = None
    img3: Optional[str] = None
    img4: Optional[str] = None
    img5: Optional[str] = None
    category: Optional[str] = None
    sort_order: int = 0
    is_featured: bool = False


class ProductDetailResponse(BaseModel):
    """商品详情响应 - 完整版"""
    id: int
    name: str
    name_kh: Optional[str] = None
    unit: str
    specs: Optional[str] = None
    barcode: Optional[str] = None
    price_usd: float
    retail_price_usd: Optional[float] = None
    stock: int
    stock_warning: int
    is_low_stock: bool
    description: Optional[str] = None
    image_url: Optional[str] = None
    img1: Optional[str] = None
    img2: Optional[str] = None
    img3: Optional[str] = None
    img4: Optional[str] = None
    img5: Optional[str] = None
    thumbnail_url: Optional[str] = None
    category: Optional[str] = None
    sort_order: int = 0
    is_active: bool
    is_featured: bool = False


class ProductCreate(BaseModel):
    """创建商品"""
    name: str = Field(max_length=100)
    name_kh: Optional[str] = None
    unit: str = "件"
    specs: Optional[str] = None
    barcode: Optional[str] = None
    price_usd: float = Field(gt=0)
    retail_price_usd: Optional[float] = None
    stock: int = Field(default=0, ge=0)
    stock_warning: int = 10
    description: Optional[str] = None
    image_url: Optional[str] = None
    img1: Optional[str] = None
    img2: Optional[str] = None
    img3: Optional[str] = None
    img4: Optional[str] = None
    img5: Optional[str] = None
    category: Optional[str] = None
    sort_order: int = 0
    is_featured: bool = False


class ProductUpdate(BaseModel):
    """更新商品"""
    name: Optional[str] = None
    name_kh: Optional[str] = None
    unit: Optional[str] = None
    specs: Optional[str] = None
    barcode: Optional[str] = None
    price_usd: Optional[float] = Field(default=None, gt=0)
    retail_price_usd: Optional[float] = None
    stock: Optional[int] = Field(default=None, ge=0)
    stock_warning: Optional[int] = None
    description: Optional[str] = None
    image_url: Optional[str] = None
    img1: Optional[str] = None
    img2: Optional[str] = None
    img3: Optional[str] = None
    img4: Optional[str] = None
    img5: Optional[str] = None
    category: Optional[str] = None
    sort_order: Optional[int] = None
    is_active: Optional[bool] = None
    is_featured: Optional[bool] = None


# ============= 订单相关 =============
class OrderItemCreate(BaseModel):
    """创建订单项"""
    product_id: int
    quantity: int = Field(gt=0)


class OrderItemResponse(BaseModel):
    """订单项响应 - 精简版"""
    product_id: int
    product_name: str
    quantity: int
    unit_price_usd: float
    subtotal_usd: float


class OrderCreate(BaseModel):
    """创建订单"""
    items: list[OrderItemCreate]
    delivery_address: Optional[str] = None
    delivery_phone: Optional[str] = None
    note: Optional[str] = None
    payment_status: PaymentStatus = PaymentStatus.UNPAID


class OrderResponse(BaseModel):
    """订单响应 - 精简版"""
    id: int
    order_no: str
    total_usd: float
    payment_status: PaymentStatus
    delivery_status: DeliveryStatus
    created_at: str


class OrderDetailResponse(BaseModel):
    """订单详情响应 - 完整版"""
    id: int
    order_no: str
    merchant_id: int
    merchant_name: str
    total_usd: float
    total_khr: float
    payment_status: PaymentStatus
    delivery_status: DeliveryStatus
    delivery_address: Optional[str] = None
    delivery_phone: Optional[str] = None
    note: Optional[str] = None
    items: list[OrderItemResponse]
    created_at: str
    unpaid_days: Optional[int] = None  # 未回款天数
    days_to_billing: Optional[int] = None  # 距离下一个结账日天数


class OrderUpdate(BaseModel):
    """更新订单"""
    payment_status: Optional[PaymentStatus] = None
    delivery_status: Optional[DeliveryStatus] = None
    delivery_person_id: Optional[int] = None
    note: Optional[str] = None


# ============= 用户管理 =============
class UserUpdate(BaseModel):
    """更新用户"""
    full_name: Optional[str] = None
    phone: Optional[str] = None
    address: Optional[str] = None
    location_url: Optional[str] = None
    credit_limit: Optional[float] = None
    billing_day: Optional[int] = None
    allow_monthly_billing: Optional[bool] = None
    is_active: Optional[bool] = None
    role: Optional[UserRole] = None
    telegram_bot_token: Optional[str] = None
    telegram_chat_id: Optional[str] = None


class PasswordChange(BaseModel):
    """修改密码"""
    old_password: str
    new_password: str = Field(min_length=6)


# ============= 公告相关 =============
class AnnouncementCreate(BaseModel):
    """创建公告"""
    type: AnnouncementType = AnnouncementType.NOTICE
    content_zh: str = ""
    content_en: str = ""
    is_active: bool = True
    sort_order: int = 0


class AnnouncementUpdate(BaseModel):
    """更新公告"""
    content_zh: Optional[str] = None
    content_en: Optional[str] = None
    is_active: Optional[bool] = None
    sort_order: Optional[int] = None


class AnnouncementResponse(BaseModel):
    """公告响应"""
    id: int
    type: AnnouncementType
    content_zh: str
    content_en: str
    is_active: bool
    sort_order: int
    created_at: str
    updated_at: str


# ============= 月结账单相关 =============
class MonthlyBillResponse(BaseModel):
    """月结账单响应"""
    id: int
    merchant_id: int
    merchant_name: Optional[str] = None
    year: int
    month: int
    total_amount: float
    paid_amount: float
    status: BillStatus
    created_at: str


class MonthlyBillUpdate(BaseModel):
    """更新月结账单"""
    paid_amount: Optional[float] = None
    status: Optional[BillStatus] = None
