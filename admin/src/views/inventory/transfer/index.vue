<template>
  <div class="transfer-manage">
    <!-- 搜索栏 -->
    <el-card class="search-card" shadow="never">
      <el-form :model="searchForm" inline>
        <el-form-item label="状态">
          <el-select
            v-model="searchForm.status"
            placeholder="全部状态"
            clearable
            style="width: 140px"
          >
            <el-option label="待审核" :value="0" />
            <el-option label="已审核" :value="1" />
            <el-option label="已出库" :value="2" />
            <el-option label="已入库" :value="3" />
            <el-option label="已取消" :value="4" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input
            v-model="searchForm.keyword"
            placeholder="调拨单号"
            clearable
            style="width: 200px"
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
          <span class="title">调拨管理</span>
          <el-button type="primary" icon="Plus" @click="handleAdd">新建调拨单</el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="transferList"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column prop="transfer_no" label="调拨单号" width="160" />
        <el-table-column label="调出仓库" width="120">
          <template #default="{ row }">
            {{ row.from_warehouse?.warehouse_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="调入仓库" width="120">
          <template #default="{ row }">
            {{ row.to_warehouse?.warehouse_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="调拨商品" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            {{ getItemsSummary(row.items) }}
          </template>
        </el-table-column>
        <el-table-column prop="total_quantity" label="数量" width="80" align="center" />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusTag(row.status)" size="small">
              {{ getStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="240" align="center" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 0"
              type="success" link size="small" @click="handleApprove(row)"
            >
              审核
            </el-button>
            <el-button
              v-if="row.status === 0"
              type="danger" link size="small" @click="handleCancel(row)"
            >
              拒绝
            </el-button>
            <el-button
              v-if="row.status === 1"
              type="warning" link size="small" @click="handleConfirmOut(row)"
            >
              确认出库
            </el-button>
            <el-button
              v-if="row.status === 2"
              type="primary" link size="small" @click="handleConfirmIn(row)"
            >
              确认入库
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

    <!-- 新建调拨单弹窗 -->
    
<el-dialog v-dialog-drag
      v-model="formDialogVisible"
      title="新建调拨单"
      width="700px"
      destroy-on-close
      top="5vh"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
      >
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="调出仓库" prop="from_warehouse_id">
              <el-select v-model="formData.from_warehouse_id" placeholder="请选择调出仓库" style="width: 100%">
                <el-option
                  v-for="item in warehouseOptions"
                  :key="item.id"
                  :label="item.warehouse_name"
                  :value="item.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="调入仓库" prop="to_warehouse_id">
              <el-select v-model="formData.to_warehouse_id" placeholder="请选择调入仓库" style="width: 100%">
                <el-option
                  v-for="item in warehouseOptions"
                  :key="item.id"
                  :label="item.warehouse_name"
                  :value="item.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="备注">
          <el-input v-model="formData.remark" type="textarea" :rows="2" placeholder="请输入备注" />
        </el-form-item>

        <el-divider content-position="left">
          调拨明细
          <el-button type="primary" link size="small" style="margin-left: 12px" @click="handleAddItem">
            <el-icon><Plus /></el-icon> 添加商品
          </el-button>
        </el-divider>

        <el-table :data="formData.items" border size="small" style="width: 100%">
          <el-table-column label="SKU" min-width="280">
            <template #default="{ row }">
              <el-select
                v-model="row.sku_id"
                filterable
                remote
                reserve-keyword
                placeholder="输入SKU编码或名称搜索"
                :remote-method="searchSKU"
                :loading="skuLoading"
                size="small"
                style="width: 100%"
              >
                <el-option
                  v-for="item in skuOptions"
                  :key="item.id"
                  :label="`${item.sku_code} - ${item.sku_name} (${item.product?.product_name || '未知商品'})`"
                  :value="item.id"
                />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="数量" width="120">
            <template #default="{ row }">
              <el-input-number v-model="row.quantity" :min="1" size="small" controls-position="right" style="width: 100%" />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="80" align="center">
            <template #default="{ $index }">
              <el-button type="danger" link size="small" @click="handleRemoveItem($index)">
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-form>
      <template #footer>
        <el-button @click="formDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">
          提交
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getTransferList,
  createTransfer,
  approveTransfer,
  confirmOut,
  confirmIn,
  cancelTransfer,
} from '@/api/transfer'
import { getWarehouseList } from '@/api/warehouse'
import { getAllSkuList } from '@/api/product'

// ==================== 搜索与列表 ====================
const loading = ref(false)
const transferList = ref([])
const warehouseOptions = ref([])
const skuOptions = ref([])
const skuLoading = ref(false)

const searchForm = reactive({
  status: '',
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
    if (searchForm.status) params.status = searchForm.status
    if (searchForm.keyword) params.keyword = searchForm.keyword

    const res = await getTransferList(params)
    transferList.value = res.data?.list || res.data || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取调拨列表失败:', error)
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

// 搜索SKU
const searchSKU = async (query) => {
  if (query.length < 2) return
  skuLoading.value = true
  try {
    const res = await getAllSkuList({ keyword: query, page_size: 20 })
    skuOptions.value = res.data?.list || res.data || []
  } catch (error) {
    console.error('搜索SKU失败:', error)
  } finally {
    skuLoading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchList()
}

const handleReset = () => {
  searchForm.status = ''
  searchForm.keyword = ''
  pagination.page = 1
  fetchList()
}

const getStatusLabel = (status) => {
  const map = { 0: '待审核', 1: '已审核', 2: '已出库', 3: '已入库', 4: '已取消' }
  return map[status] ?? status ?? '未知'
}

const getStatusTag = (status) => {
  const map = { 0: 'warning', 1: '', 2: '', 3: 'success', 4: 'info' }
  return map[status] || 'info'
}

// 获取调拨商品摘要
const getItemsSummary = (items) => {
  if (!items || items.length === 0) return '-'
  return items.map(item => {
    const productName = item.sku?.product?.name || item.sku?.sku_name || '未知商品'
    return `${productName}×${item.quantity}`
  }).join(', ')
}

// ==================== 新建调拨单 ====================
const formDialogVisible = ref(false)
const submitLoading = ref(false)
const formRef = ref(null)

const formData = reactive({
  from_warehouse_id: '',
  to_warehouse_id: '',
  remark: '',
  items: [],
})

const formRules = {
  from_warehouse_id: [{ required: true, message: '请选择调出仓库', trigger: 'change' }],
  to_warehouse_id: [{ required: true, message: '请选择调入仓库', trigger: 'change' }],
}

const handleAdd = () => {
  formData.from_warehouse_id = ''
  formData.to_warehouse_id = ''
  formData.remark = ''
  formData.items = []
  formDialogVisible.value = true
}

const handleAddItem = () => {
  formData.items.push({
    sku_id: '',
    quantity: 1,
  })
}

const handleRemoveItem = (index) => {
  formData.items.splice(index, 1)
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  if (formData.from_warehouse_id === formData.to_warehouse_id) {
    ElMessage.warning('调出仓库和调入仓库不能相同')
    return
  }

  if (formData.items.length === 0) {
    ElMessage.warning('请至少添加一条调拨明细')
    return
  }

  // 验证所有明细都已选择SKU
  const invalidItems = formData.items.filter(item => !item.sku_id)
  if (invalidItems.length > 0) {
    ElMessage.warning('请选择所有调拨明细的SKU')
    return
  }

  submitLoading.value = true
  try {
    await createTransfer({
      from_warehouse_id: formData.from_warehouse_id,
      to_warehouse_id: formData.to_warehouse_id,
      remark: formData.remark,
      items: formData.items.map(item => ({
        sku_id: item.sku_id,
        quantity: item.quantity,
      })),
    })
    ElMessage.success('调拨单创建成功')
    formDialogVisible.value = false
    fetchList()
  } catch (error) {
    console.error('创建调拨单失败:', error)
  } finally {
    submitLoading.value = false
  }
}

// ==================== 操作 ====================
const handleApprove = (row) => {
  ElMessageBox.confirm(
    `确定要审核通过调拨单 "${row.transfer_no}" 吗？`,
    '审核确认',
    {
      confirmButtonText: '确定审核',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await approveTransfer(row.id)
      ElMessage.success('审核成功')
      fetchList()
    } catch (error) {
      console.error('审核失败:', error)
    }
  }).catch(() => {})
}

const handleCancel = (row) => {
  ElMessageBox.confirm(
    `确定要拒绝调拨单 "${row.transfer_no}" 吗？拒绝后将取消该调拨单。`,
    '拒绝确认',
    {
      confirmButtonText: '确定拒绝',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await cancelTransfer(row.id)
      ElMessage.success('已拒绝')
      fetchList()
    } catch (error) {
      console.error('拒绝失败:', error)
    }
  }).catch(() => {})
}

const handleConfirmOut = (row) => {
  ElMessageBox.confirm(
    `确定要确认出库调拨单 "${row.transfer_no}" 吗？`,
    '出库确认',
    {
      confirmButtonText: '确认出库',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await confirmOut(row.id)
      ElMessage.success('出库成功')
      fetchList()
    } catch (error) {
      console.error('出库失败:', error)
    }
  }).catch(() => {})
}

const handleConfirmIn = (row) => {
  ElMessageBox.confirm(
    `确定要确认入库调拨单 "${row.transfer_no}" 吗？`,
    '入库确认',
    {
      confirmButtonText: '确认入库',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await confirmIn(row.id)
      ElMessage.success('入库成功')
      fetchList()
    } catch (error) {
      console.error('入库失败:', error)
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
.transfer-manage {
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
}
</style>
