<template>
  <div class="supplier-manage">
    <!-- 搜索栏 -->
    <el-card class="search-card" shadow="never">
      <el-form :model="searchForm" inline>
        <el-form-item label="关键词">
          <el-input
            v-model="searchForm.keyword"
            placeholder="供应商名称/编码"
            clearable
            style="width: 200px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="Search" @click="handleSearch">搜索</el-button>
          <el-button icon="Refresh" @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 数据表格 -->
    <el-card class="table-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">供应商管理</span>
          <el-button type="primary" icon="Plus" @click="handleAdd">新增供应商</el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="supplierList"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column prop="supplier_code" label="供应商编码" width="140" />
        <el-table-column prop="supplier_name" label="供应商名称" min-width="180" />
        <el-table-column prop="contact_person" label="联系人" width="120" />
        <el-table-column prop="contact_phone" label="联系电话" width="140" />
        <el-table-column label="累计采购额" width="140" align="right">
          <template #default="{ row }">
            <span class="price">¥{{ row.total_purchase_amount || 0 }}</span>
          </template>
        </el-table-column>
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

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.page_size"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @size-change="fetchList"
          @current-change="fetchList"
        />
      </div>
    </el-card>

    <!-- 新增/编辑弹窗 -->
    
<el-dialog v-dialog-drag
      v-model="formDialogVisible"
      :title="isEdit ? '编辑供应商' : '新增供应商'"
      width="560px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item label="供应商编码" prop="supplier_code">
          <el-input v-model="formData.supplier_code" placeholder="请输入供应商编码" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="供应商名称" prop="supplier_name">
          <el-input v-model="formData.supplier_name" placeholder="请输入供应商名称" />
        </el-form-item>
        <el-form-item label="联系人" prop="contact_person">
          <el-input v-model="formData.contact_person" placeholder="请输入联系人" />
        </el-form-item>
        <el-form-item label="联系电话" prop="contact_phone">
          <el-input v-model="formData.contact_phone" placeholder="请输入联系电话" />
        </el-form-item>
        <el-form-item label="地址" prop="address">
          <el-input v-model="formData.address" type="textarea" :rows="2" placeholder="请输入地址" />
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="formData.remark" type="textarea" :rows="2" placeholder="请输入备注" />
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
  getSupplierList,
  createSupplier,
  updateSupplier,
  deleteSupplier,
} from '@/api/supplier'

// ==================== 搜索与列表 ====================
const loading = ref(false)
const supplierList = ref([])

const searchForm = reactive({
  keyword: '',
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
})

const fetchList = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.page_size,
    }
    if (searchForm.keyword) params.keyword = searchForm.keyword

    const res = await getSupplierList(params)
    supplierList.value = res.data?.list || res.data || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取供应商列表失败:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchList()
}

const handleReset = () => {
  searchForm.keyword = ''
  pagination.page = 1
  fetchList()
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
  contact_person: '',
  contact_phone: '',
  address: '',
  remark: '',
  status: 1,
})

const formRules = {
  supplier_code: [{ required: true, message: '请输入供应商编码', trigger: 'blur' }],
  supplier_name: [{ required: true, message: '请输入供应商名称', trigger: 'blur' }],
  contact_person: [{ required: true, message: '请输入联系人', trigger: 'blur' }],
  contact_phone: [{ required: true, message: '请输入联系电话', trigger: 'blur' }],
}

const resetForm = () => {
  formData.id = null
  formData.supplier_code = ''
  formData.supplier_name = ''
  formData.contact_person = ''
  formData.contact_phone = ''
  formData.address = ''
  formData.remark = ''
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
  formData.supplier_code = row.supplier_code
  formData.supplier_name = row.supplier_name
  formData.contact_person = row.contact_person
  formData.contact_phone = row.contact_phone
  formData.address = row.address
  formData.remark = row.remark
  formData.status = row.status
  formDialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    const data = {
      supplier_code: formData.supplier_code,
      supplier_name: formData.supplier_name,
      contact_person: formData.contact_person,
      contact_phone: formData.contact_phone,
      address: formData.address,
      remark: formData.remark,
      status: formData.status,
    }
    if (isEdit.value) {
      await updateSupplier(formData.id, data)
      ElMessage.success('更新成功')
    } else {
      await createSupplier(data)
      ElMessage.success('创建成功')
    }
    formDialogVisible.value = false
    fetchList()
  } catch (error) {
    console.error('保存供应商失败:', error)
  } finally {
    submitLoading.value = false
  }
}

// ==================== 删除 ====================
const handleDelete = (row) => {
  ElMessageBox.confirm(
    `确定要删除供应商 "${formData.supplier_name}" 吗？此操作不可恢复。`,
    '删除确认',
    {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await deleteSupplier(row.id)
      ElMessage.success('删除成功')
      fetchList()
    } catch (error) {
      console.error('删除供应商失败:', error)
    }
  }).catch(() => {})
}

// ==================== 初始化 ====================
onMounted(() => {
  fetchList()
})
</script>

<style lang="scss" scoped>
.supplier-manage {
  .search-card {
    margin-bottom: 16px;
  }

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

  .table-card {
    .pagination-wrapper {
      display: flex;
      justify-content: flex-end;
      margin-top: 16px;
    }
  }

  .price {
    color: #f56c6c;
    font-weight: 500;
  }
}
</style>
