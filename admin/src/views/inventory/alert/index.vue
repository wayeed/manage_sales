<template>
  <div class="alert-manage">
    <!-- 搜索栏 -->
    <el-card class="search-card" shadow="never">
      <el-form :model="searchForm" inline>
        <el-form-item label="状态">
          <el-select
            v-model="searchForm.status"
            placeholder="全部状态"
            clearable
            style="width: 120px"
          >
            <el-option label="未处理" :value="0" />
            <el-option label="已处理" :value="1" />
          </el-select>
        </el-form-item>
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
          <span class="title">库存预警</span>
          <el-tag type="danger" effect="dark" size="small">
            {{ pendingCount }} 条待处理
          </el-tag>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="alertList"
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
        <el-table-column label="商品" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.product?.name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="SKU" width="140">
          <template #default="{ row }">
            {{ row.sku?.sku_code || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="current_stock" label="当前库存" width="100" align="center">
          <template #default="{ row }">
            <span class="stock-num">{{ row.current_stock }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="warning_stock" label="预警阈值" width="100" align="center" />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.alert_status === 0 ? 'danger' : 'success'" size="small">
              {{ row.alert_status === 0 ? '未处理' : '已处理' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="预警时间" width="180" />
        <el-table-column label="操作" width="140" align="center" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.alert_status === 0"
              type="primary" link size="small" @click="handleProcess(row)"
            >
              处理
            </el-button>
            <span v-else class="text-muted">已处理</span>
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

    <!-- 处理弹窗 -->
    
<el-dialog v-dialog-drag
      v-model="processDialogVisible"
      title="处理库存预警"
      width="520px"
      destroy-on-close
    >
      <div class="process-info">
        <p><strong>商品：</strong>{{ currentAlert?.product?.name }}</p>
        <p><strong>SKU：</strong>{{ currentAlert?.sku?.sku_code }}</p>
        <p><strong>仓库：</strong>{{ currentAlert?.warehouse?.warehouse_name }}</p>
        <p><strong>当前库存：</strong>{{ currentAlert?.current_stock }}</p>
        <p><strong>预警阈值：</strong>{{ currentAlert?.warning_stock }}</p>
      </div>
      <el-divider />
      <el-form
        ref="processFormRef"
        :model="processForm"
        :rules="processFormRules"
        label-width="80px"
      >
        <el-form-item label="处理方式" prop="action">
          <el-select v-model="processForm.action" placeholder="请选择处理方式" style="width: 100%">
            <el-option label="创建采购单" value="purchase" />
            <el-option label="仓库调拨" value="transfer" />
            <el-option label="调整阈值" value="adjust_threshold" />
            <el-option label="暂时忽略" value="ignore" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input
            v-model="processForm.remark"
            type="textarea"
            :rows="3"
            placeholder="请输入处理备注"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="processDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="processLoading" @click="handleProcessSubmit">
          确认处理
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getStockAlerts, handleAlert } from '@/api/inventory'
import { getWarehouseList } from '@/api/warehouse'

// ==================== 搜索与列表 ====================
const loading = ref(false)
const alertList = ref([])
const warehouseOptions = ref([])
const pendingCount = ref(0)

const searchForm = reactive({
  status: '',
  warehouse_id: '',
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
    if (searchForm.status) params.status = searchForm.status
    if (searchForm.warehouse_id) params.warehouse_id = searchForm.warehouse_id

    const res = await getStockAlerts(params)
    alertList.value = res.data?.list || res.data || []
    pagination.total = res.data?.total || 0
    pendingCount.value = res.data?.pending_count || alertList.value.filter((a) => a.alert_status === 0).length
  } catch (error) {
    console.error('获取预警列表失败:', error)
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
  searchForm.status = ''
  searchForm.warehouse_id = ''
  pagination.page = 1
  fetchList()
}

const tableRowClassName = ({ row }) => {
  if (row.alert_status === 0) return 'alert-row'
  return ''
}

// ==================== 处理预警 ====================
const processDialogVisible = ref(false)
const processLoading = ref(false)
const processFormRef = ref(null)
const currentAlert = ref(null)

const processForm = reactive({
  action: '',
  remark: '',
})

const processFormRules = {
  action: [{ required: true, message: '请选择处理方式', trigger: 'change' }],
}

const handleProcess = (row) => {
  currentAlert.value = row
  processForm.action = ''
  processForm.remark = ''
  processDialogVisible.value = true
}

const handleProcessSubmit = async () => {
  const valid = await processFormRef.value?.validate().catch(() => false)
  if (!valid) return

  processLoading.value = true
  try {
    await handleAlert(currentAlert.value.id, {
      action: processForm.action,
      remark: processForm.remark,
    })
    ElMessage.success('处理成功')
    processDialogVisible.value = false
    fetchList()
  } catch (error) {
    console.error('处理预警失败:', error)
  } finally {
    processLoading.value = false
  }
}

// ==================== 初始化 ====================
onMounted(() => {
  fetchList()
  fetchWarehouseOptions()
})
</script>

<style lang="scss" scoped>
.alert-manage {
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

  .stock-num {
    color: #f56c6c;
    font-weight: 600;
  }

  .text-muted {
    color: #909399;
    font-size: 13px;
  }

  .process-info {
    p {
      margin: 8px 0;
      font-size: 14px;
      color: #606266;
    }
  }
}

:deep(.alert-row) {
  background-color: #fef0f0 !important;
}
</style>
