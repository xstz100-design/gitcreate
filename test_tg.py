import paramiko

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('43.134.13.229', username='ubuntu', password='qer1235A@', timeout=30)

# Get bot token from .env
stdin, stdout, stderr = ssh.exec_command("grep TG_BOT_TOKEN /opt/wholesale/backend/.env | cut -d= -f2")
token = stdout.read().decode().strip()

# Try sending a message to user 酒小二 (telegram_id=7595498982)
chat_id = "7595498982"
cmd = f"curl -s 'https://api.telegram.org/bot{token}/sendMessage' -d 'chat_id={chat_id}&text=测试消息：请忽略'"
stdin2, stdout2, stderr2 = ssh.exec_command(cmd)
result = stdout2.read().decode()
print(f'Send message to 酒小二 (7595498982): {result}')

# Try sending to Junka (8298563556)
chat_id2 = "8298563556"
cmd2 = f"curl -s 'https://api.telegram.org/bot{token}/sendMessage' -d 'chat_id={chat_id2}&text=测试消息：请忽略'"
stdin3, stdout3, stderr3 = ssh.exec_command(cmd2)
result2 = stdout3.read().decode()
print(f'\nSend message to Junka (8298563556): {result2}')

# Check recent bot updates to see if users have interacted
cmd3 = f"curl -s 'https://api.telegram.org/bot{token}/getUpdates?limit=5'"
stdin4, stdout4, stderr4 = ssh.exec_command(cmd3)
result3 = stdout4.read().decode()
print(f'\nRecent bot updates: {result3[:500]}')

ssh.close()
