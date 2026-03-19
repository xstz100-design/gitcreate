
from datetime import datetime, timezone, timedelta
from fastapi import APIRouter, Depends, HTTPException, status
from sqlmodel import Session, select, func
from sqlalchemy.orm import selectinload
from app.core.database import get_session
from app.core.dependencies import get_current_user, require_admin
from app.models import Order, OrderItem, Product, User, UserRole, DeliveryStatus, PaymentStatus
from app.api.schemas import OrderCreate, OrderDetailResponse, OrderUpdate, OrderItemResponse

router = APIRouter()

# 柬埔寨时区 UTC+7
CAMBODIA_TZ = timezone(timedelta(hours=7))


def now_cambodia() -> datetime:
    """获取柬埔寨当前时间"""
    return datetime.now(CAMBODIA_TZ)


def calc_unpaid_days(order) -> int | None:
    """计算未回款天数: 未支付状态的订单返回天数"""
    if order.payment_status == PaymentStatus.CASH:
        return None
    if order.delivery_status == DeliveryStatus.CANCELLED:
        return None
    now = now_cambodia()
    created = order.created_at
    if created.tzinfo is None:
        created = created.replace(tzinfo=CAMBODIA_TZ)
    delta = now - created
    return delta.days


def calc_days_to_billing(merchant) -> int | None:
    """计算距离下一个结账日的天数"""
    if not merchant or not merchant.billing_day:
        return None
    now = now_cambodia()
    billing_day = merchant.billing_day
    # 尝试本月的结账日
    import calendar
    year, month = now.year, now.month
    max_day = calendar.monthrange(year, month)[1]
    actual_day = min(billing_day, max_day)
    from datetime import date
    billing_date = date(year, month, actual_day)
    today = now.date()
    if billing_date > today:
        return (billing_date - today).days
    # 本月结账日已过，算下个月
    if month == 12:
        year += 1
        month = 1
    else:
        month += 1
    max_day = calendar.monthrange(year, month)[1]
    actual_day = min(billing_day, max_day)
    billing_date = date(year, month, actual_day)
    return (billing_date - today).days


def _build_order_response(order, merchant_name: str = None, items_data=None) -> OrderDetailResponse:
    """统一构建订单响应"""
    if merchant_name is None:
        merchant_name = order.merchant.full_name if order.merchant else ""
    if items_data is None:
        items_data = [
            OrderItemResponse(
                id=item.id,
                product_id=item.product_id,
                product_name=item.product.name if item.product else "已删除",
                quantity=item.quantity,
                unit_price_usd=item.unit_price_usd,
                subtotal_usd=item.subtotal_usd,
            )
            for item in order.items
        ]
    return OrderDetailResponse(
        id=order.id,
        order_no=order.order_no,
        merchant_id=order.merchant_id,
        merchant_name=merchant_name,
        total_usd=order.total_usd,
        total_khr=order.total_khr,
        payment_status=order.payment_status,
        delivery_status=order.delivery_status,
        delivery_address=order.delivery_address,
        delivery_phone=order.delivery_phone,
        note=order.note,
        items=items_data,
        created_at=order.created_at.isoformat(),
        unpaid_days=calc_unpaid_days(order),
        days_to_billing=calc_days_to_billing(order.merchant) if order.payment_status == PaymentStatus.MONTHLY else None,
    )


def generate_order_no(session: Session) -> str:
    """生成订单号 - 基于当天订单计数: ORD + YYYYMMDD + 6位序号"""
    today = now_cambodia()
    date_str = today.strftime('%Y%m%d')
    prefix = f"ORD{date_str}"
    
    # 查询今天已有的订单数量
    count = session.exec(
        select(func.count(Order.id)).where(Order.order_no.startswith(prefix))
    ).one()
    
    seq = count + 1
    return f"{prefix}{seq:06d}"


