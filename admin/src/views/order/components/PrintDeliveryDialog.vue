<template>
  <el-dialog
    :model-value="visible"
    title="送货单"
    width="700px"
    :close-on-click-modal="false"
    @close="handleClose"
    class="print-delivery-dialog"
  >
    <!-- 打印区域 -->
    <div id="print-area" v-if="order" class="delivery-slip">
      <!-- 订单信息 -->
      <div class="slip-header">
        <h2 class="slip-title">送 货 单</h2>
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="订单号">{{ order.order_no }}</el-descriptions-item>
          <el-descriptions-item label="业务员">{{ order.salesman?.real_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="客户名称">{{ order.customer_name }}</el-descriptions-item>
          <el-descriptions-item label="客户电话">{{ order.customer_phone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="送货地址" :span="2">{{ order.customer_address || '-' }}</el-descriptions-item>
        </el-descriptions>
      </div>

      <!-- 商品明细 -->
      <el-table :data="order.items || []" border size="small" style="margin-top: 16px">
        <el-table-column type="index" label="序号" width="60" align="center" />
        <el-table-column prop="product_name" label="商品名称" min-width="150" />
        <el-table-column prop="sku_name" label="规格" width="120" />
        <el-table-column prop="quantity" label="数量" width="70" align="center" />
        <el-table-column label="单价" width="100" align="right">
          <template #default="{ row }">
            {{ formatCurrency(row.sale_price) }}
          </template>
        </el-table-column>
        <el-table-column label="金额" width="100" align="right">
          <template #default="{ row }">
            {{ formatCurrency((row.sale_price || 0) * (row.quantity || 0)) }}
          </template>
        </el-table-column>
      </el-table>

      <!-- 金额汇总 -->
      <div class="slip-summary">
        <div class="summary-row">
          <span>最终成交价：</span>
          <span class="amount">{{ formatCurrency(order.final_amount) }}</span>
        </div>
        <div class="summary-row">
          <span>已付金额：</span>
          <span class="amount paid">{{ formatCurrency(order.paid_amount) }}</span>
        </div>
        <div class="summary-row">
          <span>尾款金额：</span>
          <span class="amount remaining">{{ formatCurrency(order.remaining_amount) }}</span>
        </div>
        <div class="summary-row">
          <span>尾款比例：</span>
          <span :class="remainingRate > 20 ? 'amount remaining' : 'amount paid'">
            {{ remainingRate.toFixed(1) }}%
          </span>
        </div>
      </div>

      <!-- 送货备注 -->
      <div v-if="deliveryRemark" class="slip-remark">
        <div class="remark-label">备注：</div>
        <div class="remark-content">{{ deliveryRemark }}</div>
      </div>

      <!-- 审批提示 -->
      <div v-if="printApprovalStatus?.need_approval && !printApprovalStatus?.can_print" class="approval-tip">
        <el-alert
          :title="printApprovalStatus.message"
          type="warning"
          show-icon
          :closable="false"
        />
      </div>
    </div>

    <!-- 操作按钮区域（打印时隐藏） -->
    <template #footer>
      <div class="dialog-footer no-print">
        <el-button @click="handleClose">取 消</el-button>
        <el-button
          v-if="canDirectPrint"
          type="primary"
          :loading="printLoading"
          @click="handlePrint"
        >
          立即打印
        </el-button>
        <el-button
          v-if="needSubmitApproval"
          type="warning"
          :loading="submitLoading"
          @click="handleSubmitApproval"
        >
          提交审批
        </el-button>
        <el-button
          v-if="deliveryStatus === 1"
          type="success"
          :loading="confirmLoading"
          @click="handleConfirmReceived"
        >
          确认送达
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { createPrintApproval } from '@/api/order'
import { printDelivery, confirmDelivery as confirmDeliveryApi } from '@/api/delivery'
import { formatCurrency } from '@/utils/format'

const props = defineProps({
  visible: Boolean,
  order: Object,
  printApprovalStatus: Object,
  deliveryStatus: {
    type: Number,
    default: 0
  },
})

const emit = defineEmits(['update:visible', 'success'])

const printLoading = ref(false)
const submitLoading = ref(false)
const confirmLoading = ref(false)

const remainingRate = computed(() => {
  if (!props.order) return 0
  const finalAmount = parseFloat(props.order.final_amount) || 0
  const remainingAmount = parseFloat(props.order.remaining_amount) || 0
  if (finalAmount <= 0) return 0
  return (remainingAmount / finalAmount) * 100
})

const deliveryRemark = computed(() => {
  // 从送货记录中获取备注
  if (!props.order || !props.order.delivery_records || props.order.delivery_records.length === 0) {
    return ''
  }
  // 取第一条送货记录的备注
  return props.order.delivery_records[0].remark || ''
})

const canDirectPrint = computed(() => {
  // 新流程：出库确认后打开打印弹窗，无需审批，直接可打印
  if (!props.printApprovalStatus) return true
  return props.printApprovalStatus.can_print || false
})

const needSubmitApproval = computed(() => {
  // 新流程：不再需要提交审批
  if (!props.printApprovalStatus) return false
  return props.printApprovalStatus.need_approval && !props.printApprovalStatus.can_print
})

const handleClose = () => {
  emit('update:visible', false)
}

const handlePrint = async () => {
  printLoading.value = true
  try {
    // 调用打印API，更新delivery_status为1（配送中）
    await printDelivery(props.order.id)
    // 执行浏览器打印
    window.print()
    ElMessage.success('送货单已打印')
    emit('success')
  } catch (error) {
    console.error('打印失败:', error)
    ElMessage.error(error.message || '打印失败')
  } finally {
    printLoading.value = false
  }
}

const handleSubmitApproval = async () => {
  submitLoading.value = true
  try {
    const res = await createPrintApproval(props.order.id)
    const data = res.data || {}
    ElMessage.success(data.message || '审批申请已提交')
    emit('success')
    handleClose()
  } catch (error) {
    console.error('提交审批失败:', error)
  } finally {
    submitLoading.value = false
  }
}

const handleConfirmReceived = async () => {
  confirmLoading.value = true
  try {
    await confirmDeliveryApi(props.order.id)
    ElMessage.success('已确认送达')
    emit('success')
    handleClose()
  } catch (error) {
    console.error('确认送达失败:', error)
    ElMessage.error(error.message || '确认送达失败')
  } finally {
    confirmLoading.value = false
  }
}
</script>

<style lang="scss" scoped>
.delivery-slip {
  .slip-header {
    margin-bottom: 8px;

    .slip-title {
      text-align: center;
      margin-bottom: 16px;
      font-size: 20px;
      font-weight: 600;
    }
  }

  .slip-summary {
    margin-top: 16px;
    padding: 12px 16px;
    background: #f5f7fa;
    border-radius: 4px;

    .summary-row {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 6px 0;
      font-size: 14px;

      .amount {
        font-weight: 600;
        color: #f56c6c;
      }

      .paid {
        color: #67c23a;
      }

      .remaining {
        color: #f56c6c;
      }
    }
  }

  .approval-tip {
    margin-top: 16px;
  }

  .slip-remark {
    margin-top: 16px;
    padding: 12px 16px;
    background: #fffbe6;
    border: 1px solid #ffe58f;
    border-radius: 4px;

    .remark-label {
      font-weight: 600;
      color: #d48806;
      margin-bottom: 4px;
    }

    .remark-content {
      color: #595959;
      font-size: 14px;
      line-height: 1.5;
    }
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* 打印样式：隐藏按钮区域 */
@media print {
  .no-print {
    display: none !important;
  }

  /* 隐藏弹窗的遮罩层和关闭按钮 */
  :deep(.el-overlay) {
    position: static !important;
    background: none !important;
  }

  :deep(.el-dialog) {
    box-shadow: none !important;
    margin: 0 !important;
    width: 100% !important;
  }

  :deep(.el-dialog__header) {
    display: none !important;
  }

  :deep(.el-dialog__footer) {
    display: none !important;
  }

  :deep(.el-dialog__body) {
    padding: 0 !important;
  }

  /* 打印时表格边框 */
  :deep(.el-table) {
    border: 1px solid #000 !important;
  }

  :deep(.el-table th),
  :deep(.el-table td) {
    border: 1px solid #000 !important;
  }

  :deep(.el-table::before) {
    display: none;
  }
}
</style>
