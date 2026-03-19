from typing import Optional
from pydantic import BaseModel
from fastapi import APIRouter, Depends, HTTPException, status, Request
from fastapi.security import OAuth2PasswordRequestForm
from sqlmodel import Session, select, func
from app.core.database import get_session
from app.core.security import verify_password, get_password_hash, create_access_token
from app.core.dependencies import get_current_user, require_admin
from app.core.rate_limit import login_protector
from app.models import User, UserRole
from app.api.schemas import UserCreate, UserLogin, UserResponse, TokenResponse, MessageResponse, UserUpdate, PasswordChange

router = APIRouter()


def _user_response(user: User) -> UserResponse:
    """构建 UserResponse"""
    return UserResponse(
        id=user.id,
        username=user.username,
        full_name=user.full_name,
        role=user.role,
        phone=user.phone,
        address=user.address,
        credit_limit=user.credit_limit,
        billing_day=user.billing_day,
        allow_monthly_billing=user.allow_monthly_billing,
        location_url=user.location_url,
        store_photo=user.store_photo,
        telegram_bot_token=user.telegram_bot_token,
        telegram_chat_id=user.telegram_chat_id,
        must_change_password=user.must_change_password,
        is_active=user.is_active,
    )


def _generate_account_number(session: Session, role: UserRole) -> str:
    """根据角色自动生成6位账号"""
    if role == UserRole.ADMIN:
        start, end = 100001, 200000
    else:
        start, end = 200001, 299999

    # 查找该范围内已有的最大账号
    existing = session.exec(
        select(func.max(User.username))
        .where(User.username >= str(start))
        .where(User.username <= str(end))
    ).first()

    if existing:
        next_num = int(existing) + 1
    else:
        next_num = start

    if next_num > end:
        raise HTTPException(status_code=400, detail="账号已用尽")

    return str(next_num)


@router.post("/register", response_model=UserResponse, status_code=status.HTTP_201_CREATED)
def register(
    user_data: UserCreate,
    session: Session = Depends(get_session),
    current_user: User = Depends(require_admin),
):
    """注册新用户 - 仅管理员，账号自动生成"""
    # 自动生成账号
    username = _generate_account_number(session, user_data.role)

    # 创建新用户，默认密码 123456
    user = User(
        username=username,
        hashed_password=get_password_hash("123456"),
        full_name=user_data.full_name,
        role=user_data.role,
        phone=user_data.phone,
        address=user_data.address,
        location_url=user_data.location_url,
        credit_limit=user_data.credit_limit,
        billing_day=user_data.billing_day,
        allow_monthly_billing=user_data.allow_monthly_billing,
        must_change_password=True,
    )

    session.add(user)
    session.commit()
    session.refresh(user)

    return _user_response(user)


@router.post("/login", response_model=TokenResponse)
def login(
    request: Request,
    form_data: OAuth2PasswordRequestForm = Depends(),
    session: Session = Depends(get_session),
):
    """用户登录 - 含防暴力破解"""
    client_ip = request.client.host if request.client else "unknown"

    # --- 防暴力破解: 登录前检查频率 ---
    login_protector.check_login(client_ip, form_data.username)

    # 查找用户
    user = session.exec(
        select(User).where(User.username == form_data.username)
    ).first()

    if not user or not verify_password(form_data.password, user.hashed_password):
        # 记录失败
        login_protector.record_failure(client_ip, form_data.username)
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="用户名或密码错误",
            headers={"WWW-Authenticate": "Bearer"},
        )
    
    if not user.is_active:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="账号已被禁用"
        )
    
    # 创建 access token - sub必须是字符串
    access_token = create_access_token(data={"sub": str(user.id)})
    
    return TokenResponse(
        access_token=access_token,
        user=_user_response(user)
    )


@router.get("/me", response_model=UserResponse)
def get_current_user_info(
    current_user: User = Depends(get_current_user),
):
    """获取当前用户信息"""
    return _user_response(current_user)


