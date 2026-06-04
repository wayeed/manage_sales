<template>
  <div class="stocktaking-manage">
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
        <el-form-item label="状态">
          <el-select
            v-model="searchForm.status"
            placeholder="全部状态"
            clearable
            style="width: 140px"
          >
            <el-option label="盘点中" :value="0" />
            <el-option label="已提交" :value="1" />
            <el-option label="已审核" :value="2" />
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
          <span class="title">库存盘点</span>
          <el-button type="primary" icon="Plus" @click="handleCreate">新建盘点</el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="stocktakeList"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column label="盘点单号" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.stocktake_no || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="仓库" width="140">
          <template #default="{ row }">
            {{ row.warehouse?.warehouse_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.status === 0" type="warning">盘点中</el-tag>
            <el-tag v-else-if="row.status === 1" type="primary">已提交</el-tag>
            <el-tag v-else-if="row.status === 2" type="success">已审核</el-tag>
            <el-tag v-else type="info">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="商品数" width="100" align="center">
          <template #default="{ row }">
            {{ row.total_items ?? 0 }}
          </template>
        </el-table-column>
        <el-table-column label="盘盈数" width="100" align="center">
          <template #default="{ row }">
            <span class="profit-text">{{ row.profit_items ?? 0 }}</span>
          </template>
        </el-table-column>
        <el-table-column label="盘亏数" width="100" align="center">
          <template #default="{ row }">
            <span class="loss-text">{{ row.loss_items ?? 0 }}</span>
          </template>
        </el-table-column>
        <el-table-column label="创建人" width="120">
          <template #default="{ row }">
            {{ row.creator?.username || row.creator?.name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right" align="center">
          <template #default="{ row }">
            <template v-if="row.status === 0">
              <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
              <el-button type="warning" link size="small" @click="handleSubmit(row)">提交</el-button>
              <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
            </template>
            <template v-else-if="row.status === 1">
              <el-button type="success" link size="small" @click="handleApprove(row)">审核通过</el-button>
              <el-button type="danger" link size="small" @click="handleReject(row)">驳回</el-button>
            </template>
            <template v-else>
              <span class="no-action">-</span>
            </template>
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

    <!-- 新建盘点弹窗 -->
    <el-dialog
      v-model="createDialogVisible"
      title="新建盘点"
      width="480px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-width="80px"
      >
        <el-form-item label="仓库" prop="warehouse_id">
          <el-select
            v-model="createForm.warehouse_id"
            placeholder="请选择仓库"
            style="width: 100%"
          >
            <el-option
              v-for="item in warehouseOptions"
              :key="item.id"
              :label="item.warehouse_name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input
            v-model="createForm.remark"
            type="textarea"
            :rows="3"
            placeholder="请输入备注信息"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="createLoading" @click="handleCreateSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 编辑盘点弹窗 -->
    <el-dialog
      v-model="editDialogVisible"
      title="编辑盘点"
      width="900px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <div class="edit-dialog-header">
        <span>盘点单号：{{ editForm.stocktake_no }}</span>
        <span>仓库：{{ editForm.warehouse_name }}</span>
      </div>
      <el-table
        v-loading="editLoading"
        :data="editForm.items"
        border
        stripe
        style="width: 100%"
        max-height="460"
      >
        <el-table-column label="商品名称" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.sku?.product?.product_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="SKU编码" width="150">
          <template #default="{ row }">
            {{ row.sku?.sku_code || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="系统库存" width="110" align="center">
          <template #default="{ row }">
            {{ row.system_stock ?? 0 }}
          </template>
        </el-table-column>
        <el-table-column label="实际库存" width="130" align="center">
          <template #default="{ row }">
            <el-input-number
              v-model="row.actual_stock"
              :min="0"
              size="small"
              controls-position="right"
              style="width: 100px"
            />
          </template>
        </el-table-column>
        <el-table-column label="差异" width="100" align="center">
          <template #default="{ row }">
            <span :class="getDiffClass(row)">
              {{ getDiffValue(row) }}
            </span>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editSaving" @click="handleEditSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getStocktakeList,
  getStocktakeDetail,
  createStocktake,
  updateStocktake,
  submitStocktake,
  approveStocktake,
  deleteStocktake,
  getWarehouseList,
} from '@/api/inventory'

// ==================== 工具函数 ====================
const formatTime = (time) => {
  if (!time) return '-'
  return time.replace('T', ' ').substring(0, 19)
}

// ==================== 搜索与列表 ====================
const loading = ref(false)
const stocktakeList = ref([])
const warehouseOptions = ref([])

const searchForm = reactive({
  warehouse_id: '',
  status: '',
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
    if (searchForm.status !== '' && searchForm.status !== null) params.status = searchForm.status

    const res = await getStocktakeList(params)
    stocktakeList.value = res.data?.list || res.data || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取盘点列表失败:', error)
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
  searchForm.status = ''
  pagination.page = 1
  fetchList()
}

// ==================== 新建盘点 ====================
const createDialogVisible = ref(false)
const createLoading = ref(false)
const createFormRef = ref(null)
const createForm = reactive({
  warehouse_id: '',
  remark: '',
})

const createRules = {
  warehouse_id: [{ required: true, message: '请选择仓库', trigger: 'change' }],
}

const handleCreate = () => {
  createForm.warehouse_id = ''
  createForm.remark = ''
  createDialogVisible.value = true
}

const handleCreateSubmit = async () => {
  if (!createFormRef.value) return
  await createFormRef.value.validate(async (valid) => {
    if (!valid) return
    createLoading.value = true
    try {
      await createStocktake({
        warehouse_id: createForm.warehouse_id,
        remark: createForm.remark,
      })
      ElMessage.success('新建盘点成功')
      createDialogVisible.value = false
      fetchList()
    } catch (error) {
      console.error('新建盘点失败:', error)
    } finally {
      createLoading.value = false
    }
  })
}

// ==================== 编辑盘点 ====================
const editDialogVisible = ref(false)
const editLoading = ref(false)
const editSaving = ref(false)
const editForm = reactive({
  id: null,
  stocktake_no: '',
  warehouse_name: '',
  items: [],
})

const handleEdit = async (row) => {
  editForm.id = row.id
  editForm.stocktake_no = row.stocktake_no || ''
  editForm.warehouse_name = row.warehouse?.warehouse_name || ''
  editForm.items = []
  editDialogVisible.value = true
  editLoading.value = true
  try {
    const res = await getStocktakeDetail(row.id)
    const detail = res.data || {}
    editForm.items = (detail.items || []).map((item) => ({
      ...item,
    }))
  } catch (error) {
    console.error('获取盘点详情失败:', error)
    ElMessage.error('获取盘点详情失败')
    editDialogVisible.value = false
  } finally {
    editLoading.value = false
  }
}

const getDiffValue = (row) => {
  const diff = (row.actual_stock ?? 0) - (row.system_stock ?? 0)
  if (diff > 0) return `+${diff}`
  if (diff < 0) return `${diff}`
  return '0'
}

const getDiffClass = (row) => {
  const diff = (row.actual_stock ?? 0) - (row.system_stock ?? 0)
  if (diff > 0) return 'profit-text'
  if (diff < 0) return 'loss-text'
  return ''
}

const handleEditSave = async () => {
  editSaving.value = true
  try {
    const items = editForm.items.map((item) => ({
      sku_id: item.sku_id,
      system_stock: item.system_stock,
      actual_stock: item.actual_stock ?? 0,
    }))
    await updateStocktake(editForm.id, { items })
    ElMessage.success('保存成功')
    editDialogVisible.value = false
    fetchList()
  } catch (error) {
    console.error('保存盘点失败:', error)
  } finally {
    editSaving.value = false
  }
}

// ==================== 提交盘点 ====================
const handleSubmit = (row) => {
  ElMessageBox.confirm('确认提交该盘点单？提交后将无法修改。', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    try {
      await submitStocktake(row.id)
      ElMessage.success('提交成功')
      fetchList()
    } catch (error) {
      console.error('提交盘点失败:', error)
    }
  }).catch(() => {})
}

// ==================== 审核通过 ====================
const handleApprove = (row) => {
  ElMessageBox.confirm('确认审核通过该盘点单？审核后将自动调整库存。', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    try {
      await approveStocktake(row.id, { approved: true })
      ElMessage.success('审核通过')
      fetchList()
    } catch (error) {
      console.error('审核失败:', error)
    }
  }).catch(() => {})
}

// ==================== 驳回 ====================
const handleReject = (row) => {
  ElMessageBox.confirm('确认驳回该盘点单？驳回后状态将变为盘点中，可重新编辑。', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    try {
      await approveStocktake(row.id, { approved: false })
      ElMessage.success('已驳回')
      fetchList()
    } catch (error) {
      console.error('驳回失败:', error)
    }
  }).catch(() => {})
}

// ==================== 删除盘点 ====================
const handleDelete = (row) => {
  ElMessageBox.confirm('确认删除该盘点单？删除后不可恢复。', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    try {
      await deleteStocktake(row.id)
      ElMessage.success('删除成功')
      fetchList()
    } catch (error) {
      console.error('删除盘点失败:', error)
    }
  }).catch(() => {})
}

// ==================== 初始化 ====================
onMounted(() => {
  fetchList()
  fetchWarehouseOptions()
})
</script>

<style lang="scss" scoped>
.stocktaking-manage {
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

  .profit-text {
    color: #67c23a;
    font-weight: 600;
  }

  .loss-text {
    color: #f56c6c;
    font-weight: 600;
  }

  .no-action {
    color: #c0c4cc;
  }

  .edit-dialog-header {
    display: flex;
    gap: 24px;
    margin-bottom: 16px;
    font-size: 14px;
    color: #606266;
  }
}
</style>
