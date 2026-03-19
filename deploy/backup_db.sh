#!/bin/bash
# ===========================================
# 数据库自动备份脚本
# 功能: 备份SQLite数据库，保留最近7天的备份
# 可选: 通过邮件发送 或 SCP同步到远程电脑
# ===========================================

# ---- 配置 ----
DB_PATH="/opt/wholesale/backend/cambodia_wholesale.db"
BACKUP_DIR="/opt/wholesale/backups"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="wholesale_backup_${DATE}.db"
KEEP_DAYS=7

# 邮件配置（可选，需先安装 msmtp 或 ssmtp）
ENABLE_EMAIL=false
EMAIL_TO="your-email@example.com"

# SCP 同步配置（可选，需配置SSH免密登录）
ENABLE_SCP=false
SCP_USER="Administrator"
SCP_HOST="your-pc-ip"
SCP_PATH="/Users/Administrator/Desktop/db_backups/"

# ---- 开始备份 ----
echo "[$(date)] 开始数据库备份..."

# 创建备份目录
mkdir -p "$BACKUP_DIR"

# 使用 sqlite3 的 .backup 命令安全备份（不会锁表）
if command -v sqlite3 &> /dev/null; then
    sqlite3 "$DB_PATH" ".backup '$BACKUP_DIR/$BACKUP_FILE'"
else
    # 如果没有sqlite3，直接复制（WAL模式下也安全）
    cp "$DB_PATH" "$BACKUP_DIR/$BACKUP_FILE"
    # 同时复制WAL文件
    [ -f "${DB_PATH}-wal" ] && cp "${DB_PATH}-wal" "$BACKUP_DIR/${BACKUP_FILE}-wal"
fi

# 压缩备份
gzip "$BACKUP_DIR/$BACKUP_FILE"
BACKUP_GZ="$BACKUP_DIR/${BACKUP_FILE}.gz"

echo "[$(date)] 备份完成: $BACKUP_GZ ($(du -h "$BACKUP_GZ" | cut -f1))"

# ---- 清理旧备份（保留最近N天）----
find "$BACKUP_DIR" -name "wholesale_backup_*.gz" -mtime +$KEEP_DAYS -delete
echo "[$(date)] 已清理 ${KEEP_DAYS} 天前的旧备份"

# ---- 邮件发送（可选）----
if [ "$ENABLE_EMAIL" = true ]; then
    if command -v mail &> /dev/null; then
        echo "数据库备份 - $DATE" | mail -s "批发系统数据库备份 $DATE" -A "$BACKUP_GZ" "$EMAIL_TO"
        echo "[$(date)] 备份已发送到 $EMAIL_TO"
    else
        echo "[$(date)] 警告: mail 命令不可用，跳过邮件发送"
    fi
fi

# ---- SCP 同步到本地电脑（可选）----
if [ "$ENABLE_SCP" = true ]; then
    scp "$BACKUP_GZ" "${SCP_USER}@${SCP_HOST}:${SCP_PATH}"
    if [ $? -eq 0 ]; then
        echo "[$(date)] 备份已同步到 ${SCP_HOST}:${SCP_PATH}"
    else
        echo "[$(date)] 警告: SCP同步失败"
    fi
fi

echo "[$(date)] 备份流程结束"
