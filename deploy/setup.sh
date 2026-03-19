#!/bin/bash
set -e

echo "=== Cambodia Wholesale 部署脚本 ==="

# 1. 安装系统依赖
echo "[1/7] 安装系统依赖..."
apt update -y
apt install -y python3 python3-venv python3-pip nginx

# 2. 创建项目目录
echo "[2/7] 创建项目目录..."
mkdir -p /opt/wholesale/backend
mkdir -p /opt/wholesale/frontend/dist
mkdir -p /opt/wholesale/backend/uploads

# 3. 创建 Python 虚拟环境
echo "[3/7] 创建虚拟环境..."
if [ ! -d "/opt/wholesale/venv" ]; then
    python3 -m venv /opt/wholesale/venv
fi
source /opt/wholesale/venv/bin/activate

# 4. 安装 Python 依赖
echo "[4/7] 安装 Python 依赖..."
pip install --upgrade pip
pip install -r /opt/wholesale/backend/requirements.txt

# 5. 初始化数据库
echo "[5/7] 初始化数据库..."
cd /opt/wholesale/backend
python init_db.py

# 6. 配置 Nginx
echo "[6/7] 配置 Nginx..."
cp /opt/wholesale/deploy/nginx.conf /etc/nginx/sites-available/wholesale
ln -sf /etc/nginx/sites-available/wholesale /etc/nginx/sites-enabled/wholesale
rm -f /etc/nginx/sites-enabled/default
nginx -t
systemctl restart nginx
systemctl enable nginx

# 7. 配置 systemd 服务
echo "[7/7] 配置后端服务..."
cp /opt/wholesale/deploy/wholesale.service /etc/systemd/system/wholesale.service
systemctl daemon-reload
systemctl restart wholesale
systemctl enable wholesale

echo ""
echo "=== 部署完成 ==="
echo "前端: http://43.134.13.229"
echo "后端: http://43.134.13.229/api"
echo ""
echo "查看后端日志: journalctl -u wholesale -f"
echo "默认管理员: admin / admin123"
