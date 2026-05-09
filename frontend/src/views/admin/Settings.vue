<template>
  <div class="settings-page">
    <div class="page-header">
      <h2>{{ $t('settings.title') }}</h2>
    </div>

    <el-card class="setting-card" shadow="never">
      <template #header>
        <span>{{ $t('settings.roleChatIds') }}</span>
      </template>
      <el-alert
        type="info"
        :closable="false"
        show-icon
        :title="$t('settings.roleChatIdsTip')"
        style="margin-bottom: 16px;"
      />
      <el-form
        v-loading="loadingRoles"
        :model="rolesForm"
        label-width="160px"
        label-position="left"
      >
        <el-form-item :label="$t('settings.groupChatId')">
          <el-input
            v-model="rolesForm.group_chat_id"
            :placeholder="$t('settings.chatIdPlaceholder')"
            clearable
            style="max-width: 300px;"
          />
          <el-button
            link
            type="primary"
            style="margin-left: 8px;"
            @click="openChatPicker"
          >
            {{ $t('settings.detectChat') }}
          </el-button>
        </el-form-item>
        <el-form-item :label="$t('settings.deliveryGroupLink')">
          <el-input
            v-model="rolesForm.delivery_group_link"
            placeholder="https://t.me/+..."
            clearable
            style="max-width: 300px;"
          />
          <a
            v-if="rolesForm.delivery_group_link"
            :href="rolesForm.delivery_group_link"
            target="_blank"
            style="margin-left:8px;color:#409eff;font-size:13px;"
          >{{ $t('settings.openGroup') }}</a>
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="savingRoles"
            @click="saveRoles"
          >
            {{ $t('common.save') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="setting-card" shadow="never">
      <template #header>
        <span>{{ $t('settings.deliveryFee') }}</span>
      </template>
      <el-form
        v-loading="loadingFee"
        :model="feeForm"
        label-width="160px"
        label-position="left"
      >
        <el-form-item :label="$t('settings.freeDistanceKm')">
          <el-input-number
            v-model="feeForm.free_distance_km"
            :min="0"
            :step="0.5"
            :precision="2"
          />
          <span class="unit">km</span>
        </el-form-item>
        <el-form-item :label="$t('settings.feePerExtraKm')">
          <el-input-number
            v-model="feeForm.fee_per_extra_km_usd"
            :min="0"
            :step="0.1"
            :precision="2"
          />
          <span class="unit">USD / km</span>
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="savingFee"
            @click="saveFee"
          >
            {{ $t('common.save') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Google Maps 配置 -->
    <el-card class="setting-card" shadow="never">
      <template #header>
        <span>{{ $t('settings.googleMaps') }}</span>
      </template>
      <el-alert
        type="info"
        :closable="false"
        show-icon
        :title="$t('settings.googleMapsTip')"
        style="margin-bottom: 16px;"
      />
      <el-form
        v-loading="loadingMaps"
        :model="mapsForm"
        label-width="160px"
        label-position="left"
      >
        <el-form-item :label="$t('settings.googleMapsKey')">
          <el-input
            v-model="mapsForm.api_key"
            :placeholder="$t('settings.googleMapsKeyPlaceholder')"
            clearable
            show-password
            style="max-width: 420px;"
          />
        </el-form-item>
        <el-form-item :label="$t('settings.warehouseLocation')">
          <div style="width:100%;max-width:520px;">
            <div v-if="!mapsForm.api_key" style="color:#999;font-size:13px;padding:8px 0;">
              {{ $t('settings.warehouseLocationNeedKey') }}
            </div>
            <template v-else>
              <div
                ref="mapPickerEl"
                style="width:100%;height:280px;border:1px solid #dcdfe6;border-radius:6px;margin-bottom:8px;"
              />
              <div style="display:flex;align-items:center;gap:12px;font-size:13px;">
                <span v-if="mapsForm.warehouse_lat && mapsForm.warehouse_lng" style="color:#555;">
                  {{ mapsForm.warehouse_lat }}, {{ mapsForm.warehouse_lng }}
                </span>
                <span v-else style="color:#999;">{{ $t('settings.warehouseLocationTip') }}</span>
                <a
                  v-if="mapsForm.warehouse_lat && mapsForm.warehouse_lng"
                  :href="`https://maps.google.com/?q=${mapsForm.warehouse_lat},${mapsForm.warehouse_lng}`"
                  target="_blank"
                  style="color:#409eff;text-decoration:none;"
                >{{ $t('profile.viewMap') }}</a>
              </div>
            </template>
          </div>
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="savingMaps"
            @click="saveMaps"
          >
            {{ $t('common.save') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 联系方式设置 -->
    <el-card class="setting-card" shadow="never">
      <template #header>
        <span>{{ $t('settings.contactInfo') }}</span>
      </template>
      <el-form
        v-loading="loadingContact"
        :model="contactForm"
        label-width="160px"
        label-position="left"
      >
        <el-form-item :label="$t('settings.contactPhone')">
          <el-input
            v-model="contactForm.phone"
            placeholder="+855 12 345 678"
            clearable
            style="max-width: 300px;"
          />
        </el-form-item>
        <el-form-item :label="$t('settings.contactWhatsapp')">
          <el-input
            v-model="contactForm.whatsapp"
            placeholder="+855 12 345 678"
            clearable
            style="max-width: 300px;"
          />
        </el-form-item>
        <el-form-item :label="$t('settings.contactTelegram')">
          <el-input
            v-model="contactForm.telegram"
            placeholder="@username"
            clearable
            style="max-width: 300px;"
          />
        </el-form-item>
        <el-form-item :label="$t('settings.contactWechat')">
          <el-input
            v-model="contactForm.wechat"
            placeholder="WeChat ID"
            clearable
            style="max-width: 300px;"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="savingContact"
            @click="saveContact"
          >
            {{ $t('common.save') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 选择 Telegram 群/用户对话框 -->
    <el-dialog
      v-model="chatPickerVisible"
      :title="$t('settings.detectChatTitle')"
      width="600px"
    >
      <el-alert
        type="warning"
        :closable="false"
        show-icon
        :title="$t('settings.detectChatHelp')"
        style="margin-bottom: 12px;"
      />
      <div style="margin-bottom: 12px;">
        <el-button
          :loading="loadingChats"
          type="primary"
          plain
          @click="loadChats"
        >
          {{ $t('settings.refreshList') }}
        </el-button>
      </div>
      <div v-loading="loadingChats" class="chat-list">
        <div
          v-for="c in chats"
          :key="c.id"
          class="chat-item"
          @click="selectChat(c)"
        >
          <div class="chat-main">
            <div class="chat-title">
              <el-tag size="small" :type="chatTagType(c.type)">{{ c.type }}</el-tag>
              <span class="chat-name">{{ c.title }}</span>
              <span v-if="c.username" class="chat-username">@{{ c.username }}</span>
            </div>
            <div class="chat-id">{{ c.id }}</div>
            <div v-if="c.last_message" class="chat-msg">{{ c.last_message }}</div>
          </div>
          <el-button size="small" type="primary">
            {{ $t('settings.useThis') }}
          </el-button>
        </div>
        <div v-if="!loadingChats && chats.length === 0" class="empty">
          {{ $t('settings.noChats') }}
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  getRoleChatIds,
  updateRoleChatIds,
  getDeliveryFeeSettings,
  updateDeliveryFeeSettings,
  getTelegramRecentChats,
  getContactInfo,
  updateContactInfo,
  getGoogleMapsSettings,
  updateGoogleMapsSettings,
} from '@/api'

const { t } = useI18n()

const loadingRoles = ref(false)
const savingRoles = ref(false)
const rolesForm = ref({ group_chat_id: '', delivery_group_link: '' })

const loadingFee = ref(false)
const savingFee = ref(false)
const feeForm = ref({ free_distance_km: 3, fee_per_extra_km_usd: 0.5 })

// Telegram chat 检测
const chatPickerVisible = ref(false)
const loadingChats = ref(false)
const chats = ref([])

function chatTagType(t) {
  if (t === 'private') return 'info'
  if (t === 'group' || t === 'supergroup') return 'success'
  if (t === 'channel') return 'warning'
  return ''
}

function openChatPicker() {
  chatPickerVisible.value = true
  if (chats.value.length === 0) loadChats()
}

async function loadChats() {
  loadingChats.value = true
  try {
    const data = await getTelegramRecentChats()
    chats.value = data.chats || []
  } catch (e) {
    ElMessage.error(e?.message || 'Load failed')
  } finally {
    loadingChats.value = false
  }
}

function selectChat(c) {
  rolesForm.value.group_chat_id = c.id
  chatPickerVisible.value = false
  ElMessage.success(t('settings.chatSelected') || 'Selected')
}

async function loadRoles() {
  loadingRoles.value = true
  try {
    const data = await getRoleChatIds()
    rolesForm.value = {
      group_chat_id: data.group_chat_id || '',
      delivery_group_link: data.delivery_group_link || '',
    }
  } catch (e) {
    ElMessage.error(e?.message || 'Load failed')
  } finally {
    loadingRoles.value = false
  }
}

async function saveRoles() {
  savingRoles.value = true
  try {
    const data = await updateRoleChatIds({
      group_chat_id: rolesForm.value.group_chat_id || '',
      delivery_group_link: rolesForm.value.delivery_group_link || '',
    })
    rolesForm.value = {
      group_chat_id: data.group_chat_id || '',
      delivery_group_link: data.delivery_group_link || '',
    }
    ElMessage.success(t('common.saveSuccess') || 'Saved')
  } catch (e) {
    ElMessage.error(e?.message || 'Save failed')
  } finally {
    savingRoles.value = false
  }
}

async function loadFee() {
  loadingFee.value = true
  try {
    const data = await getDeliveryFeeSettings()
    feeForm.value = {
      free_distance_km: data.free_distance_km ?? 3,
      fee_per_extra_km_usd: data.fee_per_extra_km_usd ?? 0.5,
    }
  } catch (e) {
    ElMessage.error(e?.message || 'Load failed')
  } finally {
    loadingFee.value = false
  }
}

async function saveFee() {
  savingFee.value = true
  try {
    const data = await updateDeliveryFeeSettings({
      free_distance_km: feeForm.value.free_distance_km,
      fee_per_extra_km_usd: feeForm.value.fee_per_extra_km_usd,
    })
    feeForm.value = {
      free_distance_km: data.free_distance_km,
      fee_per_extra_km_usd: data.fee_per_extra_km_usd,
    }
    ElMessage.success(t('common.saveSuccess') || 'Saved')
  } catch (e) {
    ElMessage.error(e?.message || 'Save failed')
  } finally {
    savingFee.value = false
  }
}

onMounted(() => {
  loadRoles()
  loadFee()
  loadContact()
  loadMaps()
})

const loadingMaps = ref(false)
const savingMaps = ref(false)
const mapsForm = ref({ api_key: '', warehouse_lat: '', warehouse_lng: '' })

// ─── Google Maps 交互地图选点 ───
const mapPickerEl = ref(null)
let _googleMap = null
let _warehouseMarker = null

async function initMapPicker() {
  await nextTick()
  if (!mapPickerEl.value || !mapsForm.value.api_key) return
  if (!window.google?.maps) {
    await new Promise((resolve, reject) => {
      if (document.querySelector('#gmap-script')) { resolve(); return }
      const s = document.createElement('script')
      s.id = 'gmap-script'
      s.src = `https://maps.googleapis.com/maps/api/js?key=${mapsForm.value.api_key}`
      s.onload = resolve
      s.onerror = reject
      document.head.appendChild(s)
    })
  }
  const lat = parseFloat(mapsForm.value.warehouse_lat) || 11.5564
  const lng = parseFloat(mapsForm.value.warehouse_lng) || 104.9282
  _googleMap = new window.google.maps.Map(mapPickerEl.value, {
    center: { lat, lng },
    zoom: mapsForm.value.warehouse_lat ? 16 : 13,
    mapTypeControl: false,
    streetViewControl: false,
    fullscreenControl: false,
  })
  if (mapsForm.value.warehouse_lat && mapsForm.value.warehouse_lng) {
    _warehouseMarker = new window.google.maps.Marker({ position: { lat, lng }, map: _googleMap })
  }
  _googleMap.addListener('click', (e) => {
    const newLat = e.latLng.lat().toFixed(6)
    const newLng = e.latLng.lng().toFixed(6)
    mapsForm.value.warehouse_lat = newLat
    mapsForm.value.warehouse_lng = newLng
    const pos = { lat: parseFloat(newLat), lng: parseFloat(newLng) }
    if (_warehouseMarker) {
      _warehouseMarker.setPosition(pos)
    } else {
      _warehouseMarker = new window.google.maps.Marker({ position: pos, map: _googleMap })
    }
  })
}

// 当 api_key 填写完成且容器已渲染后初始化地图
watch(() => mapsForm.value.api_key, (val) => {
  if (val) initMapPicker()
})

async function loadMaps() {
  loadingMaps.value = true
  try {
    const data = await getGoogleMapsSettings()
    mapsForm.value = {
      api_key: data.api_key || '',
      warehouse_lat: data.warehouse_lat || '',
      warehouse_lng: data.warehouse_lng || '',
    }
    if (mapsForm.value.api_key) {
      initMapPicker()
    }
  } catch (e) {
    ElMessage.error(e?.message || 'Load failed')
  } finally {
    loadingMaps.value = false
  }
}

async function saveMaps() {
  savingMaps.value = true
  try {
    await updateGoogleMapsSettings(mapsForm.value)
    ElMessage.success(t('common.saveSuccess') || 'Saved')
  } catch (e) {
    ElMessage.error(e?.message || 'Save failed')
  } finally {
    savingMaps.value = false
  }
}

const loadingContact = ref(false)
const savingContact = ref(false)
const contactForm = ref({ phone: '', telegram: '', whatsapp: '', wechat: '' })

async function loadContact() {
  loadingContact.value = true
  try {
    const data = await getContactInfo()
    contactForm.value = { phone: data.phone || '', telegram: data.telegram || '', whatsapp: data.whatsapp || '', wechat: data.wechat || '' }
  } catch (e) {
    ElMessage.error(e?.message || 'Load failed')
  } finally {
    loadingContact.value = false
  }
}

async function saveContact() {
  savingContact.value = true
  try {
    const data = await updateContactInfo(contactForm.value)
    contactForm.value = { phone: data.phone || '', telegram: data.telegram || '', whatsapp: data.whatsapp || '', wechat: data.wechat || '' }
    ElMessage.success(t('settings.contactInfoSaved'))
  } catch (e) {
    ElMessage.error(e?.message || 'Save failed')
  } finally {
    savingContact.value = false
  }
}
</script>

<style scoped lang="scss">
.settings-page {
  padding: 16px;
}
.page-header {
  margin-bottom: 16px;
  h2 {
    margin: 0;
    color: #2b2b2b;
  }
}
.setting-card {
  margin-bottom: 16px;
  border-radius: 8px;
}
.unit {
  margin-left: 8px;
  color: #666;
}
.chat-list {
  max-height: 400px;
  overflow-y: auto;
}
.chat-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border: 1px solid #eee;
  border-radius: 6px;
  margin-bottom: 8px;
  cursor: pointer;
  transition: all 0.2s;
  &:hover {
    border-color: #2b2b2b;
    background: #fafafa;
  }
}
.chat-main {
  flex: 1;
  min-width: 0;
}
.chat-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
  color: #2b2b2b;
}
.chat-name {
  margin-left: 4px;
}
.chat-username {
  color: #888;
  font-weight: normal;
  font-size: 12px;
}
.chat-id {
  font-family: monospace;
  font-size: 12px;
  color: #666;
  margin-top: 2px;
}
.chat-msg {
  font-size: 12px;
  color: #999;
  margin-top: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.empty {
  text-align: center;
  color: #999;
  padding: 24px 0;
}
</style>
