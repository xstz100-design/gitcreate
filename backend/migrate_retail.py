import sqlite3
import os

# Try both possible DB paths
for db_name in ['cambodia_wholesale.db', 'wholesale.db']:
    db_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), db_name)
    if not os.path.exists(db_path):
        print(f'{db_name} not found at {db_path}')
        continue
    
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    
    # Check tables
    tables = [t[0] for t in cursor.execute("SELECT name FROM sqlite_master WHERE type='table'").fetchall()]
    print(f'{db_name} tables: {tables}')
    
    if 'products' not in tables:
        print(f'No products table in {db_name}, skipping')
        conn.close()
        continue
    
    cursor.execute('PRAGMA table_info(products)')
    cols = [c[1] for c in cursor.fetchall()]
    print(f'Existing columns: {cols}')
    
    if 'retail_price_usd' not in cols:
        cursor.execute('ALTER TABLE products ADD COLUMN retail_price_usd FLOAT')
        conn.commit()
        print(f'Column retail_price_usd added to {db_name}')
    else:
        print(f'Column retail_price_usd already exists in {db_name}')
    
    conn.close()
