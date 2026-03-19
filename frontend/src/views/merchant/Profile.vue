<template>
  <div class="profile-page">
    <h2>{{ $t('profile.title') }}</h2>

    <!-- 个人信息卡片 -->
    <el-card class="info-card">
      <template #header>
        <div class="card-header">
          <span>{{ $t('profile.personalInfo') }}</span>
        </div>
      </template>
      <el-form label-width="120px" class="profile-form">
        <el-form-item :label="$t('login.username')">
          <span class="info-value">{{ userStore.userInfo?.username }}</span>
        </el-form-item>
        <el-form-item :label="$t('profile.name')">
          <div class="editable-field">
            <span class="info-value">{{ userStore.userInfo?.full_name || $t('profile.notSet') }}</span>
            <el-button type="primary" link size="small" @click="editField('full_name', $t('profile.name'), userStore.userInfo?.full_name)">
              {{ $t('common.edit') }}
            </el-button>
          </div>
        </el-form-item>
        <el-form-item :label="$t('profile.phone')">
          <div class="editable-field">
            <span class="info-value">{{ userStore.userInfo?.phone || $t('profile.notSet') }}</span>
            <el-button type="primary" link size="small" @click="editField('phone', $t('profile.phone'), userStore.userInfo?.phone)">
              {{ $t('common.edit') }}
            </el-button>
          </div>
        </el-form-item>
        <el-form-item :label="$t('profile.address')">
          <div class="editable-field">
            <span class="info-value">{{ userStore.userInfo?.address || $t('profile.notSet') }}</span>
            <el-button type="primary" link size="small" @click="editField('address', $t('profile.address'), userStore.userInfo?.address)">
              {{ $t('common.edit') }}
            </el-button>
          </div>
        </el-form-item>
        <el-form-item :label="$t('profile.locationUrl')">
          <div class="editable-field">
            <span v-if="userStore.userInfo?.location_url" class="info-value">
              <a :href="userStore.userInfo.location_url" target="_blank" class="link-text">
                <el-icon><Location /></el-icon>
                {{ $t('profile.viewMap') }}
              </a>
            </span>
            <span v-else class="info-value">{{ $t('profile.notSet') }}</span>
            <el-button type="primary" link size="small" @click="editField('location_url', $t('profile.locationUrl'), userStore.userInfo?.location_url)">
              {{ $t('common.edit') }}
            </el-button>
          </div>
        </el-form-item>
        <el-form-item :label="$t('profile.storePhoto')">
          <div class="store-photo-field">
            <div v-if="userStore.userInfo?.store_photo" class="store-photo-preview">
              <el-image
                :src="userStore.userInfo.store_photo"
                fit="cover"
                style="width: 120px; height: 90px; border-radius: 8px;"
                :preview-src-list="[userStore.userInfo.store_photo]"
              />
              <el-button type="danger" link size="small" @click="removeStorePhoto">{{ $t('common.delete') }}</el-button>
            </div>
            <el-upload
              v-else
              class="store-photo-uploader"
              :show-file-list="false"
              :http-request="handleStorePhotoUpload"
              accept=".jpg,.jpeg,.png,.webp"
            >
              <div class="upload-area" @paste="onPasteStorePhoto">
                <el-icon :size="24"><Plus /></el-icon>
                <span>{{ $t('profile.uploadStorePhoto') }}</span>
              </div>
            </el-upload>
          </div>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 账户安全 -->
    <el-card class="info-card">
      <template #header>
        <div class="card-header">
          <span>{{ $t('profile.accountSecurity') }}</span>
        </div>
      </template>
      <el-form label-width="120px">
        <el-form-item :label="$t('profile.changePassword')">
          <el-button type="primary" @click="showPasswordDialog = true">
            {{ $t('profile.changePassword') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 帮助与支持 -->
    <el-card class="info-card">
      <template #header>
        <div class="card-header">
          <span>{{ $t('profile.helpSupport') }}</span>
        </div>
      </template>
      <el-form label-width="120px">
        <el-form-item :label="$t('profile.contactService')">
          <el-button @click="contactService">{{ $t('profile.contactHint') }}</el-button>
        </el-form-item>
        <el-form-item :label="$t('profile.aboutSystem')">
          <el-button @click="showAbout = true">{{ $t('profile.aboutSystem') }}</el-button>
        </el-form-item>
        <el-form-item :label="$t('profile.clearCache')">
          <el-button type="warning" plain @click="clearCache">{{ $t('profile.clearCache') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 编辑个人信息对话框 -->
    <el-dialog
      v-model="editDialogVisible"
      :title="$t('profile.editPrefix') + editLabel"
      width="420px"
      destroy-on-close
    >
      <el-form label-width="80px">
        <el-form-item :label="editLabel">
          <el-input
            v-model="editValue"
            :placeholder="$t('profile.inputPrefix') + editLabel"
            :type="editKey === 'address' ? 'textarea' : 'text'"
            :rows="editKey === 'address' ? 3 : undefined"
            clearable
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="saving" @click="handleSaveProfile">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- 修改密码对话框 -->
    <el-dialog
      v-model="showPasswordDialog"
      :title="$t('profile.changePassword')"
      width="420px"
      destroy-on-close
    >
      <el-form label-width="100px">
        <el-form-item :label="$t('profile.oldPassword')">
          <el-input v-model="passwordForm.old_password" type="password" show-password :placeholder="$t('profile.oldPasswordPlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('profile.newPassword')">
          <el-input v-model="passwordForm.new_password" type="password" show-password :placeholder="$t('profile.newPasswordPlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('profile.confirmPassword')">
          <el-input v-model="passwordForm.confirm_password" type="password" show-password :placeholder="$t('profile.confirmPasswordPlaceholder')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPasswordDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="saving" @click="handleChangePassword">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- 关于弹窗 -->
    <el-dialog v-model="showAbout" :title="$t('profile.aboutTitle')" width="420px">
      <div class="about-content">
        <div class="about-logo"><img src="/images/logo1.svg" alt="Logo" /></div>
        <p class="about-version">{{ $t('profile.aboutVersion') }}</p>
        <div v-if="aboutInfo.length > 0" class="about-features">
          <p v-for="(item, i) in aboutInfo" :key="i">{{ currentLang === 'zh' ? item.content_zh : item.content_en }}</p>
        </div>
        <div v-else class="about-features">
          <p>{{ $t('profile.aboutFeature1') }}</p>
          <p>{{ $t('profile.aboutFeature2') }}</p>
          <p>{{ $t('profile.aboutFeature3') }}</p>
          <p>{{ $t('profile.aboutFeature4') }}</p>
        </div>
      </div>
      <template #footer>
        <el-button type="primary" @click="showAbout = false">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus/es/components/message/index'
import { ElMessageBox } from 'element-plus/es/components/message-box/index'
import { Location, Plus } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { useCartStore } from '@/stores/cart'
import { updateProfile, changePassword, getPublicAnnouncements, uploadImage } from '@/api'
import { getCurrentLanguage } from '@/i18n'

const { t } = useI18n()
const userStore = useUserStore()
const cartStore = useCartStore()
const saving = ref(false)
const editDialogVisible = ref(false)
const showPasswordDialog = ref(false)
const showAbout = ref(false)
const currentLang = ref(getCurrentLanguage())
const contactInfo = ref([])
const aboutInfo = ref([])

onMounted(async () => {
  try {
    const [c, a] = await Promise.all([getPublicAnnouncements('contact'), getPublicAnnouncements('about')])
    contactInfo.value = c
    aboutInfo.value = a
  } catch (e) { /* silent */ }
})

const editKey = ref('')
const editLabel = ref('')
const editValue = ref('')

const editField = (key, label, currentValue) => {
  editKey.value = key
  editLabel.value = label
  editValue.value = currentValue || ''
  editDialogVisible.value = true
}

const handleSaveProfile = async () => {
  saving.value = true
  try {
    const val = editValue.value.trim()
    const data = { [editKey.value]: val }
    const updatedUser = await updateProfile(data)
    userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
    ElMessage.success(t('profile.updateSuccess'))
    editDialogVisible.value = false
  } catch (error) {
    console.error('Update failed:', error)
  } finally {
    saving.value = false
  }
}

const handleStorePhotoUpload = async (options) => {
  try {
    const res = await uploadImage(options.file)
    const url = res.url || res
    const updatedUser = await updateProfile({ store_photo: url })
    userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
    ElMessage.success(t('profile.updateSuccess'))
  } catch (error) {
    ElMessage.error(t('product.uploadFailed'))
  }
}

const removeStorePhoto = async () => {
  try {
    await updateProfile({ store_photo: '' })
    userStore.userInfo = { ...userStore.userInfo, store_photo: '' }
    ElMessage.success(t('profile.updateSuccess'))
  } catch (error) { /* silent */ }
}

const onPasteStorePhoto = async (e) => {
  const items = e.clipboardData?.items
  if (!items) return
  for (const item of items) {
    if (item.type.startsWith('image/')) {
      e.preventDefault()
      const file = item.getAsFile()
      if (file) await handleStorePhotoUpload({ file })
      break
    }
  }
}

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
  } catch (error) { console.error(error) } finally { saving.value = false }
}

const contactService = () => {
  const lang = getCurrentLanguage()
  let msg = t('profile.contactMessage')
  if (contactInfo.value.length > 0) {
    msg = contactInfo.value.map(c => lang === 'zh' ? c.content_zh : c.content_en).join('\n\n')
  }
  ElMessageBox.alert(msg, t('profile.contactTitle'), { confirmButtonText: t('profile.contactOk') })
}

const clearCache = () => {
  ElMessageBox.confirm(t('profile.clearCacheMessage'), t('profile.clearCacheTitle'), {
    confirmButtonText: t('common.confirm'),
    cancelButtonText: t('common.cancel'),
    type: 'warning',
  }).then(() => {
    cartStore.clear()
    ElMessage.success(t('profile.cacheCleared'))
  }).catch(() => {})
}
</script>

<style scoped>
.profile-page { padding: 20px; max-width: 680px; }
.profile-page h2 { margin-bottom: 20px; }
.info-card { margin-bottom: 20px; }
.card-header { font-weight: 600; }
.profile-form .el-form-item { margin-bottom: 16px; }
.info-value { color: #333; font-size: 14px; }
.editable-field { display: flex; align-items: center; gap: 12px; }
.link-text { color: #409eff; text-decoration: none; display: flex; align-items: center; gap: 4px; }
.link-text:hover { text-decoration: underline; }
.store-photo-field { display: flex; align-items: flex-start; gap: 12px; }
.store-photo-preview { display: flex; flex-direction: column; align-items: center; gap: 6px; }
.store-photo-uploader .upload-area {
  width: 120px; height: 90px; border: 2px dashed #dcdfe6; border-radius: 8px;
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 4px; cursor: pointer; transition: all 0.2s; color: #999; font-size: 12px;
}
.store-photo-uploader .upload-area:hover { border-color: #409eff; color: #409eff; }
.about-content { text-align: center; padding: 10px 0; }
.about-logo img { height: 48px; margin-bottom: 10px; }
.about-version { color: #999; font-size: 13px; margin-bottom: 16px; }
.about-features p { color: #666; font-size: 14px; margin: 6px 0; }

@media (max-width: 767px) {
  .profile-page { padding: 12px; padding-bottom: 70px; }
  .profile-page h2 { font-size: 18px; margin-bottom: 12px; }
  .el-dialog { width: 92vw !important; }
}
</style>
