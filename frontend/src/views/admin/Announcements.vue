<template>
  <div class="announcements-page">
    <div class="page-header">
      <h2>{{ $t('announcement.title') }}</h2>
    </div>

    <!-- 三个 Tab 区域 -->
    <el-tabs v-model="activeTab" @tab-change="loadData">
      <el-tab-pane :label="$t('announcement.notice')" name="notice" />
      <el-tab-pane :label="$t('announcement.contact')" name="contact" />
    </el-tabs>

    <!-- 操作栏 -->
    <div class="action-bar">
      <el-button type="primary" @click="handleAdd" :size="mobile ? 'small' : 'default'">
        <el-icon><plus /></el-icon>
        {{ $t('common.add') }}
      </el-button>
    </div>

    <!-- 桌面端: 列表 -->
    <el-table v-if="!mobile" v-loading="loading" :data="items" border>
      <el-table-column label="ID" prop="id" width="60" />
      <el-table-column :label="$t('announcement.contentZh')" min-width="200">
        <template #default="{ row }">
          <div class="content-cell">{{ row.content_zh }}</div>
        </template>
      </el-table-column>
      <el-table-column :label="$t('announcement.contentEn')" min-width="200">
        <template #default="{ row }">
          <div class="content-cell">{{ row.content_en }}</div>
        </template>
      </el-table-column>
      <el-table-column :label="$t('product.sortOrder')" prop="sort_order" width="80" />
      <el-table-column :label="$t('product.status')" width="80">
        <template #default="{ row }">
          <el-switch
            :model-value="row.is_active"
            size="small"
            @change="(val) => toggleActive(row, val)"
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
      <div v-for="row in items" :key="row.id" class="announce-card" @click="handleEdit(row)">
        <div class="card-top">
          <div class="card-content-preview">{{ row.content_zh || row.content_en }}</div>
          <el-switch
            :model-value="row.is_active"
            size="small"
            @change="(val) => toggleActive(row, val)"
            @click.stop
          />
        </div>
        <div v-if="row.content_en && row.content_zh" class="card-content-en">{{ row.content_en }}</div>
        <div class="card-bottom">
          <span class="card-sort-badge">{{ $t('product.sortOrder') }}: {{ row.sort_order }}</span>
          <el-button type="danger" link size="small" @click.stop="handleDelete(row)">{{ $t('common.delete') }}</el-button>
        </div>
      </div>
      <el-empty v-if="!loading && items.length === 0" />
    </div>

    <!-- 编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? $t('announcement.edit') : $t('announcement.add')"
      :width="mobile ? '94vw' : '600px'"
      :fullscreen="mobile"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :label-width="mobile ? '80px' : '100px'" :label-position="mobile ? 'top' : 'right'">
        <el-form-item :label="$t('announcement.contentZh')" prop="content_zh" :rules="[{ required: true, message: $t('announcement.contentRequired'), trigger: 'blur' }]">
          <el-input v-model="form.content_zh" type="textarea" :rows="4" :placeholder="$t('announcement.contentZhPlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('announcement.contentEn')" prop="content_en">
          <el-input v-model="form.content_en" type="textarea" :rows="4" :placeholder="$t('announcement.contentEnPlaceholder')" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item :label="$t('product.sortOrder')">
              <el-input-number v-model="form.sort_order" :min="0" controls-position="right" style="width: 100%;" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="$t('product.status')">
              <el-switch v-model="form.is_active" :active-text="$t('common.enabled')" :inactive-text="$t('common.disabled')" />
            </el-form-item>
          </el-col>
        </el-row>
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
import { getAnnouncements, createAnnouncement, updateAnnouncement, deleteAnnouncement } from '@/api'

const { t } = useI18n()

// 移动端检测
const mobile = ref(window.innerWidth < 768)
const onResize = () => { mobile.value = window.innerWidth < 768 }
onMounted(() => window.addEventListener('resize', onResize))
onBeforeUnmount(() => window.removeEventListener('resize', onResize))

const loading = ref(false)
const items = ref([])
const activeTab = ref('notice')
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref()

const form = reactive({
  id: null,
  content_zh: '',
  content_en: '',
  sort_order: 0,
  is_active: true,
})

const loadData = async () => {
  loading.value = true
  try {
    const data = await getAnnouncements(activeTab.value)
    items.value = data
  } catch (error) {
    console.error('加载公告失败:', error)
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  Object.assign(form, { id: null, content_zh: '', content_en: '', sort_order: 0, is_active: true })
  formRef.value?.clearValidate()
}

const handleAdd = () => {
  resetForm()
  isEdit.value = false
  dialogVisible.value = true
}

const handleEdit = (row) => {
  Object.assign(form, {
    id: row.id,
    content_zh: row.content_zh,
    content_en: row.content_en,
    sort_order: row.sort_order,
    is_active: row.is_active,
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
      const payload = {
        content_zh: form.content_zh,
        content_en: form.content_en,
        sort_order: form.sort_order,
        is_active: form.is_active,
      }
      if (isEdit.value) {
        await updateAnnouncement(form.id, payload)
        ElMessage.success(t('announcement.updated'))
      } else {
        payload.type = activeTab.value
        await createAnnouncement(payload)
        ElMessage.success(t('announcement.added'))
      }
      dialogVisible.value = false
      loadData()
    } catch (error) {
      console.error('提交失败:', error)
    } finally {
      submitting.value = false
    }
  })
}

const toggleActive = async (row, val) => {
  try {
    await updateAnnouncement(row.id, { is_active: val })
    loadData()
  } catch (error) {
    console.error('切换状态失败:', error)
  }
}

const handleDelete = async (row) => {
  const result = await ElMessageBox.confirm(
    t('announcement.deleteConfirm'),
    t('common.confirmDelete'),
    { confirmButtonText: t('common.confirm'), cancelButtonText: t('common.cancel'), type: 'warning' }
  ).catch(() => false)
  if (!result) return
  try {
    await deleteAnnouncement(row.id)
    ElMessage.success(t('announcement.deleted'))
    loadData()
  } catch (error) {
    console.error('删除失败:', error)
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.announcements-page {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.page-header h2 {
  margin: 0;
}

.action-bar {
  margin-bottom: 16px;
}

.content-cell {
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  white-space: pre-wrap;
  max-height: 44px;
  font-size: 13px;
  line-height: 1.5;
}

/* ========== 移动端卡片列表 ========== */
.mobile-card-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.announce-card {
  background: #fff;
  border-radius: 10px;
  padding: 14px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
  cursor: pointer;
  transition: box-shadow 0.2s;
}

.announce-card:active {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
}

.card-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 6px;
}

.card-content-preview {
  flex: 1;
  font-size: 14px;
  font-weight: 500;
  color: #1a1a1a;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  white-space: pre-wrap;
}

.card-content-en {
  font-size: 12px;
  color: #8c8c8c;
  line-height: 1.4;
  margin-bottom: 8px;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 8px;
  border-top: 1px solid #f5f5f5;
}

.card-sort-badge {
  font-size: 12px;
  color: #999;
}

/* ========== 移动端适配 ========== */
@media (max-width: 767px) {
  .announcements-page {
    padding: 12px;
  }

  .page-header {
    margin-bottom: 10px;
  }

  .page-header h2 {
    font-size: 18px;
  }

  .action-bar {
    margin-bottom: 10px;
  }
}
</style>
