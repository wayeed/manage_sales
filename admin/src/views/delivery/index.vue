<template>
  <div class="delivery-manage">
    <!-- 搜索栏 -->
    <el-card class="search-card" shadow="never">
      <el-form :model="searchForm" inline>
        <el-form-item label="订单编号">
          <el-input v-model="searchForm.order_no" placeholder="请输入订单编号" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="送货单号">
          <el-input v-model="searchForm.delivery_no" placeholder="请输入送货单号" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 120px">
            <el-option label="正常" :value="1" />
            <el-option label="作废" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="送货时间">
          <el-date-picker v-model="searchForm.dateRange" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" style="width: 260px" />
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
          <span class="title">送货出库记录</span>
          <el-button type="primary" icon="Plus" @click="handleCreate" v-permission="['delivery:create']">新建送货出库</el-button>
        </div>
      </template>
      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="delivery_no" label="送货单号" min-width="160" fixed />
        <el-table-column prop="order_no" label="订单编号" min-width="160">
          <template #default="{ row }">
            <el-link type="primary" @click="handleOrderDetail(row)">{{ row.order_no }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="warehouse_name" label="出库仓库" min-width="120" />
        <el-table-column prop="operator_name" label="操作人" min-width="100" />
        <el-table-column prop="delivery_type_name" label="送货类型" min-width="90" />
        <el-table-column prop="total_quantity" label="总数量" min-width="80" align="center" />
        <el-table-column prop="total_amount" label="总金额" min-width="110" align="right">
          <template #default="{ row }">
            <span class="price">{{ formatMoney(row.total_amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status_name" label="状态" min-width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status_name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="delivery_time" label="送货时间" min-width="160">
          <template #default="{ row }">
            {{ formatDateTime(row.delivery_time) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right" align="center">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleDetail(row)">详情</el-button>
            <el-button link type="danger" size="small" @click="handleCancel(row)" v-if="row.status === 1" v-permission="['delivery:cancel']">作废</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-wrapper">
        <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.page_size" :page-sizes="[10, 20, 50, 100]" :total="pagination.total" layout="total, sizes, prev, pager, next, jumper" @size-change="fetchList" @current-change="fetchList" />
      </div>
    </el-card>

    <!-- 送货出库弹窗 -->
    <DeliveryFormDialog v-model:visible="formDialogVisible" @success="fetchList" />

    <!-- 送货详情弹窗 -->
    <DeliveryDetailDialog v-model:visible="detailDialogVisible" :delivery-id="currentDeliveryId" />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getDeliveryList, cancelDelivery } from '@/api/delivery'
import { formatMoney, formatDateTime } from '@/utils/format'
import DeliveryFormDialog from './components/DeliveryFormDialog.vue'
import DeliveryDetailDialog from './components/DeliveryDetailDialog.vue'

const router = useRouter()

// ==================== 搜索与列表 ====================
const loading = ref(false)
const list = ref([])
const searchForm = reactive({
  order_no: '',
  delivery_no: '',
  status: '',
  dateRange: [],
})
const pagination = reactive({ page: 1, page_size: 20, total: 0 })

const getSearchParams = () => {
  const params = {
    page: pagination.page,
    page_size: pagination.page_size,
  }
  if (searchForm.order_no) params.order_no = searchForm.order_no
  if (searchForm.delivery_no) params.delivery_no = searchForm.delivery_no
  if (searchForm.status !== '' && searchForm.status !== null) params.status = searchForm.status
  if (searchForm.dateRange?.length === 2) {
    params.start_time = searchForm.dateRange[0] + ' 00:00:00'
    params.end_time = searchForm.dateRange[1] + ' 23:59:59'
  }
  return params
}

const fetchList = async () => {
  loading.value = true
  try {
    const res = await getDeliveryList(getSearchParams())
    list.value = res.data?.list || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取送货列表失败:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => { pagination.page = 1; fetchList() }
const handleReset = () => {
  searchForm.order_no = ''
  searchForm.delivery_no = ''
  searchForm.status = ''
  searchForm.dateRange = []
  handleSearch()
}

// ==================== 操作 ====================
const formDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const currentDeliveryId = ref(null)

const handleCreate = () => { formDialogVisible.value = true }

const handleDetail = (row) => {
  currentDeliveryId.value = row.id
  detailDialogVisible.value = true
}

const handleOrderDetail = (row) => {
  router.push(`/order/detail/${row.order_id}`)
}

const handleCancel = (row) => {
  ElMessageBox.confirm('确定要作废该送货记录吗？作废后不可恢复。', '作废确认', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    try {
      await cancelDelivery(row.id, {})
      ElMessage.success('作废成功')
      fetchList()
    } catch (error) {
      console.error('作废失败:', error)
    }
  }).catch(() => {})
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
