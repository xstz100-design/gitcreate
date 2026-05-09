import paramiko, json

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('43.134.13.229', username='ubuntu', password='qer1235A@', timeout=30)

# Test OTP request with phone number
cmd = "curl -s -X POST http://localhost:8000/api/auth/otp/request -H 'Content-Type: application/json' -d '{\"phone\":\"85832567\"}'"
stdin, stdout, stderr = ssh.exec_command(cmd)
resp = stdout.read().decode()
err = stderr.read().decode()
print('=== OTP request (phone=85832567) ===')
print('Response:', resp)
if err:
    print('Stderr:', err)

# Test with Junka's phone
cmd2 = "curl -s -X POST http://localhost:8000/api/auth/otp/request -H 'Content-Type: application/json' -d '{\"phone\":\"885236645\"}'"
stdin2, stdout2, stderr2 = ssh.exec_command(cmd2)
resp2 = stdout2.read().decode()
print('\n=== OTP request (phone=885236645 / Junka) ===')
print('Response:', resp2)

# Check backend logs
stdin3, stdout3, stderr3 = ssh.exec_command('sudo journalctl -u wholesale -n 50 --no-pager 2>&1')
logs = stdout3.read().decode()
print('\n=== Backend logs ===')
print(logs)

ssh.close()
