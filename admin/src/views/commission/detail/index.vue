<template>
  <div class="commission-detail">
    <!-- 搜索栏 -->
    <el-card class="search-card" shadow="never">
      <el-form :model="searchForm" inline>
        <el-form-item label="员工">
          <el-select
            v-model="searchForm.employee_id"
            placeholder="全部员工"
            clearable
            filterable
            style="width: 160px"
          >
            <el-option
              v-for="item in employeeOptions"
              :key="item.id"
              :label="item.real_name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="searchForm.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 240px"
          />
        </el-form-item>
        <el-form-item label="提成类型">
          <el-select
            v-model="searchForm.commission_type"
            placeholder="全部类型"
            clearable
            style="width: 140px"
          >
            <el-option label="销售提成" :value="1" />
            <el-option label="团队分润" :value="2" />
            <el-option label="基金池奖励" :value="3" />
            <el-option label="老带新奖励" :value="4" />
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
          <span class="title">提成明细</span>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="commissionList"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column label="订单号" width="180" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.order?.order_no || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="员工" width="120">
          <template #default="{ row }">
            {{ row.employee?.real_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="提成类型" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="getCommissionTypeTag(row.commission_type)" size="small">
              {{ getCommissionTypeLabel(row.commission_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="基数金额" width="130" align="right">
          <template #default="{ row }">
            <span class="price">{{ formatCurrency(row.base_amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="比例" width="100" align="center">
          <template #default="{ row }">
            {{ formatPercent(row.rate) }}
          </template>
        </el-table-column>
        <el-table-column label="提成金额" width="130" align="right">
          <template #default="{ row }">
            <span class="amount">{{ formatCurrency(row.amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusTag(row.status)" size="small">
              {{ getStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" width="180" />
        <el-table-column label="操作" width="120" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="warning" link size="small" @click="handleAdjust(row)">
              手工调整
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

    <!-- 手工调整弹窗 -->
    
<el-dialog v-dialog-drag
      v-model="adjustDialogVisible"
      title="手工调整提成"
      width="480px"
      destroy-on-close
    >
      <el-form
        ref="adjustFormRef"
        :model="adjustForm"
        :rules="adjustFormRules"
        label-width="100px"
      >
        <el-form-item label="订单号">
          <el-input :model-value="currentCommission?.order?.order_no" disabled />
        </el-form-item>
        <el-form-item label="员工">
          <el-input :model-value="currentCommission?.employee?.real_name" disabled />
        </el-form-item>
        <el-form-item label="原提成金额">
          <el-input :model-value="formatCurrency(currentCommission?.amount)" disabled />
        </el-form-item>
        <el-form-item label="调整后金额" prop="new_amount">
          <el-input-number
            v-model="adjustForm.new_amount"
            :min="0"
            :precision="2"
            controls-position="right"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="调整原因" prop="reason">
          <el-input
            v-model="adjustForm.reason"
            type="textarea"
            :rows="3"
            placeholder="请输入调整原因"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="adjustDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleAdjustSubmit">
          确定调整
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getCommissionList, manualAdjust } from '@/api/commission'
import { formatCurrency, formatPercent } from '@/utils/format'

// ==================== 搜索与列表 ====================
const loading = ref(false)
const commissionList = ref([])
const employeeOptions = ref([])

const searchForm = reactive({
  employee_id: '',
  dateRange: null,
  commission_type: '',
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
      ...getSearchParams(),
    }
    const res = await getCommissionList(params)
    commissionList.value = res.data?.list || res.data || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取提成列表失败:', error)
  } finally {
    loading.value = false
  }
}

const getSearchParams = () => {
  const params = {}
  if (searchForm.employee_id) params.employee_id = searchForm.employee_id
  if (searchForm.commission_type) params.commission_type = searchForm.commission_type
  if (searchForm.dateRange && searchForm.dateRange.length === 2) {
    params.start_date = searchForm.dateRange[0]
    params.end_date = searchForm.dateRange[1]
  }
  return params
}

const handleSearch = () => {
  pagination.page = 1
  fetchList()
}

const handleReset = () => {
  searchForm.employee_id = ''
  searchForm.dateRange = null
  searchForm.commission_type = ''
  pagination.page = 1
  fetchList()
}

// ==================== 状态映射 ====================
const getCommissionTypeLabel = (type) => {
  const map = {
    1: '销售提成',
    2: '团队分润',
    3: '基金池奖励',
    4: '老带新奖励',
  }
  return map[type] || type || '未知'
}

const getCommissionTypeTag = (type) => {
  const map = {
    1: '',
    2: 'success',
    3: 'warning',
    4: 'danger',
  }
  return map[type] || 'info'
}

const getStatusLabel = (status) => {
  const map = {
    1: '待结算',
    2: '已结算',
    3: '已发放',
    4: '已调整',
    5: '已取消',
  }
  return map[status] || status || '未知'
}

const getStatusTag = (status) => {
  const map = {
    1: 'warning',
    2: 'success',
    3: '',
    4: 'info',
    5: 'danger',
  }
  return map[status] || 'info'
}

// ==================== 手工调整 ====================
const adjustDialogVisible = ref(false)
const submitLoading = ref(false)
const adjustFormRef = ref(null)
const currentCommission = ref(null)

const adjustForm = reactive({
  new_amount: 0,
  reason: '',
})

const adjustFormRules = {
  new_amount: [{ required: true, message: '请输入调整后金额', trigger: 'blur' }],
  reason: [{ required: true, message: '请输入调整原因', trigger: 'blur' }],
}

const handleAdjust = (row) => {
  currentCommission.value = row
  adjustForm.new_amount = row.amount || 0
  adjustForm.reason = ''
  adjustDialogVisible.value = true
}

const handleAdjustSubmit = async () => {
  const valid = await adjustFormRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    await manualAdjust(currentCommission.value.id, {
      new_amount: adjustForm.new_amount,
      reason: adjustForm.reason,
    })
    ElMessage.success('调整成功')
    adjustDialogVisible.value = false
    fetchList()
  } catch (error) {
    console.error('调整提成失败:', error)
  } finally {
    submitLoading.value = false
  }
}

// ==================== 初始化 ====================
onMounted(() => {
  fetchList()
})
</script>

<style lang="scss" scoped>
.commission-detail {
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

  .amount {
    color: #e6a23c;
    font-weight: 600;
  }
}
</style>
