<template>
  <div class="login-container" :class="{ 'tg-fullscreen': isTgContext }">

    <!-- ══ Telegram 内：全屏自动登录 ══ -->
    <div v-if="isTgContext" class="tg-autoscreen">
      <div v-if="tgAutoLoading" class="tg-auto-body">
        <div class="tg-spinner"></div>
        <p>正在登录…</p>
      </div>
      <div v-else-if="tgAutoError" class="tg-auto-body tg-auto-err">
        <p>{{ tgAutoError }}</p>
        <button class="tg-retry-btn" @click="handleTgAutoLogin">重试</button>
      </div>
    </div>

    <!-- ══ 外部浏览器：登录卡片 ══ -->
    <template v-else>
      <button class="login-lang-btn" @click="toggleLang">{{ langLabel }}</button>
      <div class="login-card">
        <div class="login-header">
          <div class="logo-container">
          </div>
        </div>

        <!-- 管理员账号密码登录（默认隐藏，点「管理员入口」展开） -->
        <div v-if="showAdminForm">
          <el-form :model="form" :rules="rules" ref="formRef" class="login-form">
            <el-form-item prop="username">
              <el-input
                v-model="form.username"
                :placeholder="$t('login.accountPlaceholder')"
                size="large"
                clearable
              />
            </el-form-item>
            <el-form-item prop="password">
              <el-input
                v-model="form.password"
                type="password"
                :placeholder="$t('login.passwordPlaceholder')"
                size="large"
                show-password
                @keyup.enter="handleLogin"
              />
            </el-form-item>
            <el-form-item>
              <el-button
                type="primary"
                :loading="loading"
                size="large"
                style="width: 100%"
                @click="handleLogin"
              >
                {{ $t('login.login') }}
              </el-button>
            </el-form-item>
          </el-form>
          <div class="admin-back" @click="showAdminForm = false">← 返回 Telegram 登录</div>
        </div>

        <!-- Telegram Bot 深链登录（默认） -->
        <div v-else class="tg-login-section">
          <div v-if="!tgPolling">
            <p class="tg-browser-tip">在 Telegram 中打开即可自动登录注册</p>
            <div style="display:flex;justify-content:center;margin-bottom:8px">
              <button class="tg-btn" :disabled="tgLoading" @click="handleTgLogin">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="#fff"><path d="M12 0C5.373 0 0 5.373 0 12s5.373 12 12 12 12-5.373 12-12S18.627 0 12 0zm5.894 8.221l-1.97 9.28c-.145.658-.537.818-1.084.508l-3-2.21-1.447 1.394c-.16.16-.295.295-.605.295l.213-3.053 5.56-5.023c.242-.213-.054-.333-.373-.12L7.17 13.857l-2.96-.924c-.643-.204-.657-.643.136-.953l11.57-4.461c.537-.194 1.006.131.978.702z"/></svg>
                <span>{{ tgLoading ? '正在生成链接…' : '用 Telegram 登录' }}</span>
              </button>
            </div>
          </div>
          <div v-else class="tg-waiting">
            <div class="tg-spinner"></div>
            <p>已打开 Telegram，请点击 <b>START</b> 确认登录</p>
            <p class="tg-countdown">等待中（{{ tgCountdown }}秒后过期）</p>
            <button class="tg-cancel-btn" @click="cancelTgLogin">取消</button>
          </div>
          <div v-if="tgError" class="tg-error">{{ tgError }}</div>
        </div>

        <div class="login-footer">
          <div class="footer-links">
            <span class="forgot-link" @click="showForgotPassword">{{ $t('login.forgotPassword') }}</span>
            <span class="admin-entry" @click="showAdminForm = true">管理员入口</span>
          </div>
        </div>
      </div>
    </template>

    <!-- 忘记密码弹窗 -->
    <el-dialog v-model="forgotVisible" :title="$t('login.forgotPassword')" width="380px" center>
      <div class="forgot-content">
        <p class="forgot-tip">{{ $t('login.forgotTip') }}</p>
        <div class="forgot-contact" v-if="contactInfo">
          <p v-for="(line, idx) in contactLines" :key="idx">{{ line }}</p>
        </div>
        <div v-else class="forgot-contact">
          <p>{{ $t('login.contactAdmin') }}</p>
        </div>
      </div>
      <template #footer>
        <el-button type="primary" @click="forgotVisible = false">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- 首次登录强制改密弹窗 -->
    <el-dialog v-model="changePasswordVisible" :title="$t('login.mustChangePassword')" width="400px" :close-on-click-modal="false" :close-on-press-escape="false" :show-close="false" center>
      <div class="change-pwd-content">
        <p class="change-pwd-tip">{{ $t('login.mustChangePasswordTip') }}</p>
        <el-form :model="pwdForm" :rules="pwdRules" ref="pwdFormRef">
          <el-form-item prop="new_password">
            <el-input v-model="pwdForm.new_password" type="password" :placeholder="$t('profile.newPasswordPlaceholder')" size="large" show-password />
          </el-form-item>
          <el-form-item prop="confirm_password">
            <el-input v-model="pwdForm.confirm_password" type="password" :placeholder="$t('profile.confirmPasswordPlaceholder')" size="large" show-password @keyup.enter="handleChangePassword" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button type="primary" :loading="changingPassword" style="width: 100%;" @click="handleChangePassword">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus/es/components/message/index'
