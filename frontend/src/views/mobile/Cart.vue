<template>
  <div class="mobile-cart">
    <van-nav-bar :title="$t('cart.title')" :left-text="$t('common.back')" left-arrow fixed placeholder @click-left="handleBack" />
    
    <van-empty v-if="cartStore.items.length === 0" :description="$t('cart.empty')">
      <van-button type="primary" round @click="$router.push('/m/shop')">
        {{ $t('cart.goShopping') }}
      </van-button>
    </van-empty>
    
    <template v-else>
      <div v-if="!userStore.canOrder" class="restriction-panel">
        <div class="restriction-title">{{ restrictionTitle }}</div>
        <div class="restriction-desc">{{ restrictionDescription }}</div>
        <div class="restriction-note">{{ cartSavedTip }}</div>
        <van-button round block type="warning" class="restriction-btn" @click="goToRestrictionAction">
          {{ restrictionActionLabel }}
        </van-button>
      </div>

      <!-- 商铺信息 -->
      <van-cell-group inset style="margin-top: 10px">
        <van-cell :title="$t('cart.wholesaleShop')" is-link />
      </van-cell-group>
      
      <!-- 商品列表 -->
      <van-checkbox-group v-model="checkedItems">
        <van-swipe-cell
          v-for="item in cartStore.items"
          :key="item.id"
        >
          <div class="cart-item">
            <van-checkbox :name="item.id" />
            
            <van-image
              :src="item.image_url || '/placeholder.png'"
              width="80"
              height="80"
              fit="cover"
            />
            
            <div class="item-info">
              <div class="item-name">{{ item.name }}</div>
              <div v-if="item.name_kh" class="item-name-kh">{{ item.name_kh }}</div>
              <div v-if="item.purchase_mode && item.purchase_mode !== 'default'" class="item-mode-tag">
                {{ item.purchase_mode === 'piece' ? (item.display_unit || '件购') : (item.display_unit || '箱购') }}
              </div>
              
              <div class="item-price">
                <span class="price-usd">${{ item.price_usd }}</span>
                <span class="price-khr">{{ formatKHR(usdToKhr(item.price_usd)) }}</span>
              </div>
              
              <van-stepper
                :model-value="item.quantity"
                :min="1"
                :max="item.stock"
                @change="val => updateQuantity(item.id, val)"
                button-size="22"
                input-width="36"
              />
            </div>
          </div>
          
          <template #right>
            <van-button
              square
              type="danger"
              :text="$t('common.delete')"
              style="height: 100%"
              @click="removeItem(item.id)"
            />
          </template>
        </van-swipe-cell>
      </van-checkbox-group>
      
      <!-- 配送信息 -->
      <van-cell-group inset>
        <van-cell
          v-if="!userStore.canOrder"
          :title="cartFormTip"
          icon="warning-o"
          class="checkout-disabled-tip"
        />
        <van-field
          v-model="orderForm.delivery_address"
          :label="$t('cart.address')"
          :placeholder="$t('cart.addressPlaceholder')"
          :readonly="!formEditable"
          clearable
          @focus="handleInputFocus"
        />
        <van-field
          v-model="orderForm.delivery_phone"
          :label="$t('cart.phone')"
          type="tel"
          :placeholder="$t('cart.phonePlaceholder')"
          :readonly="!formEditable"
          clearable
          @focus="handleInputFocus"
        />
        <van-field
          v-model="orderForm.note"
          :label="$t('cart.note')"
          type="textarea"
          :placeholder="$t('cart.notePlaceholder')"
          :readonly="!formEditable"
          rows="2"
          autosize
          maxlength="200"
          show-word-limit
          @focus="handleInputFocus"
        />
        <!-- 配送费（根据地址自动估算，不需要手动输入距离） -->
        <van-cell :title="$t('cart.deliveryFee')">
          <template #value>
            <span v-if="estimatingFee" style="color:#999">计算中...</span>
            <span v-else-if="deliveryFee !== null" style="color:#ee0a24;font-weight:600">${{ deliveryFee.toFixed(2) }}</span>
            <span v-else style="color:#999">--</span>
          </template>
          <template v-if="!estimatingFee" #label>
            <span v-if="autoEstimated" style="font-size:11px;color:#52c41a">📍 已根据您的收货地址自动计算</span>
            <span v-else-if="deliveryFee === null" style="font-size:11px;color:#999">请先在个人资料中设置收货地址</span>
          </template>
        </van-cell>
      </van-cell-group>
      
      <!-- 支付方式 -->
      <van-cell-group inset>
        <van-cell :title="$t('cart.paymentMethod')" :is-link="formEditable" @click="openPaymentPicker">
          <template #value>
            <span :style="{ color: orderForm.payment_status === 'monthly' ? '#1989fa' : '#333' }">
              {{ orderForm.payment_status === 'monthly' ? $t('cart.monthlyPayment') : $t('cart.cashPayment') }}
            </span>
          </template>
        </van-cell>
      </van-cell-group>

      <!-- 支付方式选择弹窗 -->
      <van-action-sheet
        v-model:show="showPaymentPicker"
        teleport="body"
        :actions="paymentActions"
        :cancel-text="$t('common.cancel')"
        @select="onPaymentSelect"
      />
      
      <!-- 底部结算栏 -->
      <van-submit-bar
        :price="totalWithFee * 100"
        :button-text="primaryButtonText"
        @submit="handlePrimaryAction"
        :loading="submitting"
        :disabled="submitting"
      >
        <van-checkbox v-model="checkAll">{{ $t('cart.selectAll') }}</van-checkbox>
        <template #tip>
          <span v-if="deliveryFee !== null">{{ $t('cart.itemCount', { count: checkedItems.length }) }} · 运费 ${{ deliveryFee.toFixed(2) }}</span>
          <span v-else>{{ footerTipText }}</span>
        </template>
      </van-submit-bar>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { showSuccessToast, showToast, showDialog } from 'vant'
