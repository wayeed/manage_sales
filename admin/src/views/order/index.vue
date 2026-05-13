<template>
  <div class="order-manage">
    <!-- 搜索栏 -->
    <el-card class="search-card" shadow="never">
      <el-form :model="searchForm" inline>
        <el-form-item label="订单号">
          <el-input
            v-model="searchForm.order_no"
            placeholder="请输入订单号"
            clearable
            style="width: 180px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="订单状态">
          <el-select
            v-model="searchForm.status"
            placeholder="全部状态"
            clearable
            style="width: 140px"
          >
            <el-option label="待审批" :value="0" />
            <el-option label="已生效" :value="1" />
            <el-option label="已驳回" :value="2" />
            <el-option label="已取消" :value="3" />
            <el-option label="已退货" :value="4" />
          </el-select>
        </el-form-item>
        <el-form-item label="订单类型">
          <el-select
            v-model="searchForm.order_type"
            placeholder="全部类型"
            clearable
            style="width: 140px"
          >
            <el-option label="单品" :value="1" />
            <el-option label="多品" :value="2" />
            <el-option label="特殊审批" :value="3" />
            <el-option label="同行单品" :value="4" />
            <el-option label="同行多品" :value="5" />
            <el-option label="同行特批" :value="6" />
          </el-select>
        </el-form-item>
        <el-form-item label="回款状态">
          <el-select
            v-model="searchForm.payment_status"
            placeholder="全部"
            clearable
            style="width: 140px"
          >
            <el-option label="未回款" :value="0" />
            <el-option label="部分回款" :value="1" />
            <el-option label="已回款" :value="2" />
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
            style="width: 240px"
          />
        </el-form-item>
        <el-form-item label="业务员">
          <el-select
            v-model="searchForm.salesman_id"
            placeholder="全部"
            clearable
            style="width: 140px"
          >
            <el-option
              v-for="item in salesmanOptions"
              :key="item.id"
              :label="item.real_name"
              :value="item.id"
            />
          </el-select>
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
          <span class="title">订单列表</span>
          <el-button type="success" icon="Download" @click="handleExport">导出</el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="orderList"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column prop="order_no" label="订单号" width="180" fixed="left" />
        <el-table-column label="客户" width="140">
          <template #default="{ row }">
            {{ row.customer_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="业务员" width="100">
          <template #default="{ row }">
            {{ row.salesman?.real_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="订单类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getOrderTypeTag(row.order_type)" size="small">
              {{ getOrderTypeLabel(row.order_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="订单金额" width="120" align="right">
          <template #default="{ row }">
            <span class="price">{{ formatCurrency(row.final_amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="实际利润" width="120" align="right">
          <template #default="{ row }">
            <span :class="row.actual_profit >= 0 ? 'profit' : 'loss'">
              {{ formatCurrency(row.actual_profit) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="回款状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getPaymentStatusTag(row.payment_status)" size="small">
              {{ getPaymentStatusLabel(row.payment_status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="订单状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusTag(row.order_status)" size="small">
              {{ getStatusLabel(row.order_status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="200" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleDetail(row)">
              查看详情
            </el-button>
            <el-button
              v-if="row.order_status === 0"
              type="success" link size="small" @click="handleApprove(row)"
            >
              审核
            </el-button>
            <el-button
              v-if="row.order_status === 0"
              type="danger" link size="small" @click="handleCancel(row)"
            >
              取消
            </el-button>
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

    <!-- 审核弹窗 -->
    <ApproveDialog
      v-model:visible="approveDialogVisible"
      :order="currentOrder"
      @success="fetchList"
    />

    <!-- 导出弹窗 -->
    <ExportDialog
      v-model:visible="exportDialogVisible"
      :search-params="getSearchParams()"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getOrderList, cancelOrder } from '@/api/order'
import { getUserList } from '@/api/user'
import { formatCurrency } from '@/utils/format'
import ApproveDialog from './components/ApproveDialog.vue'
import ExportDialog from './components/ExportDialog.vue'

const router = useRouter()

// ==================== 搜索与列表 ====================
const loading = ref(false)
const orderList = ref([])
const salesmanOptions = ref([])

const searchForm = reactive({
  order_no: '',
  status: '',
  order_type: '',
  payment_status: '',
  dateRange: null,
  salesman_id: '',
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
      ...getSearchParams(),
    }
    const res = await getOrderList(params)
    orderList.value = res.data?.list || res.data || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取订单列表失败:', error)
  } finally {
    loading.value = false
  }
}

const getSearchParams = () => {
  const params = {}
  if (searchForm.order_no) params.keyword = searchForm.order_no
  if (searchForm.status) params.order_status = searchForm.status
  if (searchForm.order_type) params.order_type = searchForm.order_type
  if (searchForm.payment_status) params.payment_status = searchForm.payment_status
  if (searchForm.salesman_id) params.salesman_id = searchForm.salesman_id
  if (searchForm.dateRange && searchForm.dateRange.length === 2) {
    params.start_date = searchForm.dateRange[0]
    params.end_date = searchForm.dateRange[1]
  }
  return params
}

const handleSearch = () => {
  pagination.page = 1
  fetchList()
}

const handleReset = () => {
  searchForm.order_no = ''
  searchForm.status = ''
  searchForm.order_type = ''
  searchForm.payment_status = ''
  searchForm.dateRange = null
  searchForm.salesman_id = ''
  pagination.page = 1
  fetchList()
}

// ==================== 状态映射 ====================
const getStatusLabel = (status) => {
  const map = {
    0: '待审批',
    1: '已生效',
    2: '已驳回',
    3: '已取消',
    4: '已退货',
  }
  return map[status] ?? status ?? '未知'
}

const getStatusTag = (status) => {
  const map = {
    0: 'warning',
    1: 'success',
    2: 'danger',
    3: 'info',
    4: '',
  }
  return map[status] ?? 'info'
}

const getOrderTypeLabel = (type) => {
  const map = { 1: '单品', 2: '多品', 3: '特殊审批', 4: '同行单品', 5: '同行多品', 6: '同行特批' }
  return map[type] ?? type ?? '-'
}

const getOrderTypeTag = (type) => {
  const map = { 1: '', 2: 'success', 3: 'warning', 4: 'info', 5: '', 6: 'danger' }
  return map[type] ?? 'info'
}

const getPaymentStatusLabel = (status) => {
  const map = { 0: '未回款', 1: '部分回款', 2: '已回款' }
  return map[status] ?? status ?? '-'
}

const getPaymentStatusTag = (status) => {
  const map = { 0: 'danger', 1: 'warning', 2: 'success' }
  return map[status] ?? 'info'
}

// ==================== 操作 ====================
const handleDetail = (row) => {
  router.push(`/order/detail/${row.id}`)
}

const currentOrder = ref(null)
const approveDialogVisible = ref(false)

const handleApprove = (row) => {
  currentOrder.value = row
  approveDialogVisible.value = true
}

const handleCancel = (row) => {
  ElMessageBox.confirm(
    `确定要取消订单 "${row.order_no}" 吗？`,
    '取消确认',
    {
      confirmButtonText: '确定取消',
      cancelButtonText: '返回',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await cancelOrder(row.id)
      ElMessage.success('取消成功')
      fetchList()
    } catch (error) {
      console.error('取消订单失败:', error)
    }
  }).catch(() => {})
}

const exportDialogVisible = ref(false)

const handleExport = () => {
  exportDialogVisible.value = true
}

// ==================== 初始化 ====================
const fetchSalesmanOptions = async () => {
  try {
    const res = await getUserList({ page: 1, page_size: 1000 })
    const list = res.data?.list || res.data || []
    salesmanOptions.value = list.map(item => ({
      id: item.id,
      real_name: item.real_name || item.username,
    }))
  } catch (error) {
    console.error('获取业务员列表失败:', error)
  }
}

onMounted(() => {
  fetchList()
  fetchSalesmanOptions()
})
</script>

<style lang="scss" scoped>
.order-manage {
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

  .price {
    color: #f56c6c;
    font-weight: 500;
  }

  .profit {
    color: #67c23a;
    font-weight: 500;
  }

  .loss {
    color: #f56c6c;
    font-weight: 500;
  }
}
</style>
