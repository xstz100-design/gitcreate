import paramiko
import os

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('43.134.13.229', username='ubuntu', password='qer1235A@')

# 1. Update nginx config to disable cache for index.html
nginx_patch = '''
# Check current nginx config for wholesale
grep -q "no-cache" /etc/nginx/sites-available/default
if [ $? -ne 0 ]; then
    # Add no-cache for index.html before the last closing brace
    sudo sed -i '/location \\/ {/,/}/{ s|try_files.*|try_files $uri $uri/ /index.html;\\n            add_header Cache-Control "no-cache, no-store, must-revalidate";\\n            add_header Pragma "no-cache";|; }' /etc/nginx/sites-available/default
fi
'''

# Read current nginx config
stdin, stdout, stderr = ssh.exec_command('cat /etc/nginx/sites-available/default')
nginx_conf = stdout.read().decode()
print('Current nginx config length:', len(nginx_conf))

# 2. Clear old dist and upload new
print('Clearing old dist...')
stdin, stdout, stderr = ssh.exec_command('sudo rm -rf /opt/wholesale/frontend/dist/*')
stdout.read()

# Fix ownership
stdin, stdout, stderr = ssh.exec_command('sudo chown -R ubuntu:ubuntu /opt/wholesale/frontend/dist')
stdout.read()

# Upload new dist
sftp = ssh.open_sftp()
local_dist = r'c:\Users\Administrator\Desktop\vue\frontend\dist'
remote_dist = '/opt/wholesale/frontend/dist'

def upload_dir(local_path, remote_path):
    try:
        sftp.stat(remote_path)
    except FileNotFoundError:
        sftp.mkdir(remote_path)
    for item in os.listdir(local_path):
        local_item = os.path.join(local_path, item)
        remote_item = f'{remote_path}/{item}'
        if os.path.isdir(local_item):
            upload_dir(local_item, remote_item)
        else:
            sftp.put(local_item, remote_item)

upload_dir(local_dist, remote_dist)
print('Frontend uploaded!')
sftp.close()

# 3. Add no-cache header for index.html in nginx
# Check if location block for html/root exists
if 'no-cache' not in nginx_conf:
    # Add cache control inside server block, before the last }
    new_location = '''
    # No cache for index.html
    location = /index.html {
        root /opt/wholesale/frontend/dist;
        add_header Cache-Control "no-cache, no-store, must-revalidate";
        add_header Pragma "no-cache";
        add_header Expires "0";
    }
'''
    # Find where to insert - before the last closing brace of the server block
    stdin, stdout, stderr = ssh.exec_command(
        f'echo \'{new_location}\' | sudo tee /tmp/nginx_patch.conf'
    )
    stdout.read()
    
    # Insert before the last } in the config
    stdin, stdout, stderr = ssh.exec_command(
        "sudo sed -i '/^}/i \\    # No cache for index.html\\n    location = /index.html {\\n        root /opt/wholesale/frontend/dist;\\n        add_header Cache-Control \"no-cache, no-store, must-revalidate\";\\n        add_header Pragma \"no-cache\";\\n        add_header Expires \"0\";\\n    }' /etc/nginx/sites-available/default"
    )
    out = stdout.read().decode()
    err = stderr.read().decode()
    if err:
        print('nginx patch error:', err)
    else:
        print('Added no-cache for index.html')

# Test and reload nginx
stdin, stdout, stderr = ssh.exec_command('sudo nginx -t 2>&1')
nginx_test = stdout.read().decode()
print('nginx test:', nginx_test)

if 'successful' in nginx_test or 'ok' in nginx_test:
    stdin, stdout, stderr = ssh.exec_command('sudo systemctl reload nginx')
    stdout.read()
    print('Nginx reloaded!')
else:
    print('Nginx config test failed, not reloading')
    # Revert
    stdin, stdout, stderr = ssh.exec_command(
        "sudo sed -i '/# No cache for index.html/,/^    }/d' /etc/nginx/sites-available/default"
    )
    stdout.read()
    print('Reverted nginx config')

# Verify
stdin, stdout, stderr = ssh.exec_command('ls /opt/wholesale/frontend/dist/assets/ | wc -l')
print('Assets count:', stdout.read().decode().strip())

stdin, stdout, stderr = ssh.exec_command('cat /opt/wholesale/frontend/dist/index.html')
print('index.html:', stdout.read().decode().strip())

ssh.close()
print('Done!')
