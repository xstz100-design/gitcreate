<template>
  <div class="mobile-shop">
    <!-- 搜索栏（固定顶部） -->
    <van-sticky>
      <div class="search-bar">
        <van-search
          v-model="searchKeyword"
          placeholder="搜索商品"
          @search="handleSearch"
          @clear="handleSearch"
        />
      </div>
    </van-sticky>
    
    <!-- 左右分栏布局 -->
    <div class="shop-content">
      <!-- 左侧分类栏（窄） -->
      <div class="category-sidebar">
        <div
          v-for="category in categories"
          :key="category.id"
          class="category-item"
          :class="{ active: currentCategory === category.id }"
          @click="switchCategory(category.id)"
        >
          {{ category.name }}
        </div>
      </div>
      
      <!-- 右侧商品列表（宽） -->
      <div class="product-list" ref="productListRef">
        <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
          <van-list
            v-model:loading="loading"
            :finished="finished"
            finished-text="没有更多了"
            @load="onLoad"
          >
            <div
              v-for="product in filteredProducts"
              :key="product.id"
              class="product-card"
              @click="showProductDetail(product)"
            >
              <!-- 商品图片 -->
              <div class="product-image">
                <img 
                  v-if="product.image"
                  :src="product.image" 
                  :alt="product.name"
                />
                <div v-else class="image-placeholder">
                  <van-icon name="photo" size="32" />
                </div>
              </div>
              
              <!-- 商品信息 -->
              <div class="product-info">
                <div class="product-name">{{ product.name }}</div>
                <div class="product-name-kh">{{ product.name_kh }}</div>
                
                <!-- 价格行 -->
                <div class="product-price-row">
                  <div class="product-price">${{ product.price_usd }}</div>
                  <div 
                    class="stock-badge"
                    :class="{
                      'stock-low': product.is_low_stock,
                      'stock-out': product.stock <= 0
                    }"
                  >
                    {{ product.stock <= 0 ? '缺货' : product.is_low_stock ? '库存紧张' : `库存 ${product.stock}` }}
                  </div>
                </div>
                
                <!-- 快速加购（底部大拇指区域） -->
                <div class="product-actions">
                  <van-stepper
                    v-model="productQuantities[product.id]"
                    :min="0"
                    :max="product.stock"
                    :disabled="product.stock <= 0"
                    button-size="32px"
                    @change="() => handleQuantityChange(product)"
                  />
                  <van-button
                    type="primary"
                    size="small"
                    :disabled="product.stock <= 0 || !productQuantities[product.id]"
                    @click.stop="addToCart(product)"
                  >
                    加入订单
                  </van-button>
                </div>
              </div>
            </div>
          </van-list>
        </van-pull-refresh>
      </div>
    </div>
    
    <!-- 底部浮动购物车按钮 -->
    <van-goods-action v-if="cartCount > 0">
      <van-goods-action-icon
        icon="cart-o"
        :text="`订单(${cartCount})`"
        :badge="cartCount"
        @click="$router.push('/m/cart')"
      />
      <van-goods-action-button
        type="primary"
        text="去结算"
        @click="$router.push('/m/cart')"
      />
    </van-goods-action>
    
    <!-- 商品详情弹窗 -->
    <van-popup
      v-model:show="showDetail"
      position="bottom"
      round
      :style="{ height: '70%' }"
    >
      <div v-if="selectedProduct" class="product-detail">
        <div class="detail-header">
          <h3>{{ selectedProduct.name }}</h3>
          <van-icon name="cross" @click="showDetail = false" />
        </div>
        <div class="detail-content">
          <div class="detail-row">
            <span class="label">高棉语名称</span>
            <span class="value">{{ selectedProduct.name_kh }}</span>
          </div>
          <div class="detail-row">
            <span class="label">单价</span>
            <span class="value price">${{ selectedProduct.price_usd }}</span>
          </div>
          <div class="detail-row">
            <span class="label">库存</span>
            <span class="value">{{ selectedProduct.stock }}</span>
          </div>
          <div class="detail-row">
            <span class="label">最小订货量</span>
            <span class="value">{{ selectedProduct.min_order_qty }}</span>
          </div>
          <div v-if="selectedProduct.description" class="detail-row full">
            <span class="label">商品描述</span>
            <span class="value">{{ selectedProduct.description }}</span>
          </div>
        </div>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { showToast, showSuccessToast } from 'vant'
