<template>

<el-dialog v-dialog-drag
    :model-value="visible"
    title="回款录入"
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
      <el-form-item label="回款金额" prop="amount">
        <el-input-number
          v-model="formData.amount"
          :min="0.01"
          :precision="2"
          controls-position="right"
          style="width: 100%"
          placeholder="请输入回款金额"
        />
      </el-form-item>
      <el-form-item label="回款日期" prop="payment_date">
        <el-date-picker
          v-model="formData.payment_date"
          type="date"
          placeholder="请选择回款日期"
          value-format="YYYY-MM-DD"
          style="width: 100%"
        />
      </el-form-item>
      <el-form-item label="回款方式" prop="payment_method">
        <el-select v-model="formData.payment_method" placeholder="请选择回款方式" style="width: 100%">
          <el-option label="银行转账" :value="0" />
          <el-option label="现金" :value="1" />
          <el-option label="微信" :value="2" />
          <el-option label="支付宝" :value="3" />
        </el-select>
      </el-form-item>
      <el-form-item label="备注">
        <el-input
          v-model="formData.remark"
          type="textarea"
          :rows="2"
          placeholder="请输入备注（选填）"
        />
      </el-form-item>
      <el-form-item label="付款凭证">
        <el-upload
          class="voucher-uploader"
          :action="uploadUrl"
          :headers="uploadHeaders"
          :show-file-list="false"
          :on-success="handleUploadSuccess"
          :before-upload="beforeUpload"
          accept="image/*"
        >
          <el-image v-if="formData.voucher_url" :src="formData.voucher_url" fit="cover" class="voucher-preview" />
          <el-icon v-else class="voucher-uploader-icon"><Plus /></el-icon>
        </el-upload>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" :loading="submitLoading" @click="handleSubmit">
        确认录入
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getToken } from '@/utils/auth'
import { createPayment } from '@/api/payment'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false,
  },
  orderId: {
    type: [Number, String],
    default: null,
  },
})

const emit = defineEmits(['update:visible', 'success'])

const formRef = ref(null)
const submitLoading = ref(false)

const uploadUrl = '/api/upload/image'
const uploadHeaders = { Authorization: 'Bearer ' + getToken() }

const formData = reactive({
  amount: null,
  payment_date: '',
  payment_method: '',
  remark: '',
  voucher_url: '',
})

const formRules = {
  amount: [{ required: true, message: '请输入回款金额', trigger: 'blur' }],
  payment_date: [{ required: true, message: '请选择回款日期', trigger: 'change' }],
  payment_method: [{ required: true, message: '请选择回款方式', trigger: 'change' }],
}

const beforeUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) {
    ElMessage.error('只能上传图片文件')
    return false
  }
  if (!isLt5M) {
    ElMessage.error('图片大小不能超过5MB')
    return false
  }
  return true
}

const handleUploadSuccess = (response) => {
  if (response.errno === 0 && response.data && response.data.url) {
    formData.voucher_url = response.data.url
  } else {
    ElMessage.error('上传失败')
  }
}

watch(
  () => props.visible,
  (val) => {
    if (val) {
      formData.amount = null
      formData.payment_date = ''
      formData.payment_method = ''
      formData.remark = ''
      formData.voucher_url = ''
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
    await createPayment({
      order_id: props.orderId,
      amount: String(formData.amount),
      payment_date: formData.payment_date,
      payment_method: formData.payment_method,
      remark: formData.remark,
      voucher_url: formData.voucher_url,
    })
    ElMessage.success('回款录入成功')
    handleClose()
    emit('success')
  } catch (error) {
    console.error('回款录入失败:', error)
  } finally {
    submitLoading.value = false
  }
}
</script>

<style scoped>
.voucher-uploader .el-upload {
  border: 1px dashed var(--el-border-color);
  border-radius: 6px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  width: 120px;
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.voucher-uploader .el-upload:hover {
  border-color: var(--el-color-primary);
}
.voucher-preview {
  width: 120px;
  height: 120px;
}
.voucher-uploader-icon {
  font-size: 28px;
  color: #8c939d;
}
</style>
