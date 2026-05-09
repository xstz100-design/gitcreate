# 柬埔寨批发管理系统

面向柬埔寨批发业务的 B2B 下单与管理系统，支持 PC 管理后台、PC 商户端、移动商户端和 Telegram Mini App 场景。

- **线上域名**：`https://tfyx.shop`
- **服务器**：Ubuntu 22.04 / 43.134.13.229
- **后端**：Go 1.22（Gin + GORM + SQLite，无 CGO，单二进制部署）
- **前端**：Vue 3 + Vite 5，三语界面 **中文 / English / ខ្មែរ**（柬埔寨语）

---

## 整体架构


```
用户浏览器 / Telegram Mini App
            │
            ▼
       Nginx 80/443
            │
   ┌────────┴─────────┐
   │                  │
   ▼                  ▼
Vue dist          Go /api
静态资源          Gin 127.0.0.1:8000
                       │
                       ▼
                 SQLite (WAL 模式)
                 /opt/wholesale/backend/cambodia_wholesale.db
```

### 三种使用场景

| 场景 | 路由前缀 | 适配 |
|---|---|---|
| PC 管理后台 | `/admin/*` | Element Plus，管理员专用 |
| PC 商户端 | `/merchant/*` | Element Plus，商户账号 |
| 移动/Telegram 商户端 | `/m/*` | Vant，手机和 Telegram Mini App |

前端路由守卫会自动识别设备类型，移动设备访问 `/merchant/*` 时自动跳转到 `/m/*`，反之亦然。

---

## 目录结构

```
├── backend-go/               # Go 后端（生产使用）
│   ├── main.go               # 入口：路由注册、中间件挂载、服务启动
│   ├── go.mod / go.sum
│   ├── handlers/             # HTTP 处理器，按业务模块划分
│   │   ├── auth.go           # 登录、注册、用户管理、审核、Telegram 绑定
│   │   ├── products.go       # 商品 CRUD、条码查询、批量导入
│   │   ├── orders.go         # 订单创建/查询/取消/配货/配送状态
│   │   ├── categories.go     # 分类管理
│   │   ├── announcements.go  # 公告管理
│   │   ├── billing.go        # 月结账单
│   │   ├── settings.go       # 系统设置（配送费、联系信息、Telegram）
│   │   ├── upload.go         # 图片上传（压缩 + 缩略图）
│   │   └── disk_*.go         # 磁盘空间查询（跨平台）
│   ├── middleware/           # JWT 认证、角色权限守卫
│   ├── models/
│   │   └── models.go         # 全部数据模型（User、Product、Order 等）
│   ├── database/             # GORM 初始化、迁移
│   ├── services/
│   │   ├── notify.go         # Telegram 通知（下单/配送/库存预警）
│   │   ├── image.go          # 图片处理（disintegration/imaging）
│   │   └── settings.go       # 系统设置读写服务
│   ├── config/               # 环境变量读取（godotenv）
│   ├── utils/                # JWT 生成/验证、密码哈希、随机密码
│   └── bot/                  # Telegram Bot 消息发送
│
├── backend/                  # Python FastAPI 后端（已归档，不再使用）
│
├── frontend/                 # Vue 3 前端
│   ├── index.html
│   ├── vite.config.js
│   ├── package.json
│   └── src/
│       ├── api/index.js      # 全部 Axios API 封装
│       ├── components/       # SkeletonProduct、SkeletonTable
│       ├── i18n/             # 三语支持（zh.js / en.js / kh.js）
│       ├── layouts/          # AdminLayout、MerchantLayout、MobileLayout
│       ├── router/index.js   # 路由定义 + 设备自动跳转守卫
│       ├── stores/           # user.js（登录态）、cart.js（购物车持久化）
│       ├── styles/           # 全局 SCSS、Element Plus 覆盖、变量
│       ├── utils/            # request、telegram、device、format、colors
│       └── views/
│           ├── Login.vue
│           ├── admin/        # Dashboard、Products、Orders、Merchants...
│           ├── merchant/     # Products、Cart、Orders、Profile
│           └── mobile/       # Shop、Cart、Orders、Profile
│
├── deploy/                   # 部署相关
│   ├── wholesale.service     # systemd 服务定义
│   ├── nginx.conf            # Nginx 配置
│   ├── bt_nginx.conf         # 宝塔 Nginx 配置
│   ├── setup.sh              # 传统部署初始化脚本
│   ├── bt_setup.sh           # 宝塔部署脚本
│   ├── backup_db.sh          # 数据库定时备份脚本
│   └── BT_DEPLOY.md          # 宝塔部署文档
│
├── .gitignore
├── README.md
└── PROJECT_DETAIL.md
```

---

## API 路由总览

