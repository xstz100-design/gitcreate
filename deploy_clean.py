import paramiko
import os

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('43.134.13.229', username='ubuntu', password='qer1235A@')

# 1. Disable conflicting default site
print('=== Step 1: Disable default site ===')
stdin, stdout, stderr = ssh.exec_command('sudo rm -f /etc/nginx/sites-enabled/default')
stdout.read()
print('Removed default site')

# 2. Update wholesale nginx config with aggressive no-cache for HTML
print('\n=== Step 2: Update nginx config ===')
wholesale_conf = """server {
    listen 80 default_server;
    server_name _;

    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml text/javascript image/svg+xml;

    root /opt/wholesale/frontend/dist;
    index index.html;

    # SPA fallback - no cache for HTML
    location / {
        try_files $uri $uri/ /index.html;
        add_header Cache-Control "no-cache, no-store, must-revalidate";
        add_header Pragma "no-cache";
        add_header Expires "0";
    }

    # API proxy
    location /api/ {
        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_read_timeout 300;
        proxy_buffering on;
        proxy_buffer_size 8k;
        proxy_buffers 4 16k;
    }

    # Static uploads
    location /uploads/ {
        alias /opt/wholesale/backend/uploads/;
        expires 30d;
        add_header Cache-Control "public, immutable";
    }

    # Hashed assets - long cache
    location /assets/ {
        expires 1y;
        add_header Cache-Control "public, immutable";
        access_log off;
    }

    # Images
    location /images/ {
        expires 30d;
        add_header Cache-Control "public";
    }

    client_max_body_size 20M;
}
"""
sftp = ssh.open_sftp()
with sftp.open('/tmp/wholesale.conf', 'w') as f:
    f.write(wholesale_conf)

stdin, stdout, stderr = ssh.exec_command('sudo cp /tmp/wholesale.conf /etc/nginx/sites-available/wholesale')
stdout.read()
print('Updated wholesale config')

# 3. Clean old dist completely
print('\n=== Step 3: Clean deploy frontend ===')
stdin, stdout, stderr = ssh.exec_command('sudo rm -rf /opt/wholesale/frontend/dist && sudo mkdir -p /opt/wholesale/frontend/dist && sudo chown -R ubuntu:ubuntu /opt/wholesale/frontend/dist')
stdout.read()
print('Cleaned dist directory')

# 4. Upload new dist
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
sftp.close()
print('Frontend uploaded')

# 5. Test and reload nginx
print('\n=== Step 4: Reload nginx ===')
stdin, stdout, stderr = ssh.exec_command('sudo nginx -t 2>&1')
test_result = stdout.read().decode()
print(test_result)

if 'successful' in test_result:
    stdin, stdout, stderr = ssh.exec_command('sudo systemctl reload nginx')
    stdout.read()
    print('Nginx reloaded!')
else:
    print('ERROR: nginx config test failed!')

# 6. Verify
print('\n=== Step 5: Verify ===')
stdin, stdout, stderr = ssh.exec_command('curl -sI http://127.0.0.1/ | head -10')
print('Headers:', stdout.read().decode())

stdin, stdout, stderr = ssh.exec_command('curl -s http://127.0.0.1/ | head -5')
print('Body:', stdout.read().decode())

stdin, stdout, stderr = ssh.exec_command('ls /opt/wholesale/frontend/dist/assets/ | wc -l')
print('Asset files:', stdout.read().decode().strip())

ssh.close()
print('\nAll done!')
