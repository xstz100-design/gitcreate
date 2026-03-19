import paramiko
ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('43.134.13.229', username='ubuntu', password='qer1235A@')

cmds = [
    ("API /auth/me", "curl -s -o /dev/null -w '%{http_code}' http://localhost/api/auth/me"),
    ("Frontend", "curl -s -o /dev/null -w '%{http_code}' http://localhost/"),
    ("Products API", "curl -s http://localhost/api/products | python3 -c \"import sys,json; d=json.load(sys.stdin); print(f'Count: {len(d)}'); print(json.dumps(d[0], indent=2)[:300]) if d else print('empty')\""),
]

for name, cmd in cmds:
    _, stdout, stderr = ssh.exec_command(cmd)
    print(f"{name}: {stdout.read().decode().strip()}")

ssh.close()
