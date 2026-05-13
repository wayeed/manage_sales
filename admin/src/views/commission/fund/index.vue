<template>
  <div class="fund-pool-manage">
    <!-- 基金池列表 -->
    <el-card class="table-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">基金池管理</span>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="fundPoolList"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column label="周期" width="140">
          <template #default="{ row }">
            {{ row.period_value }}
          </template>
        </el-table-column>
        <el-table-column label="类型" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="row.period_type === 1 ? '' : 'success'" size="small">
              {{ row.period_type === 1 ? '月度' : '季度' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="总利润" width="140" align="right">
          <template #default="{ row }">
            <span class="price">{{ formatCurrency(row.total_profit) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="提取比例" width="110" align="center">
          <template #default="{ row }">
            {{ formatPercent(row.extract_rate) }}
          </template>
        </el-table-column>
        <el-table-column label="基金池金额" width="150" align="right">
          <template #default="{ row }">
            <span class="fund-amount">{{ formatCurrency(row.pool_amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="份数" width="90" align="center">
          <template #default="{ row }">
            {{ row.total_shares }}
          </template>
        </el-table-column>
        <el-table-column label="每份金额" width="130" align="right">
          <template #default="{ row }">
            <span class="amount">{{ formatCurrency(row.per_share_amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getFundStatusTag(row.status)" size="small">
              {{ getFundStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleViewShares(row)">
              份额详情
            </el-button>
            <el-button
              v-if="row.status === 1"
              type="success" link size="small" @click="handleSettle(row)"
            >
              结算
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

    <!-- 份额详情弹窗 -->
    
<el-dialog v-dialog-drag
      v-model="sharesDialogVisible"
      title="份额详情"
      width="700px"
      destroy-on-close
    >
      <div class="shares-info">
        <span>周期：</span>
        <strong>{{ currentFund?.period_value }}</strong>
        <el-divider direction="vertical" />
        <span>基金池金额：</span>
        <strong>{{ formatCurrency(currentFund?.pool_amount) }}</strong>
        <el-divider direction="vertical" />
        <span>份数：</span>
        <strong>{{ currentFund?.total_shares }}</strong>
        <el-divider direction="vertical" />
        <span>每份金额：</span>
        <strong>{{ formatCurrency(currentFund?.per_share_amount) }}</strong>
      </div>
      <el-divider />
      <el-table
        v-loading="sharesLoading"
        :data="sharesList"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column prop="employee_name" label="姓名" width="140" />
        <el-table-column label="个人利润" width="160" align="right">
          <template #default="{ row }">
            <span class="price">{{ formatCurrency(row.personal_profit) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="shares" label="份数" width="100" align="center" />
        <el-table-column label="奖励金额" width="160" align="right">
          <template #default="{ row }">
            <span class="amount">{{ formatCurrency(row.reward_amount) }}</span>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 结算弹窗 -->
    
<el-dialog v-dialog-drag
      v-model="settleDialogVisible"
      title="基金池结算"
      width="480px"
      destroy-on-close
    >
      <el-form
        ref="settleFormRef"
        :model="settleForm"
        :rules="settleFormRules"
        label-width="100px"
      >
        <el-form-item label="周期">
          <el-input :model-value="currentFund?.period_value" disabled />
        </el-form-item>
        <el-form-item label="基金池金额">
          <el-input :model-value="formatCurrency(currentFund?.pool_amount)" disabled />
        </el-form-item>
        <el-form-item label="结算备注" prop="remark">
          <el-input
            v-model="settleForm.remark"
            type="textarea"
            :rows="3"
            placeholder="请输入结算备注"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="settleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSettleSubmit">
          确认结算
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getFundPoolList, getFundPoolShares, settleFundPool } from '@/api/fund'
import { formatCurrency, formatPercent } from '@/utils/format'

// ==================== 列表 ====================
const loading = ref(false)
const fundPoolList = ref([])

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
    const res = await getFundPoolList(params)
    fundPoolList.value = res.data?.list || res.data || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取基金池列表失败:', error)
  } finally {
    loading.value = false
  }
}

// ==================== 状态映射 ====================
const getFundStatusLabel = (status) => {
  const map = {
    1: '待结算',
    2: '已结算',
    3: '已取消',
  }
  return map[status] || status || '未知'
}

const getFundStatusTag = (status) => {
  const map = {
    1: 'warning',
    2: 'success',
    3: 'info',
  }
  return map[status] || 'info'
}

// ==================== 份额详情 ====================
const sharesDialogVisible = ref(false)
const sharesLoading = ref(false)
const sharesList = ref([])
const currentFund = ref(null)

const handleViewShares = async (row) => {
  currentFund.value = row
  sharesDialogVisible.value = true
  sharesLoading.value = true
  try {
    const res = await getFundPoolShares(row.id)
    sharesList.value = res.data || []
  } catch (error) {
    console.error('获取份额详情失败:', error)
  } finally {
    sharesLoading.value = false
  }
}

// ==================== 结算 ====================
const settleDialogVisible = ref(false)
const submitLoading = ref(false)
const settleFormRef = ref(null)

const settleForm = reactive({
  remark: '',
})

const settleFormRules = {
  remark: [{ required: true, message: '请输入结算备注', trigger: 'blur' }],
}

const handleSettle = (row) => {
  currentFund.value = row
  settleForm.remark = ''
  settleDialogVisible.value = true
}

const handleSettleSubmit = async () => {
  const valid = await settleFormRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    await settleFundPool({
      fund_pool_id: currentFund.value.id,
      remark: settleForm.remark,
    })
    ElMessage.success('结算成功')
    settleDialogVisible.value = false
    fetchList()
  } catch (error) {
    console.error('结算失败:', error)
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
.fund-pool-manage {
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

  .fund-amount {
    color: #e6a23c;
    font-weight: 600;
  }

  .amount {
    color: #67c23a;
    font-weight: 600;
  }

  .shares-info {
    font-size: 14px;
    color: #606266;

    strong {
      color: #303133;
    }
  }
}
</style>
