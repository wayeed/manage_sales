<template>
  
<el-dialog v-dialog-drag
    :model-value="visible"
    title="导出订单"
    width="500px"
    destroy-on-close
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="formData"
      label-width="100px"
    >
      <el-form-item label="导出范围">
        <el-radio-group v-model="formData.range">
          <el-radio value="current">当前筛选结果</el-radio>
          <el-radio value="all">全部数据</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="导出字段">
        <el-checkbox-group v-model="formData.fields">
          <el-checkbox label="order_no">订单号</el-checkbox>
          <el-checkbox label="customer">客户</el-checkbox>
          <el-checkbox label="salesman">业务员</el-checkbox>
          <el-checkbox label="order_type">订单类型</el-checkbox>
          <el-checkbox label="total_amount">订单金额</el-checkbox>
          <el-checkbox label="actual_profit">实际利润</el-checkbox>
          <el-checkbox label="payment_status">回款状态</el-checkbox>
          <el-checkbox label="status">订单状态</el-checkbox>
          <el-checkbox label="created_at">创建时间</el-checkbox>
        </el-checkbox-group>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" :loading="exportLoading" @click="handleExport">
        确认导出
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, watch } from 'vue'
import { ElMessage } from 'element-plus'
import ExcelJS from 'exceljs'
import { saveAs } from 'file-saver'
import { getOrderList } from '@/api/order'
import { formatCurrency } from '@/utils/format'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false,
  },
  searchParams: {
    type: Object,
    default: () => ({}),
  },
})

const emit = defineEmits(['update:visible'])

const exportLoading = ref(false)

const formData = reactive({
  range: 'current',
  fields: [
    'order_no', 'customer', 'salesman', 'order_type',
    'total_amount', 'actual_profit', 'payment_status', 'status', 'created_at',
  ],
})

watch(
  () => props.visible,
  (val) => {
    if (val) {
      formData.range = 'current'
      formData.fields = [
        'order_no', 'customer', 'salesman', 'order_type',
        'total_amount', 'actual_profit', 'payment_status', 'status', 'created_at',
      ]
    }
  }
)

const handleClose = () => {
  emit('update:visible', false)
}

const fieldLabels = {
  order_no: '订单号',
  customer: '客户',
  salesman: '业务员',
  order_type: '订单类型',
  total_amount: '订单金额',
  actual_profit: '实际利润',
  payment_status: '回款状态',
  status: '订单状态',
  created_at: '创建时间',
}

const orderTypeMap = { 1: '零售', 2: '批发', 3: '定制', 4: '同行', 5: '礼品', 6: '退货' }
const paymentStatusMap = { 0: '未回款', 1: '部分回款', 2: '已回款', 3: '已退款' }
const statusMap = { 0: '待审批', 1: '已生效', 2: '已发货', 3: '已完成', 4: '已取消', 5: '已退货' }

const handleExport = async () => {
  if (formData.fields.length === 0) {
    ElMessage.warning('请至少选择一个导出字段')
    return
  }

  exportLoading.value = true
  try {
    // 获取数据
    const params = formData.range === 'current'
      ? { ...props.searchParams, page: 1, page_size: 10000 }
      : { page: 1, page_size: 10000 }

    const res = await getOrderList(params)
    const list = res.data?.list || res.data || []

    if (list.length === 0) {
      ElMessage.warning('没有可导出的数据')
      return
    }

    // 转换数据
    const headers = formData.fields.map((f) => fieldLabels[f])
    const rows = list.map((row) => {
      return formData.fields.map((f) => {
        switch (f) {
          case 'customer':
            return row.customer?.customer_name || '-'
          case 'salesman':
            return row.salesman?.real_name || '-'
          case 'order_type':
            return orderTypeMap[row.order_type] || row.order_type || '-'
          case 'total_amount':
            return row.final_amount || 0
          case 'actual_profit':
            return row.actual_profit || 0
          case 'payment_status':
            return paymentStatusMap[row.payment_status] || row.payment_status || '-'
          case 'status':
            return statusMap[row.status] || row.status || '-'
          default:
            return row[f] || ''
        }
      })
    })

    // 生成 Excel（使用 exceljs）
    const workbook = new ExcelJS.Workbook()
    const sheet = workbook.addWorksheet('订单列表')

    // 写入表头
    const headerRow = sheet.addRow(headers)
    headerRow.font = { bold: true }

    // 写入数据行
    rows.forEach((row) => {
      sheet.addRow(row)
    })

    const fileName = `订单列表_${new Date().toISOString().slice(0, 10)}.xlsx`
    const buffer = await workbook.xlsx.writeBuffer()
    const blob = new Blob([buffer], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
    saveAs(blob, fileName)

    ElMessage.success(`成功导出 ${list.length} 条数据`)
    handleClose()
  } catch (error) {
    console.error('导出失败:', error)
    ElMessage.error('导出失败')
  } finally {
    exportLoading.value = false
  }
}
</script>
