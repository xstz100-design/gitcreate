#!/bin/bash
set -e

# ===================================================
# 柬埔寨批发管理系统 - 宝塔面板部署脚本
# 适用于: CentOS 7/8, Ubuntu 18/20/22, Debian 10/11
# 注意: 请以 root 用户运行本脚本
# ===================================================

echo "=============================================="
echo "  柬埔寨批发管理系统 - 宝塔面板部署"
echo "=============================================="

# ---- 颜色输出 ----
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# ---- 配置 ----
PROJECT_DIR="/opt/wholesale"
BACKEND_DIR="${PROJECT_DIR}/backend"
FRONTEND_DIR="${PROJECT_DIR}/frontend"
VENV_DIR="${PROJECT_DIR}/venv"
BT_WWWROOT="/www/wwwroot"
SITE_DIR="${BT_WWWROOT}/wholesale"
SERVER_IP=$(curl -s ifconfig.me 2>/dev/null || echo "YOUR_SERVER_IP")

# ---- 检查 root 权限 ----
if [ "$(id -u)" -ne 0 ]; then
    echo -e "${RED}错误: 请使用 root 用户运行此脚本${NC}"
    exit 1
fi

# ===== 第一步: 检查/安装宝塔面板 =====
echo ""
echo -e "${GREEN}[1/8] 检查宝塔面板...${NC}"

if [ -f "/www/server/panel/BT-Panel" ] || command -v bt &> /dev/null; then
    echo "宝塔面板已安装"
    bt default 2>/dev/null || true
else
    echo "正在安装宝塔面板..."
    # 检测操作系统
    if [ -f /etc/redhat-release ]; then
        echo "检测到 CentOS/RHEL 系统"
        yum install -y wget
        wget -O install.sh https://download.bt.cn/install/install_6.0.sh && bash install.sh ed8484bec
    elif [ -f /etc/debian_version ]; then
        echo "检测到 Debian/Ubuntu 系统"
        apt-get update -y
        apt-get install -y wget
        wget -O install.sh https://download.bt.cn/install/install-ubuntu_6.0.sh && bash install.sh ed8484bec
    else
        echo -e "${RED}不支持的操作系统，请参考宝塔官网手动安装: https://www.bt.cn/new/download.html${NC}"
        exit 1
    fi
    echo -e "${GREEN}宝塔面板安装完成！请记录上方的面板地址和账号密码${NC}"
fi

# ===== 第二步: 通过宝塔安装必要软件 =====
echo ""
echo -e "${GREEN}[2/8] 检查必要软件...${NC}"

# 检查 Nginx
if [ ! -f "/www/server/nginx/sbin/nginx" ]; then
    echo -e "${YELLOW}请在宝塔面板中安装 Nginx (推荐 1.24+)${NC}"
    echo "面板地址: http://${SERVER_IP}:8888"
    echo "进入 软件商店 → Nginx → 安装"
    echo ""
    echo "安装完成后，重新运行此脚本"
    echo "或使用命令行安装: bt 16 (选择Nginx)"
    exit 1
else
    echo "Nginx 已安装 ✓"
fi

# 检查 Python
if ! command -v python3 &> /dev/null; then
    echo "安装 Python3..."
    if [ -f /etc/redhat-release ]; then
        yum install -y python3 python3-devel python3-pip
    else
        apt-get install -y python3 python3-venv python3-pip python3-dev
    fi
fi
echo "Python3 已就绪 ✓"

# 安装 Node.js (用于构建前端)
if ! command -v node &> /dev/null; then
    echo "安装 Node.js..."
    if [ -f /etc/redhat-release ]; then
        curl -fsSL https://rpm.nodesource.com/setup_18.x | bash -
        yum install -y nodejs
    else
        curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
        apt-get install -y nodejs
    fi
fi
echo "Node.js $(node -v) 已就绪 ✓"

# ===== 第三步: 创建项目目录 =====
echo ""
echo -e "${GREEN}[3/8] 创建项目目录...${NC}"
mkdir -p "${PROJECT_DIR}"
mkdir -p "${BACKEND_DIR}/uploads/thumbnails"
mkdir -p "${FRONTEND_DIR}"
mkdir -p "${SITE_DIR}"
mkdir -p "${PROJECT_DIR}/backups"
mkdir -p "${PROJECT_DIR}/logs"

# ===== 第四步: 部署后端 =====
echo ""
echo -e "${GREEN}[4/8] 部署后端...${NC}"

# 复制后端代码 (如果从本地运行)
SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
if [ -d "${SCRIPT_DIR}/backend" ]; then
    echo "从源码目录复制后端..."
    rsync -av --exclude='__pycache__' --exclude='*.pyc' --exclude='venv' \
        "${SCRIPT_DIR}/backend/" "${BACKEND_DIR}/"
