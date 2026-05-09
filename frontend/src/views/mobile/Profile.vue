<template>
  <div class="mobile-profile">
    <van-nav-bar :title="$t('profile.title')" fixed placeholder />
    
    <!-- 审核状态横幅 -->
    <div v-if="userStore.isMerchant && !userStore.isApproved" class="status-banner">
      <div v-if="userStore.approvalStatus === 'pending'" class="banner pending">
        <van-icon name="clock-o" size="20" />
        <div class="banner-text">
          <strong v-if="!userStore.profileCompleted">{{ $t('profile.pleaseCompleteProfile') }}</strong>
          <strong v-else>{{ $t('profile.pendingApproval') }}</strong>
          <p v-if="userStore.profileCompleted">{{ $t('profile.pendingApprovalTip') }}</p>
          <p v-else>{{ $t('profile.completeProfileTip') }}</p>
        </div>
      </div>
      <div v-else-if="userStore.approvalStatus === 'rejected'" class="banner rejected">
        <van-icon name="warning-o" size="20" />
        <div class="banner-text">
          <strong>{{ $t('profile.rejected') }}</strong>
          <p v-if="userStore.userInfo?.rejected_reason">{{ $t('profile.rejectedReason') }}: {{ userStore.userInfo.rejected_reason }}</p>
          <p>{{ $t('profile.rejectedTip') }}</p>
        </div>
      </div>
    </div>

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
    
    <!-- 我的地址 -->
    <van-cell-group inset style="margin-top:16px">
      <van-cell
        title="我的地址"
        icon="location-o"
        is-link
        @click="openAddressPopup"
      >
        <template #label>
          <span v-if="addressSummary" class="addr-summary-text">{{ addressSummary }}</span>
          <span v-else class="addr-summary-empty">点击设置配送地址</span>
        </template>
      </van-cell>
    </van-cell-group>

    <!-- 地址编辑弹层 -->
    <van-popup
      v-model:show="showAddressPopup"
      position="bottom"
      round
      :style="{ maxHeight: '88%' }"
      closeable
    >
      <div class="addr-popup">
        <div class="addr-popup-title">我的地址</div>
        <div class="addr-popup-body">
          <van-field
            v-model="addressForm.full_name"
            :label="$t('profile.name')"
            :placeholder="$t('profile.inputPrefix') + $t('profile.name')"
            clearable
          />
          <van-field
            :model-value="userStore.userInfo?.phone || ''"
            :label="$t('profile.phone')"
            readonly
            is-link
            @click="openPhoneEdit"
          />
          <van-field
            v-model="addressForm.address"
            :label="$t('profile.address')"
            type="textarea"
            rows="2"
            :placeholder="$t('profile.inputPrefix') + $t('profile.address')"
            autosize
            clearable
          />
          <van-cell
            :title="$t('profile.locationUrl')"
            icon="guide-o"
            is-link
            @click="openLocationPickerFromPopup"
          >
            <template #value>
              <span v-if="userStore.userInfo?.location_url" class="map-set-text">{{ $t('profile.viewMap') }}</span>
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
        </div>
        <div class="addr-popup-footer">
          <van-button type="primary" block round :loading="saving" @click="handleSaveAddress">
            {{ $t('common.save') }}
          </van-button>
        </div>
      </div>
    </van-popup>

    <!-- 资料未完成时显示提交按钮 -->
    <div v-if="userStore.isMerchant && !userStore.profileCompleted" class="submit-profile-section">
      <van-button type="primary" block round @click="submitProfileForReview" :loading="submittingProfile">
        {{ $t('profile.submitForReview') }}
      </van-button>
      <p class="submit-hint">{{ $t('profile.submitHint') }}</p>
    </div>

    <!-- 已提交待审核时显示刷新状态按钮 -->
    <div v-else-if="userStore.isMerchant && userStore.approvalStatus === 'pending'" class="submit-profile-section">
      <van-button type="primary" block round @click="checkApprovalStatus" :loading="checkingStatus" icon="replay">
        {{ $t('profile.checkStatus') }}
      </van-button>
    </div>

    <!-- 被拒绝时显示重新提交 -->
    <div v-else-if="userStore.isMerchant && userStore.approvalStatus === 'rejected'" class="submit-profile-section">
      <van-button type="warning" block round @click="submitProfileForReview" :loading="submittingProfile">
        {{ $t('profile.resubmitForReview') }}
      </van-button>
    </div>
    
    <van-cell-group inset :title="$t('profile.accountSecurity')" v-if="userStore.isAdmin">
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
        :title="$t('admin.telegramId')"
        :value="userStore.userInfo?.telegram_id || $t('profile.notSet')"
        icon="chat-o"
        is-link
        @click="editField('telegram_id', $t('admin.telegramId'), userStore.userInfo?.telegram_id)"
      />
      <van-cell
        v-if="telegramMiniAppAvailable"
        :title="userStore.userInfo?.telegram_id ? $t('admin.rebindCurrentTelegram') : $t('admin.bindCurrentTelegram')"
        icon="link-o"
        is-link
        @click="bindCurrentTelegram"
      />
      <div class="telegram-hint">
        <span>{{ telegramMiniAppAvailable ? $t('admin.telegramTipMiniApp') : $t('admin.telegramTipManual') }}</span>
      </div>
    </van-cell-group>

    <!-- 推送通知设置 -->
    <van-cell-group v-if="userStore.userInfo?.telegram_id" inset :title="$t('profile.notificationSettings')">
      <van-cell :title="$t('profile.notifyEnabled')" icon="bullhorn-o">
        <template #right-icon>
          <van-switch v-model="notifyEnabled" size="20" @change="toggleNotify" />
        </template>
      </van-cell>
      <div class="telegram-hint">
        <span>{{ $t('profile.notifyTip') }}</span>
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

    <van-dialog
      v-model:show="showPhoneVerifyDialog"
      :title="$t('profile.verifyPhone')"
      show-cancel-button
      :confirm-button-text="$t('profile.verifyAndSave')"
      :before-close="handleVerifyPhone"
    >
      <div class="edit-form">
        <van-field
          v-model="phoneVerifyForm.phone"
          :label="$t('profile.phone')"
          :placeholder="$t('register.phonePlaceholder')"
          type="tel"
          clearable
        />
        <van-field
          v-if="!canDirectVerifyPhone()"
          v-model="phoneVerifyForm.code"
          :label="$t('register.smsCodePlaceholder')"
          :placeholder="$t('register.smsCodePlaceholder')"
          clearable
        >
          <template #button>
            <van-button size="small" type="primary" plain :loading="sendingPhoneCode" @click.stop="sendPhoneCode">
              {{ $t('profile.sendTelegramCode') }}
            </van-button>
          </template>
        </van-field>
        <div class="telegram-hint">
          <span>{{ canDirectVerifyPhone() ? $t('profile.verifyPhoneMiniAppTip') : $t('profile.verifyPhoneTip') }}</span>
        </div>
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
    
    <!-- 地图选取位置弹窗 -->
    <van-dialog
      v-model:show="showMapPicker"
      :title="$t('profile.pickLocation')"
      show-cancel-button
      :confirm-button-text="$t('common.save')"
      :before-close="handleSaveLocation"
    >
      <div class="map-picker-wrap">
        <div class="map-toolbar">
          <button class="locate-btn" @click="locateCurrentPosition" :disabled="locating">
            <span v-if="locating">⏳ {{ $t('profile.locating') }}</span>
            <span v-else>📍 {{ $t('profile.locateMe') }}</span>
          </button>
        </div>
        <div id="google-map-picker" ref="mapContainerRef" class="map-container"></div>
        <van-field
          v-model="pickedLocationUrl"
          :label="$t('profile.locationUrl')"
          :placeholder="$t('profile.locationUrlPlaceholder')"
          clearable
          class="map-url-field"
        />
        <div class="map-picker-hint">{{ $t('profile.locationPickerHint') }}</div>
      </div>
    </van-dialog>

  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onActivated, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { showSuccessToast, showToast, showDialog, showImagePreview } from 'vant'
