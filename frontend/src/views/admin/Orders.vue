<template>
  <div class="orders-page">
    <div class="page-header">
      <h2>{{ $t('order.management') }}</h2>
    </div>
    
    <div class="filters">
      <el-row :gutter="10" align="middle">
        <el-col :xs="10" :sm="8" :md="6">
          <el-select v-model="filters.payment_status" :placeholder="$t('order.paymentStatus')" clearable style="width: 100%;">
            <el-option :label="$t('order.all')" value="" />
            <el-option :label="$t('order.unpaid')" value="unpaid" />
            <el-option :label="$t('order.cash')" value="cash" />
            <el-option :label="$t('order.monthly')" value="monthly" />
          </el-select>
        </el-col>
        <el-col :xs="10" :sm="8" :md="6">
          <el-select
            v-model="filters.delivery_status"
            :placeholder="$t('order.deliveryStatus')"
            clearable
            style="width: 100%;"
          >
            <el-option :label="$t('order.all')" value="" />
            <el-option :label="$t('order.deliveryPending')" value="pending" />
            <el-option :label="$t('order.delivering')" value="delivering" />
            <el-option :label="$t('order.delivered')" value="delivered" />
            <el-option :label="$t('order.cancelled')" value="cancelled" />
          </el-select>
        </el-col>
        <el-col :xs="4" :sm="4" :md="3">
          <el-button type="primary" @click="loadOrders" :size="mobile ? 'small' : 'default'" style="width:100%;padding-left:0;padding-right:0">{{ $t('common.query') }}</el-button>
        </el-col>
      </el-row>
    </div>

    <!-- 桌面端: 表格视图 -->
    <el-table
      v-if="!mobile"
      v-loading="loading"
      :data="paginatedOrders"
      border
      row-key="id"
      :default-sort="{ prop: 'created_at', order: 'descending' }"
    >
      <el-table-column :label="$t('order.orderNo')" prop="order_no" width="180" />
      <el-table-column :label="$t('order.merchant')" prop="merchant_name" width="120" />
      <el-table-column :label="$t('order.orderTime')" width="180" sortable prop="created_at">
        <template #default="{ row }">
          {{ formatDateTime(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column :label="$t('order.totalPriceUsd')" width="120" sortable prop="total_usd">
        <template #default="{ row }">
          {{ formatUSD(row.total_usd) }}
        </template>
      </el-table-column>
      <el-table-column :label="$t('order.paymentStatus')" width="130">
        <template #default="{ row }">
          <el-tag :type="getPaymentStatusType(row.payment_status)" size="small">
            {{ getPaymentStatusText(row.payment_status) }}
          </el-tag>
          <el-tag v-if="row.unpaid_days != null && row.unpaid_days > 0" type="danger" size="small" style="margin-left: 4px;">
            {{ $t('order.unpaidDaysShort', { days: row.unpaid_days }) }}
          </el-tag>
          <el-tag v-if="row.days_to_billing != null && row.payment_status === 'monthly'" type="warning" size="small" style="margin-left: 4px;">
            {{ $t('order.daysToBillingShort', { days: row.days_to_billing }) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="$t('order.deliveryStatus')" width="100">
        <template #default="{ row }">
          <el-tag :type="getDeliveryStatusType(row.delivery_status)" size="small">
            {{ getDeliveryStatusText(row.delivery_status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="$t('order.deliveryAddress')" prop="delivery_address" min-width="150" />
      <el-table-column :label="$t('common.operation')" width="180" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link @click="viewOrder(row)">
            {{ $t('common.detail') }}
          </el-button>
          <el-button type="success" link @click="handleUpdateStatus(row)">
            {{ $t('common.updateStatus') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 移动端: 卡片列表 -->
    <div v-else v-loading="loading" class="mobile-card-list">
      <div v-for="row in paginatedOrders" :key="row.id" class="order-card" @click="viewOrder(row)">
        <div class="ocard-header">
          <span class="ocard-no">{{ row.order_no }}</span>
          <span class="ocard-time">{{ formatDateTime(row.created_at) }}</span>
        </div>
        <div class="ocard-body">
          <div class="ocard-row">
            <span class="ocard-merchant">{{ row.merchant_name }}</span>
            <span class="ocard-amount">{{ formatUSD(row.total_usd) }}</span>
          </div>
          <div class="ocard-tags">
            <el-tag :type="getPaymentStatusType(row.payment_status)" size="small">
              {{ getPaymentStatusText(row.payment_status) }}
            </el-tag>
            <el-tag :type="getDeliveryStatusType(row.delivery_status)" size="small">
              {{ getDeliveryStatusText(row.delivery_status) }}
            </el-tag>
            <el-tag v-if="row.unpaid_days != null && row.unpaid_days > 0" type="danger" size="small">
              {{ $t('order.unpaidDaysShort', { days: row.unpaid_days }) }}
            </el-tag>
            <el-tag v-if="row.days_to_billing != null && row.payment_status === 'monthly'" type="warning" size="small">
              {{ $t('order.daysToBillingShort', { days: row.days_to_billing }) }}
            </el-tag>
          </div>
        </div>
        <div class="ocard-footer" @click.stop>
          <el-button type="primary" size="small" link @click="viewOrder(row)">{{ $t('common.detail') }}</el-button>
          <el-button type="success" size="small" link @click="handleUpdateStatus(row)">{{ $t('common.updateStatus') }}</el-button>
        </div>
      </div>
      <el-empty v-if="!loading && orders.length === 0" />
    </div>
    
    <div class="pagination-wrapper">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="orders.length"
        :layout="mobile ? 'prev, pager, next' : 'total, sizes, prev, pager, next, jumper'"
        :small="mobile"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
    
    <!-- 订单详情对话框 -->
    <el-dialog v-model="detailVisible" :title="$t('order.orderDetail')" :width="mobile ? '94vw' : '800px'" :fullscreen="mobile">
      <div v-if="currentOrder" class="order-detail">
        <el-descriptions :column="mobile ? 1 : 2" border size="small">
          <el-descriptions-item :label="$t('order.orderNo')">
            {{ currentOrder.order_no }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('order.merchant')">
            {{ currentOrder.merchant_name }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('order.orderTime')">
            {{ formatDateTime(currentOrder.created_at) }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('order.deliveryAddress')">
            {{ currentOrder.delivery_address || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('order.deliveryPhone')">
            {{ currentOrder.delivery_phone || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('order.paymentStatus')">
            <el-tag :type="getPaymentStatusType(currentOrder.payment_status)" size="small">
              {{ getPaymentStatusText(currentOrder.payment_status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item :label="$t('order.deliveryStatus')">
            <el-tag :type="getDeliveryStatusType(currentOrder.delivery_status)" size="small">
              {{ getDeliveryStatusText(currentOrder.delivery_status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item :label="$t('order.note')" :span="mobile ? 1 : 2">
            {{ currentOrder.note || '-' }}
          </el-descriptions-item>
        </el-descriptions>
        
        <h3 style="margin-top: 16px; font-size: 14px;">{{ $t('order.orderItems') }}</h3>

        <!-- 移动端: 商品明细卡片 -->
        <div v-if="mobile" class="order-items-mobile">
          <div v-for="item in currentOrder.items" :key="item.id" class="item-row">
            <span class="item-name">{{ item.product_name }}</span>
            <span class="item-detail">{{ formatUSD(item.unit_price_usd) }} × {{ item.quantity }}</span>
            <span class="item-subtotal">{{ formatUSD(item.subtotal_usd) }}</span>
          </div>
        </div>
        <!-- 桌面端: 商品明细表格 -->
        <el-table v-else :data="currentOrder.items" border size="small">
          <el-table-column :label="$t('product.name')" prop="product_name" />
          <el-table-column :label="$t('order.unitPrice')" width="120">
            <template #default="{ row }">
              {{ formatUSD(row.unit_price_usd) }}
            </template>
          </el-table-column>
          <el-table-column :label="$t('order.quantity')" prop="quantity" width="80" />
          <el-table-column :label="$t('order.itemSubtotal')" width="120">
            <template #default="{ row }">
              {{ formatUSD(row.subtotal_usd) }}
            </template>
          </el-table-column>
        </el-table>
        
        <div class="order-total">
          <span>{{ $t('order.orderTotal') }}:</span>
          <strong>{{ formatUSD(currentOrder.total_usd) }}</strong>
        </div>

        <!-- 移动端在详情内直接提供状态更新入口 -->
        <div v-if="mobile" class="mobile-detail-actions">
          <el-button type="success" style="width: 100%;" @click="detailVisible = false; handleUpdateStatus(currentOrder)">
            {{ $t('common.updateStatus') }}
          </el-button>
        </div>
      </div>
    </el-dialog>
    
    <!-- 更新状态对话框 -->
    <el-dialog v-model="statusVisible" :title="$t('order.updateOrderStatus')" :width="mobile ? '94vw' : '500px'" :fullscreen="mobile">
      <el-form v-if="currentOrder" :label-width="mobile ? '80px' : '100px'" :label-position="mobile ? 'top' : 'right'">
        <el-form-item :label="$t('order.paymentStatus')">
          <el-radio-group v-model="statusForm.payment_status">
            <el-radio label="unpaid">{{ $t('order.unpaid') }}</el-radio>
            <el-radio label="cash">{{ $t('order.cash') }}</el-radio>
            <el-radio label="monthly">{{ $t('order.monthly') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('order.deliveryStatus')">
          <el-radio-group v-model="statusForm.delivery_status" class="radio-group-wrap">
            <el-radio label="pending">{{ $t('order.deliveryPending') }}</el-radio>
            <el-radio label="delivering">{{ $t('order.delivering') }}</el-radio>
            <el-radio label="delivered">{{ $t('order.delivered') }}</el-radio>
            <el-radio label="cancelled">{{ $t('order.cancelled') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('cart.scheduledAt')">
          <el-date-picker
            v-model="statusForm.scheduled_at"
            type="datetime"
            :placeholder="$t('cart.scheduledAt')"
            style="width: 100%"
            value-format="YYYY-MM-DDTHH:mm:ss"
          />
        </el-form-item>
        <el-form-item :label="$t('order.note')">
          <el-input
            v-model="statusForm.note"
            type="textarea"
            :rows="3"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="statusVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button
          type="primary"
          :loading="submitting"
          @click="submitStatusUpdate"
        >
          {{ $t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus/es/components/message/index'
import { getOrders, updateOrder } from '@/api'
import {
  formatUSD,
  formatDateTime,
  getPaymentStatusText,
  getDeliveryStatusText,
} from '@/utils/format'

const { t } = useI18n()

// 移动端检测
const mobile = ref(window.innerWidth < 768)
const onResize = () => { mobile.value = window.innerWidth < 768 }
onMounted(() => window.addEventListener('resize', onResize))
onBeforeUnmount(() => window.removeEventListener('resize', onResize))

const loading = ref(false)
const orders = ref([])
const detailVisible = ref(false)
const statusVisible = ref(false)
const submitting = ref(false)
const currentOrder = ref(null)

const filters = reactive({
  payment_status: '',
  delivery_status: '',
})

// 分页
const currentPage = ref(1)
const pageSize = ref(20)

const paginatedOrders = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return orders.value.slice(start, end)
})

const handleSizeChange = () => {
  currentPage.value = 1
}

const handleCurrentChange = () => {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

const statusForm = reactive({
  payment_status: 'unpaid',
  delivery_status: 'pending',
  scheduled_at: '',
  note: '',
})

const getPaymentStatusType = (status) => {
  const map = { unpaid: 'warning', cash: 'success', monthly: 'primary' }
  return map[(status || '').toLowerCase()] || ''
}

const getDeliveryStatusType = (status) => {
  const map = {
    pending: 'info',
    delivering: 'warning',
    delivered: 'success',
    cancelled: 'danger',
  }
  return map[(status || '').toLowerCase()] || ''
}

const loadOrders = async () => {
  loading.value = true
  try {
    const data = await getOrders(filters)
    orders.value = data
  } catch (error) {
    console.error('加载订单失败:', error)
  } finally {
    loading.value = false
  }
}

const viewOrder = (order) => {
  currentOrder.value = order
  detailVisible.value = true
}

const handleUpdateStatus = (order) => {
  currentOrder.value = order
  statusForm.payment_status = order.payment_status
  statusForm.delivery_status = order.delivery_status
  statusForm.scheduled_at = order.scheduled_at ? order.scheduled_at.slice(0, 19) : ''
  statusForm.note = order.note || ''
  statusVisible.value = true
}

const submitStatusUpdate = async () => {
  submitting.value = true
  try {
    const payload = { ...statusForm }
    // 空字符串的时间字段会让 Go 后端解析失败，转成 null
    if (!payload.scheduled_at) {
      payload.scheduled_at = null
    } else {
      // 补全为带时区的 ISO 字符串以匹配 Go RFC3339
      payload.scheduled_at = new Date(payload.scheduled_at).toISOString()
    }
    await updateOrder(currentOrder.value.id, payload)
    ElMessage.success(t('order.statusUpdateSuccess'))
    statusVisible.value = false
    loadOrders()
  } catch (error) {
    console.error('更新失败:', error)
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadOrders()
})
</script>

<style scoped>
.orders-page {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.page-header h2 {
  margin: 0;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.filters {
  margin-bottom: 16px;
}

.order-detail {
  padding: 10px 0;
}

.order-total {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 12px;
  margin-top: 16px;
  padding-top: 12px;
  border-top: 2px solid #e4e7ed;
  font-size: 16px;
}

/* ========== 移动端卡片列表 ========== */
.mobile-card-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.order-card {
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 10px;
  padding: 14px;
  cursor: pointer;
  transition: box-shadow 0.2s;
  overflow: hidden;
}

.order-card:active {
  background: #f5f7fa;
}

.ocard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
  gap: 8px;
  overflow: hidden;
}

.ocard-no {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  font-family: monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  min-width: 0;
}

.ocard-time {
  font-size: 11px;
  color: #909399;
}

.ocard-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.ocard-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.ocard-merchant {
  font-size: 14px;
  color: #303133;
}

.ocard-amount {
  font-size: 16px;
  font-weight: 700;
  color: #1D4ED8;
}

.ocard-tags {
  display: flex;
  gap: 6px;
}

.ocard-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
}

/* 订单详情内的商品明细（移动端） */
.order-items-mobile {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  overflow: hidden;
}

.item-row {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  font-size: 13px;
  border-bottom: 1px solid #f0f0f0;
}

.item-row:last-child {
  border-bottom: none;
}

.item-name {
  flex: 1;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-right: 8px;
}

.item-detail {
  color: #909399;
  font-size: 12px;
  white-space: nowrap;
  margin-right: 12px;
}

.item-subtotal {
  font-weight: 600;
  color: #303133;
  white-space: nowrap;
}

.mobile-detail-actions {
  margin-top: 16px;
}

/* ========== 移动端适配 ========== */
@media (max-width: 767px) {
  .orders-page {
    padding: 12px;
  }

  .page-header h2 {
    font-size: 18px;
  }

  .order-total {
    font-size: 15px;
  }

  :deep(.el-dialog) {
    margin-top: 0 !important;
    border-radius: 12px 12px 0 0;
  }

  :deep(.el-dialog__body) {
    max-height: 75vh;
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
  }

  :deep(.el-descriptions) {
    font-size: 13px;
  }

  .radio-group-wrap {
    display: flex;
    flex-wrap: wrap;
    gap: 4px 0;
  }

  :deep(.el-form-item__label) {
    font-size: 13px;
    padding-bottom: 4px;
  }

  :deep(.el-select) {
    width: 100% !important;
  }

  .pagination-wrapper :deep(.el-pagination) {
    flex-wrap: wrap;
    justify-content: center;
  }
}
</style>