import { useCartStore } from '@/stores/cart'
import { useUserStore } from '@/stores/user'
import { createOrder, estimateDeliveryFeeByAddress } from '@/api'
import { formatKHR, usdToKhr } from '@/utils/format'
import { hapticFeedback } from '@/utils/device'

const { t } = useI18n()
const router = useRouter()
const cartStore = useCartStore()
const userStore = useUserStore()

const checkedItems = ref(cartStore.items.map(item => item.id))
const checkAll = ref(true)
const submitting = ref(false)
const clientRequestId = ref('')
const showPaymentPicker = ref(false)

// 运费估算（自动，基于用户地址 vs 仓库地址）
const distanceKm = ref(0)
const deliveryFee = ref(null)
const estimatingFee = ref(false)
const autoEstimated = ref(false)

// 根据用户保存的 Google 定位自动估算运费
const autoEstimateFromLocation = async () => {
  const locationUrl = userStore.userInfo?.location_url
  if (!locationUrl) return
  const m = locationUrl.match(/[?&]q=([-\d.]+),([-\d.]+)/)
  if (!m) return
  const destination = `${m[1]},${m[2]}`
  estimatingFee.value = true
  try {
    const res = await estimateDeliveryFeeByAddress('', destination)
    if (res.distance_km !== undefined) distanceKm.value = Number(res.distance_km)
    deliveryFee.value = res.delivery_fee_usd ?? null
    autoEstimated.value = true
  } catch {
    // 静默失败
  } finally {
    estimatingFee.value = false
  }
}

const totalWithFee = computed(() => {
  const base = cartStore.items
    .filter(item => checkedItems.value.includes(item.id))
    .reduce((sum, item) => sum + item.price_usd * item.quantity, 0)
  return base + (deliveryFee.value ?? 0)
})

const formEditable = computed(() => userStore.canOrder || userStore.orderAccessState === 'incomplete')

const handleInputFocus = (e) => {
  setTimeout(() => {
    if (e?.target) e.target.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }, 300)
}

const restrictionTitle = computed(() => {
  switch (userStore.orderAccessState) {
    case 'incomplete':
      return t('profile.orderGuideIncompleteTitle')
    case 'pending':
      return t('profile.orderGuidePendingTitle')
    case 'rejected':
      return t('profile.orderGuideRejectedTitle')
    default:
      return t('cart.checkoutDisabled')
  }
})

