<template>
  <div class="gift-manage">
    <!-- 搜索栏 -->
    <el-card class="search-card" shadow="never">
      <el-form :model="searchForm" inline>
        <el-form-item label="关键词">
          <el-input
            v-model="searchForm.keyword"
            placeholder="礼品名称/编码"
            clearable
            style="width: 200px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select
            v-model="searchForm.status"
            placeholder="全部状态"
            clearable
            style="width: 120px"
          >
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
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
              v-for="w in warehouseOptions"
              :key="w.id"
              :label="w.warehouse_name"
              :value="w.id"
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
          <span class="title">礼品管理</span>
          <el-button type="primary" icon="Plus" @click="handleAdd">新增礼品</el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="giftList"
        :key="searchForm.warehouse_id || 'all'"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column prop="gift_code" label="礼品编码" width="140" />
        <el-table-column prop="gift_name" label="礼品名称" min-width="180" />
        <el-table-column label="成本价" width="100" align="right">
          <template #default="{ row }">
            <span class="price">¥{{ row.cost_price || 0 }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="stock_quantity" label="总库存" width="100" align="center" />
        <el-table-column v-if="searchForm.warehouse_id" label="仓库" width="140" align="center">
          <template #default="{ row }">
            <span>{{ row.warehouse_name || '--' }}</span>
          </template>
        </el-table-column>
        <el-table-column v-if="searchForm.warehouse_id" label="仓库库存" width="100" align="center">
          <template #default="{ row }">
            <span :class="{ 'text-warning': (row.warehouse_stock || 0) <= 0 }">
              {{ row.warehouse_stock || 0 }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="220" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button type="success" link size="small" @click="handleStock(row)">
              入库
            </el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">
              删除
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

    <!-- 新增/编辑弹窗 -->
    
<el-dialog v-dialog-drag
      v-model="formDialogVisible"
      :title="isEdit ? '编辑礼品' : '新增礼品'"
      width="520px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="80px"
      >
        <el-form-item label="礼品编码" prop="gift_code">
          <el-input v-model="formData.gift_code" placeholder="请输入礼品编码" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="礼品名称" prop="gift_name">
          <el-input v-model="formData.gift_name" placeholder="请输入礼品名称" />
        </el-form-item>
        <el-form-item label="成本价" prop="cost_price">
          <el-input-number
            v-model="formData.cost_price"
            :min="0"
            :precision="2"
            controls-position="right"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 入库弹窗 -->
    <el-dialog
      v-model="stockDialogVisible"
      title="礼品入库"
      width="420px"
      destroy-on-close
    >
      <el-form
        ref="stockFormRef"
        :model="stockFormData"
        :rules="stockFormRules"
        label-width="80px"
      >
        <el-form-item label="礼品名称">
          <span>{{ currentGift?.gift_name }}</span>
        </el-form-item>
        <el-form-item label="当前库存">
          <span>{{ currentGift?.stock_quantity || 0 }}</span>
        </el-form-item>
        <el-form-item label="入库仓库" prop="warehouse_id">
          <el-select
            v-model="stockFormData.warehouse_id"
            placeholder="请选择仓库"
            style="width: 100%"
          >
            <el-option
              v-for="warehouse in warehouseList"
              :key="warehouse.id"
              :label="warehouse.warehouse_name"
              :value="warehouse.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="入库数量" prop="quantity">
          <el-input-number
            v-model="stockFormData.quantity"
            :min="1"
            :precision="0"
            controls-position="right"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="采购单价" prop="purchase_price">
          <el-input-number
            v-model="stockFormData.purchase_price"
            :min="0"
            :precision="2"
            controls-position="right"
            style="width: 100%"
            placeholder="选填"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="stockDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="stockSubmitLoading" @click="handleStockSubmit">
          确定入库
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getGiftList,
  createGift,
  updateGift,
  deleteGift,
  addGiftStock,
} from '@/api/gift'
import { getWarehouseList } from '@/api/warehouse'

// ==================== 搜索与列表 ====================
const loading = ref(false)
const giftList = ref([])
const warehouseOptions = ref([])

const searchForm = reactive({
  keyword: '',
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
    if (searchForm.keyword) params.keyword = searchForm.keyword
    if (searchForm.status !== '' && searchForm.status !== null) params.status = searchForm.status
    if (searchForm.warehouse_id) params.warehouse_id = searchForm.warehouse_id

    const res = await getGiftList(params)
    giftList.value = res.data?.list || res.data || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取礼品列表失败:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchList()
}

const handleReset = () => {
  searchForm.keyword = ''
  searchForm.status = ''
  searchForm.warehouse_id = ''
  pagination.page = 1
  fetchList()
}

// ==================== 新增/编辑 ====================
const formDialogVisible = ref(false)
const isEdit = ref(false)
const submitLoading = ref(false)
const formRef = ref(null)

const formData = reactive({
  id: null,
  gift_code: '',
  gift_name: '',
  cost_price: 0,
  status: 1,
})

const formRules = {
  gift_code: [{ required: true, message: '请输入礼品编码', trigger: 'blur' }],
  gift_name: [{ required: true, message: '请输入礼品名称', trigger: 'blur' }],
}

const resetForm = () => {
  formData.id = null
  formData.gift_code = ''
  formData.gift_name = ''
  formData.cost_price = 0
  formData.status = 1
}

const handleAdd = () => {
  isEdit.value = false
  resetForm()
  formDialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  formData.id = row.id
  formData.gift_code = row.gift_code
  formData.gift_name = row.gift_name
  formData.cost_price = row.cost_price || 0
  formData.status = row.status
  formDialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    const data = {
      gift_code: formData.gift_code,
      gift_name: formData.gift_name,
      cost_price: formData.cost_price,
      status: formData.status,
    }
    if (isEdit.value) {
      await updateGift(formData.id, data)
      ElMessage.success('更新成功')
    } else {
      await createGift(data)
      ElMessage.success('创建成功')
    }
    formDialogVisible.value = false
    fetchList()
  } catch (error) {
    console.error('保存礼品失败:', error)
  } finally {
    submitLoading.value = false
  }
}

// ==================== 删除 ====================
const handleDelete = (row) => {
  ElMessageBox.confirm(
    `确定要删除礼品 "${row.gift_name}" 吗？此操作不可恢复。`,
    '删除确认',
    {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await deleteGift(row.id)
      ElMessage.success('删除成功')
      fetchList()
    } catch (error) {
      console.error('删除礼品失败:', error)
    }
  }).catch(() => {})
}

