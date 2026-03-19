from datetime import datetime, timezone, timedelta
from typing import Optional
from enum import Enum
from sqlmodel import SQLModel, Field, Relationship

# 柬埔寨时区 UTC+7
CAMBODIA_TZ = timezone(timedelta(hours=7))

def _now_cambodia():
    return datetime.now(CAMBODIA_TZ)


class BillStatus(str, Enum):
    """账单状态"""
    UNPAID = "unpaid"        # 未结清
    PAID = "paid"            # 已结清
    PARTIAL = "partial"      # 部分结清


class MonthlyBill(SQLModel, table=True):
    """月结账单"""
    __tablename__ = "monthly_bills"
    
    id: Optional[int] = Field(default=None, primary_key=True)
    
    # 关联商户
    merchant_id: int = Field(foreign_key="users.id", index=True)
    
    # 账期
    year: int  # 年份
    month: int  # 月份(1-12)
    
    # 金额
    total_amount: float = Field(default=0.0)  # 月结总金额
    paid_amount: float = Field(default=0.0)   # 已付金额
    
    # 状态
    status: BillStatus = Field(default=BillStatus.UNPAID)
    
    # 时间戳
    created_at: datetime = Field(default_factory=_now_cambodia)
    updated_at: datetime = Field(default_factory=_now_cambodia)
