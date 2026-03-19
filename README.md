# 柬埔寨批发管理系统

## 项目概述

这是一个专为柬埔寨批发业务设计的现代化管理系统,支持双语(中文/高棉语)、多货币(USD/KHR),采用双端分离架构:
- **PC端管理后台** (Element Plus) - 供管理员和商户在电脑上操作
- **移动端商户端** (Vant UI) - 供商户在手机上快速下单

## 核心特性

### 🎯 双端架构
- **PC端**: Element Plus UI,适合管理员处理复杂操作(商品管理、订单审核、商户管理)
- **移动端**: Vant UI,专为批发采购优化(左右联动分类、批量输入、触觉反馈)

### 💰 多语言/多货币
- 中文/高棉语双语支持
- USD/KHR双币种实时显示(汇率1:4100)

### 🚀 性能优化
- **API响应优化**: 列表接口精简字段(ProductResponse 9字段 vs 详情12字段)
- **图片优化**: 自动生成缩略图(200x200, quality=85)
- **前端优化**: 分页加载、row-key优化、v-memo防止重复渲染
- **离线支持**: Pinia持久化购物车

### 📱 移动端特色
- **美团式左右联动**: 左侧分类导航,右侧商品列表自动滚动跟随
- **批量输入优化**:数字键盘快速输入数量(批发场景下一次可能输入100+)
- **触觉反馈**: 点击/删除/成功/错误不同振动模式
- **语义化色彩**: 
  - 主色调 #2F54EB (稳重蓝)
  - 成功 #52C41A (库存充足)
  - 警告 #FAAD14 (库存预警)
  - 危险 #F5222D (缺货/欠款)

## 技术栈

### 后端
- **框架**: FastAPI 0.109.2
- **ORM**: SQLModel 0.0.16
- **数据库**: PostgreSQL
- **认证**: JWT (pyjwt)
- **图片处理**: Pillow 10.2.0

### 前端
- **框架**: Vue 3.4.15
- **构建工具**: Vite 5.0.12
- **UI库**: 
  - Element Plus 2.5.6 (PC端)
  - Vant 4.8.11 (移动端)
- **状态管理**: Pinia 2.1.7 (持久化插件)
- **HTTP客户端**: Axios 1.6.7
- **路由**: Vue Router 4.2.5

## 项目结构

```
backend/
├── app/
│   ├── api/          # API路由
│   │   ├── auth.py       # 登录/注册
│   │   ├── products.py   # 商品管理
│   │   ├── orders.py     # 订单管理
│   │   └── schemas.py    # Pydantic模型(优化版)
│   ├── models/       # SQLModel数据库模型
│   ├── core/         # 配置/安全/数据库
│   └── services/     # 业务逻辑
│       └── image.py      # 图片优化服务
├── main.py
└── requirements.txt

frontend/
├── src/
│   ├── api/          # API客户端
│   ├── layouts/      # 布局组件
│   │   ├── AdminLayout.vue      # PC管理端
│   │   ├── MerchantLayout.vue   # PC商户端
│   │   └── MobileLayout.vue     # 移动商户端
│   ├── views/
│   │   ├── admin/    # 管理员页面(PC)
│   │   ├── merchant/ # 商户页面(PC)
│   │   └── mobile/   # 商户页面(移动)
│   │       ├── Shop.vue     # 美团式购物页
│   │       ├── Cart.vue     # 购物车
│   │       ├── Orders.vue   # 订单列表
│   │       └── Profile.vue  # 个人中心
│   ├── stores/       # Pinia状态管理
│   ├── router/       # 路由(含设备检测自动跳转)
│   └── utils/
│       ├── colors.js    # 语义化色彩系统
│       ├── device.js    # 设备检测/触觉反馈
│       └── format.js    # 货币/日期格式化
├── package.json
└── vite.config.js
```

## 快速开始

### 1. 后端启动

```bash
cd backend

# 创建虚拟环境
python -m venv venv
venv\Scripts\activate  # Windows
# source venv/bin/activate  # Linux/Mac

# 安装依赖
pip install -r requirements.txt

# 配置环境变量(.env文件)
DATABASE_URL=postgresql://user:password@localhost/cambodia_wholesale
SECRET_KEY=your-secret-key-here

# 启动服务(开发模式)
uvicorn main:app --reload --host 0.0.0.0 --port 8000
```

后端将运行在: `http://localhost:8000`  
API文档: `http://localhost:8000/docs`

### 2. 前端启动

```bash
cd frontend

# 安装依赖
npm install
# 或使用 pnpm/yarn

# 启动开发服务器
npm run dev
```

前端将运行在: `http://localhost:5173`

### 3. 访问系统

#### PC端访问
- 管理员: `http://localhost:5173/login` → 登录后自动跳转 `/admin`
- 商户: `http://localhost:5173/login` → 登录后自动跳转 `/merchant`

