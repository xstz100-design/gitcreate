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
 * 获取当前用户信息
 */
export function getCurrentUser() {
  return request({
    url: '/api/auth/me',
    method: 'get',
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
    url: '/api/upload',
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
