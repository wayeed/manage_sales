<template>
  <div class="purchase-manage">
    <!-- 搜索栏 -->
    <el-card class="search-card" shadow="never">
      <el-form :model="searchForm" inline>
        <el-form-item label="采购单号">
          <el-input
            v-model="searchForm.keyword"
            placeholder="请输入采购单号"
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
          <span class="title">采购管理</span>
          <el-button type="primary" icon="Plus" @click="handleAdd">新建采购单</el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="purchaseList"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column prop="purchase_no" label="采购单号" width="235" />
        <el-table-column label="供应商" min-width="200">
          <template #default="{ row }">
            {{ row.supplier_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="采购金额" width="120" align="right">
          <template #default="{ row }">
            <span class="price">¥{{ row.total_amount || 0 }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="total_quantity" label="采购数量" width="100" align="center" />
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
        <el-table-column label="操作" width="200" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleDetail(row)">
              详情
            </el-button>
            <el-button
              v-if="row.status === 0 || row.status === 1"
              type="primary" link size="small" @click="handleEdit(row)"
            >
              编辑
            </el-button>
            <el-button
              v-if="row.status === 0"
              type="success" link size="small" @click="handleApprove(row)"
            >
              审核
            </el-button>
            <el-button
              v-if="row.status === 1"
              type="warning" link size="small" @click="handleCreateReceipt(row)"
            >
              创建回货单
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

    <!-- 新建采购单弹窗 -->
    
<el-dialog v-dialog-drag
      v-model="formDialogVisible"
      :title="isEditMode ? '编辑采购单' : '新建采购单'"
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
          <el-table-column label="SKU" min-width="280">
            <template #default="{ row }">
              <!-- 有sku_code或sku_name时直接显示，否则显示选择器 -->
              <div v-if="row.sku_code || row.sku_name" class="sku-display">
                <span class="sku-code">{{ row.sku_code || row.sku_name }}</span>
                <span v-if="row.sku_code && row.sku_name && row.sku_code !== row.sku_name" class="sku-name">{{ row.sku_name }}</span>
                <span v-if="row.style" class="product-name">({{ row.style }})</span>
              </div>
              <el-select
                v-else
                v-model="row.sku_id"
                filterable
                remote
                reserve-keyword
                placeholder="输入SKU编码或名称搜索"
                :remote-method="searchSKU"
                :loading="skuLoading"
                size="small"
                style="width: 100%"
                @change="(val) => handleSkuSelect(val, row)"
              >
                <el-option
                  v-for="item in skuOptions"
                  :key="item.id"
                  :label="`${item.sku_code} - ${item.sku_name} (${item.product?.style || '未知款式'})`"
                  :value="item.id"
                />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="数量" width="100">
            <template #default="{ row }">
              <el-input-number v-model="row.quantity" :min="1" size="small" controls-position="right" style="width: 100%" />
            </template>
          </el-table-column>
          <el-table-column label="单价" width="120">
            <template #default="{ row }">
              <el-input-number v-model="row.purchase_price" :min="0" :precision="2" size="small" controls-position="right" style="width: 100%" />
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
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getPurchaseList,
  getPurchaseDetail,
  createPurchase,
  updatePurchase,
  approvePurchase,
  cancelPurchase,
} from '@/api/purchase'
import { getSupplierList } from '@/api/supplier'
import { getAllSkuList } from '@/api/product'

const router = useRouter()
const route = useRoute()

// ==================== 搜索与列表 ====================
const loading = ref(false)
const purchaseList = ref([])
const supplierOptions = ref([])
const skuOptions = ref([])
const skuLoading = ref(false)

// 编辑模式
const isEditMode = ref(false)
const editingOrderId = ref(null)

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

    const res = await getPurchaseList(params)
    purchaseList.value = res.data?.list || res.data || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取采购列表失败:', error)
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

// 搜索SKU
// SKU选择后回填进货价
const handleSkuSelect = (skuId, row) => {
  const selectedSku = skuOptions.value.find(s => s.id === skuId)
  if (selectedSku) {
    row.purchase_price = Number(selectedSku.product?.cost_price) || Number(selectedSku.product?.reference_cost) || 0
    row.product_name = selectedSku.product?.product_name || ''
    row.sku_name = selectedSku.sku_name || ''
    row.sku_code = selectedSku.sku_code || ''
    
    // 组合品牌和款式
    const brand = selectedSku.product?.brand || ''
    const style = selectedSku.product?.style || ''
    row.brand = brand
    row.brand_style = (brand && style) ? `${brand}${style}` : (brand || style) || ''
    
    // 如果获取不到品牌或款式，给出提示
    if (!brand && !style) {
      ElMessage.warning('该商品缺少品牌和款式信息，请补充')
    }
  }
}

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

// ==================== 新建采购单 ====================
const formDialogVisible = ref(false)
const submitLoading = ref(false)
const formRef = ref(null)

const formData = reactive({
  supplier_id: '',
  remark: '',
  items: [],
})

const formRules = {
  supplier_id: [{ required: true, message: '请选择供应商', trigger: 'change' }],
}

const handleAdd = () => {
  isEditMode.value = false
  editingOrderId.value = null
  formData.supplier_id = ''
  formData.remark = ''
  formData.items = []
  // 确保供应商列表已加载
  if (supplierOptions.value.length === 0) {
    fetchSupplierOptions()
  }
  formDialogVisible.value = true
}

const handleEdit = async (row) => {
  isEditMode.value = true
  editingOrderId.value = row.id

  try {
    const res = await getPurchaseDetail(row.id)
    const order = res.data || {}
    formData.supplier_id = order.supplier_id || ''
    formData.remark = order.remark || ''
    formData.items = (order.items || []).map(item => ({
        id: item.id,
        sku_id: item.sku_id,
        product_name: item.product_name || '',
        sku_name: item.sku_name || '',
        sku_code: item.sku?.sku_code || item.sku_code || '',
        style: item.style || item.sku?.product?.style || '',
        quantity: item.quantity,
        purchase_price: Number(item.purchase_price) || 0,
      }))

    if (supplierOptions.value.length === 0) {
      await fetchSupplierOptions()
    }

    formDialogVisible.value = true
  } catch (error) {
    console.error('获取采购单详情失败:', error)
    ElMessage.error('获取采购单详情失败')
  }
}

const handleAddItem = () => {
  formData.items.push({
    sku_id: '',
    quantity: 1,
    purchase_price: 0,
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

  // 验证所有明细都已选择SKU
  const invalidItems = formData.items.filter(item => !item.sku_id)
  if (invalidItems.length > 0) {
    ElMessage.warning('请选择所有商品明细的SKU')
    return
  }

  submitLoading.value = true
  try {
    const selectedSupplier = supplierOptions.value.find(s => s.id === formData.supplier_id)
    const payload = {
      supplier_id: formData.supplier_id,
      supplier_name: selectedSupplier?.supplier_name || '',
      remark: formData.remark,
      items: formData.items.map(item => ({
        id: item.id || undefined,
        sku_id: item.sku_id,
        product_name: item.product_name || '',
        sku_name: item.sku_name || '',
        sku_code: item.sku_code || '',
        brand_style: item.brand_style || '',
        quantity: item.quantity,
        purchase_price: item.purchase_price,
      })),
    }

    if (isEditMode.value) {
      await updatePurchase(editingOrderId.value, payload)
      ElMessage.success('采购单修改成功')
    } else {
      await createPurchase(payload)
      ElMessage.success('采购单创建成功')
    }
    formDialogVisible.value = false
    fetchList()
  } catch (error) {
    console.error(isEditMode.value ? '修改采购单失败:' : '创建采购单失败:', error)
  } finally {
    submitLoading.value = false
  }
}

// ==================== 操作 ====================
const handleDetail = (row) => {
  router.push(`/inventory/purchase/detail/${row.id}`)
}

const handleApprove = (row) => {
  ElMessageBox.confirm(
    `确定要审核通过采购单 "${row.purchase_no}" 吗？`,
    '审核确认',
    {
      confirmButtonText: '确定审核',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await approvePurchase(row.id)
      ElMessage.success('审核成功')
      fetchList()
    } catch (error) {
      console.error('审核失败:', error)
    }
  }).catch(() => {})
}

const handleCreateReceipt = (row) => {
  router.push({ 
    path: '/inventory/receipt',
    query: { 
      purchase_id: row.id,
      purchase_no: row.purchase_no,
      supplier_id: row.supplier_id 
    }
  })
}

const handleCancel = (row) => {
  ElMessageBox.confirm(
    `确定要取消采购单 "${row.purchase_no}" 吗？`,
    '取消确认',
    {
      confirmButtonText: '确定取消',
      cancelButtonText: '返回',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await cancelPurchase(row.id)
      ElMessage.success('取消成功')
      fetchList()
    } catch (error) {
      console.error('取消失败:', error)
    }
  }).catch(() => {})
}

// ==================== 初始化 ====================
onMounted(async () => {
  fetchList()
  fetchSupplierOptions()

  // 编辑模式：从详情页跳转过来
  if (route.query.mode === 'edit' && route.query.id) {
    const row = { id: Number(route.query.id) }
    await handleEdit(row)
    router.replace({ query: {} })
    return
  }

  // 检查是否有预选商品数据（从订单详情跳转）
  if (route.query.mode === 'create' && history.state?.prefillItems) {
    const prefillItems = history.state.prefillItems
    const orderNo = route.query.order_no || ''
    
    // 自动打开新建采购单弹窗
    formData.supplier_id = ''
    formData.remark = orderNo ? `由订单[${orderNo}]生成` : ''
    formData.items = prefillItems.map(item => ({
        sku_id: item.sku_id,
        product_name: item.product_name || '',
        sku_name: item.sku_name || '',
        sku_code: item.sku_code || '',
        style: item.style || '',
        quantity: item.quantity || 1,
        purchase_price: item.purchase_price || 0,
      }))
    
    formDialogVisible.value = true
    
    // 清除 URL 中的 mode 参数，避免刷新页面时重复触发
    router.replace({ query: {} })
  }
})
</script>

<style lang="scss" scoped>
.purchase-manage {
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
