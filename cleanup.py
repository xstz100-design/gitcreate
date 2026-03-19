from sqlmodel import Session, select
from app.core.database import engine
from app.models import Product

with Session(engine) as s:
    for p in s.exec(select(Product).where(Product.name.like('test%'))).all():
        print(f'Deleting: {p.id} {p.name}')
        s.delete(p)
    s.commit()
    print('Done')