import { useUserStore } from '@/stores/user'
import { useCartStore } from '@/stores/cart'
import { changePassword, updateProfile, getPublicAnnouncements, uploadImage, submitForReview, updateAdminTelegram, bindCurrentAdminTelegram, sendPhoneVerificationCode, verifyPhoneVerificationCode, verifyPhoneWithTelegram } from '@/api'
import { hapticFeedback } from '@/utils/device'
import { getCurrentLanguage } from '@/i18n'
import { getInitData, isTelegramMiniApp } from '@/utils/telegram'

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()
const cartStore = useCartStore()

const showPasswordDialog = ref(false)
const showMapPicker = ref(false)
const locating = ref(false)
const pickedLocationUrl = ref('')
const mapContainerRef = ref(null)
const contactInfo = ref([])
const submittingProfile = ref(false)
const checkingStatus = ref(false)
const notifyEnabled = ref(userStore.userInfo?.notify_enabled !== false)
const telegramMiniAppAvailable = ref(isTelegramMiniApp())
const bindingTelegram = ref(false)
// 编辑个人信息
const showEditDialog = ref(false)
const showPhoneVerifyDialog = ref(false)
const editKey = ref('')
const editLabel = ref('')
const editValue = ref('')
const sendingPhoneCode = ref(false)
const phoneVerifyForm = reactive({ phone: '', code: '' })

