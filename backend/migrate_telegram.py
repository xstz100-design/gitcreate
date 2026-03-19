import sqlite3
conn = sqlite3.connect('cambodia_wholesale.db')
c = conn.cursor()
c.execute('PRAGMA table_info(users)')
cols = [col[1] for col in c.fetchall()]
print('Existing columns:', cols)
if 'telegram_bot_token' not in cols:
    c.execute('ALTER TABLE users ADD COLUMN telegram_bot_token VARCHAR(200)')
    print('Added telegram_bot_token')
if 'telegram_chat_id' not in cols:
    c.execute('ALTER TABLE users ADD COLUMN telegram_chat_id VARCHAR(100)')
    print('Added telegram_chat_id')
conn.commit()
conn.close()
print('Done')
