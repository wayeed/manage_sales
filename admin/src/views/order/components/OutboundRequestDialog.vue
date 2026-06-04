<template>
  <el-dialog
    v-model="visible"
    title="申请出库"
    width="520px"
    :close-on-click-modal="false"
  >
    <div v-if="order">
      <!-- 订单基本信息 -->
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="订单号">{{ order.order_no }}</el-descriptions-item>
        <el-descriptions-item label="客户名称">{{ order.customer_name }}</el-descriptions-item>
        <el-descriptions-item label="订单金额">
          <span class="price">{{ formatCurrency(order.final_amount) }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="库存状态">
          <el-tag :type="stockStatusTag" size="small">{{ stockStatusLabel }}</el-tag>
        </el-descriptions-item>
      </el-descriptions>

      <!-- 尾款信息 -->
      <div style="margin-top: 16px">
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="尾款金额">
            <span :class="remainingRate > 20 ? 'loss' : 'price'">
              {{ formatCurrency(order.remaining_amount) }}
            </span>
          </el-descriptions-item>
          <el-descriptions-item label="尾款比例">
            <span :class="remainingRate > 20 ? 'loss' : 'price'">
              {{ remainingRate.toFixed(1) }}%
            </span>
          </el-descriptions-item>
        </el-descriptions>
      </div>

      <!-- 尾款比例过高警告 -->
      <el-alert
        v-if="remainingRate > 20"
        type="error"
        :closable="false"
        style="margin-top: 16px"
      >
        <template #title>
          尾款比例超过 20%，提交后将进入审批流程
        </template>
      </el-alert>

      <!-- 审批备注预览 -->
      <div v-if="remainingRate > 20" style="margin-top: 12px; padding: 12px; background: #fef0f0; border-radius: 4px; font-size: 13px; color: #f56c6c">
        审批备注：此订单尾款还有{{ formatCurrency(order.remaining_amount) }}元，送货后由业务员{{ order.salesman?.real_name || '未知' }}负责收回尾款
      </div>
    </div>

    <template #footer>
      <div style="display: flex; justify-content: flex-end; gap: 12px">
        <el-button @click="visible = false">取消</el-button>
        <el-button
          type="primary"
          :disabled="!canSubmit"
          :loading="loading"
          @click="handleSubmit"
        >
          {{ canSubmit ? '提交申请' : '库存不足，无法申请' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { createOutboundRequest } from '@/api/outbound-request'
import { formatCurrency } from '@/utils/format'

const props = defineProps({
  modelValue: Boolean,
  order: Object,
})

const emit = defineEmits(['update:modelValue', 'success'])

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})

const loading = ref(false)

// 尾款比例
const remainingRate = computed(() => {
  if (!props.order) return 0
  const finalAmount = parseFloat(props.order.final_amount) || 0
  const remainingAmount = parseFloat(props.order.remaining_amount) || 0
  if (finalAmount <= 0) return 0
  return (remainingAmount / finalAmount) * 100
})

// 库存状态
const stockStatusMap = {
  0: { label: '全部有库存', tag: 'success' },
  1: { label: '部分缺货', tag: 'warning' },
  2: { label: '全部缺货', tag: 'danger' },
}

const stockStatusLabel = computed(() => {
  const status = props.order?.stock_status
  return stockStatusMap[status]?.label || '未知'
})

const stockStatusTag = computed(() => {
  const status = props.order?.stock_status
  return stockStatusMap[status]?.tag || 'info'
})

// 是否可以提交申请（库存充足）
const canSubmit = computed(() => {
  return props.order?.stock_status === 0
})

// 提交申请
const handleSubmit = async () => {
  if (!props.order?.id) return

  loading.value = true
  try {
    await createOutboundRequest({ order_id: props.order.id })
    ElMessage.success('出库申请已提交')
    emit('success')
    visible.value = false
  } catch (error) {
    console.error('提交出库申请失败:', error)
    ElMessage.error(error.message || '提交出库申请失败')
  } finally {
    loading.value = false
  }
}
</script>

<style lang="scss" scoped>
.price { color: #f56c6c; font-weight: 500; }
.loss { color: #f56c6c; font-weight: 600; }
</style>
