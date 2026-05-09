import paramiko

host='43.134.13.229'; user='ubuntu'; password='qer1235A@'
ssh=paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect(host, username=user, password=password)

_,o,e = ssh.exec_command('sudo systemctl restart wholesale && sleep 2 && sudo systemctl is-active wholesale')
print('Restart result:', o.read().decode())

import time
time.sleep(2)

# Test if backend is now responsive
_,o,e = ssh.exec_command('curl -s -o /dev/null -w "%{http_code} in %{time_total}s" -X POST http://localhost:8000/api/auth/login -H "Content-Type: application/json" -d \'{"username":"100001","password":"test"}\' --max-time 5 2>&1')
print('Login response:', o.read().decode())

ssh.close()
