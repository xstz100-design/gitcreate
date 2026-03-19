# 宝塔面板部署指南

## 说明

本文档介绍如何通过 **宝塔面板 (BT Panel)** 部署柬埔寨批发管理系统。宝塔面板提供可视化的服务器管理界面，适合不熟悉 Linux 命令行的用户。

---

## 一、一键部署（推荐）

将项目上传到服务器后，运行一键脚本：

```bash
# 上传项目到服务器
scp -r . root@YOUR_SERVER_IP:/opt/wholesale/

# SSH 登录服务器
ssh root@YOUR_SERVER_IP

# 运行部署脚本
cd /opt/wholesale
chmod +x deploy/bt_setup.sh
bash deploy/bt_setup.sh
```

脚本会自动完成：
1. 安装宝塔面板（如未安装）
2. 检查 Nginx / Python / Node.js
3. 部署后端（虚拟环境 + 依赖 + 数据库初始化）
4. 构建前端（npm install + build）
5. 配置 Nginx 反向代理
6. 配置 Supervisor 进程守护
7. 配置每日自动备份

---

## 二、手动部署（通过宝塔面板操作）

### 2.1 安装宝塔面板

**CentOS:**
```bash
yum install -y wget && wget -O install.sh https://download.bt.cn/install/install_6.0.sh && bash install.sh ed8484bec
```

**Ubuntu / Debian:**
```bash
apt-get update && apt-get install -y wget && wget -O install.sh https://download.bt.cn/install/install-ubuntu_6.0.sh && bash install.sh ed8484bec
```

安装完成后记录面板地址和账号密码，浏览器访问 `http://服务器IP:8888`。

### 2.2 安装必要软件

在宝塔面板 → **软件商店** 中安装：

| 软件 | 推荐版本 | 说明 |
|------|---------|------|
| Nginx | 1.24+ | Web 服务器和反向代理 |
| PM2管理器 或 Supervisor管理器 | 最新版 | 管理后端进程 |

> Python 3 和 Node.js 可在面板中安装或使用命令行安装

### 2.3 部署后端

```bash
# 创建项目目录
mkdir -p /opt/wholesale/backend
mkdir -p /opt/wholesale/logs

# 上传后端代码到 /opt/wholesale/backend/

# 创建虚拟环境
python3 -m venv /opt/wholesale/venv
source /opt/wholesale/venv/bin/activate

# 安装依赖
pip install --upgrade pip
pip install -r /opt/wholesale/backend/requirements.txt

# 初始化数据库
cd /opt/wholesale/backend
python init_db.py
```

### 2.4 构建前端

```bash
# 上传前端代码
# 本地构建后上传 dist 目录更快:
cd frontend
npm install
npm run build

# 将 dist/ 目录内容上传到 /www/wwwroot/wholesale/
```

或在服务器上构建：
```bash
cd /opt/wholesale/frontend
npm install
npm run build
cp -r dist/* /www/wwwroot/wholesale/
```

### 2.5 配置网站

#### 方法 A：通过面板添加（推荐）

1. 宝塔面板 → **网站** → **添加站点**
2. 域名填写：服务器IP 或 你的域名
3. 根目录设为：`/www/wwwroot/wholesale`
4. PHP版本选：**纯静态**
5. 创建后点 **设置** → **配置文件**
6. 将 `deploy/bt_nginx.conf` 的内容粘贴进去
7. 保存并重载 Nginx

#### 方法 B：直接写配置文件

```bash
# 复制配置到宝塔 Nginx 站点目录
cp /opt/wholesale/deploy/bt_nginx.conf /www/server/panel/vhost/nginx/wholesale.conf

# 修改 server_name 为你的域名或IP
vi /www/server/panel/vhost/nginx/wholesale.conf

# 测试并重载
/www/server/nginx/sbin/nginx -t
/www/server/nginx/sbin/nginx -s reload
```

### 2.6 配置后端进程守护

#### 方法 A：使用宝塔 Supervisor（推荐）

1. 宝塔面板 → **软件商店** → 安装 **Supervisor管理器**
2. 打开 Supervisor管理器 → **添加守护进程**
3. 填写：
   - **名称**: wholesale
   - **启动命令**: `/opt/wholesale/venv/bin/uvicorn main:app --host 127.0.0.1 --port 8000 --workers 2`
   - **运行目录**: `/opt/wholesale/backend`
   - **进程数量**: 1
   - **启动用户**: root
4. 点击 **确定** 即可自动启动

#### 方法 B：使用 systemd

```bash
# 复制服务文件
cp /opt/wholesale/deploy/wholesale.service /etc/systemd/system/

# 启动服务
systemctl daemon-reload
systemctl start wholesale
systemctl enable wholesale

# 查看状态
systemctl status wholesale
```

