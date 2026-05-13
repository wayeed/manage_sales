<template>
  <div class="warehouse-manage">
    <!-- 数据表格 -->
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">仓库管理</span>
          <el-button type="primary" icon="Plus" @click="handleAdd">新增仓库</el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="warehouseList"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column prop="warehouse_code" label="仓库编码" width="140" />
        <el-table-column prop="warehouse_name" label="仓库名称" width="180" />
        <el-table-column label="类型" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="getTypeTag(row.warehouse_type)" size="small">
              {{ getTypeLabel(row.warehouse_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="address" label="地址" min-width="200" show-overflow-tooltip />
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
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
      :title="isEdit ? '编辑仓库' : '新增仓库'"
      width="560px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="80px"
      >
        <el-form-item label="仓库编码" prop="code">
          <el-input v-model="formData.code" placeholder="请输入仓库编码" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="仓库名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入仓库名称" />
        </el-form-item>
        <el-form-item label="仓库类型" prop="type">
          <el-select v-model="formData.type" placeholder="请选择仓库类型" style="width: 100%">
            <el-option label="总仓" :value="1" />
            <el-option label="门店仓" :value="2" />
            <el-option label="中转仓" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="地址" prop="address">
          <el-input v-model="formData.address" type="textarea" :rows="2" placeholder="请输入仓库地址" />
        </el-form-item>
        <el-form-item label="联系人" prop="contact_person">
          <el-input v-model="formData.contact_person" placeholder="请输入联系人" />
        </el-form-item>
        <el-form-item label="联系电话" prop="contact_phone">
          <el-input v-model="formData.contact_phone" placeholder="请输入联系电话" />
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
  getWarehouseList,
  createWarehouse,
  updateWarehouse,
  deleteWarehouse,
} from '@/api/warehouse'

// ==================== 列表 ====================
const loading = ref(false)
const warehouseList = ref([])

const fetchList = async () => {
  loading.value = true
  try {
    const res = await getWarehouseList()
    warehouseList.value = res.data || []
  } catch (error) {
    console.error('获取仓库列表失败:', error)
  } finally {
    loading.value = false
  }
}

const getTypeLabel = (type) => {
  const map = { 1: '总仓', 2: '门店仓', 3: '中转仓' }
  return map[type] || '未知'
}

const getTypeTag = (type) => {
  const map = { 1: 'danger', 2: '', 3: 'warning' }
  return map[type] || 'info'
}

// ==================== 新增/编辑 ====================
const formDialogVisible = ref(false)
const isEdit = ref(false)
const submitLoading = ref(false)
const formRef = ref(null)

const formData = reactive({
  id: null,
  code: '',
  name: '',
  type: 1,
  address: '',
  contact_person: '',
  contact_phone: '',
  status: 1,
})

const formRules = {
  code: [{ required: true, message: '请输入仓库编码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入仓库名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择仓库类型', trigger: 'change' }],
}

const resetForm = () => {
  formData.id = null
  formData.code = ''
  formData.name = ''
  formData.type = 1
  formData.address = ''
  formData.contact_person = ''
  formData.contact_phone = ''
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
  formData.code = row.warehouse_code
  formData.name = row.warehouse_name
  formData.type = row.warehouse_type
  formData.address = row.address
  formData.contact_person = row.contact_person
  formData.contact_phone = row.contact_phone
  formData.status = row.status
  formDialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    const data = {
      warehouse_code: formData.code,
      warehouse_name: formData.name,
      warehouse_type: formData.type,
      address: formData.address,
      contact_person: formData.contact_person,
      contact_phone: formData.contact_phone,
      status: formData.status,
    }
    if (isEdit.value) {
      await updateWarehouse(formData.id, data)
      ElMessage.success('更新成功')
    } else {
      await createWarehouse(data)
      ElMessage.success('创建成功')
    }
    formDialogVisible.value = false
    fetchList()
  } catch (error) {
    console.error('保存仓库失败:', error)
  } finally {
    submitLoading.value = false
  }
}

// ==================== 删除 ====================
const handleDelete = (row) => {
  ElMessageBox.confirm(
    `确定要删除仓库 "${row.warehouse_name}" 吗？此操作不可恢复。`,
    '删除确认',
    {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await deleteWarehouse(row.id)
      ElMessage.success('删除成功')
      fetchList()
    } catch (error) {
      console.error('删除仓库失败:', error)
    }
  }).catch(() => {})
}

// ==================== 初始化 ====================
onMounted(() => {
  fetchList()
})
</script>

<style lang="scss" scoped>
.warehouse-manage {
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
