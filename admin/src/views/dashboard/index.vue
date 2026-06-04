<template>
  <div class="dashboard-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16">
      <el-col :span="6">
        <CardStatistic
          title="本月销售额"
          :value="overview.total_sales"
          suffix="元"
          icon="Money"
          trend="up"
          trend-value="12.5%"
        />
      </el-col>
      <el-col :span="6">
        <CardStatistic
          title="本月订单数"
          :value="overview.total_orders"
          suffix="单"
          icon="Document"
          trend="up"
          trend-value="8.3%"
        />
      </el-col>
      <el-col :span="6">
        <CardStatistic
          title="本月利润"
          :value="overview.total_profit"
          suffix="元"
          icon="TrendCharts"
          trend="up"
          trend-value="5.2%"
        />
      </el-col>
      <el-col :span="6">
        <CardStatistic
          title="待审批订单"
          :value="overview.pending_orders"
          suffix="单"
          icon="Bell"
          trend="down"
          trend-value="3.1%"
        />
      </el-col>
    </el-row>

    <!-- 中间：销售趋势折线图 -->
    <el-row :gutter="16" style="margin-top: 16px;">
      <el-col :span="24">
        <el-card shadow="hover" class="chart-card">
          <template #header>
            <div class="card-header">
              <span>销售趋势（近30天）</span>
            </div>
          </template>
          <LineChart
            :x-data="trendXData"
            :series-data="trendSeriesData"
            height="320px"
          />
        </el-card>
      </el-col>
    </el-row>

    <!-- 下方左：待办事项 / 下方右：业绩排行 -->
    <el-row :gutter="16" style="margin-top: 16px;">
      <el-col :span="10">
        <el-card shadow="hover" class="chart-card">
          <template #header>
            <div class="card-header">
              <span>待办事项</span>
              <el-badge :value="todoTotal" type="danger" />
            </div>
          </template>
          <el-timeline>
            <el-timeline-item
              v-for="item in todoList"
              :key="item.id"
              :type="item.timelineType"
              :hollow="item.hollow"
              :timestamp="item.timestamp"
              placement="top"
            >
              <div class="todo-item" @click="handleTodoClick(item)">
                <div class="todo-header">
                  <el-tag :type="item.tagType" size="small" effect="plain">{{ item.tag }}</el-tag>
                  <span class="todo-text">{{ item.text }}</span>
                </div>
                <div class="todo-count">
                  <el-badge :value="item.count" :type="item.badgeType" />
                </div>
              </div>
            </el-timeline-item>
          </el-timeline>
        </el-card>
      </el-col>
      <el-col :span="14">
        <el-card shadow="hover" class="chart-card">
          <template #header>
            <div class="card-header">
              <span>业绩排行 TOP5</span>
            </div>
          </template>
          <BarChart
            :x-data="rankingXData"
            :series-data="rankingSeriesData"
            height="320px"
          />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getDashboardOverview } from '@/api/dashboard'
import { getSalesTrend, getPerformanceRanking } from '@/api/report'
import CardStatistic from '@/components/Charts/CardStatistic.vue'
import LineChart from '@/components/Charts/LineChart.vue'
import BarChart from '@/components/Charts/BarChart.vue'

const router = useRouter()

// ==================== 概览数据 ====================
const overview = ref({
  total_sales: 0,
  total_orders: 0,
  total_profit: 0,
  pending_orders: 0,
  pending_payments: 0,
  low_stock_count: 0,
})

const fetchOverview = async () => {
  try {
    const res = await getDashboardOverview()
    if (res.data) {
      overview.value = res.data
    }
  } catch (error) {
    console.error('获取仪表盘概览失败:', error)
  }
}

// ==================== 销售趋势（近30天） ====================
const trendXData = ref([])
const trendSeriesData = ref([])

