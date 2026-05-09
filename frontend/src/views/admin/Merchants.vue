<template>
  <div class="page-container">
    <div class="page-header">
      <h2>{{ $t('admin.merchants') }}</h2>
      <div class="header-btns">
        <el-button type="primary" @click="handleAdd('merchant')" :size="mobile ? 'small' : 'default'">
          <el-icon><plus /></el-icon>
          {{ $t('admin.addMerchant') }}
        </el-button>
        <el-button v-if="isSuperAdmin" type="warning" plain @click="handleAdd('admin')" :size="mobile ? 'small' : 'default'">
          <el-icon><plus /></el-icon>
          {{ $t('admin.addAdmin') }}
        </el-button>
      </div>
    </div>

    <el-row :gutter="mobile ? 10 : 24" class="mb-lg">
      <el-col :xs="12" :sm="12" :md="6">
        <div class="stat-card">
          <div class="stat-icon" style="background: #2563EB">
            <el-icon :size="mobile ? 18 : 24"><User /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ merchants.length }}</div>
            <div class="stat-label">{{ $t('admin.totalUsers') }}</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="12" :md="6">
        <div class="stat-card">
          <div class="stat-icon" style="background: #DC2626">
            <el-icon :size="mobile ? 18 : 24"><Avatar /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ adminCount }}</div>
            <div class="stat-label">{{ $t('admin.adminUsers') }}</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="12" :md="6">
        <div class="stat-card">
          <div class="stat-icon" style="background: #16A34A">
            <el-icon :size="mobile ? 18 : 24"><Shop /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ merchantCount }}</div>
            <div class="stat-label">{{ $t('admin.merchantUsers') }}</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="12" :md="6">
        <div class="stat-card">
          <div class="stat-icon" style="background: #0891B2">
            <el-icon :size="mobile ? 18 : 24"><CircleCheck /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ activeCount }}</div>
            <div class="stat-label">{{ $t('admin.activeUsers') }}</div>
          </div>
        </div>
      </el-col>
    </el-row>

    <div class="content-card">

      <el-table v-if="!mobile" v-loading="loading" :data="filteredUsers" stripe class="modern-user-table">
        <el-table-column :label="$t('admin.userInfoColumn')" min-width="270">
          <template #default="{ row }">
            <div class="user-primary-cell">
              <div class="user-id-badge">#{{ row.id }}</div>
              <div class="user-main-block">
                <div class="user-main-top">
                  <span class="user-name">{{ row.full_name || '—' }}</span>
                  <el-tag :type="getRoleType(row.role)" size="small">{{ getRoleText(row.role) }}</el-tag>
                  <el-tag v-if="row.is_super_admin" type="warning" size="small">{{ $t('admin.superAdmin') }}</el-tag>
                </div>
                <div class="user-sub-line">
                  <span class="user-account">{{ row.username }}</span>
                  <span v-if="row.phone && row.phone !== row.username" class="user-extra">{{ row.phone }}</span>
                </div>
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column :label="$t('admin.contactAddressColumn')" min-width="260">
          <template #default="{ row }">
            <div class="stack-cell">
              <div class="stack-line emphasis">{{ row.address || '—' }}</div>
              <div class="stack-line muted">
                <a v-if="row.location_url" :href="row.location_url" target="_blank" class="map-link" @click.stop>{{ $t('profile.viewMap') }}</a>
                <span v-else>—</span>
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column :label="$t('admin.billingInfoColumn')" min-width="180">
          <template #default="{ row }">
            <div v-if="row.role === 'merchant'" class="stack-cell">
              <div class="stack-line emphasis">${{ row.credit_limit || 0 }}</div>
              <div class="stack-line muted">{{ getBillingText(row) }}</div>
            </div>
            <div v-else class="stack-cell">
              <div class="stack-line emphasis">—</div>
            </div>
          </template>
        </el-table-column>

        <el-table-column :label="$t('admin.accountStateColumn')" min-width="120">
          <template #default="{ row }">
            <div class="status-stack">
              <el-tag :type="row.is_active ? 'success' : 'info'" size="small">{{ row.is_active ? $t('common.enabled') : $t('common.disabled') }}</el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column :label="$t('common.operation')" width="240" fixed="right">
          <template #default="{ row }">
            <div class="action-group action-chip-group">
              <el-button v-if="canEditRow(row)" type="primary" plain size="small" @click="handleEdit(row)">{{ $t('common.edit') }}</el-button>
              <el-button
                v-if="canManageSuperAdminRow(row)"
                :type="row.is_super_admin ? 'warning' : 'success'"
                plain
                size="small"
                @click="handleToggleSuperAdmin(row, !row.is_super_admin)"
              >{{ row.is_super_admin ? $t('admin.demoteSuperAdmin') : $t('admin.promoteSuperAdmin') }}</el-button>
              <el-button v-if="canDeleteRow(row)" type="danger" plain size="small" @click="handleDeleteUser(row)">{{ $t('common.delete') }}</el-button>
              <el-switch
                v-if="canToggleActive(row)"
                :model-value="row.is_active"
                size="small"
                @change="(val) => handleToggleActive(row, val)"
              />
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div v-else v-loading="loading" class="mobile-card-list">
        <div v-for="row in filteredUsers" :key="row.id" class="user-card modern-user-card" @click="handleEdit(row)">
          <div class="card-top-v2">
            <div class="card-meta-v2">
              <div class="card-name-v2">{{ row.full_name || row.username }}</div>
              <div class="card-sub-v2">
                <span v-if="row.phone" class="card-phone">{{ row.phone }}</span>
                <span v-else-if="row.username" class="card-phone">{{ row.username }}</span>
                <span v-if="row.telegram_id" class="card-tg-id">· TG: {{ row.telegram_id }}</span>
              </div>
            </div>
            <el-switch
              v-if="canToggleActive(row)"
              :model-value="row.is_active"
              @change="(val) => handleToggleActive(row, val)"
              @click.stop
            />
          </div>
          <div class="card-tags-v2">
            <el-tag :type="getRoleType(row.role)" size="small">{{ getRoleText(row.role) }}</el-tag>
            <el-tag v-if="row.is_super_admin" type="warning" size="small">{{ $t('admin.superAdmin') }}</el-tag>
            <el-tag v-if="canToggleActive(row) && !row.is_active" type="info" size="small">{{ $t('common.disabled') }}</el-tag>
          </div>
          <div class="card-actions-v2">
            <el-button v-if="canEditRow(row)" type="primary" plain size="small" @click.stop="handleEdit(row)">{{ $t('common.edit') }}</el-button>
            <el-button
              v-if="canManageSuperAdminRow(row)"
              :type="row.is_super_admin ? 'warning' : 'success'"
              plain
              size="small"
              @click.stop="handleToggleSuperAdmin(row, !row.is_super_admin)"
            >{{ row.is_super_admin ? $t('admin.demoteSuperAdmin') : $t('admin.promoteSuperAdmin') }}</el-button>
            <el-button v-if="canDeleteRow(row)" type="danger" plain size="small" @click.stop="handleDeleteUser(row)">{{ $t('common.delete') }}</el-button>
          </div>
        </div>
        <el-empty v-if="!loading && filteredUsers.length === 0" />
      </div>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      :width="mobile ? '94vw' : '560px'"
      :fullscreen="mobile"
      @open="onDialogOpen"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="computedRules"
        :label-width="mobile ? '80px' : '100px'"
        :label-position="mobile ? 'top' : 'right'"
      >
        <el-alert
          v-if="!isEdit && isSuperAdmin"
          class="dialog-role-tip"
          type="info"
          :closable="false"
          show-icon
          :description="$t('admin.dialogRoleTip')"
        />
        <el-form-item v-if="isEdit" :label="$t('profile.phone') + '/' + $t('login.account')">
          <el-input :model-value="form.username" disabled />
        </el-form-item>
        <el-form-item :label="$t('profile.name')" prop="full_name">
          <el-input v-model="form.full_name" />
        </el-form-item>
        <el-form-item :label="$t('product.status')" prop="role">
          <el-select v-model="form.role" :disabled="isEdit" style="width: 100%;">
            <el-option :label="$t('role.merchant')" value="merchant" />
            <el-option v-if="isSuperAdmin" :label="$t('role.admin')" value="admin" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="!isEdit" :label="$t('profile.phone')" prop="phone" :required="form.role === 'merchant'">
          <el-input v-model="form.phone" :placeholder="form.role === 'merchant' ? $t('admin.phonePlaceholder') : ''" />
        </el-form-item>
        <el-form-item :label="$t('profile.address')" prop="address">
          <el-input v-model="form.address" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item :label="$t('profile.locationUrl')" prop="location_url">
          <div style="display:flex;gap:8px;align-items:center;width:100%">
            <el-input v-model="form.location_url" :placeholder="$t('profile.locationUrlPlaceholder')" style="flex:1">
              <template #append v-if="form.location_url">
                <a :href="form.location_url" target="_blank" style="color: #409eff; text-decoration: none;">{{ $t('profile.viewMap') }}</a>
              </template>
            </el-input>
            <el-button size="small" @click="openAdminMapPicker">📍 {{ $t('profile.pickLocation') }}</el-button>
          </div>
        </el-form-item>
        <el-form-item v-if="form.role === 'merchant'" :label="$t('admin.allowMonthlyBilling')">
          <el-switch v-model="form.allow_credit" />
        </el-form-item>
        <el-form-item v-if="form.role === 'merchant'" :label="$t('admin.creditLimit')" prop="credit_limit">
          <el-input-number v-model="form.credit_limit" :min="0" :step="100" :precision="2" style="width: 100%;" disabled />
        </el-form-item>
        <el-form-item v-if="form.role === 'merchant'" :label="$t('admin.billingDay')" prop="billing_cycle_days">
          <el-input-number v-model="form.billing_cycle_days" :min="1" :max="365" :step="1" :placeholder="$t('admin.billingDayPlaceholder')" controls-position="right" style="width: 100%;" />
        </el-form-item>
        <el-form-item v-if="isEdit && !form.is_super_admin" :label="$t('product.status')">
          <el-switch v-model="form.is_active" :active-text="$t('product.onSale')" :inactive-text="$t('product.offSale')" />
        </el-form-item>
        <template v-if="form.role === 'admin' && isEdit">
          <el-divider>{{ $t('admin.telegramSettings') }}</el-divider>
          <el-form-item :label="$t('admin.telegramId')" prop="telegram_id">
            <el-input v-model="form.telegram_id" :placeholder="$t('admin.telegramIdPlaceholder')" clearable />
          </el-form-item>
          <div class="telegram-tip">
            <el-icon><InfoFilled /></el-icon>
            <span>{{ $t('admin.telegramTip') }}</span>
          </div>
        </template>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="newAccountVisible" :title="$t('admin.userAdded')" :width="mobile ? '92vw' : '380px'" center :close-on-click-modal="false">
      <div class="new-account-info">
        <p class="account-tip">{{ $t('admin.userCreatedTip') }}</p>
        <div class="account-box">
          <span class="account-label">{{ $t('login.account') }}</span>
          <span class="account-number">{{ newAccountNumber }}</span>
        </div>
        <p class="account-pwd-tip">{{ $t('admin.temporaryPassword') }}: <b>{{ newAccountPassword }}</b></p>
      </div>
      <template #footer>
        <el-button type="primary" @click="newAccountVisible = false">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- 地图选点对话框 -->
    <el-dialog v-model="showAdminMapPicker" :title="$t('profile.pickLocation')" :width="mobile ? '94vw' : '560px'" :fullscreen="false" destroy-on-close :close-on-click-modal="false">
      <div class="map-picker-wrap">
        <div style="display:flex;justify-content:flex-end;margin-bottom:8px">
          <el-button size="small" type="primary" :loading="adminLocating" @click="adminLocateMe">
            📍 {{ adminLocating ? $t('profile.locating') : $t('profile.locateMe') }}
          </el-button>
        </div>
        <div ref="adminMapContainerRef" style="width:100%;height:320px;background:#eee;border-radius:6px"></div>
        <el-input v-model="adminPickedUrl" :placeholder="$t('profile.locationUrlPlaceholder')" clearable style="margin-top:10px" />
        <div style="font-size:12px;color:#999;margin-top:6px">{{ $t('profile.locationPickerHint') }}</div>
      </div>
      <template #footer>
        <el-button @click="showAdminMapPicker = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="confirmAdminMapPick">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { Plus, InfoFilled, User, Avatar, Shop, CircleCheck } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus/es/components/message/index'
