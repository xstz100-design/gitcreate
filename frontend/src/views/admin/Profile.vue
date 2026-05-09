<template>
  <div class="profile-page">
    <h2>{{ $t('admin.myProfile') }}</h2>

    <!-- 个人信息 -->
    <el-card class="info-card">
      <template #header>
        <div class="card-header">
          <span>{{ $t('profile.personalInfo') }}</span>
          <el-button type="primary" link size="small" @click="openEditName">{{ $t('common.edit') }}</el-button>
        </div>
      </template>
      <div class="info-rows">
        <div class="info-row">
          <span class="info-label">{{ $t('login.username') }}</span>
          <span class="info-val">{{ userStore.userInfo?.username }}</span>
        </div>
        <div class="info-row">
          <span class="info-label">{{ $t('profile.name') }}</span>
          <span class="info-val">{{ userStore.userInfo?.full_name || $t('profile.notSet') }}</span>
        </div>
        <div class="info-row">
          <span class="info-label">{{ $t('profile.phone') }}</span>
          <span class="info-val">{{ userStore.userInfo?.phone || $t('profile.notSet') }}</span>
        </div>
      </div>
    </el-card>

    <!-- 账户安全 -->
    <el-card class="info-card">
      <template #header>
        <div class="card-header">
          <span>{{ $t('profile.accountSecurity') }}</span>
        </div>
      </template>
      <el-form :label-width="mobile ? '0' : '120px'" :label-position="mobile ? 'top' : 'right'">
        <el-form-item :label="$t('profile.changePassword')">
          <el-button type="primary" @click="showPasswordDialog = true">
            {{ $t('profile.changePassword') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Telegram 通知设置 -->
    <el-card class="info-card">
      <template #header>
        <div class="card-header">
          <span>{{ $t('admin.telegramSettings') }}</span>
        </div>
      </template>
      <el-form :label-width="mobile ? '0' : '120px'" :label-position="mobile ? 'top' : 'right'">
        <el-form-item :label="$t('admin.telegramId')">
          <div class="editable-field">
            <span class="info-value">{{ userStore.userInfo?.telegram_id || $t('profile.notSet') }}</span>
            <el-button type="primary" link size="small" @click="openEditTelegram">{{ $t('common.edit') }}</el-button>
          </div>
        </el-form-item>
        <el-form-item v-if="telegramMiniAppAvailable" :label="$t('admin.bindCurrentTelegram')">
          <el-button type="primary" :loading="bindingTelegram" @click="bindCurrentTelegram">
            {{ userStore.userInfo?.telegram_id ? $t('admin.rebindCurrentTelegram') : $t('admin.bindCurrentTelegram') }}
          </el-button>
        </el-form-item>
        <div class="notify-tip">{{ telegramMiniAppAvailable ? $t('admin.telegramTipMiniApp') : $t('admin.telegramTipManual') }}</div>
      </el-form>
    </el-card>

    <!-- 推送通知 -->
    <el-card class="info-card" v-if="userStore.userInfo?.telegram_id">
      <template #header>
        <div class="card-header">
          <span>{{ $t('profile.notificationSettings') }}</span>
        </div>
      </template>
      <el-form :label-width="mobile ? '0' : '120px'" :label-position="mobile ? 'top' : 'right'">
        <el-form-item :label="$t('profile.notifyEnabled')">
          <el-switch v-model="notifyEnabled" @change="toggleNotify" />
        </el-form-item>
        <div class="notify-tip">{{ $t('profile.notifyTip') }}</div>
      </el-form>
    </el-card>

    <!-- 帮助与支持 -->
    <el-card class="info-card">
      <template #header>
        <div class="card-header">
          <span>{{ $t('profile.helpSupport') }}</span>
        </div>
      </template>
      <el-form :label-width="mobile ? '0' : '120px'" :label-position="mobile ? 'top' : 'right'">
        <el-form-item :label="$t('profile.contactService')">
          <el-button @click="contactService">{{ contactPhone ? contactPhone : $t('profile.contactHint') }}</el-button>
        </el-form-item>
        <el-form-item :label="$t('profile.clearCache')">
          <el-button type="warning" plain @click="clearCache">{{ $t('profile.clearCache') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 退出登录 -->
    <div class="logout-section">
      <el-button type="danger" plain style="width:100%" @click="handleLogout">
        {{ $t('profile.logout') }}
      </el-button>
    </div>

    <!-- 修改密码对话框 -->
    <el-dialog
      v-model="showPasswordDialog"
      :title="$t('profile.changePassword')"
      :width="mobile ? '92vw' : '420px'"
      destroy-on-close
    >
      <el-form label-width="100px" :label-position="mobile ? 'top' : 'right'">
        <el-form-item :label="$t('profile.oldPassword')">
          <el-input v-model="passwordForm.old_password" type="password" show-password :placeholder="$t('profile.oldPasswordPlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('profile.newPassword')">
          <el-input v-model="passwordForm.new_password" type="password" show-password :placeholder="$t('profile.newPasswordPlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('profile.confirmPassword')">
          <el-input v-model="passwordForm.confirm_password" type="password" show-password :placeholder="$t('profile.confirmPasswordPlaceholder')" @keyup.enter="handleChangePassword" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPasswordDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="saving" @click="handleChangePassword">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- 编辑姓名 -->
    <el-dialog
      v-model="editNameVisible"
      :title="$t('profile.editPrefix') + $t('profile.name')"
      :width="mobile ? '92vw' : '360px'"
      destroy-on-close
    >
      <el-form label-width="80px" :label-position="mobile ? 'top' : 'right'">
        <el-form-item :label="$t('profile.name')">
          <el-input v-model="editNameVal" :placeholder="$t('profile.inputPrefix') + $t('profile.name')" clearable />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editNameVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="savingProfile" @click="handleSaveName">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- 编辑 Telegram ID -->
    <el-dialog
      v-model="editTgVisible"
      :title="$t('admin.telegramId')"
      :width="mobile ? '92vw' : '360px'"
      destroy-on-close
    >
      <el-form label-width="80px" :label-position="mobile ? 'top' : 'right'">
        <el-form-item :label="$t('admin.telegramId')">
          <el-input v-model="editTgVal" placeholder="123456789" clearable />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editTgVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="savingTg" @click="handleSaveTelegram">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { changePassword, updateProfile, updateAdminTelegram, bindCurrentAdminTelegram, getContactInfo } from '@/api'
import { isTelegramMiniApp, getInitData } from '@/utils/telegram'

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()

const mobile = ref(window.innerWidth < 768)
const onResize = () => { mobile.value = window.innerWidth < 768 }
onMounted(async () => {
  window.addEventListener('resize', onResize)
  try {
    const info = await getContactInfo()
    contactPhone.value = info.phone || ''
  } catch {}
})
onBeforeUnmount(() => window.removeEventListener('resize', onResize))

// 密码修改
const saving = ref(false)
const showPasswordDialog = ref(false)
const passwordForm = reactive({ old_password: '', new_password: '', confirm_password: '' })

const handleChangePassword = async () => {
  if (!passwordForm.old_password || !passwordForm.new_password || !passwordForm.confirm_password) {
    ElMessage.warning(t('profile.fillRequired'))
    return
  }
  if (passwordForm.new_password.length < 6) {
    ElMessage.warning(t('profile.passwordMinLength'))
    return
  }
  if (passwordForm.new_password !== passwordForm.confirm_password) {
    ElMessage.warning(t('profile.passwordMismatch'))
    return
  }
  saving.value = true
  try {
    await changePassword({ old_password: passwordForm.old_password, new_password: passwordForm.new_password })
    ElMessage.success(t('profile.passwordChanged'))
    showPasswordDialog.value = false
    Object.assign(passwordForm, { old_password: '', new_password: '', confirm_password: '' })
  } catch {}
  finally { saving.value = false }
}

// 编辑姓名
const editNameVisible = ref(false)
const editNameVal = ref('')
const savingProfile = ref(false)

const openEditName = () => {
  editNameVal.value = userStore.userInfo?.full_name || ''
  editNameVisible.value = true
}

const handleSaveName = async () => {
  savingProfile.value = true
  try {
    const updated = await updateProfile({ full_name: editNameVal.value.trim() })
    userStore.userInfo = { ...userStore.userInfo, ...updated }
    ElMessage.success(t('profile.updateSuccess'))
    editNameVisible.value = false
  } catch {}
  finally { savingProfile.value = false }
}

// Telegram 设置
const editTgVisible = ref(false)
const editTgVal = ref('')
const savingTg = ref(false)
const bindingTelegram = ref(false)
const telegramMiniAppAvailable = ref(isTelegramMiniApp())
const notifyEnabled = ref(userStore.userInfo?.notify_enabled !== false)

const openEditTelegram = () => {
  editTgVal.value = userStore.userInfo?.telegram_id || ''
  editTgVisible.value = true
}

const handleSaveTelegram = async () => {
  savingTg.value = true
  try {
    const updated = await updateAdminTelegram({ telegram_id: editTgVal.value.trim() })
    userStore.userInfo = { ...userStore.userInfo, ...updated }
    ElMessage.success(t('profile.updateSuccess'))
    editTgVisible.value = false
  } catch {}
  finally { savingTg.value = false }
}

const bindCurrentTelegram = async () => {
  const initData = getInitData()
  if (!initData) {
    ElMessage.warning(t('admin.notInTelegram'))
    return
  }
  bindingTelegram.value = true
  try {
    const updated = await bindCurrentAdminTelegram(initData)
    userStore.userInfo = { ...userStore.userInfo, ...updated }
    ElMessage.success(t('admin.telegramBound'))
  } catch {}
  finally { bindingTelegram.value = false }
}

const toggleNotify = async (val) => {
  try {
    const updated = await updateProfile({ notify_enabled: val })
    userStore.userInfo = { ...userStore.userInfo, ...updated }
    ElMessage.success(t('profile.updateSuccess'))
  } catch {
    notifyEnabled.value = !val
    ElMessage.error(t('profile.updateFailed'))
  }
}

// 帮助与支持
const contactPhone = ref('')

const contactService = () => {
  if (contactPhone.value) {
    window.open('tel:' + contactPhone.value)
  } else {
    ElMessage.info(t('profile.contactHint'))
  }
}

const clearCache = () => {
  localStorage.clear()
  sessionStorage.clear()
  ElMessage.success(t('profile.cacheCleared'))
}

// 退出登录
const handleLogout = () => {
  ElMessageBox.confirm(t('admin.logoutConfirm'), t('admin.hint'), {
    confirmButtonText: t('common.confirm'),
    cancelButtonText: t('common.cancel'),
    type: 'warning',
  }).then(() => {
    userStore.logout()
    router.push('/login')
  }).catch(() => {})
}
</script>

<style scoped>
.profile-page { padding: 20px; max-width: 680px; }
.profile-page h2 { margin-bottom: 20px; }
.info-card { margin-bottom: 16px; }
.card-header { display: flex; justify-content: space-between; align-items: center; font-weight: 600; }
.info-rows { display: flex; flex-direction: column; gap: 12px; }
.info-row { display: flex; align-items: center; gap: 8px; }
.info-label { width: 90px; color: #6b7280; font-size: 14px; flex-shrink: 0; }
.info-val { color: #111827; font-size: 14px; }
.editable-field { display: flex; align-items: center; gap: 8px; }
.info-value { color: #333; font-size: 14px; }
.notify-tip { color: #999; font-size: 12px; margin-top: 4px; padding-left: 120px; }
.logout-section { margin-top: 24px; padding-bottom: 40px; }

@media (max-width: 767px) {
  .profile-page { padding: 12px; padding-bottom: 70px; max-width: 100%; }
  .profile-page h2 { font-size: 18px; margin-bottom: 12px; }
  .info-label { width: 60px; }
  .notify-tip { padding-left: 0; }
}
</style>