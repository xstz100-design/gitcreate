<template>
  <div class="mobile-profile">
    <van-nav-bar :title="$t('profile.title')" fixed placeholder />
    
    <!-- 用户信息卡片 -->
    <div class="profile-header">
      <div class="avatar">
        <van-icon name="user-o" size="32" color="#fff" />
      </div>
      <div class="user-info">
        <div class="user-name">{{ userStore.userInfo?.full_name || $t('profile.notLoggedIn') }}</div>
        <div class="user-phone">{{ userStore.userInfo?.phone || '' }}</div>
      </div>
    </div>
    
    <!-- 个人信息（可编辑） -->
    <van-cell-group inset :title="$t('profile.personalInfo')">
      <van-cell
        :title="$t('profile.name')"
        :value="userStore.userInfo?.full_name || $t('profile.notSet')"
        icon="contact"
        is-link
        @click="editField('full_name', $t('profile.name'), userStore.userInfo?.full_name)"
      />
      <van-cell
        :title="$t('profile.phone')"
        :value="userStore.userInfo?.phone || $t('profile.notSet')"
        icon="phone-o"
        is-link
        @click="editField('phone', $t('profile.phone'), userStore.userInfo?.phone)"
      />
      <van-cell
        :title="$t('profile.address')"
        :value="userStore.userInfo?.address || $t('profile.notSet')"
        icon="location-o"
        is-link
        @click="editField('address', $t('profile.address'), userStore.userInfo?.address)"
      />
      <van-cell
        :title="$t('profile.locationUrl')"
        icon="guide-o"
        is-link
        @click="editField('location_url', $t('profile.locationUrl'), userStore.userInfo?.location_url)"
      >
        <template #value>
          <a
            v-if="userStore.userInfo?.location_url"
            :href="userStore.userInfo.location_url"
            target="_blank"
            @click.stop
            class="map-link"
          >{{ $t('profile.viewMap') }}</a>
          <span v-else>{{ $t('profile.notSet') }}</span>
        </template>
      </van-cell>
      <van-cell :title="$t('profile.storePhoto')" icon="photograph">
        <template #value>
          <div class="store-photo-cell">
            <van-image
              v-if="userStore.userInfo?.store_photo"
              :src="userStore.userInfo.store_photo"
              width="60"
              height="45"
              fit="cover"
              radius="4"
              @click="previewStorePhoto"
            />
            <van-uploader
              v-else
              :after-read="onStorePhotoRead"
              :max-count="1"
              accept="image/*"
            >
              <van-button size="mini" type="primary" plain>{{ $t('profile.uploadStorePhoto') }}</van-button>
            </van-uploader>
            <van-icon v-if="userStore.userInfo?.store_photo" name="cross" class="remove-photo" @click="removeStorePhoto" />
          </div>
        </template>
      </van-cell>
    </van-cell-group>
    
    <van-cell-group inset :title="$t('profile.accountSecurity')">
      <van-cell
        :title="$t('profile.changePassword')"
        is-link
        icon="lock"
        @click="showPasswordDialog = true"
      />
    </van-cell-group>
    
    <!-- Telegram 通知设置 (仅管理员) -->
    <van-cell-group v-if="userStore.userInfo?.role === 'admin'" inset :title="$t('admin.telegramSettings')">
      <van-cell
        :title="$t('admin.telegramBotToken')"
        :value="userStore.userInfo?.telegram_bot_token ? '已设置' : $t('profile.notSet')"
        icon="chat-o"
        is-link
        @click="editField('telegram_bot_token', $t('admin.telegramBotToken'), userStore.userInfo?.telegram_bot_token)"
      />
      <van-cell
        :title="$t('admin.telegramChatId')"
        :value="userStore.userInfo?.telegram_chat_id || $t('profile.notSet')"
        icon="comment-o"
        is-link
        @click="editField('telegram_chat_id', $t('admin.telegramChatId'), userStore.userInfo?.telegram_chat_id)"
      />
      <div class="telegram-hint">
        <span>{{ $t('admin.telegramTip') }}</span>
      </div>
    </van-cell-group>
    
    <van-cell-group inset :title="$t('profile.helpSupport')">
      <van-cell
        :title="$t('profile.contactService')"
        is-link
        icon="service-o"
        @click="contactService"
        :label="$t('profile.contactHint')"
      />
      <van-cell
        :title="$t('profile.aboutSystem')"
        is-link
        icon="info-o"
        @click="showAbout = true"
      />
      <van-cell
        :title="$t('profile.clearCache')"
        is-link
        icon="delete-o"
        @click="clearCache"
      />
    </van-cell-group>
    
    <div class="logout-section">
      <van-button
        plain
        block
        type="default"
        @click="handleLogout"
      >
        {{ $t('profile.logout') }}
      </van-button>
    </div>
    
    <!-- 编辑个人信息弹窗 -->
    <van-dialog
      v-model:show="showEditDialog"
      :title="$t('profile.editPrefix') + editLabel"
      show-cancel-button
      :before-close="handleSaveProfile"
    >
      <div class="edit-form">
        <van-field
          v-model="editValue"
          :label="editLabel"
          :placeholder="$t('profile.inputPrefix') + editLabel"
          :type="editKey === 'phone' ? 'tel' : 'text'"
          clearable
        />
      </div>
    </van-dialog>
    
    <!-- 修改密码弹窗 -->
    <van-dialog
      v-model:show="showPasswordDialog"
      :title="$t('profile.changePassword')"
      show-cancel-button
      :before-close="handleChangePassword"
    >
      <div class="password-form">
        <van-field
          v-model="passwordForm.old_password"
          type="password"
          :label="$t('profile.oldPassword')"
          :placeholder="$t('profile.oldPasswordPlaceholder')"
        />
        <van-field
          v-model="passwordForm.new_password"
          type="password"
          :label="$t('profile.newPassword')"
          :placeholder="$t('profile.newPasswordPlaceholder')"
        />
        <van-field
          v-model="passwordForm.confirm_password"
          type="password"
          :label="$t('profile.confirmPassword')"
          :placeholder="$t('profile.confirmPasswordPlaceholder')"
        />
      </div>
    </van-dialog>
    
    <!-- 关于弹窗 -->
    <van-dialog
      v-model:show="showAbout"
      :title="$t('profile.aboutTitle')"
      :confirm-button-text="$t('common.confirm')"
    >
      <div class="about-content">
        <div class="about-logo">
          <img src="/images/logo1.svg" alt="Logo" />
        </div>
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
    </van-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { showSuccessToast, showToast, showDialog, showImagePreview } from 'vant'
