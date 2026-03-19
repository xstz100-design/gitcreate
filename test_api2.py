import paramiko

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('43.134.13.229', username='ubuntu', password='qer1235A@')

# Reset 100001 password to known value
sftp = ssh.open_sftp()
script = """
import sys
sys.path.insert(0, '/opt/wholesale/backend')
from app.core.security import get_password_hash
import sqlite3
conn = sqlite3.connect('/opt/wholesale/backend/cambodia_wholesale.db')
c = conn.cursor()
new_hash = get_password_hash('admin888')
c.execute("UPDATE users SET hashed_password = ?, must_change_password = 0 WHERE username = '100001'", (new_hash,))
print(f'Updated {c.rowcount} row(s)')
conn.commit()
conn.close()
"""
with sftp.open('/tmp/reset_pwd.py', 'w') as f:
    f.write(script)
sftp.close()

stdin, stdout, stderr = ssh.exec_command('/opt/wholesale/venv/bin/python3 /tmp/reset_pwd.py')
print(stdout.read().decode())
print(stderr.read().decode())

# Now test login
stdin, stdout, stderr = ssh.exec_command(
    'curl -s -X POST http://127.0.0.1:8000/api/auth/login '
    '-d "username=100001&password=admin888"'
)
resp = stdout.read().decode()
print('Login:', resp[:200])

import json
try:
    token = json.loads(resp)['access_token']
except:
    print('Cannot get token, exiting')
    ssh.close()
    exit()

# Test filters
for status in ['', 'cash', 'unpaid', 'monthly']:
    qs = f'?payment_status={status}' if status else ''
    stdin, stdout, stderr = ssh.exec_command(
        f'curl -s "http://127.0.0.1:8000/api/orders{qs}" '
        f'-H "Authorization: Bearer {token}"'
    )
    r = stdout.read().decode()
    try:
        data = json.loads(r)
        label = status if status else 'ALL'
        print(f'{label}: {len(data)} orders')
    except:
        print(f'{status} error: {r[:200]}')

# Test user update
stdin, stdout, stderr = ssh.exec_command(
    f'curl -s -X PATCH "http://127.0.0.1:8000/api/auth/users/5" '
    f'-H "Authorization: Bearer {token}" '
    f'-H "Content-Type: application/json" '
    f"""-d '{{"full_name":"Test Admin"}}'"""
)
print('Update test:', stdout.read().decode()[:300])

# Revert
stdin, stdout, stderr = ssh.exec_command(
    f'curl -s -X PATCH "http://127.0.0.1:8000/api/auth/users/5" '
    f'-H "Authorization: Bearer {token}" '
    f'-H "Content-Type: application/json" '
    f"""-d '{{"full_name":"xuyang wang"}}'"""
)
print('Revert:', stdout.read().decode()[:200])

ssh.close()
print('Done!')
