<template>
  <div class="receipt-detail">
    <!-- 返回按钮 -->
    <div class="header-bar">
      <el-button @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
    </div>

    <!-- 基本信息卡片 -->
    <el-card class="info-card" shadow="never">
      <div class="info-header">
        <div class="order-title">
          <span class="order-no">{{ receiptDetail?.receipt_no }}</span>
          <el-tag :type="getStatusTag(receiptDetail?.status)" size="small" style="margin-left: 12px">
            {{ getStatusLabel(receiptDetail?.status) }}
          </el-tag>
        </div>
        <div class="action-buttons">
          <el-button
            v-if="receiptDetail?.status === 0"
            type="success"
            icon="Check"
            @click="handleApprove"
          >
            审核通过
          </el-button>
          <el-button
            v-if="receiptDetail?.status === 1"
            type="warning"
            icon="Box"
            @click="handleReceive"
          >
            入库
          </el-button>
          <el-button
            v-if="receiptDetail?.status === 0"
            type="danger"
            icon="Close"
            @click="handleCancel"
          >
            取消
          </el-button>
        </div>
      </div>

      <el-row :gutter="20" class="info-row">
        <el-col :span="6">
          <div class="info-item">
            <span class="label">供应商：</span>
            <span class="value">{{ receiptDetail?.supplier_name || '-' }}</span>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-item">
            <span class="label">发货金额：</span>
            <span class="value price">¥{{ receiptDetail?.total_amount || 0 }}</span>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-item">
            <span class="label">发货数量：</span>
            <span class="value">{{ receiptDetail?.total_quantity || 0 }} 件</span>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-item">
            <span class="label">已入库：</span>
            <span class="value">{{ receivedQuantity }} 件</span>
          </div>
        </el-col>
      </el-row>

      <el-row :gutter="20" class="info-row">
        <el-col :span="6">
          <div class="info-item">
            <span class="label">创建人：</span>
            <span class="value">{{ receiptDetail?.created_by_name || '-' }}</span>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-item">
            <span class="label">创建时间：</span>
            <span class="value">{{ formatTime(receiptDetail?.created_at) }}</span>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-item">
            <span class="label">审核人：</span>
            <span class="value">{{ receiptDetail?.audited_by_name || '-' }}</span>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-item">
            <span class="label">审核时间：</span>
            <span class="value">{{ formatTime(receiptDetail?.audited_at) }}</span>
          </div>
        </el-col>
      </el-row>

      <div v-if="receiptDetail?.remark" class="remark-item">
        <span class="label">备注：</span>
        <span class="value">{{ receiptDetail.remark }}</span>
      </div>
    </el-card>

    <!-- 商品明细卡片 -->
    <el-card class="detail-card" shadow="never">
      <template #header>
        <span class="card-title">商品明细</span>
      </template>

      <el-table
        v-loading="loading"
        :data="receiptDetail?.items"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column prop="sku_code" label="SKU编码" width="160" />
        <el-table-column prop="product_name" label="商品名称" width="180" />
        <el-table-column prop="sku_name" label="规格" width="200" />
        <el-table-column prop="brand_style" label="品牌款式" width="150" />
        <el-table-column label="发货数量" width="120" align="center">
          <template #default="{ row }">
            {{ row.ship_quantity || 0 }}
          </template>
        </el-table-column>
        <el-table-column label="已入库数量" width="120" align="center">
          <template #default="{ row }">
            {{ row.receive_quantity || 0 }}
          </template>
        </el-table-column>
        <el-table-column label="成本价" width="120" align="right">
          <template #default="{ row }">
            ¥{{ row.cost_price || 0 }}
          </template>
        </el-table-column>
        <el-table-column label="小计" width="140" align="right">
          <template #default="{ row }">
            ¥{{ getSubtotal(row) }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>

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
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Check, Box, Close } from '@element-plus/icons-vue'
import { getReceiptDetail, approveReceipt, receiveReceipt, cancelReceipt } from '@/api/receipt'
import { getWarehouseList } from '@/api/warehouse'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const receiptDetail = ref({})
const warehouseOptions = ref([])

const receiveDialogVisible = ref(false)
const receiveLoading = ref(false)
const receiveFormRef = ref(null)
const receiveItems = ref([])

