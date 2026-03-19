<template>
  <div class="merchants-page">
    <div class="page-header">
      <h2>{{ $t('admin.merchants') }}</h2>
      <el-button type="primary" @click="handleAdd" :size="mobile ? 'small' : 'default'">
        <el-icon><plus /></el-icon>
        {{ $t('admin.addMerchant') }}
      </el-button>
    </div>

    <!-- 桌面端: 表格视图 -->
    <el-table v-if="!mobile" v-loading="loading" :data="merchants" border>
      <el-table-column label="ID" prop="id" width="60" />
      <el-table-column :label="$t('login.username')" prop="username" width="120" />
      <el-table-column :label="$t('profile.name')" prop="full_name" width="120" />
      <el-table-column :label="$t('profile.phone')" prop="phone" width="140" />
      <el-table-column :label="$t('profile.address')" prop="address" min-width="180" />
      <el-table-column :label="$t('profile.locationUrl')" width="100">
        <template #default="{ row }">
          <a v-if="row.location_url" :href="row.location_url" target="_blank" class="map-link" @click.stop>
            {{ $t('profile.viewMap') }}
          </a>
          <span v-else style="color: #ccc;">—</span>
        </template>
      </el-table-column>
      <el-table-column :label="$t('product.status')" width="100">
        <template #default="{ row }">
          <el-tag :type="getRoleType(row.role)" size="small">
            {{ getRoleText(row.role) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="$t('admin.creditLimit')" width="110">
        <template #default="{ row }">
          <span v-if="row.role === 'merchant'">${{ row.credit_limit || 0 }}</span>
          <span v-else style="color: #ccc;">—</span>
        </template>
      </el-table-column>
      <el-table-column :label="$t('admin.allowMonthlyBilling')" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.role === 'merchant'" :type="row.allow_monthly_billing ? 'success' : 'info'" size="small">
            {{ row.allow_monthly_billing ? $t('common.enabled') : $t('common.disabled') }}
          </el-tag>
          <span v-else style="color: #ccc;">—</span>
        </template>
      </el-table-column>
      <el-table-column :label="$t('admin.billingDay')" width="90">
        <template #default="{ row }">
          <span v-if="row.role === 'merchant' && row.billing_day">{{ row.billing_day }}号</span>
          <span v-else style="color: #ccc;">—</span>
        </template>
      </el-table-column>
      <el-table-column :label="$t('product.status')" width="80">
        <template #default="{ row }">
          <el-switch
            v-if="row.username !== '100001'"
            :model-value="row.is_active"
            size="small"
            @change="(val) => handleToggleActive(row, val)"
          />
          <el-tag v-else type="success" size="small">{{ $t('common.enabled') }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="$t('common.operation')" width="200" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link size="small" @click="handleEdit(row)">{{ $t('common.edit') }}</el-button>
          <el-button v-if="isSuperAdmin && row.username !== '100001'" type="warning" link size="small" @click="handleResetPassword(row)">{{ $t('admin.resetPassword') }}</el-button>
          <el-button v-if="isSuperAdmin && row.username !== '100001'" type="danger" link size="small" @click="handleDeleteUser(row)">{{ $t('common.delete') }}</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 移动端: 卡片列表 -->
    <div v-else v-loading="loading" class="mobile-card-list">
      <div v-for="row in merchants" :key="row.id" class="user-card" @click="handleEdit(row)">
        <div class="card-top">
          <div class="card-user-info">
            <span class="card-username">{{ row.username }}</span>
            <el-tag :type="getRoleType(row.role)" size="small">{{ getRoleText(row.role) }}</el-tag>
          </div>
          <el-switch
            v-if="row.username !== '100001'"
            :model-value="row.is_active"
            size="small"
            @change="(val) => handleToggleActive(row, val)"
            @click.stop
          />
          <el-tag v-else type="success" size="small">{{ $t('common.enabled') }}</el-tag>
        </div>
        <div class="card-body">
          <div v-if="row.full_name" class="card-row">
            <span class="card-label">{{ $t('profile.name') }}</span>
            <span class="card-value">{{ row.full_name }}</span>
          </div>
          <div v-if="row.phone" class="card-row">
            <span class="card-label">{{ $t('profile.phone') }}</span>
            <span class="card-value">{{ row.phone }}</span>
          </div>
          <div v-if="row.address" class="card-row">
            <span class="card-label">{{ $t('profile.address') }}</span>
            <span class="card-value text-ellipsis">{{ row.address }}</span>
          </div>
          <div v-if="row.location_url" class="card-row">
            <span class="card-label">{{ $t('profile.locationUrl') }}</span>
            <a :href="row.location_url" target="_blank" class="map-link" @click.stop>{{ $t('profile.viewMap') }}</a>
          </div>
          <div v-if="row.role === 'merchant'" class="card-row">
            <span class="card-label">{{ $t('admin.creditLimit') }}</span>
            <span class="card-value">${{ row.credit_limit || 0 }}</span>
          </div>
          <div v-if="row.role === 'merchant'" class="card-row">
            <span class="card-label">{{ $t('admin.allowMonthlyBilling') }}</span>
            <el-tag :type="row.allow_monthly_billing ? 'success' : 'info'" size="small">
              {{ row.allow_monthly_billing ? $t('common.enabled') : $t('common.disabled') }}
            </el-tag>
          </div>
          <div v-if="row.role === 'merchant' && row.billing_day" class="card-row">
            <span class="card-label">{{ $t('admin.billingDay') }}</span>
            <span class="card-value">{{ row.billing_day }}号</span>
          </div>
        </div>
        <div v-if="isSuperAdmin && row.username !== '100001'" class="card-footer">
          <el-button type="danger" size="small" link @click.stop="handleDeleteUser(row)">{{ $t('common.delete') }}</el-button>
        </div>
      </div>
      <el-empty v-if="!loading && merchants.length === 0" />
    </div>

    <!-- 添加/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? $t('admin.editUser') : $t('admin.addUser')"
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
        <el-form-item v-if="isEdit" :label="$t('login.account')">
          <el-input :model-value="form.username" disabled />
        </el-form-item>
        <el-form-item :label="$t('profile.name')" prop="full_name">
          <el-input v-model="form.full_name" />
        </el-form-item>
        <el-form-item :label="$t('product.status')" prop="role">
          <el-select v-model="form.role" :disabled="isEdit" style="width: 100%;">
            <el-option :label="$t('role.merchant')" value="merchant" />
            <el-option :label="$t('role.admin')" value="admin" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('profile.phone')" prop="phone">
          <el-input v-model="form.phone" />
        </el-form-item>
        <el-form-item :label="$t('profile.address')" prop="address">
          <el-input v-model="form.address" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item :label="$t('profile.locationUrl')" prop="location_url">
          <el-input v-model="form.location_url" :placeholder="$t('profile.locationUrlPlaceholder')">
            <template #append v-if="form.location_url">
              <a :href="form.location_url" target="_blank" style="color: #409eff; text-decoration: none;">{{ $t('profile.viewMap') }}</a>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item v-if="form.role === 'merchant'" :label="$t('admin.allowMonthlyBilling')">
          <el-switch v-model="form.allow_monthly_billing" />
        </el-form-item>
        <el-form-item v-if="form.role === 'merchant'" :label="$t('admin.creditLimit')" prop="credit_limit">
          <el-input-number v-model="form.credit_limit" :min="0" :step="100" :precision="2" style="width: 100%;" disabled />
        </el-form-item>
        <el-form-item v-if="form.role === 'merchant'" :label="$t('admin.billingDay')" prop="billing_day">
          <el-input-number v-model="form.billing_day" :min="1" :max="31" :step="1" :placeholder="$t('admin.billingDayPlaceholder')" controls-position="right" style="width: 100%;" />
        </el-form-item>
        <el-form-item v-if="isEdit && form.username !== '100001'" :label="$t('product.status')">
          <el-switch v-model="form.is_active" :active-text="$t('product.onSale')" :inactive-text="$t('product.offSale')" />
        </el-form-item>
        <!-- Telegram 设置 (仅管理员角色) -->
        <template v-if="form.role === 'admin' && isEdit">
          <el-divider>{{ $t('admin.telegramSettings') }}</el-divider>
          <el-form-item :label="$t('admin.telegramBotToken')" prop="telegram_bot_token">
            <el-input v-model="form.telegram_bot_token" :placeholder="$t('admin.telegramBotTokenPlaceholder')" clearable />
          </el-form-item>
          <el-form-item :label="$t('admin.telegramChatId')" prop="telegram_chat_id">
            <el-input v-model="form.telegram_chat_id" :placeholder="$t('admin.telegramChatIdPlaceholder')" clearable />
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

    <!-- 新用户账号展示弹窗 -->
    <el-dialog v-model="newAccountVisible" :title="$t('admin.userAdded')" width="380px" center :close-on-click-modal="false">
      <div class="new-account-info">
        <p class="account-tip">{{ $t('admin.accountGenerated') }}</p>
        <div class="account-box">
          <span class="account-label">{{ $t('login.account') }}</span>
          <span class="account-number">{{ newAccountNumber }}</span>
        </div>
        <p class="account-pwd-tip">{{ $t('admin.defaultPassword') }}: <b>123456</b></p>
      </div>
      <template #footer>
        <el-button type="primary" @click="newAccountVisible = false">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { Plus, InfoFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus/es/components/message/index'
import { getUserList, register, updateUser, resetUserPassword, deleteUser } from '@/api'
import { ElMessageBox } from 'element-plus/es/components/message-box/index'
import { getRoleText } from '@/utils/format'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const isSuperAdmin = computed(() => userStore.userInfo?.username === '100001')

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
  billing_day: null,
  allow_monthly_billing: false,
  is_active: true,
  telegram_bot_token: '',
  telegram_chat_id: '',
})

// 新用户账号展示
const newAccountVisible = ref(false)
const newAccountNumber = ref('')

const computedRules = computed(() => {
  const base = {
    full_name: [{ required: true, message: t('admin.fullNameRequired'), trigger: 'blur' }],
    role: [{ required: true, message: t('admin.roleRequired'), trigger: 'change' }],
  }
  return base
})

const getRoleType = (role) => {
  const map = { admin: 'danger', merchant: 'success' }
  return map[role] || ''
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
    billing_day: null,
    allow_monthly_billing: false,
    is_active: true,
    telegram_bot_token: '',
    telegram_chat_id: '',
  })
  formRef.value?.clearValidate()
}

