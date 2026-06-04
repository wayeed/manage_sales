<template>
  <div class="sales-report">
    <!-- 筛选栏 -->
    <el-card shadow="never" class="filter-card">
      <el-row :gutter="16" align="middle">
        <el-col :span="8">
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
        <el-col :span="6">
          <el-radio-group v-model="dimension" @change="handleSearch">
            <el-radio-button value="day">按日</el-radio-button>
            <el-radio-button value="week">按周</el-radio-button>
            <el-radio-button value="month">按月</el-radio-button>
          </el-radio-group>
        </el-col>
        <el-col :span="4">
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-col>
      </el-row>
    </el-card>

    <!-- 统计卡片 -->
    <el-row :gutter="16" style="margin-top: 16px;">
      <el-col :span="6">
        <CardStatistic title="总销售额" :value="summary.totalSales" suffix="元" icon="Money" />
      </el-col>
      <el-col :span="6">
        <CardStatistic title="订单数" :value="summary.orderCount" suffix="单" icon="Document" />
      </el-col>
      <el-col :span="6">
        <CardStatistic title="客单价" :value="summary.avgPrice" suffix="元" icon="Coin" :decimals="2" />
      </el-col>
      <el-col :span="6">
        <CardStatistic title="利润率" :value="summary.profitRate" suffix="%" icon="TrendCharts" :decimals="1" />
      </el-col>
    </el-row>

    <!-- 销售趋势折线图 -->
    <el-card shadow="hover" class="chart-card" style="margin-top: 16px;">
      <template #header>
        <div class="card-header">
          <span>销售趋势</span>
        </div>
      </template>
      <LineChart
        :x-data="trendXData"
        :series-data="trendSeriesData"
        height="350px"
      />
    </el-card>

    <!-- 销售明细表格 -->
    <el-card shadow="hover" class="chart-card" style="margin-top: 16px;">
      <template #header>
        <div class="card-header">
          <span>销售明细</span>
        </div>
      </template>
      <el-table :data="tableData" border stripe v-loading="loading" style="width: 100%">
        <el-table-column prop="date" label="日期" width="140" />
        <el-table-column prop="orderCount" label="订单数" width="100" align="center" />
        <el-table-column prop="salesAmount" label="销售额" width="140" align="right">
          <template #default="{ row }">
            {{ formatCurrency(row.salesAmount) }}
          </template>
        </el-table-column>
        <el-table-column prop="costAmount" label="成本" width="140" align="right">
          <template #default="{ row }">
            {{ formatCurrency(row.costAmount) }}
          </template>
        </el-table-column>
        <el-table-column prop="profit" label="利润" width="140" align="right">
          <template #default="{ row }">
            {{ formatCurrency(row.profit) }}
          </template>
        </el-table-column>
        <el-table-column prop="profitRate" label="利润率" width="100" align="center">
          <template #default="{ row }">
            <span :class="row.profitRate >= 0 ? 'text-success' : 'text-danger'">
              {{ formatPercent(row.profitRate) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="avgPrice" label="客单价" width="120" align="right">
          <template #default="{ row }">
            {{ formatCurrency(row.avgPrice) }}
          </template>
        </el-table-column>
        <el-table-column prop="returnCount" label="退货数" width="100" align="center" />
        <el-table-column prop="returnRate" label="退货率" min-width="100" align="center">
          <template #default="{ row }">
            {{ formatPercent(row.returnRate) }}
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.page_size"
          :total="pagination.total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSearch"
          @current-change="handleSearch"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getSalesSummary, getSalesTrend } from '@/api/report'
import { formatCurrency, formatPercent } from '@/utils/format'
import CardStatistic from '@/components/Charts/CardStatistic.vue'
import LineChart from '@/components/Charts/LineChart.vue'

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
const dimension = ref('day')
const loading = ref(false)

const handleSearch = () => {
  fetchSummary()
  fetchTrend()
  fetchTableData()
}

const handleReset = () => {
  dateRange.value = getDefaultDateRange()
  dimension.value = 'day'
  handleSearch()
}

// ==================== 统计摘要 ====================
const summary = reactive({
  totalSales: 0,
  orderCount: 0,
  avgPrice: 0,
  profitRate: 0,
})

const fetchSummary = async () => {
  try {
    const params = {}
    if (dateRange.value?.length === 2) {
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    }
    const res = await getSalesSummary(params)
    if (res.data) {
      // 适配后端下划线命名到前端驼峰命名
      summary.totalSales = res.data.total_sales || 0
      summary.orderCount = res.data.total_orders || 0
      summary.avgPrice = res.data.avg_order_value || 0
      summary.profitRate = res.data.profit_rate || 0
    }
  } catch (error) {
    console.error('获取销售摘要失败:', error)
  }
}

// ==================== 趋势图 ====================
const trendXData = ref([])
const trendSeriesData = ref([])

const fetchTrend = async () => {
  try {
    const params = { dimension: dimension.value }
    if (dateRange.value?.length === 2) {
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    }
    const res = await getSalesTrend(params)
    if (res.data && Array.isArray(res.data)) {
      // 后端返回数组格式，每个元素包含 date, sales_amount, order_count
      trendXData.value = res.data.map(item => item.date)
      trendSeriesData.value = [
        {
          name: '销售额',
          data: res.data.map(item => item.sales_amount || 0),
          itemStyle: { color: '#1890ff' },
          areaStyle: { color: 'rgba(24,144,255,0.1)' },
        },
        {
          name: '订单数',
          data: res.data.map(item => item.order_count || 0),
          yAxisIndex: 1,
          itemStyle: { color: '#52c41a' },
        },
      ]
    }
  } catch (error) {
    console.error('获取销售趋势失败:', error)
    initMockTrend()
  }
}

const initMockTrend = () => {
  const days = []
  const salesData = []
  const orderData = []
  for (let i = 1; i <= 30; i++) {
    days.push(`04-${String(i).padStart(2, '0')}`)
    salesData.push(Math.floor(Math.random() * 80000 + 20000))
    orderData.push(Math.floor(Math.random() * 40 + 10))
  }
  trendXData.value = days
  trendSeriesData.value = [
    { name: '销售额', data: salesData, itemStyle: { color: '#1890ff' }, areaStyle: { color: 'rgba(24,144,255,0.1)' } },
    { name: '订单数', data: orderData, yAxisIndex: 1, itemStyle: { color: '#52c41a' } },
  ]
}

// ==================== 明细表格 ====================
const tableData = ref([])
const pagination = reactive({
  page: 1,
  page_size: 10,
  total: 0,
})

const fetchTableData = async () => {
  loading.value = true
  try {
    const params = { dimension: dimension.value }
    if (dateRange.value?.length === 2) {
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    }
    // 使用趋势数据作为表格数据
    const res = await getSalesTrend(params)
    if (res.data && Array.isArray(res.data)) {
      // 转换后端数据为前端表格格式
      tableData.value = res.data.map(item => ({
        date: item.date,
        orderCount: item.order_count || 0,
        salesAmount: item.sales_amount || 0,
        costAmount: item.cost || 0,
        profit: item.profit || 0,
        profitRate: item.sales_amount > 0 ? ((item.profit || 0) / item.sales_amount * 100).toFixed(2) : 0,
        avgPrice: item.order_count > 0 ? ((item.sales_amount || 0) / item.order_count).toFixed(2) : 0,
        returnCount: 0,
        returnRate: 0,
      }))
      pagination.total = tableData.value.length
    }
  } catch (error) {
    console.error('获取销售明细失败:', error)
    initMockTable()
  } finally {
    loading.value = false
  }
}

const initMockTable = () => {
  const list = []
  for (let i = 1; i <= 10; i++) {
    const sales = Math.floor(Math.random() * 80000 + 20000)
    const cost = Math.floor(sales * (0.5 + Math.random() * 0.2))
    list.push({
      date: `2024-04-${String(i).padStart(2, '0')}`,
      orderCount: Math.floor(Math.random() * 40 + 10),
      salesAmount: sales,
      costAmount: cost,
      profit: sales - cost,
      profitRate: ((sales - cost) / sales * 100).toFixed(1),
      avgPrice: Math.floor(sales / (Math.floor(Math.random() * 40 + 10))),
      returnCount: Math.floor(Math.random() * 3),
      returnRate: (Math.random() * 5).toFixed(1),
    })
  }
  tableData.value = list
  pagination.total = 30
}

// ==================== 初始化 ====================
onMounted(() => {
  handleSearch()
})
</script>

<style lang="scss" scoped>
.sales-report {
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

  .pagination-wrapper {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }

  .text-success {
    color: #52c41a;
  }

  .text-danger {
    color: #f5222d;
  }
}
</style>
