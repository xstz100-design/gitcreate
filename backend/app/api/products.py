from fastapi import APIRouter, Depends, HTTPException, status, Query
from sqlmodel import Session, select
from app.core.database import get_session
from app.core.dependencies import get_current_user, require_admin
from app.models import Product, User
from app.api.schemas import ProductCreate, ProductUpdate, ProductResponse, ProductDetailResponse

router = APIRouter()


def _product_response(p) -> ProductResponse:
    return ProductResponse(
        id=p.id, name=p.name, name_kh=p.name_kh, unit=p.unit,
        specs=p.specs, barcode=p.barcode, price_usd=p.price_usd,
        retail_price_usd=p.retail_price_usd, stock=p.stock,
        stock_warning=getattr(p, 'stock_warning', 10),
        is_low_stock=p.is_low_stock, description=p.description,
        image_url=p.image_url, img1=p.img1, img2=p.img2, img3=p.img3,
        img4=p.img4, img5=p.img5, category=p.category,
        sort_order=p.sort_order, is_active=p.is_active, is_featured=p.is_featured,
    )


@router.post("", response_model=ProductResponse, status_code=status.HTTP_201_CREATED)
def create_product(
    product_data: ProductCreate,
    session: Session = Depends(get_session),
    current_user: User = Depends(require_admin),
):
    """创建商品 - 仅管理员"""
    product = Product(**product_data.model_dump())
    session.add(product)
    session.commit()
    session.refresh(product)
    return _product_response(product)


@router.get("", response_model=list[ProductResponse])
def list_products(
    category: str | None = None,
    is_active: bool | None = True,
    low_stock_only: bool = False,
    session: Session = Depends(get_session),
    current_user: User = Depends(get_current_user),
):
    """获取商品列表"""
    query = select(Product)
    if is_active is not None:
        query = query.where(Product.is_active == is_active)
    if category:
        query = query.where(Product.category == category)
    if low_stock_only:
        query = query.where(Product.stock <= Product.stock_warning)
    products = session.exec(query).all()
    return [_product_response(p) for p in products]


@router.get("/{product_id}", response_model=ProductDetailResponse)
def get_product(
    product_id: int,
    session: Session = Depends(get_session),
    current_user: User = Depends(get_current_user),
):
    """获取商品详情"""
    product = session.get(Product, product_id)
    if not product:
        raise HTTPException(status_code=404, detail="商品不存在")
    return ProductDetailResponse(
        id=product.id, name=product.name, name_kh=product.name_kh,
        unit=product.unit, specs=product.specs, barcode=product.barcode,
        price_usd=product.price_usd, retail_price_usd=product.retail_price_usd,
        stock=product.stock, stock_warning=product.stock_warning,
        is_low_stock=product.is_low_stock, description=product.description,
        image_url=product.image_url, img1=product.img1, img2=product.img2,
        img3=product.img3, img4=product.img4, img5=product.img5,
        thumbnail_url=product.image_url, category=product.category,
        sort_order=product.sort_order, is_active=product.is_active,
        is_featured=product.is_featured,
    )


@router.patch("/{product_id}", response_model=ProductResponse)
def update_product(
    product_id: int,
    product_data: ProductUpdate,
    session: Session = Depends(get_session),
    current_user: User = Depends(require_admin),
):
    """更新商品 - 仅管理员"""
    product = session.get(Product, product_id)
    if not product:
        raise HTTPException(status_code=404, detail="商品不存在")
    for key, value in product_data.model_dump(exclude_unset=True).items():
        setattr(product, key, value)
    session.add(product)
    session.commit()
    session.refresh(product)
    # Telegram 通知: 库存预警
    try:
        if product.stock <= product.stock_warning:
            from app.services.telegram import notify_admins_low_stock
            notify_admins_low_stock(session, product_name=product.name,
                current_stock=product.stock, warning_level=product.stock_warning)
    except Exception:
        pass
    return _product_response(product)


@router.delete("/{product_id}")
def delete_product(
    product_id: int,
    session: Session = Depends(get_session),
    current_user: User = Depends(require_admin),
):
    """删除商品(软删除) - 仅管理员"""
    product = session.get(Product, product_id)
    if not product:
        raise HTTPException(status_code=404, detail="商品不存在")
    product.is_active = False
    session.add(product)
    session.commit()
    return {"message": "商品已删除"}