const handleAdd = () => {
  resetForm()
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
    billing_day: row.billing_day || null,
    allow_monthly_billing: row.allow_monthly_billing || false,
    is_active: row.is_active,
    telegram_bot_token: row.telegram_bot_token || '',
    telegram_chat_id: row.telegram_chat_id || '',
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
        billing_day: form.billing_day,
        allow_monthly_billing: form.allow_monthly_billing,
        is_active: form.is_active,
        telegram_bot_token: form.telegram_bot_token || null,
        telegram_chat_id: form.telegram_chat_id || null,
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
        billing_day: form.billing_day,
        allow_monthly_billing: form.allow_monthly_billing,
      })
      // 展示自动生成的账号
      newAccountNumber.value = res.username
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

// 重置密码 - 二次确认
const handleResetPassword = async (row) => {
  try {
    // 第一次确认
    await ElMessageBox.confirm(
      t('admin.resetPasswordConfirm', { name: row.full_name || row.username }),
      t('admin.hint'),
      { confirmButtonText: t('common.confirm'), cancelButtonText: t('common.cancel'), type: 'warning' }
    )
    // 第二次确认 - 避免误触
    await ElMessageBox.confirm(
      `⚠️ ${t('admin.resetPasswordDoubleConfirm', { name: row.full_name || row.username })}`,
      t('admin.hint'),
      { confirmButtonText: t('admin.confirmReset'), cancelButtonText: t('common.cancel'), type: 'error', confirmButtonClass: 'el-button--danger' }
    )
    await resetUserPassword(row.id)
    ElMessage.success(t('admin.passwordResetSuccess'))
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
.merchants-page {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
}

/* ========== 移动端卡片列表 ========== */
.mobile-card-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.user-card {
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 10px;
  padding: 14px;
  cursor: pointer;
  transition: box-shadow 0.2s;
  overflow: hidden;
}

.user-card:active {
  background: #f5f7fa;
}

.card-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.card-user-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.card-username {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 60vw;
}

.card-body {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.card-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
  line-height: 1.5;
}

.card-label {
  color: #909399;
  flex-shrink: 0;
  margin-right: 12px;
}

.card-value {
  color: #303133;
  text-align: right;
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
  .merchants-page {
    padding: 12px;
  }

  .page-header h2 {
    font-size: 18px;
    margin: 0;
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