// ── 我的地址弹层 ──
const showAddressPopup = ref(false)
const saving = ref(false)
const addressForm = reactive({ full_name: '', address: '' })

const addressSummary = computed(() => {
  const info = userStore.userInfo
  if (!info) return ''
  return [info.full_name, info.phone, info.address].filter(Boolean).join(' · ')
})

const openAddressPopup = () => {
  addressForm.full_name = userStore.userInfo?.full_name || ''
  addressForm.address = userStore.userInfo?.address || ''
  showAddressPopup.value = true
  hapticFeedback('light')
}

const openPhoneEdit = () => {
  handlePhoneEdit()
}

const openLocationPickerFromPopup = () => {
  openLocationPicker()
}

const handleSaveAddress = async () => {
  saving.value = true
  try {
    const updatedUser = await updateProfile({
      full_name: addressForm.full_name.trim(),
      address: addressForm.address.trim(),
    })
    userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
    hapticFeedback('success')
    showSuccessToast(t('profile.updateSuccess'))
    showAddressPopup.value = false
  } catch {
    showToast(t('profile.updateFailed'))
  } finally {
    saving.value = false
  }
}

const canDirectVerifyPhone = () => telegramMiniAppAvailable.value && !!getInitData()

const editField = (key, label, currentValue) => {
  editKey.value = key
  editLabel.value = label
  editValue.value = currentValue || ''
  showEditDialog.value = true
  hapticFeedback('light')
}

const handlePhoneEdit = () => {
  if (!userStore.userInfo?.telegram_id) {
    editField('phone', t('profile.phone'), userStore.userInfo?.phone)
    return
  }
  phoneVerifyForm.phone = userStore.userInfo?.phone || ''
  phoneVerifyForm.code = ''
  showPhoneVerifyDialog.value = true
  hapticFeedback('light')
}

// ─────────── 地图位置选取 ───────────
const GOOGLE_MAPS_API_KEY = 'AIzaSyBLMXrpizbEE4f36sUCerOasHVWM8Doumc'
let googleMapInstance = null
let googleMapMarker = null

const loadGoogleMapsScript = () => {
  if (window.google?.maps) return Promise.resolve()
  if (document.getElementById('google-maps-script')) {
    return new Promise(resolve => {
      const check = setInterval(() => {
        if (window.google?.maps) { clearInterval(check); resolve() }
      }, 100)
    })
  }
  return new Promise((resolve, reject) => {
    const s = document.createElement('script')
    s.id = 'google-maps-script'
    s.src = `https://maps.googleapis.com/maps/api/js?key=${GOOGLE_MAPS_API_KEY}&language=zh-CN`
    s.async = true
    s.defer = true
    s.onload = resolve
    s.onerror = reject
    document.head.appendChild(s)
  })
}

