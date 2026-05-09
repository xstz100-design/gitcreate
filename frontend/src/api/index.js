import request from '@/utils/request'

// ============= 认证相关 =============

/**
 * 用户登录
 */
export function login(username, password) {
  const formData = new FormData()
  formData.append('username', username)
  formData.append('password', password)
  
  return request({
    url: '/api/auth/login',
    method: 'post',
    data: formData,
  })
}

/**
 * Telegram Mini App 免登录
 */
export function telegramAuth(initData) {
  return request({
    url: '/api/auth/telegram-auth',
    method: 'post',
    data: { init_data: initData },
  })
}

/**
 * 获取当前用户信息
 */
export function getCurrentUser() {
  return request({
    url: '/api/auth/me',
    method: 'get',
  })
}

export function getDashboardMetrics(days = 7) {
  return request({
    url: '/api/auth/dashboard-metrics',
    method: 'get',
    params: { days },
  })
}

/**
 * 注册用户 - 仅管理员
 */
export function register(userData) {
  return request({
    url: '/api/auth/register',
    method: 'post',
    data: userData,
  })
}

/**
 * 获取待审核用户列表 - 仅管理员
 */
export function getPendingUsers() {
  return request({
    url: '/api/auth/pending-users',
    method: 'get',
  })
}

/**
 * 获取所有注册用户(可按状态筛选) - 仅管理员
 */
export function getAllRegistrations(status = null) {
  return request({
    url: '/api/auth/all-registrations',
    method: 'get',
    params: status ? { status } : {},
  })
}

/**
 * 审核用户 - 仅管理员
 */
export function approveUser(userId, approved, rejectedReason = null) {
  return request({
    url: `/api/auth/users/${userId}/approve`,
    method: 'post',
    data: { 
      approved, 
      rejected_reason: rejectedReason 
    },
  })
}

/**
 * 获取待审核用户数量 - 仅管理员
 */
export function getPendingCount() {
  return request({
    url: '/api/auth/pending-count',
    method: 'get',
  })
}

/**
 * 获取用户列表 - 仅管理员
 */
export function getUserList(role = null) {
  return request({
    url: '/api/auth/users',
    method: 'get',
    params: { role },
  })
}

/**
 * 更新用户信息 - 仅管理员
 */
export function updateUser(id, data) {
  return request({
    url: `/api/auth/users/${id}`,
    method: 'patch',
    data,
  })
}

/**
 * 修改密码
 */
export function changePassword(data) {
  return request({
    url: '/api/auth/change-password',
    method: 'post',
    data,
  })
}

/**
 * 重置用户密码 - 仅管理员
 */
export function resetUserPassword(userId) {
  return request({
    url: `/api/auth/users/${userId}/reset-password`,
    method: 'post',
  })
}

export function setUserSuperAdmin(userId, isSuperAdmin) {
  return request({
    url: `/api/auth/users/${userId}/super-admin`,
    method: 'post',
    data: { is_super_admin: isSuperAdmin },
  })
}

/**
 * 删除用户 - 仅管理员
 */
export function deleteUser(userId) {
  return request({
    url: `/api/auth/users/${userId}`,
    method: 'delete',
  })
}

/**
 * 更新个人信息 - 当前用户
 */
export function updateProfile(data) {
  return request({
    url: '/api/auth/me',
    method: 'patch',
    data,
  })
}

/**
 * 管理员绑定/解绑Telegram ID
 */
export function updateAdminTelegram(data) {
  return request({
    url: '/api/auth/me/telegram',
    method: 'patch',
    data,
  })
}

export function bindCurrentAdminTelegram(initData) {
  return request({
    url: '/api/auth/me/telegram/bind-current',
    method: 'post',
    data: { init_data: initData },
  })
}

export function sendPhoneVerificationCode(data) {
  return request({
    url: '/api/auth/phone-verification/send',
    method: 'post',
    data,
  })
}

export function verifyPhoneVerificationCode(data) {
  return request({
    url: '/api/auth/phone-verification/verify',
    method: 'post',
    data,
  })
}

export function verifyPhoneWithTelegram(data) {
  return request({
    url: '/api/auth/phone-verification/telegram-verify',
    method: 'post',
    data,
  })
}

/**
 * Telegram requestContact 一键关联手机号账号
 */
export function telegramContactLink(initData, contactData) {
  return request({
    url: '/api/auth/telegram-contact-link',
    method: 'post',
    data: { init_data: initData, contact_data: contactData },
  })
}

/**
 * Mini App 中用账号密码关联已有账号
 */
export function telegramLinkLogin(initData, username, password) {
  return request({
    url: '/api/auth/telegram-link-login',
    method: 'post',
    data: { init_data: initData, username, password },
  })
}

