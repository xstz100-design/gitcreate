import paramiko

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('43.134.13.229', username='ubuntu', password='qer1235A@')

# Check which index.html is served
stdin, stdout, stderr = ssh.exec_command('cat /opt/wholesale/frontend/dist/index.html')
print('index.html:', stdout.read().decode())

# Check nginx config for cache settings
stdin, stdout, stderr = ssh.exec_command('cat /etc/nginx/sites-available/default')
nginx_conf = stdout.read().decode()
print('\n--- NGINX CONFIG ---')
print(nginx_conf)

# Check if location for frontend root correctly set
stdin, stdout, stderr = ssh.exec_command('curl -s -I http://127.0.0.1/index.html | head -15')
print('\n--- HTTP HEADERS for index.html ---')
print(stdout.read().decode())

# Check if the Merchants chunk is the latest
stdin, stdout, stderr = ssh.exec_command('ls -la /opt/wholesale/frontend/dist/assets/Merchants*')
print('\n--- Merchants files ---')
print(stdout.read().decode())

# Check delivery word in latest Merchants
stdin, stdout, stderr = ssh.exec_command("grep -c 'delivery' /opt/wholesale/frontend/dist/assets/Merchants-*.js 2>/dev/null")
print('delivery in Merchants:', stdout.read().decode())

# Check if the main bundle references correct Merchants
stdin, stdout, stderr = ssh.exec_command("grep -oP 'Merchants-[a-zA-Z0-9_-]+' /opt/wholesale/frontend/dist/assets/index-*.js | head -5")
print('Merchants refs:', stdout.read().decode())

ssh.close()
print('Done!')
