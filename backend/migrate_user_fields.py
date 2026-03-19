"""
迁移脚本: 为 users 表添加 location_url 和 store_photo 字段
"""
import sqlite3
import os

DB_PATH = os.path.join(os.path.dirname(__file__), 'cambodia_wholesale.db')

def migrate():
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()
    
    # 获取 users 表现有列
    cursor.execute("PRAGMA table_info(users)")
    columns = [col[1] for col in cursor.fetchall()]
    
    added = []
    
    if 'location_url' not in columns:
        cursor.execute("ALTER TABLE users ADD COLUMN location_url TEXT DEFAULT NULL")
        added.append('location_url')
    
    if 'store_photo' not in columns:
        cursor.execute("ALTER TABLE users ADD COLUMN store_photo TEXT DEFAULT NULL")
        added.append('store_photo')
    
    conn.commit()
    conn.close()
    
    if added:
        print(f"✅ 已添加字段: {', '.join(added)}")
    else:
        print("✅ 字段已存在，无需迁移")

if __name__ == '__main__':
    migrate()