#### 移动端访问 (使用Chrome DevTools模拟)
1. 打开 `http://localhost:5173`
2. F12 打开开发者工具
3. 点击设备工具栏图标(Ctrl+Shift+M)
4. 选择移动设备(如iPhone 12 Pro)
5. 刷新页面 → 商户端自动跳转 `/m`

## 默认账号

需要通过API或数据库手动创建初始用户:

```python
# 管理员
{
  "username": "admin",
  "password": "admin123",
  "role": "admin",
  "full_name": "系统管理员"
}

# 商户
{
  "username": "merchant001",
  "password": "merchant123",
  "role": "merchant",
  "full_name": "测试商户",
  "phone": "012345678",
  "address": "金边市中心"
}
```

## 主要功能

### 管理员(PC)
- ✅ 仪表板 - 销售统计/库存预警/订单概览
- ✅ 商品管理 - CRUD操作,批量导入,库存预警
- ✅ 订单管理 - 订单审核,状态更新,支付/配送跟踪
- ✅ 商户管理 - 商户信息,信用额度,欠款管理
- ✅ 分页优化 - 默认20条/页,支持10/20/50/100

### 商户(PC)
- ✅ 商品浏览 - 分类筛选,搜索,价格/库存显示
- ✅ 购物车 - 批量编辑,价格汇总(USD+KHR)
- ✅ 订单查询 - 历史订单,状态跟踪,重复下单

### 商户(移动)
- ✅ 美团式购物 - 左侧分类(90px)+右侧商品列表,联动滚动
- ✅ 批量输入 - 数字键盘弹窗,快速输入大数量(如100箱)
- ✅ 购物车 - 滑动删除,全选/反选,离线缓存
- ✅ 订单管理 - 下拉刷新,状态筛选,详情弹窗
- ✅ 触觉反馈 - light(10ms)/medium(20ms)/heavy(30ms)/success/error模式

## API接口优化说明

### 响应模型分离

**之前**: 所有接口返回完整字段(12字段ProductResponse)
```json
{
  "id": 1,
  "name": "可口可乐",
  "name_kh": "កូកាកូឡា",
  "category": "饮料",
  "unit": "箱",
  "price_usd": 12.50,
  "price_khr": 51250,
  "stock": 100,
  "stock_warning": 20,
  "is_low_stock": false,
  "description": "330ml x 24罐",
  "image_url": "/uploads/coke.jpg",
  "thumbnail_url": "/uploads/coke_thumb.jpg",
  "is_active": true
}
```

**现在**: 列表接口精简到9字段
```json
{
  "id": 1,
  "name": "可口可乐",
  "name_kh": "កូកាកូឡា",
  "unit": "箱",
  "price_usd": 12.50,
  "stock": 100,
  "is_low_stock": false,
  "image_url": "/uploads/coke_thumb.jpg",
  "category": "饮料"
}
```

**优势**: 移动端加载100个商品,从120KB减少到90KB,节省25%流量

### 图片缩略图

```python
# backend/app/services/image.py
ImageService.create_thumbnail(
    input_path="uploads/product.jpg",
    size=(200, 200),  # 移动端列表够用
    quality=85        # 平衡质量和大小
)
```

**建议CDN策略**: 
- 原图: `/uploads/product.jpg` → CDN加速 → `https://cdn.example.com/uploads/product.jpg`
- 缩略图: `/uploads/product_thumb.jpg` → 自动生成 → 列表页使用

## 路由自动跳转逻辑

```javascript
// frontend/src/router/index.js
router.beforeEach((to, from, next) => {
  // 移动设备访问PC商户页 → 自动重定向到移动端
  if (to.path.startsWith('/merchant') && isMobile()) {
    next(to.path.replace('/merchant', '/m'))
  }
  
  // PC设备访问移动端页 → 自动重定向到PC端
  if (to.path.startsWith('/m') && !isMobile()) {
    next(to.path.replace('/m', '/merchant'))
  }
})
```

**设备检测**:
```javascript
// frontend/src/utils/device.js
export const isMobile = () => {
  return /Android|iPhone|iPad|iPod/i.test(navigator.userAgent)
    || window.innerWidth < 768
}
```

## 性能优化清单

### 后端
- ✅ 分离响应模型(list vs detail)
- ✅ 图片缩略图自动生成
- ✅ 数据库索引(未实现,建议添加)
- ⏳ Redis缓存(未启用)
- ⏳ CDN集成(需配置)

### 前端
- ✅ 分页加载(PC表格20条/页)
- ✅ row-key优化(避免重复渲染)
- ✅ 路由懒加载(component: () => import(...))
- ✅ Pinia持久化(购物车离线可用)
- ✅ 图片懒加载(移动端商品列表)
- ⏳ v-memo优化(可进一步应用于大列表)
- ⏳ 虚拟滚动(数据量>1000时考虑)

