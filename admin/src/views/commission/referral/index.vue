<template>
  <div class="referral-manage">
    <!-- 引荐关系列表 -->
    <el-card class="table-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">老带新管理</span>
          <el-button type="primary" icon="Plus" @click="handleAdd">新增引荐关系</el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="referralList"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column label="引荐人" width="140">
          <template #default="{ row }">
            {{ row.referrer?.real_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="被引荐人" width="140">
          <template #default="{ row }">
            {{ row.referred?.real_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getReferralStatusTag(row.status)" size="small">
              {{ getReferralStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="奖励比例" width="120" align="center">
          <template #default="{ row }">
            {{ formatPercent(row.reward_rate) }}
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="200" show-overflow-tooltip />
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="160" align="center" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 1"
              type="danger" link size="small" @click="handleTerminate(row)"
            >
              终止关系
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

    <!-- 新增引荐关系弹窗 -->
    
<el-dialog v-dialog-drag
      v-model="formDialogVisible"
      title="新增引荐关系"
      width="520px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item label="引荐人" prop="referrer_id">
          <el-select
            v-model="formData.referrer_id"
            placeholder="请选择引荐人"
            filterable
            style="width: 100%"
          >
            <el-option
              v-for="item in employeeOptions"
              :key="item.id"
              :label="item.real_name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="被引荐人" prop="referred_id">
          <el-select
            v-model="formData.referred_id"
            placeholder="请选择被引荐人"
            filterable
            style="width: 100%"
          >
            <el-option
              v-for="item in employeeOptions"
              :key="item.id"
              :label="item.real_name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="奖励比例" prop="reward_rate">
          <el-input-number
            v-model="formData.reward_rate"
            :min="0"
            :max="100"
            :precision="1"
            :step="0.5"
            controls-position="right"
            style="width: 100%"
          />
          <span style="margin-left: 8px; color: #909399">%</span>
        </el-form-item>
        <el-form-item label="备注">
          <el-input
            v-model="formData.remark"
            type="textarea"
            :rows="3"
            placeholder="请输入备注"
          />
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
import { get, post } from '@/api/request'
import { formatPercent } from '@/utils/format'

// ==================== 列表 ====================
const loading = ref(false)
const referralList = ref([])
const employeeOptions = ref([])

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
    const res = await get('/referrals', params)
    referralList.value = res.data?.list || res.data || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取引荐关系列表失败:', error)
  } finally {
    loading.value = false
  }
}

const fetchEmployees = async () => {
  try {
    const res = await get('/employees', { page_size: 999 })
    employeeOptions.value = res.data?.list || res.data || []
  } catch (error) {
    console.error('获取员工列表失败:', error)
  }
}

// ==================== 状态映射 ====================
const getReferralStatusLabel = (status) => {
  const map = {
    1: '生效中',
    2: '已终止',
  }
  return map[status] || status || '未知'
}

const getReferralStatusTag = (status) => {
  const map = {
    1: 'success',
    2: 'info',
  }
  return map[status] || 'info'
}

// ==================== 新增引荐关系 ====================
const formDialogVisible = ref(false)
const submitLoading = ref(false)
const formRef = ref(null)

const formData = reactive({
  referrer_id: '',
  referred_id: '',
  reward_rate: 0,
  remark: '',
})

const formRules = {
  referrer_id: [{ required: true, message: '请选择引荐人', trigger: 'change' }],
  referred_id: [{ required: true, message: '请选择被引荐人', trigger: 'change' }],
  reward_rate: [{ required: true, message: '请输入奖励比例', trigger: 'blur' }],
}

const handleAdd = () => {
  formData.referrer_id = ''
  formData.referred_id = ''
  formData.reward_rate = 0
  formData.remark = ''
  formDialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  if (formData.referrer_id === formData.referred_id) {
    ElMessage.warning('引荐人和被引荐人不能相同')
    return
  }

  submitLoading.value = true
  try {
    await post('/referrals', {
      referrer_id: formData.referrer_id,
      referred_id: formData.referred_id,
      reward_rate: formData.reward_rate,
      remark: formData.remark,
    })
    ElMessage.success('新增成功')
    formDialogVisible.value = false
    fetchList()
  } catch (error) {
    console.error('新增引荐关系失败:', error)
  } finally {
    submitLoading.value = false
  }
}

// ==================== 终止关系 ====================
const handleTerminate = (row) => {
  ElMessageBox.confirm(
    `确定要终止 "${row.referrer?.real_name}" 与 "${row.referred?.real_name}" 的引荐关系吗？`,
    '终止确认',
    {
      confirmButtonText: '确定终止',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await post(`/referrals/${row.id}/terminate`)
      ElMessage.success('终止成功')
      fetchList()
    } catch (error) {
      console.error('终止引荐关系失败:', error)
    }
  }).catch(() => {})
}

// ==================== 初始化 ====================
onMounted(() => {
  fetchList()
  fetchEmployees()
})
</script>

<style lang="scss" scoped>
.referral-manage {
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
}
</style>
