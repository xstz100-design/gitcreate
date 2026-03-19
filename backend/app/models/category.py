from datetime import datetime
from typing import Optional
from sqlmodel import SQLModel, Field


class Category(SQLModel, table=True):
    """商品分类表"""
    __tablename__ = "categories"
    
    id: Optional[int] = Field(default=None, primary_key=True)
    name: str = Field(max_length=50, unique=True, index=True)
    sort_order: int = Field(default=0)  # 排序序号，越小越靠前
    is_active: bool = Field(default=True)
    created_at: datetime = Field(default_factory=datetime.utcnow)
