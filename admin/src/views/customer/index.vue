<template>
  <div class="customer-manage">
    <!-- 搜索栏 -->
    <el-card class="search-card" shadow="never">
      <el-form :model="searchForm" inline>
        <el-form-item label="关键词">
          <el-input
            v-model="searchForm.keyword"
            placeholder="姓名/手机号/编码"
            clearable
            style="width: 200px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="客户等级">
          <el-select
            v-model="searchForm.level"
            placeholder="全部等级"
            clearable
            style="width: 140px"
          >
            <el-option label="普通" :value="0" />
            <el-option label="银卡" :value="1" />
            <el-option label="金卡" :value="2" />
            <el-option label="钻石" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select
            v-model="searchForm.status"
            placeholder="全部状态"
            clearable
            style="width: 120px"
          >
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
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
          <span class="title">客户列表</span>
          <el-button type="primary" icon="Plus" @click="handleAdd">新增客户</el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="customerList"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column prop="customer_code" label="客户编码" width="120" />
        <el-table-column prop="customer_name" label="姓名" width="100" />
        <el-table-column prop="phone" label="手机号" width="130" />
        <el-table-column label="原手机号" width="130">
          <template #default="{ row }">
            {{ row.original_phone || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="客户来源" width="100" align="center">
          <template #default="{ row }">
            {{ getSourceTypeLabel(row.source_type) }}
          </template>
        </el-table-column>
        <el-table-column label="性别" width="70" align="center">
          <template #default="{ row }">
            {{ row.gender === 1 ? '男' : row.gender === 2 ? '女' : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="等级" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="getLevelTag(row.level)" size="small">
              {{ getLevelLabel(row.level) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="负责业务员" width="100">
          <template #default="{ row }">
            {{ row.salesman?.real_name || row.salesman_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="total_orders" label="订单数" width="80" align="center" />
        <el-table-column label="累计消费" width="110" align="right">
          <template #default="{ row }">
            <span class="price">{{ formatCurrency(row.total_amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="最后下单" width="110" align="center">
          <template #default="{ row }">
            {{ row.last_order_at ? row.last_order_at.substring(0, 10) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="70" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button type="success" link size="small" @click="handleFollowUp(row)">
              跟进
            </el-button>
            <el-button
              :type="row.status === 1 ? 'warning' : 'success'"
              link
              size="small"
              @click="handleToggleStatus(row)"
            >
              {{ row.status === 1 ? '禁用' : '启用' }}
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
    <el-dialog
      v-model="formDialogVisible"
      :title="isEdit ? '编辑客户' : '新增客户'"
      width="650px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="90px"
      >
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="客户姓名" prop="customer_name">
              <el-input v-model="formData.customer_name" placeholder="请输入客户姓名" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="手机号" prop="phone">
              <el-input v-model="formData.phone" placeholder="请输入手机号" maxlength="11" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="客户编码">
              <el-input v-model="formData.customer_code" placeholder="系统自动生成" disabled />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="邮箱">
              <el-input v-model="formData.email" placeholder="请输入邮箱" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="性别">
              <el-radio-group v-model="formData.gender">
                <el-radio :value="0">未知</el-radio>
                <el-radio :value="1">男</el-radio>
                <el-radio :value="2">女</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="生日">
              <el-date-picker
                v-model="formData.birthday"
                type="date"
                placeholder="选择生日"
                value-format="YYYY-MM-DD"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="客户等级" prop="level">
              <el-select v-model="formData.level" placeholder="请选择客户等级" style="width: 100%">
                <el-option label="普通" :value="0" />
                <el-option label="银卡" :value="1" />
                <el-option label="金卡" :value="2" />
                <el-option label="钻石" :value="3" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="客户来源" prop="source_type">
              <el-select v-model="formData.source_type" placeholder="请选择客户来源" style="width: 100%">
                <el-option label="自然进店" :value="0" />
                <el-option label="主动邀约" :value="1" />
                <el-option label="同行带单" :value="2" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="状态">
              <el-radio-group v-model="formData.status">
                <el-radio :value="1">启用</el-radio>
                <el-radio :value="0">禁用</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="isEdit && formData.original_phone">
            <el-form-item label="原手机号">
              <el-input v-model="formData.original_phone" disabled />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="地址">
          <el-input v-model="formData.address" type="textarea" :rows="2" placeholder="请输入地址" />
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

    <!-- 跟进记录弹窗 -->
    <el-dialog
      v-model="followUpDialogVisible"
      title="跟进记录"
      width="650px"
      destroy-on-close
    >
      <div class="follow-up-header">
        <span class="customer-name">{{ currentCustomer?.customer_name }}</span>
        <span class="customer-phone">{{ currentCustomer?.phone }}</span>
      </div>

      <el-timeline v-if="followUpList.length > 0">
        <el-timeline-item
          v-for="item in followUpList"
          :key="item.id"
          :timestamp="formatTime(item.created_at)"
          placement="top"
        >
          <el-card shadow="never" class="follow-up-card">
            <div class="follow-up-content">
              <div class="follow-up-text">{{ item.content }}</div>
              <div class="follow-up-meta">
                <span>跟进人：{{ item.follower?.real_name || item.follower?.username || '-' }}</span>
                <span>方式：{{ getFollowUpMethodLabel(item.follow_type) }}</span>
                <span v-if="item.is_deal === 1" class="deal-tag">已成交</span>
              </div>
              <div v-if="item.next_follow_date" class="follow-up-next">
                下次跟进：{{ item.next_follow_date?.substring(0, 10) }}
                <span v-if="item.next_follow_content"> - {{ item.next_follow_content }}</span>
              </div>
            </div>
          </el-card>
        </el-timeline-item>
      </el-timeline>

      <el-empty v-else description="暂无跟进记录" />

      <div class="add-follow-up">
        <el-divider />
        <el-form :inline="true" :model="followUpForm">
          <el-form-item>
            <el-select v-model="followUpForm.method" placeholder="跟进方式" style="width: 120px">
              <el-option label="电话" value="phone" />
              <el-option label="微信" value="wechat" />
              <el-option label="上门" value="visit" />
              <el-option label="其他" value="other" />
            </el-select>
          </el-form-item>
          <el-form-item style="flex: 1">
            <el-input v-model="followUpForm.content" placeholder="请输入跟进内容" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="addFollowUpLoading" @click="handleAddFollowUp">
              添加
            </el-button>
          </el-form-item>
        </el-form>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getCustomerList,
  createCustomer,
  updateCustomer,
  deleteCustomer,
  getFollowUps,
  addFollowUp,
} from '@/api/customer'
import { formatCurrency } from '@/utils/format'

// ==================== 搜索与列表 ====================
const loading = ref(false)
const customerList = ref([])

const searchForm = reactive({
  keyword: '',
  level: '',
  status: '',
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
    if (searchForm.level !== '' && searchForm.level !== null) params.level = searchForm.level
    if (searchForm.status !== '' && searchForm.status !== null) params.status = searchForm.status

    const res = await getCustomerList(params)
    customerList.value = res.data?.list || res.data || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取客户列表失败:', error)
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
  searchForm.level = ''
  searchForm.status = ''
  pagination.page = 1
  fetchList()
}

// ==================== 等级映射 ====================
const getLevelLabel = (level) => {
  const map = { 0: '普通', 1: '银卡', 2: '金卡', 3: '钻石' }
  return map[level] ?? level ?? '-'
}

const getSourceTypeLabel = (type) => {
  const map = { 0: '自然进店', 1: '主动邀约', 2: '同行带单' }
  return map[type] ?? '-'
}

const getLevelTag = (level) => {
  const map = { 0: 'info', 1: '', 2: 'warning', 3: 'success' }
  return map[level] ?? 'info'
}

// ==================== 新增/编辑 ====================
const formDialogVisible = ref(false)
const submitLoading = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const editingId = ref(null)

const formData = reactive({
  customer_name: '',
  phone: '',
  original_phone: '',
  customer_code: '',
  email: '',
  gender: 0,
  birthday: '',
  level: 0,
  source_type: 0,
  status: 1,
  address: '',
  remark: '',
})

const formRules = {
  customer_name: [{ required: true, message: '请输入客户姓名', trigger: 'blur' }],
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' },
  ],
  source_type: [{ required: true, message: '请选择客户来源', trigger: 'change' }],
}

const resetFormData = () => {
  formData.customer_name = ''
  formData.phone = ''
  formData.original_phone = ''
  formData.customer_code = ''
  formData.email = ''
  formData.gender = 0
  formData.birthday = ''
  formData.level = 0
  formData.source_type = 0
  formData.status = 1
  formData.address = ''
  formData.remark = ''
}

const handleAdd = () => {
  isEdit.value = false
  editingId.value = null
  resetFormData()
  formDialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  editingId.value = row.id
  formData.customer_name = row.customer_name || ''
  formData.phone = row.phone || ''
  formData.original_phone = row.original_phone || ''
  formData.customer_code = row.customer_code || ''
  formData.email = row.email || ''
  formData.gender = row.gender ?? 0
  formData.birthday = row.birthday ? row.birthday.substring(0, 10) : ''
  formData.level = row.level ?? 0
  formData.source_type = row.source_type ?? 0
  formData.status = row.status ?? 1
  formData.address = row.address || ''
  formData.remark = row.remark || ''
  formDialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    const data = {
      customer_name: formData.customer_name,
      phone: formData.phone,
      email: formData.email,
      gender: formData.gender,
      birthday: formData.birthday || null,
      level: formData.level,
      source_type: formData.source_type,
      status: formData.status,
      address: formData.address,
      remark: formData.remark,
    }

    if (isEdit.value) {
      await updateCustomer(editingId.value, data)
      ElMessage.success('编辑成功')
    } else {
      await createCustomer(data)
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

// ==================== 启用/禁用 ====================
const handleToggleStatus = (row) => {
  const action = row.status === 1 ? '禁用' : '启用'
  ElMessageBox.confirm(
    `确定要${action}客户 "${row.customer_name}" 吗？`,
    `${action}确认`,
    {
      confirmButtonText: `确定${action}`,
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await updateCustomer(row.id, { status: row.status === 1 ? 0 : 1 })
      ElMessage.success(`${action}成功`)
      fetchList()
    } catch (error) {
      console.error(`${action}失败:`, error)
    }
  }).catch(() => {})
}

// ==================== 删除 ====================
const handleDelete = (row) => {
  ElMessageBox.confirm(
    `确定要删除客户 "${row.customer_name}" 吗？`,
    '删除确认',
    {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await deleteCustomer(row.id)
      ElMessage.success('删除成功')
      fetchList()
    } catch (error) {
      console.error('删除失败:', error)
    }
  }).catch(() => {})
}

// ==================== 跟进记录 ====================
const followUpDialogVisible = ref(false)
const followUpList = ref([])
const currentCustomer = ref(null)
const addFollowUpLoading = ref(false)
const followUpForm = reactive({
  method: 'phone',
  content: '',
})

const getFollowUpMethodLabel = (followType) => {
  const map = { 0: '其他', 1: '电话', 2: '微信', 3: '上门' }
  return map[followType] ?? followType ?? '-'
}

const formatTime = (time) => {
  if (!time) return '-'
  return time.replace('T', ' ').substring(0, 19)
}

const handleFollowUp = async (row) => {
  currentCustomer.value = row
  followUpForm.method = 'phone'
  followUpForm.content = ''
  followUpDialogVisible.value = true
  await fetchFollowUps(row.id)
}

const fetchFollowUps = async (customerId) => {
  try {
    const res = await getFollowUps(customerId)
    followUpList.value = res.data || []
  } catch (error) {
    console.error('获取跟进记录失败:', error)
  }
}

const handleAddFollowUp = async () => {
  if (!followUpForm.content.trim()) {
    ElMessage.warning('请输入跟进内容')
    return
  }

  addFollowUpLoading.value = true
  try {
    await addFollowUp(currentCustomer.value.id, {
      follow_type: followUpForm.method === 'phone' ? 1 : followUpForm.method === 'wechat' ? 2 : followUpForm.method === 'visit' ? 3 : 0,
      content: followUpForm.content,
    })
    ElMessage.success('跟进记录添加成功')
    followUpForm.content = ''
    await fetchFollowUps(currentCustomer.value.id)
  } catch (error) {
    console.error('添加跟进记录失败:', error)
  } finally {
    addFollowUpLoading.value = false
  }
}

// ==================== 初始化 ====================
onMounted(() => {
  fetchList()
})
</script>

<style lang="scss" scoped>
.customer-manage {
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

  .follow-up-header {
    display: flex;
    align-items: center;
    gap: 16px;
    margin-bottom: 20px;

    .customer-name {
      font-size: 16px;
      font-weight: 600;
      color: #303133;
    }

    .customer-phone {
      color: #909399;
    }
  }

  .follow-up-card {
    .follow-up-content {
      .follow-up-text {
        color: #303133;
        line-height: 1.6;
      }

      .follow-up-meta {
        display: flex;
        gap: 16px;
        margin-top: 8px;
        color: #909399;
        font-size: 13px;

        .deal-tag {
          color: #67c23a;
          font-weight: 500;
        }
      }

      .follow-up-next {
        margin-top: 8px;
        padding: 6px 10px;
        background: #f0f7ff;
        border-radius: 4px;
        color: #409eff;
        font-size: 13px;
      }
    }
  }

  .add-follow-up {
    margin-top: 8px;
  }
}
</style>
