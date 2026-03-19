<template>
  <div class="categories-page">
    <div class="page-header">
      <h2>{{ $t('category.title') }}</h2>
      <el-button type="primary" @click="handleAdd" :size="mobile ? 'small' : 'default'">
        <el-icon><plus /></el-icon>
        {{ $t('category.addCategory') }}
      </el-button>
    </div>

    <!-- 桌面端: 表格视图 -->
    <el-table v-if="!mobile" v-loading="loading" :data="categories" border row-key="id" :default-sort="{ prop: 'sort_order', order: 'ascending' }">
      <el-table-column :label="$t('category.sortOrder')" prop="sort_order" width="80" sortable />
      <el-table-column label="ID" prop="id" width="60" />
      <el-table-column :label="$t('category.categoryName')" prop="name" min-width="200" />
      <el-table-column :label="$t('common.status')" width="100">
        <template #default="{ row }">
          <el-switch
            :model-value="row.is_active"
            size="small"
            @change="(val) => handleToggleActive(row, val)"
          />
        </template>
      </el-table-column>
      <el-table-column :label="$t('category.createdAt')" width="180">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column :label="$t('common.operation')" width="150" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link size="small" @click="handleEdit(row)">{{ $t('common.edit') }}</el-button>
          <el-button type="danger" link size="small" @click="handleDelete(row)">{{ $t('common.delete') }}</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 移动端: 卡片列表 -->
    <div v-else v-loading="loading" class="mobile-card-list">
      <div v-for="row in categories" :key="row.id" class="category-card" @click="handleEdit(row)">
        <div class="card-top">
          <div class="card-name-area">
            <span class="card-sort">{{ row.sort_order }}</span>
            <span class="card-name">{{ row.name }}</span>
          </div>
          <el-switch
            :model-value="row.is_active"
            size="small"
            @change="(val) => handleToggleActive(row, val)"
            @click.stop
          />
        </div>
        <div class="card-bottom">
          <span class="card-time">{{ formatDate(row.created_at) }}</span>
          <el-button type="danger" link size="small" @click.stop="handleDelete(row)">{{ $t('common.delete') }}</el-button>
        </div>
      </div>
      <el-empty v-if="!loading && categories.length === 0" />
    </div>

    <!-- 添加/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? $t('category.editCategory') : $t('category.addCategory')"
      :width="mobile ? '94vw' : '440px'"
      :fullscreen="mobile"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="rules" :label-width="mobile ? '70px' : '80px'" :label-position="mobile ? 'top' : 'right'">
        <el-form-item :label="$t('category.name')" prop="name">
          <el-input v-model="form.name" :placeholder="$t('category.namePlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('category.sortOrder')" prop="sort_order">
          <el-input-number v-model="form.sort_order" :min="0" controls-position="right" />
          <span style="margin-left: 8px; color: #999; font-size: 12px;">{{ $t('category.sortHint') }}</span>
        </el-form-item>
        <el-form-item v-if="isEdit" :label="$t('common.status')">
          <el-switch v-model="form.is_active" :active-text="$t('common.enabled')" :inactive-text="$t('common.disabled')" />
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
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { Plus } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus/es/components/message/index'
import { ElMessageBox } from 'element-plus/es/components/message-box/index'
import { getAllCategories, createCategory, updateCategory, deleteCategory } from '@/api'

const { t } = useI18n()

// 移动端检测
const mobile = ref(window.innerWidth < 768)
const onResize = () => { mobile.value = window.innerWidth < 768 }
onMounted(() => window.addEventListener('resize', onResize))
onBeforeUnmount(() => window.removeEventListener('resize', onResize))

const loading = ref(false)
const categories = ref([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref()

const form = reactive({
  id: null,
  name: '',
  sort_order: 0,
  is_active: true,
})

const rules = {
  name: [{ required: true, message: () => t('category.nameRequired'), trigger: 'blur' }],
}

const formatDate = (str) => {
  if (!str) return ''
  const d = new Date(str)
  return d.toLocaleDateString('zh-CN') + ' ' + d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

const loadCategories = async () => {
  loading.value = true
  try {
    categories.value = await getAllCategories()
  } catch (error) {
    console.error('加载分类失败:', error)
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  Object.assign(form, { id: null, name: '', sort_order: 0, is_active: true })
  formRef.value?.clearValidate()
}

const handleAdd = () => {
  resetForm()
  isEdit.value = false
  dialogVisible.value = true
}

const handleEdit = (row) => {
  Object.assign(form, { id: row.id, name: row.name, sort_order: row.sort_order, is_active: row.is_active })
  isEdit.value = true
  dialogVisible.value = true
}

const handleToggleActive = async (row, val) => {
  try {
    await updateCategory(row.id, { is_active: val })
    row.is_active = val
    ElMessage.success(val ? t('common.enabled') : t('common.disabled'))
  } catch (error) {
    ElMessage.error(t('common.operationFailed'))
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      if (isEdit.value) {
        await updateCategory(form.id, { name: form.name, sort_order: form.sort_order, is_active: form.is_active })
        ElMessage.success(t('category.categoryUpdated'))
      } else {
        await createCategory({ name: form.name, sort_order: form.sort_order })
        ElMessage.success(t('category.categoryAdded'))
      }
      dialogVisible.value = false
      loadCategories()
    } catch (error) {
      console.error('提交失败:', error)
    } finally {
      submitting.value = false
    }
  })
}

const handleDelete = async (row) => {
  const result = await ElMessageBox.confirm(
    t('category.deleteConfirm', { name: row.name }),
    t('common.confirmDelete'),
    { confirmButtonText: t('common.confirm'), cancelButtonText: t('common.cancel'), type: 'warning' }
  ).catch(() => false)
  if (!result) return
  try {
    await deleteCategory(row.id)
    ElMessage.success(t('category.categoryDeleted'))
    loadCategories()
  } catch (error) {
    console.error('删除失败:', error)
  }
}

onMounted(() => {
  loadCategories()
})
</script>

<style scoped>
.categories-page {
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

/* ========== 移动端卡片列表 ========== */
.mobile-card-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.category-card {
  background: #fff;
  border-radius: 10px;
  padding: 14px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
  cursor: pointer;
  transition: box-shadow 0.2s;
}

.category-card:active {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
}

.card-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.card-name-area {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  min-width: 0;
}

.card-sort {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  background: #409eff;
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  border-radius: 6px;
  flex-shrink: 0;
}

.card-name {
  font-size: 15px;
  font-weight: 600;
  color: #1a1a1a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-time {
  font-size: 12px;
  color: #999;
}

/* ========== 移动端适配 ========== */
@media (max-width: 767px) {
  .categories-page {
    padding: 12px;
  }

  .page-header {
    margin-bottom: 14px;
  }

  .page-header h2 {
    font-size: 18px;
  }
}
</style>