/**
 * 浏览器端：Telegram Login Widget 第三方登录
 */
export function telegramWidgetLogin(authData) {
  return request({
    url: '/api/auth/telegram-widget-login',
    method: 'post',
    data: authData,
    suppressError: true,
  })
}

/**
 * Bot 深链登录：创建 token，获取 bot_url
 */
export function botLoginCreate() {
  return request({
    url: '/api/auth/bot-login/create',
    method: 'post',
    suppressError: true,
  })
}

/**
 * Bot 深链登录：轮询确认结果
 */
export function botLoginVerify(token) {
  return request({
    url: `/api/auth/bot-login/verify?token=${encodeURIComponent(token)}`,
    method: 'get',
    suppressError: true,
  })
}

/**
 * 浏览器端：请求 Telegram 验证码
 */
export function requestLoginOTP(phone) {
  return request({
    url: '/api/auth/otp/request',
    method: 'post',
    data: { phone },
    suppressError: true,  // 由 Login.vue 自行处理 bot 引导逻辑
  })
}

/**
 * 浏览器端：验证码登录
 */
export function verifyLoginOTP(phone, code) {
  return request({
    url: '/api/auth/otp/verify',
    method: 'post',
    data: { phone, code },
  })
}

/**
 * 提交/重新提交资料审核
 */
export function submitForReview() {
  return request({
    url: '/api/auth/submit-review',
    method: 'post',
  })
}

// ============= 商品相关 =============

/**
 * 获取商品列表
 */
export function getProducts(params = {}) {
  return request({
    url: '/api/products',
    method: 'get',
    params,
  })
}

/**
 * 获取商品详情
 */
export function getProduct(id) {
  return request({
    url: `/api/products/${id}`,
    method: 'get',
  })
}

/**
 * 创建商品 - 仅管理员
 */
export function createProduct(data) {
  return request({
    url: '/api/products',
    method: 'post',
    data,
  })
}

/**
 * 更新商品 - 仅管理员
 */
export function updateProduct(id, data) {
  return request({
    url: `/api/products/${id}`,
    method: 'patch',
    data,
  })
}

/**
 * 删除商品 - 仅管理员
 */
export function deleteProduct(id) {
  return request({
    url: `/api/products/${id}`,
    method: 'delete',
  })
}

// ============= 订单相关 =============

/**
 * 获取订单列表
 */
export function getOrders(params = {}) {
  // 过滤掉空值参数，避免后端枚举校验失败
  const cleanParams = Object.fromEntries(
    Object.entries(params).filter(([, v]) => v !== '' && v !== null && v !== undefined)
  )
  return request({
    url: '/api/orders',
    method: 'get',
    params: cleanParams,
  })
}

/**
 * 获取订单详情
 */
export function getOrder(id) {
  return request({
    url: `/api/orders/${id}`,
    method: 'get',
  })
}

/**
 * 创建订单
 */
export function createOrder(data) {
  return request({
    url: '/api/orders',
    method: 'post',
    data,
  })
}

export function getDeliveryFeeSettings() {
  return request({
    url: '/api/settings/delivery-fee',
    method: 'get',
  })
}

export function updateDeliveryFeeSettings(data) {
  return request({
    url: '/api/settings/delivery-fee',
    method: 'patch',
    data,
  })
}

export function estimateDeliveryFee(distance_km) {
  return request({
    url: '/api/settings/delivery-fee/estimate',
    method: 'post',
    data: { distance_km },
  })
}

/** 获取配货员/配送员 Telegram chat id */
export function getRoleChatIds() {
  return request({
    url: '/api/settings/role-chat-ids',
    method: 'get',
  })
}

/** 设置配货员/配送员 Telegram chat id */
export function updateRoleChatIds(data) {
  return request({
    url: '/api/settings/role-chat-ids',
    method: 'put',
    data,
  })
}

/** 抓取近期与 Bot 交互过的 chat（用于管理员快速选择群） */
export function getTelegramRecentChats() {
  return request({
    url: '/api/settings/telegram-recent-chats',
    method: 'get',
  })
}

export function getContactInfo() {
  return request({
    url: '/api/settings/contact-info',
    method: 'get',
  })
}

export function updateContactInfo(data) {
  return request({
    url: '/api/settings/contact-info',
    method: 'patch',
    data,
  })
}

/**
 * 更新订单 - 仅管理员
 */
export function updateOrder(id, data) {
  return request({
    url: `/api/orders/${id}`,
    method: 'patch',
    data,
  })
}

/**
 * 取消订单 - 商户取消
 */
