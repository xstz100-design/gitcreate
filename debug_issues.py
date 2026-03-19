import paramiko

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('43.134.13.229', username='ubuntu', password='qer1235A@')

# Check DB values
sftp = ssh.open_sftp()
script = 'import sqlite3\n'
script += 'conn = sqlite3.connect("/opt/wholesale/backend/cambodia_wholesale.db")\n'
script += 'c = conn.cursor()\n'
script += 'c.execute("SELECT DISTINCT payment_status, COUNT(id) FROM orders GROUP BY payment_status")\n'
script += 'for row in c.fetchall():\n'
script += '    print(f"  {row[0]}: {row[1]} orders")\n'
script += 'conn.close()\n'

with sftp.open('/tmp/check_db.py', 'w') as f:
    f.write(script)
sftp.close()

stdin, stdout, stderr = ssh.exec_command('/opt/wholesale/venv/bin/python3 /tmp/check_db.py')
print('DB payment_status values:')
print(stdout.read().decode())
err = stderr.read().decode()
if err:
    print('ERROR:', err)

# Check 100001 user info
sftp = ssh.open_sftp()
script2 = 'import sqlite3\n'
script2 += 'conn = sqlite3.connect("/opt/wholesale/backend/cambodia_wholesale.db")\n'
script2 += 'c = conn.cursor()\n'
script2 += 'c.execute("SELECT id, username, full_name, role, phone, address, is_active FROM users WHERE username=\'100001\'")\n'
script2 += 'row = c.fetchone()\n'
script2 += 'print(f"User: id={row[0]}, username={row[1]}, name={row[2]}, role={row[3]}, phone={row[4]}, addr={row[5]}, active={row[6]}")\n'
script2 += 'conn.close()\n'

with sftp.open('/tmp/check_user.py', 'w') as f:
    f.write(script2)
sftp.close()

stdin, stdout, stderr = ssh.exec_command('/opt/wholesale/venv/bin/python3 /tmp/check_user.py')
print('100001 user info:')
print(stdout.read().decode())
err = stderr.read().decode()
if err:
    print('ERROR:', err)

ssh.close()