fi

# 创建虚拟环境
if [ ! -d "${VENV_DIR}" ]; then
    echo "创建 Python 虚拟环境..."
    python3 -m venv "${VENV_DIR}"
fi
source "${VENV_DIR}/bin/activate"

# 安装依赖
echo "安装 Python 依赖..."
pip install --upgrade pip
pip install -r "${BACKEND_DIR}/requirements.txt"

# 初始化数据库
echo "初始化数据库..."
cd "${BACKEND_DIR}"
python init_db.py

deactivate

# ===== 第五步: 构建前端 =====
echo ""
echo -e "${GREEN}[5/8] 构建前端...${NC}"

if [ -d "${SCRIPT_DIR}/frontend" ]; then
    echo "从源码目录复制前端..."
    rsync -av --exclude='node_modules' --exclude='dist' \
        "${SCRIPT_DIR}/frontend/" "${FRONTEND_DIR}/"
fi

cd "${FRONTEND_DIR}"
if [ -f "package.json" ]; then
    echo "安装前端依赖..."
    npm install
    echo "构建前端..."
    npm run build
    # 将构建产物复制到宝塔站点目录
    rsync -av "${FRONTEND_DIR}/dist/" "${SITE_DIR}/"
    echo "前端构建完成 ✓"
else
    echo -e "${YELLOW}未找到 package.json，请手动构建前端并上传到 ${SITE_DIR}${NC}"
fi

# ===== 第六步: 配置宝塔 Nginx 站点 =====
echo ""
echo -e "${GREEN}[6/8] 配置 Nginx 站点...${NC}"

BT_NGINX_VHOST="/www/server/panel/vhost/nginx"
BT_NGINX_CONF="${BT_NGINX_VHOST}/wholesale.conf"
mkdir -p "${BT_NGINX_VHOST}"

cat > "${BT_NGINX_CONF}" << NGINX_EOF
server {
    listen 80;
    server_name ${SERVER_IP};

    # ---------- 日志 ----------
    access_log /www/wwwlogs/wholesale.log;
    error_log  /www/wwwlogs/wholesale.error.log;

    # ---------- 前端静态文件 ----------
    root ${SITE_DIR};
    index index.html;

    # Vue Router history 模式
    location / {
        try_files \$uri \$uri/ /index.html;
    }

    # ---------- API 反向代理 ----------
    location /api/ {
        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_read_timeout 300;
        proxy_connect_timeout 60;
        proxy_send_timeout 300;
    }

    # ---------- 上传文件 ----------
    location /uploads/ {
        alias ${BACKEND_DIR}/uploads/;
        expires 30d;
        add_header Cache-Control "public, immutable";
    }

    # ---------- 静态资源缓存 ----------
    location /assets/ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # ---------- 安全头 ----------
    add_header X-Content-Type-Options nosniff;
    add_header X-Frame-Options SAMEORIGIN;
    add_header X-XSS-Protection "1; mode=block";

    # ---------- 上传大小限制 ----------
    client_max_body_size 20M;

    # ---------- 禁止访问隐藏文件 ----------
    location ~ /\. {
        deny all;
        access_log off;
        log_not_found off;
    }
}
NGINX_EOF

echo "Nginx 配置已写入: ${BT_NGINX_CONF}"

# 测试 Nginx 配置
/www/server/nginx/sbin/nginx -t
if [ $? -eq 0 ]; then
    /www/server/nginx/sbin/nginx -s reload
    echo "Nginx 重载成功 ✓"
else
    echo -e "${RED}Nginx 配置有误，请检查${NC}"
fi

# ===== 第七步: 配置 Supervisor 进程守护 =====
echo ""
echo -e "${GREEN}[7/8] 配置 Supervisor 进程守护...${NC}"

# 检查宝塔 Supervisor 插件
BT_SUPERVISOR_DIR="/www/server/panel/plugin/supervisor"
if [ -d "${BT_SUPERVISOR_DIR}" ]; then
    echo "宝塔 Supervisor 插件已安装 ✓"
else
    echo -e "${YELLOW}提示: 建议在宝塔面板 → 软件商店 → 安装 Supervisor管理器${NC}"
    echo "安装后可通过面板图形化管理后端进程"
fi

# 创建 Supervisor 配置文件
SUPERVISOR_CONF_DIR="/www/server/panel/plugin/supervisor/config"
if [ -d "${SUPERVISOR_CONF_DIR}" ]; then
    cat > "${SUPERVISOR_CONF_DIR}/wholesale.conf" << SUP_EOF
