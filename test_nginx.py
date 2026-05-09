import paramiko

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('43.134.13.229', username='ubuntu', password='qer1235A@', timeout=30)

# Test through nginx (HTTPS)
print('=== Test via nginx (localhost:443 → backend) ===')
stdin, stdout, stderr = ssh.exec_command(
    "curl -s -k -X POST https://localhost/api/auth/otp/request "
    "-H 'Content-Type: application/json' "
    "-d '{\"phone\":\"85832567\"}'"
)
print('Response:', stdout.read().decode())

# Check nginx config for API proxy
print('\n=== nginx config (api proxy section) ===')
stdin2, stdout2, stderr2 = ssh.exec_command("grep -A 5 'api' /etc/nginx/nginx.conf | head -40")
print(stdout2.read().decode())

# Check what domains nginx listens on
print('\n=== nginx server_name ===')
stdin3, stdout3, stderr3 = ssh.exec_command("grep -E 'server_name|listen' /etc/nginx/nginx.conf")
print(stdout3.read().decode())

# Check if the service is accessible from outside
print('\n=== curl from server to itself via domain ===')
stdin4, stdout4, stderr4 = ssh.exec_command(
    "curl -s -k -X POST https://khmerai.cn/api/auth/otp/request "
    "-H 'Content-Type: application/json' "
    "-d '{\"phone\":\"85832567\"}' --max-time 10"
)
print('Via khmerai.cn:', stdout4.read().decode())

ssh.close()
