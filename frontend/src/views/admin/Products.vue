<template>
  <div class="products-page">
    <div class="page-header">
      <h2>{{ $t('admin.products') }}</h2>
      <div class="header-btns">
        <el-button @click="showInventoryPreview" :size="mobile ? 'small' : 'default'">
          <el-icon><document /></el-icon>
          {{ $t('product.inventoryPreview') }}
        </el-button>
        <el-button type="primary" @click="handleAdd" :size="mobile ? 'small' : 'default'">
          <el-icon><plus /></el-icon>
          {{ $t('product.addProduct') }}
        </el-button>
      </div>
    </div>

    <!-- 桌面端: 表格视图 -->
    <el-table
      v-if="!mobile"
      v-loading="loading"
      :data="paginatedProducts"
      border
      row-key="id"
      :default-sort="{ prop: 'sort_order', order: 'ascending' }"
    >
      <el-table-column :label="$t('product.sortOrder')" prop="sort_order" width="70" sortable />
      <el-table-column label="ID" prop="id" width="60" sortable />
      <el-table-column :label="$t('product.image')" width="80">
        <template #default="{ row }">
          <el-image
            v-if="row.img1"
            :src="row.img1"
            :preview-src-list="getImageList(row)"
            fit="cover"
            style="width: 50px; height: 50px; border-radius: 4px;"
          />
          <span v-else style="color: #ccc; font-size: 12px;">{{ $t('product.noImage') }}</span>
        </template>
      </el-table-column>
      <el-table-column :label="$t('product.name')" prop="name" min-width="140" />
      <el-table-column :label="$t('product.nameKh')" prop="name_kh" width="130" />
      <el-table-column :label="$t('product.category')" prop="category" width="100" />
      <el-table-column :label="$t('product.unit')" prop="unit" width="70" />
      <el-table-column :label="$t('product.price')" width="110" sortable prop="price_usd">
        <template #default="{ row }">
          {{ formatUSD(row.price_usd) }}
        </template>
      </el-table-column>
      <el-table-column :label="$t('product.retailPrice')" width="110" sortable prop="retail_price_usd">
        <template #default="{ row }">
          <span v-if="row.retail_price_usd" style="color: #d4a017; font-weight: 600;">
            {{ formatUSD(row.retail_price_usd) }}
          </span>
          <span v-else style="color: #ccc;">-</span>
        </template>
      </el-table-column>
      <el-table-column :label="$t('product.stock')" width="90" sortable prop="stock">
        <template #default="{ row }">
          <el-tag :type="row.is_low_stock ? 'warning' : 'success'" size="small">
            {{ row.stock }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="$t('product.stockWarning')" prop="stock_warning" width="75" />
      <el-table-column :label="$t('product.status')" width="85">
        <template #default="{ row }">
          <el-switch
            :model-value="row.is_active"
            size="small"
            :active-text="$t('product.onSale')"
            :inactive-text="$t('product.offSale')"
            inline-prompt
            @change="(val) => toggleActive(row, val)"
          />
        </template>
      </el-table-column>
      <el-table-column :label="$t('product.featured')" width="75">
        <template #default="{ row }">
          <el-switch
            :model-value="row.is_featured"
            size="small"
            @change="(val) => toggleFeatured(row, val)"
          />
        </template>
      </el-table-column>
      <el-table-column :label="$t('common.operation')" width="140" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link size="small" @click="handleEdit(row)">{{ $t('common.edit') }}</el-button>
          <el-button type="danger" link size="small" @click="handleDelete(row)">{{ $t('common.delete') }}</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 移动端: 卡片列表 -->
    <div v-else v-loading="loading" class="mobile-card-list">
      <div v-for="row in paginatedProducts" :key="row.id" class="product-card" @click="handleEdit(row)">
        <div class="card-left">
          <el-image
            v-if="row.img1"
            :src="row.img1"
            fit="cover"
            class="card-img"
          />
          <div v-else class="card-img card-img-empty">
            <el-icon :size="20"><goods /></el-icon>
          </div>
        </div>
        <div class="card-right">
          <div class="card-title-row">
            <span class="card-name">{{ row.name }}</span>
            <el-tag :type="row.is_active ? 'success' : 'info'" size="small" @click.stop="toggleActive(row, !row.is_active)">
              {{ row.is_active ? $t('product.onSale') : $t('product.offSale') }}
            </el-tag>
          </div>
          <div class="card-meta">
            <span v-if="row.category" class="card-cat">{{ row.category }}</span>
            <span v-if="row.unit" class="card-unit">/ {{ row.unit }}</span>
          </div>
          <div class="card-bottom">
            <span class="card-price">{{ formatUSD(row.price_usd) }}</span>
            <div class="card-stock-control" @click.stop>
              <button class="stock-btn stock-minus" @click="openQuickDelta(row, 'pick')" :disabled="row.stock <= 0">−</button>
              <span class="stock-num" :class="{ 'stock-warn': row.is_low_stock }" @click="openStockDrawer(row)">{{ row.stock }}</span>
              <button class="stock-btn stock-plus" @click="openQuickDelta(row, 'restock')">+</button>
            </div>
          </div>
        </div>
        <div class="card-actions" @click.stop>
          <el-button type="danger" link size="small" @click="handleDelete(row)">
            <el-icon><delete /></el-icon>
          </el-button>
        </div>
      </div>
      <el-empty v-if="!loading && products.length === 0" />
    </div>

    <!-- 库存快捷编辑抽屉 -->
    <el-drawer
      v-model="stockDrawerVisible"
      :title="$t('admin.quickStockEdit')"
      direction="btt"
      size="auto"
      :with-header="true"
    >
      <div v-if="stockDrawerProduct" class="stock-drawer-content">
        <div class="drawer-product-info">
          <span class="drawer-product-name">{{ stockDrawerProduct.name }}</span>
          <el-tag :type="stockDrawerProduct.is_low_stock ? 'warning' : 'success'" size="small">
            {{ $t('product.stock') }}: {{ stockDrawerProduct.stock }}
          </el-tag>
        </div>
        <div class="drawer-form">
          <div class="drawer-form-item">
            <label>{{ $t('product.restockPick') }}</label>
            <div class="delta-input-row">
              <el-radio-group v-model="stockDrawerForm.deltaMode" size="small">
                <el-radio-button value="restock">{{ $t('product.restock') }}</el-radio-button>
                <el-radio-button value="pick">{{ $t('product.pick') }}</el-radio-button>
              </el-radio-group>
              <el-input-number v-model="stockDrawerForm.deltaAmount" :min="0" controls-position="right" style="flex: 1;" />
            </div>
            <div class="delta-preview">
              {{ $t('product.currentStock') }}: {{ stockDrawerProduct.stock }}
              →
              <span :class="{ 'stock-warn': previewStock <= (stockDrawerForm.stock_warning || stockDrawerProduct.stock_warning) }">
                {{ previewStock }}
              </span>
            </div>
          </div>
          <div class="drawer-form-item">
            <label>{{ $t('product.stockWarning') }}</label>
            <el-input-number v-model="stockDrawerForm.stock_warning" :min="0" controls-position="right" style="width: 100%;" />
          </div>
          <div class="drawer-form-item">
            <label>{{ $t('product.price') }}</label>
            <el-input-number v-model="stockDrawerForm.price_usd" :min="0" :step="0.1" :precision="2" controls-position="right" style="width: 100%;" />
          </div>
        </div>
        <el-button type="primary" round :loading="stockDrawerSaving" @click="saveStockDrawer" style="width: 100%; margin-top: 16px;">
          {{ $t('common.save') }}
        </el-button>
      </div>
    </el-drawer>

    <!-- 快捷补货/取货弹窗 -->
    <el-dialog
      v-model="quickDeltaVisible"
      :title="quickDeltaMode === 'restock' ? $t('product.restock') : $t('product.pick')"
      width="320px"
      append-to-body
      destroy-on-close
    >
      <div v-if="quickDeltaProduct" class="quick-delta-content">
        <div class="quick-delta-info">
          <span>{{ quickDeltaProduct.name }}</span>
          <el-tag :type="quickDeltaProduct.is_low_stock ? 'warning' : 'success'" size="small">
            {{ $t('product.stock') }}: {{ quickDeltaProduct.stock }}
          </el-tag>
        </div>
        <div class="quick-delta-input">
          <span class="delta-label">{{ quickDeltaMode === 'restock' ? '+' : '−' }}</span>
          <el-input-number v-model="quickDeltaAmount" :min="1" :max="quickDeltaMode === 'pick' ? quickDeltaProduct.stock : 99999" controls-position="right" style="flex: 1;" />
        </div>
        <div class="quick-delta-preview">
          {{ quickDeltaProduct.stock }} {{ quickDeltaMode === 'restock' ? '+' : '−' }} {{ quickDeltaAmount }}
          = <b>{{ quickDeltaMode === 'restock' ? quickDeltaProduct.stock + quickDeltaAmount : quickDeltaProduct.stock - quickDeltaAmount }}</b>
        </div>
      </div>
      <template #footer>
        <el-button @click="quickDeltaVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="quickDeltaSaving" @click="confirmQuickDelta">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- 库存预览弹窗 -->
    <el-dialog
      v-model="inventoryPreviewVisible"
      :title="$t('product.inventoryPreview')"
      :width="mobile ? '94vw' : '600px'"
      destroy-on-close
    >
      <div class="inventory-preview">
        <div class="inventory-summary">
          <div class="inv-stat inv-total">
            <div class="inv-num">{{ products.length }}</div>
            <div class="inv-label">{{ $t('admin.totalProducts') }}</div>
          </div>
          <div class="inv-stat inv-ok">
            <div class="inv-num">{{ normalStockCount }}</div>
            <div class="inv-label">{{ $t('product.stockNormal') }}</div>
          </div>
          <div class="inv-stat inv-warn">
            <div class="inv-num">{{ lowStockCount }}</div>
            <div class="inv-label">{{ $t('product.lowStock') }}</div>
          </div>
          <div class="inv-stat inv-danger">
            <div class="inv-num">{{ outOfStockCount }}</div>
            <div class="inv-label">{{ $t('product.outOfStock') }}</div>
          </div>
        </div>
        <el-divider />
        <div class="inv-filter-row">
          <el-radio-group v-model="invFilter" size="small">
            <el-radio-button value="all">{{ $t('product.all') }}</el-radio-button>
            <el-radio-button value="low">{{ $t('product.lowStock') }}</el-radio-button>
            <el-radio-button value="out">{{ $t('product.outOfStock') }}</el-radio-button>
          </el-radio-group>
        </div>
        <div class="inv-list">
          <div v-for="p in filteredInventory" :key="p.id" class="inv-item">
            <span class="inv-dot" :class="getInvDotClass(p)"></span>
            <span class="inv-name">{{ p.name }}</span>
            <span class="inv-stock" :class="getInvDotClass(p)">{{ p.stock }} / {{ p.stock_warning }}</span>
          </div>
          <el-empty v-if="filteredInventory.length === 0" :description="$t('common.noData')" />
        </div>
      </div>
    </el-dialog>

    <div class="pagination-wrapper">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="products.length"
        :layout="mobile ? 'prev, pager, next' : 'total, sizes, prev, pager, next, jumper'"
        :small="mobile"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>

    <!-- 编辑/添加对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? $t('product.editProduct') : $t('product.addProduct')"
      :width="mobile ? '94vw' : '680px'"
      :fullscreen="mobile"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        :label-width="mobile ? '80px' : '100px'"
        :label-position="mobile ? 'top' : 'right'"
      >
        <el-row :gutter="mobile ? 0 : 16">
          <el-col :xs="24" :sm="12">
            <el-form-item :label="$t('product.name')" prop="name">
              <el-input v-model="form.name" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12">
            <el-form-item :label="$t('product.nameKh')" prop="name_kh">
              <el-input v-model="form.name_kh" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="mobile ? 0 : 16">
          <el-col :xs="24" :sm="12">
            <el-form-item :label="$t('product.category')" prop="category">
              <el-select v-model="form.category" filterable allow-create default-first-option :placeholder="$t('product.selectCategory')" style="width: 100%;">
                <el-option
                  v-for="cat in categoryOptions"
                  :key="cat"
                  :label="cat"
                  :value="cat"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="12" :sm="6">
            <el-form-item :label="$t('product.unit')" prop="unit">
              <el-input v-model="form.unit" :placeholder="$t('product.unitPlaceholder')" />
            </el-form-item>
          </el-col>
          <el-col :xs="12" :sm="6">
            <el-form-item :label="$t('product.sortOrder')" prop="sort_order">
              <el-input-number v-model="form.sort_order" :min="0" controls-position="right" style="width: 100%;" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="mobile ? 0 : 16">
          <el-col :xs="12" :sm="6">
            <el-form-item :label="$t('product.price')" prop="price_usd">
              <el-input-number v-model="form.price_usd" :min="0" :step="0.1" :precision="2" controls-position="right" style="width: 100%;" />
            </el-form-item>
          </el-col>
          <el-col :xs="12" :sm="6">
            <el-form-item :label="$t('product.retailPrice')" prop="retail_price_usd">
              <el-input-number v-model="form.retail_price_usd" :min="0" :step="0.1" :precision="2" controls-position="right" style="width: 100%;" />
            </el-form-item>
          </el-col>
          <el-col :xs="12" :sm="6">
            <el-form-item :label="$t('product.stock')" prop="stock">
              <el-input-number v-model="form.stock" :min="0" controls-position="right" style="width: 100%;" />
            </el-form-item>
          </el-col>
          <el-col :xs="12" :sm="6">
            <el-form-item :label="$t('product.stockWarning')" prop="stock_warning">
              <el-input-number v-model="form.stock_warning" :min="0" controls-position="right" style="width: 100%;" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item :label="$t('product.description')" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>

        <el-row :gutter="mobile ? 0 : 16">
          <el-col :xs="12" :sm="6">
            <el-form-item :label="$t('product.specs')" prop="specs">
              <el-input v-model="form.specs" :placeholder="$t('product.specsPlaceholder')" />
            </el-form-item>
          </el-col>
          <el-col :xs="12" :sm="6">
            <el-form-item :label="$t('product.minOrderQty')" prop="min_order_qty">
              <el-input-number v-model="form.min_order_qty" :min="1" controls-position="right" style="width: 100%;" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12">
            <el-form-item :label="$t('product.barcode')" prop="barcode">
              <el-input v-model="form.barcode" :placeholder="$t('product.barcodePlaceholder')" clearable>
                <template #prefix>
                  <el-icon><Goods /></el-icon>
                </template>
              </el-input>
            </el-form-item>
          </el-col>
        </el-row>

        <!-- 图片上传区域 -->
        <el-form-item :label="$t('product.images')">
          <div class="image-upload-area" @paste="onPasteImage">
            <div v-for="idx in 5" :key="idx" class="image-slot">
              <div v-if="form[`img${idx}`]" class="image-preview">
                <el-image :src="form[`img${idx}`]" fit="cover" style="width: 100%; height: 100%;" />
                <div class="image-actions">
                  <el-icon @click="removeImage(idx)"><delete /></el-icon>
                </div>
              </div>
              <el-upload
                v-else
                class="image-uploader"
                :show-file-list="false"
                :http-request="(opt) => handleUpload(opt, idx)"
                accept=".jpg,.jpeg,.png,.webp,.gif"
              >
                <div class="upload-placeholder">
                  <el-icon><plus /></el-icon>
                  <span>{{ $t('product.imageN', { n: idx }) }}</span>
                </div>
              </el-upload>
            </div>
          </div>
          <div class="upload-tip">{{ $t('product.uploadTip5') }} | {{ $t('product.pasteTip') }}</div>
        </el-form-item>

        <el-form-item :label="$t('product.status')" prop="is_active">
          <el-switch v-model="form.is_active" :active-text="$t('product.onSale')" :inactive-text="$t('product.offSale')" />
        </el-form-item>

        <el-form-item :label="$t('product.featured')" prop="is_featured">
          <el-switch v-model="form.is_featured" :active-text="$t('product.featured')" :inactive-text="''" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { Plus, Delete, Goods } from '@element-plus/icons-vue'
import { Document } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus/es/components/message/index'
import { ElMessageBox } from 'element-plus/es/components/message-box/index'
import { getProducts, createProduct, updateProduct, deleteProduct, uploadImage, getAllCategories } from '@/api'
import { formatUSD } from '@/utils/format'

const { t } = useI18n()

// 移动端检测
const mobile = ref(window.innerWidth < 768)
const onResize = () => { mobile.value = window.innerWidth < 768 }
onMounted(() => window.addEventListener('resize', onResize))
onBeforeUnmount(() => window.removeEventListener('resize', onResize))

const loading = ref(false)
const products = ref([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref()
const categoryOptions = ref([])

// 分页
const currentPage = ref(1)
const pageSize = ref(20)

const paginatedProducts = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return products.value.slice(start, end)
})

const handleSizeChange = () => { currentPage.value = 1 }
const handleCurrentChange = () => { window.scrollTo({ top: 0, behavior: 'smooth' }) }

const form = reactive({
  id: null,
  name: '',
  name_kh: '',
  category: '',
  unit: '',
  price_usd: null,
  retail_price_usd: null,
  stock: 0,
  stock_warning: 10,
  description: '',
  image_url: '',
  img1: '',
  img2: '',
  img3: '',
  img4: '',
  img5: '',
  specs: '',
  barcode: '',
  sort_order: 0,
  is_active: true,
  is_featured: false,
})

const rules = {
  name: [{ required: true, message: () => t('product.nameRequired'), trigger: 'blur' }],
  price_usd: [
    { required: true, message: () => t('product.priceRequired'), trigger: 'blur' },
    { type: 'number', min: 0.01, message: () => t('product.priceRequired'), trigger: 'blur' },
  ],
}

// 获取图片列表用于预览
const getImageList = (row) => {
  return [row.img1, row.img2, row.img3, row.img4, row.img5].filter(Boolean)
}

// 图片上传
const handleUpload = async (options, idx) => {
  try {
    const res = await uploadImage(options.file)
    form[`img${idx}`] = res.url
    // 自动同步 img1 到 image_url 作为兼容
    if (idx === 1) form.image_url = res.url
    ElMessage.success(`${t('product.image')}${idx} ✓`)
  } catch (error) {
    ElMessage.error(t('product.uploadFailed'))
  }
}

// 移除图片
const removeImage = (idx) => {
  form[`img${idx}`] = ''
  if (idx === 1) form.image_url = ''
}

// 粘贴图片 - 自动填充到第一个空位
const onPasteImage = async (e) => {
  const items = e.clipboardData?.items
  if (!items) return
  for (const item of items) {
    if (item.type.startsWith('image/')) {
      e.preventDefault()
      const file = item.getAsFile()
      if (!file) return
      // 找到第一个空的图片槽
      const emptyIdx = [1, 2, 3, 4, 5].find(i => !form[`img${i}`])
      if (!emptyIdx) {
        ElMessage.warning(t('product.uploadTip'))
        return
      }
      await handleUpload({ file }, emptyIdx)
      break
    }
  }
}

// 加载分类选项
const loadCategories = async () => {
  try {
    const data = await getAllCategories()
    categoryOptions.value = data.map(c => c.name)
  } catch {
    // 如果分类API未就绪，从现有产品中提取
    const cats = [...new Set(products.value.map(p => p.category).filter(Boolean))]
    categoryOptions.value = cats
  }
}

const loadProducts = async () => {
  loading.value = true
  try {
    const data = await getProducts({ is_active: null })
    products.value = data
  } catch (error) {
    console.error('加载商品失败:', error)
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  Object.assign(form, {
    id: null,
    name: '',
    name_kh: '',
    category: '',
    unit: '',
    price_usd: null,
    retail_price_usd: null,
    stock: 0,
    stock_warning: 10,
    description: '',
    image_url: '',
    img1: '',
    img2: '',
    img3: '',
    img4: '',
    img5: '',
    specs: '',
    barcode: '',
    sort_order: 0,
    is_active: true,
    is_featured: false,
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
    ...row,
    retail_price_usd: row.retail_price_usd || null,
    img1: row.img1 || '',
    img2: row.img2 || '',
    img3: row.img3 || '',
    img4: row.img4 || '',
    img5: row.img5 || '',
    specs: row.specs || '',
    barcode: row.barcode || '',
    sort_order: row.sort_order || 0,
    is_featured: row.is_featured || false,
  })
  isEdit.value = true
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      const payload = { ...form }
      delete payload.id
      // 清理空字符串为null，避免后端验证问题
      for (const key of Object.keys(payload)) {
        if (payload[key] === '') payload[key] = null
      }
      if (isEdit.value) {
        await updateProduct(form.id, payload)
        ElMessage.success(t('product.productUpdated'))
      } else {
        delete payload.is_active  // 创建时由后端默认
        await createProduct(payload)
        ElMessage.success(t('product.productAdded'))
      }
      dialogVisible.value = false
      loadProducts()
    } catch (error) {
      console.error('提交失败:', error)
    } finally {
      submitting.value = false
    }
  })
}

