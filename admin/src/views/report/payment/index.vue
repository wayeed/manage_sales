<template>
  <div class="payment-report">
    <!-- 筛选栏 -->
    <el-card shadow="never" class="filter-card">
      <el-row :gutter="16" align="middle">
        <el-col :span="10">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 100%"
            @change="handleSearch"
          />
        </el-col>
        <el-col :span="4">
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-col>
      </el-row>
    </el-card>

    <!-- 统计卡片 -->
    <el-row :gutter="16" style="margin-top: 16px;">
      <el-col :span="8">
        <CardStatistic title="回款率" :value="summary.paymentRate" suffix="%" icon="CircleCheck" trend="up" trend-value="3.2%" />
      </el-col>
      <el-col :span="8">
        <CardStatistic title="回款总额" :value="summary.totalPayment" suffix="元" icon="Money" trend="up" trend-value="10.5%" />
      </el-col>
      <el-col :span="8">
        <CardStatistic title="未回款总额" :value="summary.unpaidAmount" suffix="元" icon="Warning" trend="down" trend-value="5.8%" />
      </el-col>
    </el-row>

    <!-- 回款方式分布饼图 + 月度回款趋势 -->
    <el-row :gutter="16" style="margin-top: 16px;">
      <el-col :span="10">
        <el-card shadow="hover" class="chart-card">
          <template #header>
            <div class="card-header">
              <span>回款方式分布</span>
            </div>
          </template>
          <PieChart :series-data="paymentMethodPieData" height="350px" />
        </el-card>
      </el-col>
      <el-col :span="14">
        <el-card shadow="hover" class="chart-card">
          <template #header>
            <div class="card-header">
              <span>月度回款趋势</span>
            </div>
          </template>
          <LineChart
            :x-data="trendXData"
            :series-data="trendSeriesData"
            height="350px"
          />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getPaymentAnalysis } from '@/api/report'
import CardStatistic from '@/components/Charts/CardStatistic.vue'
import LineChart from '@/components/Charts/LineChart.vue'
import PieChart from '@/components/Charts/PieChart.vue'

// ==================== 筛选条件 ====================
// 初始化默认日期范围为最近30天
const getDefaultDateRange = () => {
  const end = new Date()
  const start = new Date()
  start.setDate(start.getDate() - 30)
  return [
    start.toISOString().split('T')[0],
    end.toISOString().split('T')[0],
  ]
}
const dateRange = ref(getDefaultDateRange())
const loading = ref(false)

const handleSearch = () => {
  fetchPaymentData()
}

const handleReset = () => {
  dateRange.value = getDefaultDateRange()
  handleSearch()
}

// ==================== 统计摘要 ====================
const summary = reactive({
  paymentRate: 0,
  totalPayment: 0,
  unpaidAmount: 0,
})

// ==================== 饼图 ====================
const paymentMethodPieData = ref([])

// ==================== 趋势图 ====================
const trendXData = ref([])
const trendSeriesData = ref([])

const fetchPaymentData = async () => {
  loading.value = true
  try {
    const params = {}
    if (dateRange.value?.length === 2) {
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    }
    const res = await getPaymentAnalysis(params)
    if (res.data) {
      summary.paymentRate = res.data.paymentRate || 0
      summary.totalPayment = res.data.totalPayment || 0
      summary.unpaidAmount = res.data.unpaidAmount || 0

      paymentMethodPieData.value = res.data.methodDistribution || []
      trendXData.value = res.data.trendMonths || []
      trendSeriesData.value = [
        {
          name: '回款金额',
          data: res.data.trendPayments || [],
          itemStyle: { color: '#1890ff' },
          areaStyle: { color: 'rgba(24,144,255,0.1)' },
        },
      ]
    }
  } catch (error) {
    console.error('获取回款分析失败:', error)
    initMockData()
  } finally {
    loading.value = false
  }
}

const initMockData = () => {
  summary.paymentRate = 85.6
  summary.totalPayment = 1286000
  summary.unpaidAmount = 216000

  paymentMethodPieData.value = [
    { name: '银行转账', value: 520000 },
    { name: '微信支付', value: 310000 },
    { name: '支付宝', value: 256000 },
    { name: '现金', value: 120000 },
    { name: '刷卡', value: 80000 },
  ]

  const months = ['01月', '02月', '03月', '04月', '05月', '06月', '07月', '08月', '09月', '10月', '11月', '12月']
  trendXData.value = months
  trendSeriesData.value = [
    {
      name: '回款金额',
      data: [85000, 92000, 88000, 105000, 112000, 98000, 120000, 108000, 125000, 115000, 130000, 108000],
      itemStyle: { color: '#1890ff' },
      areaStyle: { color: 'rgba(24,144,255,0.1)' },
    },
  ]
}

// ==================== 初始化 ====================
onMounted(() => {
  handleSearch()
})
</script>

<style lang="scss" scoped>
.payment-report {
  .filter-card {
    border-radius: 8px;

    :deep(.el-card__body) {
      padding: 16px 20px;
    }
  }

  .chart-card {
    border-radius: 8px;
  }

  .card-header {
    font-size: 16px;
    font-weight: 600;
    color: #303133;
  }
}
</style>
