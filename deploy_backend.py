import paramiko
import os

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('43.134.13.229', username='ubuntu', password='qer1235A@', timeout=60)

def run(cmd):
    _, stdout, stderr = ssh.exec_command(cmd)
    out = stdout.read().decode()
    err = stderr.read().decode()
    print(f'$ {cmd[:100]}')
    if out: print(out[:500])
    if err and 'warning' not in err.lower(): print('[err]', err[:300])
    return out

local_binary = r'E:\Program_ayang\vue\wholesale'
remote_tmp = '/opt/wholesale/backend/wholesale_new'
remote_bin = '/opt/wholesale/backend/wholesale'

sftp = ssh.open_sftp()

print('开始上传后端二进制文件...')
sftp.put(local_binary, remote_tmp)
print('上传完成！')

sftp.close()

run(f'sudo chmod +x {remote_tmp}')
run('sudo systemctl stop wholesale')
run(f'sudo mv {remote_tmp} {remote_bin}')
run(f'sudo chmod +x {remote_bin}')
run('sudo systemctl start wholesale')
import time; time.sleep(2)
run('sudo systemctl status wholesale --no-pager -l')

ssh.close()
print('\n=== 后端部署完成')