@router.post("", response_model=OrderDetailResponse, status_code=status.HTTP_201_CREATED)
def create_order(
    order_data: OrderCreate,
    session: Session = Depends(get_session),
    current_user: User = Depends(get_current_user),
):
    """创建订单 - 商户下单"""
    if not order_data.items:
        raise HTTPException(status_code=400, detail="订单必须包含至少一个商品")
    
    # 验证月结权限
    if order_data.payment_status == PaymentStatus.MONTHLY:
        if not current_user.allow_monthly_billing:
            raise HTTPException(status_code=400, detail="您没有月结权限，请联系管理员开通")
    
    # 验证库存并计算总价
    total_usd = 0.0
    order_items_data = []
    
    for item in order_data.items:
        product = session.get(Product, item.product_id)
        if not product:
            raise HTTPException(status_code=404, detail=f"商品 ID {item.product_id} 不存在")
        
        if not product.is_active:
            raise HTTPException(status_code=400, detail=f"商品 {product.name} 已下架")
        
        if product.stock < item.quantity:
            raise HTTPException(
                status_code=400,
                detail=f"商品 {product.name} 库存不足，当前库存: {product.stock}"
            )
        
        # 预扣库存
        product.stock -= item.quantity
        
        # 计算小计
        subtotal = product.price_usd * item.quantity
        total_usd += subtotal
        
        order_items_data.append({
            "product_id": product.id,
            "product": product,
            "quantity": item.quantity,
            "unit_price_usd": product.price_usd,
            "subtotal_usd": subtotal,
        })
    
    # 创建订单
    order = Order(
        order_no=generate_order_no(session),
        merchant_id=current_user.id,
        total_usd=total_usd,
        payment_status=order_data.payment_status,
        delivery_address=order_data.delivery_address or current_user.address,
        delivery_phone=order_data.delivery_phone or current_user.phone,
        note=order_data.note,
    )
    
    session.add(order)
    session.flush()  # 获取 order.id
    
    # 创建订单项
    for item_data in order_items_data:
        order_item = OrderItem(
            order_id=order.id,
            product_id=item_data["product_id"],
            quantity=item_data["quantity"],
            unit_price_usd=item_data["unit_price_usd"],
            subtotal_usd=item_data["subtotal_usd"],
        )
        session.add(order_item)
    
    # 月结订单: 累加用户月结金额
    if order_data.payment_status == PaymentStatus.MONTHLY:
        current_user.credit_limit = (current_user.credit_limit or 0) + total_usd
        session.add(current_user)
    
    session.commit()
    session.refresh(order)
    
    # Telegram 通知: 新订单
    try:
        from app.services.telegram import notify_admins_new_order
        notify_admins_new_order(
            session,
            order_no=order.order_no,
            merchant_name=current_user.full_name,
            total_usd=order.total_usd,
            items=[
                {
                    "product_name": item_data["product"].name,
                    "quantity": item_data["quantity"],
                    "subtotal_usd": item_data["subtotal_usd"],
                }
                for item_data in order_items_data
            ],
        )
    except Exception:
        pass  # 通知失败不影响订单创建
    
    # Telegram 通知: 库存预警
    try:
        from app.services.telegram import notify_admins_low_stock
        for item_data in order_items_data:
            product = item_data["product"]
            session.refresh(product)
            if product.stock <= product.stock_warning:
                notify_admins_low_stock(
                    session,
                    product_name=product.name,
                    current_stock=product.stock,
                    warning_level=product.stock_warning,
                )
    except Exception:
        pass
    
    # 构建响应
    return _build_order_response(
        order,
        merchant_name=current_user.full_name,
        items_data=[
            OrderItemResponse(
                id=item_data["product_id"],
                product_id=item_data["product_id"],
                product_name=item_data["product"].name,
                quantity=item_data["quantity"],
                unit_price_usd=item_data["unit_price_usd"],
                subtotal_usd=item_data["subtotal_usd"],
            )
            for item_data in order_items_data
        ],
    )


@router.get("", response_model=list[OrderDetailResponse])
def list_orders(
    merchant_id: int | None = None,
    payment_status: PaymentStatus | None = None,
    delivery_status: DeliveryStatus | None = None,
    session: Session = Depends(get_session),
    current_user: User = Depends(get_current_user),
):
    """获取订单列表"""
    query = select(Order).options(
        selectinload(Order.items).selectinload(OrderItem.product),
        selectinload(Order.merchant),
    )
    
    # 商户只能看自己的订单
    if current_user.role == UserRole.MERCHANT:
        query = query.where(Order.merchant_id == current_user.id)
    elif merchant_id:  # 管理员可以按商户筛选
        query = query.where(Order.merchant_id == merchant_id)
    
    if payment_status:
        query = query.where(Order.payment_status == payment_status)
    
    if delivery_status:
        query = query.where(Order.delivery_status == delivery_status)
    
    orders = session.exec(query.order_by(Order.created_at.desc())).all()
    
    return [_build_order_response(order) for order in orders]