// 将地图中心移到指定坐标并打上标记
const placeMapMarker = (latlng) => {
  if (!googleMapInstance) return
  if (googleMapMarker) googleMapMarker.setMap(null)
  googleMapMarker = new window.google.maps.Marker({ position: latlng, map: googleMapInstance })
  googleMapInstance.panTo(latlng)
  const lat = (typeof latlng.lat === 'function' ? latlng.lat() : latlng.lat).toFixed(6)
  const lng = (typeof latlng.lng === 'function' ? latlng.lng() : latlng.lng).toFixed(6)
  pickedLocationUrl.value = `https://maps.google.com/?q=${lat},${lng}`
}

// 定位当前位置按钮
const locateCurrentPosition = () => {
  if (!navigator.geolocation) {
    showToast(t('profile.geolocationUnsupported'))
    return
  }
  locating.value = true
  navigator.geolocation.getCurrentPosition(
    (pos) => {
      locating.value = false
      const latlng = { lat: pos.coords.latitude, lng: pos.coords.longitude }
      if (googleMapInstance) {
        googleMapInstance.setCenter(latlng)
        googleMapInstance.setZoom(17)
        placeMapMarker(latlng)
      }
    },
    () => {
      locating.value = false
      showToast(t('profile.geolocationFailed'))
    },
    { enableHighAccuracy: true, timeout: 10000 }
  )
}

const openLocationPicker = async () => {
  pickedLocationUrl.value = userStore.userInfo?.location_url || ''
  showMapPicker.value = true
  hapticFeedback('light')
  await nextTick()
  try {
    await loadGoogleMapsScript()
    await nextTick()
    const container = mapContainerRef.value
    if (!container || !window.google?.maps) return

    // Phnom Penh default center
    const defaultCenter = { lat: 11.5564, lng: 104.9282 }
    let initCenter = defaultCenter

    // If existing URL has coordinates, center there
    const existing = pickedLocationUrl.value
    if (existing) {
      const m = existing.match(/[?&]q=([\-\d.]+),([\-\d.]+)/) || existing.match(/@([\-\d.]+),([\-\d.]+)/)
      if (m) initCenter = { lat: parseFloat(m[1]), lng: parseFloat(m[2]) }
    }

    googleMapInstance = new window.google.maps.Map(container, {
      center: initCenter,
      zoom: 15,
      mapTypeControl: false,
      streetViewControl: false,
      fullscreenControl: false,
      // greedy：让地图独占单/双指手势，支持在弹窗内双指缩放和拖动
      gestureHandling: 'greedy',
    })

    if (existing && initCenter !== defaultCenter) {
      googleMapMarker = new window.google.maps.Marker({ position: initCenter, map: googleMapInstance })
    } else if (!existing) {
      // 没有已存坐标时，尝试自动定位
      if (navigator.geolocation) {
        locating.value = true
        navigator.geolocation.getCurrentPosition(
          (pos) => {
            locating.value = false
            const latlng = { lat: pos.coords.latitude, lng: pos.coords.longitude }
            googleMapInstance?.setCenter(latlng)
            googleMapInstance?.setZoom(17)
          },
          () => { locating.value = false },
          { enableHighAccuracy: true, timeout: 8000 }
        )
      }
    }

    googleMapInstance.addListener('click', (e) => {
      placeMapMarker(e.latLng)
    })
  } catch {
    showToast(t('profile.mapLoadFailed'))
  }
}

const openMapLink = () => {
  const url = userStore.userInfo?.location_url
  if (!url) return
  const tg = window.Telegram?.WebApp
  if (tg?.openLink) tg.openLink(url)
  else window.open(url, '_blank')
}

const handleSaveLocation = async (action) => {
  if (action !== 'confirm') {
    googleMapInstance = null; googleMapMarker = null
    return true
  }
  if (!pickedLocationUrl.value.trim()) {
    showToast(t('profile.inputRequired'))
    return false
  }
  try {
    const updatedUser = await updateProfile({ location_url: pickedLocationUrl.value.trim() })
    userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
    hapticFeedback('success')
    showSuccessToast(t('profile.updateSuccess'))
    googleMapInstance = null; googleMapMarker = null
    return true
  } catch {
    showToast(t('profile.updateFailed'))
    return false
  }
}

