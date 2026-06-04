<template>
  <div class="transaction-manage">
    <!-- 搜索栏 -->
    <el-card class="search-card" shadow="never">
      <el-form :model="searchForm" inline>
        <el-form-item label="仓库">
          <el-select
            v-model="searchForm.warehouse_id"
            placeholder="全部仓库"
            clearable
            style="width: 160px"
          >
            <el-option
              v-for="item in warehouseOptions"
              :key="item.id"
              :label="item.warehouse_name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="类型">
          <el-select
            v-model="searchForm.type"
            placeholder="全部类型"
            clearable
            style="width: 140px"
          >
            <el-option label="采购入库" :value="1" />
            <el-option label="销售出库" :value="2" />
            <el-option label="调拨出库" :value="3" />
            <el-option label="调拨入库" :value="4" />
            <el-option label="盘盈" :value="5" />
            <el-option label="盘亏" :value="6" />
            <el-option label="礼品出库" :value="7" />
            <el-option label="礼品入库" :value="8" />
            <el-option label="库存锁定" :value="9" />
            <el-option label="库存解锁" :value="10" />
            <el-option label="锁定转出库" :value="11" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="searchForm.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 260px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="Search" @click="handleSearch">搜索</el-button>
          <el-button icon="Refresh" @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 数据表格 -->
    <el-card class="table-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">库存流水</span>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="transactionList"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column label="类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getTypeTag(row.transaction_type)" size="small">
              {{ getTypeLabel(row.transaction_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="商品" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.sku?.product?.product_name || row.sku?.sku_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="SKU" width="140">
          <template #default="{ row }">
            {{ row.sku?.sku_code || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="仓库" width="120">
          <template #default="{ row }">
            {{ row.warehouse?.warehouse_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="数量" width="80" align="center">
          <template #default="{ row }">
            <span :class="isOutType(row.transaction_type) ? 'text-danger' : 'text-success'">
              {{ isOutType(row.transaction_type) ? '-' : '+' }}{{ Math.abs(row.quantity) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="before_stock" label="变动前" width="90" align="center" />
        <el-table-column prop="after_stock" label="变动后" width="90" align="center" />
        <el-table-column label="成本" width="100" align="right">
          <template #default="{ row }">
            <span v-if="row.total_cost && Number(row.total_cost) > 0">¥{{ row.total_cost }}</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="时间" width="170">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作人" width="100">
          <template #default="{ row }">
            {{ row.creator?.real_name || '-' }}
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
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getInventoryTransactions } from '@/api/inventory'
import { getWarehouseList } from '@/api/warehouse'

// ==================== 搜索与列表 ====================
const loading = ref(false)
const transactionList = ref([])
const warehouseOptions = ref([])

const searchForm = reactive({
  warehouse_id: '',
  type: '',
  dateRange: [],
})

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
    if (searchForm.warehouse_id) params.warehouse_id = searchForm.warehouse_id
    if (searchForm.type) params.type = searchForm.type
    if (searchForm.dateRange && searchForm.dateRange.length === 2) {
      params.start_date = searchForm.dateRange[0]
      params.end_date = searchForm.dateRange[1]
    }

    const res = await getInventoryTransactions(params)
    transactionList.value = res.data?.list || res.data || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取流水列表失败:', error)
  } finally {
    loading.value = false
  }
}

const fetchWarehouseOptions = async () => {
  try {
    const res = await getWarehouseList()
    warehouseOptions.value = res.data?.list || res.data || []
  } catch (error) {
    console.error('获取仓库列表失败:', error)
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchList()
}

const handleReset = () => {
  searchForm.warehouse_id = ''
  searchForm.type = ''
  searchForm.dateRange = []
  pagination.page = 1
  fetchList()
}

const getTypeLabel = (type) => {
  const map = {
    1: '采购入库',
    2: '销售出库',
    3: '调拨出库',
    4: '调拨入库',
    5: '盘盈',
    6: '盘亏',
    7: '礼品出库',
    8: '礼品入库',
    9: '库存锁定',
    10: '库存解锁',
    11: '锁定转出库',
  }
  return map[type] ?? type ?? '未知'
}

const getTypeTag = (type) => {
  const map = {
    1: 'success',
    2: 'danger',
    3: 'warning',
    4: '',
    5: 'success',
    6: 'danger',
    7: 'warning',
    8: '',
    9: 'info',
    10: 'success',
    11: 'warning',
  }
  return map[type] || 'info'
}

// 判断是否为出库类型（库存减少）
const isOutType = (type) => [2, 3, 6, 7, 9, 11].includes(type)

const formatTime = (time) => {
  if (!time) return '-'
  return time.replace('T', ' ').substring(0, 19)
}

// ==================== 初始化 ====================
onMounted(() => {
  fetchList()
  fetchWarehouseOptions()
})
</script>

<style lang="scss" scoped>
.transaction-manage {
  .search-card {
    margin-bottom: 16px;
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

  .table-card {
    .pagination-wrapper {
      display: flex;
      justify-content: flex-end;
      margin-top: 16px;
    }
  }

  .text-success {
    color: #67c23a;
    font-weight: 500;
  }

  .text-danger {
    color: #f56c6c;
    font-weight: 500;
  }

  .text-muted {
    color: #909399;
  }
}
</style>
