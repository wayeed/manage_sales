<template>
  <el-dialog
    v-model="visible"
    title="出库确认"
    width="800px"
    :close-on-click-modal="false"
    class="outbound-dialog"
  >
    <div v-if="order" class="outbound-content">
      <!-- 仓库选择 -->
      <el-form :model="form" label-width="100px" class="warehouse-form">
        <el-form-item label="出库仓库" required>
          <el-select v-model="form.warehouse_id" placeholder="请选择出库仓库" style="width: 100%">
            <el-option
              v-for="wh in warehouseList"
              :key="wh.id"
              :label="wh.warehouse_name"
              :value="wh.id"
            />
          </el-select>
        </el-form-item>
      </el-form>

      <!-- 商品库存状态 -->
      <div class="stock-section">
        <div class="section-title">
          <span>商品库存状态</span>
          <el-tag v-if="allStockSufficient" type="success" size="small">库存充足</el-tag>
          <el-tag v-else type="danger" size="small">库存不足</el-tag>
        </div>

        <el-table :data="stockList" border size="small" style="width: 100%">
          <el-table-column prop="product_name" label="商品名称" min-width="150" />
          <el-table-column prop="sku_name" label="规格" width="120" />
          <el-table-column label="订单数量" width="90" align="center">
            <template #default="{ row }">
              <span class="order-qty">{{ row.quantity }}</span>
            </template>
          </el-table-column>
          <el-table-column label="可用库存" width="100" align="center">
            <template #default="{ row }">
              <span :class="{ 'insufficient': row.locked_qty < row.quantity }">
                {{ row.available_qty }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="锁定库存" width="100" align="center">
            <template #default="{ row }">
              <span class="locked-qty">{{ row.locked_qty || 0 }}</span>
            </template>
          </el-table-column>
          <el-table-column label="批次信息" min-width="150">
            <template #default="{ row }">
              <div v-if="row.batches && row.batches.length > 0" class="batch-list">
                <el-tag
                  v-for="batch in row.batches"
                  :key="batch.batch_id"
                  size="small"
                  type="info"
                  class="batch-tag"
                >
                  {{ batch.batch_no }}: {{ batch.quantity }}件
                </el-tag>
              </div>
              <span v-else class="no-batch">-</span>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="80" align="center">
            <template #default="{ row }">
              <el-tag v-if="(row.batches && row.batches.length > 0 && Number(row.locked_qty || 0) >= Number(row.quantity || 0)) || Number(row.available_qty || 0) >= Number(row.quantity || 0)" type="success" size="small">充足</el-tag>
              <el-tag v-else type="danger" size="small">不足</el-tag>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 出库信息 -->
      <div class="delivery-info">
        <div class="section-title">出库信息</div>
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="订单号">{{ order.order_no }}</el-descriptions-item>
          <el-descriptions-item label="客户">{{ order.customer_name }}</el-descriptions-item>
          <el-descriptions-item label="收货人">{{ order.customer_name }}</el-descriptions-item>
          <el-descriptions-item label="联系电话">{{ order.customer_phone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="送货地址" :span="2">{{ order.customer_address || '-' }}</el-descriptions-item>
        </el-descriptions>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button
          type="primary"
          :disabled="!canConfirm"
          :loading="loading"
          @click="handleConfirm"
        >
          确认出库并打印送货单
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { getWarehouseList } from '@/api/warehouse'
import { getOrderStockStatus, createDelivery } from '@/api/delivery'

const props = defineProps({
  modelValue: Boolean,
  order: Object,
})

const emit = defineEmits(['update:modelValue', 'success', 'print'])

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})

const loading = ref(false)
const warehouseList = ref([])
const stockList = ref([])

const form = ref({
  warehouse_id: null,
})

// 计算属性
const allStockSufficient = computed(() => {
  return stockList.value.every(item => {
    const locked = Number(item.locked_qty || 0)
    const qty = Number(item.quantity || 0)
    const hasBatches = item.batches && item.batches.length > 0
    // 有批次且锁定库存足够 → 充足；无批次但可用库存足够 → 充足
    if (hasBatches) return locked >= qty
    return Number(item.available_qty || 0) >= qty
  })
})

const canConfirm = computed(() => {
  return form.value.warehouse_id && allStockSufficient.value && stockList.value.length > 0
})

// 获取库存状态描述
const getStockStatus = (row) => {
  const locked = Number(row.locked_qty || 0)
  const available = Number(row.available_qty || 0)
  const qty = Number(row.quantity || 0)
  const hasBatches = row.batches && row.batches.length > 0
  if (hasBatches && locked >= qty) return { text: '锁定充足', type: 'success' }
  if (available >= qty) return { text: '可用充足', type: 'warning' }
  return { text: '库存不足', type: 'danger' }
}

// 获取仓库列表
const fetchWarehouses = async () => {
  try {
    const res = await getWarehouseList({ page: 1, page_size: 100 })
    warehouseList.value = res.data?.list || res.data || []
    // 默认选中第一个仓库
    if (warehouseList.value.length > 0 && !form.value.warehouse_id) {
      form.value.warehouse_id = warehouseList.value[0].id
    }
  } catch (error) {
    console.error('获取仓库列表失败:', error)
  }
}

// 获取库存状态
const fetchStockStatus = async () => {
  if (!props.order?.id || !form.value.warehouse_id) return

  try {
    const res = await getOrderStockStatus(props.order.id, form.value.warehouse_id)
    stockList.value = res.data || []
  } catch (error) {
    console.error('获取库存状态失败:', error)
    ElMessage.error('获取库存状态失败')
  }
}

// 监听弹窗显示
watch(() => props.modelValue, (val) => {
  if (val && props.order) {
    form.value.warehouse_id = null
    stockList.value = []
    fetchWarehouses().then(() => {
      fetchStockStatus()
    })
  }
})

// 监听仓库变化
watch(() => form.value.warehouse_id, () => {
  fetchStockStatus()
})

const handleClose = () => {
  visible.value = false
}

const handleConfirm = async () => {
  if (!canConfirm.value) return

  loading.value = true
  try {
    // 创建送货记录（实际出库），传入商品明细
    const items = (props.order.items || []).map(item => ({
      order_item_id: item.id,
      quantity: item.quantity,
    }))
    await createDelivery({
      order_id: props.order.id,
      warehouse_id: form.value.warehouse_id,
      receiver_name: props.order.customer_name,
      receiver_phone: props.order.customer_phone,
      receiver_address: props.order.customer_address,
      items,
    })

    ElMessage.success('出库成功')
    emit('success')
    emit('print') // 触发打印
    handleClose()
  } catch (error) {
    console.error('出库失败:', error)
    ElMessage.error(error.message || '出库失败')
  } finally {
    loading.value = false
  }
}
</script>

<style lang="scss" scoped>
.outbound-dialog {
  .outbound-content {
    max-height: 60vh;
    overflow-y: auto;
  }

  .warehouse-form {
    margin-bottom: 20px;
    padding: 16px;
    background: #f5f7fa;
    border-radius: 4px;
  }

  .section-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 600;
    font-size: 14px;
    margin-bottom: 12px;
    padding-bottom: 8px;
    border-bottom: 1px solid #e4e7ed;
  }

  .stock-section {
    margin-bottom: 20px;

    .order-qty {
      font-weight: 600;
      color: #409eff;
    }

    .locked-qty {
      color: #e6a23c;
    }

    .insufficient {
      color: #f56c6c;
      font-weight: 600;
    }

    .batch-list {
      display: flex;
      flex-wrap: wrap;
      gap: 4px;

      .batch-tag {
        font-size: 11px;
      }
    }

    .no-batch {
      color: #909399;
    }
  }

  .delivery-info {
    .section-title {
      margin-top: 16px;
    }
  }

  .dialog-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }
}
</style>
