<template>
  <div class="receipt-manage">
    <!-- 搜索栏 -->
    <el-card class="search-card" shadow="never">
      <el-form :model="searchForm" inline>
        <el-form-item label="回货单号">
          <el-input
            v-model="searchForm.keyword"
            placeholder="请输入回货单号"
            clearable
            style="width: 200px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="供应商">
          <el-select
            v-model="searchForm.supplier_id"
            placeholder="全部供应商"
            clearable
            style="width: 180px"
          >
            <el-option
              v-for="item in supplierOptions"
              :key="item.id"
              :label="item.supplier_name"
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
            <el-option label="待审核" :value="0" />
            <el-option label="已审核" :value="1" />
            <el-option label="已入库" :value="2" />
            <el-option label="已取消" :value="3" />
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
          <span class="title">回货单管理</span>
          <el-button type="primary" icon="Plus" @click="handleAdd">新建回货单</el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="receiptList"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column prop="receipt_no" label="回货单号" width="235" />
        <el-table-column label="供应商" min-width="300">
          <template #default="{ row }">
            {{ row.supplier_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="发货金额" width="120" align="right">
          <template #default="{ row }">
            <span class="price">¥{{ row.total_amount || 0 }}</span>
          </template>
        </el-table-column>
        <el-table-column label="发货数量" width="100" align="center">
          <template #default="{ row }">
            {{ row.total_quantity || 0 }}
          </template>
        </el-table-column>
        <el-table-column label="已入库数量" width="120" align="center">
          <template #default="{ row }">
            {{ getReceivedQuantity(row) }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusTag(row.status)" size="small">
              {{ getStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleDetail(row)">
              详情
            </el-button>
            <el-button
              v-if="row.status === 1"
              type="warning" link size="small" @click="handleReceive(row)"
            >
              入库
            </el-button>
            <el-button
              v-if="row.status === 0"
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

    <!-- 新建回货单弹窗 -->
    <el-dialog
      v-model="formDialogVisible"
      title="新建回货单"
      width="900px"
      destroy-on-close
      top="5vh"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item label="回货单号" prop="receipt_no">
          <el-input
            v-model="formData.receipt_no"
            placeholder="请输入回货单号"
            style="width: 100%"
            maxlength="32"
          />
        </el-form-item>
        <el-form-item label="供应商" prop="supplier_id">
          <el-select v-model="formData.supplier_id" placeholder="请选择供应商" style="width: 100%">
            <el-option
              v-for="item in supplierOptions"
              :key="item.id"
              :label="item.supplier_name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="formData.remark" type="textarea" :rows="2" placeholder="请输入备注" />
        </el-form-item>

        <el-divider content-position="left">
          商品明细
          <el-button type="primary" link size="small" style="margin-left: 12px" @click="handleAddItem">
            <el-icon><Plus /></el-icon> 添加商品
          </el-button>
        </el-divider>

        <el-table :data="formData.items" border size="small" style="width: 100%">
          <el-table-column label="采购明细" min-width="280">
            <template #default="{ row }">
              <div v-if="row.product_name || row.sku_code" class="sku-display">
                <span class="sku-code">{{ row.sku_code || row.product_name }}</span>
                <span v-if="row.sku_name" class="sku-name">{{ row.sku_name }}</span>
                <span v-if="row.brand_style" class="product-name">({{ row.brand_style }})</span>
              </div>
              <el-select
                v-else
                v-model="row.purchase_item_id"
                filterable
                remote
                reserve-keyword
                placeholder="搜索采购商品"
                :remote-method="searchPurchaseItems"
                :loading="purchaseItemLoading"
                size="small"
                style="width: 100%"
                @change="(val) => handlePurchaseItemSelect(val, row)"
              >
                <el-option
                  v-for="item in purchaseItemOptions"
                  :key="item.id"
                  :label="`${item.sku_code} - ${item.sku_name} (${item.brand_style || '未知款式'}) 剩余:${item.remaining_quantity}`"
                  :value="item.id"
                />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="发货数量" width="120">
            <template #default="{ row }">
              <el-input-number v-model="row.ship_quantity" :min="1" size="small" controls-position="right" style="width: 100%" />
            </template>
          </el-table-column>
          <el-table-column label="成本价" width="120">
            <template #default="{ row }">
              <el-input-number v-model="row.cost_price" :min="0" :precision="2" size="small" controls-position="right" style="width: 100%" />
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

    <!-- 入库确认弹窗 -->
    <el-dialog
      v-model="receiveDialogVisible"
      title="确认入库"
      width="500px"
      destroy-on-close
    >
      <el-form
        ref="receiveFormRef"
        :model="receiveForm"
        :rules="receiveRules"
        label-width="100px"
      >
        <el-form-item label="回货单号">
          <span>{{ currentReceipt?.receipt_no }}</span>
        </el-form-item>
        <el-form-item label="入库仓库" prop="warehouse_id">
          <el-select v-model="receiveForm.warehouse_id" placeholder="请选择入库仓库" style="width: 100%">
            <el-option
              v-for="item in warehouseOptions"
              :key="item.id"
              :label="item.warehouse_name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="入库明细">
          <el-table :data="receiveItems" border size="small" style="width: 100%">
            <el-table-column prop="sku_code" label="SKU" width="120" />
            <el-table-column prop="product_name" label="商品名称" width="150" />
            <el-table-column label="可入库数量" width="100" align="center">
              <template #default="{ row }">
                {{ row.remaining }}
              </template>
            </el-table-column>
            <el-table-column label="入库数量" width="100">
              <template #default="{ row }">
                <el-input-number v-model="row.receive_quantity" :min="0" :max="row.remaining" size="small" controls-position="right" style="width: 100%" />
              </template>
            </el-table-column>
          </el-table>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="receiveForm.remark" type="textarea" :rows="2" placeholder="入库备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="receiveDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="receiveLoading" @click="handleReceiveSubmit">
          确认入库
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  getReceiptList,
  getReceiptDetail,
  createReceipt,
  approveReceipt,
  receiveReceipt,
  cancelReceipt,
} from '@/api/receipt'
import { getSupplierList } from '@/api/supplier'
import { getWarehouseList } from '@/api/warehouse'
import { getPurchaseList, getPurchaseDetail } from '@/api/purchase'

const router = useRouter()
const route = useRoute()

// ==================== 搜索与列表 ====================
const loading = ref(false)
const receiptList = ref([])
const supplierOptions = ref([])
const warehouseOptions = ref([])
const purchaseItemOptions = ref([])
const purchaseItemLoading = ref(false)

const searchForm = reactive({
  keyword: '',
  supplier_id: '',
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
    if (searchForm.keyword) params.keyword = searchForm.keyword
    if (searchForm.supplier_id) params.supplier_id = searchForm.supplier_id
    if (searchForm.status) params.status = searchForm.status

    const res = await getReceiptList(params)
    receiptList.value = res.data?.list || res.data || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取回货单列表失败:', error)
  } finally {
    loading.value = false
  }
}

const fetchSupplierOptions = async () => {
  try {
    const res = await getSupplierList({ page_size: 100 })
    supplierOptions.value = res.data?.list || res.data || []
  } catch (error) {
    console.error('获取供应商列表失败:', error)
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

const searchPurchaseItems = async (query) => {
  if (query.length < 2) return
  purchaseItemLoading.value = true
  try {
    const res = await getPurchaseList({ keyword: query, page_size: 20, status: 1, with_items: true })
    const purchases = res.data?.list || res.data || []
    
    const items = []
    purchases.forEach(purchase => {
      if (purchase.items) {
        purchase.items.forEach(item => {
          const remaining = (item.quantity || 0) - (item.received_quantity || 0)
          if (remaining > 0) {
            // 组合品牌和款式
            const brand = item.brand || item.sku?.product?.brand || ''
            const style = item.sku?.product?.style || ''
            const brandStyle = item.brand_style || ((brand && style) ? `${brand}${style}` : (brand || style) || '')
            
            items.push({
              id: item.id,
              purchase_order_id: purchase.id,
              purchase_no: purchase.purchase_no,
              sku_id: item.sku_id,
              sku_code: item.sku_code || item.sku?.sku_code || '',
              product_name: item.product_name || item.sku?.product?.product_name || '',
              sku_name: item.sku_name || item.sku?.sku_name || '',
              brand: brand,
              brand_style: brandStyle,
              cost_price: Number(item.purchase_price) || 0,
              remaining_quantity: remaining,
            })
          }
        })
      }
    })
    purchaseItemOptions.value = items
  } catch (error) {
    console.error('搜索采购商品失败:', error)
  } finally {
    purchaseItemLoading.value = false
  }
}

const handlePurchaseItemSelect = (itemId, row) => {
  const selectedItem = purchaseItemOptions.value.find(i => i.id === itemId)
  if (selectedItem) {
    row.purchase_item_id = selectedItem.id
    row.sku_id = selectedItem.sku_id
    row.product_name = selectedItem.product_name
    row.sku_code = selectedItem.sku_code
    row.sku_name = selectedItem.sku_name
    row.brand_style = selectedItem.brand_style
    row.cost_price = selectedItem.cost_price
    row.ship_quantity = selectedItem.remaining_quantity
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchList()
}

const handleReset = () => {
  searchForm.keyword = ''
  searchForm.supplier_id = ''
  searchForm.status = ''
  pagination.page = 1
  fetchList()
}

const getStatusLabel = (status) => {
  const map = { 0: '待审核', 1: '已审核', 2: '已入库', 3: '已取消' }
  return map[status] ?? status ?? '未知'
}

const getStatusTag = (status) => {
  const map = { 0: 'warning', 1: '', 2: 'success', 3: 'info' }
  return map[status] || 'info'
}

const formatTime = (time) => {
  if (!time) return '-'
  return time.replace('T', ' ').substring(0, 19)
}

const getReceivedQuantity = (row) => {
  if (!row.items || row.items.length === 0) return 0
  return row.items.reduce((sum, item) => sum + (item.receive_quantity || 0), 0)
}

// ==================== 新建回货单 ====================
const formDialogVisible = ref(false)
const submitLoading = ref(false)
const formRef = ref(null)

const formData = reactive({
  receipt_no: '',
  supplier_id: '',
  remark: '',
  items: [],
})

const formRules = {
  receipt_no: [{ required: true, message: '请输入回货单号', trigger: 'blur' }],
  supplier_id: [{ required: true, message: '请选择供应商', trigger: 'change' }],
}

const handleAdd = () => {
  formData.receipt_no = ''
  formData.supplier_id = ''
  formData.remark = ''
  formData.items = []
  if (supplierOptions.value.length === 0) {
    fetchSupplierOptions()
  }
  formDialogVisible.value = true
}

const handleAddItem = () => {
  formData.items.push({
    purchase_item_id: '',
    sku_id: '',
    product_name: '',
    sku_code: '',
    sku_name: '',
    brand_style: '',
    ship_quantity: 1,
    cost_price: 0,
  })
}

const handleRemoveItem = (index) => {
  formData.items.splice(index, 1)
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  if (formData.items.length === 0) {
    ElMessage.warning('请至少添加一条商品明细')
    return
  }

  submitLoading.value = true
  try {
    const selectedSupplier = supplierOptions.value.find(s => s.id === formData.supplier_id)
    const payload = {
      receipt_no: formData.receipt_no,
      supplier_id: formData.supplier_id,
      supplier_name: selectedSupplier?.supplier_name || '',
      remark: formData.remark,
      items: formData.items.map(item => ({
        purchase_item_id: item.purchase_item_id || undefined,
        sku_id: item.sku_id,
        product_name: item.product_name || '',
        sku_name: item.sku_name || '',
        sku_code: item.sku_code || '',
        brand_style: item.brand_style || '',
        ship_quantity: item.ship_quantity,
        cost_price: item.cost_price,
      })),
    }

    await createReceipt(payload)
    ElMessage.success('回货单创建成功')
    formDialogVisible.value = false
    fetchList()
  } catch (error) {
    console.error('创建回货单失败:', error)
  } finally {
    submitLoading.value = false
  }
}

// ==================== 入库确认 ====================
const receiveDialogVisible = ref(false)
const receiveLoading = ref(false)
const receiveFormRef = ref(null)
const currentReceipt = ref(null)
const receiveItems = ref([])

const receiveForm = reactive({
  warehouse_id: '',
  remark: '',
})

const receiveRules = {
  warehouse_id: [{ required: true, message: '请选择入库仓库', trigger: 'change' }],
}

const handleReceive = async (row) => {
  currentReceipt.value = row
  receiveForm.warehouse_id = ''
  receiveForm.remark = ''
  
  // 获取回货单详情，确保包含 items
  try {
    const res = await getReceiptDetail(row.id)
    const detail = res.data || {}
    
    receiveItems.value = (detail.items || []).map(item => ({
      receipt_item_id: item.id,
      sku_code: item.sku_code || '',
      product_name: item.product_name || '',
      remaining: item.ship_quantity - (item.receive_quantity || 0),
      receive_quantity: item.ship_quantity - (item.receive_quantity || 0),
    }))
  } catch (error) {
    console.error('获取回货单详情失败:', error)
    receiveItems.value = (row.items || []).map(item => ({
      receipt_item_id: item.id,
      sku_code: item.sku_code || '',
      product_name: item.product_name || '',
      remaining: item.ship_quantity - (item.receive_quantity || 0),
      receive_quantity: item.ship_quantity - (item.receive_quantity || 0),
    }))
  }
  
  if (warehouseOptions.value.length === 0) {
    fetchWarehouseOptions()
  }
  receiveDialogVisible.value = true
}

const handleReceiveSubmit = async () => {
  const valid = await receiveFormRef.value?.validate().catch(() => false)
  if (!valid) return

  const items = receiveItems.value.filter(item => item.receive_quantity > 0)
  if (items.length === 0) {
    ElMessage.warning('请至少选择一个商品入库')
    return
  }

  receiveLoading.value = true
  try {
    await receiveReceipt(currentReceipt.value.id, {
      warehouse_id: receiveForm.warehouse_id,
      remark: receiveForm.remark,
      items: items.map(item => ({
        receipt_item_id: item.receipt_item_id,
        receive_quantity: item.receive_quantity,
      })),
    })
    ElMessage.success('入库成功')
    receiveDialogVisible.value = false
    fetchList()
  } catch (error) {
    console.error('入库失败:', error)
  } finally {
    receiveLoading.value = false
  }
}

// ==================== 操作 ====================
const handleDetail = (row) => {
  router.push(`/inventory/receipt/detail/${row.id}`)
}

const handleCancel = (row) => {
  ElMessageBox.confirm(
    `确定要取消回货单 "${row.receipt_no}" 吗？`,
    '取消确认',
    {
      confirmButtonText: '确定取消',
      cancelButtonText: '返回',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await cancelReceipt(row.id)
      ElMessage.success('取消成功')
      fetchList()
    } catch (error) {
      console.error('取消失败:', error)
    }
  }).catch(() => {})
}

// ==================== 初始化 ====================
onMounted(() => {
  fetchList()
  fetchSupplierOptions()
  fetchWarehouseOptions()

  // 检查是否从采购单跳转过来
  const purchaseId = route.query.purchase_id
  const supplierId = route.query.supplier_id
  const purchaseNo = route.query.purchase_no
  
  if (purchaseId && supplierId) {
    // 自动打开新建回货单弹窗，并填充供应商信息
    handleAddWithPurchase(purchaseId, supplierId, purchaseNo)
  }
})

const handleAddWithPurchase = async (purchaseId, supplierId, purchaseNo) => {
  formData.receipt_no = ''
  formData.supplier_id = Number(supplierId)
  formData.remark = purchaseNo ? `关联采购单[${purchaseNo}]` : ''
  formData.items = []
  
  // 获取采购单详情，自动填充商品明细
  try {
    const res = await getPurchaseDetail(purchaseId)
    const purchase = res.data || {}
    if (purchase.items && purchase.items.length > 0) {
      purchase.items.forEach(item => {
        const remaining = (item.quantity || 0) - (item.received_quantity || 0)
        if (remaining > 0) {
          // 组合品牌和款式
          const brand = item.brand || item.sku?.product?.brand || ''
          const style = item.sku?.product?.style || ''
          const brandStyle = item.brand_style || ((brand && style) ? `${brand}${style}` : (brand || style) || '')
          
          formData.items.push({
            purchase_item_id: item.id,
            sku_id: item.sku_id,
            product_name: item.product_name || item.sku?.product?.product_name || '',
            sku_code: item.sku?.sku_code || item.sku_code || '',
            sku_name: item.sku_name || item.sku?.sku_name || '',
            brand_style: brandStyle,
            ship_quantity: remaining,
            cost_price: Number(item.purchase_price) || 0,
          })
        }
      })
    }
  } catch (error) {
    console.error('获取采购单详情失败:', error)
  }
  
  formDialogVisible.value = true
  // 清除URL参数
  router.replace({ query: {} })
}
</script>

<style lang="scss" scoped>
.receipt-manage {
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

  .sku-display {
    padding: 4px 8px;
    background: #f5f7fa;
    border-radius: 4px;
    font-size: 13px;

    .sku-code {
      font-weight: 500;
      color: #303133;
    }

    .sku-name {
      color: #606266;
      margin-left: 8px;
      &::before {
        content: '-';
        margin-right: 8px;
        color: #dcdfe6;
      }
    }

    .product-name {
      color: #909399;
      margin-left: 4px;
    }
  }
}
</style>