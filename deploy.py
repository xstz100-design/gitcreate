"""
Cambodia Wholesale 自动部署脚本
使用 paramiko 通过 SSH 上传文件并配置服务器
"""
import os
import sys
import stat
import paramiko
from scp import SCPClient
import tarfile
import tempfile

# 服务器配置
SERVER = "43.134.13.229"
USER = "root"
PASSWORD = "qer1235A@"
REMOTE_DIR = "/opt/wholesale"

# 本地路径
BASE_DIR = os.path.dirname(os.path.abspath(__file__))
BACKEND_DIR = os.path.join(BASE_DIR, "backend")
FRONTEND_DIST = os.path.join(BASE_DIR, "frontend", "dist")
DEPLOY_DIR = os.path.join(BASE_DIR, "deploy")


def create_archive():
    """打包项目文件为 tar.gz"""
    archive_path = os.path.join(tempfile.gettempdir(), "wholesale_deploy.tar.gz")
    print(f"[1/5] 打包项目文件 -> {archive_path}")

    with tarfile.open(archive_path, "w:gz") as tar:
        # 添加后端文件（排除 __pycache__, .db, uploads 中的图片）
        for root, dirs, files in os.walk(BACKEND_DIR):
            # 排除目录
            dirs[:] = [d for d in dirs if d not in ("__pycache__", ".venv", "venv", "uploads")]
            for f in files:
                if f.endswith((".pyc", ".db")):
                    continue
                full_path = os.path.join(root, f)
                arcname = os.path.join("backend", os.path.relpath(full_path, BACKEND_DIR))
                tar.add(full_path, arcname=arcname)
                
        # 添加前端 dist
        for root, dirs, files in os.walk(FRONTEND_DIST):
            for f in files:
                full_path = os.path.join(root, f)
                arcname = os.path.join("frontend", "dist", os.path.relpath(full_path, FRONTEND_DIST))
                tar.add(full_path, arcname=arcname)

        # 添加部署配置文件
        for f in os.listdir(DEPLOY_DIR):
            full_path = os.path.join(DEPLOY_DIR, f)
            if os.path.isfile(full_path):
                tar.add(full_path, arcname=os.path.join("deploy", f))

    size_mb = os.path.getsize(archive_path) / (1024 * 1024)
    print(f"   打包完成: {size_mb:.1f} MB")
    return archive_path


def ssh_connect():
    """建立 SSH 连接"""
    print(f"[2/5] 连接服务器 {SERVER}...")
    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    client.connect(SERVER, username=USER, password=PASSWORD, timeout=30)
    print("   连接成功")
    return client


def upload_files(ssh_client, archive_path):
    """上传文件到服务器"""
    print("[3/5] 上传文件到服务器...")
    
    # 先创建目录
    run_cmd(ssh_client, f"mkdir -p {REMOTE_DIR}")
    
    # 上传压缩包
    with SCPClient(ssh_client.get_transport(), progress=progress) as scp:
        scp.put(archive_path, f"{REMOTE_DIR}/deploy.tar.gz")
    
    print("\n   上传完成，解压中...")
    
    # 解压
    run_cmd(ssh_client, f"cd {REMOTE_DIR} && tar xzf deploy.tar.gz && rm deploy.tar.gz")
    print("   解压完成")