import { useUserStore } from '@/stores/user'
import { useCartStore } from '@/stores/cart'
import { changePassword, updateProfile, getPublicAnnouncements, uploadImage } from '@/api'
import { hapticFeedback } from '@/utils/device'
import { getCurrentLanguage } from '@/i18n'

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()
const cartStore = useCartStore()

const showAbout = ref(false)
const showPasswordDialog = ref(false)
const contactInfo = ref([])
const aboutInfo = ref([])

// 编辑个人信息
const showEditDialog = ref(false)
const editKey = ref('')
const editLabel = ref('')
const editValue = ref('')

const editField = (key, label, currentValue) => {
  editKey.value = key
  editLabel.value = label
  editValue.value = currentValue || ''
  showEditDialog.value = true
  hapticFeedback('light')
}

const handleSaveProfile = async (action) => {
  if (action !== 'confirm') return true
  
  if (!editValue.value.trim()) {
    showToast(t('profile.inputRequired'))
    return false
  }
  
  try {
    const data = { [editKey.value]: editValue.value.trim() }
    const updatedUser = await updateProfile(data)
    userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
    hapticFeedback('success')
    showSuccessToast(t('profile.updateSuccess'))
    return true
  } catch (error) {
    showToast(t('profile.updateFailed'))
    return false
  }
}

// 门面照片上传
const onStorePhotoRead = async (file) => {
  try {
    const res = await uploadImage(file.file)
    const url = res.url || res
    const updatedUser = await updateProfile({ store_photo: url })
    userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
    hapticFeedback('success')
    showSuccessToast(t('profile.updateSuccess'))
  } catch (error) {
    showToast(t('product.uploadFailed'))
  }
}

const previewStorePhoto = () => {
  if (userStore.userInfo?.store_photo) {
    showImagePreview([userStore.userInfo.store_photo])
  }
}

const removeStorePhoto = async () => {
  try {
    await updateProfile({ store_photo: '' })
    userStore.userInfo = { ...userStore.userInfo, store_photo: '' }
    hapticFeedback('success')
    showSuccessToast(t('profile.updateSuccess'))
  } catch (error) { /* silent */ }
}

