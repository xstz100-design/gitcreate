<template>
  <div class="picking-page">
    <div class="page-header">
      <h2>{{ $t('picker.title') }}</h2>
    </div>

    <div class="filters">
      <el-radio-group v-model="filterStatus" @change="loadOrders">
        <el-radio-button label="pending">{{ $t('picker.pending') }}</el-radio-button>
        <el-radio-button label="picked">{{ $t('picker.picked') }}</el-radio-button>
        <el-radio-button label="all">{{ $t('picker.all') }}</el-radio-button>
      </el-radio-group>
      <el-button type="primary" plain @click="loadOrders" style="margin-left: 12px;">
        {{ $t('common.refresh') }}
      </el-button>
    </div>

    <el-table v-loading="loading" :data="filteredOrders" border style="width:100%">
      <el-table-column :label="$t('picker.orderNo')" prop="order_no" width="200" />
      <el-table-column :label="$t('order.orderTime')" width="180">
        <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column :label="$t('picker.items')" width="100" align="center">
        <template #default="{ row }">{{ countItems(row) }}</template>
      </el-table-column>
      <el-table-column :label="$t('picker.pickedAt')" width="180">
        <template #default="{ row }">
          <el-tag v-if="row.picked_at" type="success" size="small">
            {{ formatDateTime(row.picked_at) }}
          </el-tag>
          <el-tag v-else type="warning" size="small">{{ $t('picker.pending') }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="$t('common.action')" width="240">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="openItems(row)">
            {{ $t('picker.confirmPick') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <div v-if="!loading && filteredOrders.length === 0" class="empty">
      {{ $t('picker.noOrders') }}
    </div>

    <!-- 配货清单弹窗：仅商品名 + 图 + 数量 + 条码 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="640px"
      align-center
    >
      <div v-loading="dialogLoading">
        <div v-if="currentDetail" class="picker-list">
          <div
            v-for="(it, idx) in currentDetail.items"
            :key="idx"
            class="picker-item"
          >
            <el-image
              v-if="it.image"
              :src="it.image"
              fit="cover"
              class="picker-img"
            />
            <div v-else class="picker-img placeholder" />
            <div class="picker-info">
              <div class="picker-name">{{ it.name }}</div>
              <div v-if="it.barcode" class="picker-barcode">{{ it.barcode }}</div>
              <div v-if="it.specs" class="picker-specs">{{ it.specs }}</div>
            </div>
            <div class="picker-qty">
              × <b>{{ it.quantity }}</b>
              <span v-if="it.unit" class="picker-unit">{{ it.unit }}</span>
            </div>
          </div>
          <div class="picker-total">
            {{ $t('picker.totalItems', { n: currentDetail.items.length, q: totalQty }) }}
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button
          v-if="currentDetail && !currentDetail.picked_at"
          type="primary"
          :loading="confirming"
          @click="confirmPick"
        >
          {{ $t('picker.confirmPick') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { getOrders, getPickerItems, markOrderPicked } from '@/api'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()

const loading = ref(false)
const orders = ref([])
const filterStatus = ref('pending')

const dialogVisible = ref(false)
const dialogLoading = ref(false)
const confirming = ref(false)
const currentDetail = ref(null)
const currentOrderId = ref(null)

const filteredOrders = computed(() => {
  if (filterStatus.value === 'pending') {
    return orders.value.filter(o => !o.picked_at && o.delivery_status !== 'cancelled')
  }
  if (filterStatus.value === 'picked') {
    return orders.value.filter(o => !!o.picked_at)
  }
  return orders.value
})

const dialogTitle = computed(() => {
  if (!currentDetail.value) return t('picker.title')
  return `${t('picker.orderNo')}: ${currentDetail.value.order_no}`
})

const totalQty = computed(() => {
  if (!currentDetail.value) return 0
  return currentDetail.value.items.reduce((s, it) => s + (it.quantity || 0), 0)
})

function countItems(row) {
  return (row.items || []).length
}

async function loadOrders() {
  loading.value = true
  try {
    const data = await getOrders({})
    orders.value = data || []
  } catch (e) {
    ElMessage.error(e?.message || 'Load failed')
  } finally {
    loading.value = false
  }
}

async function openItems(row) {
  currentOrderId.value = row.id
  dialogVisible.value = true
  dialogLoading.value = true
  currentDetail.value = null
  try {
    currentDetail.value = await getPickerItems(row.id)
  } catch (e) {
    ElMessage.error(e?.message || 'Load failed')
    dialogVisible.value = false
  } finally {
    dialogLoading.value = false
  }
}

async function confirmPick() {
  try {
    await ElMessageBox.confirm(t('picker.confirmPickTip'), t('picker.confirmPick'), {
      type: 'warning',
    })
  } catch {
    return
  }
  confirming.value = true
  try {
    await markOrderPicked(currentOrderId.value)
    ElMessage.success(t('picker.pickedSuccess'))
    dialogVisible.value = false
    await loadOrders()
  } catch (e) {
    ElMessage.error(e?.message || 'Failed')
  } finally {
    confirming.value = false
  }
}

onMounted(loadOrders)
</script>

<style scoped lang="scss">
.picking-page {
  padding: 16px;
}
.page-header {
  margin-bottom: 16px;
  h2 {
    margin: 0;
    color: #2b2b2b;
  }
}
.filters {
  margin-bottom: 16px;
  display: flex;
  align-items: center;
}
.empty {
  text-align: center;
  color: #999;
  padding: 40px 0;
}
.picker-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.picker-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px;
  border: 1px solid #eee;
  border-radius: 8px;
}
.picker-img {
  width: 64px;
  height: 64px;
  border-radius: 6px;
  flex-shrink: 0;
  background: #f5f5f5;
  &.placeholder {
    background: #f0f0f0;
  }
}
.picker-info {
  flex: 1;
  min-width: 0;
}
.picker-name {
  font-weight: 600;
  font-size: 16px;
  color: #2b2b2b;
  margin-bottom: 4px;
}
.picker-barcode {
  font-family: monospace;
  font-size: 14px;
  color: #666;
}
.picker-specs {
  font-size: 12px;
  color: #999;
}
.picker-qty {
  font-size: 18px;
  color: #d44e4e;
  white-space: nowrap;
  .picker-unit {
    margin-left: 4px;
    font-size: 12px;
    color: #999;
  }
}
.picker-total {
  text-align: right;
  padding: 8px 4px;
  color: #666;
  border-top: 1px dashed #ddd;
}
</style>
