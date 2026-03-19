from fastapi import APIRouter, Depends, HTTPException, status
from sqlmodel import Session, select
from app.core.database import get_session
from app.core.dependencies import get_current_user, require_admin
from app.models import Announcement, AnnouncementType, User
from app.api.schemas import AnnouncementCreate, AnnouncementUpdate, AnnouncementResponse

router = APIRouter()


def _to_response(a: Announcement) -> AnnouncementResponse:
    return AnnouncementResponse(
        id=a.id,
        type=a.type,
        content_zh=a.content_zh,
        content_en=a.content_en,
        is_active=a.is_active,
        sort_order=a.sort_order,
        created_at=a.created_at.isoformat() if a.created_at else "",
        updated_at=a.updated_at.isoformat() if a.updated_at else "",
    )


@router.get("/public", response_model=list[AnnouncementResponse])
def list_public_announcements(
    type: AnnouncementType | None = None,
    session: Session = Depends(get_session),
):
    """获取公开公告 - 无需认证"""
    query = select(Announcement).where(Announcement.is_active == True)
    if type:
        query = query.where(Announcement.type == type)
    query = query.order_by(Announcement.sort_order.asc(), Announcement.id.desc())
    items = session.exec(query).all()
    return [_to_response(a) for a in items]


@router.get("", response_model=list[AnnouncementResponse])
def list_announcements(
    type: AnnouncementType | None = None,
    session: Session = Depends(get_session),
    current_user: User = Depends(require_admin),
):
    """获取全部公告 - 管理员"""
    query = select(Announcement)
    if type:
        query = query.where(Announcement.type == type)
    query = query.order_by(Announcement.sort_order.asc(), Announcement.id.desc())
    items = session.exec(query).all()
    return [_to_response(a) for a in items]


@router.post("", response_model=AnnouncementResponse, status_code=status.HTTP_201_CREATED)
def create_announcement(
    data: AnnouncementCreate,
    session: Session = Depends(get_session),
    current_user: User = Depends(require_admin),
):
    """创建公告 - 管理员"""
    announcement = Announcement(**data.model_dump())
    session.add(announcement)
    session.commit()
    session.refresh(announcement)
    return _to_response(announcement)


@router.patch("/{announcement_id}", response_model=AnnouncementResponse)
def update_announcement(
    announcement_id: int,
    data: AnnouncementUpdate,
    session: Session = Depends(get_session),
    current_user: User = Depends(require_admin),
):
    """更新公告 - 管理员"""
    announcement = session.get(Announcement, announcement_id)
    if not announcement:
        raise HTTPException(status_code=404, detail="公告不存在")
    
    from datetime import datetime, timezone, timedelta
    CAMBODIA_TZ = timezone(timedelta(hours=7))
    
    for key, value in data.model_dump(exclude_unset=True).items():
        setattr(announcement, key, value)
    announcement.updated_at = datetime.now(CAMBODIA_TZ).replace(tzinfo=None)
    
    session.add(announcement)
    session.commit()
    session.refresh(announcement)
    return _to_response(announcement)


@router.delete("/{announcement_id}")
def delete_announcement(
    announcement_id: int,
    session: Session = Depends(get_session),
    current_user: User = Depends(require_admin),
):
    """删除公告 - 管理员"""
    announcement = session.get(Announcement, announcement_id)
    if not announcement:
        raise HTTPException(status_code=404, detail="公告不存在")
    
    session.delete(announcement)
    session.commit()
    return {"message": "公告已删除"}
