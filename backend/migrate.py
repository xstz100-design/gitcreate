import sqlite3
conn = sqlite3.connect('/opt/wholesale/backend/cambodia_wholesale.db')
c = conn.cursor()
try:
    c.execute('ALTER TABLE products ADD COLUMN is_featured BOOLEAN DEFAULT 0')
    print('Added is_featured column')
except Exception as e:
    print(f'is_featured: {e}')
try:
    c.execute('''CREATE TABLE IF NOT EXISTS announcements (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        type VARCHAR(10) NOT NULL DEFAULT 'notice',
        content_zh TEXT NOT NULL DEFAULT '',
        content_en TEXT NOT NULL DEFAULT '',
        is_active BOOLEAN DEFAULT 1,
        sort_order INTEGER DEFAULT 0,
        created_at TIMESTAMP,
        updated_at TIMESTAMP
    )''')
    print('Created announcements table')
except Exception as e:
    print(f'announcements: {e}')
conn.commit()
conn.close()
print('Migration done')
