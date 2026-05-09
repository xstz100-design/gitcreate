import paramiko
import os

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('43.134.13.229', username='ubuntu', password='qer1235A@', timeout=60)

def run(cmd):
    _, stdout, stderr = ssh.exec_command(cmd)
    out = stdout.read().decode()
    err = stderr.read().decode()
    print(f'$ {cmd[:80]}')
    if out: print(out[:300])
    if err and 'warning' not in err.lower(): print('[err]', err[:200])
    return out

# 确保目标目录存在
run('sudo mkdir -p /opt/wholesale/frontend/dist/assets')
run('sudo chown -R ubuntu:ubuntu /opt/wholesale/frontend/')

dist_dir = r'E:\Program_ayang\vue\frontend\dist'
remote_base = '/opt/wholesale/frontend/dist'

sftp = ssh.open_sftp()

def upload_dir(local_path, remote_path):
    os.makedirs(local_path, exist_ok=False) if False else None
    entries = os.listdir(local_path)
    for entry in entries:
        local_entry = os.path.join(local_path, entry)
        remote_entry = remote_path + '/' + entry
        if os.path.isdir(local_entry):
            try:
                sftp.mkdir(remote_entry)
            except:
                pass
            upload_dir(local_entry, remote_entry)
        else:
            sftp.put(local_entry, remote_entry)

print('开始上传 dist ...')
# 先清空旧文件
run(f'rm -rf {remote_base}/* 2>/dev/null; true')

upload_dir(dist_dir, remote_base)
print('上传完成！')

# 验证
run(f'ls -la {remote_base}/')
run(f'ls {remote_base}/assets/ | wc -l')

# reload nginx
run('sudo nginx -t && sudo nginx -s reload')
print('\n=== nginx 已 reload')

# 测试首页
run('curl -s -o /dev/null -w "%{http_code}" https://khmerai.cn/')

sftp.close()
ssh.close()
print('\n=== 部署完成')