import { getUserList, register, updateUser, deleteUser, approveUser, setUserSuperAdmin } from '@/api'
import { ElMessageBox } from 'element-plus/es/components/message-box/index'
import { getRoleText } from '@/utils/format'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const isSuperAdmin = computed(() => userStore.isSuperAdmin)

// ──────────────── 地图选点 ────────────────
const GMAP_KEY = 'AIzaSyBLMXrpizbEE4f36sUCerOasHVWM8Doumc'
const showAdminMapPicker = ref(false)
const adminPickedUrl = ref('')
const adminLocating = ref(false)
const adminMapContainerRef = ref(null)
let adminMapInstance = null
let adminMapMarker = null

const loadGMapScript = () => {
  if (document.getElementById('google-maps-script')) {
    return window.google?.maps ? Promise.resolve() : new Promise(r => { const t = setInterval(() => { if (window.google?.maps) { clearInterval(t); r() } }, 100) })
  }
  return new Promise((resolve, reject) => {
    const s = document.createElement('script')
    s.id = 'google-maps-script'
    s.src = `https://maps.googleapis.com/maps/api/js?key=${GMAP_KEY}&language=zh-CN`
    s.async = true; s.defer = true; s.onload = resolve; s.onerror = reject
    document.head.appendChild(s)
  })
}