import { useUserStore } from '@/stores/user'
import { useI18n } from 'vue-i18n'
import { setLanguage, getCurrentLanguage } from '@/i18n'
import { changePassword, getPublicAnnouncements, telegramAuth, botLoginCreate, botLoginVerify } from '@/api'

const TG_BOT_USERNAME = 'testshopmouy_bot'

const router = useRouter()
const userStore = useUserStore()
const { t } = useI18n()
const currentLang = ref(getCurrentLanguage())

const toggleLang = () => {
  const order = ['zh', 'en', 'kh']
  const idx = order.indexOf(currentLang.value)
  const newLang = order[(idx + 1) % order.length]
  setLanguage(newLang)
  currentLang.value = newLang
}

const langLabel = computed(() => {
  const next = { zh: 'English', en: 'ខ្មែរ', kh: '中文' }
  return next[currentLang.value] || 'English'
})

const showAdminForm = ref(false)
const formRef = ref()

// 检测是否在 Telegram 环境内（任何设备）
const isTgContext = typeof window !== 'undefined' && (
  !!(window.Telegram?.WebApp?.initData) ||
  !!(window.TelegramWebviewProxy) ||
  /Telegram/i.test(navigator.userAgent)
)

// —— TG 环境：自动免登录 ——
const tgAutoLoading = ref(false)
const tgAutoError = ref('')

const handleTgAutoLogin = async () => {
  const initData = window.Telegram?.WebApp?.initData
  if (!initData) {
    tgAutoError.value = '无法获取 Telegram 身份信息，请尝试关闭并重新打开'
    return
  }
  tgAutoLoading.value = true
  tgAutoError.value = ''
  try {
    const data = await telegramAuth(initData)
    userStore.token = data.access_token
    userStore.userInfo = data.user
    await new Promise(resolve => setTimeout(resolve, 100))
    ElMessage.success(data.is_new_user ? '账号已自动创建，欢迎加入！' : t('login.loginSuccess'))
    navigateByRole()
  } catch (e) {
    const msg = e?.response?.data?.detail || e?.response?.data?.message
    if (e?.response?.status === 403) {
      tgAutoError.value = '账号已被禁用，请联系管理员'
    } else {
      tgAutoError.value = msg || '自动登录失败，请重试'
    }
  } finally {
    tgAutoLoading.value = false
  }
}

