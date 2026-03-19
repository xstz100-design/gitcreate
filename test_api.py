import paramiko
import json

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('43.134.13.229', username='ubuntu', password='qer1235A@')

# Login first to get token (OAuth2 form format)
stdin, stdout, stderr = ssh.exec_command(
    'curl -s -X POST http://127.0.0.1:8000/api/auth/login '
    '-d "username=100001&password=Abc123456"'
)
login_resp = stdout.read().decode()
try:
    token = json.loads(login_resp)['access_token']
    print('Got token OK')
except:
    print('Login resp:', login_resp[:300])
    ssh.close()
    exit()

# Test filter: no filter
stdin, stdout, stderr = ssh.exec_command(
    f'curl -s "http://127.0.0.1:8000/api/orders" '
    f'-H "Authorization: Bearer {token}"'
)
data = json.loads(stdout.read().decode())
print(f'No filter: {len(data)} orders')

# Test filter: cash
stdin, stdout, stderr = ssh.exec_command(
    f'curl -s "http://127.0.0.1:8000/api/orders?payment_status=cash" '
    f'-H "Authorization: Bearer {token}"'
)
resp = stdout.read().decode()
try:
    data = json.loads(resp)
    print(f'cash filter: {len(data)} orders')
except:
    print(f'cash filter error: {resp[:300]}')

# Test filter: unpaid
stdin, stdout, stderr = ssh.exec_command(
    f'curl -s "http://127.0.0.1:8000/api/orders?payment_status=unpaid" '
    f'-H "Authorization: Bearer {token}"'
)
resp = stdout.read().decode()
try:
    data = json.loads(resp)
    print(f'unpaid filter: {len(data)} orders')
except:
    print(f'unpaid filter error: {resp[:300]}')

# Test filter: monthly
stdin, stdout, stderr = ssh.exec_command(
    f'curl -s "http://127.0.0.1:8000/api/orders?payment_status=monthly" '
    f'-H "Authorization: Bearer {token}"'
)
resp = stdout.read().decode()
try:
    data = json.loads(resp)
    print(f'monthly filter: {len(data)} orders')
except:
    print(f'monthly filter error: {resp[:300]}')

# Test update user 100001 (id=5)
stdin, stdout, stderr = ssh.exec_command(
    f'curl -s -X PATCH "http://127.0.0.1:8000/api/auth/users/5" '
    f'-H "Authorization: Bearer {token}" '
    f'-H "Content-Type: application/json" '
    f'-d \'{{"full_name":"TestAdmin123"}}\''
)
resp = stdout.read().decode()
print(f'Update 100001: {resp[:300]}')

# Revert
stdin, stdout, stderr = ssh.exec_command(
    f'curl -s -X PATCH "http://127.0.0.1:8000/api/auth/users/5" '
    f'-H "Authorization: Bearer {token}" '
    f'-H "Content-Type: application/json" '
    f'-d \'{{"full_name":"xuyang wang"}}\''
)
resp = stdout.read().decode()
print(f'Revert: {resp[:200]}')

ssh.close()
print('Done!')
