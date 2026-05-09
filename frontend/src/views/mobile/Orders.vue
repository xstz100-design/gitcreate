<template>
  <div class="mobile-orders">
    <van-nav-bar :title="$t('order.title')" fixed placeholder />
    
    <!-- 订单状态筛选 -->
    <van-tabs v-model:active="activeTab" sticky offset-top="46px">
      <van-tab :title="$t('order.all')" />
      <van-tab :title="$t('order.deliveryPending')" />
      <van-tab :title="$t('order.delivering')" />
      <van-tab :title="$t('order.delivered')" />
    </van-tabs>
    
    <!-- 订单列表 -->
    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <van-list
        v-model:loading="loading"
        :finished="finished"
        :finished-text="$t('common.noMore')"
        @load="onLoad"
      >
        <div
          v-for="order in filteredOrders"
          :key="order.id"
          class="order-card"
          @click="showOrderDetail(order)"
        >
          <!-- 订单头部 -->
          <div class="order-header">
            <span class="order-no">{{ order.order_no }}</span>
            <div class="order-tags">
              <van-tag :type="getPaymentStatusType(order.payment_status)" plain size="medium">
                {{ getPaymentStatusText(order.payment_status) }}
              </van-tag>
              <van-tag :type="getDeliveryStatusType(order.delivery_status)" size="medium">
                {{ getDeliveryStatusText(order.delivery_status) }}
              </van-tag>
              <van-tag v-if="order.unpaid_days != null && order.unpaid_days > 0" type="danger" size="medium">
                {{ $t('order.unpaidDaysShort', { days: order.unpaid_days }) }}
              </van-tag>
              <van-tag v-if="order.days_to_billing != null && order.payment_status === 'monthly'" type="warning" size="medium">
                {{ $t('order.daysToBillingShort', { days: order.days_to_billing }) }}
              </van-tag>
            </div>
          </div>
          
          <!-- 订单商品列表 (最多显示3个) -->
          <div class="order-items">
            <div
              v-for="(item, index) in (order.items || []).slice(0, 3)"
              :key="index"
              class="order-item"
            >
              <span class="item-name">{{ item.product_name }}</span>
              <span class="item-price-qty">
                <span class="item-unit-price">${{ item.unit_price_usd }}</span>
                <span class="item-times">×</span>
                <span class="item-qty">{{ item.quantity }}</span>
              </span>
              <span class="item-subtotal">${{ item.subtotal_usd }}</span>
            </div>
            <div v-if="(order.items || []).length > 3" class="more-items">
              {{ $t('order.moreItems', { count: order.items.length - 3 }) }}
            </div>
          </div>
          
          <!-- 订单底部 -->
          <div class="order-footer">
            <div class="order-time">{{ formatDateTime(order.created_at) }}</div>
            <div class="order-price">
              <span class="price-label">{{ $t('order.totalPrice') }}:</span>
              <span class="price-usd">${{ order.total_usd }}</span>
            </div>
          </div>
        </div>
        
        <van-empty
          v-if="!loading && filteredOrders.length === 0"
          :description="$t('order.noOrders')"
        />
      </van-list>
    </van-pull-refresh>
    
    <!-- 订单详情弹窗 -->
    <van-popup
      v-model:show="showDetail"
      position="bottom"
      :style="{ height: '80%' }"
      round
      closeable
    >
      <div v-if="currentOrder" class="detail-popup">
        <div class="popup-header">
          <h3>{{ $t('order.orderDetail') }}</h3>
          <van-tag :type="getDeliveryStatusType(currentOrder.delivery_status)">
            {{ getDeliveryStatusText(currentOrder.delivery_status) }}
          </van-tag>
        </div>
        
        <div class="popup-content">
          <!-- 订单信息 -->
          <van-cell-group inset>
            <van-cell :title="$t('order.orderNo')" :value="currentOrder.order_no" />
            <van-cell :title="$t('order.orderTime')" :value="formatDateTime(currentOrder.created_at)" />
            <van-cell
              :title="$t('order.paymentStatus')"
              :value="getPaymentStatusText(currentOrder.payment_status)"
            />
            <van-cell
              :title="$t('order.deliveryStatus')"
              :value="getDeliveryStatusText(currentOrder.delivery_status)"
            />
            <van-cell
              :title="$t('order.deliveryAddress')"
              :value="currentOrder.delivery_address || '-'"
            />
            <van-cell
              :title="$t('order.deliveryPhone')"
              :value="currentOrder.delivery_phone || '-'"
            />
          </van-cell-group>
          
          <!-- 商品列表 -->
          <div class="detail-items">
            <div class="section-title">{{ $t('order.itemList') }}</div>
            <div
              v-for="item in (currentOrder.items || [])"
              :key="item.product_id"
              class="detail-item"
            >
              <div class="item-info">
                <div class="item-name">{{ item.product_name }}</div>
                <div class="item-spec">{{ formatUSD(item.unit_price_usd) }} × {{ item.quantity }}</div>
              </div>
              <div class="item-subtotal">{{ formatUSD(item.subtotal_usd) }}</div>
            </div>
          </div>
          
          <!-- 价格汇总 -->
          <div class="price-summary">
            <div class="summary-row">
              <span>{{ $t('order.subtotal') }}:</span>
              <span class="price-usd">${{ currentOrder.total_usd }}</span>
            </div>
            <div class="summary-row total">
              <span>{{ $t('order.payableAmount') }}:</span>
              <span class="price-usd">${{ currentOrder.total_usd }}</span>
            </div>
            <div class="summary-row">
              <span></span>
              <span class="price-khr">{{ formatKHR(currentOrder.total_khr || usdToKhr(currentOrder.total_usd)) }}</span>
            </div>
          </div>
          
          <!-- 备注 -->
          <van-cell-group v-if="currentOrder.note" inset>
            <van-cell :title="$t('order.note')" :value="currentOrder.note" />
          </van-cell-group>

          <!-- 取消订单 -->
          <div v-if="canCancel(currentOrder)" class="cancel-section">
            <van-button type="danger" plain round block :loading="cancelling" @click="handleCancel(currentOrder)">
              {{ $t('order.cancelOrder') }}
            </van-button>
          </div>

          <!-- 联系客服（送货中/已送达时） -->
          <div v-if="showContactBtn(currentOrder) && contactPhone" class="contact-section">
            <a :href="'tel:' + contactPhone" class="contact-btn-link">
              <van-button type="primary" plain round block icon="phone-o">
                联系客服 / ទំនាក់ទំនង {{ contactPhone }}
              </van-button>
            </a>
          </div>
        </div>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { showToast, showDialog } from 'vant'