const fetchTrendData = async () => {
  try {
    // 计算近30天日期范围
    const endDate = new Date()
    const startDate = new Date()
    startDate.setDate(startDate.getDate() - 29)
    
    const formatDate = (date) => {
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      return `${year}-${month}-${day}`
    }
    
    const res = await getSalesTrend({
      start_date: formatDate(startDate),
      end_date: formatDate(endDate),
      dimension: 'day'
    })
    
    if (res.data && Array.isArray(res.data)) {
      const list = res.data
      trendXData.value = list.map(item => {
        const date = new Date(item.date)
        return `${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
      })
      trendSeriesData.value = [
        {
          name: '销售额',
          data: list.map(item => parseFloat(item.sales_amount) || 0),
          itemStyle: { color: '#1890ff' },
          areaStyle: { color: 'rgba(24,144,255,0.1)' },
        },
        {
          name: '订单数',
          data: list.map(item => item.order_count || 0),
          yAxisIndex: 1,
          itemStyle: { color: '#52c41a' },
        },
      ]
    }
  } catch (error) {
    console.error('获取销售趋势失败:', error)
  }
}

// ==================== 待办事项 ====================
const todoList = computed(() => [
  {
    id: 1,
    tag: '待审批',
    tagType: 'warning',
    text: '待审批订单',
    count: overview.value.pending_orders || 0,
    badgeType: 'warning',
    timelineType: 'warning',
    hollow: false,
    timestamp: '最新',
    route: '/order/index',
  },
  {
    id: 2,
    tag: '待审核',
    tagType: 'danger',
    text: '待审核回款',
    count: overview.value.pending_payments || 0,
    badgeType: 'danger',
    timelineType: 'danger',
    hollow: false,
    timestamp: '今天',
    route: '/report/payment',
  },
  {
    id: 3,
    tag: '预警',
    tagType: 'info',
    text: '库存预警',
    count: overview.value.low_stock_count || 0,
    badgeType: 'info',
    timelineType: 'primary',
    hollow: false,
    timestamp: '今天',
    route: '/inventory/alert',
  },
])

const todoTotal = computed(() => {
  return todoList.value.reduce((sum, item) => sum + item.count, 0)
})

const handleTodoClick = (item) => {
  if (item.route) {
    router.push(item.route)
  }
}

// ==================== 业绩排行 TOP5 ====================
const rankingXData = ref([])
const rankingSeriesData = ref([])

const fetchRankingData = async () => {
  try {
    const now = new Date()
    const res = await getPerformanceRanking({
      period_type: 'month',
      period_value: `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`,
      rank_by: 'salesman',
      limit: 5
    })
    
    if (res.data && Array.isArray(res.data)) {
      const list = res.data.slice(0, 5)
      rankingXData.value = list.map(item => item.name)
      rankingSeriesData.value = [
        {
          name: '销售额',
          data: list.map(item => parseFloat(item.sales_amount) || 0),
          itemStyle: {
            color: (params) => {
              const colors = ['#f5222d', '#fa8c16', '#1890ff', '#52c41a', '#722ed1']
              return colors[params.dataIndex] || '#1890ff'
            },
          },
          label: {
            show: true,
            position: 'top',
            formatter: (params) => `${(params.value / 10000).toFixed(1)}万`,
            fontSize: 12,
            color: '#606266',
          },
        },
      ]
    }
  } catch (error) {
    console.error('获取业绩排行失败:', error)
  }
}

// ==================== 初始化 ====================
onMounted(() => {
  fetchOverview()
  fetchTrendData()
  fetchRankingData()
})
</script>

<style lang="scss" scoped>
.dashboard-container {
  width: 100%;
}

.chart-card {
  border-radius: 8px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.todo-item {
  cursor: pointer;
  transition: background-color 0.2s;
  padding: 4px 0;

  &:hover {
    .todo-header {
      color: #1890ff;
    }
  }

  .todo-header {
    display: flex;
    align-items: center;
    gap: 8px;
    transition: color 0.2s;
  }

  .todo-text {
    font-size: 14px;
    color: #303133;
    font-weight: 500;
  }

  .todo-count {
    margin-top: 4px;
  }
}
</style>
