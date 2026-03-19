import paramiko
import os

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('43.134.13.229', username='ubuntu', password='qer1235A@')
sftp = ssh.open_sftp()

local_dist = r'c:\Users\Administrator\Desktop\vue\frontend\dist'
remote_dist = '/opt/wholesale/frontend/dist'

# First fix permissions
stdin, stdout, stderr = ssh.exec_command('sudo chown -R ubuntu:ubuntu /opt/wholesale/frontend/dist')
stdout.read()

def upload_dir(local_path, remote_path):
    """递归上传目录"""
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
print('Frontend uploaded successfully!')

sftp.close()
ssh.close()