import { getOrders, cancelOrder, getContactInfo } from '@/api'
import {
  formatUSD,
  formatKHR,
  usdToKhr,
  formatDateTime,
  getPaymentStatusText,
  getDeliveryStatusText,
} from '@/utils/format'
import { hapticFeedback } from '@/utils/device'

const { t } = useI18n()

const activeTab = ref(0)
const orders = ref([])
const loading = ref(false)
const finished = ref(false)
const refreshing = ref(false)
const showDetail = ref(false)
const currentOrder = ref(null)
const cancelling = ref(false)
const contactPhone = ref('')

// 加载客服电话
const loadContactPhone = async () => {
  try {
    const info = await getContactInfo()
    contactPhone.value = info.phone || ''
  } catch {
    // 静默处理
  }
}

onMounted(() => {
  loadContactPhone()
})

// 根据Tab筛选订单
const filteredOrders = computed(() => {
  const statusMap = ['', 'pending', 'delivering', 'delivered']
  const status = statusMap[activeTab.value]
  if (!status) return orders.value
  return orders.value.filter(order => order.delivery_status === status)
})

// 配送状态类型
const getDeliveryStatusType = (status) => {
  const map = {
    pending: 'default',
    delivering: 'warning',
    delivered: 'success',
    cancelled: 'danger',
  }
  return map[(status || '').toLowerCase()] || 'default'
}

// 支付状态类型
const getPaymentStatusType = (status) => {
  const map = {
    unpaid: 'warning',
    cash: 'success',
    monthly: 'primary',
  }
  return map[(status || '').toLowerCase()] || 'default'
}

// 加载订单
const loadOrders = async () => {
  try {
    const data = await getOrders()
    orders.value = data
    finished.value = true
  } catch (error) {
    showToast(t('order.loadFailed'))
  }
}

