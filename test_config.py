import paramiko

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('43.134.13.229', username='ubuntu', password='qer1235A@', timeout=30)

# Check .env file
print('=== .env file ===')
stdin, stdout, stderr = ssh.exec_command('cat /opt/wholesale/backend/.env 2>/dev/null || echo "NO .env FILE"')
env_content = stdout.read().decode()
# Mask sensitive values before printing
lines = env_content.splitlines()
masked = []
for line in lines:
    if 'TOKEN' in line.upper() or 'SECRET' in line.upper() or 'PASSWORD' in line.upper():
        key = line.split('=')[0] if '=' in line else line
        val = line.split('=', 1)[1] if '=' in line else ''
        masked.append(f'{key}={"*" * min(len(val), 8)}... (length={len(val)})')
    else:
        masked.append(line)
print('\n'.join(masked))

# Check systemd service env vars
print('\n=== systemd service file ===')
stdin2, stdout2, stderr2 = ssh.exec_command('cat /etc/systemd/system/wholesale.service')
print(stdout2.read().decode())

# Check if TG_BOT_TOKEN is set
print('\n=== Check TG_BOT_TOKEN is set ===')
stdin3, stdout3, stderr3 = ssh.exec_command('sudo -u ubuntu sh -c "cat /opt/wholesale/backend/.env | grep TG_BOT_TOKEN | wc -c"')
result = stdout3.read().decode().strip()
print(f'TG_BOT_TOKEN line length: {result} chars (0 = not found)')

# Send a test Telegram message directly using the bot token from .env
print('\n=== Test Telegram API connectivity ===')
stdin4, stdout4, stderr4 = ssh.exec_command(
    'source /opt/wholesale/backend/.env 2>/dev/null; '
    'TOKEN=$(grep TG_BOT_TOKEN /opt/wholesale/backend/.env | cut -d= -f2); '
    'echo "Token length: ${#TOKEN}"; '
    'if [ -n "$TOKEN" ]; then '
    '  curl -s "https://api.telegram.org/bot${TOKEN}/getMe" | head -c 200; '
    'else echo "NO TOKEN"; fi'
)
print(stdout4.read().decode())

ssh.close()