// ==================== 入库 ====================
const stockDialogVisible = ref(false)
const stockSubmitLoading = ref(false)
const stockFormRef = ref(null)
const currentGift = ref(null)
const warehouseList = ref([])

const stockFormData = reactive({
  warehouse_id: null,
  quantity: 1,
  purchase_price: 0,
})

const stockFormRules = {
  warehouse_id: [
    { required: true, message: '请选择入库仓库', trigger: 'change' },
  ],
  quantity: [
    { required: true, message: '请输入入库数量', trigger: 'blur' },
    { type: 'number', min: 1, message: '入库数量必须大于0', trigger: 'blur' },
  ],
}

const resetStockForm = () => {
  stockFormData.warehouse_id = null
  stockFormData.quantity = 1
  stockFormData.purchase_price = 0
}

// 加载仓库列表
const fetchWarehouseList = async () => {
  try {
    const res = await getWarehouseList({ page: 1, page_size: 100 })
    warehouseList.value = res.data?.list || res.data || []
    // 如果只有一个仓库，默认选中
    if (warehouseList.value.length === 1) {
      stockFormData.warehouse_id = warehouseList.value[0].id
    }
  } catch (error) {
    console.error('获取仓库列表失败:', error)
  }
}

const handleStock = async (row) => {
  currentGift.value = row
  await fetchWarehouseList()
  resetStockForm()
  stockDialogVisible.value = true
}

const handleStockSubmit = async () => {
  const valid = await stockFormRef.value?.validate().catch(() => false)
  if (!valid) return

  stockSubmitLoading.value = true
  try {
    await addGiftStock(currentGift.value.id, {
      warehouse_id: stockFormData.warehouse_id,
      quantity: stockFormData.quantity,
      purchase_price: stockFormData.purchase_price,
    })
    ElMessage.success('入库成功')
    stockDialogVisible.value = false
    fetchList()
  } catch (error) {
    console.error('礼品入库失败:', error)
  } finally {
    stockSubmitLoading.value = false
  }
}

// ==================== 初始化 ====================
const fetchWarehouseOptions = async () => {
  try {
    const res = await getWarehouseList({ page: 1, page_size: 100 })
    warehouseOptions.value = res.data?.list || res.data || []
  } catch (error) {
    console.error('获取仓库列表失败:', error)
  }
}

onMounted(() => {
  fetchWarehouseOptions()
  fetchList()
})
</script>

<style lang="scss" scoped>
.gift-manage {
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

  .text-warning {
    color: #e6a23c;
    font-weight: 500;
  }
}
</style>