@router.get("/{order_id}", response_model=OrderDetailResponse)
def get_order(
    order_id: int,
    session: Session = Depends(get_session),
    current_user: User = Depends(get_current_user),
):
    """获取订单详情"""
    order = session.exec(
        select(Order)
        .where(Order.id == order_id)
        .options(
            selectinload(Order.items).selectinload(OrderItem.product),
            selectinload(Order.merchant),
        )
    ).first()
    if not order:
        raise HTTPException(status_code=404, detail="订单不存在")
    
    # 商户只能看自己的订单
    if current_user.role == UserRole.MERCHANT and order.merchant_id != current_user.id:
        raise HTTPException(status_code=403, detail="无权访问此订单")
    
    return _build_order_response(order)


@router.patch("/{order_id}", response_model=OrderDetailResponse)
def update_order(
    order_id: int,
    order_data: OrderUpdate,
    session: Session = Depends(get_session),
    current_user: User = Depends(require_admin),
):
    """更新订单状态 - 仅管理员"""
    order = session.exec(
        select(Order)
        .where(Order.id == order_id)
        .options(
            selectinload(Order.items).selectinload(OrderItem.product),
            selectinload(Order.merchant),
        )
    ).first()
    if not order:
        raise HTTPException(status_code=404, detail="订单不存在")
    
    # 记录旧的支付状态，用于月结金额联动
    old_payment_status = order.payment_status
    
    # 更新字段
    for key, value in order_data.model_dump(exclude_unset=True).items():
        setattr(order, key, value)
    
    # 支付状态变更时联动月结金额
    if order_data.payment_status is not None and order_data.payment_status != old_payment_status:
        merchant = session.get(User, order.merchant_id)
        if merchant:
            # 旧状态是月结 → 扣减月结金额
            if old_payment_status == PaymentStatus.MONTHLY:
                merchant.credit_limit = max(0, (merchant.credit_limit or 0) - order.total_usd)
            # 新状态是月结 → 增加月结金额
            if order_data.payment_status == PaymentStatus.MONTHLY:
                merchant.credit_limit = (merchant.credit_limit or 0) + order.total_usd
            session.add(merchant)
    
    # 如果标记为已签收,记录时间
    if order_data.delivery_status == DeliveryStatus.DELIVERED and not order.delivered_at:
        order.delivered_at = now_cambodia()
    
    session.add(order)
    session.commit()
    session.refresh(order)
    
    return _build_order_response(order)


@router.post("/{order_id}/cancel", response_model=OrderDetailResponse)
def cancel_order(
    order_id: int,
    session: Session = Depends(get_session),
    current_user: User = Depends(get_current_user),
):
    """取消订单 - 商户在管理员确认签收/已支付之前可以取消"""
    order = session.exec(
        select(Order)
        .where(Order.id == order_id)
        .options(
            selectinload(Order.items).selectinload(OrderItem.product),
            selectinload(Order.merchant),
        )
    ).first()
    if not order:
        raise HTTPException(status_code=404, detail="订单不存在")

    # 商户只能取消自己的订单
    if current_user.role == UserRole.MERCHANT and order.merchant_id != current_user.id:
        raise HTTPException(status_code=403, detail="无权操作此订单")

    # 已签收或现结的订单不允许取消
    if order.delivery_status == DeliveryStatus.DELIVERED:
        raise HTTPException(status_code=400, detail="已签收的订单不允许取消")
    if order.delivery_status == DeliveryStatus.CANCELLED:
        raise HTTPException(status_code=400, detail="订单已取消")

    # 回滚库存
    for item in order.items:
        product = session.get(Product, item.product_id)
        if product:
            product.stock += item.quantity

    # 月结订单取消时扣减用户月结金额
    if order.payment_status == PaymentStatus.MONTHLY:
        merchant = session.get(User, order.merchant_id)
        if merchant:
            merchant.credit_limit = max(0, (merchant.credit_limit or 0) - order.total_usd)
            session.add(merchant)

    order.delivery_status = DeliveryStatus.CANCELLED
    session.add(order)
    session.commit()
    session.refresh(order)

    return _build_order_response(order)