@router.get("/users", response_model=list[UserResponse])
def list_users(
    role: UserRole | None = None,
    session: Session = Depends(get_session),
    current_user: User = Depends(require_admin),
):
    """获取用户列表 - 仅管理员"""
    query = select(User)
    if role:
        query = query.where(User.role == role)
    
    users = session.exec(query).all()
    return [_user_response(user) for user in users]


@router.patch("/users/{user_id}", response_model=UserResponse)
def update_user(
    user_id: int,
    user_data: UserUpdate,
    session: Session = Depends(get_session),
    current_user: User = Depends(require_admin),
):
    """更新用户信息 - 仅管理员"""
    user = session.get(User, user_id)
    if not user:
        raise HTTPException(status_code=404, detail="用户不存在")
    
    # 保护超级管理员(100001)不能被禁用
    if user.username == "100001" and user_data.is_active is not None and not user_data.is_active:
        raise HTTPException(status_code=400, detail="不能禁用超级管理员")
    
    for key, value in user_data.model_dump(exclude_unset=True).items():
        setattr(user, key, value)
    
    session.add(user)
    session.commit()
    session.refresh(user)
    
    return _user_response(user)


@router.post("/users/{user_id}/reset-password", response_model=MessageResponse)
def reset_password(
    user_id: int,
    session: Session = Depends(get_session),
    current_user: User = Depends(require_admin),
):
    """重置用户密码为123456 - 仅超级管理员(100001)"""
    if current_user.username != "100001":
        raise HTTPException(status_code=403, detail="仅超级管理员可重置密码")
    user = session.get(User, user_id)
    if not user:
        raise HTTPException(status_code=404, detail="用户不存在")
    user.hashed_password = get_password_hash("123456")
    user.must_change_password = True
    session.add(user)
    session.commit()
    return MessageResponse(message="密码已重置为123456")


@router.delete("/users/{user_id}", response_model=MessageResponse)
def delete_user(
    user_id: int,
    session: Session = Depends(get_session),
    current_user: User = Depends(require_admin),
):
    """删除用户 - 仅最高管理员(100001)"""
    if current_user.username != "100001":
        raise HTTPException(status_code=403, detail="仅最高管理员可删除用户")
    if user_id == current_user.id:
        raise HTTPException(status_code=400, detail="不能删除自己")
    user = session.get(User, user_id)
    if not user:
        raise HTTPException(status_code=404, detail="用户不存在")
    # 检查是否有关联订单
    from app.models import Order
    order_count = session.exec(
        select(func.count(Order.id)).where(Order.merchant_id == user_id)
    ).one()
    if order_count > 0:
        raise HTTPException(status_code=400, detail=f"该用户有 {order_count} 条关联订单，无法删除，建议停用")
    session.delete(user)
    session.commit()
    return MessageResponse(message="用户已删除")


@router.post("/change-password", response_model=MessageResponse)
def change_password(
    data: PasswordChange,
    session: Session = Depends(get_session),
    current_user: User = Depends(get_current_user),
):
    """修改密码 - 当前用户"""
    if not verify_password(data.old_password, current_user.hashed_password):
        raise HTTPException(status_code=400, detail="原密码错误")
    
    current_user.hashed_password = get_password_hash(data.new_password)
    current_user.must_change_password = False
    session.add(current_user)
    session.commit()
    
    return MessageResponse(message="密码修改成功")


class ProfileUpdate(BaseModel):
    """用户自行更新个人信息"""
    full_name: Optional[str] = None
    phone: Optional[str] = None
    address: Optional[str] = None
    location_url: Optional[str] = None
    store_photo: Optional[str] = None
    telegram_bot_token: Optional[str] = None
    telegram_chat_id: Optional[str] = None


@router.patch("/me", response_model=UserResponse)
def update_profile(
    data: ProfileUpdate,
    session: Session = Depends(get_session),
    current_user: User = Depends(get_current_user),
):
    """更新个人信息 - 当前用户"""
    for key, value in data.model_dump(exclude_unset=True).items():
        setattr(current_user, key, value)
    
    session.add(current_user)
    session.commit()
    session.refresh(current_user)
    
    return _user_response(current_user)