所有接口以 `/api` 为前缀，分以下权限层级：
- **公开**：无需 Token
- **需登录**：JWT Token（管理员 + 商户均可）
- **管理员**：`role = admin`
- **超级管理员**：`is_super_admin = true`

### 认证 `/api/auth`

| 方法 | 路径 | 权限 | 说明 |
|---|---|---|---|
| POST | `/login` | 公开 | 用户名密码登录，返回 JWT |
| POST | `/telegram-auth` | 公开 | Telegram Mini App 免登录 |
| POST | `/phone-verification/send` | 公开 | 发送手机验证码（存根） |
| POST | `/phone-verification/verify` | 公开 | 验证手机验证码（存根） |
| GET | `/me` | 需登录 | 获取当前用户信息 |
| PATCH | `/me` | 需登录 | 更新当前用户资料 |
| PATCH | `/me/telegram` | 需登录 | 更新 Telegram ID |
| POST | `/me/telegram/bind-current` | 需登录 | 绑定 Telegram（验证 initData） |
| POST | `/change-password` | 需登录 | 修改密码 |
| POST | `/submit-review` | 需登录 | 商户提交/重新提交资料审核 |
| GET | `/users` | 管理员 | 用户列表 |
| POST | `/register` | 管理员 | 创建新用户（返回临时密码） |
| GET | `/users/:id` | 管理员 | 用户详情 |
| PATCH | `/users/:id` | 管理员 | 更新用户信息 |
| DELETE | `/users/:id` | 管理员 | 停用用户 |
| POST | `/users/:id/approve` | 管理员 | 审核通过/拒绝商户 |
| GET | `/pending-users` | 管理员 | 待审核用户列表 |
| GET | `/pending-count` | 管理员 | 待审核用户数量 |
| GET | `/all-registrations` | 管理员 | 所有注册用户（可按状态筛选） |
| POST | `/users/:id/reset-password` | 管理员 | 重置密码（返回临时密码） |
| GET | `/dashboard` | 管理员 | 仪表盘统计 |
| PATCH | `/users/:id/super-admin` | 超级管理员 | 设置/取消超级管理员 |

### 商品 `/api/products`

| 方法 | 路径 | 权限 | 说明 |
|---|---|---|---|
| GET | `` | 需登录 | 商品列表（支持搜索、分类、分页） |
| GET | `/barcode/:barcode` | 需登录 | 按条码查询商品 |
| GET | `/:id` | 需登录 | 商品详情 |
| POST | `` | 管理员 | 新增商品 |
| GET | `/import/template` | 管理员 | 下载导入模板 CSV |
| POST | `/import` | 管理员 | 批量导入（待实现） |
| PATCH | `/:id` | 管理员 | 更新商品 |
| DELETE | `/:id` | 管理员 | 软删除商品 |

### 订单 `/api/orders`

| 方法 | 路径 | 权限 | 说明 |
|---|---|---|---|
| GET | `` | 需登录 | 订单列表（商户只看自己的） |
| POST | `` | 需登录 | 创建订单（扣库存、幂等校验） |
| GET | `/picker/items/:orderId` | 需登录 | 配货员视图 |
| GET | `/:id` | 需登录 | 订单详情 |
| PATCH | `/:id` | 需登录 | 更新订单 |
| POST | `/:id/cancel` | 需登录 | 取消订单（仅 pending，恢复库存） |
| POST | `/:id/pick` | 需登录 | 标记配货完成 |
| DELETE | `/:id` | 管理员 | 软删除订单（恢复库存） |

### 其他

| 前缀 | 说明 |
|---|---|
| `/api/categories` | 分类：公开查询、管理员 CRUD |
| `/api/announcements` | 公告：公开查询、管理员 CRUD |
| `/api/billing` | 月结账单：生成、列表、更新 |
| `/api/settings` | 配送费、联系信息、Telegram Chat ID 管理 |
| `/api/upload/image` | 图片上传（自动压缩 + 缩略图） |

---

## 核心业务流程

### 1. 商户注册与审核

```
商户注册（自助 or 管理员代建）
        │
        ▼
   完善个人资料（姓名/电话/地址）
        │
        ▼
   提交审核  POST /api/auth/submit-review
        │
        ▼
   管理员在审核页面处理
        │
   ┌────┴────┐
   ▼         ▼
通过        拒绝（记录拒绝原因）
   │
   ▼
商户可正常下单
```

审核状态流转：`pending` → `approved` / `rejected`

未通过审核的商户可浏览商品和使用购物车，提交订单时被拦截。

### 5. 语言切换

语言按钮在所有布局（AdminLayout、MerchantLayout、MobileLayout）及登录页均可切换，按顺序循环：

```
中文 → English → ខ្មែរ → 中文
```

语言偏好通过 `localStorage` 持久化。翻译文件结构：