const toggleFeatured = async (row, val) => {
  try {
    await updateProduct(row.id, { is_featured: val })
    row.is_featured = val
    ElMessage.success(val ? t('product.featuredOn') : t('product.featuredOff'))
  } catch (error) {
    console.error('切换推荐失败:', error)
  }
}

// ====== 库存快捷操作 ======
const stockDrawerVisible = ref(false)
const stockDrawerProduct = ref(null)
const stockDrawerSaving = ref(false)
const stockDrawerForm = reactive({
  deltaMode: 'restock',
  deltaAmount: 0,
  stock_warning: 10,
  price_usd: 0,
})

// 计算抽屉中的预览库存
const previewStock = computed(() => {
  if (!stockDrawerProduct.value) return 0
  const base = stockDrawerProduct.value.stock
  return stockDrawerForm.deltaMode === 'restock'
    ? base + stockDrawerForm.deltaAmount
    : Math.max(0, base - stockDrawerForm.deltaAmount)
})

// 快捷补货/取货弹窗
const quickDeltaVisible = ref(false)
const quickDeltaProduct = ref(null)
const quickDeltaMode = ref('restock')
const quickDeltaAmount = ref(1)
const quickDeltaSaving = ref(false)

// 库存预览
const inventoryPreviewVisible = ref(false)
const invFilter = ref('all')

