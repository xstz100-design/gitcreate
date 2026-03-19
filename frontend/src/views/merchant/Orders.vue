<template>
  <div class="orders-page">
    <!-- 状态筛选 Tabs -->
    <div class="status-tabs">
      <div
        v-for="(tab, index) in tabs"
        :key="index"
        class="tab-item"
        :class="{ active: activeTab === index }"
        @click="activeTab = index"
      >
        {{ tab.label }}
      </div>
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="loading-state">
      <el-icon class="is-loading" :size="24"><Loading /></el-icon>
      <span>{{ $t('common.loading') }}</span>
    </div>

    <!-- 空状态 -->
    <el-empty v-else-if="filteredOrders.length === 0" :description="$t('order.noOrders')" />

    <!-- 订单卡片列表 -->
    <div v-else class="order-list">
      <div
        v-for="order in filteredOrders"
        :key="order.id"
        class="order-card"
        @click="viewOrder(order)"
      >
        <!-- 卡片头部 -->
        <div class="card-header">
          <span class="order-no">{{ order.order_no }}</span>
          <el-tag :type="getDeliveryStatusType(order.delivery_status)" size="small" round>
            {{ getDeliveryStatusText(order.delivery_status) }}
          </el-tag>
        </div>

        <!-- 商品摘要 -->
        <div class="card-items">
          <div
            v-for="(item, idx) in order.items.slice(0, 3)"
            :key="idx"
            class="item-row"
          >
            <span class="item-name">{{ item.product_name }}</span>
            <span class="item-qty">x{{ item.quantity }}</span>
            <span class="item-price">${{ item.subtotal_usd }}</span>
          </div>
          <div v-if="order.items.length > 3" class="more-items">
            {{ $t('order.moreItems', { count: order.items.length - 3 }) }}
          </div>
        </div>

        <!-- 卡片底部 -->
        <div class="card-footer">
          <span class="order-time">{{ formatDateTime(order.created_at) }}</span>
          <div class="order-total">
            <span v-if="order.days_to_billing != null && order.payment_status === 'monthly'" class="billing-badge">
              {{ $t('order.daysToBillingShort', { days: order.days_to_billing }) }}
            </span>
            <span class="total-label">{{ $t('common.total') }}</span>
            <span class="total-price">${{ order.total_usd }}</span>
          </div>
        </div>
      </div>
    </div>
    
    <!-- 订单详情弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="$t('order.orderDetail')"
      width="600px"
      top="6vh"
      destroy-on-close
    >
      <div v-if="currentOrder" class="order-detail">
        <!-- 订单状态头 -->
        <div class="detail-status-bar">
          <div class="status-left">
            <el-tag :type="getDeliveryStatusType(currentOrder.delivery_status)" size="large">
              {{ getDeliveryStatusText(currentOrder.delivery_status) }}
            </el-tag>
            <el-tag :type="getPaymentStatusType(currentOrder.payment_status)" size="small">
              {{ getPaymentStatusText(currentOrder.payment_status) }}
            </el-tag>
          </div>
          <span class="order-no-detail">{{ currentOrder.order_no }}</span>
        </div>

        <!-- 订单信息 -->
        <div class="detail-section">
          <div class="section-title">{{ $t('order.orderInfo') }}</div>
          <div class="info-grid">
            <div class="info-item">
              <span class="info-label">{{ $t('order.orderTime') }}</span>
              <span class="info-value">{{ formatDateTime(currentOrder.created_at) }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">{{ $t('order.deliveryAddress') }}</span>
              <span class="info-value">{{ currentOrder.delivery_address || '-' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">{{ $t('order.deliveryPhone') }}</span>
              <span class="info-value">{{ currentOrder.delivery_phone || '-' }}</span>
            </div>
            <div v-if="currentOrder.note" class="info-item full">
              <span class="info-label">{{ $t('order.note') }}</span>
              <span class="info-value">{{ currentOrder.note }}</span>
            </div>
          </div>
        </div>

        <!-- 商品清单 -->
        <div class="detail-section">
          <div class="section-title">{{ $t('order.itemList') }}</div>
          <div class="items-list">
            <div
              v-for="item in currentOrder.items"
              :key="item.product_id"
              class="detail-item"
            >
              <div class="detail-item-info">
                <span class="detail-item-name">{{ item.product_name }}</span>
                <span class="detail-item-spec">{{ formatUSD(item.unit_price_usd) }} × {{ item.quantity }}</span>
              </div>
              <span class="detail-item-subtotal">{{ formatUSD(item.subtotal_usd) }}</span>
            </div>
          </div>
        </div>

        <!-- 价格汇总 -->
        <div class="price-summary">
          <div class="summary-row total-row">
            <span>{{ $t('order.payableAmount') }}</span>
            <div class="summary-price">
              <span class="total-usd">${{ currentOrder.total_usd }}</span>
              <span class="total-khr">{{ formatKHR(currentOrder.total_khr) }}</span>
            </div>
          </div>
        </div>

        <!-- 取消订单按钮 -->
        <div v-if="canCancel(currentOrder)" class="cancel-section">
          <el-button type="danger" plain round :loading="cancelling" @click="handleCancel(currentOrder)">
            {{ $t('order.cancelOrder') }}
          </el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { Loading } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus/es/components/message/index'
import { ElMessageBox } from 'element-plus/es/components/message-box/index'
import { getOrders, cancelOrder } from '@/api'
import {
  formatUSD,
  formatKHR,
  formatDateTime,
  getPaymentStatusText,
  getDeliveryStatusText,
} from '@/utils/format'

const { t } = useI18n()

const loading = ref(false)
const orders = ref([])
const dialogVisible = ref(false)
const currentOrder = ref(null)
const activeTab = ref(0)
const cancelling = ref(false)

const tabs = computed(() => [
  { label: t('order.all'), status: '' },
  { label: t('order.deliveryPending'), status: 'pending' },
  { label: t('order.delivering'), status: 'delivering' },
  { label: t('order.completed'), status: 'delivered' },
  { label: t('order.cancelled'), status: 'cancelled' },
])

// 根据Tab筛选订单
const filteredOrders = computed(() => {
  const status = tabs.value[activeTab.value].status
  if (!status) return orders.value
  return orders.value.filter(order => order.delivery_status === status)
})

// 支付状态标签类型
const getPaymentStatusType = (status) => {
  const map = { unpaid: 'warning', cash: 'success', monthly: 'primary' }
  return map[status] || ''
}

// 配送状态标签类型
const getDeliveryStatusType = (status) => {
  const map = { pending: 'info', delivering: 'warning', delivered: 'success', cancelled: 'danger' }
  return map[status] || ''
}

// 加载订单列表
const loadOrders = async () => {
  loading.value = true
  try {
    const data = await getOrders()
    orders.value = data
  } catch (error) {
    console.error('加载订单失败:', error)
  } finally {
    loading.value = false
  }
}

// 查看订单详情
const viewOrder = (order) => {
  currentOrder.value = order
  dialogVisible.value = true
}

// 判断订单是否可以取消
const canCancel = (order) => {
  return order.delivery_status !== 'delivered' &&
    order.delivery_status !== 'cancelled'
}

// 取消订单
const handleCancel = async (order) => {
  try {
    await ElMessageBox.confirm(t('order.cancelConfirm'), t('common.confirm'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    })
  } catch { return }

  cancelling.value = true
  try {
    await cancelOrder(order.id)
    ElMessage.success(t('order.cancelSuccess'))
    dialogVisible.value = false
    loadOrders()
  } catch (error) {
    console.error('取消订单失败:', error)
  } finally {
    cancelling.value = false
  }
}

onMounted(() => {
  loadOrders()
})
</script>

<style scoped>
.orders-page {
  min-height: 100%;
  background: #f7f7f7;
}

/* 状态Tab */
.status-tabs {
  display: flex;
  background: #fff;
  border-bottom: 1px solid #f0f0f0;
  position: sticky;
  top: 0;
  z-index: 10;
}

.tab-item {
  flex: 1;
  text-align: center;
  padding: 14px 0;
  font-size: 14px;
  color: #666;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition: all 0.25s;
}

.tab-item.active {
  color: #409eff;
  font-weight: 600;
  border-bottom-color: #409eff;
}

.tab-item:hover {
  color: #409eff;
}

/* 加载状态 */
.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 60px 0;
  color: #999;
  font-size: 14px;
}

/* 订单列表 */
.order-list {
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.order-card {
  background: #fff;
  border-radius: 12px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
}

.order-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
  transform: translateY(-1px);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 12px;
  border-bottom: 1px solid #f5f5f5;
}

.order-no {
  font-size: 13px;
  color: #999;
  font-family: monospace;
}

.card-items {
  padding: 12px 0;
}

.item-row {
  display: flex;
  align-items: center;
  padding: 4px 0;
  font-size: 14px;
}

.item-name {
  flex: 1;
  color: #1a1a1a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-qty {
  width: 48px;
  text-align: center;
  color: #999;
  font-size: 13px;
}

.item-price {
  width: 72px;
  text-align: right;
  color: #1a1a1a;
  font-weight: 500;
}

.more-items {
  font-size: 13px;
  color: #999;
  padding: 4px 0;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 12px;
  border-top: 1px solid #f5f5f5;
}

.order-time {
  font-size: 12px;
  color: #999;
}

.order-total {
  display: flex;
  align-items: center;
  gap: 6px;
}

.total-label {
  font-size: 13px;
  color: #999;
}

.total-price {
  font-size: 20px;
  font-weight: 700;
  color: #f5222d;
}

/* 订单详情弹窗 */
.order-detail {
  padding: 0;
}

.detail-status-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 0 16px;
  border-bottom: 1px solid #f0f0f0;
}

.status-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.order-no-detail {
  font-size: 13px;
  color: #999;
  font-family: monospace;
}

.detail-section {
  padding: 16px 0;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 12px;
}

.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-item.full {
  grid-column: 1 / -1;
}

.info-label {
  font-size: 12px;
  color: #999;
}

.info-value {
  font-size: 14px;
  color: #1a1a1a;
}

.items-list {
  background: #fafafa;
  border-radius: 8px;
  overflow: hidden;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
}

.detail-item:last-child {
  border-bottom: none;
}

.detail-item-info {
  flex: 1;
}

.detail-item-name {
  font-size: 14px;
  color: #1a1a1a;
  display: block;
  margin-bottom: 2px;
}

.detail-item-spec {
  font-size: 12px;
  color: #999;
}

.detail-item-subtotal {
  font-size: 15px;
  font-weight: 600;
  color: #1a1a1a;
}

.price-summary {
  background: #fafafa;
  border-radius: 8px;
  padding: 16px;
  margin-top: 4px;
}

.summary-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.total-row {
  font-size: 15px;
  font-weight: 600;
}

.summary-price {
  text-align: right;
}

.total-usd {
  font-size: 22px;
  font-weight: 700;
  color: #f5222d;
  display: block;
}

.total-khr {
  font-size: 12px;
  color: #999;
}

.cancel-section {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
  text-align: center;
}

.unpaid-badge {
  display: inline-block;
  background: #f56c6c;
  color: #fff;
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 8px;
  margin-right: 6px;
  font-weight: 600;
}

.billing-badge {
  display: inline-block;
  background: #e6a23c;
  color: #fff;
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 8px;
  margin-right: 6px;
  font-weight: 600;
}

/* === 移动端适配 === */
@media (max-width: 767px) {
  .orders-page {
    padding: 12px;
    padding-bottom: 70px;
  }

  .order-card {
    padding: 12px;
  }

  .el-dialog {
    width: 92vw !important;
  }
}
</style>
