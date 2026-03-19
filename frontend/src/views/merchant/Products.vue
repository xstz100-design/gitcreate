<template>
  <div class="products-page">
    <!-- 搜索栏 -->
    <div class="search-bar">
      <el-input
        v-model="searchKeyword"
        :placeholder="$t('product.searchPlaceholder')"
        :prefix-icon="Search"
        clearable
        size="large"
        class="search-input"
      />
    </div>

    <!-- 公告栏 -->
    <div v-if="notices.length > 0" class="notice-bar">
      <el-icon class="notice-icon"><bell /></el-icon>
      <div class="notice-scroll">
        <div class="notice-track" :style="{ animationDuration: notices.length * 15 + 's' }">
          <span v-for="(n, i) in notices" :key="i" class="notice-item">{{ currentLang === 'zh' ? n.content_zh : n.content_en }}&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;</span>
        </div>
      </div>
    </div>

    <!-- 分类导航 -->
    <div class="category-tabs">
      <div
        v-for="cat in allCategories"
        :key="cat.id"
        class="tab-item"
        :class="{ active: activeCat === cat.id }"
        @click="activeCat = cat.id"
      >
        {{ cat.name }}
      </div>
    </div>

    <!-- 骨架屏 -->
    <div v-if="loading" class="product-grid">
      <div v-for="i in 8" :key="i" class="skeleton-card">
        <div class="skeleton-img"></div>
        <div class="skeleton-line w60"></div>
        <div class="skeleton-line w40"></div>
      </div>
    </div>

    <!-- 空状态 -->
    <el-empty v-else-if="filteredProducts.length === 0" :description="$t('common.noData')" />

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
            <el-icon :size="32" color="#ccc"><Picture /></el-icon>
          </div>
          <span v-if="product.stock <= 0" class="badge badge-out">{{ $t('product.outOfStock') }}</span>
          <span v-else-if="product.is_low_stock" class="badge badge-low">{{ $t('product.lowStock') }}</span>
          <span v-if="product.is_featured" class="badge badge-featured">{{ $t('product.recommended') }}</span>
        </div>

        <!-- 商品名称 -->
        <div class="card-name">{{ product.name }}</div>

        <!-- 价格行 -->
        <div class="card-price">
          <span class="price-val">${{ product.price_usd }}</span>
          <span class="price-unit">/{{ product.unit }}</span>
        </div>

        <!-- 建议零售价 -->
        <div v-if="product.retail_price_usd" class="card-retail">
          <span class="retail-label">{{ $t('product.retailPrice') }}:</span>
          <span class="retail-val">${{ product.retail_price_usd }}</span>
        </div>

        <!-- 加购操作 -->
        <div class="card-action" @click.stop>
          <div v-if="getCartQty(product.id) > 0" class="qty-control">
            <button class="qty-btn" @click="decrementCart(product)">−</button>
            <input
              v-if="editingQtyId === product.id"
              class="qty-input"
              type="number"
              :value="getCartQty(product.id)"
              @blur="onQtyInputBlur(product, $event)"
              @keyup.enter="$event.target.blur()"
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

    <!-- 底部结算栏 -->
    <div v-if="cartStore.totalCount > 0" class="checkout-bar">
      <div class="cart-info" @click="$router.push('/merchant/cart')">
        <div class="cart-icon-wrap">
          <el-icon :size="24"><ShoppingCart /></el-icon>
          <span class="cart-badge">{{ cartStore.totalCount }}</span>
        </div>
        <div class="cart-total">
          <span class="total-label">{{ $t('common.total') }}</span>
          <span class="total-price">${{ cartStore.totalPrice.toFixed(2) }}</span>
        </div>
      </div>
      <button class="checkout-btn" @click="$router.push('/merchant/cart')">{{ $t('product.checkout') }}</button>
    </div>

    <!-- 商品详情弹窗 -->
    <el-dialog
      v-model="detailVisible"
      :title="currentProduct?.name"
      width="520px"
      top="10vh"
      destroy-on-close
    >
      <div v-if="currentProduct" class="detail-body">
        <div class="detail-image">
          <div v-if="getProductImages(currentProduct).length > 1" class="detail-carousel">
            <div class="carousel-container" ref="carouselRef">
              <div class="carousel-track" :style="{ transform: `translateX(-${carouselIndex * 100}%)` }">
                <div v-for="(img, idx) in getProductImages(currentProduct)" :key="idx" class="carousel-slide">
                  <img :src="img" class="detail-img" />
                </div>
              </div>
            </div>
            <div class="carousel-indicators">
              <span v-for="(_, idx) in getProductImages(currentProduct)" :key="idx" class="indicator-dot" :class="{ active: carouselIndex === idx }" @click="carouselIndex = idx"></span>
            </div>
            <button v-if="carouselIndex > 0" class="carousel-arrow carousel-prev" @click="carouselIndex--">&lt;</button>
            <button v-if="carouselIndex < getProductImages(currentProduct).length - 1" class="carousel-arrow carousel-next" @click="carouselIndex++">&gt;</button>
          </div>
          <template v-else>
            <img
              v-if="getProductImage(currentProduct)"
              :src="getProductImage(currentProduct)"
              class="detail-img"
            />
            <div v-else class="detail-img-empty">
              <el-icon :size="48" color="#ccc"><Picture /></el-icon>
            </div>
          </template>
        </div>

        <div class="detail-info">
          <div class="detail-row">
            <span class="label">{{ $t('product.name') }}</span>
            <span class="value">{{ currentProduct.name }}</span>
          </div>
          <div v-if="currentProduct.name_kh" class="detail-row">
            <span class="label">{{ $t('product.nameKh') }}</span>
            <span class="value">{{ currentProduct.name_kh }}</span>
          </div>
          <div class="detail-row">
            <span class="label">{{ $t('product.price') }}</span>
            <span class="value price">${{ currentProduct.price_usd }}
              <em>≈ {{ formatKHR(usdToKhr(currentProduct.price_usd)) }}</em>
            </span>
          </div>
          <div v-if="currentProduct.retail_price_usd" class="detail-row detail-retail-row">
            <span class="label">{{ $t('product.retailPrice') }}</span>
            <span class="value retail-price">${{ currentProduct.retail_price_usd }}
              <em>≈ {{ formatKHR(usdToKhr(currentProduct.retail_price_usd)) }}</em>
            </span>
          </div>
          <div class="detail-row">
            <span class="label">{{ $t('product.stock') }}</span>
            <span class="value">{{ currentProduct.stock }} {{ currentProduct.unit }}</span>
          </div>
          <div v-if="currentProduct.specs" class="detail-row">
            <span class="label">{{ $t('product.specs') }}</span>
            <span class="value">{{ currentProduct.specs }}</span>
          </div>
          <div class="detail-row">
            <span class="label">{{ $t('product.category') }}</span>
            <span class="value">{{ currentProduct.category || $t('product.uncategorized') }}</span>
          </div>
          <div v-if="currentProduct.description" class="detail-row">
            <span class="label">{{ $t('product.description') }}</span>
            <span class="value">{{ currentProduct.description }}</span>
          </div>
        </div>

        <div class="detail-action">
          <span class="qty-label">{{ $t('product.buyQty') }}</span>
          <el-input-number
            v-model="detailQty"
            :min="1"
            :max="currentProduct.stock"
          />
        </div>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button
          type="primary"
          :disabled="currentProduct?.stock <= 0"
          @click="addFromDetail"
        >
          {{ $t('product.addToCartFull') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus/es/components/message/index'
import { Search, ShoppingCart, Picture, Bell } from '@element-plus/icons-vue'
import { getProducts, getCategories, getPublicAnnouncements } from '@/api'
import { useCartStore } from '@/stores/cart'
import { formatKHR, usdToKhr } from '@/utils/format'
import { getCurrentLanguage } from '@/i18n'

const { t } = useI18n()
const cartStore = useCartStore()

const loading = ref(true)
const products = ref([])
const searchKeyword = ref('')
const activeCat = ref('all')
const allCategories = ref([{ id: 'all', name: t('product.all') }])
const notices = ref([])
const currentLang = ref(getCurrentLanguage())

// 详情弹窗
const detailVisible = ref(false)
const currentProduct = ref(null)
const detailQty = ref(1)
const carouselIndex = ref(0)

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

// 加载分类
const loadCategories = async () => {
  try {
    const data = await getCategories()
    allCategories.value = [
      { id: 'all', name: t('product.all') },
      ...data.map(c => ({ id: c.name, name: c.name })),
    ]
  } catch {
    // 回退：从商品中提取
  }
}

// 过滤后的商品列表
const filteredProducts = computed(() => {
  let list = products.value
  if (activeCat.value !== 'all') {
    list = list.filter(p => p.category === activeCat.value)
  }
  if (searchKeyword.value.trim()) {
    const kw = searchKeyword.value.toLowerCase()
    list = list.filter(p =>
      p.name.toLowerCase().includes(kw) ||
      (p.name_kh && p.name_kh.includes(kw))
    )
  }
  // 推荐商品置顶
  return [...list].sort((a, b) => {
    if (a.is_featured && !b.is_featured) return -1
    if (!a.is_featured && b.is_featured) return 1
    return (a.sort_order || 0) - (b.sort_order || 0)
  })
})

// 获取商品首图
const getProductImage = (product) => {
  return product.img1 || product.image_url || ''
}

// 购物车数量
const getCartQty = (id) => {
  const item = cartStore.items.find(i => i.id === id)
  return item ? item.quantity : 0
}

const incrementCart = (product) => {
  if (product.stock <= 0) return
  cartStore.addItem(product, 1)
  ElMessage.success({ message: t('product.addedToCart'), duration: 1000, grouping: true })
}

const decrementCart = (product) => {
  const current = getCartQty(product.id)
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

// 商品详情
const showDetail = (product) => {
  currentProduct.value = product
  detailQty.value = 1
  carouselIndex.value = 0
  detailVisible.value = true
}

const addFromDetail = () => {
  if (!currentProduct.value) return
  cartStore.addItem(currentProduct.value, detailQty.value)
  ElMessage.success(t('product.addedToCart'))
  detailVisible.value = false
}

// 加载商品列表
const loadProducts = async () => {
  loading.value = true
  try {
    const data = await getProducts({ is_active: true })
    products.value = data
  } catch (error) {
    console.error('加载商品失败:', error)
  } finally {
    loading.value = false
  }
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

onMounted(async () => {
  await Promise.all([loadCategories(), loadProducts(), loadNotices()])
  // 从商品中补充分类
  if (allCategories.value.length <= 1) {
    const cats = new Set()
    products.value.forEach(p => { if (p.category) cats.add(p.category) })
    cats.forEach(c => allCategories.value.push({ id: c, name: c }))
  }
})
</script>

<style scoped>
.products-page {
  min-height: 100%;
  background: #f7f7f7;
  padding-bottom: 80px;
}

/* 搜索栏 */
.search-bar {
  padding: 16px 20px 8px;
  background: #fff;
}

.search-input :deep(.el-input__wrapper) {
  border-radius: 8px;
  background: #f5f5f5;
  box-shadow: none !important;
}

/* 公告栏 */
.notice-bar {
  display: flex;
  align-items: center;
  background: #fffbe6;
  margin: 0 20px;
  padding: 8px 14px;
  border-radius: 6px;
  border: 1px solid #fff1b8;
  overflow: hidden;
}

.notice-icon {
  color: #fa8c16;
  font-size: 18px;
  flex-shrink: 0;
  margin-right: 10px;
}

.notice-scroll {
  flex: 1;
  overflow: hidden;
  white-space: nowrap;
}

.notice-track {
  display: inline-block;
  white-space: nowrap;
  animation: noticeScroll linear 1;
  animation-fill-mode: forwards;
}

.notice-item {
  font-size: 13px;
  color: #ad6800;
}

@keyframes noticeScroll {
  0% { transform: translateX(100%); }
  100% { transform: translateX(-100%); }
}

/* 分类导航 */
.category-tabs {
  display: flex;
  overflow-x: auto;
  background: #fff;
  padding: 0 12px;
  border-bottom: 1px solid #f0f0f0;
  scrollbar-width: none;
}

.category-tabs::-webkit-scrollbar {
  display: none;
}

.tab-item {
  flex-shrink: 0;
  padding: 12px 18px;
  font-size: 14px;
  color: #666;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition: all 0.25s;
  white-space: nowrap;
}

.tab-item.active {
  color: #409eff;
  font-weight: 600;
  border-bottom-color: #409eff;
}

.tab-item:hover {
  color: #409eff;
}

/* 商品网格 */
.product-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
  padding: 16px;
}

.product-card {
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
  transition: all 0.25s;
  cursor: pointer;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
}

.product-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.1);
}

/* 正方形图片 */
.card-image {
  position: relative;
  width: 100%;
  padding-bottom: 100%;
  background: #fafafa;
  overflow: hidden;
}

.card-image img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.image-placeholder {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f5f5;
}

.badge {
  position: absolute;
  top: 8px;
  right: 8px;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  color: #fff;
}

.badge-out {
  background: #f56c6c;
}

.badge-low {
  background: #e6a23c;
}

.badge-featured {
  top: 8px;
  left: 8px;
  right: auto;
  background: #f5222d;
}

/* 商品名称 */
.card-name {
  padding: 10px 12px 0;
  font-size: 14px;
  font-weight: 600;
  color: #1a1a1a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 价格 */
.card-price {
  padding: 4px 12px 0;
}

.price-val {
  font-size: 18px;
  font-weight: 700;
  color: #f5222d;
}

.price-unit {
  font-size: 12px;
  color: #999;
}

/* 建议零售价 */
.card-retail {
  padding: 2px 12px 0;
  font-size: 12px;
}

.retail-label {
  color: #b8860b;
}

.retail-val {
  color: #d4a017;
  font-weight: 600;
}

.detail-retail-row .value.retail-price {
  font-size: 18px;
  font-weight: 700;
  color: #d4a017;
}

.detail-retail-row .value.retail-price em {
  font-style: normal;
  font-size: 13px;
  font-weight: 400;
  color: #999;
  margin-left: 6px;
}

/* 加购操作 */
.card-action {
  padding: 10px 12px 14px;
  display: flex;
  justify-content: flex-end;
}

.qty-control {
  display: flex;
  align-items: center;
  gap: 0;
  background: #f5f5f5;
  border-radius: 20px;
  overflow: hidden;
}

.qty-btn {
  width: 30px;
  height: 30px;
  border: none;
  background: #409eff;
  color: #fff;
  font-size: 16px;
  cursor: pointer;
  transition: background 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.qty-btn:hover {
  background: #66b1ff;
}

.qty-num {
  width: 32px;
  text-align: center;
  font-size: 14px;
  font-weight: 600;
  color: #1a1a1a;
  cursor: pointer;
}

.qty-input {
  width: 42px;
  text-align: center;
  font-size: 14px;
  font-weight: 600;
  border: 1px solid #409eff;
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

.add-btn {
  padding: 6px 18px;
  border: none;
  background: #409eff;
  color: #fff;
  border-radius: 20px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.add-btn:hover {
  background: #66b1ff;
}

.add-btn:disabled {
  background: #c0c4cc;
  cursor: not-allowed;
}

/* 底部结算栏 */
.checkout-bar {
  position: fixed;
  bottom: 0;
  left: 200px;
  right: 0;
  height: 60px;
  background: #fff;
  border-top: 1px solid #eee;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  z-index: 100;
  box-shadow: 0 -2px 12px rgba(0, 0, 0, 0.06);
}

.cart-info {
  display: flex;
  align-items: center;
  gap: 14px;
  cursor: pointer;
}

.cart-icon-wrap {
  position: relative;
  width: 40px;
  height: 40px;
  background: #409eff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}

.cart-badge {
  position: absolute;
  top: -4px;
  right: -4px;
  min-width: 18px;
  height: 18px;
  background: #f5222d;
  color: #fff;
  border-radius: 9px;
  font-size: 11px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 4px;
}

.cart-total {
  display: flex;
  flex-direction: column;
}

.total-label {
  font-size: 12px;
  color: #999;
}

.total-price {
  font-size: 20px;
  font-weight: 700;
  color: #f5222d;
}

.checkout-btn {
  padding: 0 36px;
  height: 40px;
  background: #409eff;
  color: #fff;
  border: none;
  border-radius: 20px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

.checkout-btn:hover {
  background: #66b1ff;
}

/* 骨架屏 */
.skeleton-card {
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
  padding-bottom: 12px;
}

.skeleton-img {
  width: 100%;
  padding-bottom: 100%;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}

.skeleton-line {
  height: 14px;
  margin: 10px 12px 0;
  border-radius: 4px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}

.skeleton-line.w60 { width: 60%; }
.skeleton-line.w40 { width: 40%; }

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

/* 商品详情弹窗 */
.detail-body {
  padding: 0;
}

.detail-image {
  width: 100%;
  height: 280px;
  background: #f5f5f5;
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 20px;
  position: relative;
}

.detail-carousel {
  width: 100%;
  height: 100%;
  position: relative;
}

.carousel-container {
  width: 100%;
  height: 100%;
  overflow: hidden;
}

.carousel-track {
  display: flex;
  height: 100%;
  transition: transform 0.3s ease;
}

.carousel-slide {
  flex-shrink: 0;
  width: 100%;
  height: 100%;
}

.carousel-indicators {
  position: absolute;
  bottom: 10px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: 6px;
}

.indicator-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: rgba(255,255,255,0.5);
  cursor: pointer;
  transition: background 0.2s;
}

.indicator-dot.active {
  background: #409eff;
}

.carousel-arrow {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: none;
  background: rgba(0,0,0,0.3);
  color: #fff;
  font-size: 14px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.carousel-prev { left: 8px; }
.carousel-next { right: 8px; }

.carousel-arrow:hover {
  background: rgba(0,0,0,0.5);
}

.detail-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.detail-img-empty {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.detail-info {
  display: flex;
  flex-direction: column;
  gap: 14px;
  margin-bottom: 20px;
}

.detail-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.detail-row .label {
  flex-shrink: 0;
  width: 80px;
  font-size: 13px;
  color: #999;
}

.detail-row .value {
  font-size: 14px;
  color: #1a1a1a;
}

.detail-row .value.price {
  font-size: 20px;
  font-weight: 700;
  color: #f5222d;
}

.detail-row .value.price em {
  font-style: normal;
  font-size: 13px;
  font-weight: 400;
  color: #999;
  margin-left: 6px;
}

.detail-action {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 0;
  border-top: 1px solid #f0f0f0;
}

.detail-action .qty-label {
  font-size: 14px;
  color: #666;
}

/* === 移动端适配 === */
@media (max-width: 767px) {
  .products-page {
    padding-bottom: 120px;
  }

  .search-bar {
    padding: 10px 12px 6px;
  }

  .notice-bar {
    margin: 0 12px;
  }

  .product-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
    padding: 10px;
  }

  .card-name {
    font-size: 13px;
    padding: 8px 8px 0;
  }

  .card-price {
    padding: 2px 8px 0;
  }

  .price-val {
    font-size: 15px;
  }

  .card-retail {
    padding: 1px 8px 0;
    font-size: 11px;
  }

  .card-action {
    padding: 6px 8px 10px;
  }

  .checkout-bar {
    left: 0 !important;
    bottom: 50px;
  }

  .el-dialog {
    width: 92vw !important;
  }
}
</style>