// —— 外部浏览器：Bot 深链登录（带轮询） ——
const tgLoading = ref(false)
const tgPolling = ref(false)
const tgCountdown = ref(300)
const tgError = ref('')
let tgPollTimer = null
let tgCountdownTimer = null
let tgCurrentToken = ''

const stopPolling = () => {
  clearInterval(tgPollTimer)
  clearInterval(tgCountdownTimer)
  tgPollTimer = null
  tgCountdownTimer = null
}

const cancelTgLogin = () => {
  stopPolling()
  tgPolling.value = false
  tgCurrentToken = ''
  tgCountdown.value = 300
}

const handleTgLogin = async () => {
  tgError.value = ''
  tgLoading.value = true
  try {
    const data = await botLoginCreate()
    tgCurrentToken = data.token
    window.open(data.bot_url, '_blank')
    tgPolling.value = true
    tgCountdown.value = 300
    tgPollTimer = setInterval(async () => {
      try {
        const result = await botLoginVerify(tgCurrentToken)
        if (result && result.access_token) {
          stopPolling()
          tgPolling.value = false
          userStore.token = result.access_token
          userStore.userInfo = result.user
          await new Promise(resolve => setTimeout(resolve, 100))
          ElMessage.success(t('login.loginSuccess'))
          navigateByRole()
        }
      } catch (e) {
        const status = e?.response?.status
        if (status === 410 || status === 404) {
          stopPolling(); tgPolling.value = false
          tgError.value = '登录链接已过期，请重新点击登录'
        } else if (status === 403) {
          stopPolling(); tgPolling.value = false
          tgError.value = '账号已被禁用，请联系管理员'
        }
      }
    }, 2000)
    tgCountdownTimer = setInterval(() => {
      tgCountdown.value--
      if (tgCountdown.value <= 0) {
        stopPolling(); tgPolling.value = false
        tgError.value = '等待超时，请重新点击登录'
      }
    }, 1000)
  } catch {
    tgError.value = '生成登录链接失败，请重试'
  } finally {
    tgLoading.value = false
  }
}

onMounted(() => {
  // 在 Telegram 环境内就自动登录
  if (isTgContext) {
    handleTgAutoLogin()
  }
})

onUnmounted(() => {
  stopPolling()
})
const loading = ref(false)
const forgotVisible = ref(false)


const contactInfo = ref('')
// 修复 XSS: 不再用 v-html，改为按行拆分渲染纯文本
const contactLines = computed(() => contactInfo.value ? contactInfo.value.split('\n') : [])

const form = reactive({
  username: '',
  password: '',
})

const rules = {
  username: [{ required: true, message: () => t('login.accountRequired'), trigger: 'blur' }],
  password: [{ required: true, message: () => t('login.passwordRequired'), trigger: 'blur' }],
}

// 强制改密
const changePasswordVisible = ref(false)
const changingPassword = ref(false)
const pwdFormRef = ref()
const pwdForm = reactive({
  new_password: '',
  confirm_password: '',
})

const pwdRules = {
  new_password: [
    { required: true, message: () => t('profile.newPasswordPlaceholder'), trigger: 'blur' },
    { min: 6, message: () => t('profile.passwordMinLength'), trigger: 'blur' },
  ],
  confirm_password: [
    { required: true, message: () => t('profile.confirmPasswordPlaceholder'), trigger: 'blur' },
    { validator: (rule, value, callback) => {
      if (value !== pwdForm.new_password) {
        callback(new Error(t('profile.passwordMismatch')))
      } else {
        callback()
      }
    }, trigger: 'blur' },
  ],
}

let tempOldPassword = ''

const navigateByRole = () => {
  if (userStore.isAdmin) {
    router.push('/admin')
  } else {
    router.push('/m/shop')
  }
}