| 文件 | 语言 |
|---|---|
| `src/i18n/zh.js` | 中文 |
| `src/i18n/en.js` | English |
| `src/i18n/kh.js` | ខ្មែរ（柬埔寨语）|

---

### 2. 下单流程

```
商户加购物车（Pinia 持久化）
        │
        ▼
确认地址/电话/备注/付款方式
前端生成 client_request_id
        │
        ▼
POST /api/orders
        │
   后端依次校验：
   ① 审核状态 approved
   ② 月结权限
   ③ 幂等：merchant_id + client_request_id
   ④ 逐项检查库存
   ⑤ 原子扣减库存
   ⑥ 创建订单和明细
        │
        ▼
返回订单详情 + 触发 Telegram 通知
```

### 3. 配送流程

```
delivery_status: pending
        │ 管理员安排配送
        ▼
delivery_status: delivering
（触发 Telegram 通知配送员）
        │ 配货员标记
        ▼
picked_at / picked_by_id 记录
        │ 完成送货
        ▼
delivery_status: delivered
delivered_at 记录
```

商户或管理员可在 `pending` 状态下取消订单，库存自动恢复。

### 4. 月结账单

```
管理员触发  POST /api/billing/generate
        │
        ▼
按月汇总各商户信用订单金额
生成 MonthlyBill 记录
        │
        ▼
管理员查看/标记已收款
PATCH /api/billing/:id
```

---

## 数据模型

### User

| 字段 | 说明 |
|---|---|
| `role` | `admin` / `merchant` |
| `is_super_admin` | 超级管理员标志 |
| `approval_status` | `pending` / `approved` / `rejected` |
| `allow_credit` | 是否允许月结 |
| `billing_cycle_days` | 账期天数 |
| `credit_limit` | 信用额度（美元） |
| `telegram_id` | Telegram 用户 ID |
| `notify_enabled` | 是否接收 Telegram 通知 |
| `must_change_password` | 首次登录强制改密 |
| `is_active` | 账号是否启用 |

### Product

| 字段 | 说明 |
|---|---|
| `name` / `name_kh` / `name_en` | 中/高棉/英文名 |
| `brand` | 品牌 |
| `barcode` | 条码（唯一） |
| `price_usd` | 批发价（美元） |
| `stock` | 当前库存 |
| `is_active` | 是否上架 |
| `is_deleted` | 软删除标志 |

### Order

| 字段 | 说明 |
|---|---|
| `order_no` | 订单号（系统生成，唯一） |
| `merchant_id` | 所属商户 |
| `payment_status` | `unpaid` / `cash` / `credit` |
| `delivery_status` | `pending` / `delivering` / `delivered` / `cancelled` |
| `client_request_id` | 防重复提交唯一标识 |
| `picked_at` / `picked_by_id` | 配货时间和操作人 |
| `delivered_at` | 签收时间 |
| `is_deleted` | 软删除标志 |

### StockLedger（库存流水）

| 字段 | 说明 |
|---|---|
| `product_id` | 关联商品 |
| `order_id` | 关联订单（可为空） |
| `delta` | 变动量（负数=扣减，正数=回补） |
| `stock_after` | 变动后库存快照 |
| `reason` | `order_create` / `order_cancel` / `order_delete` / `manual_adjust` / `import` |
| `operator_id` | 操作人（可为空） |
| `note` | 备注 |

---

## 权限模型

```
超级管理员（is_super_admin = true）
  ├── 拥有管理员全部能力
  ├── 可设置/取消其他账号的超级管理员状态
  └── 保护：不能删除/降级最后一个超级管理员

普通管理员（role = admin）
  ├── 管理商品、订单、分类、公告、商户、审核
  └── 无法操作超级管理员专属功能

商户（role = merchant）
  ├── 只能访问自己的资料和订单
  └── 购物车、下单、查看历史、取消未处理订单
```

---

## 安全机制

| 机制 | 实现 |
|---|---|
| JWT 认证 | `golang-jwt/jwt v5`，每个受保护请求验证 Token |
| 密码存储 | bcrypt rounds=12 |
| 防重复下单 | `merchant_id + client_request_id` 唯一索引；重复请求返回 HTTP 200 及已有订单（幂等） |
| 库存保护 | 原子 SQL 扣减避免超卖；所有变动写入 `stock_ledger` 流水表 |
| 商户隔离 | 订单 API 全部校验 `merchant_id = current_user_id` |
| 软删除 | 商品和订单使用 `is_deleted`，历史数据不丢失 |
| 上传安全 | MIME 类型白名单、UUID 重命名、图片压缩 |

---

## 技术栈

### 后端（Go 1.22）

