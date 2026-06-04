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
        <span class="price">{{ formatCurrency(order?.final_amount) }}</span>
      </el-form-item>
      <el-form-item label="退货金额" prop="return_amount">
        <el-input-number
          v-model="formData.return_amount"
          :min="0"
          :max="order?.final_amount || 0"
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
      <!-- 出库后退货才需要选择退入仓库 -->
      <el-form-item v-if="isDelivered" label="退货仓库" prop="warehouse_id">
        <el-select
          v-model="formData.warehouse_id"
          placeholder="请选择退货仓库"
          style="width: 100%"
        >
          <el-option
            v-for="wh in warehouseList"
            :key="wh.id"
            :label="wh.warehouse_name"
            :value="wh.id"
          />
        </el-select>
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
import { ref, reactive, watch, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { returnOrder } from '@/api/order'
import { getWarehouseList } from '@/api/warehouse'
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
const warehouseList = ref([])

// 是否已出库（delivery_status > 0 表示已出库）
const isDelivered = computed(() => {
  return props.order?.delivery_status > 0 || props.order?.outbound_confirmed === true
})

const formData = reactive({
  return_amount: 0,
  profit_deduction: 0,
  warehouse_id: null,
  reason: '',
})

// 动态表单校验规则
const formRules = computed(() => ({
  return_amount: [{ required: true, message: '请输入退货金额', trigger: 'blur' }],
  warehouse_id: isDelivered.value ? [{ required: true, message: '请选择退货仓库', trigger: 'change' }] : [],
  reason: [{ required: true, message: '请输入退货原因', trigger: 'blur' }],
}))

// 加载仓库列表
const loadWarehouseList = async () => {
  try {
    const res = await getWarehouseList({ page: 1, page_size: 100 })
    // API 返回的数据直接是数组
    warehouseList.value = Array.isArray(res.data) ? res.data : (res.data?.list || [])
  } catch (err) {
    console.error('加载仓库列表失败:', err)
  }
}

watch(
  () => props.visible,
  (val) => {
    if (val) {
      formData.return_amount = 0
      formData.profit_deduction = 0
      formData.warehouse_id = null
      formData.reason = ''
      // 只有出库后才需要加载仓库列表
      if (isDelivered.value) {
        loadWarehouseList()
      }
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
    const params = {
      return_amount: Number(formData.return_amount) || 0,
      return_profit: Number(formData.profit_deduction) || 0,
      reason: formData.reason,
    }
    // 出库后退货才传递仓库ID
    if (isDelivered.value) {
      params.warehouse_id = formData.warehouse_id
    }
    await returnOrder(props.order.id, params)
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