const lowStockCount = computed(() => products.value.filter(p => p.stock > 0 && p.is_low_stock).length)
const outOfStockCount = computed(() => products.value.filter(p => p.stock <= 0).length)
const normalStockCount = computed(() => products.value.filter(p => !p.is_low_stock && p.stock > 0).length)

const filteredInventory = computed(() => {
  if (invFilter.value === 'low') return products.value.filter(p => p.stock > 0 && p.is_low_stock)
  if (invFilter.value === 'out') return products.value.filter(p => p.stock <= 0)
  return [...products.value].sort((a, b) => a.stock - b.stock)
})

const getInvDotClass = (p) => {
  if (p.stock <= 0) return 'inv-danger'
  if (p.is_low_stock) return 'inv-warn'
  return 'inv-ok'
}

const showInventoryPreview = () => {
  invFilter.value = 'all'
  inventoryPreviewVisible.value = true
}

// 打开快捷补货/取货弹窗
const openQuickDelta = (row, mode) => {
  quickDeltaProduct.value = row
  quickDeltaMode.value = mode
  quickDeltaAmount.value = 1
  quickDeltaVisible.value = true
}

// 确认快捷补货/取货
const confirmQuickDelta = async () => {
  if (!quickDeltaProduct.value) return
  const row = quickDeltaProduct.value
  const delta = quickDeltaMode.value === 'restock' ? quickDeltaAmount.value : -quickDeltaAmount.value
  const newStock = Math.max(0, row.stock + delta)
  quickDeltaSaving.value = true
  try {
    await updateProduct(row.id, { stock: newStock })
    row.stock = newStock
    row.is_low_stock = newStock <= (row.stock_warning || 0)
    ElMessage.success(t('product.productUpdated'))
    quickDeltaVisible.value = false
  } catch (error) {
    ElMessage.error(t('common.operationFailed'))
  } finally {
    quickDeltaSaving.value = false
  }
}