[program:wholesale]
command=${VENV_DIR}/bin/uvicorn main:app --host 127.0.0.1 --port 8000 --workers 2
directory=${BACKEND_DIR}
user=root
autostart=true
autorestart=true
startsecs=5
startretries=3
redirect_stderr=true
stdout_logfile=${PROJECT_DIR}/logs/uvicorn.log
stdout_logfile_maxbytes=10MB
stdout_logfile_backups=5
environment=PATH="${VENV_DIR}/bin:/usr/bin:/bin"
SUP_EOF
    echo "Supervisor 配置已生成"
    
    # 尝试重载 Supervisor
    supervisorctl reread 2>/dev/null && supervisorctl update 2>/dev/null || true
fi

# 同时创建 systemd 服务作为备用方案
cat > /etc/systemd/system/wholesale.service << SERVICE_EOF
[Unit]
Description=Cambodia Wholesale Backend (uvicorn)
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${BACKEND_DIR}
ExecStart=${VENV_DIR}/bin/uvicorn main:app --host 127.0.0.1 --port 8000 --workers 2
Restart=always
RestartSec=5
Environment=PATH=${VENV_DIR}/bin:/usr/bin:/bin
StandardOutput=append:${PROJECT_DIR}/logs/uvicorn.log
StandardError=append:${PROJECT_DIR}/logs/uvicorn_error.log

[Install]
WantedBy=multi-user.target
SERVICE_EOF

systemctl daemon-reload

# 如果没有Supervisor，就用systemd启动
if [ ! -d "${SUPERVISOR_CONF_DIR}" ]; then
    echo "使用 systemd 启动后端服务..."
    systemctl restart wholesale
    systemctl enable wholesale
fi

echo "后端进程守护配置完成 ✓"

# ===== 第八步: 配置定时备份 =====
echo ""
echo -e "${GREEN}[8/8] 配置数据库定时备份...${NC}"

# 复制备份脚本
if [ -f "${SCRIPT_DIR}/deploy/backup_db.sh" ]; then
    cp "${SCRIPT_DIR}/deploy/backup_db.sh" "${PROJECT_DIR}/backup_db.sh"
    chmod +x "${PROJECT_DIR}/backup_db.sh"
fi

# 添加宝塔计划任务 (通过crontab, 宝塔面板也会读取)
CRON_CMD="0 3 * * * /bin/bash ${PROJECT_DIR}/backup_db.sh >> ${PROJECT_DIR}/logs/backup.log 2>&1"
(crontab -l 2>/dev/null | grep -v "backup_db.sh"; echo "${CRON_CMD}") | crontab -
echo "每日凌晨3点自动备份已配置 ✓"

# ===== 配置防火墙 =====
echo ""
echo -e "${GREEN}配置防火墙...${NC}"

# 宝塔自带防火墙管理，但确保必要端口开放
if command -v firewall-cmd &> /dev/null; then
    firewall-cmd --permanent --add-port=80/tcp 2>/dev/null || true
    firewall-cmd --permanent --add-port=443/tcp 2>/dev/null || true
    firewall-cmd --permanent --add-port=8888/tcp 2>/dev/null || true
    firewall-cmd --reload 2>/dev/null || true
elif command -v ufw &> /dev/null; then
    ufw allow 80/tcp 2>/dev/null || true
    ufw allow 443/tcp 2>/dev/null || true
    ufw allow 8888/tcp 2>/dev/null || true
fi
echo "防火墙端口已开放 (80, 443, 8888) ✓"

# ===== 完成 =====
echo ""
echo "=============================================="
echo -e "${GREEN}  部署完成！${NC}"
echo "=============================================="
echo ""
echo "  📌 宝塔面板:  http://${SERVER_IP}:8888"
echo "  📌 网站前端:  http://${SERVER_IP}"
echo "  📌 API 文档:  http://${SERVER_IP}/api/docs"
echo ""
echo "  📁 项目目录:  ${PROJECT_DIR}"
echo "  📁 站点目录:  ${SITE_DIR}"
echo "  📁 日志目录:  ${PROJECT_DIR}/logs"
echo "  📁 备份目录:  ${PROJECT_DIR}/backups"
echo ""
echo "  🔑 默认管理员: admin / admin123"
echo ""
echo "----------------------------------------------"
echo "  常用命令:"
echo "  查看后端日志:  tail -f ${PROJECT_DIR}/logs/uvicorn.log"
echo "  重启后端:      systemctl restart wholesale"
echo "  重启 Nginx:    /www/server/nginx/sbin/nginx -s reload"
echo "  宝塔命令:      bt"
echo "=============================================="
