<template>
  
<el-dialog v-dialog-drag
    :model-value="visible"
    title="订单退货"
    width="500px"
    destroy-on-close
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      label-width="120px"
    >
      <el-form-item label="订单号">
        <span>{{ order?.order_no }}</span>
      </el-form-item>
      <el-form-item label="最终成交价">
        <span class="price">{{ formatCurrency(order?.final_price) }}</span>
      </el-form-item>
      <el-form-item label="退货金额" prop="return_amount">
        <el-input-number
          v-model="formData.return_amount"
          :min="0"
          :max="order?.final_price || 0"
          :precision="2"
          controls-position="right"
          style="width: 100%"
          placeholder="请输入退货金额"
        />
      </el-form-item>
      <el-form-item label="利润冲减" prop="profit_deduction">
        <el-input-number
          v-model="formData.profit_deduction"
          :precision="2"
          controls-position="right"
          style="width: 100%"
          placeholder="请输入利润冲减金额"
        />
      </el-form-item>
      <el-form-item label="退货原因" prop="reason">
        <el-input
          v-model="formData.reason"
          type="textarea"
          :rows="3"
          placeholder="请输入退货原因"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="warning" :loading="submitLoading" @click="handleSubmit">
        确认退货
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { returnOrder } from '@/api/order'
import { formatCurrency } from '@/utils/format'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false,
  },
  order: {
    type: Object,
    default: null,
  },
})

const emit = defineEmits(['update:visible', 'success'])

const formRef = ref(null)
const submitLoading = ref(false)

const formData = reactive({
  return_amount: 0,
  profit_deduction: 0,
  reason: '',
})

const formRules = {
  return_amount: [{ required: true, message: '请输入退货金额', trigger: 'blur' }],
  reason: [{ required: true, message: '请输入退货原因', trigger: 'blur' }],
}

watch(
  () => props.visible,
  (val) => {
    if (val) {
      formData.return_amount = 0
      formData.profit_deduction = 0
      formData.reason = ''
    }
  }
)

const handleClose = () => {
  emit('update:visible', false)
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    await returnOrder(props.order.id, {
      return_amount: formData.return_amount,
      profit_deduction: formData.profit_deduction,
      reason: formData.reason,
    })
    ElMessage.success('退货操作成功')
    handleClose()
    emit('success')
  } catch (error) {
    console.error('退货失败:', error)
  } finally {
    submitLoading.value = false
  }
}
</script>

<style scoped>
.price {
  color: #f56c6c;
  font-weight: 500;
}
</style>
