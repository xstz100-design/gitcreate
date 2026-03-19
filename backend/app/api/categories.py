"""分类管理接口"""
from fastapi import APIRouter, Depends, HTTPException, status
from sqlmodel import Session, select
from pydantic import BaseModel, Field
from typing import Optional
from app.core.database import get_session
from app.core.dependencies import get_current_user, require_admin
from app.models import Category, User

router = APIRouter()


# ========== Schemas ==========

class CategoryCreate(BaseModel):
    name: str = Field(max_length=50)
    sort_order: int = 0


class CategoryUpdate(BaseModel):
    name: Optional[str] = None
    sort_order: Optional[int] = None
    is_active: Optional[bool] = None


class CategoryResponse(BaseModel):
    id: int
    name: str
    sort_order: int
    is_active: bool


# ========== Routes ==========

@router.get("", response_model=list[CategoryResponse])
def list_categories(
    session: Session = Depends(get_session),
    current_user: User = Depends(get_current_user),
):
    """获取分类列表"""
    categories = session.exec(
        select(Category).where(Category.is_active == True).order_by(Category.sort_order)
    ).all()
    return categories


@router.get("/all", response_model=list[CategoryResponse])
def list_all_categories(
    session: Session = Depends(get_session),
    current_user: User = Depends(require_admin),
):
    """获取所有分类(含禁用) - 仅管理员"""
    categories = session.exec(
        select(Category).order_by(Category.sort_order)
    ).all()
    return categories


@router.post("", response_model=CategoryResponse, status_code=status.HTTP_201_CREATED)
def create_category(
    data: CategoryCreate,
    session: Session = Depends(get_session),
    current_user: User = Depends(require_admin),
):
    """创建分类 - 仅管理员"""
    existing = session.exec(
        select(Category).where(Category.name == data.name)
    ).first()
    if existing:
        raise HTTPException(status_code=400, detail="分类名称已存在")
    
    category = Category(**data.model_dump())
    session.add(category)
    session.commit()
    session.refresh(category)
    return category


@router.patch("/{category_id}", response_model=CategoryResponse)
def update_category(
    category_id: int,
    data: CategoryUpdate,
    session: Session = Depends(get_session),
    current_user: User = Depends(require_admin),
):
    """更新分类 - 仅管理员"""
    category = session.get(Category, category_id)
    if not category:
        raise HTTPException(status_code=404, detail="分类不存在")
    
    for key, value in data.model_dump(exclude_unset=True).items():
        setattr(category, key, value)
    
    session.add(category)
    session.commit()
    session.refresh(category)
    return category


@router.delete("/{category_id}")
def delete_category(
    category_id: int,
    session: Session = Depends(get_session),
    current_user: User = Depends(require_admin),
):
    """删除分类(硬删除) - 仅管理员"""
    category = session.get(Category, category_id)
    if not category:
        raise HTTPException(status_code=404, detail="分类不存在")
    
    # 检查是否有商品使用该分类
    from app.models import Product
    products_using = session.exec(
        select(Product).where(Product.category == category.name, Product.is_active == True)
    ).first()
    if products_using:
        raise HTTPException(status_code=400, detail="该分类下有商品，无法删除，请先移除相关商品的分类")
    
    session.delete(category)
    session.commit()
    return {"message": "分类已删除"}
