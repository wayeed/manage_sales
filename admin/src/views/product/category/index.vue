<template>
  <div class="category-manage">
    <!-- 品类列表 -->
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">品类管理</span>
          <el-button type="primary" icon="Plus" @click="handleAdd">新增品类</el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="categoryList"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column prop="category_code" label="品类编码" width="160" />
        <el-table-column prop="category_name" label="品类名称" width="200" />
        <el-table-column prop="sort_order" label="排序" width="100" align="center" />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新增/编辑弹窗 -->
    
<el-dialog v-dialog-drag
      v-model="formDialogVisible"
      :title="isEdit ? '编辑品类' : '新增品类'"
      width="480px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="80px"
      >
        <el-form-item label="品类编码" prop="category_code">
          <el-input v-model="formData.category_code" placeholder="请输入品类编码" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="品类名称" prop="category_name">
          <el-input v-model="formData.category_name" placeholder="请输入品类名称" />
        </el-form-item>
        <el-form-item label="排序" prop="sort_order">
          <el-input-number v-model="formData.sort_order" :min="0" :max="9999" style="width: 100%" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getCategoryList,
  createCategory,
  updateCategory,
  deleteCategory,
} from '@/api/category'

// ==================== 列表 ====================
const loading = ref(false)
const categoryList = ref([])

const fetchList = async () => {
  loading.value = true
  try {
    const res = await getCategoryList({ page_size: 100 })
    categoryList.value = res.data?.list || res.data || []
  } catch (error) {
    console.error('获取品类列表失败:', error)
  } finally {
    loading.value = false
  }
}

const formatTime = (time) => {
  if (!time) return '-'
  return time.replace('T', ' ').substring(0, 19)
}

// ==================== 新增/编辑 ====================
const formDialogVisible = ref(false)
const isEdit = ref(false)
const submitLoading = ref(false)
const formRef = ref(null)

const formData = reactive({
  id: null,
  category_code: '',
  category_name: '',
  sort_order: 0,
  status: 1,
})

const formRules = {
  category_code: [{ required: true, message: '请输入品类编码', trigger: 'blur' }],
  category_name: [{ required: true, message: '请输入品类名称', trigger: 'blur' }],
}

const resetForm = () => {
  formData.id = null
  formData.category_code = ''
  formData.category_name = ''
  formData.sort_order = 0
  formData.status = 1
}

const handleAdd = () => {
  isEdit.value = false
  resetForm()
  formDialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  formData.id = row.id
  formData.category_code = row.category_code
  formData.category_name = row.category_name
  formData.sort_order = row.sort_order || 0
  formData.status = row.status
  formDialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    const data = {
      category_code: formData.category_code,
      category_name: formData.category_name,
      sort_order: formData.sort_order,
      status: formData.status,
    }
    if (isEdit.value) {
      await updateCategory(formData.id, data)
      ElMessage.success('更新成功')
    } else {
      await createCategory(data)
      ElMessage.success('创建成功')
    }
    formDialogVisible.value = false
    fetchList()
  } catch (error) {
    console.error('保存品类失败:', error)
  } finally {
    submitLoading.value = false
  }
}

// ==================== 删除 ====================
const handleDelete = (row) => {
  ElMessageBox.confirm(
    `确定要删除品类 "${row.category_name}" 吗？此操作不可恢复。`,
    '删除确认',
    {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await deleteCategory(row.id)
      ElMessage.success('删除成功')
      fetchList()
    } catch (error) {
      console.error('删除品类失败:', error)
    }
  }).catch(() => {})
}

// ==================== 初始化 ====================
onMounted(() => {
  fetchList()
})
</script>

<style lang="scss" scoped>
.category-manage {
  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;

    .title {
      font-size: 16px;
      font-weight: 600;
      color: #303133;
    }
  }
}
</style>
