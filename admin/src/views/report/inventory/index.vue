<template>
  <div class="inventory-report">
    <!-- 统计卡片 -->
    <el-row :gutter="16">
      <el-col :span="8">
        <CardStatistic title="总SKU数" :value="summary.totalSku" suffix="个" icon="Goods" />
      </el-col>
      <el-col :span="8">
        <CardStatistic title="库存预警数" :value="summary.alertCount" suffix="项" icon="Warning" trend="down" trend-value="12.3%" />
      </el-col>
      <el-col :span="8">
        <CardStatistic title="滞销商品数" :value="summary.slowMovingCount" suffix="项" icon="Timer" trend="up" trend-value="5.6%" />
      </el-col>
    </el-row>

    <!-- 库存周转率趋势 -->
    <el-card shadow="hover" class="chart-card" style="margin-top: 16px;">
      <template #header>
        <div class="card-header">
          <span>库存周转率趋势</span>
        </div>
      </template>
      <LineChart
        :x-data="turnoverXData"
        :series-data="turnoverSeriesData"
        height="350px"
      />
    </el-card>

    <!-- 滞销商品列表 -->
    <el-card shadow="hover" class="chart-card" style="margin-top: 16px;">
      <template #header>
        <div class="card-header">
          <span>滞销商品列表</span>
        </div>
      </template>
      <el-table :data="slowMovingList" border stripe v-loading="loading" style="width: 100%">
        <el-table-column prop="productName" label="商品名称" min-width="150" />
        <el-table-column prop="sku" label="SKU" width="140" />
        <el-table-column prop="category" label="分类" width="100" />
        <el-table-column prop="stock" label="库存数量" width="100" align="center" />
        <el-table-column prop="stockValue" label="库存金额" width="140" align="right">
          <template #default="{ row }">
            {{ formatCurrency(row.stockValue) }}
          </template>
        </el-table-column>
        <el-table-column prop="lastSaleDate" label="最后销售日期" width="130" />
        <el-table-column prop="unsoldDays" label="滞销天数" width="100" align="center">
          <template #default="{ row }">
            <span :class="row.unsoldDays > 180 ? 'text-danger' : row.unsoldDays > 90 ? 'text-warning' : ''">
              {{ row.unsoldDays }}天
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="suggestion" label="建议" min-width="120">
          <template #default="{ row }">
            <el-tag :type="row.unsoldDays > 180 ? 'danger' : 'warning'" size="small">
              {{ row.unsoldDays > 180 ? '建议促销清仓' : '建议降价处理' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getInventoryAnalysis } from '@/api/report'
import { formatCurrency } from '@/utils/format'
import CardStatistic from '@/components/Charts/CardStatistic.vue'
import LineChart from '@/components/Charts/LineChart.vue'

// ==================== 统计摘要 ====================
const summary = reactive({
  totalSku: 0,
  alertCount: 0,
  slowMovingCount: 0,
})

// ==================== 周转率趋势 ====================
const turnoverXData = ref([])
const turnoverSeriesData = ref([])

// ==================== 滞销列表 ====================
const slowMovingList = ref([])
const loading = ref(false)

const fetchInventoryData = async () => {
  loading.value = true
  try {
    const res = await getInventoryAnalysis()
    if (res.data) {
      summary.totalSku = res.data.totalSku || 0
      summary.alertCount = res.data.alertCount || 0
      summary.slowMovingCount = res.data.slowMovingCount || 0

      turnoverXData.value = res.data.turnoverMonths || []
      turnoverSeriesData.value = [
        {
          name: '周转率',
          data: res.data.turnoverRates || [],
          itemStyle: { color: '#1890ff' },
          areaStyle: { color: 'rgba(24,144,255,0.1)' },
        },
      ]

      slowMovingList.value = res.data.slowMovingList || []
    }
  } catch (error) {
    console.error('获取库存分析失败:', error)
    initMockData()
  } finally {
    loading.value = false
  }
}

const initMockData = () => {
  summary.totalSku = 856
  summary.alertCount = 23
  summary.slowMovingCount = 15

  const months = ['01月', '02月', '03月', '04月', '05月', '06月', '07月', '08月', '09月', '10月', '11月', '12月']
  turnoverXData.value = months
  turnoverSeriesData.value = [
    {
      name: '周转率',
      data: [2.1, 2.3, 2.0, 2.5, 2.8, 2.4, 3.0, 2.6, 3.2, 2.9, 3.5, 3.1],
      itemStyle: { color: '#1890ff' },
      areaStyle: { color: 'rgba(24,144,255,0.1)' },
    },
  ]

  slowMovingList.value = [
    { productName: '复古银胸针F001', sku: 'SKU-F001', category: '银饰', stock: 45, stockValue: 6750, lastSaleDate: '2024-01-15', unsoldDays: 210 },
    { productName: '红玛瑙手链G002', sku: 'SKU-G002', category: '宝石', stock: 30, stockValue: 12000, lastSaleDate: '2024-02-20', unsoldDays: 175 },
    { productName: '钛钢戒指H003', sku: 'SKU-H003', category: '其他', stock: 60, stockValue: 3000, lastSaleDate: '2024-03-10', unsoldDays: 156 },
    { productName: '琥珀吊坠I004', sku: 'SKU-I004', category: '宝石', stock: 20, stockValue: 8000, lastSaleDate: '2024-04-05', unsoldDays: 130 },
    { productName: '合金手镯J005', sku: 'SKU-J005', category: '其他', stock: 80, stockValue: 2400, lastSaleDate: '2024-05-01', unsoldDays: 104 },
  ]
}

// ==================== 初始化 ====================
onMounted(() => {
  fetchInventoryData()
})
</script>

<style lang="scss" scoped>
.inventory-report {
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

  .text-warning {
    color: #faad14;
  }
}
</style>
