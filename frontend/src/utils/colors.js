/**
 * 统一色彩系统 - 语义化设计
 */

export const colors = {
  // 主色 - 深蓝，代表稳重批发
  primary: '#2F54EB',
  primaryLight: '#597EF7',
  primaryDark: '#1D39C4',
  
  // 成功/有货 - 绿色  
  success: '#52C41A',
  successLight: '#73D13D',
  successDark: '#389E0D',
  
  // 警告/库存不足 - 橙色
  warning: '#FAAD14',
  warningLight: '#FFC53D',
  warningDark: '#D48806',
  
  // 危险/欠款/缺货 - 红色
  danger: '#F5222D',
  dangerLight: '#FF4D4F',
  dangerDark: '#CF1322',
  
  // 中性色
  text: '#262626',
  textSecondary: '#8C8C8C',
  textDisabled: '#BFBFBF',
  border: '#D9D9D9',
  background: '#F5F5F5',
  white: '#FFFFFF',
}

/**
 * 获取支付状态颜色
 */
export function getPaymentStatusColor(status) {
  const colorMap = {
    pending: colors.warning,
    paid: colors.success,
    credit: colors.primary,
  }
  return colorMap[status] || colors.textSecondary
}

/**
 * 获取配送状态颜色
 */
export function getDeliveryStatusColor(status) {
  const colorMap = {
    pending: colors.textSecondary,
    delivering: colors.warning,
    delivered: colors.success,
    cancelled: colors.danger,
  }
  return colorMap[status] || colors.textSecondary
}

/**
 * 获取库存状态颜色
 */
export function getStockStatusColor(stock, warning = 10) {
  if (stock === 0) return colors.danger
  if (stock <= warning) return colors.warning
  return colors.success
}