// 商品上架/下架切换
const toggleActive = async (row, val) => {
  try {
    await updateProduct(row.id, { is_active: val })
    row.is_active = val
    ElMessage.success(val ? t('product.statusOnSale') : t('product.statusOffSale'))
  } catch (error) {
    ElMessage.error(t('common.operationFailed'))
  }
}

// 打开库存抽屉
const openStockDrawer = (row) => {
  stockDrawerProduct.value = row
  Object.assign(stockDrawerForm, {
    deltaMode: 'restock',
    deltaAmount: 0,
    stock_warning: row.stock_warning,
    price_usd: row.price_usd,
  })
  stockDrawerVisible.value = true
}

// 保存抽屉编辑
const saveStockDrawer = async () => {
  if (!stockDrawerProduct.value) return
  stockDrawerSaving.value = true
  try {
    const newStock = previewStock.value
    await updateProduct(stockDrawerProduct.value.id, {
      stock: newStock,
      stock_warning: stockDrawerForm.stock_warning,
      price_usd: stockDrawerForm.price_usd,
    })
    stockDrawerProduct.value.stock = newStock
    stockDrawerProduct.value.stock_warning = stockDrawerForm.stock_warning
    stockDrawerProduct.value.price_usd = stockDrawerForm.price_usd
    stockDrawerProduct.value.is_low_stock = newStock <= stockDrawerForm.stock_warning
    ElMessage.success(t('product.productUpdated'))
    stockDrawerVisible.value = false
    loadProducts()
  } catch (error) {
    console.error('保存失败:', error)
  } finally {
    stockDrawerSaving.value = false
  }
}

