<template>
  <div class="stock-manage">
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
        <el-form-item label="商品搜索">
          <el-input
            v-model="searchForm.keyword"
            placeholder="商品名称/SKU编码"
            clearable
            style="width: 220px"
            @keyup.enter="handleSearch"
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
          <span class="title">库存查询</span>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="stockList"
        border
        stripe
        style="width: 100%"
        :row-class-name="tableRowClassName"
      >
        <el-table-column label="仓库" width="140">
          <template #default="{ row }">
            {{ row.warehouse?.warehouse_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="商品名称" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.sku?.sku_name || row.sku?.product?.product_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="商品编码" width="160">
          <template #default="{ row }">
            <span>{{ row.sku?.product?.product_code || '-' }}</span>
            <span v-if="row.sku?.color || row.sku?.size" class="sku-info">
              ({{ [row.sku?.color, row.sku?.size].filter(Boolean).join(' / ') }})
            </span>
          </template>
        </el-table-column>
        <el-table-column label="品牌" width="120">
          <template #default="{ row }">
            {{ row.sku?.product?.brand || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="款式" width="160">
          <template #default="{ row }">
            {{ row.sku?.product?.style || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="总库存" width="100" align="center">
          <template #default="{ row }">
            {{ row.stock_quantity ?? 0 }}
          </template>
        </el-table-column>
        <el-table-column label="可用库存" width="100" align="center">
          <template #default="{ row }">
            {{ row.available_quantity ?? 0 }}
          </template>
        </el-table-column>
        <el-table-column label="锁定库存" width="100" align="center">
          <template #default="{ row }">
            {{ row.locked_quantity ?? 0 }}
          </template>
        </el-table-column>
        <el-table-column label="在途库存" width="100" align="center">
          <template #default="{ row }">
            {{ row.in_transit_quantity ?? 0 }}
          </template>
        </el-table-column>
        <el-table-column label="待分配库存" width="100" align="center">
          <template #default="{ row }">
            {{ row.pending_quantity ?? 0 }}
          </template>
        </el-table-column>
        <el-table-column label="预警状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag
              v-if="(row.stock_quantity ?? 0) <= (row.warning_stock ?? 10)"
              type="danger"
              size="small"
              effect="dark"
            >
              预警
            </el-tag>
            <el-tag v-else type="success" size="small">正常</el-tag>
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
import { getStockList } from '@/api/inventory'
import { getWarehouseList } from '@/api/warehouse'

// ==================== 搜索与列表 ====================
const loading = ref(false)
const stockList = ref([])
const warehouseOptions = ref([])

const searchForm = reactive({
  warehouse_id: '',
  keyword: '',
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
    if (searchForm.keyword) params.keyword = searchForm.keyword

    const res = await getStockList(params)
    stockList.value = res.data?.list || res.data || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取库存列表失败:', error)
  } finally {
    loading.value = false
  }
}

const fetchWarehouseOptions = async () => {
  try {
    const res = await getWarehouseList()
    warehouseOptions.value = res.data || []
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
  searchForm.keyword = ''
  pagination.page = 1
  fetchList()
}

// 预警行高亮
const tableRowClassName = ({ row }) => {
  if (row.is_alert) return 'alert-row'
  return ''
}

// ==================== 初始化 ====================
onMounted(() => {
  fetchList()
  fetchWarehouseOptions()
})
</script>

<style lang="scss" scoped>
.stock-manage {
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

  .sku-info {
    color: #909399;
    font-size: 12px;
    margin-left: 4px;
  }
}

:deep(.alert-row) {
  background-color: #fef0f0 !important;
}
</style>
