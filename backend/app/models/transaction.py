from datetime import datetime
from typing import Optional
from enum import Enum
from sqlmodel import SQLModel, Field, Relationship


class TransactionType(str, Enum):
    """交易类型"""
    PAYMENT = "payment"      # 收款
    REFUND = "refund"       # 退款
    EXPENSE = "expense"     # 支出(采购等)
    ADJUSTMENT = "adjustment"  # 调整


class Transaction(SQLModel, table=True):
    """账流水表 - 记录所有金钱进出"""
    __tablename__ = "transactions"
    
    id: Optional[int] = Field(default=None, primary_key=True)
    
    # 关联用户(商户或员工)
    user_id: Optional[int] = Field(default=None, foreign_key="users.id")
    user: Optional["User"] = Relationship(
        back_populates="transactions",
        sa_relationship_kwargs={"foreign_keys": "[Transaction.user_id]"}
    )
    
    # 金额和类型
    amount_usd: float
    transaction_type: TransactionType
    
    # 关联订单(可选)
    order_id: Optional[int] = Field(default=None, foreign_key="orders.id")
    
    # 描述
    description: str = Field(max_length=200)
    note: Optional[str] = Field(default=None, max_length=500)
    
    # 时间戳
    created_at: datetime = Field(default_factory=datetime.utcnow)
    created_by: Optional[int] = Field(default=None, foreign_key="users.id")  # 操作人
    
    @property
    def amount_khr(self) -> float:
        """金额(瑞尔)"""
        from app.core.config import settings
        return self.amount_usd * settings.USD_TO_KHR_RATE
