<template>
  <div class="outbound-approval">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">出库审批</span>
        </div>
      </template>

      <!-- Tab 切换 -->
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="待主管审批" name="supervisor" />
        <el-tab-pane label="待财务审批" name="finance" />
      </el-tabs>

      <!-- 列表表格 -->
      <el-table v-loading="loading" :data="tableData" border stripe style="width: 100%">
        <el-table-column prop="order.order_no" label="订单号" min-width="160" show-overflow-tooltip />
        <el-table-column prop="order.customer_name" label="客户名称" min-width="140" show-overflow-tooltip />
        <el-table-column prop="applicant_name" label="业务员" width="100" />
        <el-table-column prop="remaining_amount" label="尾款金额" width="120" align="right">
          <template #default="{ row }">
            {{ formatCurrency(row.remaining_amount) }}
          </template>
        </el-table-column>
        <el-table-column prop="remaining_rate" label="尾款比例" width="100" align="center">
          <template #default="{ row }">
            {{ (row.remaining_rate || 0).toFixed(2) }}%
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="申请备注" min-width="160" show-overflow-tooltip />
        <el-table-column prop="created_at" label="申请时间" width="170">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right" align="center">
          <template #default="{ row }">
            <el-popconfirm
              title="确认审批通过该出库申请？"
              confirm-button-text="确认"
              cancel-button-text="取消"
              @confirm="handleApprove(row)"
            >
              <template #reference>
                <el-button type="success" size="small" :loading="row._approving">审批通过</el-button>
              </template>
            </el-popconfirm>
            <el-button type="danger" size="small" :loading="row._rejecting" @click="handleReject(row)">拒绝</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.page_size"
          :total="pagination.total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPendingOutboundRequests, supervisorApprove, financeApprove, rejectOutboundRequest } from '@/api/outbound-request'
import { formatDate, formatCurrency } from '@/utils/format'

const activeTab = ref('supervisor')
const loading = ref(false)
const tableData = ref([])
const pagination = reactive({ page: 1, page_size: 20, total: 0 })

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getPendingOutboundRequests({
      role: activeTab.value,
      page: pagination.page,
      page_size: pagination.page_size,
    })
    tableData.value = res.data?.list || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取审批列表失败:', error)
  } finally {
    loading.value = false
  }
}

const handleApprove = async (row) => {
  row._approving = true
  try {
    if (activeTab.value === 'supervisor') {
      await supervisorApprove(row.id, { remark: '' })
    } else {
      await financeApprove(row.id, { remark: '' })
    }
    ElMessage.success('审批通过')
    fetchData()
  } catch (error) {
    ElMessage.error(error.message || '审批失败')
  } finally {
    row._approving = false
  }
}

const handleReject = async (row) => {
  try {
    const { value } = await ElMessageBox.prompt('请输入拒绝原因', '拒绝出库申请', {
      confirmButtonText: '确认拒绝',
      cancelButtonText: '取消',
      inputType: 'textarea',
    })
    row._rejecting = true
    await rejectOutboundRequest(row.id, { remark: value || '拒绝' })
    ElMessage.success('已拒绝')
    fetchData()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '操作失败')
    }
  } finally {
    row._rejecting = false
  }
}

const handleTabChange = () => {
  pagination.page = 1
  fetchData()
}

const handleSizeChange = (size) => {
  pagination.page_size = size
  pagination.page = 1
  fetchData()
}

const handlePageChange = (page) => {
  pagination.page = page
  fetchData()
}

onMounted(() => fetchData())
</script>

<style scoped>
.outbound-approval {
  padding: 20px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-header .title {
  font-size: 18px;
  font-weight: 600;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
