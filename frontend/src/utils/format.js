import i18n from '@/i18n'

/**
 * 美元转瑞尔
 */
export function usdToKhr(usd, rate = 4000) {
  const val = Number(usd) * rate
  return isNaN(val) ? 0 : Math.round(val)
}

/**
 * 瑞尔转美元
 */
export function khrToUsd(khr, rate = 4000) {
  const val = Number(khr) / rate
  return isNaN(val) ? '0.00' : val.toFixed(2)
}

/**
 * 格式化货币 (USD)
 */
export function formatUSD(amount) {
  const val = Number(amount)
  return `$${isNaN(val) ? '0.00' : val.toFixed(2)}`
}

/**
 * 格式化货币 (KHR)
 */
export function formatKHR(amount) {
  const val = Number(amount)
  if (isNaN(val) || val === 0) return '0 ៛'
  return `${val.toLocaleString()} ៛`
}

/**
 * 格式化日期时间
 */
export function formatDateTime(dateString) {
  const date = new Date(dateString)
  const lang = i18n.global.locale.value === 'en' ? 'en-US' : 'zh-CN'
  return date.toLocaleString(lang, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

/**
 * 支付状态文本
 */
export function getPaymentStatusText(status) {
  const t = i18n.global.t
  const key = (status || '').toLowerCase()
  const map = {
    unpaid: t('order.unpaid'),
    cash: t('order.cash'),
    monthly: t('order.monthly'),
  }
  return map[key] || status
}

/**
 * 配送状态文本
 */
export function getDeliveryStatusText(status) {
  const t = i18n.global.t
  const key = (status || '').toLowerCase()
  const map = {
    pending: t('order.deliveryPending'),
    delivering: t('order.delivering'),
    delivered: t('order.delivered'),
    cancelled: t('order.cancelled'),
  }
  return map[key] || status
}

/**
 * 用户角色文本
 */
export function getRoleText(role) {
  const t = i18n.global.t
  const map = {
    admin: t('role.admin'),
    merchant: t('role.merchant'),
  }
  return map[role] || role
}
