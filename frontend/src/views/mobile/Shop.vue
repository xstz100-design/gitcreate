<template>
  <div class="mobile-shop">
    <!-- 第一层：搜索栏 (Sticky) -->
    <div class="search-header">
      <div class="search-box">
        <van-icon name="search" class="search-icon" />
        <input
          v-model="searchKeyword"
          type="text"
          :placeholder="$t('product.searchPlaceholder')"
          class="search-input"
          @input="onSearch"
        />
        <van-icon
          v-if="searchKeyword"
          name="clear"
          class="clear-icon"
          @click="searchKeyword = ''; onSearch()"
        />
      </div>
      <button class="contact-btn" @click="contactVisible = true">
        <van-icon name="phone-o" />
        <span>{{ $t('common.contactUs') }}</span>
      </button>
    </div>

    <!-- 公告栏 -->
    <van-notice-bar
      v-if="notices.length > 0"
      :text="notices.map(n => currentLang === 'zh' ? n.content_zh : n.content_en).join('          ')"
      left-icon="volume-o"
      :speed="50"
      background="#fffbe6"
      color="#ad6800"
    />

    <div v-if="!userStore.canOrder" class="browse-guide">
      <div class="guide-copy">
        <div class="guide-title">{{ restrictionTitle }}</div>
        <div class="guide-desc">{{ restrictionDescription }}</div>
      </div>
      <van-button size="small" round type="warning" @click="goToRestrictionAction">
        {{ restrictionActionLabel }}
      </van-button>
    </div>

    <!-- 新品推荐 -->
    <div v-if="newProducts.length > 0" class="new-products-section">
      <div class="np-header">
        <div class="np-title-group">
          <span class="np-icon">🔥</span>
          <span class="np-title">{{ $t('product.newProducts') }}</span>
        </div>
        <span class="np-more" @click="scrollToProducts">{{ $t('common.viewAll') }} ›</span>
      </div>
      <div class="np-scroll">
        <div
          v-for="product in newProducts"
          :key="product.id"
          class="np-card"
          @click="showDetail(product)"
        >
          <div class="np-img">
            <img v-if="getProductImage(product)" :src="getProductImage(product)" :alt="product.name" loading="lazy" />
            <div v-else class="np-img-placeholder">
              <van-icon name="photo-o" size="20" color="#ccc" />
            </div>
            <span class="np-badge">NEW</span>
          </div>
          <div class="np-info">
            <div class="np-name">{{ product.name }}</div>
            <div class="np-price">${{ product.price_usd }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 第二层：品类导航 (横向滚动 Tabs) -->
    <div class="category-tabs">
      <div class="tabs-scroll" ref="tabsRef">
        <div
          v-for="cat in categories"
          :key="cat.id"
          class="tab-item"
          :class="{ active: activeCat === cat.id }"
          @click="switchCategory(cat.id)"
        >
          {{ cat.name }}
        </div>
      </div>
    </div>

    <!-- 第三层：商品列表 (Grid) -->
    <div class="product-area" ref="scrollContainer">
      <!-- 骨架屏 -->
      <div v-if="loading" class="skeleton-grid">
        <div v-for="i in 6" :key="i" class="skeleton-card">
          <div class="skeleton-img shimmer"></div>
          <div class="skeleton-name shimmer"></div>
          <div class="skeleton-price shimmer"></div>
        </div>
      </div>

      <!-- 空状态 -->
      <van-empty v-else-if="filteredProducts.length === 0" :description="$t('common.noData')" />

      <!-- 商品网格 -->
      <div v-else class="product-grid">
        <div
          v-for="product in filteredProducts"
          :key="product.id"
          class="product-card"
          @click="showDetail(product)"
        >
          <!-- 正方形图片 -->
          <div class="card-image">
            <img
              v-if="getProductImage(product)"
              :src="getProductImage(product)"
              :alt="product.name"
              loading="lazy"
            />
            <div v-else class="image-placeholder">
              <van-icon name="photo-o" size="28" color="#ccc" />
            </div>
            <!-- 库存标记 -->
            <span v-if="product.stock <= 0" class="badge-out">{{ $t('product.outOfStock') }}</span>
            <span v-else-if="product.is_low_stock" class="badge-low">{{ $t('product.lowStock') }}</span>
            <!-- 推荐标记 -->
            <span v-if="product.is_featured" class="badge-featured">{{ $t('product.recommended') }}</span>
          </div>

          <!-- 名称 -->
          <div class="card-name">{{ product.name }}</div>

          <!-- 参考价 -->
          <div class="card-price">
            <span class="price-val">${{ product.price_usd }}</span>
            <span class="price-unit">/{{ product.unit }}</span>
          </div>

          <!-- 加购按钮 -->
          <div class="card-action" @click.stop>
            <button
              v-if="!userStore.canOrder"
              class="add-btn guide-btn"
              @click="goToRestrictionAction"
            >
              {{ restrictionActionLabel }}
            </button>
            <div v-else-if="getCartQty(product.id) > 0" class="qty-control">
              <button class="qty-btn" @click="decrementCart(product)">−</button>
              <input
                v-if="editingQtyId === product.id"
                class="qty-input"
                type="number"
                :value="getCartQty(product.id)"
                @blur="onQtyInputBlur(product, $event)"
                @keyup.enter="$event.target.blur()"
                ref="qtyInputRef"
                autofocus
              />
              <span v-else class="qty-num" @click="startEditQty(product)">{{ getCartQty(product.id) }}</span>
              <button class="qty-btn" @click="incrementCart(product)">+</button>
            </div>
            <button
              v-else
              class="add-btn"
              :disabled="product.stock <= 0"
              @click="incrementCart(product)"
            >
              {{ $t('product.addToCart') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 底部结算栏 -->
    <div v-if="cartStore.totalCount > 0" class="checkout-bar">
      <div class="cart-info" @click="goToCart">
        <div class="cart-icon-wrap">
          <van-icon name="shopping-cart-o" size="24" />
          <span class="cart-badge">{{ cartStore.totalCount }}</span>
        </div>
        <div class="cart-total">
          <span class="total-label">{{ $t('common.total') }}</span>
          <span class="total-price">${{ cartStore.totalPrice.toFixed(2) }}</span>
        </div>
      </div>
      <button class="checkout-btn" @click="goToCart">{{ $t('product.checkout') }}</button>
    </div>

    <!-- 商品详情弹窗 -->
    <van-popup
      v-model:show="detailVisible"
      position="bottom"
      teleport="body"
      :style="{ height: '85%' }"
      round
      closeable
    >
      <div v-if="currentProduct" class="detail-sheet" @focusin="handleDetailFocusin">
        <!-- 图片轮播 -->
        <van-swipe class="detail-swipe" :autoplay="0" indicator-color="#1D4ED8">
          <van-swipe-item v-for="(img, idx) in getProductImages(currentProduct)" :key="idx">
            <img :src="img" class="detail-img" @click="previewImage(img, getProductImages(currentProduct))" />
          </van-swipe-item>
          <van-swipe-item v-if="getProductImages(currentProduct).length === 0">
            <div class="detail-img-empty">
              <van-icon name="photo-o" size="48" color="#ccc" />
            </div>
          </van-swipe-item>
        </van-swipe>

        <div class="detail-body">
          <h3 class="detail-name">{{ currentProduct.name }}</h3>
          <p v-if="currentProduct.name_kh" class="detail-kh">{{ currentProduct.name_kh }}</p>

          <!-- 购买规格表（参考电商下单界面：规格/单价/数量三列） -->
          <div class="purchase-table">
            <div class="pt-header">
              <span class="pt-col-spec">{{ $t('product.specs') }}</span>
              <span class="pt-col-price">{{ $t('product.tableUnitPrice') }}</span>
              <span class="pt-col-qty">{{ $t('product.tableQty') }}</span>
            </div>
            <!-- 默认单件 -->
            <div class="pt-row">
              <span class="pt-spec">{{ currentProduct.unit_name || currentProduct.unit || $t('product.unitFallback') }}</span>
              <span class="pt-price">${{ Number(currentProduct.price_usd).toFixed(2) }}<em>/{{ currentProduct.unit || $t('product.unitFallback') }}</em></span>
              <div class="pt-stepper" @click.stop>
                <van-stepper
                  v-model="detailQtyDefault"
                  :min="0"
                  :max="currentProduct.stock"
                  integer
                  button-size="26px"
                  input-width="36px"
                />
              </div>
            </div>
            <!-- 箱装（只要设了每箱件数就显示，价格=单独设置价 或 参考价×件数） -->
            <div v-if="currentProduct.pieces_per_package" class="pt-row">
              <span class="pt-spec">
                {{ currentProduct.pack_name || $t('product.caseFallback') }}
                <span class="pt-spec-hint">({{ currentProduct.pieces_per_package }}{{ currentProduct.unit || $t('product.unitFallback') }}/{{ currentProduct.pack_name || $t('product.caseFallback') }})</span>
              </span>
              <span class="pt-price">
                <template v-if="currentProduct.price_per_package_usd">
                  <s class="pt-price-original">${{ (Number(currentProduct.price_usd) * Number(currentProduct.pieces_per_package)).toFixed(2) }}</s>
                  ${{ Number(currentProduct.price_per_package_usd).toFixed(2) }}
                </template>
                <template v-else>
                  ${{ (Number(currentProduct.price_usd) * Number(currentProduct.pieces_per_package)).toFixed(2) }}
                </template>
                <em>/{{ currentProduct.pack_name || $t('product.caseFallback') }}</em>
              </span>
              <div class="pt-stepper" @click.stop>
                <van-stepper
                  v-model="detailQtyPackage"
                  :min="0"
                  integer
                  button-size="26px"
                  input-width="36px"
                />
              </div>
            </div>
          </div>

          <div class="detail-meta">
            <span v-if="currentProduct.specs" class="meta-item">{{ $t('product.specs') }}：{{ currentProduct.specs }}</span>
            <span>{{ $t('product.stock') }}：{{ currentProduct.stock }} {{ currentProduct.unit }}</span>
            <span>{{ $t('product.category') }}：{{ currentProduct.category || $t('product.uncategorized') }}</span>
          </div>

          <!-- 商品描述 -->
          <div v-if="currentProduct.description" class="detail-desc">
            <div class="desc-title">{{ $t('product.description') }}</div>
            <div class="desc-content">{{ currentProduct.description }}</div>
          </div>

          <!-- 最小订货量 -->
          <div v-if="currentProduct.min_order_qty > 1" class="detail-notice">
            {{ $t('product.minOrderQty') }}：{{ currentProduct.min_order_qty }} {{ currentProduct.unit }}
          </div>
        </div>

        <!-- 加入购物车操作栏 -->
        <div class="detail-footer">
          <button
            class="detail-add-btn"
            :disabled="currentProduct.stock <= 0 || (!detailQtyDefault && !detailQtyPackage)"
            @click="handleDetailPrimaryAction"
          >
            {{ userStore.canOrder ? $t('product.addToCart') : restrictionActionLabel }}
          </button>
        </div>
      </div>
    </van-popup>
    <!-- 商品详情弹窗 -->
    <van-popup
      v-model:show="contactVisible"
      position="bottom"
      teleport="body"
      round
      :style="{ padding: '24px 20px 32px' }"
    >
      <div class="contact-sheet">
        <div class="contact-title">{{ $t('common.contactUs') }}</div>
        <div v-if="!contactInfo.phone && !contactInfo.telegram && !contactInfo.whatsapp && !contactInfo.wechat" class="contact-empty">
          联系方式尚未设置
        </div>
        <a v-if="contactInfo.phone" :href="'tel:' + contactInfo.phone" class="contact-item">
          <van-icon name="phone-o" size="20" />
          <span>{{ contactInfo.phone }}</span>
        </a>
        <a v-if="contactInfo.whatsapp" :href="'https://wa.me/' + contactInfo.whatsapp.replace(/\D/g, '')" target="_blank" class="contact-item">
          <van-icon name="chat-o" size="20" />
          <span>WhatsApp: {{ contactInfo.whatsapp }}</span>
        </a>
        <a v-if="contactInfo.telegram" :href="'https://t.me/' + contactInfo.telegram.replace(/^@/, '')" target="_blank" class="contact-item">
          <van-icon name="send-gift-o" size="20" />
          <span>Telegram: {{ contactInfo.telegram }}</span>
        </a>
        <div v-if="contactInfo.wechat" class="contact-item" @click="copyWechat">
          <van-icon name="wechat-o" size="20" />
          <span>WeChat: {{ contactInfo.wechat }}</span>
        </div>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { showSuccessToast, showToast, showImagePreview } from 'vant'
import { getProducts, getCategories, getPublicAnnouncements, getContactInfo } from '@/api'
import { useCartStore } from '@/stores/cart'
import { useUserStore } from '@/stores/user'
import { formatKHR, usdToKhr } from '@/utils/format'
import { hapticFeedback } from '@/utils/device'
import { getCurrentLanguage } from '@/i18n'

const router = useRouter()
const cartStore = useCartStore()
const userStore = useUserStore()
const { t } = useI18n()

const loading = ref(true)
const products = ref([])
const searchKeyword = ref('')
const activeCat = ref('all')
const categories = ref([{ id: 'all', name: t('product.all') }])
const notices = ref([])
const currentLang = ref(getCurrentLanguage())

// 联系我
const contactVisible = ref(false)
const contactInfo = ref({ phone: '', telegram: '', whatsapp: '', wechat: '' })

const loadContactInfo = async () => {
  try {
    contactInfo.value = await getContactInfo()
  } catch {
    // 静默失败
  }
}

const copyWechat = () => {
  if (!contactInfo.value.wechat) return
  navigator.clipboard?.writeText(contactInfo.value.wechat).then(() => {
    showSuccessToast('已复制微信号')
  }).catch(() => {
    showToast('WeChat: ' + contactInfo.value.wechat)
  })
}

// 键盘弹出后滚动输入框到可视区
const handleDetailFocusin = (e) => {
  if (e.target?.tagName === 'INPUT' || e.target?.tagName === 'TEXTAREA') {
    setTimeout(() => {
      e.target.scrollIntoView({ behavior: 'smooth', block: 'center' })
    }, 350)
  }
}

// 详情弹窗
const detailVisible = ref(false)
const currentProduct = ref(null)
const detailQtyDefault = ref(0)
const detailQtyPackage = ref(0)

const restrictionTitle = computed(() => {
  switch (userStore.orderAccessState) {
    case 'incomplete':
      return t('profile.orderGuideIncompleteTitle')
    case 'pending':
      return t('profile.orderGuidePendingTitle')
    case 'rejected':
      return t('profile.orderGuideRejectedTitle')
    default:
      return t('product.browseOnlyTip')
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
      return t('product.browseOnlyTip')
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
      return t('product.addToCart')
  }
})

// 加载分类
const loadCategories = async () => {
  try {
    const data = await getCategories()
    categories.value = [
      { id: 'all', name: t('product.all') },
      ...data.map(c => ({ id: c.name, name: c.name })),
    ]
  } catch {
    // 回退：从商品中提取分类
  }
}

// 加载商品
const loadProducts = async () => {
  loading.value = true
  try {
    const data = await getProducts({ is_active: true })
    products.value = data
  } catch {
    showToast(t('product.loadFailed'))
  } finally {
    loading.value = false
  }
}

// 过滤商品
const filteredProducts = computed(() => {
  let list = products.value
  if (activeCat.value !== 'all') {
    list = list.filter(p => p.category === activeCat.value)
  }
  if (searchKeyword.value.trim()) {
    const kw = searchKeyword.value.toLowerCase()
    list = list.filter(p =>
      p.name.toLowerCase().includes(kw) ||
      (p.name_kh && p.name_kh.toLowerCase().includes(kw)) ||
      (p.name_en && p.name_en.toLowerCase().includes(kw)) ||
      (p.brand && p.brand.toLowerCase().includes(kw))
    )
  }
  // 推荐商品置顶
  return [...list].sort((a, b) => {
    if (a.is_featured && !b.is_featured) return -1
    if (!a.is_featured && b.is_featured) return 1
    return (a.sort_order || 0) - (b.sort_order || 0)
  })
})

// 新品推荐 — 仅管理员勾选推荐的商品
const newProducts = computed(() => {
  if (!products.value.length) return []
  return [...products.value]
    .filter(p => p.is_featured)
    .sort((a, b) => b.id - a.id)
    .slice(0, 10)
})

// 滚动到商品区域
const scrollToProducts = () => {
  const el = document.querySelector('.product-area')
  if (el) el.scrollIntoView({ behavior: 'smooth' })
}

// 获取商品首图
const getProductImage = (product) => {
  return product.img1 || product.image_url || ''
}

// 获取商品所有图片
const getProductImages = (product) => {
  const imgs = []
  if (product.img1) imgs.push(product.img1)
  if (product.img2) imgs.push(product.img2)
  if (product.img3) imgs.push(product.img3)
  if (product.img4) imgs.push(product.img4)
  if (product.img5) imgs.push(product.img5)
  if (imgs.length === 0 && product.image_url) imgs.push(product.image_url)
  return imgs
}

// 购物车数量
const getCartQty = (id) => {
  const item = cartStore.items.find(i => i.id === id)
  return item ? item.quantity : 0
}

const incrementCart = (product) => {
  if (product.stock <= 0) return
  if (!userStore.canOrder) {
    goToRestrictionAction()
    return
  }
  hapticFeedback('light')
  cartStore.addItem(product, 1)
}

const decrementCart = (product) => {
  const current = getCartQty(product.id)
  hapticFeedback('light')
  if (current <= 1) {
    cartStore.removeItem(product.id)
  } else {
    cartStore.updateQuantity(product.id, current - 1)
  }
}

// 点击数量进入编辑模式
const editingQtyId = ref(null)

const startEditQty = (product) => {
  editingQtyId.value = product.id
}

const onQtyInputBlur = (product, event) => {
  const val = parseInt(event.target.value)
  if (!isNaN(val) && val > 0) {
    const qty = Math.min(val, product.stock)
    cartStore.updateQuantity(product.id, qty)
  } else if (val === 0 || event.target.value === '') {
    cartStore.removeItem(product.id)
  }
  editingQtyId.value = null
}

// 品类切换
const switchCategory = (id) => {
  activeCat.value = id
  hapticFeedback('light')
}

const onSearch = () => {
  // 搜索时重置分类到全部，确保搜索结果不受分类过滤影响
  if (searchKeyword.value.trim()) {
    activeCat.value = 'all'
  }
}

// 商品详情
const showDetail = (product) => {
  currentProduct.value = product
  detailQtyDefault.value = 0
  detailQtyPackage.value = 0
  detailVisible.value = true
  hapticFeedback('medium')
}

// 图片预览（全屏大图）
const previewImage = (current, images) => {
  showImagePreview({
    images,
    startPosition: images.indexOf(current),
    closeable: true,
  })
}

const addFromDetail = () => {
  if (!currentProduct.value) return
  if (!userStore.canOrder) {
    goToRestrictionAction()
    return
  }
  let added = false
  if (detailQtyDefault.value > 0) {
    cartStore.addItem(currentProduct.value, detailQtyDefault.value, 'default')
    added = true
  }
  if (detailQtyPackage.value > 0 && currentProduct.value.pieces_per_package) {
    cartStore.addItem(currentProduct.value, detailQtyPackage.value, 'package')
    added = true
  }
  if (added) {
    hapticFeedback('success')
    showSuccessToast(t('product.addedToCart'))
    detailVisible.value = false
  } else {
    showToast('请先选择购买数量')
  }
}

const goToRestrictionAction = () => {
  showToast(restrictionDescription.value)
  router.push('/m/profile')
}

const handleDetailPrimaryAction = () => {
  if (!userStore.canOrder) {
    goToRestrictionAction()
    return
  }
  addFromDetail()
}

// 加载公告
const loadNotices = async () => {
  try {
    const data = await getPublicAnnouncements('notice')
    notices.value = data
  } catch {
    // 静默处理
  }
}

const goToCart = () => {
  hapticFeedback('medium')
  router.push('/m/cart')
}

onMounted(async () => {
  await Promise.all([loadCategories(), loadProducts(), loadNotices(), loadContactInfo()])
  // 从商品中补充分类
  if (categories.value.length <= 1) {
    const cats = new Set()
    products.value.forEach(p => { if (p.category) cats.add(p.category) })
    cats.forEach(c => categories.value.push({ id: c, name: c }))
  }
})
</script>

<style scoped>
.mobile-shop {
  min-height: var(--tg-viewport-height, 100vh);
  background: var(--bg-gray, #f7f7f7);
  padding-bottom: calc(60px + var(--tg-content-safe-area-inset-bottom, env(safe-area-inset-bottom, 0px)));
  overflow: visible;
}

/* ---------- 搜索栏 ---------- */
.search-header {
  background: #fff;
  padding: 8px 12px;
  border-bottom: 1px solid #F0F0F0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.search-box {
  flex: 1;
  display: flex;
  align-items: center;
  background: #f5f5f5;
  border-radius: 4px;
  padding: 0 10px;
  height: 36px;
}

.contact-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
  padding: 0 10px;
  height: 36px;
  border: 1.5px solid #d44e4e;
  border-radius: 6px;
  background: #fff;
  color: #d44e4e;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
}

.contact-sheet {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.contact-title {
  font-size: 17px;
  font-weight: 700;
  color: #1a1a1a;
  margin-bottom: 4px;
  text-align: center;
}

.contact-empty {
  text-align: center;
  color: #aaa;
  font-size: 14px;
  padding: 12px 0;
}

.contact-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  background: #f8f8f8;
  border-radius: 10px;
  color: #1a1a1a;
  font-size: 15px;
  text-decoration: none;
  border: 1px solid #eee;
}

.search-icon {
  color: #999;
  font-size: 16px;
  margin-right: 6px;
}

.search-input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: 14px;
  color: #1a1a1a;
}

.search-input::placeholder {
  color: #bfbfbf;
}

.clear-icon {
  color: #999;
  font-size: 16px;
  margin-left: 4px;
}

/* ---------- 公告栏 ---------- */
.notice-bar {
  display: flex;
  align-items: center;
  background: #fffbe6;
  padding: 6px 12px;
  border-bottom: 1px solid #fff1b8;
  overflow: hidden;
}

.notice-icon {
  color: #fa8c16;
  font-size: 16px;
  flex-shrink: 0;
  margin-right: 8px;
}

.notice-scroll {
  flex: 1;
  overflow: hidden;
  white-space: nowrap;
}

.notice-track {
  display: inline-block;
  white-space: nowrap;
  animation: noticeScroll linear infinite;
}

.notice-item {
  font-size: 13px;
  color: #ad6800;
}

.browse-guide {
  margin: 8px 12px 0;
  padding: 12px;
  border-radius: 10px;
  background: #fff7e6;
  border: 1px solid #ffd591;
  display: flex;
  align-items: center;
  gap: 12px;
}

.guide-copy {
  flex: 1;
  min-width: 0;
}

.guide-title {
  font-size: 14px;
  font-weight: 700;
  color: #ad6800;
}

.guide-desc {
  margin-top: 4px;
  font-size: 12px;
  line-height: 1.5;
  color: #8c5a00;
}

.guide-btn {
  background: #fa8c16;
}

@keyframes noticeScroll {
  0% { transform: translateX(100%); }
  100% { transform: translateX(-100%); }
}

/* ---------- 新品推荐 ---------- */
.new-products-section {
  background: #fff;
  margin: 0;
  border-bottom: 1px solid #f0f0f0;
}

.np-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px 6px;
}

.np-title-group {
  display: flex;
  align-items: center;
  gap: 4px;
}

.np-icon {
  font-size: 16px;
}

.np-title {
  font-size: 15px;
  font-weight: 700;
  color: #1a1a1a;
}

.np-more {
  font-size: 12px;
  color: #999;
}

.np-scroll {
  display: flex;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
  padding: 0 12px 12px;
  gap: 8px;
}

.np-scroll::-webkit-scrollbar {
  display: none;
}

.np-card {
  flex-shrink: 0;
  width: 100px;
  background: #fafafa;
  border-radius: 6px;
  overflow: hidden;
  border: 1px solid #f0f0f0;
}

.np-img {
  position: relative;
  width: 100%;
  height: 100px;
  background: #f5f5f5;
  overflow: hidden;
}

.np-img img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.np-img-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.np-badge {
  position: absolute;
  top: 4px;
  left: 4px;
  background: linear-gradient(135deg, #ff4d4f, #ff7a45);
  color: #fff;
  font-size: 9px;
  font-weight: 700;
  padding: 1px 5px;
  border-radius: 3px;
  letter-spacing: 0.5px;
}

.np-info {
  padding: 6px 6px 8px;
}

.np-name {
  font-size: 12px;
  font-weight: 500;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.np-price {
  font-size: 13px;
  font-weight: 700;
  color: #e8222e;
  margin-top: 2px;
}

/* ---------- 品类导航 ---------- */
.category-tabs {
  position: sticky;
  top: 0;
  z-index: 99;
  background: #fff;
  border-bottom: 1px solid #F0F0F0;
}

.tabs-scroll {
  display: flex;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
  padding: 0 4px;
}

.tabs-scroll::-webkit-scrollbar {
  display: none;
}

.tab-item {
  flex-shrink: 0;
  padding: 10px 16px;
  font-size: 13px;
  color: #666;
  white-space: nowrap;
  position: relative;
  transition: color 0.2s;
}

.tab-item.active {
  color: var(--primary-color, #2b2b2b);
  font-weight: 600;
}

.tab-item.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 20px;
  height: 2px;
  background: var(--primary-color, #2b2b2b);
  border-radius: 1px;
}

/* ---------- 商品区域 ---------- */
.product-area {
  padding: 8px;
}

/* ---------- 商品网格 ---------- */
.product-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
}

.product-card {
  background: #fff;
  border-radius: 4px;
  overflow: hidden;
  border: 1px solid var(--border-color, #eee);
}

.card-image {
  position: relative;
  width: 100%;
  padding-top: 100%;
  background: #fafafa;
  overflow: hidden;
}

.card-image img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.image-placeholder {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fafafa;
}

.badge-out,
.badge-low {
  position: absolute;
  top: 6px;
  right: 6px;
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 2px;
  font-weight: 500;
}

.badge-out {
  background: #fef2f2;
  color: #dc2626;
  border: 1px solid #dc2626;
}

.badge-low {
  background: #fff7ed;
  color: #ea580c;
  border: 1px solid #ea580c;
}

.badge-featured {
  position: absolute;
  top: 6px;
  left: 6px;
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 2px;
  font-weight: 500;
  background: #fff0f0;
  color: #e8222e;
  border: 1px solid #e8222e;
}

.card-name {
  padding: 8px 8px 2px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #1a1a1a);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-price {
  padding: 0 8px 0;
}

.price-val {
  font-size: 16px;
  font-weight: 700;
  color: var(--price-color, #d44e4e);
}

.price-unit {
  font-size: 11px;
  color: #999;
}

.card-retail {
  padding: 0 8px 4px;
  font-size: 11px;
  color: #999;
}

.retail-label {
  color: #bbb;
}

.retail-val {
  color: #e6a23c;
  font-weight: 500;
}

.card-action {
  padding: 4px 8px 8px;
}

.add-btn {
  width: 100%;
  height: 30px;
  border: 1px solid var(--border-color, #eee);
  background: #fff;
  color: var(--primary-color, #2b2b2b);
  font-size: 13px;
  font-weight: 500;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s;
}

.add-btn:active {
  background: #f5f5f5;
  transform: scale(0.97);
}

.add-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.qty-control {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0;
}

.qty-btn {
  width: 30px;
  height: 30px;
  border: 1px solid var(--border-color, #eee);
  background: #fff;
  font-size: 16px;
  font-weight: 500;
  color: var(--primary-color, #2b2b2b);
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.qty-btn:active {
  background: #f0f0f0;
}

.qty-num {
  width: 36px;
  text-align: center;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #1a1a1a);
  cursor: pointer;
}

.qty-input {
  width: 42px;
  text-align: center;
  font-size: 14px;
  font-weight: 600;
  border: 1px solid var(--color-primary, #1D4ED8);
  border-radius: 4px;
  outline: none;
  padding: 2px 0;
  -moz-appearance: textfield;
}
.qty-input::-webkit-inner-spin-button,
.qty-input::-webkit-outer-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

/* ---------- 底部结算栏 ---------- */
.checkout-bar {
  position: fixed;
  bottom: calc(50px + env(safe-area-inset-bottom, 0px));
  left: 0;
  right: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  background: #fff;
  border-top: 1px solid #F0F0F0;
  padding: 8px 12px;
}

.cart-info {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 10px;
}

.cart-icon-wrap {
  position: relative;
}

.cart-badge {
  position: absolute;
  top: -6px;
  right: -8px;
  background: var(--price-color, #d44e4e);
  color: #fff;
  font-size: 10px;
  min-width: 16px;
  height: 16px;
  line-height: 16px;
  text-align: center;
  border-radius: 8px;
  padding: 0 4px;
  font-weight: 600;
}

.total-label {
  font-size: 12px;
  color: #999;
}

.total-price {
  font-size: 18px;
  font-weight: 700;
  color: var(--price-color, #d44e4e);
  margin-left: 4px;
}

.checkout-btn {
  padding: 0 24px;
  height: 36px;
  background: var(--primary-color, #2b2b2b);
  color: #fff;
  font-size: 14px;
  font-weight: 500;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.checkout-btn:active {
  opacity: 0.85;
  transform: scale(0.97);
}

/* ---------- 骨架屏 ---------- */
.skeleton-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
}

.skeleton-card {
  background: #fff;
  border-radius: 4px;
  overflow: hidden;
  border: 1px solid var(--border-color, #eee);
}

.skeleton-img {
  width: 100%;
  padding-top: 100%;
}

.skeleton-name {
  height: 16px;
  margin: 8px;
  width: 60%;
  border-radius: 2px;
}

.skeleton-price {
  height: 18px;
  margin: 0 8px 8px;
  width: 40%;
  border-radius: 2px;
}

.shimmer {
  background: linear-gradient(90deg, #f0f0f0 0%, #f8f8f8 50%, #f0f0f0 100%);
  background-size: 200% 100%;
  animation: shimmer 1.5s ease-in-out infinite;
}

@keyframes shimmer {
  0% { background-position: -200% 0; }
  100% { background-position: 200% 0; }
}

/* ---------- 详情弹窗 ---------- */
.detail-sheet {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.detail-swipe {
  height: 280px;
  background: #fafafa;
}

.detail-img {
  width: 100%;
  height: 280px;
  object-fit: cover;
  cursor: pointer;
}

.detail-img-empty {
  width: 100%;
  height: 280px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fafafa;
}

.detail-body {
  flex: 1;
  padding: 16px;
  padding-bottom: 0;
  overflow-y: auto;
}

.detail-name {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary, #1a1a1a);
  margin: 0 0 4px;
}

.detail-kh {
  font-size: 13px;
  color: #999;
  margin: 0 0 12px;
}

/* ===== 购买规格表 ===== */
.purchase-table {
  margin: 12px 0;
  border: 1px solid #eef0f3;
  border-radius: 10px;
  overflow: hidden;
}

.pt-header {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  background: #f8fafc;
  padding: 8px 12px;
  font-size: 12px;
  color: #8c8c8c;
  font-weight: 500;
}

.pt-row {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  align-items: center;
  padding: 10px 12px;
  border-top: 1px solid #eef0f3;
  transition: background 0.15s;
}

.pt-row:active {
  background: #f5f8ff;
}

.pt-spec {
  font-size: 13px;
  color: #333;
  font-weight: 500;
}

.pt-spec-hint {
  display: block;
  font-size: 11px;
  color: #999;
  font-weight: 400;
  margin-top: 1px;
}

.pt-price {
  font-size: 14px;
  font-weight: 700;
  color: #d44e4e;
}

.pt-price em {
  font-style: normal;
  font-size: 11px;
  font-weight: 400;
  color: #999;
}

.pt-price-original {
  font-size: 11px;
  font-weight: 400;
  color: #bbb;
  text-decoration: line-through;
  margin-right: 3px;
}

.pt-col-spec { }
.pt-col-price { }
.pt-col-qty { text-align: right; }

.pt-stepper {
  display: flex;
  justify-content: flex-end;
}

:deep(.purchase-table .van-stepper__input) {
  background: transparent;
}

.detail-mode-toggle {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.mode-btn {
  flex: 1;
  padding: 8px 4px;
  border: 1.5px solid #ddd;
  border-radius: 8px;
  background: #f5f5f5;
  color: #666;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s;
}

.mode-btn.active {
  border-color: var(--primary-color, #1D4ED8);
  background: #EFF6FF;
  color: var(--primary-color, #1D4ED8);
  font-weight: 600;
}

.mode-hint {
  font-size: 11px;
  color: inherit;
  opacity: 0.8;
}

.detail-price-row {
  display: flex;
  align-items: baseline;
  gap: 10px;
  margin-bottom: 12px;
}

.detail-price {
  font-size: 24px;
  font-weight: 700;
  color: var(--price-color, #d44e4e);
}

.detail-unit {
  font-size: 13px;
  color: #888;
}

.detail-khr {
  font-size: 12px;
  color: #999;
}

.detail-retail-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  padding: 8px 12px;
  background: #fffbeb;
  border-radius: 6px;
  border: 1px solid #fde68a;
}

.retail-tag {
  font-size: 12px;
  color: #92400e;
  font-weight: 500;
}

.retail-price {
  font-size: 16px;
  font-weight: 700;
  color: #d97706;
}

.retail-khr {
  font-size: 12px;
  color: #b45309;
}

.detail-meta {
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: #666;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.detail-desc {
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.desc-title {
  font-size: 14px;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 6px;
}

.desc-content {
  font-size: 13px;
  color: #666;
  line-height: 1.6;
  white-space: pre-wrap;
}

.detail-notice {
  font-size: 13px;
  color: #ea580c;
  background: #fff7ed;
  padding: 8px 12px;
  border-radius: 6px;
  margin-bottom: 12px;
}

.detail-footer {
  padding: 10px 16px calc(10px + env(safe-area-inset-bottom, 0px));
  background: #fff;
  border-top: 1px solid #f0f0f0;
  flex-shrink: 0;
}

.qty-label {
  font-size: 14px;
  color: #333;
  font-weight: 500;
}

.detail-add-btn {
  width: 100%;
  height: 44px;
  background: var(--primary-color, #2b2b2b);
  color: #fff;
  font-size: 15px;
  font-weight: 500;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.detail-add-btn:active {
  opacity: 0.85;
  transform: scale(0.98);
}

.detail-add-btn:disabled {
  opacity: 0.4;
}
</style>
