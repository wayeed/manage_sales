<template>
  <div class="order-detail">
    <el-card v-loading="loading" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">订单详情</span>
          <el-button icon="ArrowLeft" @click="handleBack">返回列表</el-button>
        </div>
      </template>

      <template v-if="detail">
        <!-- 基本信息 -->
        <el-card class="info-card" shadow="never">
          <template #header>
            <span class="section-title">基本信息</span>
          </template>
          <el-descriptions :column="3" border>
            <el-descriptions-item label="订单号">{{ detail.order_no }}</el-descriptions-item>
            <el-descriptions-item label="订单状态">
              <el-tag :type="getStatusTag(detail.order_status)" size="small">
                {{ getStatusLabel(detail.order_status) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="订单类型">
              <el-tag :type="getOrderTypeTag(detail.order_type)" size="small">
                {{ getOrderTypeLabel(detail.order_type) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="创建时间">{{ detail.created_at }}</el-descriptions-item>
            <el-descriptions-item label="业务员">{{ detail.salesman?.real_name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="客户名称">{{ detail.customer_name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="客户手机">{{ detail.customer_phone || '-' }}</el-descriptions-item>
            <el-descriptions-item label="客户等级">{{ detail.customer?.level || '-' }}</el-descriptions-item>
            <el-descriptions-item label="备注" :span="3">{{ detail.remark || '-' }}</el-descriptions-item>
          </el-descriptions>
        </el-card>

        <!-- 金额信息 -->
        <el-card class="info-card" shadow="never">
          <template #header>
            <span class="section-title">金额信息</span>
          </template>
          <el-descriptions :column="4" border>
            <el-descriptions-item label="挂牌总价">
              <span class="price">{{ formatCurrency(detail.total_list_price) }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="实际销售价">
              <span class="price">{{ formatCurrency(detail.total_sale_price) }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="折扣金额">
              <span class="discount">-{{ formatCurrency(detail.discount_amount) }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="最终成交价">
              <span class="price">{{ formatCurrency(detail.final_amount) }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="商品成本">
              <span class="cost">{{ formatCurrency(detail.total_cost) }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="礼品成本">
              <span class="cost">{{ formatCurrency(detail.gift_cost) }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="实际利润">
              <span :class="detail.actual_profit >= 0 ? 'profit' : 'loss'">
                {{ formatCurrency(detail.actual_profit) }}
              </span>
            </el-descriptions-item>
            <el-descriptions-item label="利润率">
              <span :class="detail.profit_rate >= 0 ? 'profit' : 'loss'">
                {{ formatPercent(detail.profit_rate) }}
              </span>
            </el-descriptions-item>
          </el-descriptions>
        </el-card>

        <!-- 商品明细 -->
        <el-card class="info-card" shadow="never">
          <template #header>
            <span class="section-title">商品明细</span>
          </template>
          <el-table
            :data="detail.items || []"
            border
            stripe
            style="width: 100%"
            show-summary
            :summary-method="getItemSummary"
          >
            <el-table-column prop="product_name" label="商品名称" min-width="160" />
            <el-table-column prop="sku_name" label="SKU" width="140" />
            <el-table-column prop="quantity" label="数量" width="80" align="center" />
            <el-table-column label="挂牌价" width="110" align="right">
              <template #default="{ row }">
                <span>{{ formatCurrency(row.list_price) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="销售价" width="110" align="right">
              <template #default="{ row }">
                <span class="price">{{ formatCurrency(row.sale_price) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="折扣率" width="90" align="center">
              <template #default="{ row }">
                <span>{{ formatPercent(row.discount_rate) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="单位成本" width="110" align="right">
              <template #default="{ row }">
                <span class="cost">{{ formatCurrency(row.unit_cost) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="成本合计" width="120" align="right">
              <template #default="{ row }">
                <span class="cost">{{ formatCurrency((row.quantity || 0) * (row.unit_cost || 0)) }}</span>
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <!-- 礼品信息 -->
        <el-card v-if="detail.gifts && detail.gifts.length > 0" class="info-card" shadow="never">
          <template #header>
            <span class="section-title">礼品信息</span>
          </template>
          <el-table
            :data="detail.gifts"
            border
            stripe
            style="width: 100%"
            show-summary
            :summary-method="getGiftSummary"
          >
            <el-table-column prop="gift_name" label="礼品名称" min-width="160" />
            <el-table-column prop="quantity" label="数量" width="100" align="center" />
            <el-table-column label="成本价" width="120" align="right">
              <template #default="{ row }">
                <span class="cost">{{ formatCurrency(row.cost_price) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="成本合计" width="120" align="right">
              <template #default="{ row }">
                <span class="cost">{{ formatCurrency((row.quantity || 0) * (row.cost_price || 0)) }}</span>
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <!-- 回款记录 -->
        <el-card class="info-card" shadow="never">
          <template #header>
            <div class="section-header">
              <span class="section-title">回款记录</span>
              <el-button
                v-if="detail.order_status === 1"
                type="primary"
                size="small"
                icon="Plus"
                @click="handleAddPayment"
              >
                回款录入
              </el-button>
            </div>
          </template>
          <el-table
            :data="detail.payments || []"
            border
            stripe
            style="width: 100%"
          >
            <el-table-column prop="payment_no" label="回款流水号" width="200" />
            <el-table-column label="金额" width="120" align="right">
              <template #default="{ row }">
                <span class="price">{{ formatCurrency(row.amount) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="payment_date" label="日期" width="120" />
            <el-table-column label="方式" width="100" align="center">
              <template #default="{ row }">
                {{ getPaymentMethodLabel(row.payment_method) }}
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="getPaymentStatusTag(row.status)" size="small">
                  {{ getPaymentStatusLabel(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="remark" label="备注" min-width="140" />
            <el-table-column label="操作" width="100" align="center">
              <template #default="{ row }">
                <el-button
                  v-if="row.status === 0"
                  type="success" link size="small" @click="handleApprovePayment(row)"
                >
                  确认
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <!-- 操作按钮 -->
        <div class="action-bar">
          <el-button
            v-if="detail.order_status === 0"
            type="success"
            @click="handleApprove"
          >
            审核订单
          </el-button>
          <el-button
            v-if="detail.order_status === 1"
            type="warning"
            @click="handleReturn"
          >
            退货
          </el-button>
        </div>
      </template>
    </el-card>

    <!-- 审核弹窗 -->
    <ApproveDialog
      v-model:visible="approveDialogVisible"
      :order="detail"
      @success="fetchDetail"
    />

    <!-- 退货弹窗 -->
    <ReturnDialog
      v-model:visible="returnDialogVisible"
      :order="detail"
      @success="fetchDetail"
    />

    <!-- 回款录入弹窗 -->
    <PaymentDialog
      v-model:visible="paymentDialogVisible"
      :order-id="detail?.id"
      @success="fetchDetail"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getOrderDetail } from '@/api/order'
import { approvePayment } from '@/api/payment'
import { formatCurrency, formatPercent } from '@/utils/format'
import ApproveDialog from './components/ApproveDialog.vue'
import ReturnDialog from './components/ReturnDialog.vue'
import PaymentDialog from './components/PaymentDialog.vue'

const router = useRouter()
const route = useRoute()

const loading = ref(false)
const detail = ref(null)

const fetchDetail = async () => {
  loading.value = true
  try {
    const res = await getOrderDetail(route.params.id)
    // 后端返回结构: { order, items, gifts, payments }
    const data = res.data || {}
    detail.value = {
      ...data.order,
      items: data.items || [],
      gifts: data.gifts || [],
      payments: data.payments || [],
    }
  } catch (error) {
    console.error('获取订单详情失败:', error)
    ElMessage.error('获取订单详情失败')
  } finally {
    loading.value = false
  }
}

// ==================== 状态映射 ====================
const getStatusLabel = (status) => {
  const map = {
    0: '待审批',
    1: '已生效',
    2: '已驳回',
    3: '已取消',
    4: '已退货',
  }
  return map[status] ?? status ?? '未知'
}

const getStatusTag = (status) => {
  const map = {
    0: 'warning',
    1: 'success',
    2: 'danger',
    3: 'info',
    4: '',
  }
  return map[status] ?? 'info'
}

const getOrderTypeLabel = (type) => {
  const map = { 1: '单品', 2: '多品', 3: '特殊审批', 4: '同行单品', 5: '同行多品', 6: '同行特批' }
  return map[type] ?? type ?? '-'
}

const getOrderTypeTag = (type) => {
  const map = { 1: '', 2: 'success', 3: 'warning', 4: 'info', 5: '', 6: 'danger' }
  return map[type] ?? 'info'
}

const getPaymentMethodLabel = (method) => {
  const map = { 0: '银行转账', 1: '现金', 2: '微信', 3: '支付宝' }
  return map[method] || method || '-'
}

const getPaymentStatusLabel = (status) => {
  const map = { 0: '待确认', 1: '已确认' }
  return map[status] || status || '-'
}

const getPaymentStatusTag = (status) => {
  const map = { 0: 'warning', 1: 'success' }
  return map[status] || 'info'
}

// ==================== 合计行 ====================
const getItemSummary = (param) => {
  const { columns, data } = param
  const sums = []
  columns.forEach((column, index) => {
    if (index === 0) {
      sums[index] = '合计'
      return
    }
    if (index === 2) {
      const total = data.reduce((sum, row) => sum + (row.quantity || 0), 0)
      sums[index] = total
      return
    }
    if (index === 3) {
      const total = data.reduce((sum, row) => sum + (row.list_price || 0) * (row.quantity || 0), 0)
      sums[index] = formatCurrency(total)
      return
    }
    if (index === 4) {
      const total = data.reduce((sum, row) => sum + (row.sale_price || 0) * (row.quantity || 0), 0)
      sums[index] = formatCurrency(total)
      return
    }
    if (index === 7) {
      const total = data.reduce((sum, row) => sum + (row.quantity || 0) * (row.unit_cost || 0), 0)
      sums[index] = formatCurrency(total)
      return
    }
    sums[index] = ''
  })
  return sums
}

const getGiftSummary = (param) => {
  const { columns, data } = param
  const sums = []
  columns.forEach((column, index) => {
    if (index === 0) {
      sums[index] = '合计'
      return
    }
    if (index === 1) {
      const total = data.reduce((sum, row) => sum + (row.quantity || 0), 0)
      sums[index] = total
      return
    }
    if (index === 3) {
      const total = data.reduce((sum, row) => sum + (row.quantity || 0) * (row.cost_price || 0), 0)
      sums[index] = formatCurrency(total)
      return
    }
    sums[index] = ''
  })
  return sums
}

// ==================== 操作 ====================
const handleBack = () => {
  router.push('/order/index')
}

const approveDialogVisible = ref(false)
const returnDialogVisible = ref(false)
const paymentDialogVisible = ref(false)

const handleApprove = () => {
  approveDialogVisible.value = true
}

const handleReturn = () => {
  returnDialogVisible.value = true
}

const handleAddPayment = () => {
  paymentDialogVisible.value = true
}

const handleApprovePayment = async (row) => {
  try {
    await approvePayment(row.id, { approved: true })
    ElMessage.success('回款确认成功')
    fetchDetail()
  } catch (error) {
    console.error('确认回款失败:', error)
  }
}

onMounted(() => {
  fetchDetail()
})
</script>

<style lang="scss" scoped>
.order-detail {
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

  .info-card {
    margin-bottom: 16px;

    .section-title {
      font-size: 15px;
      font-weight: 600;
      color: #303133;
    }

    .section-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
    }
  }

  .price {
    color: #f56c6c;
    font-weight: 500;
  }

  .cost {
    color: #909399;
  }

  .discount {
    color: #e6a23c;
  }

  .profit {
    color: #67c23a;
    font-weight: 500;
  }

  .loss {
    color: #f56c6c;
    font-weight: 500;
  }

  .action-bar {
    margin-top: 24px;
    text-align: center;
  }
}
</style>
