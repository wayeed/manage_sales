<template>
  
<el-dialog v-dialog-drag
    :model-value="visible"
    title="订单审核"
    width="500px"
    destroy-on-close
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      label-width="100px"
    >
      <el-form-item label="订单号">
        <span>{{ order?.order_no }}</span>
      </el-form-item>
      <el-form-item label="审核结果" prop="result">
        <el-radio-group v-model="formData.result">
          <el-radio value="approved">通过</el-radio>
          <el-radio value="rejected">驳回</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="审核备注" prop="remark">
        <el-input
          v-model="formData.remark"
          type="textarea"
          :rows="3"
          placeholder="请输入审核备注"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" :loading="submitLoading" @click="handleSubmit">
        确认
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { approveOrder } from '@/api/order'

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
  result: 'approved',
  remark: '',
})

const formRules = {
  result: [{ required: true, message: '请选择审核结果', trigger: 'change' }],
}

watch(
  () => props.visible,
  (val) => {
    if (val) {
      formData.result = 'approved'
      formData.remark = ''
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
    await approveOrder(props.order.id, {
      result: formData.result,
      remark: formData.remark,
    })
    ElMessage.success(formData.result === 'approved' ? '审核通过' : '已驳回')
    handleClose()
    emit('success')
  } catch (error) {
    console.error('审核失败:', error)
  } finally {
    submitLoading.value = false
  }
}
</script>