const handleDelete = async (row) => {
  const result = await ElMessageBox.confirm(
    t('product.deleteConfirm', { name: row.name }),
    t('product.deleteTitle'),
    { confirmButtonText: t('common.confirm'), cancelButtonText: t('common.cancel'), type: 'warning' }
  ).catch(() => false)
  if (!result) return
  try {
    await deleteProduct(row.id)
    ElMessage.success(t('product.productDeleted'))
    loadProducts()
  } catch (error) {
    console.error('删除失败:', error)
  }
}

onMounted(async () => {
  await loadProducts()
  loadCategories()
})
</script>

<style scoped>
.products-page {
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

.header-btns {
  display: flex;
  gap: 8px;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.image-upload-area {
  display: flex;
  gap: 12px;
}

.image-slot {
  width: 100px;
  height: 100px;
  border-radius: 6px;
  overflow: hidden;
}

.image-preview {
  position: relative;
  width: 100%;
  height: 100%;
  border: 1px solid #eee;
  border-radius: 6px;
  overflow: hidden;
}

.image-actions {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0,0,0,0.4);
  opacity: 0;
  transition: opacity 0.2s;
  cursor: pointer;
}

.image-preview:hover .image-actions {
  opacity: 1;
}

.image-actions .el-icon {
  color: #fff;
  font-size: 22px;
}

.image-uploader {
  width: 100px;
  height: 100px;
}

.image-uploader :deep(.el-upload) {
  width: 100%;
  height: 100%;
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: border-color 0.2s;
}

.image-uploader :deep(.el-upload:hover) {
  border-color: #409EFF;
}

.upload-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  color: #999;
  font-size: 12px;
}