const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
})

// 修改密码
const handleChangePassword = async (action) => {
  if (action !== 'confirm') return true
  
  if (!passwordForm.old_password || !passwordForm.new_password) {
    showToast(t('profile.fillRequired'))
    return false
  }
  
  if (passwordForm.new_password.length < 6) {
    showToast(t('profile.passwordMinLength'))
    return false
  }
  
  if (passwordForm.new_password !== passwordForm.confirm_password) {
    showToast(t('profile.passwordMismatch'))
    return false
  }
  
  try {
    await changePassword({
      old_password: passwordForm.old_password,
      new_password: passwordForm.new_password,
    })
    hapticFeedback('success')
    showSuccessToast(t('profile.passwordChanged'))
    passwordForm.old_password = ''
    passwordForm.new_password = ''
    passwordForm.confirm_password = ''
    return true
  } catch (error) {
    return false
  }
}

// 联系客服
const contactService = () => {
  hapticFeedback('light')
  const lang = getCurrentLanguage()
  let msg = t('profile.contactMessage')
  if (contactInfo.value.length > 0) {
    msg = contactInfo.value.map(c => lang === 'zh' ? c.content_zh : c.content_en).join('\n\n')
  }
  showDialog({
    title: t('profile.contactTitle'),
    message: msg,
    confirmButtonText: t('profile.contactOk'),
  })
}

// 清除缓存
const clearCache = async () => {
  const confirmed = await showDialog({
    title: t('profile.clearCacheTitle'),
    message: t('profile.clearCacheMessage'),
    showCancelButton: true,
    confirmButtonText: t('common.confirm'),
    cancelButtonText: t('common.cancel'),
  }).catch(() => false)
  
  if (confirmed !== false) {
    cartStore.clear()
    hapticFeedback('success')
    showSuccessToast(t('profile.cacheCleared'))
  }
}

// 退出登录
const handleLogout = async () => {
  const confirmed = await showDialog({
    title: t('profile.logoutTitle'),
    message: t('profile.logoutMessage'),
    showCancelButton: true,
    confirmButtonText: t('common.confirm'),
    cancelButtonText: t('common.cancel'),
  }).catch(() => false)
  
  if (confirmed !== false) {
    userStore.logout()
    cartStore.clear()
    hapticFeedback('success')
    router.push('/login')
  }
}

const currentLang = ref(getCurrentLanguage())

// 加载客服和关于信息
const loadContactAbout = async () => {
  try {
    const [contact, about] = await Promise.all([
      getPublicAnnouncements('contact'),
      getPublicAnnouncements('about'),
    ])
    contactInfo.value = contact
    aboutInfo.value = about
  } catch {
    // 静默处理
  }
}

onMounted(() => {
  loadContactAbout()
})
</script>

<style scoped>
.mobile-profile {
  min-height: 100vh;
  background: var(--bg-gray, #f7f7f7);
  padding-bottom: 60px;
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 28px 20px;
  background: var(--primary-color, #2b2b2b);
  color: #fff;
}

.avatar {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: rgba(255,255,255,0.15);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.user-name {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 4px;
}

.user-phone {
  font-size: 13px;
  opacity: 0.8;
}

.logout-section {
  padding: 24px 16px;
}

.password-form {
  padding: 8px 0;
}

.about-content {
  padding: 20px;
  text-align: center;
}

.about-logo {
  display: flex;
  justify-content: center;
  margin-bottom: 16px;
  
  img {
    width: 200px;
    max-width: 100%;
    height: auto;
    object-fit: contain;
  }
}

.about-version {
  font-size: 13px;
  color: #999;
  margin: 0 0 16px;
}

.about-features {
  text-align: left;
  padding: 12px 16px;
  background: #f5f5f5;
  border-radius: 4px;
}

.about-features p {
  font-size: 13px;
  color: #666;
  margin: 0;
  padding: 4px 0;
}

.about-features p::before {
  content: '• ';
  color: #999;
}

.map-link {
  color: #1989fa;
  text-decoration: none;
}

.store-photo-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.remove-photo {
  color: #ee0a24;
  font-size: 16px;
  cursor: pointer;
}

.telegram-hint {
  padding: 8px 16px 12px;
  font-size: 12px;
  color: #999;
  line-height: 1.5;
}
</style>
