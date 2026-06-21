<template>
  <div class="purchase-detail">
    <el-card v-loading="loading" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">采购单详情</span>
          <el-button icon="ArrowLeft" @click="handleBack">返回列表</el-button>
        </div>
      </template>

      <template v-if="detail">
        <!-- 基本信息 -->
        <el-descriptions :column="3" border class="detail-section">
          <el-descriptions-item label="采购单号">{{ detail.purchase_no }}</el-descriptions-item>
          <el-descriptions-item label="供应商">{{ detail.supplier_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusTag(detail.status)" size="small">
              {{ getStatusLabel(detail.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="采购金额">
            <span class="price">¥{{ detail.total_amount || 0 }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="采购数量">{{ detail.total_quantity || 0 }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="备注" :span="3">{{ detail.remark || '-' }}</el-descriptions-item>
        </el-descriptions>

        <!-- 商品明细 -->
        <el-divider content-position="left">商品明细</el-divider>
        <el-table
          :data="detail.items || []"
          border
          stripe
          style="width: 100%"
        >
          <el-table-column label="SKU编码" width="120">
            <template #default="{ row }">
              {{ row.sku?.sku_code || '-' }}
            </template>
          </el-table-column>
          <el-table-column label="商品名称" min-width="180">
            <template #default="{ row }">
              {{ row.sku_name || row.sku?.product?.product_name || '-' }} 
            </template>
          </el-table-column>
          <el-table-column label="品牌款式" width="160">
            <template #default="{ row }">
              {{ row.brand_style || row.sku?.product?.style || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="quantity" label="采购数量" width="100" align="center" />
          <el-table-column label="单价" width="120" align="right">
            <template #default="{ row }">
              <span class="price">¥{{ row.purchase_price || 0 }}</span>
            </template>
          </el-table-column>
          <el-table-column label="小计" width="120" align="right">
            <template #default="{ row }">
              <span class="price">¥{{ row.subtotal || ((row.quantity || 0) * (row.purchase_price || 0)).toFixed(2) }}</span>
            </template>
          </el-table-column>
        </el-table>

        <!-- 操作按钮 -->
        <div class="action-bar">
          <el-button
            v-if="Number(detail.status) === 0 || Number(detail.status) === 1"
            type="primary"
            @click="handleEdit"
          >
            编辑采购单
          </el-button>
          <el-button
            v-if="Number(detail.status) === 0"
            type="success"
            @click="handleApprove"
          >
            审核通过
          </el-button>
          <el-button
            v-if="Number(detail.status) === 1"
            type="warning"
            @click="handleCreateReceipt"
          >
            创建回货单
          </el-button>
          <el-button
            v-if="Number(detail.status) === 0"
            type="danger"
            @click="handleCancel"
          >
            取消采购单
          </el-button>
        </div>
      </template>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getPurchaseDetail,
  approvePurchase,
  cancelPurchase,
} from '@/api/purchase'

const router = useRouter()
const route = useRoute()

const loading = ref(false)
const detail = ref(null)

const fetchDetail = async () => {
  loading.value = true
  try {
    const res = await getPurchaseDetail(route.params.id)
    detail.value = res.data || null
  } catch (error) {
    console.error('获取采购单详情失败:', error)
    ElMessage.error('获取采购单详情失败')
  } finally {
    loading.value = false
  }
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

const handleBack = () => {
  router.push('/inventory/purchase')
}

const handleEdit = () => {
  router.push(`/inventory/purchase?mode=edit&id=${detail.value.id}`)
}

const handleApprove = async () => {
  try {
    await ElMessageBox.confirm('确定要审核通过该采购单吗？', '审核确认', {
      confirmButtonText: '确定审核',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await approvePurchase(detail.value.id)
    ElMessage.success('审核成功')
    fetchDetail()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('审核失败:', error)
    }
  }
}

const handleCreateReceipt = () => {
  router.push({ 
    path: '/inventory/receipt',
    query: { 
      purchase_id: detail.value.id,
      purchase_no: detail.value.purchase_no,
      supplier_id: detail.value.supplier_id 
    }
  })
}

const handleCancel = async () => {
  try {
    await ElMessageBox.confirm('确定要取消该采购单吗？', '取消确认', {
      confirmButtonText: '确定取消',
      cancelButtonText: '返回',
      type: 'warning',
    })
    await cancelPurchase(detail.value.id)
    ElMessage.success('取消成功')
    fetchDetail()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('取消失败:', error)
    }
  }
}

onMounted(() => {
  fetchDetail()
})
</script>

<style lang="scss" scoped>
.purchase-detail {
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

  .detail-section {
    margin-bottom: 16px;
  }

  .price {
    color: #f56c6c;
    font-weight: 500;
  }

  .action-bar {
    margin-top: 24px;
    text-align: center;
  }
}
</style>
