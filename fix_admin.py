import paramiko

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('43.134.13.229', username='ubuntu', password='qer1235A@')

sftp = ssh.open_sftp()

# Write a temp script on the server
import tempfile
script = (
    "import sqlite3\n"
    "conn = sqlite3.connect('/opt/wholesale/backend/cambodia_wholesale.db')\n"
    "c = conn.cursor()\n"
    "c.execute(\"UPDATE users SET is_active = 1 WHERE username = '100001'\")\n"
    "print('Updated', c.rowcount, 'row(s)')\n"
    "conn.commit()\n"
    "conn.close()\n"
)

with sftp.open('/tmp/fix_admin.py', 'w') as f:
    f.write(script)

stdin, stdout, stderr = ssh.exec_command('/opt/wholesale/venv/bin/python3 /tmp/fix_admin.py')
print(stdout.read().decode())
print(stderr.read().decode())
ssh.close()
print('Done!')
