<template>
  <div class="ranking-report">
    <!-- 筛选栏 -->
    <el-card shadow="never" class="filter-card">
      <el-row :gutter="16" align="middle">
        <el-col :span="6">
          <el-select v-model="rankDimension" placeholder="排行维度" style="width: 100%" @change="handleSearch">
            <el-option label="业务员排行" value="salesman" />
            <el-option label="品类排行" value="category" />
            <el-option label="商品排行" value="product" />
          </el-select>
        </el-col>
        <el-col :span="5">
          <el-select v-model="periodType" style="width: 100%" @change="handleSearch">
            <el-option label="月度" value="month" />
            <el-option label="季度" value="quarter" />
            <el-option label="年度" value="year" />
          </el-select>
        </el-col>
        <el-col :span="5">
          <el-date-picker
            v-model="periodValue"
            :type="periodType === 'month' ? 'month' : periodType === 'quarter' ? 'month' : 'year'"
            placeholder="选择周期"
            :format="periodType === 'month' ? 'YYYY-MM' : periodType === 'quarter' ? 'YYYY-MM' : 'YYYY'"
            :value-format="periodType === 'month' ? 'YYYY-MM' : periodType === 'quarter' ? 'YYYY-MM' : 'YYYY'"
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

    <!-- 排行榜表格 -->
    <el-card shadow="hover" class="chart-card" style="margin-top: 16px;">
      <template #header>
        <div class="card-header">
          <span>{{ dimensionLabel }}排行榜</span>
        </div>
      </template>
      <el-table :data="tableData" border stripe v-loading="loading" style="width: 100%">
        <el-table-column label="排名" width="80" align="center">
          <template #default="{ row, $index }">
            <el-tag
              v-if="$index < 3"
              :type="['danger', 'warning', ''][$index]"
              effect="dark"
              round
              size="small"
            >
              {{ $index + 1 }}
            </el-tag>
            <span v-else>{{ $index + 1 }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" :label="dimensionLabel" min-width="150" />
        <el-table-column prop="salesAmount" label="销售额" width="140" align="right" sortable>
          <template #default="{ row }">
            {{ formatCurrency(row.salesAmount) }}
          </template>
        </el-table-column>
        <el-table-column prop="orderCount" label="订单数" width="100" align="center" sortable />
        <el-table-column prop="profit" label="利润" width="140" align="right" sortable>
          <template #default="{ row }">
            {{ formatCurrency(row.profit) }}
          </template>
        </el-table-column>
        <el-table-column prop="profitRate" label="利润率" width="100" align="center" sortable>
          <template #default="{ row }">
            <span :class="row.profitRate >= 0 ? 'text-success' : 'text-danger'">
              {{ formatPercent(row.profitRate) }}
            </span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 柱状图展示 TOP10 -->
    <el-card shadow="hover" class="chart-card" style="margin-top: 16px;">
      <template #header>
        <div class="card-header">
          <span>{{ dimensionLabel }}销售额 TOP10</span>
        </div>
      </template>
      <BarChart
        :x-data="chartXData"
        :series-data="chartSeriesData"
        height="400px"
      />
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getPerformanceRanking } from '@/api/report'
import { formatCurrency, formatPercent } from '@/utils/format'
import BarChart from '@/components/Charts/BarChart.vue'

// ==================== 筛选条件 ====================
const rankDimension = ref('salesman')
const periodType = ref('month')
const getDefaultPeriodValue = () => {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
}
const periodValue = ref(getDefaultPeriodValue())
const loading = ref(false)

const dimensionLabel = computed(() => {
  const map = { salesman: '业务员', category: '品类', product: '商品' }
  return map[rankDimension.value] || '业务员'
})

const handleSearch = () => {
  fetchRankingData()
}

const handleReset = () => {
  rankDimension.value = 'salesman'
  periodType.value = 'month'
  periodValue.value = getDefaultPeriodValue()
  handleSearch()
}

// ==================== 排行数据 ====================
const tableData = ref([])
const chartXData = ref([])
const chartSeriesData = ref([])

const fetchRankingData = async () => {
  loading.value = true
  try {
    const params = {
      period_type: periodType.value,
      period_value: periodValue.value || '',
      rank_by: rankDimension.value,
    }
    const res = await getPerformanceRanking(params)
    if (res.data && Array.isArray(res.data)) {
      // 适配后端下划线命名到前端驼峰命名
      tableData.value = res.data.map(item => ({
        name: item.name,
        salesAmount: item.sales_amount || 0,
        orderCount: item.order_count || 0,
        profit: item.profit || 0,
        profitRate: item.profit_rate || 0,
      }))
      chartXData.value = tableData.value.slice(0, 10).map((item) => item.name)
      chartSeriesData.value = [
        {
          name: '销售额',
          data: tableData.value.slice(0, 10).map((item) => item.salesAmount),
          itemStyle: {
            color: (params) => {
              const colors = ['#f5222d', '#fa8c16', '#1890ff', '#52c41a', '#722ed1', '#13c2c2', '#eb2f96', '#faad14', '#a0d911', '#2f54eb']
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
    console.error('获取排行数据失败:', error)
    initMockData()
  } finally {
    loading.value = false
  }
}

const initMockData = () => {
  const names = rankDimension.value === 'salesman'
    ? ['张三', '李四', '王五', '赵六', '钱七', '孙八', '周九', '吴十']
    : rankDimension.value === 'category'
    ? ['黄金饰品', '银饰', '钻石', '翡翠', '珍珠', '铂金']
    : ['金项链A001', '钻戒B002', '翡翠手镯C003', '银手链D004', '珍珠耳环E005']

  const list = names.map((name) => {
    const sales = Math.floor(Math.random() * 200000 + 30000)
    const orders = Math.floor(Math.random() * 50 + 5)
    const profit = Math.floor(sales * (0.2 + Math.random() * 0.15))
    return {
      name,
      salesAmount: sales,
      orderCount: orders,
      profit,
      profitRate: ((profit / sales) * 100).toFixed(1),
    }
  })
  list.sort((a, b) => b.salesAmount - a.salesAmount)

  tableData.value = list
  chartXData.value = list.slice(0, 10).map((item) => item.name)
  chartSeriesData.value = [
    {
      name: '销售额',
      data: list.slice(0, 10).map((item) => item.salesAmount),
      itemStyle: {
        color: (params) => {
          const colors = ['#f5222d', '#fa8c16', '#1890ff', '#52c41a', '#722ed1', '#13c2c2', '#eb2f96', '#faad14', '#a0d911', '#2f54eb']
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
  handleSearch()
})
</script>

<style lang="scss" scoped>
.ranking-report {
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

  .text-success {
    color: #52c41a;
  }

  .text-danger {
    color: #f5222d;
  }
}
</style>
