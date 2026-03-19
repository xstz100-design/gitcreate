/**
 * 设备检测工具
 */

/**
 * 判断是否为移动设备
 * 结合 user-agent 和屏幕宽度综合判断
 */
export function isMobile() {
  const ua = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(
    navigator.userAgent
  )
  const smallScreen = (window.innerWidth || document.documentElement.clientWidth) < 768
  return ua || smallScreen
}

/**
 * 判断是否为触摸设备
 */
export function isTouchDevice() {
  return 'ontouchstart' in window || navigator.maxTouchPoints > 0
}

/**
 * 获取视口宽度
 */
export function getViewportWidth() {
  return window.innerWidth || document.documentElement.clientWidth
}

/**
 * 判断是否为小屏幕 (< 768px)
 */
export function isSmallScreen() {
  return getViewportWidth() < 768
}

/**
 * 触觉反馈 (仅移动端支持)
 */
export function hapticFeedback(style = 'light') {
  if ('vibrate' in navigator) {
    // 不同强度的震动
    const patterns = {
      light: [10],      // 轻微
      medium: [20],     // 中等
      heavy: [30],      // 强烈
      success: [10, 50, 10], // 成功
      error: [50, 100, 50],  // 错误
    }
    navigator.vibrate(patterns[style] || patterns.light)
  }
}