const sendPhoneCode = async () => {
  if (canDirectVerifyPhone()) {
    showToast(t('profile.verifyPhoneMiniAppTip'))
    return
  }
  if (!userStore.userInfo?.telegram_id) {
    showToast(t('profile.telegramBindRequired'))
    return
  }
  if (!phoneVerifyForm.phone.trim()) {
    showToast(t('profile.pleaseInputPhone'))
    return
  }
  sendingPhoneCode.value = true
  try {
    await sendPhoneVerificationCode({ phone: phoneVerifyForm.phone.trim() })
    hapticFeedback('success')
    showSuccessToast(t('profile.telegramCodeSent'))
  } finally {
    sendingPhoneCode.value = false
  }
}

const handleVerifyPhone = async (action) => {
  if (action !== 'confirm') return true
  if (!phoneVerifyForm.phone.trim()) {
    showToast(t('profile.pleaseInputPhone'))
    return false
  }
  if (canDirectVerifyPhone()) {
    try {
      const updatedUser = await verifyPhoneWithTelegram({
        phone: phoneVerifyForm.phone.trim(),
        init_data: getInitData(),
      })
      userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
      notifyEnabled.value = updatedUser.notify_enabled !== false
      hapticFeedback('success')
      showSuccessToast(t('profile.phoneVerified'))
      return true
    } catch {
      return false
    }
  }
  if (!phoneVerifyForm.code.trim()) {
    showToast(t('profile.phoneCodeRequired'))
    return false
  }
  try {
    const updatedUser = await verifyPhoneVerificationCode({
      phone: phoneVerifyForm.phone.trim(),
      code: phoneVerifyForm.code.trim(),
    })
    userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
    notifyEnabled.value = updatedUser.notify_enabled !== false
    hapticFeedback('success')
    showSuccessToast(t('profile.phoneVerified'))
    return true
  } catch {
    return false
  }
}

const handleSaveProfile = async (action) => {
  if (action !== 'confirm') return true
  
  if (!editValue.value.trim()) {
    showToast(t('profile.inputRequired'))
    return false
  }
  
  try {
    let updatedUser
    if (editKey.value === 'telegram_id') {
      // 管理员绑定 Telegram ID 使用专用接口
      const telegram_id = parseInt(editValue.value.trim())
      if (isNaN(telegram_id)) {
        showToast(t('admin.telegramIdInvalid') || 'Telegram ID 格式错误')
        return false
      }
      updatedUser = await updateAdminTelegram({ telegram_id })
    } else {
      const data = { [editKey.value]: editValue.value.trim() }
      updatedUser = await updateProfile(data)
    }
    userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
    hapticFeedback('success')
    showSuccessToast(t('profile.updateSuccess'))
    return true
  } catch (error) {
    showToast(t('profile.updateFailed'))
    return false
  }
}

const bindCurrentTelegram = async () => {
  if (!telegramMiniAppAvailable.value) {
    showToast(t('admin.telegramBindMiniAppOnly'))
    return
  }

  const initData = getInitData()
  if (!initData) {
    showToast(t('admin.telegramBindMiniAppOnly'))
    return
  }

  bindingTelegram.value = true
  try {
    const updatedUser = await bindCurrentAdminTelegram(initData)
    userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
    hapticFeedback('success')
    showSuccessToast(t('admin.telegramBindSuccess'))
  } finally {
    bindingTelegram.value = false
  }
}

const toggleNotify = async (val) => {
  try {
    const updatedUser = await updateProfile({ notify_enabled: val })
    userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
    hapticFeedback('success')
    showSuccessToast(t('profile.updateSuccess'))
  } catch (error) {
    notifyEnabled.value = !val
    showToast(t('profile.updateFailed'))
  }
}

// 提交资料等待审核
const submitProfileForReview = async () => {
  const info = userStore.userInfo
  if (!info?.full_name || info.full_name.startsWith('TG_')) {
    showToast(t('profile.pleaseInputName'))
    return
  }
  if (!info?.phone) {
    showToast(t('profile.pleaseInputPhone'))
    return
  }
  if (!info?.address) {
    showToast(t('profile.pleaseInputAddress'))
    return
  }
  submittingProfile.value = true
  try {
    await submitForReview()
    await userStore.fetchUserInfo()
    hapticFeedback('success')
    showSuccessToast(t('profile.profileSubmitted'))
  } catch (error) {
    // request.js 拦截器已显示错误
  } finally {
    submittingProfile.value = false
  }
}