| 依赖 | 用途 |
|---|---|
| `gin-gonic/gin v1.9.1` | HTTP 框架 |
| `gorm.io/gorm v1.25.10` | ORM |
| `glebarez/sqlite v1.11.0` | SQLite 纯 Go 驱动（无 CGO） |
| `golang-jwt/jwt v5.2.1` | JWT 认证 |
| `golang.org/x/crypto` | bcrypt 密码哈希 |
| `disintegration/imaging v1.6.2` | 图片压缩和缩略图 |
| `joho/godotenv v1.5.1` | .env 加载 |
| `google/uuid v1.6.0` | UUID 生成 |

### 前端

| 依赖 | 用途 |
|---|---|
| Vue 3 + Vite 5 | 核心框架 + 构建 |
| Vue Router 4 | 路由 |
| Pinia + persistedstate | 状态管理 + 持久化 |
| Element Plus | PC 端 UI |
| Vant 4 | 移动端 UI |
| vue-i18n | 中文 / English / ខ្មែរ 三语切换 |
| Axios | HTTP 请求 |
| Sass | CSS 预处理 |

---

## 本地开发

### 启动 Go 后端

```bash
cd backend-go
go mod download
go run .
# 默认监听 :8000
```

### 启动前端

```bash
cd frontend
npm install
npm run dev
# 默认地址：http://localhost:5173
```

### 环境变量（`backend-go/.env`）

| 变量 | 说明 | 默认值 |
|---|---|---|
| `PORT` | 监听端口 | `8000` |
| `GIN_MODE` | `release` / `debug` | `debug` |
| `SECRET_KEY` | JWT 签名密钥 | 内置默认（生产必须覆盖） |
| `DATABASE_PATH` | SQLite 文件路径 | `./cambodia_wholesale.db` |
| `TG_BOT_TOKEN` | Telegram Mini App 验证 Token | 可选 |
| `TELEGRAM_BOT_TOKEN` | 通知 Bot Token | 可选 |
| `USD_TO_KHR_RATE` | 美元兑瑞尔汇率 | `4000` |

---

## 编译与部署

### 交叉编译 Linux 二进制

```powershell
# Windows PowerShell
$env:GOOS = "linux"; $env:GOARCH = "amd64"; $env:CGO_ENABLED = "0"
go build -o wholesale .
```

```bash
# macOS / Linux
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o wholesale .
```

### 服务器目录约定

| 路径 | 内容 |
|---|---|
| `/opt/wholesale/backend/` | Go 二进制、.env、uploads/ |
| `/opt/wholesale/backend/cambodia_wholesale.db` | SQLite 数据库 |
| `/opt/wholesale/backend/uploads/` | 用户上传图片 |
| `/opt/wholesale/frontend/` | Vue dist 静态文件 |

### 更新部署

```bash
# 1. 本地编译（Windows）
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"
go build -o wholesale .

# 2. 上传并重启
scp wholesale ubuntu@43.134.13.229:/tmp/wholesale_new
ssh ubuntu@43.134.13.229 "sudo mv /tmp/wholesale_new /opt/wholesale/backend/wholesale && sudo chmod +x /opt/wholesale/backend/wholesale && sudo systemctl restart wholesale"
```

### systemd 服务（`deploy/wholesale.service`）

```ini
[Service]
Type=simple
User=root
WorkingDirectory=/opt/wholesale/backend
ExecStart=/opt/wholesale/backend/wholesale
Restart=always
RestartSec=5
Environment=GIN_MODE=release
```

---

## 数据库备份

```bash
# 脚本位置：deploy/backup_db.sh
# 建议通过 cron 每日凌晨执行
crontab -e
# 0 3 * * * /opt/wholesale/deploy/backup_db.sh
```

默认保留最近 7 天备份，输出到 `/opt/wholesale/backups/`。

---

## 初始账号

- 超级管理员用户名：`100001` / `100002`
- 初始密码由初始化脚本随机生成，首次登录后强制修改
- 管理员重置密码后，系统返回一次性临时密码，登录后立即修改

---

## 常见问题

**商户无法下单**
- `approval_status` 必须为 `approved`
- 资料完整（姓名、电话、地址）
- 月结付款需要 `allow_credit = true`

**管理员用户管理页白屏**
- 确认 `/admin/merchants` 路由已在 `router/index.js` 注册，否则页面空白

**移动端白屏**
- 确认是否跳转到了 `/m/shop`
- 检查 Telegram Mini App 视口高度初始化
- 确认 Vant 全局样式已引入

**订单重复**
- 前端：下单后按钮立即禁用
- 后端：`merchant_id + client_request_id` 幂等检查
- 数据库：唯一索引兜底

**忘记管理员密码**
```bash
# 在服务器上直接更新（需要 bcrypt 哈希）
sqlite3 /opt/wholesale/backend/cambodia_wholesale.db \
  "UPDATE users SET hashed_password='$HASH' WHERE username='100001';"
```
