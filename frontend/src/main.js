import { createApp } from 'vue'
import { createPinia } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'

// Vant UI - 移动端
import Vant from 'vant'
import 'vant/lib/index.css'
import '@vant/touch-emulator'

// i18n
import i18n, { getCurrentLanguage } from './i18n'

// 全局样式
import '@/styles/global.scss'

// Telegram Mini App
import { isTelegramMiniApp, getInitData, initBrowserViewportVars, initTelegramWebApp, getTelegramWebApp } from '@/utils/telegram'
import { telegramContactLink, telegramLinkLogin } from '@/api'

import App from './App.vue'
import router from './router'
import { useUserStore } from '@/stores/user'

const app = createApp(App)

const pinia = createPinia()
pinia.use(piniaPluginPersistedstate)
app.use(pinia)
app.use(router)
app.use(i18n)
app.use(ElementPlus, { locale: getCurrentLanguage() === 'en' ? en : zhCn })
app.use(Vant)

initBrowserViewportVars()

/**
 * 检查登录后的用户是否是未关联的自动创建账号（username=tg_xxx 且没有 phone）
 */
function isAutoCreatedAccount(user) {
  return user && user.username && user.username.startsWith('tg_') && !user.phone
}

/**
 * requestContact 一键授权：调用 Telegram 官方弹窗，获取用户手机号
 * 返回 Promise<contactData | null>
 */
function requestTelegramContact() {
  return new Promise((resolve) => {
    const webApp = getTelegramWebApp()
    if (!webApp || typeof webApp.requestContact !== 'function') {
      resolve(null)
      return
    }
    webApp.requestContact((shared, result) => {
      if (shared && result) {
        resolve(result)
      } else {
        resolve(null)
      }
    })
  })
}

/**
 * 弹出账号关联底部弹窗
 * 优先使用 requestContact（一键），失败则引导手动输入账号密码
 */
async function promptAccountLink(userStore) {
  const webApp = getTelegramWebApp()
  const initData = getInitData()

  // 方式一：尝试 requestContact（Telegram 原生一键授权）
  const hasRequestContact = webApp && typeof webApp.requestContact === 'function'

  if (hasRequestContact) {
    // 弹出 Telegram 原生共享手机号弹窗
    const contactResult = await requestTelegramContact()
    if (contactResult) {
      try {
        const data = await telegramContactLink(initData, contactResult)
        userStore.token = data.access_token
        userStore.userInfo = data.user
        return true
      } catch (e) {
        console.warn('[TG关联] requestContact 关联失败，降级到密码关联:', e)
      }
    }
  }

  // 方式二：降级 — 弹出 Vant Dialog 让用户输入账号密码
  const { showDialog, closeDialog, showToast } = await import('vant')

  return new Promise((resolve) => {
    // 构造表单容器
    const container = document.createElement('div')
    container.style.cssText = 'padding:16px 0'
    const phoneInput = document.createElement('input')
    phoneInput.placeholder = '手机号（账号）'
    phoneInput.style.cssText = 'width:100%;padding:10px;margin-bottom:10px;border:1px solid #eee;border-radius:8px;font-size:15px;box-sizing:border-box'
    const pwdInput = document.createElement('input')
    pwdInput.type = 'password'
    pwdInput.placeholder = '登录密码'
    pwdInput.style.cssText = 'width:100%;padding:10px;border:1px solid #eee;border-radius:8px;font-size:15px;box-sizing:border-box'
    container.appendChild(phoneInput)
    container.appendChild(pwdInput)

    showDialog({
      title: '关联已有账号',
      message: container,
      confirmButtonText: '确认关联',
      cancelButtonText: '暂不关联',
      showCancelButton: true,
    }).then(async () => {
      const username = phoneInput.value.trim()
      const password = pwdInput.value
      if (!username || !password) {
        showToast('请填写账号和密码')
        resolve(false)
        return
      }
      try {
        const data = await telegramLinkLogin(initData, username, password)
        userStore.token = data.access_token
        userStore.userInfo = data.user
        resolve(true)
      } catch (e) {
        showToast(e?.response?.data?.detail || '关联失败，请检查账号密码')
        resolve(false)
      }
    }).catch(() => {
      resolve(false)
    })
  })
}

// Telegram Mini App 自动登录
async function bootstrapTelegram() {
  if (!isTelegramMiniApp()) return false

  initTelegramWebApp()

  const userStore = useUserStore()
  // 已有 token 则验证 token 有效性
  if (userStore.isLoggedIn) {
    try {
      await userStore.fetchUserInfo()
      // 已登录但仍是自动创建账号，再次提示关联（可选）
      if (isAutoCreatedAccount(userStore.userInfo)) {
        await promptAccountLink(userStore)
      }
      return true
    } catch {
      // token 失效，重新登录
      userStore.logout()
    }
  }

  try {
    await userStore.telegramLogin(getInitData())

    // 登录成功后检查是否需要关联已有账号
    if (isAutoCreatedAccount(userStore.userInfo)) {
      await promptAccountLink(userStore)
    }

    return true
  } catch (e) {
    console.error('Telegram 自动登录失败:', e)
    return false
  }
}

async function syncTelegramEntryRoute() {
  if (!isTelegramMiniApp()) return

  const userStore = useUserStore()
  const currentPath = router.currentRoute.value.path

  // 管理员和商户在 Telegram miniapp 中均使用移动端界面
  if (userStore.isLoggedIn) {
    if (currentPath === '/' || currentPath === '/login' || currentPath.startsWith('/admin')) {
      await router.replace('/m/shop')
    }
  }
}

bootstrapTelegram().finally(async () => {
  await syncTelegramEntryRoute()
  app.mount('#app')
})
