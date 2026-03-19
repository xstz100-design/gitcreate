"""快速更新部署脚本 - 上传改动文件并执行迁移"""
import os
import paramiko

HOST = "43.134.13.229"
USER = "ubuntu"
PASSWORD = "qer1235A@"
REMOTE_BASE = "/opt/wholesale"
LOCAL_BASE = os.path.dirname(os.path.abspath(__file__))

BACKEND_FILES = [
    "backend/app/models/base.py",
    "backend/app/models/order.py",
    "backend/app/models/monthly_bill.py",
    "backend/app/models/category.py",
    "backend/app/models/__init__.py",
    "backend/app/api/schemas.py",
    "backend/app/api/auth.py",
    "backend/app/api/products.py",
    "backend/app/api/orders.py",
    "backend/app/api/billing.py",
    "backend/app/api/categories.py",
    "backend/main.py",
]

MIGRATION_SCRIPT = '''
import sqlite3
db_path = '/opt/wholesale/backend/cambodia_wholesale.db'
conn = sqlite3.connect(db_path)
c = conn.cursor()

def add_col(table, column, col_type, default=None):
    c.execute(f"PRAGMA table_info({table})")
    cols = [row[1] for row in c.fetchall()]
    if column not in cols:
        defstr = f" DEFAULT {default}" if default is not None else ""
        c.execute(f"ALTER TABLE {table} ADD COLUMN {column} {col_type}{defstr}")
        print(f"  Added {table}.{column}")
    else:
        print(f"  {table}.{column} already exists")

add_col('users', 'must_change_password', 'BOOLEAN', '1')
add_col('users', 'billing_day', 'INTEGER', 'NULL')
add_col('products', 'specs', 'VARCHAR(100)', 'NULL')
add_col('products', 'img4', 'VARCHAR(500)', 'NULL')
add_col('products', 'img5', 'VARCHAR(500)', 'NULL')
add_col('products', 'barcode', 'VARCHAR(100)', 'NULL')
add_col('users', 'allow_monthly_billing', 'BOOLEAN', '0')

# Migrate old payment_status values to new ones (SQLAlchemy stores enum NAME in uppercase)
c.execute("UPDATE orders SET payment_status = 'UNPAID' WHERE payment_status = 'PENDING'")
c.execute("UPDATE orders SET payment_status = 'CASH' WHERE payment_status = 'PAID'")
c.execute("UPDATE orders SET payment_status = 'MONTHLY' WHERE payment_status = 'CREDIT'")
rows_updated = c.execute("SELECT changes()").fetchone()[0]
print(f"  Migrated payment_status values (last batch: {rows_updated} rows)")

# Create monthly_bills table if not exists
c.execute("""
CREATE TABLE IF NOT EXISTS monthlybill (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    merchant_id INTEGER NOT NULL REFERENCES users(id),
    year INTEGER NOT NULL,
    month INTEGER NOT NULL,
    total_amount REAL NOT NULL DEFAULT 0,
    paid_amount REAL NOT NULL DEFAULT 0,
    status VARCHAR(10) NOT NULL DEFAULT 'unpaid',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)
""")
# Check if index exists
c.execute("SELECT name FROM sqlite_master WHERE type='index' AND name='ix_monthlybill_merchant_id'")
if not c.fetchone():
    c.execute("CREATE INDEX ix_monthlybill_merchant_id ON monthlybill(merchant_id)")
    print("  Created index ix_monthlybill_merchant_id")
else:
    print("  Index ix_monthlybill_merchant_id already exists")
print("  monthlybill table ready")

# Admin should not be forced to change password
c.execute("UPDATE users SET must_change_password = 0 WHERE username = 'admin'")
print("  Admin must_change_password set to 0")

conn.commit()
conn.close()
print("Migration complete!")
'''


def sftp_upload_dir(sftp, local_dir, remote_dir):
    for item in os.listdir(local_dir):
        local_path = os.path.join(local_dir, item)
        remote_path = f"{remote_dir}/{item}"
        if os.path.isdir(local_path):
            try:
                sftp.stat(remote_path)
            except FileNotFoundError:
                sftp.mkdir(remote_path)
            sftp_upload_dir(sftp, local_path, remote_path)
        else:
            print(f"    {item}")
            sftp.put(local_path, remote_path)


def run_cmd(ssh, cmd):
    stdin, stdout, stderr = ssh.exec_command(cmd, timeout=120)
    out = stdout.read().decode()
    err = stderr.read().decode()
    return out, err


def main():
    print(f"Connecting to {HOST}...")
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    ssh.connect(HOST, username=USER, password=PASSWORD, timeout=15)
    sftp = ssh.open_sftp()
    print("Connected!\n")

    # 1. Upload backend
    print("=== 1. Uploading backend files ===")
    for rel in BACKEND_FILES:
        local = os.path.join(LOCAL_BASE, rel)
        remote = f"{REMOTE_BASE}/{rel}"
        print(f"  {rel}")
        sftp.put(local, remote)
    print()

    # 2. Upload frontend dist
    print("=== 2. Uploading frontend dist ===")
    remote_dist = f"{REMOTE_BASE}/frontend/dist"
    # Clean old dist and recreate with correct ownership
    run_cmd(ssh, f"sudo rm -rf {remote_dist} && sudo mkdir -p {remote_dist}/assets && sudo chown -R ubuntu:ubuntu {remote_dist}")
    local_dist = os.path.join(LOCAL_BASE, "frontend", "dist")
    sftp_upload_dir(sftp, local_dist, remote_dist)
    print()

    # 3. Fix permissions
    print("=== 3. Fixing permissions ===")
    out, err = run_cmd(ssh, f"sudo chown -R www-data:www-data {remote_dist} && sudo chmod -R 755 {remote_dist}")
    print("  Done\n")

    # 4. DB migration
    print("=== 4. Running DB migration ===")
    with sftp.open("/tmp/migrate_update.py", "w") as f:
        f.write(MIGRATION_SCRIPT)
    out, err = run_cmd(ssh, "/opt/wholesale/venv/bin/python3 /tmp/migrate_update.py")
    print(out)
    if err:
        print(f"  Error: {err}")

    # 5. Restart backend
    print("=== 5. Restarting services ===")
    out, err = run_cmd(ssh, "sudo systemctl restart wholesale")
    out2, _ = run_cmd(ssh, "sudo systemctl is-active wholesale")
    print(f"  wholesale: {out2.strip()}")

    out, err = run_cmd(ssh, "sudo systemctl reload nginx")
    out2, _ = run_cmd(ssh, "sudo systemctl is-active nginx")
    print(f"  nginx: {out2.strip()}")

    # 6. Quick health check
    print("\n=== 6. Health check ===")
    out, err = run_cmd(ssh, "curl -s -o /dev/null -w '%{http_code}' http://localhost/api/auth/me || echo 'FAIL'")
    print(f"  API /auth/me: {out}")
    out, err = run_cmd(ssh, "curl -s -o /dev/null -w '%{http_code}' http://localhost/ || echo 'FAIL'")
    print(f"  Frontend: {out}")

    sftp.close()
    ssh.close()
    print(f"\n=== Deployment complete! http://{HOST} ===")


if __name__ == "__main__":
    main()