// 刷新审核状态
const checkApprovalStatus = async () => {
  checkingStatus.value = true
  try {
    await userStore.fetchUserInfo()
    hapticFeedback('light')
    if (userStore.isApproved) {
      hapticFeedback('success')
      showSuccessToast(t('profile.approved'))
      // 审核通过，延迟后跳转到商城
      setTimeout(() => router.push('/m/shop'), 800)
    } else {
      showToast(t('profile.pendingApproval'))
    }
  } catch {
    showToast(t('profile.updateFailed'))
  } finally {
    checkingStatus.value = false
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

// 加载客服信息
const loadContactAbout = async () => {
  try {
    const contact = await getPublicAnnouncements('contact')
    contactInfo.value = contact
  } catch {
    // 静默处理
  }
}

onMounted(() => {
  loadContactAbout()
  // 刷新用户信息，确保审核状态是最新的
  userStore.fetchUserInfo()
})

// keep-alive 激活时也刷新用户信息
onActivated(() => {
  userStore.fetchUserInfo()
})
</script>

<style scoped>
.mobile-profile {
  min-height: var(--tg-viewport-height, 100vh);
  background: var(--bg-gray, #f7f7f7);
  padding-bottom: calc(60px + var(--tg-content-safe-area-inset-bottom, env(safe-area-inset-bottom, 0px)));
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
  cursor: pointer;
}

.map-picker-wrap {
  padding: 0;
}

.map-toolbar {
  display: flex;
  justify-content: flex-end;
  padding: 8px 12px 4px;
}

.locate-btn {
  background: #1976d2;
  color: #fff;
  border: none;
  border-radius: 20px;
  padding: 6px 14px;
  font-size: 13px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
}

.locate-btn:disabled {
  background: #90bce8;
  cursor: not-allowed;
}

.map-container {
  width: 100%;
  height: 260px;
  background: #eee;
}

.map-url-field {
  margin-top: 4px;
}

.map-picker-hint {
  font-size: 12px;
  color: #999;
  padding: 4px 16px 8px;
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

/* 我的地址 */
.addr-summary-text {
  font-size: 12px;
  color: #666;
  line-height: 1.6;
}

.addr-summary-empty {
  font-size: 12px;
  color: #aaa;
}

.addr-popup {
  display: flex;
  flex-direction: column;
  max-height: 88vh;
}

.addr-popup-title {
  text-align: center;
  padding: 16px 16px 8px;
  font-size: 16px;
  font-weight: 600;
  color: #1a1a1a;
  border-bottom: 1px solid #f5f5f5;
  flex-shrink: 0;
}

.addr-popup-body {
  overflow-y: auto;
  flex: 1;
  padding-bottom: 8px;
}

.map-set-text {
  color: #1d4ed8;
  font-size: 13px;
}

.addr-popup-footer {
  padding: 12px 16px calc(12px + env(safe-area-inset-bottom, 0px));
  border-top: 1px solid #f5f5f5;
  flex-shrink: 0;
}

.telegram-hint {
  padding: 8px 16px 12px;
  font-size: 12px;
  color: #999;
  line-height: 1.5;
}

/* 审核状态横幅 */
.status-banner {
  margin: 0 12px 8px;
}
.banner {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px 16px;
  border-radius: 8px;
  margin-top: 8px;
}
.banner.pending {
  background: #FFF7E6;
  border: 1px solid #FFD591;
  color: #AD6800;
}
.banner.rejected {
  background: #FFF1F0;
  border: 1px solid #FFA39E;
  color: #CF1322;
}
.banner-text strong {
  display: block;
  font-size: 14px;
  margin-bottom: 4px;
}
.banner-text p {
  font-size: 12px;
  margin: 0;
  opacity: 0.85;
}

/* 提交审核按钮 */
.submit-profile-section {
  padding: 16px;
}
.submit-hint {
  text-align: center;
  font-size: 12px;
  color: #999;
  margin: 8px 0 0;
}
</style>