.upload-placeholder .el-icon {
  font-size: 22px;
}

.upload-tip {
  font-size: 12px;
  color: #999;
  margin-top: 6px;
}

/* ========== 移动端卡片列表 ========== */
.mobile-card-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.product-card {
  display: flex;
  gap: 12px;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 10px;
  padding: 12px;
  cursor: pointer;
  transition: box-shadow 0.2s;
  align-items: center;
  overflow: hidden;
}

.product-card:active {
  background: #f5f7fa;
}

.card-left {
  flex-shrink: 0;
}

.card-img {
  width: 80px;
  height: 80px;
  border-radius: 8px;
  overflow: hidden;
}

.card-img-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
  color: #c0c4cc;
}

.card-right {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.card-title-row {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
  overflow: hidden;
}

.card-name {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  min-width: 0;
}

.card-meta {
  font-size: 12px;
  color: #909399;
}

.card-cat {
  background: #f0f2f5;
  padding: 1px 6px;
  border-radius: 3px;
}

.card-unit {
  margin-left: 4px;
}

.card-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 2px;
}

.card-price {
  font-size: 15px;
  font-weight: 700;
  color: #1D4ED8;
}

.card-actions {
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

/* 库存快捷控制 */
.card-stock-control {
  display: flex;
  align-items: center;
  background: #f5f5f5;
  border-radius: 6px;
  border: 1px solid #e4e7ed;
  overflow: hidden;
}

.stock-btn {
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  color: #606266;
  font-size: 16px;
  font-weight: 700;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
}

.stock-btn:hover {
  background: #e8e8e8;
}

.stock-btn:active {
  background: #d9d9d9;
}

.stock-minus:disabled {
  color: #c0c4cc;
  cursor: not-allowed;
}

.stock-num {
  min-width: 36px;
  text-align: center;
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  cursor: pointer;
  padding: 0 4px;
}

.stock-num.stock-warn {
  color: #E6A23C;
}

/* 库存抽屉 */
.stock-drawer-content {
  padding: 0 8px 20px;
}

.drawer-product-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.drawer-product-name {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.drawer-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.drawer-form-item label {
  display: block;
  font-size: 13px;
  color: #909399;
  margin-bottom: 6px;
}

/* ========== 移动端适配 ========== */
@media (max-width: 767px) {
  .products-page {
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

  .image-upload-area {
    flex-wrap: wrap;
  }

  .image-slot {
    width: 80px;
    height: 80px;
  }

  .image-uploader {
    width: 80px;
    height: 80px;
  }

  .pagination-wrapper :deep(.el-pagination) {
    flex-wrap: wrap;
    justify-content: center;
  }
}

/* ========== 补货/取货弹窗 ========== */
.quick-delta-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.quick-delta-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

.quick-delta-input {
  display: flex;
  align-items: center;
  gap: 12px;
}

.delta-label {
  font-size: 24px;
  font-weight: 700;
  color: #409EFF;
  width: 28px;
  text-align: center;
}

.quick-delta-preview {
  text-align: center;
  font-size: 14px;
  color: #909399;
  padding: 8px;
  background: #f5f7fa;
  border-radius: 6px;
}

/* ========== 库存抽屉delta输入 ========== */
.delta-input-row {
  display: flex;
  gap: 10px;
  align-items: center;
}

.delta-preview {
  margin-top: 8px;
  font-size: 13px;
  color: #606266;
  padding: 6px 10px;
  background: #f5f7fa;
  border-radius: 4px;
}

.delta-preview .stock-warn {
  color: #E6A23C;
  font-weight: 700;
}

/* ========== 库存预览弹窗 ========== */
.inventory-preview {
  max-height: 60vh;
  overflow-y: auto;
}

.inventory-summary {
  display: flex;
  justify-content: space-around;
  gap: 10px;
}

.inv-stat {
  text-align: center;
  flex: 1;
  padding: 10px 0;
}

.inv-num {
  font-size: 24px;
  font-weight: 700;
}

.inv-label {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.inv-total .inv-num { color: #303133; }
.inv-ok .inv-num { color: #67c23a; }
.inv-warn .inv-num { color: #E6A23C; }
.inv-danger .inv-num { color: #f56c6c; }

.inv-filter-row {
  margin: 12px 0;
}

.inv-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: 300px;
  overflow-y: auto;
}

.inv-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-radius: 6px;
  background: #fafafa;
}

.inv-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.inv-dot.inv-ok { background: #67c23a; }
.inv-dot.inv-warn { background: #E6A23C; }
.inv-dot.inv-danger { background: #f56c6c; }

.inv-name {
  flex: 1;
  font-size: 14px;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.inv-stock {
  font-weight: 600;
  font-size: 13px;
  flex-shrink: 0;
}

.inv-stock.inv-ok { color: #67c23a; }
.inv-stock.inv-warn { color: #E6A23C; }
.inv-stock.inv-danger { color: #f56c6c; }
</style>