def progress(filename, size, sent):
    """SCP 进度回调"""
    pct = sent / size * 100
    bar = "█" * int(pct // 5) + "░" * (20 - int(pct // 5))
    sys.stdout.write(f"\r   [{bar}] {pct:.0f}% ({sent//1024//1024}MB/{size//1024//1024}MB)")
    sys.stdout.flush()


def setup_server(ssh_client):
    """在服务器上执行安装和配置"""
    print("[4/5] 配置服务器...")

    commands = [
        # 安装系统依赖
        ("安装系统依赖", "apt update -y && apt install -y python3 python3-venv python3-pip nginx"),
        
        # 创建 uploads 目录
        ("创建上传目录", f"mkdir -p {REMOTE_DIR}/backend/uploads"),
        
        # 创建虚拟环境
        ("创建 Python 虚拟环境", f"python3 -m venv {REMOTE_DIR}/venv || true"),
        
        # 安装 Python 依赖
        ("安装 Python 依赖", f"{REMOTE_DIR}/venv/bin/pip install --upgrade pip && {REMOTE_DIR}/venv/bin/pip install -r {REMOTE_DIR}/backend/requirements.txt"),
        
        # 初始化数据库
        ("初始化数据库", f"cd {REMOTE_DIR}/backend && {REMOTE_DIR}/venv/bin/python init_db.py"),
        
        # 配置 Nginx
        ("配置 Nginx", f"""
cp {REMOTE_DIR}/deploy/nginx.conf /etc/nginx/sites-available/wholesale
ln -sf /etc/nginx/sites-available/wholesale /etc/nginx/sites-enabled/wholesale
rm -f /etc/nginx/sites-enabled/default
nginx -t && systemctl restart nginx && systemctl enable nginx
"""),
        
        # 配置 systemd 服务
        ("配置后端服务", f"""
cp {REMOTE_DIR}/deploy/wholesale.service /etc/systemd/system/wholesale.service
systemctl daemon-reload
systemctl restart wholesale
systemctl enable wholesale
"""),
    ]

    for desc, cmd in commands:
        print(f"   {desc}...")
        stdout, stderr = run_cmd(ssh_client, cmd)
        if stderr and "warning" not in stderr.lower() and "deprecat" not in stderr.lower():
            # 只显示真正的错误（忽略警告和 deprecation）
            err_lines = [l for l in stderr.split('\n') if l.strip() and 'warning' not in l.lower()]
            if err_lines:
                print(f"   ⚠ {err_lines[-1][:100]}")


def verify_deployment(ssh_client):
    """验证部署结果"""
    print("[5/5] 验证部署...")
    
    # 检查服务状态
    stdout, _ = run_cmd(ssh_client, "systemctl is-active wholesale")
    backend_ok = "active" in stdout
    
    stdout, _ = run_cmd(ssh_client, "systemctl is-active nginx")
    nginx_ok = "active" in stdout
    
    # 检查端口
    stdout, _ = run_cmd(ssh_client, "ss -tlnp | grep -E ':80|:8000'")
    
    print(f"   Nginx:  {'✓ 运行中' if nginx_ok else '✗ 未运行'}")
    print(f"   后端:   {'✓ 运行中' if backend_ok else '✗ 未运行'}")
    print(f"   端口:   {stdout.strip()}")
    
    return backend_ok and nginx_ok


def run_cmd(ssh_client, cmd):
    """执行远程命令"""
    stdin, stdout, stderr = ssh_client.exec_command(cmd, timeout=300)
    out = stdout.read().decode("utf-8", errors="replace")
    err = stderr.read().decode("utf-8", errors="replace")
    return out, err


def main():
    print("=" * 50)
    print("Cambodia Wholesale 自动部署")
    print("=" * 50)
    print()

    # 1. 打包
    archive_path = create_archive()

    # 2. 连接
    ssh_client = ssh_connect()

    try:
        # 3. 上传
        upload_files(ssh_client, archive_path)

        # 4. 配置
        setup_server(ssh_client)

        # 5. 验证
        ok = verify_deployment(ssh_client)

        print()
        print("=" * 50)
        if ok:
            print("✓ 部署成功!")
            print(f"  访问地址: http://{SERVER}")
            print(f"  管理员: admin / admin123")
        else:
            print("⚠ 部署可能有问题，请检查服务器日志:")
            print(f"  journalctl -u wholesale -f")
        print("=" * 50)
    finally:
        ssh_client.close()
        # 清理临时文件
        if os.path.exists(archive_path):
            os.remove(archive_path)


if __name__ == "__main__":
    main()