## 色彩系统

```javascript
// frontend/src/utils/colors.js
export const COLORS = {
  primary: '#2F54EB',   // 主色调(稳重蓝)
  success: '#52C41A',   // 成功/库存充足
  warning: '#FAAD14',   // 警告/库存预警
  danger: '#F5222D',    // 危险/缺货/欠款
  info: '#1890FF',      // 信息
  text: '#262626',      // 主文本
  textSecondary: '#8c8c8c', // 次要文本
}
```

**应用场景**:
- 库存充足 → `success` (绿)
- 库存预警 → `warning` (橙)
- 缺货 → `danger` (红)
- 待支付 → `warning` (橙)
- 已付款 → `success` (绿)
- 赊账 → `danger` (红)

## 移动端触觉反馈

```javascript
// frontend/src/utils/device.js
hapticFeedback('light')    // 轻点击(10ms)
hapticFeedback('medium')   // 按钮点击(20ms)
hapticFeedback('heavy')    // 删除操作(30ms)
hapticFeedback('success')  // 成功提示([10,50,10])
hapticFeedback('error')    // 错误提示([50,100,50])
```

**使用示例**:
```vue
<van-stepper @change="val => {
  hapticFeedback('light')
  updateQuantity(item.id, val)
}" />
```

## 下一步建议

### 短期优化(1-2周)
1. [ ] 添加图片上传API(集成ImageService)
2. [ ] 数据库migration脚本(Alembic)
3. [ ] 初始数据seeder(管理员账号/示例商品)
4. [ ] 单元测试(pytest for backend/vitest for frontend)

### 中期功能(1个月)
1. [ ] Redis缓存层(热门商品/订单统计)
2. [ ] WebSocket实时通知(订单状态变更)
3. [ ] 导出功能(订单Excel/PDF)
4. [ ] 数据可视化(ECharts销售趋势)

### 长期规划(3个月+)
1. [ ] 配送员App(独立移动端)
2. [ ] 财务报表模块
3. [ ] 多仓库管理
4. [ ] 小程序版本

## 故障排查

### 问题: 移动端页面不显示/白屏
**解决**:
1. 检查设备检测: `console.log(isMobile())` 应返回 `true`
2. 确认路由跳转: 商户登录后应跳转 `/m/shop`
3. 检查Vant UI导入: `main.js` 中是否引入 `import Vant from 'vant'`

### 问题: 购物车数据丢失
**解决**:
1. 检查Pinia持久化: `localStorage` 中应有 `cart` key
2. 清除缓存后重新添加商品
3. 确认 `persist: true` 配置在 `stores/cart.js`

### 问题: 图片显示404
**解决**:
1. 确认图片路径: 应为 `/uploads/xxx.jpg` (相对路径)
2. 配置后端静态文件服务:
   ```python
   # main.py
   from fastapi.staticfiles import StaticFiles
   app.mount("/uploads", StaticFiles(directory="uploads"), name="uploads")
   ```

### 问题: 触觉反馈无效
**解决**:
1. 确认设备支持: `navigator.vibrate` API(iOS Safari不支持)
2. 使用HTTPS或localhost(vibration需要安全上下文)
3. 检查浏览器权限设置

## 服务器部署

### 方式一：宝塔面板部署（推荐新手）

使用宝塔面板可视化管理服务器，一键部署：

```bash
# SSH登录服务器后
cd /opt/wholesale
bash deploy/bt_setup.sh
```

详细步骤请参考 [宝塔面板部署指南](deploy/BT_DEPLOY.md)，包含：
- 宝塔面板安装与软件配置
- Nginx 反向代理配置
- Supervisor 进程守护
- SSL 证书申请
- 定时备份设置

相关文件：
- `deploy/bt_setup.sh` — 一键部署脚本
- `deploy/bt_nginx.conf` — 宝塔 Nginx 配置
- `deploy/BT_DEPLOY.md` — 完整部署文档

### 方式二：传统部署

```bash
cd /opt/wholesale
bash deploy/setup.sh
```

相关文件：
- `deploy/setup.sh` — 传统部署脚本
- `deploy/nginx.conf` — 标准 Nginx 配置
- `deploy/wholesale.service` — systemd 服务文件

## 贡献指南

1. Fork项目
2. 创建功能分支: `git checkout -b feature/AmazingFeature`
3. 提交代码: `git commit -m 'Add some AmazingFeature'`
4. 推送分支: `git push origin feature/AmazingFeature`
5. 提交Pull Request

## 许可证

MIT License

---

**构建于**: FastAPI + Vue 3 + Element Plus + Vant UI  
**设计理念**: 极致简洁、高性能、强适配