export function cancelOrder(id) {
  return request({
    url: `/api/orders/${id}/cancel`,
    method: 'post',
  })
}

// ============= 批量导入 / 条码 =============

/** 通过条码查询商品 */
export function getProductByBarcode(barcode) {
  return request({
    url: `/api/products/barcode/${encodeURIComponent(barcode)}`,
    method: 'get',
  })
}

/** 下载批量导入模板 URL（直接 a 链接打开） */
export function getProductImportTemplateUrl() {
  return '/api/products/import/template'
}

/** 上传 CSV 批量导入商品 */
export function importProducts(formData) {
  return request({
    url: '/api/products/import',
    method: 'post',
    data: formData,
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

/** 查询临近过期商品（管理员） */
export function listExpiringProducts(days = 30) {
  return request({
    url: '/api/products/expiring',
    method: 'get',
    params: { days },
  })
}

// ============= 配送费按地址估算 =============
export function estimateDeliveryFeeByAddress(origin, destination) {
  return request({
    url: '/api/settings/delivery-fee/estimate-by-address',
    method: 'post',
    data: { origin, destination },
  })
}

// ============= Google Maps 设置 =============
export function getGoogleMapsSettings() {
  return request({
    url: '/api/settings/google-maps',
    method: 'get',
  })
}

export function updateGoogleMapsSettings(data) {
  return request({
    url: '/api/settings/google-maps',
    method: 'patch',
    data,
  })
}

// ============= 配货 =============

/** 获取配货员视图：仅商品清单 */
export function getPickerItems(orderId) {
  return request({
    url: `/api/orders/picker/items/${orderId}`,
    method: 'get',
  })
}

/** 标记订单已配货 */
export function markOrderPicked(orderId) {
  return request({
    url: `/api/orders/${orderId}/pick`,
    method: 'post',
  })
}

// ============= 分类相关 =============

/**
 * 获取分类列表(活跃)
 */
export function getCategories() {
  return request({
    url: '/api/categories',
    method: 'get',
  })
}

/**
 * 获取所有分类(含禁用) - 管理员
 */
export function getAllCategories() {
  return request({
    url: '/api/categories/all',
    method: 'get',
  })
}

/**
 * 创建分类 - 管理员
 */
export function createCategory(data) {
  return request({
    url: '/api/categories',
    method: 'post',
    data,
  })
}

/**
 * 更新分类 - 管理员
 */
export function updateCategory(id, data) {
  return request({
    url: `/api/categories/${id}`,
    method: 'patch',
    data,
  })
}

/**
 * 删除分类 - 管理员
 */
export function deleteCategory(id) {
  return request({
    url: `/api/categories/${id}`,
    method: 'delete',
  })
}

// ============= 图片上传 =============

/**
 * 上传图片
 */
export function uploadImage(file) {
  const formData = new FormData()
  formData.append('file', file)
  return request({
    url: '/api/upload/image',
    method: 'post',
    data: formData,
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

// ============= 公告相关 =============

/**
 * 获取公开公告（无需认证）
 */
export function getPublicAnnouncements(type = null) {
  return request({
    url: '/api/announcements/public',
    method: 'get',
    params: type ? { type } : {},
  })
}

/**
 * 获取全部公告 - 管理员
 */
export function getAnnouncements(type = null) {
  return request({
    url: '/api/announcements',
    method: 'get',
    params: type ? { type } : {},
  })
}

/**
 * 创建公告 - 管理员
 */
export function createAnnouncement(data) {
  return request({
    url: '/api/announcements',
    method: 'post',
    data,
  })
}

/**
 * 更新公告 - 管理员
 */
export function updateAnnouncement(id, data) {
  return request({
    url: `/api/announcements/${id}`,
    method: 'patch',
    data,
  })
}

/**
 * 删除公告 - 管理员
 */
export function deleteAnnouncement(id) {
  return request({
    url: `/api/announcements/${id}`,
    method: 'delete',
  })
}

// ============= 月结账单相关 =============

/**
 * 获取月结账单列表
 */
export function getMonthlyBills(params = {}) {
  const cleanParams = Object.fromEntries(
    Object.entries(params).filter(([, v]) => v !== '' && v !== null && v !== undefined)
  )
  return request({
    url: '/api/billing',
    method: 'get',
    params: cleanParams,
  })
}

/**
 * 生成月结账单
 */
export function generateMonthlyBills(year, month) {
  return request({
    url: '/api/billing/generate',
    method: 'post',
    params: { year, month },
  })
}

/**
 * 更新月结账单
 */
export function updateMonthlyBill(id, data) {
  return request({
    url: `/api/billing/${id}`,
    method: 'patch',
    data,
  })
}