const placeAdminMarker = (latlng) => {
  if (!adminMapInstance) return
  if (adminMapMarker) adminMapMarker.setMap(null)
  adminMapMarker = new window.google.maps.Marker({ position: latlng, map: adminMapInstance })
  adminMapInstance.panTo(latlng)
  const lat = (typeof latlng.lat === 'function' ? latlng.lat() : latlng.lat).toFixed(6)
  const lng = (typeof latlng.lng === 'function' ? latlng.lng() : latlng.lng).toFixed(6)
  adminPickedUrl.value = `https://maps.google.com/?q=${lat},${lng}`
}

const adminLocateMe = () => {
  if (!navigator.geolocation) { ElMessage.warning('浏览器不支持定位'); return }
  adminLocating.value = true
  navigator.geolocation.getCurrentPosition(
    (pos) => {
      adminLocating.value = false
      const ll = { lat: pos.coords.latitude, lng: pos.coords.longitude }
      if (adminMapInstance) { adminMapInstance.setCenter(ll); adminMapInstance.setZoom(17); placeAdminMarker(ll) }
    },
    () => { adminLocating.value = false; ElMessage.warning('无法定位') },
    { enableHighAccuracy: true, timeout: 10000 }
  )
}

const openAdminMapPicker = async () => {
  adminPickedUrl.value = form.location_url || ''
  showAdminMapPicker.value = true
  await nextTick()
  try {
    await loadGMapScript()
    await nextTick()
    const container = adminMapContainerRef.value
    if (!container || !window.google?.maps) return
    const defaultCenter = { lat: 11.5564, lng: 104.9282 }
    let initCenter = defaultCenter
    const existing = adminPickedUrl.value
    if (existing) {
      const m = existing.match(/[?&]q=([\-\d.]+),([\-\d.]+)/) || existing.match(/@([\-\d.]+),([\-\d.]+)/)
      if (m) initCenter = { lat: parseFloat(m[1]), lng: parseFloat(m[2]) }
    }
    adminMapInstance = new window.google.maps.Map(container, {
      center: initCenter, zoom: 15,
      mapTypeControl: false, streetViewControl: false, fullscreenControl: false,
      gestureHandling: 'greedy',
    })
    if (existing && initCenter !== defaultCenter) {
      adminMapMarker = new window.google.maps.Marker({ position: initCenter, map: adminMapInstance })
    } else if (!existing && navigator.geolocation) {
      adminLocating.value = true
      navigator.geolocation.getCurrentPosition(
        (pos) => { adminLocating.value = false; adminMapInstance?.setCenter({ lat: pos.coords.latitude, lng: pos.coords.longitude }); adminMapInstance?.setZoom(17) },
        () => { adminLocating.value = false }, { enableHighAccuracy: true, timeout: 8000 }
      )
    }
    adminMapInstance.addListener('click', (e) => placeAdminMarker(e.latLng))
  } catch { ElMessage.warning('地图加载失败') }
}

