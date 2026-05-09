<template>
  <div class="page-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <h2>{{ $t('admin.dashboard') }}</h2>
      <div class="header-actions">
        <el-button @click="loadStats" :icon="Refresh" :size="mobile ? 'small' : 'default'">{{ $t('admin.refreshData') }}</el-button>
      </div>
    </div>
    
    <!-- 三色库存预警横条 -->
    <div v-if="stats.lowStockProducts > 0" class="alert-bar alert-danger" @click="stockFilter = 'low'">
      <el-icon class="alert-icon"><warning /></el-icon>
      <span class="alert-text">
        {{ $t('admin.lowStockAlert', { count: stats.lowStockProducts }) }}
      </span>
    </div>
    
    <!-- 核心数据卡片 -->
    <el-row :gutter="mobile ? 10 : 24" class="mb-lg">
      <el-col :xs="12" :sm="8" :md="4">
        <div class="stat-card clickable" :class="{ active: stockFilter === 'low' }" @click="stockFilter = stockFilter === 'low' ? '' : 'low'">
          <div class="stat-icon" style="background: #DC2626">
            <el-icon :size="mobile ? 18 : 24"><warning /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value" style="color: #DC2626">{{ stats.lowStockProducts }}</div>
            <div class="stat-label">{{ $t('admin.stockUrgent') }}</div>
          </div>
        </div>
      </el-col>
      
      <el-col :xs="12" :sm="8" :md="4">
        <div class="stat-card clickable" :class="{ active: stockFilter === 'slow' }" @click="stockFilter = stockFilter === 'slow' ? '' : 'slow'">
          <div class="stat-icon" style="background: #EA580C">
            <el-icon :size="mobile ? 18 : 24"><clock /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value" style="color: #EA580C">{{ stats.slowMoving }}</div>
            <div class="stat-label">{{ $t('admin.slowMoving') }}</div>
          </div>
        </div>
      </el-col>
      
      <el-col :xs="12" :sm="8" :md="4">
        <div class="stat-card">
          <div class="stat-icon" style="background: #16A34A">
            <el-icon :size="mobile ? 18 : 24"><goods /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.totalProducts }}</div>
            <div class="stat-label">{{ $t('admin.totalProducts') }}</div>
          </div>
        </div>
      </el-col>
      
      <el-col :xs="12" :sm="8" :md="4">
        <div class="stat-card">
          <div class="stat-icon" style="background: #0891B2">
            <el-icon :size="mobile ? 18 : 24"><list /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.pendingOrders }}</div>
            <div class="stat-label">{{ $t('admin.pendingOrders') }}</div>
          </div>
        </div>
      </el-col>

      <el-col :xs="12" :sm="8" :md="4">
        <div class="stat-card">
          <div class="stat-icon" style="background: #0F766E">
            <el-icon :size="mobile ? 18 : 24"><user /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.activeUsers }}</div>
            <div class="stat-label">{{ $t('admin.activeUsers') }}</div>
          </div>
        </div>
      </el-col>

      <el-col :xs="12" :sm="8" :md="4">
        <div class="stat-card">
          <div class="stat-icon" style="background: #2563EB">
            <el-icon :size="mobile ? 18 : 24"><goods /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.pageViews }}</div>
            <div class="stat-label">{{ $t('admin.pageViews') }}</div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 7 日 PV/AU 趋势 -->
    <div class="content-card mb-lg">
      <div class="table-header">
        <span class="table-title">{{ $t('admin.pvAuTrend') }}</span>
        <span class="table-subtitle">{{ $t('admin.last7Days') }}</span>
      </div>
      <div v-if="mobile" class="dash-list">
        <div class="dash-item dash-header">
          <span class="dash-item-name" style="font-weight:600">{{ $t('common.date') }}</span>
          <span class="dash-item-val" style="font-weight:600">PV</span>
          <span class="dash-item-val" style="font-weight:600">AU</span>
        </div>
        <div v-for="m in metricsData" :key="m.date" class="dash-item">
          <span class="dash-item-name">{{ m.date.slice(5) }}</span>
          <span class="dash-item-val">{{ m.page_views }}</span>
          <span class="dash-item-val" style="color:#0891B2">{{ m.active_users }}</span>
        </div>
        <div v-if="metricsData.length === 0" class="dash-empty">暂无数据</div>
      </div>
      <el-table v-else :data="metricsData" style="width:100%" size="small">
        <el-table-column :label="$t('common.date')" prop="date" width="130" />
        <el-table-column :label="$t('admin.pageViews')" prop="page_views" align="right" width="100" />
        <el-table-column :label="$t('admin.dailyActiveUsers')" prop="active_users" align="right" width="100">
          <template #default="{ row }">
            <span style="color:#0891B2;font-weight:600">{{ row.active_users }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="$t('admin.pvBar')" min-width="160">
          <template #default="{ row }">
            <div class="mini-bar-wrap">
              <div class="mini-bar pv-bar" :style="{ width: maxPV ? (row.page_views / maxPV * 100) + '%' : '0' }"></div>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-row :gutter="16" style="align-items: stretch;">
      <el-col :xs="24" :sm="12" style="display: flex; flex-direction: column;">
        <div class="content-card" style="flex: 1;">
          <div class="data-table">
            <div class="table-header">
              <span class="table-title">
                {{ stockFilter === 'low' ? $t('admin.lowStockTitle') : stockFilter === 'slow' ? $t('admin.slowMovingTitle') : $t('admin.lowStockTitle') }}
              </span>
              <el-button text type="primary" size="small" @click="$router.push('/admin/products')">
                {{ $t('common.viewAll') }} →
              </el-button>
            </div>
            <!-- 移动端: 简洁列表 -->
            <div v-if="mobile" class="dash-list">
              <div v-for="row in displayProducts" :key="row.id" class="dash-item">
                <div class="dash-item-left">
                  <span class="stock-dot" :class="getStockColor(row)"></span>
                  <span class="dash-item-name">{{ row.name }}</span>
                </div>
                <span class="dash-item-val" :class="getStockColor(row)">{{ row.stock }}<span class="dash-item-warn"> / {{ row.stock_warning }}</span></span>
              </div>
              <div v-if="displayProducts.length === 0" class="dash-empty">—</div>
            </div>
            <!-- 桌面端: 表格 -->
            <el-table v-else :data="displayProducts" row-key="id" style="width: 100%" size="small" :header-cell-style="{ whiteSpace: 'nowrap' }">
              <el-table-column :label="$t('product.name')" prop="name" min-width="150">
                <template #default="{ row }">
                  <div style="display: flex; align-items: center; gap: 8px;">
                    <span class="stock-dot" :class="getStockColor(row)"></span>
                    <span class="text-primary font-medium">{{ row.name }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="库存" width="100" align="right">
                <template #default="{ row }">
                  <span :class="getStockColor(row)" class="font-semibold">{{ row.stock }}</span>
                </template>
              </el-table-column>
              <el-table-column width="100" align="right">
                <template #header><span style="white-space: nowrap;">预警</span></template>
                <template #default="{ row }">{{ row.stock_warning }}</template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </el-col>
      
      <el-col :xs="24" :sm="12" style="display: flex; flex-direction: column;">
        <div class="content-card" style="flex: 1;">
          <div class="data-table">
            <div class="table-header">
              <span class="table-title">{{ $t('admin.pendingOrdersTitle') }}</span>
              <el-button text type="primary" size="small" @click="$router.push('/admin/orders')">
                {{ $t('common.viewAll') }} →
              </el-button>
            </div>
            <!-- 移动端: 简洁列表 -->
            <div v-if="mobile" class="dash-list">
              <div v-for="row in pendingOrders" :key="row.id" class="dash-item">
                <span class="dash-item-name">{{ row.merchant_name }}</span>
                <span class="dash-item-val amount">{{ formatUSD(row.total_usd) }}</span>
              </div>
              <div v-if="pendingOrders.length === 0" class="dash-empty">—</div>
            </div>
            <!-- 桌面端: 表格 -->
            <el-table v-else :data="pendingOrders" row-key="id" style="width: 100%" size="small">
              <el-table-column :label="$t('order.orderNo')" prop="order_no" width="140">
                <template #default="{ row }">
                  <span class="text-primary font-medium">{{ row.order_no }}</span>
                </template>
              </el-table-column>
              <el-table-column :label="$t('order.merchant')" prop="merchant_name" min-width="120" />
              <el-table-column :label="$t('order.amount')" width="120" align="right">
                <template #default="{ row }">
                  <span class="amount">{{ formatUSD(row.total_usd) }}</span>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { Goods, Warning, List, User, Refresh, Clock } from '@element-plus/icons-vue'
import { getProducts, getOrders, getUserList, getDashboardMetrics } from '@/api'
import { formatUSD } from '@/utils/format'

// 移动端检测
const mobile = ref(window.innerWidth < 768)
const onResize = () => { mobile.value = window.innerWidth < 768 }
onMounted(() => window.addEventListener('resize', onResize))
onBeforeUnmount(() => window.removeEventListener('resize', onResize))

const stockFilter = ref('')

const stats = ref({
  totalProducts: 0,
  lowStockProducts: 0,
  slowMoving: 0,
  pendingOrders: 0,
  totalMerchants: 0,
  activeUsers: 0,
  pageViews: 0,
})

const allProducts = ref([])
const lowStockProducts = ref([])
const slowMovingProducts = ref([])
const pendingOrders = ref([])
const metricsData = ref([])

const maxPV = computed(() => {
  if (!metricsData.value.length) return 1
  return Math.max(...metricsData.value.map(m => m.page_views || 0), 1)
})

// 根据选中的筛选展示不同的商品列表（默认显示预警，无"全部"选项）
const displayProducts = computed(() => {
  if (stockFilter.value === 'slow') return slowMovingProducts.value
  return lowStockProducts.value
})

// 商品颜色分级
const getStockColor = (product) => {
  if (product.stock <= product.stock_warning) return 'text-danger'
  return 'text-normal'
}

const loadStats = async () => {
  try {
    const products = await getProducts()
    allProducts.value = products
    stats.value.totalProducts = products.length
    lowStockProducts.value = products.filter((p) => p.is_low_stock).slice(0, 8)
    stats.value.lowStockProducts = products.filter((p) => p.is_low_stock).length

    // 慢销商品: 库存 > 预警值但库存 > 0 且排序较高（简单判断: 有库存但不是低库存）
    // 由于没有 last_sold_at 字段，用 updated_at 近似判断
    const now = new Date()
    const thirtyDaysAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)
    const slowItems = products.filter(p => {
      const updatedAt = new Date(p.updated_at)
      return p.stock > p.stock_warning && p.is_active && updatedAt < thirtyDaysAgo
    })
    slowMovingProducts.value = slowItems.slice(0, 8)
    stats.value.slowMoving = slowItems.length
    
    const orders = await getOrders({ delivery_status: 'pending' })
    stats.value.pendingOrders = orders.length
    pendingOrders.value = orders.slice(0, 5)
    
    const merchants = await getUserList('merchant')
    stats.value.totalMerchants = merchants.length

    const metric = await getDashboardMetrics(7)
    const metrics = metric.metrics || []
    metricsData.value = [...metrics].reverse() // newest last for table display
    const totalAU = metrics.reduce((sum, m) => sum + (m.active_users || 0), 0)
    const totalPV = metrics.reduce((sum, m) => sum + (m.page_views || 0), 0)
    stats.value.activeUsers = totalAU
    stats.value.pageViews = totalPV
  } catch (error) {
    console.error('加载统计数据失败:', error)
  }
}

onMounted(() => {
  loadStats()
})
</script>

<style scoped>
/* ========== 移动端仪表盘列表 ========== */
.dash-list {
  display: flex;
  flex-direction: column;
}

.dash-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid #f0f0f0;
  font-size: 13px;
}

