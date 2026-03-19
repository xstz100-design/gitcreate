<template>
  <div class="login-container">
    <button class="login-lang-btn" @click="toggleLang">{{ currentLang === 'zh' ? 'English' : '中文' }}</button>
    <div class="login-card">
      <div class="login-header">
        <div class="logo-container">
          <img src="/images/logo1.svg" alt="Logo" class="logo-image" />
        </div>
      </div>
      
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
      
      <div class="login-footer">
        <span class="forgot-link" @click="showForgotPassword">{{ $t('login.forgotPassword') }}</span>
      </div>
    </div>

    <!-- 忘记密码弹窗 -->
    <el-dialog v-model="forgotVisible" :title="$t('login.forgotPassword')" width="380px" center>
      <div class="forgot-content">
        <p class="forgot-tip">{{ $t('login.forgotTip') }}</p>
        <div class="forgot-contact" v-if="contactInfo">
          <div v-html="contactInfo"></div>
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
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus/es/components/message/index'
import { useUserStore } from '@/stores/user'
import { useI18n } from 'vue-i18n'
import { setLanguage, getCurrentLanguage } from '@/i18n'
import { changePassword, getPublicAnnouncements } from '@/api'

const router = useRouter()
const userStore = useUserStore()
const { t } = useI18n()
const currentLang = ref(getCurrentLanguage())

const toggleLang = () => {
  const newLang = currentLang.value === 'zh' ? 'en' : 'zh'
  setLanguage(newLang)
  currentLang.value = newLang
}

const formRef = ref()
const loading = ref(false)
const forgotVisible = ref(false)
const contactInfo = ref('')

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
  } else if (userStore.isMerchant) {
    router.push('/merchant')
  } else {
    router.push('/')
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
      contactInfo.value = content.replace(/\n/g, '<br>')
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
    margin-bottom: 20px;
  }
  
  .logo-image {
    width: 280px;
    max-width: 100%;
    height: auto;
    object-fit: contain;
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

.forgot-link {
  color: #409eff;
  font-size: 13px;
  cursor: pointer;
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
