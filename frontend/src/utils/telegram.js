/**
 * Telegram Mini App 工具模块
 */

/**
 * 获取 Telegram WebApp 实例
 */
export function getTelegramWebApp() {
  return window.Telegram?.WebApp || null
}

function setTelegramCssVar(name, value) {
  if (value === undefined || value === null || Number.isNaN(value)) return
  document.documentElement.style.setProperty(name, `${Math.max(0, Math.round(value))}px`)
}

function getBrowserViewportHeight() {
  return window.visualViewport?.height || window.innerHeight || document.documentElement.clientHeight || 0
}

function syncBrowserViewportVars() {
  const browserHeight = getBrowserViewportHeight()
  const stableHeight = Math.max(browserHeight, window.innerHeight || 0, document.documentElement.clientHeight || 0)

  setTelegramCssVar('--tg-viewport-height', browserHeight)
  setTelegramCssVar('--tg-viewport-stable-height', stableHeight)
}

let browserViewportInited = false

export function initBrowserViewportVars() {
  if (typeof window === 'undefined') return

  syncBrowserViewportVars()

  if (browserViewportInited) return
  browserViewportInited = true

  const handleViewportChange = () => syncBrowserViewportVars()

  window.addEventListener('resize', handleViewportChange, { passive: true })
  window.addEventListener('orientationchange', handleViewportChange, { passive: true })
  window.visualViewport?.addEventListener('resize', handleViewportChange, { passive: true })
  window.visualViewport?.addEventListener('scroll', handleViewportChange, { passive: true })
}

function syncTelegramViewportVars(webApp) {
  const browserHeight = getBrowserViewportHeight()
  setTelegramCssVar('--tg-viewport-height', webApp.viewportHeight || browserHeight)
  setTelegramCssVar('--tg-viewport-stable-height', webApp.viewportStableHeight || browserHeight)

  const safeArea = webApp.safeAreaInset || {}
  setTelegramCssVar('--tg-safe-area-inset-top', safeArea.top || 0)
  setTelegramCssVar('--tg-safe-area-inset-bottom', safeArea.bottom || 0)
  setTelegramCssVar('--tg-safe-area-inset-left', safeArea.left || 0)
  setTelegramCssVar('--tg-safe-area-inset-right', safeArea.right || 0)

  const contentSafeArea = webApp.contentSafeAreaInset || {}
  setTelegramCssVar('--tg-content-safe-area-inset-top', contentSafeArea.top || 0)
  setTelegramCssVar('--tg-content-safe-area-inset-bottom', contentSafeArea.bottom || 0)
  setTelegramCssVar('--tg-content-safe-area-inset-left', contentSafeArea.left || 0)
  setTelegramCssVar('--tg-content-safe-area-inset-right', contentSafeArea.right || 0)
}

function syncTelegramTheme(webApp) {
  const theme = webApp.themeParams || {}
  const isDark = (webApp.colorScheme || '').toLowerCase() === 'dark'

  // 应用 UI 已统一为白底 + 渐变细线（暗色模式时改用 Telegram 主题底色）
  const appBg = isDark ? (theme.bg_color || '#1f2024') : '#ffffff'
  const background = isDark ? (theme.bg_color || '#1f2024') : '#f8f9fa'
  const header = appBg
  const bottomBar = isDark ? (theme.bottom_bar_bg_color || theme.secondary_bg_color || appBg) : '#ffffff'

  // 同步 data-theme 给 CSS 暗色变量切换
  document.documentElement.dataset.theme = isDark ? 'dark' : 'light'

  webApp.setBackgroundColor(background)

  if (webApp.setHeaderColor) {
    webApp.setHeaderColor(header)
  }

  if (webApp.setBottomBarColor) {
    webApp.setBottomBarColor(bottomBar)
  }
}

/**
 * 判断当前是否在 Telegram Mini App 环境中运行
 */
export function isTelegramMiniApp() {
  const webApp = getTelegramWebApp()
  return !!(webApp && webApp.initData && webApp.initData.length > 0)
}

/**
 * 获取 initData 字符串（用于发送给后端验证）
 */
export function getInitData() {
  const webApp = getTelegramWebApp()
  return webApp?.initData || ''
}

/**
 * 获取 Telegram 用户信息
 */
export function getTelegramUser() {
  const webApp = getTelegramWebApp()
  return webApp?.initDataUnsafe?.user || null
}

/**
 * 初始化 Telegram WebApp（设置主题、展开等）
 */
export function initTelegramWebApp() {
  const webApp = getTelegramWebApp()
  if (!webApp) return

  initBrowserViewportVars()

  // 通知 Telegram 客户端应用已准备好
  webApp.ready()

  // 展开到全屏
  webApp.expand()

  syncTelegramViewportVars(webApp)
  syncTelegramTheme(webApp)

  webApp.onEvent('viewportChanged', () => syncTelegramViewportVars(webApp))
  webApp.onEvent('safeAreaChanged', () => syncTelegramViewportVars(webApp))
  webApp.onEvent('contentSafeAreaChanged', () => syncTelegramViewportVars(webApp))
  webApp.onEvent('themeChanged', () => syncTelegramTheme(webApp))
}
