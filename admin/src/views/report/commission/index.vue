<template>
  <div class="commission-report">
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
        <CardStatistic title="提成总额" :value="summary.totalCommission" suffix="元" icon="Money" trend="up" trend-value="6.8%" />
      </el-col>
      <el-col :span="12">
        <CardStatistic title="人均提成" :value="summary.avgCommission" suffix="元" icon="User" trend="up" trend-value="3.2%" />
      </el-col>
    </el-row>

    <!-- 提成分布饼图 + 月度提成趋势 -->
    <el-row :gutter="16" style="margin-top: 16px;">
      <el-col :span="10">
        <el-card shadow="hover" class="chart-card">
          <template #header>
            <div class="card-header">
              <span>提成类型分布</span>
            </div>
          </template>
          <PieChart :series-data="typePieData" height="350px" />
        </el-card>
      </el-col>
      <el-col :span="14">
        <el-card shadow="hover" class="chart-card">
          <template #header>
            <div class="card-header">
              <span>月度提成趋势</span>
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
import { getCommissionAnalysis } from '@/api/report'
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
  fetchCommissionData()
}

const handleReset = () => {
  dateRange.value = getDefaultDateRange()
  handleSearch()
}

// ==================== 统计摘要 ====================
const summary = reactive({
  totalCommission: 0,
  avgCommission: 0,
})

// ==================== 饼图 ====================
const typePieData = ref([])

// ==================== 趋势图 ====================
const trendXData = ref([])
const trendSeriesData = ref([])

const fetchCommissionData = async () => {
  loading.value = true
  try {
    const params = {}
    if (dateRange.value?.length === 2) {
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    }
    const res = await getCommissionAnalysis(params)
    if (res.data) {
      summary.totalCommission = res.data.totalCommission || 0
      summary.avgCommission = res.data.avgCommission || 0

      typePieData.value = res.data.typeDistribution || []
      trendXData.value = res.data.trendMonths || []
      trendSeriesData.value = [
        {
          name: '提成总额',
          data: res.data.trendCommissions || [],
          itemStyle: { color: '#1890ff' },
          areaStyle: { color: 'rgba(24,144,255,0.1)' },
        },
      ]
    }
  } catch (error) {
    console.error('获取提成分析失败:', error)
    initMockData()
  } finally {
    loading.value = false
  }
}

const initMockData = () => {
  summary.totalCommission = 186500
  summary.avgCommission = 6217

  // 提成类型：业务员提成/同行分成/团队分润/基金池/老带新
  typePieData.value = [
    { name: '业务员提成', value: 85000 },
    { name: '同行分成', value: 32000 },
    { name: '团队分润', value: 38000 },
    { name: '基金池', value: 21500 },
    { name: '老带新', value: 10000 },
  ]

  const months = ['01月', '02月', '03月', '04月', '05月', '06月', '07月', '08月', '09月', '10月', '11月', '12月']
  trendXData.value = months
  trendSeriesData.value = [
    {
      name: '提成总额',
      data: [12000, 13500, 11800, 15000, 16200, 14500, 17800, 15600, 18200, 16800, 19500, 17600],
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
.commission-report {
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