.dash-item:last-child {
  border-bottom: none;
}

.dash-item-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
}

.dash-item-name {
  color: #303133;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-right: 12px;
}

.dash-item-val {
  font-weight: 600;
  white-space: nowrap;
  min-width: 44px;
  text-align: right;
}

.dash-item-warn {
  font-weight: 400;
  color: #909399;
  font-size: 12px;
}

.dash-empty {
  padding: 16px 0;
  text-align: center;
  color: #c0c4cc;
}

.dash-header {
  border-bottom: 2px solid #e6e8eb;
  padding-bottom: 6px;
}

.table-subtitle {
  font-size: 12px;
  color: #909399;
  margin-left: 8px;
}

.mini-bar-wrap {
  height: 12px;
  background: #f0f0f0;
  border-radius: 6px;
  overflow: hidden;
}

.mini-bar {
  height: 100%;
  border-radius: 6px;
  transition: width 0.3s;
}

.pv-bar {
  background: linear-gradient(90deg, #3B82F6, #6366F1);
}

/* 三色圆点 */
.stock-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.stock-dot.text-danger {
  background: #DC2626;
}

.stock-dot.text-warning {
  background: #EA580C;
}

.stock-dot.text-normal {
  background: #16A34A;
}

.text-danger {
  color: #DC2626;
}

.text-warning {
  color: #EA580C;
}

.text-normal {
  color: #16A34A;
}

.amount {
  color: #1D4ED8;
}

.delivery-save-col {
  display: flex;
  align-items: end;
}

/* 可点击的统计卡片 */
.stat-card.clickable {
  cursor: pointer;
  transition: all 0.25s;
  border: 2px solid transparent;
}

.stat-card.clickable:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.12);
}

.stat-card.clickable.active {
  border-color: #409eff;
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.2);
}

/* ========== 移动端适配 ========== */
@media (max-width: 767px) {
  .page-container {
    padding: 12px !important;
  }

  .page-header {
    flex-direction: row;
    align-items: center;
    gap: 0;
  }

  .page-header h2 {
    font-size: 18px;
  }

  .stat-card {
    margin-bottom: 10px;
    padding: 12px !important;
  }

  .stat-icon {
    width: 36px !important;
    height: 36px !important;
    min-width: 36px;
  }

  .stat-value {
    font-size: 20px !important;
  }

  .stat-label {
    font-size: 11px !important;
  }

  .mb-lg {
    margin-bottom: 10px !important;
  }

  .content-card {
    margin-bottom: 12px;
    padding: 12px !important;
    border-radius: 10px !important;
  }

  .table-header {
    margin-bottom: 4px;
  }

  .table-title {
    font-size: 14px !important;
  }

  .alert-bar {
    font-size: 13px;
    padding: 8px 12px !important;
    margin-bottom: 10px;
    border-radius: 8px !important;
    cursor: pointer;
  }
}
</style>