<template>
  <div class="delivery-manage">
    <!-- 搜索栏 -->
    <el-card class="search-card" shadow="never">
      <el-form :model="searchForm" inline>
        <el-form-item label="订单编号">
          <el-input v-model="searchForm.order_no" placeholder="请输入订单编号" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="客户姓名">
          <el-input v-model="searchForm.customer_name" placeholder="请输入客户姓名" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="业务员">
          <el-input v-model="searchForm.salesman_name" placeholder="请输入业务员" clearable @keyup.enter="handleSearch" />
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
          <span class="title">待送货订单</span>
          <el-tag type="warning">仅显示已审核未配送的订单</el-tag>
        </div>
      </template>
      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="order_no" label="订单编号" min-width="160" fixed>
          <template #default="{ row }">
            <el-link type="primary" @click="handleOrderDetail(row)">{{ row.order_no }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="customer_name" label="客户姓名" min-width="100" />
        <el-table-column prop="customer_phone" label="客户电话" min-width="120" />
        <el-table-column prop="customer_address" label="收货地址" min-width="200" show-overflow-tooltip />
        <el-table-column prop="salesman_name" label="业务员" min-width="100" />
        <el-table-column prop="total_quantity" label="商品数量" min-width="80" align="center" />
        <el-table-column prop="total_amount" label="订单金额" min-width="110" align="right">
          <template #default="{ row }">
            <span class="price">{{ formatMoney(row.total_amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="order_time" label="下单时间" min-width="160">
          <template #default="{ row }">
            {{ formatDateTime(row.order_time) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleDelivery(row)" v-permission="['delivery:create']">送货出库</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-wrapper">
        <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.page_size" :page-sizes="[10, 20, 50, 100]" :total="pagination.total" layout="total, sizes, prev, pager, next, jumper" @size-change="fetchList" @current-change="fetchList" />
      </div>
    </el-card>

    <!-- 送货出库弹窗 -->
    <DeliveryFormDialog v-model:visible="formDialogVisible" :order="selectedOrder" @success="fetchList" />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getPendingDeliveryOrders } from '@/api/delivery'
import { formatMoney, formatDateTime } from '@/utils/format'
import DeliveryFormDialog from './components/DeliveryFormDialog.vue'

const router = useRouter()

// ==================== 搜索与列表 ====================
const loading = ref(false)
const list = ref([])
const searchForm = reactive({
  order_no: '',
  customer_name: '',
  salesman_name: '',
})
const pagination = reactive({ page: 1, page_size: 20, total: 0 })

const getSearchParams = () => {
  const params = {
    page: pagination.page,
    page_size: pagination.page_size,
  }
  if (searchForm.order_no) params.order_no = searchForm.order_no
  if (searchForm.customer_name) params.customer_name = searchForm.customer_name
  if (searchForm.salesman_name) params.salesman_name = searchForm.salesman_name
  return params
}

const fetchList = async () => {
  loading.value = true
  try {
    const res = await getPendingDeliveryOrders(getSearchParams())
    list.value = res.data?.list || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取待送货订单失败:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => { pagination.page = 1; fetchList() }
const handleReset = () => {
  searchForm.order_no = ''
  searchForm.customer_name = ''
  searchForm.salesman_name = ''
  handleSearch()
}

// ==================== 操作 ====================
const formDialogVisible = ref(false)
const selectedOrder = ref(null)

const handleDelivery = (row) => {
  selectedOrder.value = row
  formDialogVisible.value = true
}

const handleOrderDetail = (row) => {
  router.push(`/order/detail/${row.order_id}`)
}

// ==================== 初始化 ====================
onMounted(() => { fetchList() })
</script>

<style lang="scss" scoped>
.delivery-manage {
  .search-card {
    margin-bottom: 16px;
  }
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    .title {
      font-size: 16px;
      font-weight: 600;
    }
  }
  .price {
    color: #f56c6c;
    font-weight: 500;
  }
  .pagination-wrapper {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }
}
</style>
