<template>
  <div class="products-page">
    <div class="page-header">
      <h2>{{ $t('admin.products') }}</h2>
      <div class="header-btns">
        <el-button @click="showInventoryPreview" :size="mobile ? 'small' : 'default'">
          <el-icon><document /></el-icon>
          {{ $t('product.inventoryPreview') }}
        </el-button>
        <el-button type="success" plain @click="importVisible = true" :size="mobile ? 'small' : 'default'">
          <el-icon><upload /></el-icon>
          {{ $t('product.batchImport') }}
        </el-button>
        <el-button type="primary" @click="handleAdd" :size="mobile ? 'small' : 'default'">
          <el-icon><plus /></el-icon>
          {{ $t('product.addProduct') }}
        </el-button>
      </div>
    </div>

    <!-- 搜索和分类过滤 -->
    <div class="search-filter-bar">
      <el-input
        v-model="adminSearchKeyword"
        :placeholder="$t('product.searchPlaceholder')"
        clearable
        :prefix-icon="Search"
        :style="mobile ? '' : 'max-width:300px'"
        @clear="adminSearchKeyword = ''"
      />
      <el-select v-model="adminFilterCat" clearable :placeholder="$t('product.all') + ' ' + $t('admin.categories')" style="width:150px">
        <el-option v-for="c in adminCategories" :key="c" :label="c" :value="c" />
      </el-select>
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
      <el-table-column :label="$t('product.nameEn')" prop="name_en" width="130" />
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
      <el-table-column :label="$t('product.expiryDate')" width="110" prop="expiry_date">
        <template #default="{ row }">
          <span v-if="row.expiry_date" :style="isExpiringSoon(row.expiry_date) ? 'color:#f56c6c;font-weight:600' : ''">
            {{ formatDate(row.expiry_date) }}
          </span>
          <span v-else style="color:#ccc">-</span>
        </template>
      </el-table-column>
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
      :width="mobile ? '92vw' : '320px'"
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
        :total="filteredAdminProducts.length"
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
      :width="mobile ? '100vw' : '820px'"
      :fullscreen="mobile"
      destroy-on-close
      class="product-form-dialog"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <el-collapse v-model="openSections" class="form-collapse">

          <!-- 一、供应商核心信息 -->
          <el-collapse-item name="supplier" :title="$t('productForm.sectionSupplier')">
            <el-row :gutter="12">
              <el-col :span="24">
                <el-form-item :label="$t('productForm.supplierName')">
                  <el-input v-model="form.supplier_name" />
                </el-form-item>
              </el-col>
            </el-row>
          </el-collapse-item>

          <!-- 二、商品基础主信息 -->
          <el-collapse-item name="basic" :title="$t('productForm.sectionBasic')">
            <el-row :gutter="12">
              <el-col :xs="8" :sm="4">
                <el-form-item :label="$t('productForm.no')">
                  <el-input v-model="form.sort_order" inputmode="numeric" style="width:100%" />
                </el-form-item>
              </el-col>
              <el-col :xs="16" :sm="10">
                <el-form-item :label="$t('productForm.barcode')">
                  <el-input v-model="form.barcode" clearable>
                    <template #append>
                      <el-button :icon="Camera" @click="openBarcodeScan" :title="$t('product.scanBarcode')" />
                    </template>
                  </el-input>
                </el-form-item>
              </el-col>
              <el-col :xs="12" :sm="5">
                <el-form-item :label="$t('productForm.brand')">
                  <el-input v-model="form.brand" />
                </el-form-item>
              </el-col>
              <el-col :xs="12" :sm="5">
                <el-form-item :label="$t('product.category')">
                  <el-select v-model="form.category" filterable allow-create default-first-option :placeholder="$t('product.selectCategory')" style="width:100%">
                    <el-option v-for="cat in categoryOptions" :key="cat" :label="cat" :value="cat" />
                  </el-select>
                </el-form-item>
              </el-col>
            </el-row>

            <el-row :gutter="12">
              <el-col :xs="24" :sm="12">
                <el-form-item :label="$t('productForm.productName')" prop="name">
                  <el-input v-model="form.name" />
                </el-form-item>
              </el-col>
              <el-col :xs="12" :sm="6">
                <el-form-item :label="$t('productForm.unitWeightValue')">
                  <el-input v-model="form.unit_weight_value" inputmode="decimal" placeholder="0.00" />
                </el-form-item>
              </el-col>
              <el-col :xs="12" :sm="6">
                <el-form-item :label="$t('productForm.unitWeightUnit')">
                  <el-select v-model="form.unit_weight_unit" style="width:100%">
                    <el-option label="G" value="G" />
                    <el-option label="ML" value="ML" />
                    <el-option label="Pcs" value="Pcs" />
                  </el-select>
                </el-form-item>
              </el-col>
            </el-row>

            <el-row :gutter="12">
              <el-col :xs="24" :sm="12">
                <el-form-item :label="$t('productForm.packingFormat')">
                  <el-input v-model="form.packing_format" />
                </el-form-item>
              </el-col>
              <el-col :xs="12" :sm="6">
                <el-form-item :label="$t('productForm.packSize')">
                  <el-input v-model="form.pack_size" inputmode="numeric" style="width:100%" />
                </el-form-item>
              </el-col>
              <el-col :xs="12" :sm="6">
                <el-form-item :label="$t('productForm.shelfLifeDays')">
                  <el-input v-model="form.shelf_life_days" inputmode="numeric" style="width:100%" />
                </el-form-item>
              </el-col>
            </el-row>

            <el-row :gutter="12">
              <el-col :xs="12" :sm="6">
                <el-form-item :label="$t('productForm.costUSD')" prop="price_usd">
                  <el-input v-model="form.price_usd" inputmode="decimal" placeholder="0.00" style="width:100%" />
                </el-form-item>
              </el-col>
              <el-col :xs="12" :sm="6">
                <el-form-item :label="$t('productForm.retailPriceUSD')">
                  <el-input v-model="form.retail_price_usd" inputmode="decimal" placeholder="0.00" style="width:100%" />
                </el-form-item>
              </el-col>
              <el-col :xs="12" :sm="6">
                <el-form-item :label="$t('productForm.gpPercent')">
                  <el-input v-model="form.gp_percent" inputmode="decimal" placeholder="0.00" style="width:100%">
                    <template #suffix>%</template>
                  </el-input>
                </el-form-item>
              </el-col>
              <el-col :xs="12" :sm="6">
                <el-form-item :label="$t('productForm.principleCompany')">
                  <el-input v-model="form.principle_company" />
                </el-form-item>
              </el-col>
            </el-row>

            <el-row :gutter="12">
              <el-col :xs="24" :sm="12">
                <el-form-item :label="$t('productForm.countryOfOrigin')">
                  <el-input v-model="form.country_of_origin" />
                </el-form-item>
              </el-col>
              <el-col :xs="12" :sm="6">
                <el-form-item :label="$t('product.productionDate')">
                  <el-date-picker v-model="form.production_date" type="date" value-format="YYYY-MM-DD" style="width:100%" />
                </el-form-item>
              </el-col>
              <el-col :xs="12" :sm="6">
                <el-form-item :label="$t('product.expiryDate')">
                  <el-date-picker v-model="form.expiry_date" type="date" value-format="YYYY-MM-DD" style="width:100%" />
                </el-form-item>
              </el-col>
            </el-row>

            <!-- 商品图片 -->
            <el-form-item :label="$t('productForm.productPicture')">
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

            <!-- 库存与状态 -->
            <el-row :gutter="12">
              <el-col :xs="12" :sm="4">
                <el-form-item :label="$t('product.stock')">
                  <el-input v-model="form.stock" inputmode="numeric" placeholder="0" style="width:100%" />
                </el-form-item>
              </el-col>
              <el-col :xs="12" :sm="4">
                <el-form-item :label="$t('product.stockWarning')">
                  <el-input v-model="form.stock_warning" inputmode="numeric" placeholder="10" style="width:100%" />
                </el-form-item>
              </el-col>
              <el-col :xs="12" :sm="8">
                <el-form-item :label="$t('product.status')">
                  <el-switch v-model="form.is_active" :active-text="$t('product.onSale')" :inactive-text="$t('product.offSale')" />
                </el-form-item>
              </el-col>
              <el-col :xs="12" :sm="8">
                <el-form-item :label="$t('product.featured')">
                  <el-switch v-model="form.is_featured" />
                </el-form-item>
              </el-col>
            </el-row>
          </el-collapse-item>

          <!-- 三、商品多语言名称 -->
          <el-collapse-item name="names" :title="$t('productForm.sectionNames')">
            <el-row :gutter="12">
              <el-col :xs="24" :sm="8">
                <el-form-item :label="$t('productForm.nameEn')">
                  <el-input v-model="form.name_en" />
                </el-form-item>
              </el-col>
              <el-col :xs="24" :sm="8">
                <el-form-item :label="$t('productForm.nameKh')">
                  <el-input v-model="form.name_kh" />
                </el-form-item>
              </el-col>
              <el-col :xs="24" :sm="8">
                <el-form-item :label="$t('productForm.nameZh')">
                  <el-input v-model="form.name" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-form-item :label="$t('product.description')">
              <el-input v-model="form.description" type="textarea" :rows="2" />
            </el-form-item>
          </el-collapse-item>

          <!-- 四、包装规格层级 -->
          <el-collapse-item name="packing" :title="$t('productForm.sectionPacking')">
            <el-row :gutter="12">
              <el-col :xs="8" :sm="8">
                <el-form-item :label="$t('productForm.unitLabel')">
                  <el-input v-model="form.unit" :placeholder="$t('product.unitPlaceholder')" />
                </el-form-item>
              </el-col>
              <el-col :xs="8" :sm="8">
                <el-form-item :label="$t('productForm.packLabel')">
                  <el-input v-model="form.pack_name" :placeholder="$t('product.packNamePlaceholder')" />
                </el-form-item>
              </el-col>
              <el-col :xs="8" :sm="8">
                <el-form-item :label="$t('productForm.caseLabel')">
                  <el-input v-model="form.unit_name" :placeholder="$t('product.unitNamePlaceholder')" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="12">
              <el-col :xs="8" :sm="8">
                <el-form-item :label="$t('productForm.innerPackPerCase')">
                  <el-input v-model="form.inner_pack_per_case" inputmode="numeric" style="width:100%" />
                </el-form-item>
              </el-col>
              <el-col :xs="8" :sm="8">
                <el-form-item :label="$t('productForm.unitPerInnerPack')">
                  <el-input v-model="form.unit_per_inner_pack" inputmode="numeric" style="width:100%" />
                </el-form-item>
              </el-col>
              <el-col :xs="8" :sm="8">
                <el-form-item :label="$t('productForm.unitPerCase')">
                  <el-input v-model="form.unit_per_case" inputmode="numeric" style="width:100%" />
                </el-form-item>
              </el-col>
            </el-row>
          </el-collapse-item>

          <!-- 五、成本与价格核算 -->
          <el-collapse-item name="cost" :title="$t('productForm.sectionCost')">
            <el-row :gutter="12">
              <el-col :xs="12" :sm="8">
                <el-form-item :label="$t('productForm.costPerCase')">
                  <el-input v-model="form.cost_per_case" inputmode="decimal" placeholder="0.00" style="width:100%" />
                </el-form-item>
              </el-col>
              <el-col :xs="12" :sm="8">
                <el-form-item :label="$t('productForm.dcPercent')">
                  <el-input v-model="form.dc_percent" inputmode="decimal" placeholder="0.00" style="width:100%">
                    <template #suffix>%</template>
                  </el-input>
                </el-form-item>
              </el-col>
              <el-col :xs="12" :sm="8">
                <el-form-item :label="$t('productForm.netCostPerCase')">
                  <el-input v-model="form.net_cost_per_case" inputmode="decimal" placeholder="0.00" style="width:100%" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="12">
              <el-col :xs="12" :sm="8">
                <el-form-item :label="$t('productForm.netCostPerUnit')">
                  <el-input v-model="form.net_cost_per_unit" inputmode="decimal" placeholder="0.00" style="width:100%" />
                </el-form-item>
              </el-col>
              <el-col :xs="12" :sm="8">
                <el-form-item :label="$t('productForm.inclVat')">
                  <el-input v-model="form.price_incl_vat" inputmode="decimal" placeholder="0.00" style="width:100%" />
                </el-form-item>
              </el-col>
              <el-col :xs="12" :sm="8">
                <el-form-item :label="$t('productForm.exclVat')">
                  <el-input v-model="form.price_excl_vat" inputmode="decimal" placeholder="0.00" style="width:100%" />
                </el-form-item>
              </el-col>
            </el-row>
          </el-collapse-item>

          <!-- 六、包装尺寸与重量 -->
          <el-collapse-item name="dimensions" :title="$t('productForm.sectionDimensions')">
            <div class="dim-group-title">{{ $t('productForm.smallestUnit') }}</div>
            <el-row :gutter="12">
              <el-col :span="6"><el-form-item :label="$t('productForm.widthCm')"><el-input v-model="form.unit_width_cm" inputmode="decimal" /></el-form-item></el-col>
              <el-col :span="6"><el-form-item :label="$t('productForm.lengthCm')"><el-input v-model="form.unit_length_cm" inputmode="decimal" /></el-form-item></el-col>
              <el-col :span="6"><el-form-item :label="$t('productForm.heightCm')"><el-input v-model="form.unit_height_cm" inputmode="decimal" /></el-form-item></el-col>
              <el-col :span="6"><el-form-item :label="$t('productForm.weightKg')"><el-input v-model="form.unit_weight_kg" inputmode="decimal" /></el-form-item></el-col>
            </el-row>
            <div class="dim-group-title">{{ $t('productForm.middlePack') }}</div>
            <el-row :gutter="12">
              <el-col :span="6"><el-form-item :label="$t('productForm.widthCm')"><el-input v-model="form.pack_width_cm" inputmode="decimal" /></el-form-item></el-col>
              <el-col :span="6"><el-form-item :label="$t('productForm.lengthCm')"><el-input v-model="form.pack_length_cm" inputmode="decimal" /></el-form-item></el-col>
              <el-col :span="6"><el-form-item :label="$t('productForm.heightCm')"><el-input v-model="form.pack_height_cm" inputmode="decimal" /></el-form-item></el-col>
              <el-col :span="6"><el-form-item :label="$t('productForm.weightKg')"><el-input v-model="form.pack_weight_kg" inputmode="decimal" /></el-form-item></el-col>
            </el-row>
            <div class="dim-group-title">{{ $t('productForm.outerCase') }}</div>
            <el-row :gutter="12">
              <el-col :span="6"><el-form-item :label="$t('productForm.widthCm')"><el-input v-model="form.case_width_cm" inputmode="decimal" /></el-form-item></el-col>
              <el-col :span="6"><el-form-item :label="$t('productForm.lengthCm')"><el-input v-model="form.case_length_cm" inputmode="decimal" /></el-form-item></el-col>
              <el-col :span="6"><el-form-item :label="$t('productForm.heightCm')"><el-input v-model="form.case_height_cm" inputmode="decimal" /></el-form-item></el-col>
              <el-col :span="6"><el-form-item :label="$t('productForm.weightKg')"><el-input v-model="form.case_weight_kg" inputmode="decimal" /></el-form-item></el-col>
            </el-row>
          </el-collapse-item>

        </el-collapse>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- 批量导入对话框 -->
    <el-dialog v-model="importVisible" :title="$t('product.batchImport')" :width="mobile ? '94vw' : '520px'" :fullscreen="mobile">
      <div class="import-tips">
        <p>{{ $t('product.batchImportTip') }}</p>
        <a :href="templateUrl" download class="template-link">{{ $t('product.downloadTemplate') }}</a>
      </div>
      <el-upload
        ref="uploadRef"
        :auto-upload="false"
        :limit="1"
        accept=".csv,.xlsx"
        :on-change="onImportFileChange"
        :on-remove="() => (importFile = null)"
      >
        <el-button type="primary">{{ $t('product.selectFile') }}</el-button>
      </el-upload>
      <el-checkbox v-model="importOverwrite" style="margin-top: 12px;">
        {{ $t('product.overwriteExisting') }}
      </el-checkbox>
      <div v-if="importResult" class="import-result">
        <p>{{ $t('product.imported') }}: {{ importResult.created }} / {{ $t('product.updated') }}: {{ importResult.updated }} / {{ $t('product.skipped') }}: {{ importResult.skipped }}</p>
        <ul v-if="importResult.errors?.length" class="import-errors">
          <li v-for="(err, i) in importResult.errors" :key="i">{{ err }}</li>
        </ul>
      </div>
      <template #footer>
        <el-button @click="importVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="importing" :disabled="!importFile" @click="submitImport">
          {{ $t('product.startImport') }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 扫码对话框 -->
    <el-dialog v-model="scanVisible" :title="$t('product.scanBarcode')" :width="mobile ? '94vw' : '480px'" :fullscreen="mobile" @close="stopScan">
      <div v-if="scanError" class="scan-error">{{ scanError }}</div>
      <video v-show="!scanError" ref="scanVideoRef" class="scan-video" autoplay muted playsinline></video>
      <div class="scan-manual">
        <el-input v-model="manualBarcode" :placeholder="$t('product.enterBarcode')" @keyup.enter="confirmManual">
          <template #append>
            <el-button @click="confirmManual">{{ $t('common.confirm') }}</el-button>
          </template>
        </el-input>
      </div>
      <template #footer>
        <el-button @click="scanVisible = false">{{ $t('common.cancel') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { Plus, Delete, Goods, Upload, Camera, Search } from '@element-plus/icons-vue'
import { Document } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus/es/components/message/index'
import { ElMessageBox } from 'element-plus/es/components/message-box/index'
import { getProducts, createProduct, updateProduct, deleteProduct, uploadImage, getAllCategories, importProducts, getProductImportTemplateUrl } from '@/api'
import { formatUSD } from '@/utils/format'

const formatDate = (d) => {
  if (!d) return ''
  return new Date(d).toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}
const isExpiringSoon = (d, days = 30) => {
  if (!d) return false
  return (new Date(d) - Date.now()) / 86400000 <= days
}

const { t } = useI18n()

// 移动端检测
const mobile = ref(window.innerWidth < 768)
const onResize = () => { mobile.value = window.innerWidth < 768 }
onMounted(() => window.addEventListener('resize', onResize))
onBeforeUnmount(() => {
  window.removeEventListener('resize', onResize)
  stopScan()
})

const loading = ref(false)
const products = ref([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref()
const categoryOptions = ref([])

// 批量导入 / 扫码
const importVisible = ref(false)
const importFile = ref(null)
const importOverwrite = ref(false)
const importing = ref(false)
const importResult = ref(null)
const uploadRef = ref()
const templateUrl = getProductImportTemplateUrl()

const scanVisible = ref(false)
const scanError = ref('')
const scanVideoRef = ref()
const manualBarcode = ref('')
let scanStream = null
let scanDetector = null
let scanRAF = 0

const onImportFileChange = (file) => {
  importFile.value = file?.raw || null
  importResult.value = null
}

const submitImport = async () => {
  if (!importFile.value) return
  importing.value = true
  importResult.value = null
  try {
    const fd = new FormData()
    fd.append('file', importFile.value)
    const data = await importProducts(fd)
    importResult.value = data
    ElMessage.success(t('product.importDone'))
    await loadProducts()
  } catch (e) {
    console.error(e)
  } finally {
    importing.value = false
  }
}

const stopScan = () => {
  if (scanRAF) cancelAnimationFrame(scanRAF)
  scanRAF = 0
  scanDetector = null
  if (scanStream) {
    scanStream.getTracks().forEach((t) => t.stop())
    scanStream = null
  }
}

const openBarcodeScan = async () => {
  scanError.value = ''
  manualBarcode.value = ''
  scanVisible.value = true
  await nextTick()

  if (!('BarcodeDetector' in window)) {
    scanError.value = t('product.scanNotSupported')
    return
  }
  try {
    scanStream = await navigator.mediaDevices.getUserMedia({
      video: { facingMode: 'environment' },
    })
    if (scanVideoRef.value) {
      scanVideoRef.value.srcObject = scanStream
    }
    scanDetector = new window.BarcodeDetector({
      formats: ['ean_13', 'ean_8', 'code_128', 'code_39', 'upc_a', 'upc_e', 'qr_code'],
    })
    const tick = async () => {
      if (!scanDetector || !scanVideoRef.value) return
      try {
        const codes = await scanDetector.detect(scanVideoRef.value)
        if (codes && codes.length > 0) {
          form.barcode = codes[0].rawValue
          ElMessage.success(t('product.scanSuccess'))
          scanVisible.value = false
          stopScan()
          return
        }
      } catch {
        // noop
      }
      scanRAF = requestAnimationFrame(tick)
    }
    scanRAF = requestAnimationFrame(tick)
  } catch (e) {
    scanError.value = t('product.scanCameraError')
  }
}

const confirmManual = () => {
  if (!manualBarcode.value) return
  form.barcode = manualBarcode.value.trim()
  scanVisible.value = false
  stopScan()
}

// 分页
const currentPage = ref(1)
const pageSize = ref(20)
const adminSearchKeyword = ref('')
const adminFilterCat = ref('')

const adminCategories = computed(() => {
  const cats = [...new Set(products.value.map(p => p.category).filter(Boolean))]
  return cats.sort()
})

const filteredAdminProducts = computed(() => {
  let list = products.value
  if (adminFilterCat.value) {
    list = list.filter(p => p.category === adminFilterCat.value)
  }
  if (adminSearchKeyword.value.trim()) {
    const kw = adminSearchKeyword.value.toLowerCase()
    list = list.filter(p =>
      (p.name && p.name.toLowerCase().includes(kw)) ||
      (p.name_kh && p.name_kh.toLowerCase().includes(kw)) ||
      (p.name_en && p.name_en.toLowerCase().includes(kw)) ||
      (p.category && p.category.toLowerCase().includes(kw)) ||
      (p.brand && p.brand.toLowerCase().includes(kw))
    )
  }
  return list
})

const paginatedProducts = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredAdminProducts.value.slice(start, end)
})

const handleSizeChange = () => { currentPage.value = 1 }
const handleCurrentChange = () => { window.scrollTo({ top: 0, behavior: 'smooth' }) }

const form = reactive({
  id: null,
  // 一、供应商
  supplier_name: '',
  // 二、基础信息
  sort_order: 0,
  barcode: '',
  brand: '',
  category: '',
  name: '',
  unit_weight_value: '',
  unit_weight_unit: 'G',
  packing_format: '',
  pack_size: '',
  price_usd: '',
  retail_price_usd: '',
  gp_percent: '',
  shelf_life_days: '',
  principle_company: '',
  country_of_origin: '',
  production_date: null,
  expiry_date: null,
  image_url: '',
  img1: '', img2: '', img3: '', img4: '', img5: '',
  stock: 0,
  stock_warning: 10,
  is_active: true,
  is_featured: false,
  // 三、多语言名称
  name_en: '',
  name_kh: '',
  description: '',
  // 四、包装规格层级
  unit: '',
  pack_name: '',
  unit_name: '',
  inner_pack_per_case: '',
  unit_per_inner_pack: '',
  unit_per_case: '',
  pieces_per_package: '',
  // 五、成本与价格
  cost_per_case: '',
  dc_percent: '',
  net_cost_per_case: '',
  net_cost_per_unit: '',
  price_incl_vat: '',
  price_excl_vat: '',
  price_per_piece_usd: '',
  price_per_package_usd: '',
  // 六、尺寸重量
  unit_width_cm: '', unit_length_cm: '', unit_height_cm: '', unit_weight_kg: '',
  pack_width_cm: '', pack_length_cm: '', pack_height_cm: '', pack_weight_kg: '',
  case_width_cm: '', case_length_cm: '', case_height_cm: '', case_weight_kg: '',
  specs: '',
})

const openSections = ref(['supplier', 'basic', 'names', 'packing', 'cost', 'dimensions'])

const rules = {
  name: [{ required: true, message: () => t('product.nameRequired'), trigger: 'blur' }],
  price_usd: [
    { required: true, message: () => t('product.priceRequired'), trigger: 'blur' },
    {
      validator: (_rule, value, callback) => {
        const n = parseFloat(value)
        if (isNaN(n) || n <= 0) callback(new Error(t('product.priceRequired')))
        else callback()
      },
      trigger: 'blur',
    },
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
    supplier_name: '',
    sort_order: 0, barcode: '', brand: '', category: '', name: '',
    unit_weight_value: '', unit_weight_unit: 'G',
    packing_format: '', pack_size: '',
    price_usd: '', retail_price_usd: '', gp_percent: '',
    shelf_life_days: '', principle_company: '', country_of_origin: '',
    production_date: null, expiry_date: null,
    image_url: '', img1: '', img2: '', img3: '', img4: '', img5: '',
    stock: 0, stock_warning: 10, is_active: true, is_featured: false,
    name_en: '', name_kh: '', description: '',
    unit: '', pack_name: '', unit_name: '',
    inner_pack_per_case: '', unit_per_inner_pack: '', unit_per_case: '',
    pieces_per_package: '',
    cost_per_case: '', dc_percent: '', net_cost_per_case: '',
    net_cost_per_unit: '', price_incl_vat: '', price_excl_vat: '',
    price_per_piece_usd: '', price_per_package_usd: '',
    unit_width_cm: '', unit_length_cm: '', unit_height_cm: '', unit_weight_kg: '',
    pack_width_cm: '', pack_length_cm: '', pack_height_cm: '', pack_weight_kg: '',
    case_width_cm: '', case_length_cm: '', case_height_cm: '', case_weight_kg: '',
    specs: '',
  })
  formRef.value?.clearValidate()
}

const toStr = (v) => (v != null ? String(v) : '')

const handleAdd = () => {
  resetForm()
  isEdit.value = false
  dialogVisible.value = true
}

const handleEdit = (row) => {
  Object.assign(form, {
    ...row,
    supplier_name: row.supplier_name || '',
    name: row.name || '',
    name_en: row.name_en || '',
    name_kh: row.name_kh || '',
    brand: row.brand || '',
    country_of_origin: row.country_of_origin || '',
    packing_format: row.packing_format || '',
    principle_company: row.principle_company || '',
    unit: row.unit || '',
    unit_name: row.unit_name || '',
    pack_name: row.pack_name || '',
    barcode: row.barcode || '',
    description: row.description || '',
    specs: row.specs || '',
    unit_weight_unit: row.unit_weight_unit || 'G',
    img1: row.img1 || '', img2: row.img2 || '', img3: row.img3 || '',
    img4: row.img4 || '', img5: row.img5 || '',
    production_date: row.production_date ? row.production_date.substring(0, 10) : null,
    expiry_date: row.expiry_date ? row.expiry_date.substring(0, 10) : null,
    sort_order: row.sort_order || 0,
    is_active: row.is_active !== false,
    is_featured: row.is_featured || false,
    // numeric string fields
    price_usd: toStr(row.price_usd),
    retail_price_usd: toStr(row.retail_price_usd),
    price_per_piece_usd: toStr(row.price_per_piece_usd),
    price_per_package_usd: toStr(row.price_per_package_usd),
    pieces_per_package: toStr(row.pieces_per_package),
    unit_weight_value: toStr(row.unit_weight_value),
    pack_size: toStr(row.pack_size),
    gp_percent: toStr(row.gp_percent),
    shelf_life_days: toStr(row.shelf_life_days),
    inner_pack_per_case: toStr(row.inner_pack_per_case),
    unit_per_inner_pack: toStr(row.unit_per_inner_pack),
    unit_per_case: toStr(row.unit_per_case),
    cost_per_case: toStr(row.cost_per_case),
    dc_percent: toStr(row.dc_percent),
    net_cost_per_case: toStr(row.net_cost_per_case),
    net_cost_per_unit: toStr(row.net_cost_per_unit),
    price_incl_vat: toStr(row.price_incl_vat),
    price_excl_vat: toStr(row.price_excl_vat),
    unit_width_cm: toStr(row.unit_width_cm), unit_length_cm: toStr(row.unit_length_cm),
    unit_height_cm: toStr(row.unit_height_cm), unit_weight_kg: toStr(row.unit_weight_kg),
    pack_width_cm: toStr(row.pack_width_cm), pack_length_cm: toStr(row.pack_length_cm),
    pack_height_cm: toStr(row.pack_height_cm), pack_weight_kg: toStr(row.pack_weight_kg),
    case_width_cm: toStr(row.case_width_cm), case_length_cm: toStr(row.case_length_cm),
    case_height_cm: toStr(row.case_height_cm), case_weight_kg: toStr(row.case_weight_kg),
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
      // 将数字字段从字符串转换为数字
      const floatFields = [
        'price_usd', 'retail_price_usd', 'price_per_piece_usd', 'price_per_package_usd',
        'unit_weight_value', 'pack_size', 'gp_percent',
        'cost_per_case', 'dc_percent', 'net_cost_per_case', 'net_cost_per_unit', 'price_incl_vat', 'price_excl_vat',
        'unit_width_cm', 'unit_length_cm', 'unit_height_cm', 'unit_weight_kg',
        'pack_width_cm', 'pack_length_cm', 'pack_height_cm', 'pack_weight_kg',
        'case_width_cm', 'case_length_cm', 'case_height_cm', 'case_weight_kg',
      ]
      const intFields = [
        'pieces_per_package', 'stock', 'stock_warning', 'sort_order',
        'shelf_life_days', 'inner_pack_per_case', 'unit_per_inner_pack', 'unit_per_case',
      ]
      for (const key of floatFields) {
        const n = parseFloat(payload[key])
        payload[key] = isNaN(n) ? null : n
      }
      for (const key of intFields) {
        const n = parseInt(payload[key])
        payload[key] = isNaN(n) ? null : n
      }
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

.search-filter-bar {
  display: flex;
  gap: 10px;
  align-items: stretch;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.search-filter-bar :deep(.el-input__wrapper),
.search-filter-bar :deep(.el-select__wrapper) {
  height: 32px;
  box-sizing: border-box;
}

@media (max-width: 767px) {
  .search-filter-bar {
    gap: 8px;
  }
  .search-filter-bar :deep(.el-input),
  .search-filter-bar :deep(.el-select) {
    flex: 1 1 0;
    min-width: 0;
    max-width: none !important;
    width: auto !important;
  }
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

/* ========== 商品表单折叠面板 ========== */
.form-collapse :deep(.el-collapse-item__header) {
  font-weight: 600;
  color: var(--el-color-primary);
  background: #f5f7fa;
  padding: 0 12px;
  border-radius: 4px;
  margin-bottom: 2px;
}
.form-collapse :deep(.el-collapse-item__content) {
  padding: 12px 0 0;
}
.dim-group-title {
  font-size: 13px;
  font-weight: 600;
  color: #555;
  margin: 8px 0 4px;
  padding-left: 6px;
  border-left: 3px solid var(--el-color-primary);
}
.product-form-dialog :deep(.el-dialog__body) {
  max-height: 72vh;
  overflow-y: auto;
  padding: 16px 20px;
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
    max-height: calc(100dvh - 130px);
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

.import-tips {
  margin-bottom: 12px;
  font-size: 13px;
  color: #606266;
}
.template-link {
  color: #409eff;
  margin-left: 8px;
  text-decoration: underline;
}
.import-result {
  margin-top: 12px;
  padding: 8px 12px;
  background: #f5f7fa;
  border-radius: 4px;
  font-size: 13px;
}
.import-errors {
  margin: 6px 0 0 16px;
  color: #f56c6c;
  max-height: 160px;
  overflow-y: auto;
}
.scan-video {
  width: 100%;
  max-height: 60vh;
  background: #000;
  border-radius: 4px;
}
.scan-error {
  padding: 12px;
  color: #f56c6c;
  background: #fef0f0;
  border-radius: 4px;
  margin-bottom: 12px;
}
.scan-manual {
  margin-top: 12px;
}
</style>