const handleLogin = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  
  loading.value = true
  try {
    const data = await userStore.login(form.username, form.password)
    
    // 等待持久化写入
    await new Promise(resolve => setTimeout(resolve, 100))
    
    // 检查首次登录强制改密
    if (data.user?.must_change_password) {
      tempOldPassword = form.password
      changePasswordVisible.value = true
      return
    }
    
    ElMessage.success(t('login.loginSuccess'))
    navigateByRole()
  } catch (error) {
    console.error('[登录] 登录失败:', error)
  } finally {
    loading.value = false
  }
}

const handleChangePassword = async () => {
  if (!pwdFormRef.value) return
  try {
    await pwdFormRef.value.validate()
  } catch {
    return
  }
  
  changingPassword.value = true
  try {
    await changePassword({
      old_password: tempOldPassword,
      new_password: pwdForm.new_password,
    })
    changePasswordVisible.value = false
    if (userStore.userInfo) {
      userStore.userInfo.must_change_password = false
    }
    ElMessage.success(t('profile.passwordChanged'))
    navigateByRole()
  } catch (error) {
    ElMessage.error(t('profile.updateFailed'))
  } finally {
    changingPassword.value = false
  }
}

// 忘记密码
const showForgotPassword = async () => {
  try {
    const data = await getPublicAnnouncements('contact')
    if (data && data.length > 0) {
      const item = data[0]
      const content = currentLang.value === 'zh' ? item.content_zh : item.content_en
      // 修复 XSS: 不再将内容转为 HTML，保留原始文本
      contactInfo.value = content
    }
  } catch {
    contactInfo.value = ''
  }
  forgotVisible.value = true
}
</script>

<style scoped lang="scss">
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: #F8F9FA;
  padding: 24px;
  position: relative;
}

.login-lang-btn {
  position: absolute;
  top: 20px;
  right: 20px;
  padding: 6px 16px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  background: #fff;
  color: #333;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  z-index: 10;
}

.login-tabs {
  margin-bottom: 4px;
  :deep(.el-tabs__nav-wrap::after) {
    height: 1px;
  }
}

/* ── Telegram 全屏自动登录 ── */
.tg-fullscreen {
  background: linear-gradient(160deg, #0d1b2a 0%, #1b2d3e 45%, #0f3460 100%);
  min-height: 100vh;
  justify-content: center;
  align-items: center;
}

.tg-autoscreen {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 24px;
  padding: 40px 20px;
}

.tg-auto-logo {
  height: 60px;
  filter: brightness(0) invert(1);
  opacity: 0.9;
}

.tg-auto-body {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
  color: rgba(255,255,255,0.85);
  font-size: 15px;
}

.tg-auto-err {
  color: #ff7875;
}

/* ── 管理员密码表单（外部浏览器） ── */
.admin-back {
  text-align: center;
  font-size: 13px;
  color: #409eff;
  cursor: pointer;
  margin-top: 4px;
  &:hover { text-decoration: underline; }
}

.admin-entry {
  font-size: 12px;
  color: #bbb;
  cursor: pointer;
  margin-left: 12px;
  &:hover { color: #666; }
}

/* ── Telegram 登录区域 ── */
.tg-login-section {

  padding: 12px 0 8px;
}

.tg-auto {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 16px 0;
  gap: 12px;
}

.tg-auto-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  color: #666;
  font-size: 14px;
}

.tg-auto-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  color: #f56c6c;
  font-size: 13px;
  text-align: center;
}

.tg-retry-btn {
  padding: 6px 20px;
  background: #409eff;
  color: #fff;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  &:hover { opacity: 0.88; }
}

.tg-browser-tip {
  text-align: center;
  font-size: 13px;
  color: #666;
  margin-bottom: 14px;
}

.tg-btn {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  background: #229ED9;
  color: #fff;
  border: none;
  border-radius: 10px;
  padding: 12px 28px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s;
  &:hover { opacity: 0.88; }
  &:disabled { opacity: 0.6; cursor: not-allowed; }
}

