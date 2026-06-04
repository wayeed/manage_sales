<template>
  <div class="commission-summary">
    <!-- 月份选择 -->
    <el-card class="filter-card" shadow="never">
      <el-form :model="filterForm" inline>
        <el-form-item label="统计月份">
          <el-date-picker
            v-model="filterForm.month"
            type="month"
            placeholder="请选择月份"
            value-format="YYYY-MM"
            style="width: 160px"
            @change="fetchSummary"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="Search" @click="fetchSummary">查询</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 汇总统计 -->
    <el-row :gutter="16" class="summary-row">
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-item">
            <div class="stat-label">销售提成合计</div>
            <div class="stat-value sales">{{ formatCurrency(summaryStats.total_sales) }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-item">
            <div class="stat-label">团队分润合计</div>
            <div class="stat-value team">{{ formatCurrency(summaryStats.total_team) }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-item">
            <div class="stat-label">基金池奖励合计</div>
            <div class="stat-value fund">{{ formatCurrency(summaryStats.total_fund) }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-item">
            <div class="stat-label">老带新奖励合计</div>
            <div class="stat-value referral">{{ formatCurrency(summaryStats.total_referral) }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 员工提成汇总表格 -->
    <el-card class="table-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">员工提成汇总</span>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="summaryList"
        border
        stripe
        show-summary
        :summary-method="getSummaryRow"
        style="width: 100%"
      >
        <el-table-column prop="employee_name" label="员工姓名" width="140" fixed="left" />
        <el-table-column label="销售提成" width="140" align="right">
          <template #default="{ row }">
            <span class="amount">{{ formatCurrency(row.sales_commission) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="团队分润" width="140" align="right">
          <template #default="{ row }">
            <span class="amount">{{ formatCurrency(row.team_share) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="基金池奖励" width="140" align="right">
          <template #default="{ row }">
            <span class="amount">{{ formatCurrency(row.fund_pool_reward) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="老带新奖励" width="140" align="right">
          <template #default="{ row }">
            <span class="amount">{{ formatCurrency(row.referral_reward) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="合计" width="160" align="right" fixed="right">
          <template #default="{ row }">
            <span class="total-amount">{{ formatCurrency(row.total) }}</span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getCommissionSummary } from '@/api/commission'
import { formatCurrency } from '@/utils/format'

// ==================== 筛选 ====================
const filterForm = reactive({
  month: '',
})

// ==================== 汇总统计 ====================
const summaryStats = reactive({
  total_sales: 0,
  total_team: 0,
  total_fund: 0,
  total_referral: 0,
})

// ==================== 列表 ====================
const loading = ref(false)
const summaryList = ref([])

const fetchSummary = async () => {
  if (!filterForm.month) {
    return
  }
  loading.value = true
  try {
    const params = { month: filterForm.month }
    const res = await getCommissionSummary(params)
    const list = res.data?.list || res.data || []
    summaryList.value = list
    // 从列表数据中计算汇总统计
    summaryStats.total_sales = list.reduce((sum, row) => sum + (Number(row.sales_commission) || 0), 0)
    summaryStats.total_team = list.reduce((sum, row) => sum + (Number(row.team_share) || 0), 0)
    summaryStats.total_fund = list.reduce((sum, row) => sum + (Number(row.fund_pool_reward) || 0), 0)
    summaryStats.total_referral = list.reduce((sum, row) => sum + (Number(row.referral_reward) || 0), 0)
  } catch (error) {
    console.error('获取提成汇总失败:', error)
  } finally {
    loading.value = false
  }
}

// ==================== 合计行 ====================
const getSummaryRow = ({ columns, data }) => {
  const sums = []
  columns.forEach((column, index) => {
    if (index === 0) {
      sums[index] = '合计'
      return
    }
    const propMap = {
      1: 'sales_commission',
      2: 'team_share',
      3: 'fund_pool_reward',
      4: 'referral_reward',
      5: 'total',
    }
    const prop = propMap[index]
    if (prop) {
      const total = data.reduce((sum, row) => sum + (Number(row[prop]) || 0), 0)
      sums[index] = formatCurrency(total)
    } else {
      sums[index] = ''
    }
  })
  return sums
}

// ==================== 初始化 ====================
onMounted(() => {
  // 默认当前月份
  const now = new Date()
  filterForm.month = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
  fetchSummary()
})
</script>

<style lang="scss" scoped>
.commission-summary {
  .filter-card {
    margin-bottom: 16px;
  }

  .summary-row {
    margin-bottom: 16px;
  }

  .stat-card {
    .stat-item {
      text-align: center;
      padding: 8px 0;
    }

    .stat-label {
      font-size: 14px;
      color: #909399;
      margin-bottom: 8px;
    }

    .stat-value {
      font-size: 20px;
      font-weight: 600;

      &.sales {
        color: #409eff;
      }
      &.team {
        color: #67c23a;
      }
      &.fund {
        color: #e6a23c;
      }
      &.referral {
        color: #f56c6c;
      }
    }
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

  .amount {
    color: #606266;
    font-weight: 500;
  }

  .total-amount {
    color: #e6a23c;
    font-weight: 700;
    font-size: 15px;
  }
}
</style>
