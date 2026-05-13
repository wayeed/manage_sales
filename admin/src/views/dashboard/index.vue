<template>
  <div class="dashboard-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16">
      <el-col :span="6">
        <CardStatistic
          title="本月销售额"
          :value="overview.monthSalesAmount"
          suffix="元"
          icon="Money"
          trend="up"
          trend-value="12.5%"
        />
      </el-col>
      <el-col :span="6">
        <CardStatistic
          title="本月订单数"
          :value="overview.monthOrderCount"
          suffix="单"
          icon="Document"
          trend="up"
          trend-value="8.3%"
        />
      </el-col>
      <el-col :span="6">
        <CardStatistic
          title="本月利润"
          :value="overview.monthProfit"
          suffix="元"
          icon="TrendCharts"
          trend="up"
          trend-value="5.2%"
        />
      </el-col>
      <el-col :span="6">
        <CardStatistic
          title="待审批订单"
          :value="overview.pendingOrders"
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
import CardStatistic from '@/components/Charts/CardStatistic.vue'
import LineChart from '@/components/Charts/LineChart.vue'
import BarChart from '@/components/Charts/BarChart.vue'

const router = useRouter()

// ==================== 概览数据 ====================
const overview = ref({
  monthSalesAmount: 0,
  monthOrderCount: 0,
  monthProfit: 0,
  pendingOrders: 0,
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

const initTrendData = () => {
  const days = []
  const salesData = []
  const orderData = []
  const now = new Date()
  for (let i = 29; i >= 0; i--) {
    const d = new Date(now)
    d.setDate(d.getDate() - i)
    const month = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    days.push(`${month}-${day}`)
    salesData.push(Math.floor(Math.random() * 50000 + 10000))
    orderData.push(Math.floor(Math.random() * 30 + 5))
  }
  trendXData.value = days
  trendSeriesData.value = [
    {
      name: '销售额',
      data: salesData,
      itemStyle: { color: '#1890ff' },
      areaStyle: { color: 'rgba(24,144,255,0.1)' },
    },
    {
      name: '订单数',
      data: orderData,
      yAxisIndex: 1,
      itemStyle: { color: '#52c41a' },
    },
  ]
}

// ==================== 待办事项 ====================
const todoList = ref([
  {
    id: 1,
    tag: '待审批',
    tagType: 'warning',
    text: '待审批订单',
    count: 12,
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
    count: 5,
    badgeType: 'danger',
    timelineType: 'danger',
    hollow: false,
    timestamp: '今天 14:30',
    route: '/report/payment',
  },
  {
    id: 3,
    tag: '预警',
    tagType: 'info',
    text: '库存预警',
    count: 8,
    badgeType: 'info',
    timelineType: 'primary',
    hollow: false,
    timestamp: '今天 10:00',
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

const initRankingData = () => {
  const names = ['张三', '李四', '王五', '赵六', '钱七']
  rankingXData.value = names
  rankingSeriesData.value = [
    {
      name: '销售额',
      data: [128000, 96000, 85000, 72000, 65000],
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

// ==================== 初始化 ====================
onMounted(() => {
  fetchOverview()
  initTrendData()
  initRankingData()
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
