from datetime import datetime, timezone, timedelta
from fastapi import APIRouter, Depends, HTTPException
from sqlmodel import Session, select, func
from app.core.database import get_session
from app.core.dependencies import get_current_user, require_admin
from app.models import MonthlyBill, BillStatus, Order, PaymentStatus, DeliveryStatus, User, UserRole
from app.api.schemas import MonthlyBillResponse, MonthlyBillUpdate, MessageResponse

router = APIRouter()

CAMBODIA_TZ = timezone(timedelta(hours=7))


def _bill_response(bill, merchant_name: str = None) -> MonthlyBillResponse:
    return MonthlyBillResponse(
        id=bill.id,
        merchant_id=bill.merchant_id,
        merchant_name=merchant_name or "",
        year=bill.year,
        month=bill.month,
        total_amount=bill.total_amount,
        paid_amount=bill.paid_amount,
        status=bill.status,
        created_at=bill.created_at.isoformat() if bill.created_at else "",
    )


@router.get("", response_model=list[MonthlyBillResponse])
def list_monthly_bills(
    merchant_id: int | None = None,
    year: int | None = None,
    month: int | None = None,
    session: Session = Depends(get_session),
    current_user: User = Depends(get_current_user),
):
    """获取月结账单列表"""
    query = select(MonthlyBill)

    if current_user.role == UserRole.MERCHANT:
        query = query.where(MonthlyBill.merchant_id == current_user.id)
    elif merchant_id:
        query = query.where(MonthlyBill.merchant_id == merchant_id)

    if year:
        query = query.where(MonthlyBill.year == year)
    if month:
        query = query.where(MonthlyBill.month == month)

    bills = session.exec(query.order_by(MonthlyBill.year.desc(), MonthlyBill.month.desc())).all()

    # 获取商户名称
    merchant_ids = list(set(b.merchant_id for b in bills))
    merchants = {}
    if merchant_ids:
        users = session.exec(select(User).where(User.id.in_(merchant_ids))).all()
        merchants = {u.id: u.full_name for u in users}

    return [_bill_response(b, merchants.get(b.merchant_id, "")) for b in bills]


@router.post("/generate", response_model=MessageResponse)
def generate_monthly_bills(
    year: int,
    month: int,
    session: Session = Depends(get_session),
    current_user: User = Depends(require_admin),
):
    """生成指定月份的月结账单 - 管理员"""
    # 获取所有有月结日的商户
    merchants = session.exec(
        select(User).where(User.role == UserRole.MERCHANT, User.billing_day.isnot(None))
    ).all()

    count = 0
    for merchant in merchants:
        # 检查是否已生成
        existing = session.exec(
            select(MonthlyBill).where(
                MonthlyBill.merchant_id == merchant.id,
                MonthlyBill.year == year,
                MonthlyBill.month == month,
            )
        ).first()
        if existing:
            continue

        # 计算该月赊账/待付款订单总额
        # 月结周期: 上月billing_day到本月billing_day
        billing_day = merchant.billing_day or 1

        # 计算起止日期
        if month == 1:
            start_year, start_month = year - 1, 12
        else:
            start_year, start_month = year, month - 1

        import calendar
        max_day_start = calendar.monthrange(start_year, start_month)[1]
        max_day_end = calendar.monthrange(year, month)[1]
        actual_start_day = min(billing_day, max_day_start)
        actual_end_day = min(billing_day, max_day_end)

        start_date = datetime(start_year, start_month, actual_start_day, tzinfo=CAMBODIA_TZ)
        end_date = datetime(year, month, actual_end_day, tzinfo=CAMBODIA_TZ)

        # 查询该时段内该商户的订单
        total = session.exec(
            select(func.coalesce(func.sum(Order.total_usd), 0)).where(
                Order.merchant_id == merchant.id,
                Order.created_at >= start_date,
                Order.created_at < end_date,
                Order.delivery_status != DeliveryStatus.CANCELLED,
            )
        ).one()

        bill = MonthlyBill(
            merchant_id=merchant.id,
            year=year,
            month=month,
            total_amount=float(total),
            paid_amount=0.0,
            status=BillStatus.UNPAID if float(total) > 0 else BillStatus.PAID,
        )
        session.add(bill)
        count += 1

    session.commit()
    return MessageResponse(message=f"已生成 {count} 条月结账单")


@router.patch("/{bill_id}", response_model=MonthlyBillResponse)
def update_monthly_bill(
    bill_id: int,
    data: MonthlyBillUpdate,
    session: Session = Depends(get_session),
    current_user: User = Depends(require_admin),
):
    """更新月结账单 - 管理员"""
    bill = session.get(MonthlyBill, bill_id)
    if not bill:
        raise HTTPException(status_code=404, detail="账单不存在")

    for key, value in data.model_dump(exclude_unset=True).items():
        setattr(bill, key, value)

    # 自动计算状态
    if bill.paid_amount >= bill.total_amount:
        bill.status = BillStatus.PAID
    elif bill.paid_amount > 0:
        bill.status = BillStatus.PARTIAL
    else:
        bill.status = BillStatus.UNPAID

    bill.updated_at = datetime.now(CAMBODIA_TZ)
    session.add(bill)
    session.commit()
    session.refresh(bill)

    merchant = session.get(User, bill.merchant_id)
    return _bill_response(bill, merchant.full_name if merchant else "")
