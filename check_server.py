import paramiko

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('43.134.13.229', username='ubuntu', password='qer1235A@')

# Check which Merchants chunk is referenced
stdin, stdout, stderr = ssh.exec_command("grep -oP 'Merchants-[a-zA-Z0-9_-]+' /opt/wholesale/frontend/dist/assets/index-AI884Qon.js")
print('Merchants refs in index bundle:', stdout.read().decode().strip())

# Check index files
stdin, stdout, stderr = ssh.exec_command("ls -la /opt/wholesale/frontend/dist/assets/index-*")
print(stdout.read().decode())

# Check if delivery still in i18n
stdin, stdout, stderr = ssh.exec_command("grep -c 'delivery' /opt/wholesale/frontend/dist/assets/index-AI884Qon.js")
print('delivery occurrences in index:', stdout.read().decode().strip())

# Also check the Merchants JS content for 'delivery'
stdin, stdout, stderr = ssh.exec_command("grep -c 'delivery' /opt/wholesale/frontend/dist/assets/Merchants-DOq9Q-bb.js")
print('delivery in new Merchants:', stdout.read().decode().strip())

stdin, stdout, stderr = ssh.exec_command("grep -c 'delivery' /opt/wholesale/frontend/dist/assets/Merchants-q-PATdLE.js")
print('delivery in old Merchants:', stdout.read().decode().strip())

ssh.close()