### 2.7 配置 SSL 证书（可选但推荐）

1. 宝塔面板 → **网站** → 点击站点 → **SSL**
2. 选择 **Let's Encrypt** → 填入域名 → 申请
3. 勾选 **强制HTTPS**
4. 证书会自动续期

> 注意：使用 Let's Encrypt 需要有域名指向服务器，IP 地址无法申请

### 2.8 配置定时备份

#### 方法 A：宝塔面板计划任务

1. 宝塔面板 → **计划任务**
2. 任务类型：**Shell脚本**
3. 任务名称：`wholesale 数据库备份`
4. 执行周期：每天凌晨 3:00
5. 脚本内容：
```bash
/bin/bash /opt/wholesale/backup_db.sh
```

#### 方法 B：已在一键脚本中通过 crontab 配置

---

## 三、目录结构（部署后）

```
/opt/wholesale/              # 项目根目录
├── backend/                 # 后端代码
│   ├── main.py
│   ├── app/
│   ├── uploads/             # 上传文件
│   └── cambodia_wholesale.db
├── frontend/                # 前端源码（可选保留）
├── venv/                    # Python虚拟环境
├── logs/                    # 日志目录
│   ├── uvicorn.log
│   └── backup.log
├── backups/                 # 数据库备份
├── deploy/                  # 部署配置文件
│   ├── bt_setup.sh
│   ├── bt_nginx.conf
│   └── backup_db.sh
└── backup_db.sh             # 备份脚本副本

/www/wwwroot/wholesale/      # 宝塔站点目录 (前端dist)
├── index.html
└── assets/
```

---

## 四、常用运维命令

```bash
# ---- 服务管理 ----
systemctl status wholesale         # 查看后端状态
systemctl restart wholesale        # 重启后端
systemctl stop wholesale           # 停止后端

# ---- 日志查看 ----
tail -f /opt/wholesale/logs/uvicorn.log     # 后端日志
tail -f /www/wwwlogs/wholesale.log          # Nginx 访问日志
tail -f /www/wwwlogs/wholesale.error.log    # Nginx 错误日志

# ---- Nginx ----
/www/server/nginx/sbin/nginx -t             # 测试配置
/www/server/nginx/sbin/nginx -s reload      # 重载配置

# ---- 宝塔命令 ----
bt                      # 宝塔命令行菜单
bt default              # 查看面板入口地址和账号
bt restart              # 重启面板
bt stop                 # 停止面板

# ---- 更新代码 ----
cd /opt/wholesale/backend && git pull       # 更新后端
cd /opt/wholesale/frontend && git pull && npm run build
cp -r dist/* /www/wwwroot/wholesale/        # 更新前端
systemctl restart wholesale                 # 重启后端
```

---

## 五、常见问题

### Q: 打开网站显示 502 Bad Gateway？
**A:** 后端服务未启动。检查后端进程：
```bash
systemctl status wholesale
# 或查看日志
tail -50 /opt/wholesale/logs/uvicorn.log
```

### Q: 页面刷新后 404？
**A:** Nginx 未配置 Vue Router history 模式的 try_files 回退。确认 Nginx 配置中包含：
```nginx
location / {
    try_files $uri $uri/ /index.html;
}
```

### Q: 图片上传后无法显示？
**A:** 检查 Nginx 的 uploads 别名路径和文件权限：
```bash
ls -la /opt/wholesale/backend/uploads/
chmod -R 755 /opt/wholesale/backend/uploads/
```

### Q: 宝塔面板无法访问？
**A:** 检查防火墙是否放行 8888 端口：
```bash
# CentOS
firewall-cmd --permanent --add-port=8888/tcp && firewall-cmd --reload

# Ubuntu
ufw allow 8888/tcp
```
也需要在云服务器安全组中放行该端口。

### Q: 如何修改后端端口？
**A:** 修改 Supervisor/systemd 中的启动命令端口号，同时更新 Nginx 配置中 `proxy_pass` 的端口。

---

## 六、安全建议

1. **修改默认密码**: 部署后立即修改 admin 密码
2. **修改 SECRET_KEY**: 编辑 `/opt/wholesale/backend/.env` 设置强随机密钥
3. **修改宝塔默认端口**: 宝塔面板 → 设置 → 面板端口（改为非标准端口）
4. **开启宝塔防火墙**: 软件商店 → 系统防火墙 → 安装并配置规则
5. **启用 SSL**: 有域名的情况下务必配置 HTTPS
6. **定期更新**: 保持宝塔面板和系统软件更新