import { getProducts } from '@/api'
import { useCartStore } from '@/stores/cart'
import { vibrate } from '@/utils/device'

const cartStore = useCartStore()

// 分类数据
const categories = ref([
  { id: 'all', name: '全部' },
  { id: 'beverage', name: '饮料' },
  { id: 'food', name: '食品' },
  { id: 'daily', name: '日用' },
  { id: 'other', name: '其他' },
])

const currentCategory = ref('all')
const searchKeyword = ref('')

// 商品数据
const products = ref([])
const productQuantities = reactive({})

// 列表状态
const loading = ref(false)
const refreshing = ref(false)
const finished = ref(false)
const productListRef = ref(null)

// 商品详情
const showDetail = ref(false)
const selectedProduct = ref(null)

// 计算购物车数量
const cartCount = computed(() => cartStore.items.length)

// 过滤商品
const filteredProducts = computed(() => {
  let result = products.value
  
  // 分类筛选
  if (currentCategory.value !== 'all') {
    result = result.filter(p => p.category === currentCategory.value)
  }
  
  // 搜索筛选
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(p =>
      p.name.toLowerCase().includes(keyword) ||
      p.name_kh?.toLowerCase().includes(keyword)
    )
  }
  
  return result
})

// 加载商品
const loadProducts = async () => {
  try {
    loading.value = true
    const data = await getProducts()
    products.value = data
    
    // 初始化数量
    data.forEach(p => {
      if (!productQuantities[p.id]) {
        productQuantities[p.id] = 0
      }
    })
    
    finished.value = true
  } catch (error) {
    showToast('加载失败')
    console.error('加载商品失败:', error)
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

// 下拉刷新
const onRefresh = () => {
  finished.value = false
  loadProducts()
}

// 加载更多
const onLoad = () => {
  if (!products.value.length) {
    loadProducts()
  } else {
    finished.value = true
  }
}

// 切换分类
const switchCategory = (id) => {
  currentCategory.value = id
  vibrate('light')
  
  // 滚动到顶部
  if (productListRef.value) {
    productListRef.value.scrollTop = 0
  }
}

// 搜索
const handleSearch = () => {
  vibrate('light')
}

// 数量变化
const handleQuantityChange = (product) => {
  vibrate('light')
}

// 加入购物车
const addToCart = (product) => {
  const quantity = productQuantities[product.id]
  if (!quantity) {
    showToast('请输入数量')
    return
  }
  
  if (quantity > product.stock) {
    showToast('库存不足')
    return
  }
  
  if (quantity < product.min_order_qty) {
    showToast(`最小订货量 ${product.min_order_qty}`)
    return
  }
  
  cartStore.addItem({
    ...product,
    quantity,
  })
  
  vibrate('medium')
  showSuccessToast('已加入订单')
  productQuantities[product.id] = 0
}

// 显示商品详情
const showProductDetail = (product) => {
  selectedProduct.value = product
  showDetail.value = true
  vibrate('light')
}

onMounted(() => {
  loadProducts()
})
</script>

<style scoped lang="scss">
.mobile-shop {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: #F8F9FA;
}

.search-bar {
  background-color: #FFFFFF;
  border-bottom: 1px solid #E8E8E8;
  
  :deep(.van-search) {
    padding: 8px 12px;
  }
  
  :deep(.van-search__content) {
    background-color: #F5F6F7;
    border-radius: 2px;
    border: none;
  }
  
  :deep(.van-field__control) {
    color: #1A1A1A;
    font-size: 14px;
    
    &::placeholder {
      color: #BFBFBF;
    }
  }
}

.shop-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

// 左侧分类栏（窄，80px）
.category-sidebar {
  width: 80px;
  background-color: #F5F6F7;
  border-right: 1px solid #E8E8E8;
  overflow-y: auto;
  flex-shrink: 0;
  
  .category-item {
    height: 56px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 13px;
    color: #4A4A4A;
    font-weight: 500;
    cursor: pointer;
    transition: all 150ms cubic-bezier(0.4, 0.0, 0.2, 1);
    position: relative;
    
    &::after {
      content: '';
      position: absolute;
      left: 0;
      top: 50%;
      transform: translateY(-50%);
      width: 0;
      height: 24px;
      background-color: #1D4ED8;
      transition: width 200ms cubic-bezier(0.4, 0.0, 0.2, 1);
    }
    
    &.active {
      background-color: #FFFFFF;
      color: #1D4ED8;
      font-weight: 600;
      
      &::after {
        width: 3px;
      }
    }
  }
}

// 右侧商品列表（宽）
.product-list {
  flex: 1;
  overflow-y: auto;
  background-color: #FFFFFF;
  
  :deep(.van-pull-refresh) {
    min-height: 100%;
  }
  
  :deep(.van-list) {
    padding: 8px;
  }
}

.product-card {
  display: flex;
  gap: 12px;
  padding: 12px;
  background-color: #FFFFFF;
  border: 1px solid #E8E8E8;
  border-radius: 2px;
  margin-bottom: 8px;
  transition: border-color 150ms;
  
  &:active {
    border-color: #D9D9D9;
    background-color: #FAFAFA;
  }
  
  .product-image {
    width: 80px;
    height: 80px;
    flex-shrink: 0;
    border-radius: 2px;
    overflow: hidden;
    background-color: #F5F6F7;
    
    img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
    
    .image-placeholder {
      width: 100%;
      height: 100%;
      display: flex;
      align-items: center;
      justify-content: center;
      color: #BFBFBF;
    }
  }
  
  .product-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-width: 0;
  }
  
  .product-name {
    font-size: 15px;
    font-weight: 600;
    color: #1A1A1A;
    line-height: 1.4;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  
  .product-name-kh {
    font-size: 12px;
    color: #8C8C8C;
    line-height: 1.4;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  
  .product-price-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }
  
  .product-price {
    font-size: 18px;
    font-weight: 700;
    color: #EA580C;
    letter-spacing: -0.3px;
  }
  
  .stock-badge {
    font-size: 11px;
    font-weight: 500;
    padding: 2px 6px;
    border-radius: 2px;
    background-color: #F0FDF4;
    color: #16A34A;
    border: 1px solid #16A34A;
    
    &.stock-low {
      background-color: #FFF7ED;
      color: #EA580C;
      border-color: #EA580C;
    }
    
    &.stock-out {
      background-color: #FEF2F2;
      color: #DC2626;
      border-color: #DC2626;
    }
  }
  
  .product-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: auto;
    
    :deep(.van-stepper) {
      flex: 1;
      
      .van-stepper__input {
        color: #1A1A1A;
        font-weight: 600;
        font-size: 14px;
      }
    }
    
    :deep(.van-button--small) {
      height: 32px;
      padding: 0 16px;
      font-size: 13px;
      font-weight: 500;
      border-radius: 2px;
    }
  }
}

// 底部浮动购物车
:deep(.van-goods-action) {
  border-top: 1px solid #E8E8E8;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.04);
  
  .van-goods-action-icon {
    border-right: 1px solid #E8E8E8;
  }
  
  .van-goods-action-button {
    border-radius: 0;
  }
}

// 商品详情弹窗
.product-detail {
  padding: 16px;
  height: 100%;
  display: flex;
  flex-direction: column;
  
  .detail-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-bottom: 16px;
    border-bottom: 1px solid #E8E8E8;
    margin-bottom: 16px;
    
    h3 {
      font-size: 18px;
      font-weight: 600;
      color: #1A1A1A;
      margin: 0;
    }
    
    .van-icon {
      font-size: 20px;
      color: #8C8C8C;
      padding: 4px;
    }
  }
  
  .detail-content {
    flex: 1;
    overflow-y: auto;
  }
  
  .detail-row {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    padding: 12px 0;
    border-bottom: 1px solid #F0F0F0;
    
    &.full {
      flex-direction: column;
      gap: 8px;
    }
    
    .label {
      font-size: 14px;
      color: #8C8C8C;
      font-weight: 500;
    }
    
    .value {
      font-size: 14px;
      color: #1A1A1A;
      font-weight: 500;
      text-align: right;
      
      &.price {
        font-size: 20px;
        font-weight: 700;
        color: #EA580C;
      }
    }
  }
}
</style>