.tg-waiting {
  text-align: center;
  color: #555;
  font-size: 14px;
  line-height: 1.8;
}

.tg-countdown {
  font-size: 12px;
  color: #aaa;
  margin-top: 2px;
}

.tg-cancel-btn {
  margin-top: 8px;
  background: none;
  border: 1px solid #ddd;
  border-radius: 6px;
  padding: 4px 16px;
  font-size: 13px;
  color: #999;
  cursor: pointer;
  &:hover { color: #666; }
}

.tg-error {
  color: #f56c6c;
  font-size: 13px;
  text-align: center;
  margin-top: 10px;
}

.tg-spinner {
  display: inline-block;
  width: 30px;
  height: 30px;
  border: 3px solid #e0e0e0;
  border-top-color: #229ED9;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }

.login-lang-btn:hover {
  border-color: #1D4ED8;
  color: #1D4ED8;
}

.login-card {
  width: 100%;
  max-width: 420px;
  background-color: #FFFFFF;
  border: 1px solid #E8E8E8;
  border-radius: 4px;
  padding: 48px 40px;
}

.login-header {
  text-align: center;
  margin-bottom: 40px;
  
  .logo-container {
    display: flex;
    justify-content: center;
    align-items: center;
    margin: -48px -40px 20px;
    padding: 24px 20px;
    background: #1d4ed8;
    border-radius: 4px 4px 0 0;
  }
  
  .logo-image {
    width: 240px;
    max-width: 100%;
    height: auto;
    object-fit: contain;
    filter: brightness(0) invert(1);
  }
}

.login-form {
  :deep(.el-form-item) {
    margin-bottom: 24px;
    
    &:last-child {
      margin-bottom: 0;
      margin-top: 32px;
    }
  }
  
  :deep(.el-input__wrapper) {
    height: 48px;
    box-shadow: none;
    border: 1px solid #E8E8E8;
    border-radius: 2px;
    transition: border-color 200ms cubic-bezier(0.4, 0.0, 0.2, 1);
    
    &:hover {
      border-color: #D9D9D9;
    }
    
    &.is-focus {
      border-color: #1D4ED8;
    }
  }
  
  :deep(.el-input__inner) {
    font-size: 14px;
    color: #1A1A1A;
    
    &::placeholder {
      color: #BFBFBF;
    }
  }
  
  :deep(.el-button--large) {
    height: 48px;
    font-size: 15px;
    font-weight: 500;
    letter-spacing: 0.3px;
    border-radius: 2px;
    
    &:active {
      transform: scale(0.98);
      transition: transform 150ms cubic-bezier(0.0, 0.0, 0.2, 1);
    }
  }
}

.login-footer {
  margin-top: 24px;
  text-align: center;
  padding-top: 24px;
  border-top: 1px solid #F0F0F0;
}

.footer-links {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 12px;
}

.divider {
  color: #dcdfe6;
}

.forgot-link {
  color: #409eff;
  font-size: 13px;
  cursor: pointer;
  &:hover {
    text-decoration: underline;
  }
}

.register-link {
  color: #409eff;
  font-size: 13px;
  text-decoration: none;
  &:hover {
    text-decoration: underline;
  }
}

.forgot-content {
  text-align: center;
}

.forgot-tip {
  color: #606266;
  font-size: 14px;
  margin-bottom: 16px;
}

.forgot-contact {
  padding: 16px;
  background: #f5f7fa;
  border-radius: 6px;
  color: #303133;
  font-size: 14px;
  line-height: 1.8;
}

.change-pwd-content {
  .change-pwd-tip {
    color: #E6A23C;
    font-size: 14px;
    margin-bottom: 20px;
    text-align: center;
  }
}

// 响应式
@media (max-width: 768px) {
  .login-card {
    padding: 32px 24px;
  }
  
  .login-header h1 {
    font-size: 20px;
  }
}
</style>
