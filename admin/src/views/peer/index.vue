<template>
  <div class="peer-manage">
    <!-- 搜索栏 -->
    <el-card class="search-card" shadow="never">
      <el-form :model="searchForm" inline>
        <el-form-item label="关键词">
          <el-input
            v-model="searchForm.keyword"
            placeholder="姓名/手机号"
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
          <span class="title">同行列表</span>
          <el-button type="primary" icon="Plus" @click="handleAdd">新增同行</el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="peerList"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column prop="peer_name" label="姓名" width="120" />
        <el-table-column prop="phone" label="手机号" width="140" />
        <el-table-column prop="company" label="公司" width="160" />
        <el-table-column prop="total_orders" label="累计带单数" width="120" align="center" />
        <el-table-column label="累计成交额" width="140" align="right">
          <template #default="{ row }">
            <span class="price">{{ formatCurrency(row.total_amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="累计分成" width="130" align="right">
          <template #default="{ row }">
            <span class="commission">{{ formatCurrency(row.total_commission) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '停用' }}
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
      :title="isEdit ? '编辑同行' : '新增同行'"
      width="550px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item label="姓名" prop="peer_name">
          <el-input v-model="formData.peer_name" placeholder="请输入姓名" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="formData.phone" placeholder="请输入手机号" maxlength="11" />
        </el-form-item>
        <el-form-item label="身份证号" prop="id_card">
          <el-input v-model="formData.id_card" placeholder="请输入身份证号" maxlength="18" />
        </el-form-item>
        <el-form-item label="公司" prop="company">
          <el-input v-model="formData.company" placeholder="请输入公司名称" />
        </el-form-item>
        <el-form-item label="银行卡号" prop="bank_account">
          <el-input v-model="formData.bank_account" placeholder="请输入银行卡号" />
        </el-form-item>
        <el-form-item label="开户行" prop="bank_name">
          <el-input v-model="formData.bank_name" placeholder="请输入开户行名称" />
        </el-form-item>
        <el-form-item label="分成比例" prop="commission_rate">
          <el-input-number
            v-model="formData.commission_rate"
            :min="0"
            :max="100"
            :precision="1"
            controls-position="right"
            style="width: 100%"
            placeholder="请输入分成比例，0表示使用系统默认"
          />
          <span style="margin-left: 8px; color: #909399">%</span>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">停用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="formData.remark" type="textarea" :rows="2" placeholder="请输入备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">
          确认
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPeerList, createPeer, updatePeer, deletePeer } from '@/api/peer'
import { formatCurrency } from '@/utils/format'

// ==================== 搜索与列表 ====================
const loading = ref(false)
const peerList = ref([])

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

    const res = await getPeerList(params)
    peerList.value = res.data?.list || res.data || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取同行列表失败:', error)
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
const submitLoading = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const editingId = ref(null)

const formData = reactive({
  peer_name: '',
  phone: '',
  id_card: '',
  company: '',
  bank_account: '',
  bank_name: '',
  commission_rate: 0,
  status: 1,
  remark: '',
})

const formRules = {
  peer_name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' },
  ],
  company: [{ required: true, message: '请输入公司名称', trigger: 'blur' }],
  status: [{ required: true, message: '请选择状态', trigger: 'change' }],
}

const handleAdd = () => {
  isEdit.value = false
  editingId.value = null
  formData.peer_name = ''
  formData.phone = ''
  formData.id_card = ''
  formData.company = ''
  formData.bank_account = ''
  formData.bank_name = ''
  formData.commission_rate = 0
  formData.status = 1
  formData.remark = ''
  formDialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  editingId.value = row.id
  formData.peer_name = row.peer_name || ''
  formData.phone = row.phone || ''
  formData.id_card = row.id_card || ''
  formData.company = row.company || ''
  formData.bank_account = row.bank_account || ''
  formData.bank_name = row.bank_name || ''
  formData.commission_rate = row.commission_rate ? (row.commission_rate * 100) : 0
  formData.status = row.status ?? 1
  formData.remark = row.remark || ''
  formDialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    const data = {
      peer_name: formData.peer_name,
      phone: formData.phone,
      id_card: formData.id_card,
      company: formData.company,
      bank_account: formData.bank_account,
      bank_name: formData.bank_name,
      commission_rate: formData.commission_rate / 100,
      status: formData.status,
      remark: formData.remark,
    }
    if (isEdit.value) {
      await updatePeer(editingId.value, data)
      ElMessage.success('编辑成功')
    } else {
      await createPeer(data)
      ElMessage.success('新增成功')
    }
    formDialogVisible.value = false
    fetchList()
  } catch (error) {
    console.error('操作失败:', error)
  } finally {
    submitLoading.value = false
  }
}

// ==================== 删除 ====================
const handleDelete = (row) => {
  ElMessageBox.confirm(
    `确定要删除同行 "${row.peer_name}" 吗？`,
    '删除确认',
    {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await deletePeer(row.id)
      ElMessage.success('删除成功')
      fetchList()
    } catch (error) {
      console.error('删除失败:', error)
    }
  }).catch(() => {})
}

// ==================== 初始化 ====================
onMounted(() => {
  fetchList()
})
</script>

<style lang="scss" scoped>
.peer-manage {
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

  .commission {
    color: #e6a23c;
    font-weight: 500;
  }
}
</style>
