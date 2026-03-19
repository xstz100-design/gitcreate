from datetime import datetime, timezone, timedelta
from typing import Optional
from enum import Enum
from sqlmodel import SQLModel, Field

CAMBODIA_TZ = timezone(timedelta(hours=7))

def _now_cambodia():
    return datetime.now(CAMBODIA_TZ)


class AnnouncementType(str, Enum):
    """公告类型"""
    NOTICE = "notice"      # 滚动公告
    CONTACT = "contact"    # 联系客服信息
    ABOUT = "about"        # 关于系统信息


class Announcement(SQLModel, table=True):
    """公告/系统信息表"""
    __tablename__ = "announcements"

    id: Optional[int] = Field(default=None, primary_key=True)
    type: AnnouncementType = Field(default=AnnouncementType.NOTICE)
    content_zh: str = Field(max_length=2000)  # 中文内容
    content_en: str = Field(default="", max_length=2000)  # 英文内容
    is_active: bool = Field(default=True)
    sort_order: int = Field(default=0)
    created_at: datetime = Field(default_factory=_now_cambodia)
    updated_at: datetime = Field(default_factory=_now_cambodia)