// 下拉刷新
const onRefresh = async () => {
  refreshing.value = true
  await loadOrders()
  refreshing.value = false
  showToast(t('common.refreshSuccess'))
}

// 加载更多
const onLoad = async () => {
  if (orders.value.length === 0) {
    await loadOrders()
  } else {
    finished.value = true
  }
  loading.value = false
}

// 显示订单详情
const showOrderDetail = (order) => {
  hapticFeedback('medium')
  currentOrder.value = order
  showDetail.value = true
}

// 判断订单是否可取消（仅待配送状态可取消）
const canCancel = (order) => {
  return order.delivery_status === 'pending'
}

// 判断是否显示联系客服按钮（送货中或已送达不可取消时）
const showContactBtn = (order) => {
  return order.delivery_status === 'delivering' || order.delivery_status === 'delivered'
}

// 取消订单
const handleCancel = async (order) => {
  try {
    await showDialog({
      title: t('order.cancelOrder'),
      message: t('order.cancelConfirm'),
      showCancelButton: true,
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
    })
  } catch { return }

  cancelling.value = true
  try {
    await cancelOrder(order.id)
    hapticFeedback('success')
    showToast(t('order.cancelSuccess'))
    showDetail.value = false
    await loadOrders()
  } catch (error) {
    hapticFeedback('error')
  } finally {
    cancelling.value = false
  }
}
</script>

<style scoped>
.mobile-orders {
  min-height: var(--tg-viewport-height, 100vh);
  background: #f5f5f5;
}

.order-card {
  margin: 10px;
  padding: 12px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.08);
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 8px;
  border-bottom: 1px solid #f0f0f0;
  margin-bottom: 12px;
}

.order-no {
  font-size: 13px;
  color: #8c8c8c;
  flex-shrink: 1;
  overflow: hidden;
  text-overflow: ellipsis;
}

.order-tags {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.order-items {
  margin-bottom: 12px;
}

.order-item {
  display: flex;
  align-items: center;
  padding: 4px 0;
  font-size: 14px;
  gap: 8px;
}

.order-item .item-name {
  flex: 1;
  color: #262626;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-price-qty {
  display: flex;
  align-items: center;
  gap: 2px;
  font-size: 12px;
  color: #8c8c8c;
  flex-shrink: 0;
}

.item-times {
  margin: 0 1px;
}

.item-subtotal {
  font-size: 14px;
  font-weight: 600;
  color: #262626;
  flex-shrink: 0;
  min-width: 55px;
  text-align: right;
}

.more-items {
  padding: 4px 0;
  font-size: 13px;
  color: #8c8c8c;
}

.order-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
}

.order-time {
  font-size: 12px;
  color: #8c8c8c;
}

.order-price {
  display: flex;
  align-items: center;
  gap: 6px;
}

.price-label {
  font-size: 13px;
  color: #8c8c8c;
}

.price-usd {
  font-size: 18px;
  font-weight: 700;
  color: #f5222d;
}

/* 详情弹窗 */
.detail-popup {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.popup-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #f0f0f0;
}

.popup-header h3 {
  margin: 0;
  font-size: 18px;
}

.popup-content {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
}

.section-title {
  padding: 12px 16px;
  font-size: 14px;
  font-weight: 600;
  color: #262626;
}

.detail-items {
  background: #fff;
  border-radius: 8px;
  margin: 10px 0;
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

.item-info {
  flex: 1;
}

.detail-item .item-name {
  font-size: 15px;
  color: #262626;
  margin-bottom: 4px;
}

.item-spec {
  font-size: 12px;
  color: #8c8c8c;
}

.detail-item .item-subtotal {
  font-size: 15px;
  font-weight: 600;
  color: #262626;
}

.price-summary {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  margin: 10px 0;
}

.summary-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 0;
  font-size: 14px;
}

.summary-row.total {
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
  margin-top: 6px;
  font-size: 16px;
  font-weight: 600;
}

.summary-row .price-usd {
  font-size: 20px;
}

.price-khr {
  font-size: 12px;
  color: #8c8c8c;
}

.cancel-section {
  padding: 16px;
}

.contact-section {
  padding: 8px 16px 16px;
}

.contact-btn-link {
  text-decoration: none;
  display: block;
}
</style>