const restrictionDescription = computed(() => {
  if (userStore.orderAccessState === 'rejected' && userStore.userInfo?.rejected_reason) {
    return `${t('profile.rejectedReason')}：${userStore.userInfo.rejected_reason}。${t('profile.orderGuideRejectedDesc')}`
  }
  switch (userStore.orderAccessState) {
    case 'incomplete':
      return t('profile.orderGuideIncompleteDesc')
    case 'pending':
      return t('profile.orderGuidePendingDesc')
    case 'rejected':
      return t('profile.orderGuideRejectedDesc')
    default:
      return t('cart.checkoutDisabled')
  }
})

const restrictionActionLabel = computed(() => {
  switch (userStore.orderAccessState) {
    case 'incomplete':
      return t('profile.completeProfileAction')
    case 'pending':
      return t('profile.viewApprovalStatusAction')
    case 'rejected':
      return t('profile.resubmitForReview')
    default:
      return t('cart.submitOrder')
  }
})

const cartSavedTip = computed(() => {
  switch (userStore.orderAccessState) {
    case 'incomplete':
      return t('cart.draftInfoTip')
    case 'pending':
      return t('cart.savedForApprovalTip')
    case 'rejected':
      return t('cart.savedForResubmitTip')
    default:
      return t('cart.itemCount', { count: checkedItems.value.length })
  }
})

const cartFormTip = computed(() => {
  return formEditable.value ? t('cart.draftInfoTip') : cartSavedTip.value
})

const primaryButtonText = computed(() => {
  return userStore.canOrder ? t('cart.submitOrder') : restrictionActionLabel.value
})

const footerTipText = computed(() => {
  return userStore.canOrder ? t('cart.itemCount', { count: checkedItems.value.length }) : cartSavedTip.value
})

const paymentActions = computed(() => {
  const actions = [
    { name: t('cart.cashPayment'), value: 'cash' },
  ]
  if (userStore.userInfo?.allow_credit) {
    actions.push({ name: t('cart.monthlyPayment'), value: 'monthly' })
  }
  return actions
})

