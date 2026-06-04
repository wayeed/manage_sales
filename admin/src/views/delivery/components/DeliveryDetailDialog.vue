<template>
  <el-dialog v-model="dialogVisible" title="送货出库详情" width="700px" destroy-on-close v-dialog-drag>
    <div v-loading="loading">
      <template v-if="detail">
        <!-- 基本信息 -->
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="送货单号">{{ detail.delivery_no }}</el-descriptions-item>
          <el-descriptions-item label="订单编号">
            <el-link type="primary" @click="goOrderDetail">{{ detail.order_no }}</el-link>
          </el-descriptions-item>
          <el-descriptions-item label="出库仓库">{{ detail.warehouse_name }}</el-descriptions-item>
          <el-descriptions-item label="操作人">{{ detail.operator_name }}</el-descriptions-item>
          <el-descriptions-item label="送货类型">{{ detail.delivery_type_name }}</el-descriptions-item>
          <el-descriptions-item label="送货时间">{{ formatDateTime(detail.delivery_time) }}</el-descriptions-item>
          <el-descriptions-item label="物流单号">{{ detail.logistics_no || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="detail.status === 1 ? 'success' : 'danger'" size="small">{{ detail.status_name }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="收货人">{{ detail.receiver_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="联系电话">{{ detail.receiver_phone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="收货地址" :span="2">{{ detail.receiver_address || '-' }}</el-descriptions-item>
          <el-descriptions-item label="备注" :span="2">{{ detail.remark || '-' }}</el-descriptions-item>
        </el-descriptions>

        <!-- 商品明细 -->
        <h4 style="margin: 16px 0 8px">商品明细</h4>
        <el-table :data="detail.items" border size="small" show-summary :summary-method="getSummary">
          <el-table-column prop="product_name" label="商品名称" min-width="140" />
          <el-table-column prop="sku_name" label="规格" min-width="120" />
          <el-table-column prop="quantity" label="数量" width="70" align="center" />
          <el-table-column prop="unit_cost" label="单位成本" width="100" align="right">
            <template #default="{ row }">{{ formatMoney(row.unit_cost) }}</template>
          </el-table-column>
          <el-table-column prop="total_cost" label="总成本" width="110" align="right">
            <template #default="{ row }">
              <span class="cost">{{ formatMoney(row.total_cost) }}</span>
            </template>
          </el-table-column>
        </el-table>

        <!-- 合计信息 -->
        <div class="summary-row">
          <span>总数量: <b>{{ detail.total_quantity }}</b></span>
          <span>总金额: <b class="price">{{ formatMoney(detail.total_amount) }}</b></span>
        </div>
      </template>
    </div>

    <template #footer>
      <el-button @click="handleClose">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { getDeliveryDetail } from '@/api/delivery'
import { formatMoney, formatDateTime } from '@/utils/format'

const props = defineProps({
  visible: { type: Boolean, default: false },
  deliveryId: { type: [Number, null], default: null },
})
const emit = defineEmits(['update:visible'])

const router = useRouter()

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
})

const loading = ref(false)
const detail = ref(null)

watch(() => props.visible, async (val) => {
  if (val && props.deliveryId) {
    await fetchDetail()
  }
})

const fetchDetail = async () => {
  loading.value = true
  try {
    const res = await getDeliveryDetail(props.deliveryId)
    detail.value = res.data || null
  } catch (error) {
    console.error('获取送货详情失败:', error)
  } finally {
    loading.value = false
  }
}

const goOrderDetail = () => {
  if (detail.value?.order_id) {
    handleClose()
    router.push(`/order/detail/${detail.value.order_id}`)
  }
}

const getSummary = (param) => {
  const { columns, data } = param
  const sums = []
  columns.forEach((col, index) => {
    if (index === 0) { sums[index] = '合计'; return }
    if (col.property === 'quantity') {
      sums[index] = data.reduce((sum, row) => sum + (row.quantity || 0), 0)
    } else if (col.property === 'total_cost') {
      const total = data.reduce((sum, row) => sum + (parseFloat(row.total_cost) || 0), 0)
      sums[index] = formatMoney(total)
    } else {
      sums[index] = ''
    }
  })
  return sums
}

const handleClose = () => { emit('update:visible', false) }
</script>

<script>
export default { name: 'DeliveryDetailDialog' }
</script>

<style lang="scss" scoped>
.summary-row {
  display: flex;
  justify-content: flex-end;
  gap: 24px;
  margin-top: 12px;
  font-size: 14px;
  .price {
    color: #f56c6c;
  }
  .cost {
    color: #909399;
  }
}
</style>