const receiveForm = reactive({
  warehouse_id: '',
  remark: '',
})

const receiveRules = {
  warehouse_id: [{ required: true, message: '请选择入库仓库', trigger: 'change' }],
}

const receivedQuantity = computed(() => {
  if (!receiptDetail.value.items || receiptDetail.value.items.length === 0) return 0
  return receiptDetail.value.items.reduce((sum, item) => sum + (item.receive_quantity || 0), 0)
})

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

const getSubtotal = (row) => {
  const price = Number(row.cost_price) || 0
  const qty = row.ship_quantity || 0
  return (price * qty).toFixed(2)
}

const goBack = () => {
  router.push('/inventory/receipt')
}

const fetchDetail = async () => {
  const id = route.params.id
  if (!id) return

  loading.value = true
  try {
    const res = await getReceiptDetail(id)
    receiptDetail.value = res.data || {}
  } catch (error) {
    console.error('获取回货单详情失败:', error)
    ElMessage.error('获取回货单详情失败')
  } finally {
    loading.value = false
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

const handleApprove = () => {
  ElMessageBox.confirm(
    `确定要审核通过回货单 "${receiptDetail.value.receipt_no}" 吗？`,
    '审核确认',
    {
      confirmButtonText: '确定审核',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await approveReceipt(route.params.id)
      ElMessage.success('审核成功')
      fetchDetail()
    } catch (error) {
      console.error('审核失败:', error)
    }
  }).catch(() => {})
}

const handleReceive = () => {
  receiveForm.warehouse_id = ''
  receiveForm.remark = ''
  
  receiveItems.value = (receiptDetail.value.items || []).map(item => ({
    receipt_item_id: item.id,
    sku_code: item.sku_code || '',
    product_name: item.product_name || '',
    remaining: item.ship_quantity - (item.receive_quantity || 0),
    receive_quantity: item.ship_quantity - (item.receive_quantity || 0),
  }))

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
    await receiveReceipt(route.params.id, {
      warehouse_id: receiveForm.warehouse_id,
      remark: receiveForm.remark,
      items: items.map(item => ({
        receipt_item_id: item.receipt_item_id,
        receive_quantity: item.receive_quantity,
      })),
    })
    ElMessage.success('入库成功')
    receiveDialogVisible.value = false
    fetchDetail()
  } catch (error) {
    console.error('入库失败:', error)
  } finally {
    receiveLoading.value = false
  }
}

const handleCancel = () => {
  ElMessageBox.confirm(
    `确定要取消回货单 "${receiptDetail.value.receipt_no}" 吗？`,
    '取消确认',
    {
      confirmButtonText: '确定取消',
      cancelButtonText: '返回',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await cancelReceipt(route.params.id)
      ElMessage.success('取消成功')
      fetchDetail()
    } catch (error) {
      console.error('取消失败:', error)
    }
  }).catch(() => {})
}

onMounted(() => {
  fetchDetail()
  fetchWarehouseOptions()
})
</script>

<style lang="scss" scoped>
.receipt-detail {
  .header-bar {
    margin-bottom: 16px;
  }

  .info-card {
    margin-bottom: 16px;

    .info-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 20px;
      padding-bottom: 16px;
      border-bottom: 1px solid #ebeef5;

      .order-title {
        display: flex;
        align-items: center;

        .order-no {
          font-size: 18px;
          font-weight: 600;
          color: #303133;
        }
      }

      .action-buttons {
        display: flex;
        gap: 8px;
      }
    }

    .info-row {
      margin-bottom: 16px;

      .info-item {
        display: flex;
        align-items: center;

        .label {
          color: #606266;
          font-size: 14px;
          margin-right: 8px;
        }

        .value {
          color: #303133;
          font-size: 14px;

          &.price {
            color: #f56c6c;
            font-weight: 500;
          }
        }
      }
    }

    .remark-item {
      display: flex;
      padding-top: 16px;
      border-top: 1px solid #ebeef5;

      .label {
        color: #606266;
        font-size: 14px;
        margin-right: 8px;
      }

      .value {
        color: #303133;
        font-size: 14px;
      }
    }
  }

  .detail-card {
    .card-title {
      font-size: 16px;
      font-weight: 600;
      color: #303133;
    }
  }
}
</style>