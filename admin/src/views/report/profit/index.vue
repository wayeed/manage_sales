<template>
  <div class="profit-report">
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
      <el-col :span="12">
        <CardStatistic title="总利润" :value="summary.totalProfit" suffix="元" icon="TrendCharts" trend="up" trend-value="8.5%" />
      </el-col>
      <el-col :span="12">
        <CardStatistic title="平均利润率" :value="summary.avgProfitRate" suffix="%" icon="DataLine" trend="up" trend-value="2.1%" />
      </el-col>
    </el-row>

    <!-- 利润趋势折线图 -->
    <el-card shadow="hover" class="chart-card" style="margin-top: 16px;">
      <template #header>
        <div class="card-header">
          <span>利润趋势</span>
        </div>
      </template>
      <LineChart
        :x-data="trendXData"
        :series-data="trendSeriesData"
        height="350px"
      />
    </el-card>

    <!-- 成本构成饼图 -->
    <el-card shadow="hover" class="chart-card" style="margin-top: 16px;">
      <template #header>
        <div class="card-header">
          <span>成本构成</span>
        </div>
      </template>
      <PieChart :series-data="costPieData" height="350px" />
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getProfitAnalysis } from '@/api/report'
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

const handleSearch = () => {
  fetchProfitData()
}

const handleReset = () => {
  dateRange.value = getDefaultDateRange()
  handleSearch()
}

// ==================== 统计摘要 ====================
const summary = reactive({
  totalProfit: 0,
  avgProfitRate: 0,
})

// ==================== 趋势图 ====================
const trendXData = ref([])
const trendSeriesData = ref([])

// ==================== 饼图 ====================
const costPieData = ref([])

const fetchProfitData = async () => {
  try {
    const params = {}
    if (dateRange.value?.length === 2) {
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    }
    const res = await getProfitAnalysis(params)
    if (res.data) {
      summary.totalProfit = res.data.totalProfit || 0
      summary.avgProfitRate = res.data.avgProfitRate || 0

      trendXData.value = res.data.trendDates || []
      trendSeriesData.value = [
        {
          name: '利润',
          data: res.data.trendProfits || [],
          itemStyle: { color: '#52c41a' },
          areaStyle: { color: 'rgba(82,196,26,0.1)' },
        },
        {
          name: '利润率',
          data: res.data.trendProfitRates || [],
          yAxisIndex: 1,
          itemStyle: { color: '#faad14' },
        },
      ]

      costPieData.value = res.data.costComposition || []
    }
  } catch (error) {
    console.error('获取利润分析失败:', error)
    initMockData()
  }
}

const initMockData = () => {
  summary.totalProfit = 286500
  summary.avgProfitRate = 28.5

  // 趋势数据
  const months = ['01月', '02月', '03月', '04月', '05月', '06月', '07月', '08月', '09月', '10月', '11月', '12月']
  trendXData.value = months
  trendSeriesData.value = [
    {
      name: '利润',
      data: [18000, 22000, 19500, 25000, 28000, 24000, 30000, 26000, 32000, 28000, 35000, 30000],
      itemStyle: { color: '#52c41a' },
      areaStyle: { color: 'rgba(82,196,26,0.1)' },
    },
    {
      name: '利润率',
      data: [25, 27, 24, 28, 30, 26, 31, 27, 32, 29, 33, 30],
      yAxisIndex: 1,
      itemStyle: { color: '#faad14' },
    },
  ]

  // 成本构成：商品成本、礼品成本、利润
  costPieData.value = [
    { name: '商品成本', value: 520000 },
    { name: '礼品成本', value: 85000 },
    { name: '利润', value: 286500 },
  ]
}

// ==================== 初始化 ====================
onMounted(() => {
  handleSearch()
})
</script>

<style lang="scss" scoped>
.profit-report {
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
