import paramiko
import os

host='43.134.13.229'; user='ubuntu'; password='qer1235A@'
binary = r'backend-go\wholesale'
remote = '/opt/wholesale/backend/wholesale'

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect(host, username=user, password=password)

sftp = ssh.open_sftp()
sftp.put(binary, remote + '.new')
sftp.close()

for cmd in [
    f'mv {remote}.new {remote}',
    f'chmod +x {remote}',
    'sudo systemctl restart wholesale',
    'sleep 2',
    'sudo systemctl is-active wholesale',
]:
    _, o, e = ssh.exec_command(cmd)
    out = o.read().decode().strip()
    if out:
        print(out)

import time
time.sleep(2)

# Test responsiveness
_, o, _ = ssh.exec_command('curl -s -o /dev/null -w "%{http_code} in %{time_total}s" -X POST http://localhost:8000/api/auth/login -H "Content-Type: application/json" -d \'{"username":"100001","password":"test"}\' --max-time 5')
print('Login test:', o.read().decode())

ssh.close()
print('Deploy complete')