const buildClientRequestId = () => {
  return `mobile-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
}

const onPaymentSelect = (action) => {
  orderForm.value.payment_status = action.value
  showPaymentPicker.value = false
}

const openPaymentPicker = () => {
  if (!formEditable.value) {
    goToRestrictionAction()
    return
  }
  showPaymentPicker.value = true
}

const orderForm = ref({
  delivery_address: '',
  delivery_phone: '',
  payment_status: 'cash',
  note: '',
})

// 自动填充个人资料中的默认地址和电话，并根据定位自动估算运费
onMounted(() => {
  if (userStore.userInfo?.address) {
    orderForm.value.delivery_address = userStore.userInfo.address
  }
  if (userStore.userInfo?.phone) {
    orderForm.value.delivery_phone = userStore.userInfo.phone
  }
  autoEstimateFromLocation()
})

// 选中商品的总价
const totalPrice = computed(() => {
  return cartStore.items
    .filter(item => checkedItems.value.includes(item.id))
    .reduce((sum, item) => sum + item.price_usd * item.quantity, 0)
})

// 全选/取消全选
watch(checkAll, (val) => {
  if (val) {
    checkedItems.value = cartStore.items.map(item => item.id)
  } else {
    checkedItems.value = []
  }
})

watch(checkedItems, (val) => {
  checkAll.value = val.length === cartStore.items.length
})

// 更新数量
const updateQuantity = (id, quantity) => {
  hapticFeedback('light')
  cartStore.updateQuantity(id, quantity)
}

// 删除商品
const removeItem = async (id) => {
  hapticFeedback('heavy')
  const confirmed = await showDialog({
    title: t('cart.deleteConfirm'),
    message: t('cart.deleteMessage'),
    showCancelButton: true,
    confirmButtonText: t('common.confirm'),
    cancelButtonText: t('common.cancel'),
  }).catch(() => false)
  
  if (confirmed) {
    cartStore.removeItem(id)
    checkedItems.value = checkedItems.value.filter(itemId => itemId !== id)
  }
}

// 提交订单
const goToRestrictionAction = () => {
  showDialog({ message: restrictionDescription.value })
  router.push('/m/profile')
}

const handleBack = () => {
  if (window.history.length > 1) {
    router.back()
    return
  }
  router.replace('/m/shop')
}

const handleSubmit = async () => {
  if (submitting.value) {
    return
  }

  if (!userStore.canOrder) {
    goToRestrictionAction()
    return
  }

  if (checkedItems.value.length === 0) {
    showDialog({ message: t('cart.selectItems') })
    return
  }
  
  // 地址校验：表单地址和个人资料默认地址都没有时才拦截
  const address = orderForm.value.delivery_address?.trim() || userStore.userInfo?.address?.trim()
  if (!address) {
    showDialog({ message: t('cart.addressRequired') })
    return
  }
  
  submitting.value = true
  clientRequestId.value = clientRequestId.value || buildClientRequestId()
  hapticFeedback('medium')
  
  try {
    const items = cartStore.items
      .filter(item => checkedItems.value.includes(item.id))
      .map(item => ({
        product_id: item.id,
        quantity: item.quantity,
        purchase_mode: item.purchase_mode || 'default',
      }))

    if (items.length === 0) {
      showToast(t('cart.selectItems'))
      submitting.value = false
      return
    }
    
    await createOrder({
      items,
      ...orderForm.value,
      distance_km: distanceKm.value,
      client_request_id: clientRequestId.value,
    })
    
    // 清除已下单的商品
    checkedItems.value.forEach(id => {
      cartStore.removeItem(id)
    })
    
    hapticFeedback('success')
    showSuccessToast(t('cart.orderSuccess'))
    clientRequestId.value = ''
    router.push('/m/orders')
  } catch (error) {
    hapticFeedback('error')
    const msg = error?.message || ''
    if (msg.includes('Items') && msg.includes('min')) {
      showToast(t('cart.selectItems'))
    } else {
      showToast(msg || t('common.requestFailed'))
    }
  } finally {
    submitting.value = false
  }
}

const handlePrimaryAction = async () => {
  if (submitting.value) {
    return
  }
  if (!userStore.canOrder) {
    goToRestrictionAction()
    return
  }
  await handleSubmit()
}
</script>

<style scoped>
.mobile-cart {
  min-height: var(--tg-viewport-height, 100vh);
  background: #f5f5f5;
  padding-bottom: calc(100px + var(--tg-content-safe-area-inset-bottom, env(safe-area-inset-bottom, 0px)));
}

/* van-submit-bar 默认 bottom:0，会被导航栏遮挡，需偏移导航栏高度 */
:deep(.van-submit-bar) {
  bottom: calc(var(--van-tabbar-height, 50px) + env(safe-area-inset-bottom, 0px));
}

.restriction-panel {
  margin: 10px 12px 0;
  padding: 14px;
  background: #fff7e6;
  border: 1px solid #ffd591;
  border-radius: 12px;
}

.restriction-title {
  font-size: 15px;
  font-weight: 700;
  color: #ad6800;
}

.restriction-desc,
.restriction-note {
  margin-top: 6px;
  font-size: 13px;
  line-height: 1.6;
  color: #8c5a00;
}

.restriction-btn {
  margin-top: 12px;
}

.cart-item {
  display: flex;
  align-items: center;
  padding: 12px;
  background: #fff;
  gap: 12px;
}

.item-info {
  flex: 1;
  min-width: 0;
}

.item-name {
  font-size: 15px;
  font-weight: 500;
  color: #262626;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 4px;
}

.item-name-kh {
  font-size: 12px;
  color: #8c8c8c;
  margin-bottom: 4px;
}

.item-mode-tag {
  display: inline-block;
  font-size: 11px;
  color: #1D4ED8;
  background: #EFF6FF;
  border-radius: 4px;
  padding: 1px 6px;
  margin-bottom: 6px;
}

.item-price {
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.price-usd {
  font-size: 16px;
  font-weight: 700;
  color: #f5222d;
}

.price-khr {
  font-size: 12px;
  color: #8c8c8c;
}

.checkout-disabled-tip {
  margin-bottom: 8px;
}
</style>
