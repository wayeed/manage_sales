<template>
  <el-dialog v-model="dialogVisible" title="送货出库" width="700px" destroy-on-close v-dialog-drag>
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <!-- 订单信息（只读） -->
      <el-divider content-position="left">订单信息</el-divider>
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="订单编号">
            <el-input :model-value="form.order_no" disabled />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="客户姓名">
            <el-input :model-value="form.customer_name" disabled />
          </el-form-item>
        </el-col>
      </el-row>

      <!-- 送货信息 -->
      <el-divider content-position="left">送货信息</el-divider>
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="出库仓库" prop="warehouse_id">
            <el-select v-model="form.warehouse_id" placeholder="请选择仓库" style="width: 100%">
              <el-option v-for="w in warehouseList" :key="w.id" :label="w.warehouse_name" :value="w.id" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="送货类型" prop="delivery_type">
            <el-select v-model="form.delivery_type" placeholder="请选择" style="width: 100%">
              <el-option label="自送" :value="1" />
              <el-option label="物流" :value="2" />
              <el-option label="快递" :value="3" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="送货时间" prop="delivery_time">
            <el-date-picker v-model="form.delivery_time" type="datetime" placeholder="请选择送货时间" value-format="YYYY-MM-DD HH:mm:ss" style="width: 100%" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="物流单号">
            <el-input v-model="form.logistics_no" placeholder="物流/快递单号" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="收货人">
            <el-input v-model="form.receiver_name" placeholder="收货人姓名" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="联系电话">
            <el-input v-model="form.receiver_phone" placeholder="收货人电话" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="收货地址">
        <el-input v-model="form.receiver_address" placeholder="收货地址" />
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="form.remark" type="textarea" :rows="2" placeholder="备注信息" />
      </el-form-item>

      <!-- 商品明细 -->
      <el-divider content-position="left">商品明细</el-divider>
      <el-table :data="form.items" border size="small">
        <el-table-column prop="product_name" label="商品名称" min-width="140" />
        <el-table-column prop="sku_name" label="规格" min-width="120" />
        <el-table-column prop="quantity" label="订单数量" width="90" align="center" />
        <el-table-column label="送货数量" width="120" align="center">
          <template #default="{ row }">
            <el-input-number v-model="row.delivery_qty" :min="1" :max="row.quantity" size="small" controls-position="right" style="width: 90px" />
          </template>
        </el-table-column>
      </el-table>
    </el-form>

    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">确认送货出库</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { createDelivery } from '@/api/delivery'
import { getWarehouseList } from '@/api/inventory'

const props = defineProps({
  visible: { type: Boolean, default: false },
  order: { type: Object, default: null },
})
const emit = defineEmits(['update:visible', 'success'])

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
})

const formRef = ref(null)
const submitting = ref(false)
const warehouseList = ref([])

const form = reactive({
  order_id: null,
  order_no: '',
  customer_name: '',
  warehouse_id: null,
  delivery_type: 1,
  delivery_time: '',
  logistics_no: '',
  receiver_name: '',
  receiver_phone: '',
  receiver_address: '',
  remark: '',
  items: [],
})

const rules = {
  warehouse_id: [{ required: true, message: '请选择出库仓库', trigger: 'change' }],
  delivery_type: [{ required: true, message: '请选择送货类型', trigger: 'change' }],
  delivery_time: [{ required: true, message: '请选择送货时间', trigger: 'change' }],
}

// 监听弹窗打开
watch(() => props.visible, async (val) => {
  if (val) {
    await nextTick()
    resetForm()
    if (props.order) {
      form.order_id = props.order.order_id
      form.order_no = props.order.order_no
      form.customer_name = props.order.customer_name
      form.receiver_name = props.order.customer_name || ''
      form.receiver_phone = props.order.customer_phone || ''
      form.receiver_address = props.order.customer_address || ''
      // 设置商品明细
      if (props.order.items?.length) {
        form.items = props.order.items.map(item => ({
          order_item_id: item.order_item_id,
          sku_id: item.sku_id,
          product_name: item.product_name,
          sku_name: item.sku_name,
          quantity: item.quantity,
          delivery_qty: item.quantity,
        }))
      }
    }
    // 加载仓库列表
    fetchWarehouses()
  }
})

const fetchWarehouses = async () => {
  try {
    const res = await getWarehouseList({ page: 1, page_size: 100, status: 1 })
    warehouseList.value = res.data?.list || res.data || []
    // 默认选中主仓库
    const mainWarehouse = warehouseList.value.find(w => w.warehouse_type === 1)
    if (mainWarehouse) form.warehouse_id = mainWarehouse.id
  } catch (error) {
    console.error('获取仓库列表失败:', error)
  }
}

const resetForm = () => {
  form.order_id = null
  form.order_no = ''
  form.customer_name = ''
  form.warehouse_id = null
  form.delivery_type = 1
  form.delivery_time = ''
  form.logistics_no = ''
  form.receiver_name = ''
  form.receiver_phone = ''
  form.receiver_address = ''
  form.remark = ''
  form.items = []
  formRef.value?.clearValidate()
}

const handleClose = () => { emit('update:visible', false) }

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const data = {
      order_id: form.order_id,
      warehouse_id: form.warehouse_id,
      delivery_type: form.delivery_type,
      delivery_time: form.delivery_time,
      logistics_no: form.logistics_no,
      receiver_name: form.receiver_name,
      receiver_phone: form.receiver_phone,
      receiver_address: form.receiver_address,
      remark: form.remark,
      items: form.items.map(item => ({
        order_item_id: item.order_item_id,
        quantity: item.delivery_qty,
      })),
    }
    await createDelivery(data)
    ElMessage.success('送货出库成功')
    handleClose()
    emit('success')
  } catch (error) {
    console.error('送货出库失败:', error)
  } finally {
    submitting.value = false
  }
}
</script>

<script>
export default { name: 'DeliveryFormDialog' }
</script>

<style lang="scss" scoped>
.el-divider {
  margin: 16px 0 12px;
}
</style>