const confirmAdminMapPick = () => {
  form.location_url = adminPickedUrl.value
  showAdminMapPicker.value = false
  adminMapInstance = null; adminMapMarker = null
}
// ───────────────────────────────────────────────

const { t } = useI18n()

// 移动端检测
const mobile = ref(window.innerWidth < 768)
const onResize = () => { mobile.value = window.innerWidth < 768 }
onMounted(() => window.addEventListener('resize', onResize))
onBeforeUnmount(() => window.removeEventListener('resize', onResize))
const loading = ref(false)
const merchants = ref([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref()

const form = reactive({
  id: null,
  username: '',
  password: '',
  full_name: '',
  role: 'merchant',
  phone: '',
  address: '',
  location_url: '',
  credit_limit: 0,
  billing_cycle_days: null,
  allow_credit: false,
  is_active: true,
  telegram_id: '',
  is_super_admin: false,
})

// 新用户账号展示
const newAccountVisible = ref(false)
const newAccountNumber = ref('')
const newAccountPassword = ref('')

const computedRules = computed(() => {
  const base = {
    full_name: [{ required: true, message: t('admin.fullNameRequired'), trigger: 'blur' }],
    role: [{ required: true, message: t('admin.roleRequired'), trigger: 'change' }],
  }
  // 商户必须填写手机号(手机号即账号)
  if (form.role === 'merchant' && !isEdit.value) {
    base.phone = [{ required: true, message: t('admin.phoneRequired'), trigger: 'blur' }]
  }
  return base
})

const getRoleType = (role) => {
  const map = { admin: 'danger', merchant: 'success' }
  return map[role] || ''
}

const getApprovalType = (status) => {
  const map = { pending: 'warning', approved: 'success', rejected: 'danger' }
  return map[status] || 'info'
}

const getApprovalText = (status) => {
  const map = {
    pending: t('admin.statusPending'),
    approved: t('admin.statusApproved'),
    rejected: t('admin.statusRejected'),
  }
  return map[status] || status || '—'
}

const isProtectedUser = (row) => !!row?.is_super_admin

const canEditRow = (row) => row.role !== 'admin' || isSuperAdmin.value || row.id === userStore.userInfo?.id

const canToggleActive = (row) => !isProtectedUser(row) && (row.role !== 'admin' || isSuperAdmin.value)

const canDeleteRow = (row) => isSuperAdmin.value && !isProtectedUser(row)

const canManageSuperAdminRow = (row) => isSuperAdmin.value && row.role === 'admin' && row.id !== userStore.userInfo?.id

const pendingCount = computed(() => merchants.value.filter((row) => row.role === 'merchant' && row.approval_status === 'pending').length)

const adminCount = computed(() => merchants.value.filter((row) => row.role === 'admin').length)

const merchantCount = computed(() => merchants.value.filter((row) => row.role === 'merchant').length)

const activeCount = computed(() => merchants.value.filter((row) => row.is_active).length)

const searchKeyword = ref('')

const filteredUsers = computed(() => merchants.value)

const dialogTitle = computed(() => {
  const roleText = form.role === 'admin' ? t('role.admin') : t('role.merchant')
  return isEdit.value ? t('admin.editUserWithRole', { role: roleText }) : t('admin.addUserWithRole', { role: roleText })
})

const getBillingText = (row) => {
  if (row.role !== 'merchant') {
    return t('admin.adminAccountHint')
  }

  if (!row.allow_credit) {
    return t('admin.cashOnly')
  }

  return row.billing_cycle_days ? `${row.billing_cycle_days}${t('admin.billingDayUnit')}` : t('admin.allowMonthlyBilling')
}

const loadMerchants = async () => {
  loading.value = true
  try {
    const data = await getUserList()
    merchants.value = data
  } catch (error) {
    console.error('加载用户失败:', error)
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  Object.assign(form, {
    id: null,
    username: '',
    password: '',
    full_name: '',
    role: 'merchant',
    phone: '',
    address: '',
    location_url: '',
    credit_limit: 0,
    billing_cycle_days: null,
    allow_credit: false,
    is_active: true,
    telegram_id: '',
    is_super_admin: false,
  })
  newAccountPassword.value = ''
  formRef.value?.clearValidate()
}

const handleAdd = (role = 'merchant') => {
  resetForm()
  form.role = role
  isEdit.value = false
  dialogVisible.value = true
}

const handleEdit = (row) => {
  Object.assign(form, {
    id: row.id,
    username: row.username,
    password: '',
    full_name: row.full_name || '',
    role: row.role,
    phone: row.phone || '',
    address: row.address || '',
    location_url: row.location_url || '',
    credit_limit: row.credit_limit || 0,
    billing_cycle_days: row.billing_cycle_days || null,
    allow_credit: row.allow_credit || false,
    is_active: row.is_active,
    telegram_id: row.telegram_id || '',
    is_super_admin: !!row.is_super_admin,
  })
  isEdit.value = true
  dialogVisible.value = true
}

// 快速切换启用/禁用
const handleToggleActive = async (row, val) => {
  try {
    await updateUser(row.id, { is_active: val })
    row.is_active = val
    ElMessage.success(val ? t('common.enabled') : t('common.disabled'))
  } catch (error) {
    ElMessage.error(t('common.operationFailed'))
  }
}

const onDialogOpen = () => {
  // dialog 打开后清除之前的校验状态
  setTimeout(() => {
    formRef.value?.clearValidate()
  }, 50)
}

const handleSubmit = async () => {
  if (!formRef.value) {
    ElMessage.error('表单未初始化，请重新打开')
    return
  }
  try {
    await formRef.value.validate()
  } catch {
    return // 校验不通过
  }
  submitting.value = true
  try {
    if (isEdit.value) {
      const payload = {
        full_name: form.full_name,
        phone: form.phone,
        address: form.address,
        location_url: form.location_url,
        credit_limit: form.credit_limit,
        billing_cycle_days: form.billing_cycle_days,
        allow_credit: form.allow_credit,
        is_active: form.is_active,
        telegram_id: form.telegram_id || null,
      }
      await updateUser(form.id, payload)
      ElMessage.success(t('admin.userUpdated'))
      // 如果编辑的是自己，同步更新本地用户信息
      if (form.id === userStore.userInfo?.id) {
        await userStore.fetchUserInfo()
      }
    } else {
      const res = await register({
        full_name: form.full_name,
        role: form.role,
        phone: form.phone,
        address: form.address,
        location_url: form.location_url,
        credit_limit: form.credit_limit,
        billing_cycle_days: form.billing_cycle_days,
        allow_credit: form.allow_credit,
      })
      // 展示自动生成的账号
      newAccountNumber.value = res.user.username
      newAccountPassword.value = res.temporary_password
      newAccountVisible.value = true
    }
    dialogVisible.value = false
    loadMerchants()
  } catch (error) {
    console.error('提交失败:', error)
    // request.js 拦截器已显示错误消息
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadMerchants()
})

const handleToggleSuperAdmin = async (row, enabled) => {
  try {
    await ElMessageBox.confirm(
      enabled
        ? t('admin.promoteSuperAdminConfirm', { name: row.full_name || row.username })
        : t('admin.demoteSuperAdminConfirm', { name: row.full_name || row.username }),
      t('admin.hint'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: enabled ? 'warning' : 'info',
      }
    )
    await setUserSuperAdmin(row.id, enabled)
    ElMessage.success(t('admin.superAdminUpdated'))
    loadMerchants()
  } catch {
    // 用户取消
  }
}

// 删除用户
const handleDeleteUser = async (row) => {
  try {
    await ElMessageBox.confirm(
      t('admin.deleteUserConfirm', { name: row.full_name || row.username }),
      t('admin.hint'),
      { confirmButtonText: t('common.confirm'), cancelButtonText: t('common.cancel'), type: 'error' }
    )
    await deleteUser(row.id)
    ElMessage.success(t('admin.userDeleted'))
    loadMerchants()
  } catch {
    // 用户取消或删除失败
  }
}
</script>

<style scoped>
.header-btns {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.card-detail-rows {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin: 8px 0 4px;
  padding: 8px 10px;
  background: #f8fafc;
  border-radius: 8px;
}

.card-detail-row {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  font-size: 12px;
  color: #374151;
  line-height: 1.5;
}

.detail-icon {
  flex-shrink: 0;
  font-size: 13px;
  line-height: 1.5;
}

.detail-val {
  flex: 1;
  min-width: 0;
  word-break: break-word;
}

.text-clamp {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.modern-user-table {
  --el-table-border-color: #eef2f7;
  --el-table-header-bg-color: #f8fafc;
  --el-table-row-hover-bg-color: #f8fafc;
}

.dialog-role-tip {
  margin-bottom: 16px;
}

.user-primary-cell {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.user-id-badge {
  min-width: 42px;
  height: 42px;
  border-radius: 12px;
  background: #eff6ff;
  color: #1d4ed8;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
}

.user-main-block {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.user-main-top {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.user-name {
  font-size: 15px;
  font-weight: 700;
  color: #111827;
}

.user-sub-line {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  color: #6b7280;
  font-size: 13px;
}

.user-account {
  font-weight: 600;
  color: #374151;
}

.user-extra {
  color: #9ca3af;
}

.stack-cell {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.stack-line {
  line-height: 1.5;
}

.stack-line.emphasis {
  color: #111827;
  font-weight: 600;
}

.stack-line.muted {
  color: #6b7280;
  font-size: 13px;
}

.status-stack {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 6px;
}

.action-group {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.action-group :deep(.el-button) {
  margin-left: 0;
  margin-right: 0;
}

.action-chip-group :deep(.el-button) {
  border-radius: 999px;
  padding-left: 12px;
  padding-right: 12px;
}

:deep(.filter-group .el-radio-group) {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

:deep(.filter-group .el-radio-button__inner) {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  white-space: nowrap;
}

.count-badge {
  display: inline-flex;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: #f56c6c;
  color: #fff;
  font-size: 12px;
  line-height: 1;
  margin-left: 4px;
}

/* ========== 移动端卡片列表 ========== */
.mobile-card-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.user-card {
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 14px;
  padding: 12px;
  cursor: pointer;
  transition: border-color 0.2s, transform 0.2s;
  overflow: hidden;
}

.user-card:hover {
  border-color: #cbd5e1;
}

.modern-user-card:active {
  transform: scale(0.995);
}

/* ====== 新版紧凑卡片 ====== */
.user-card {
  border: none;
  box-shadow: 0 1px 4px rgba(0,0,0,0.08), 0 0 0 1px rgba(0,0,0,0.04);
  border-radius: 12px;
  padding: 10px 12px 8px;
  background: #fff;
  cursor: pointer;
  transition: box-shadow 0.2s, transform 0.15s;
}
.user-card:hover {
  box-shadow: 0 3px 10px rgba(0,0,0,0.12), 0 0 0 1px rgba(0,0,0,0.05);
}
.card-top-v2 {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}
.card-meta-v2 {
  min-width: 0;
  flex: 1;
}
.card-name-v2 {
  font-size: 15px;
  font-weight: 700;
  color: #111827;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.3;
}
.card-sub-v2 {
  font-size: 12px;
  color: #6b7280;
  margin-top: 2px;
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}
.card-phone {
  font-size: 13px;
  color: #1d4ed8;
  font-weight: 600;
}
.card-tg-id {
  font-size: 11px;
  color: #9ca3af;
}
.card-tags-v2 {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 6px;
}
.card-actions-v2 {
  display: flex;
  gap: 6px;
  margin-top: 8px;
  flex-wrap: wrap;
}
.card-actions-v2 :deep(.el-button) {
  padding: 4px 12px;
  height: 28px;
  font-size: 12px;
}

.card-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.card-user-info {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.card-user-info-modern {
  flex-direction: column;
  min-width: 0;
  gap: 6px;
}

.card-user-meta {
  display: flex;
  flex-direction: column;
  min-width: 0;
  gap: 2px;
}

.card-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.card-username {
  font-size: 15px;
  font-weight: 700;
  color: #111827;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100%;
}

.card-account {
  color: #6b7280;
  font-size: 12px;
}

.card-body {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.compact-card .card-body {
  gap: 4px;
}

.card-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  font-size: 13px;
  line-height: 1.5;
  gap: 12px;
}

.card-label {
  color: #6b7280;
  flex-shrink: 0;
  min-width: 74px;
}

.card-value {
  color: #111827;
  text-align: right;
}

.card-value-text {
  white-space: normal;
  word-break: break-word;
}

.single-line {
  align-items: flex-start;
}

.main-name {
  font-weight: 600;
  max-width: 65%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.reject-line {
  margin-top: 2px;
}

.text-danger {
  color: #f56c6c;
}

.inline-danger {
  color: #dc2626;
  font-size: 12px;
  line-height: 1.5;
}

.text-ellipsis {
  max-width: 60%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.map-link {
  color: #409eff;
  text-decoration: none;
  font-size: 13px;
  cursor: pointer;
}

.map-link:hover {
  text-decoration: underline;
}

.card-footer-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 10px;
}

/* Telegram 提示 */
.telegram-tip {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  padding: 10px 12px;
  background: #f0f9ff;
  border-radius: 6px;
  font-size: 12px;
  color: #909399;
  line-height: 1.5;
  margin-top: -8px;
  margin-bottom: 16px;
}

.telegram-tip .el-icon {
  color: #409eff;
  margin-top: 2px;
  flex-shrink: 0;
}

/* ========== 移动端适配 ========== */
@media (max-width: 767px) {
  .stat-card {
    margin-bottom: 10px;
  }

  .header-btns {
    width: 100%;
  }

  .header-btns :deep(.el-button) {
    flex: 1;
    min-width: 0;
  }

  .card-user-info {
    flex-wrap: wrap;
  }

  .user-card {
    padding: 10px 12px;
    border-radius: 12px;
  }

  .card-top {
    margin-bottom: 0;
  }

  .card-tags :deep(.el-tag) {
    --el-tag-size: 20px;
  }

  .card-footer-wrap {
    margin-top: 8px;
    gap: 4px 6px;
  }

  .action-chip-group :deep(.el-button) {
    padding-left: 10px;
    padding-right: 10px;
  }

  :deep(.el-dialog) {
    margin-top: 0 !important;
    border-radius: 12px 12px 0 0;
  }

  :deep(.el-dialog__body) {
    max-height: 70vh;
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
  }

  :deep(.el-form-item__label) {
    font-size: 13px;
    padding-bottom: 4px;
  }

  :deep(.el-input-number) {
    width: 100% !important;
  }

  :deep(.el-select) {
    width: 100% !important;
  }
}

/* 新用户账号展示 */
.new-account-info {
  text-align: center;
}

.account-tip {
  color: #67c23a;
  font-size: 14px;
  margin-bottom: 16px;
}

.account-box {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 16px;
  background: #f0f9ff;
  border-radius: 8px;
  margin-bottom: 12px;
}

.account-label {
  color: #909399;
  font-size: 14px;
}

.account-number {
  font-size: 24px;
  font-weight: 700;
  color: #1D4ED8;
  letter-spacing: 2px;
}

.account-pwd-tip {
  color: #909399;
  font-size: 13px;
}
</style>
