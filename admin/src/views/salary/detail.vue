<template>
  <div class="salary-detail">
    <!-- 返回按钮 -->
    <div class="page-header">
      <el-button icon="ArrowLeft" @click="handleBack">返回列表</el-button>
      <span class="page-title">工资详情</span>
      <el-tag :type="getSalaryStatusTag(salaryData.status)" size="large">
        {{ getSalaryStatusLabel(salaryData.status) }}
      </el-tag>
    </div>

    <!-- 工资基本信息 -->
    <el-card class="info-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">基本信息</span>
          <div class="header-actions">
            <el-button
              v-if="salaryData.status === 1"
              type="warning"
              @click="handleConfirm"
            >
              审核确认
            </el-button>
            <el-button
              v-if="salaryData.status === 2"
              type="success"
              @click="handlePay"
            >
              发放工资
            </el-button>
            <el-button
              type="primary"
              icon="Download"
              @click="handleExport"
            >
              导出工资条
            </el-button>
          </div>
        </div>
      </template>

      <el-descriptions :column="4" border>
        <el-descriptions-item label="员工">
          {{ salaryData.employee?.real_name || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="月份">
          {{ salaryData.salary_month }}
        </el-descriptions-item>
        <el-descriptions-item label="基本工资">
          <span class="amount">{{ formatCurrency(salaryData.base_salary) }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="销售提成">
          <span class="commission">{{ formatCurrency(salaryData.sales_commission) }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="团队分润">
          <span class="commission">{{ formatCurrency(salaryData.team_commission) }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="基金池奖励">
          <span class="commission">{{ formatCurrency(salaryData.fund_reward) }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="老带新奖励">
          <span class="commission">{{ formatCurrency(salaryData.referral_reward) }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="扣款">
          <span class="deduction">{{ formatCurrency(salaryData.deduction) }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="应发总额">
          <span class="gross">{{ formatCurrency(salaryData.gross_salary) }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="实发总额">
          <span class="net">{{ formatCurrency(salaryData.net_salary) }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="发放方式">
          {{ getPayMethodLabel(salaryData.pay_method) }}
        </el-descriptions-item>
        <el-descriptions-item label="发放时间">
          {{ salaryData.paid_at || '-' }}
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <!-- 工资明细表格 -->
    <el-card class="detail-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">工资明细</span>
        </div>
      </template>

      <el-table
        v-loading="detailLoading"
        :data="detailList"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column label="项目类型" width="140">
          <template #default="{ row }">
            <el-tag :type="getItemTypeTag(row.item_type)" size="small">
              {{ getItemTypeLabel(row.item_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="item_name" label="项目名称" min-width="200" show-overflow-tooltip />
        <el-table-column label="金额" width="160" align="right">
          <template #default="{ row }">
            <span :class="row.amount >= 0 ? 'income' : 'deduction'">
              {{ row.amount >= 0 ? '+' : '' }}{{ formatCurrency(row.amount) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="200" show-overflow-tooltip />
      </el-table>
    </el-card>

    <!-- 发放弹窗 -->
    
<el-dialog v-dialog-drag
      v-model="payDialogVisible"
      title="发放工资"
      width="480px"
      destroy-on-close
    >
      <el-form
        ref="payFormRef"
        :model="payForm"
        :rules="payFormRules"
        label-width="100px"
      >
        <el-form-item label="员工">
          <el-input :model-value="salaryData.employee?.real_name" disabled />
        </el-form-item>
        <el-form-item label="实发总额">
          <el-input :model-value="formatCurrency(salaryData.net_salary)" disabled />
        </el-form-item>
        <el-form-item label="发放方式" prop="pay_method">
          <el-select v-model="payForm.pay_method" placeholder="请选择发放方式" style="width: 100%">
            <el-option label="银行转账" value="bank_transfer" />
            <el-option label="现金" value="cash" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input
            v-model="payForm.remark"
            type="textarea"
            :rows="3"
            placeholder="请输入备注"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="payDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handlePaySubmit">
          确认发放
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getSalaryDetail, confirmSalary, paySalary, exportSalarySlip } from '@/api/salary'
import { formatCurrency } from '@/utils/format'

const route = useRoute()
const router = useRouter()

const salaryId = route.params.id

// ==================== 工资数据 ====================
const salaryData = ref({})
const detailLoading = ref(false)
const detailList = ref([])

const fetchDetail = async () => {
  try {
    const res = await getSalaryDetail(salaryId)
    salaryData.value = res.data || {}
    detailList.value = res.data?.items || []
  } catch (error) {
    console.error('获取工资详情失败:', error)
  }
}

// ==================== 状态映射 ====================
const getSalaryStatusLabel = (status) => {
  const map = {
    1: '草稿',
    2: '已确认',
    3: '已发放',
  }
  return map[status] || status || '未知'
}

const getSalaryStatusTag = (status) => {
  const map = {
    1: 'info',
    2: 'warning',
    3: 'success',
  }
  return map[status] || 'info'
}

const getPayMethodLabel = (method) => {
  const map = {
    bank_transfer: '银行转账',
    cash: '现金',
    other: '其他',
  }
  return map[method] || method || '-'
}

const getItemTypeLabel = (type) => {
  const map = {
    base: '基本工资',
    commission: '提成',
    team_share: '团队分润',
    fund_pool: '基金池奖励',
    referral: '老带新奖励',
    deduction: '扣款',
    other: '其他',
  }
  return map[type] || type || '未知'
}

const getItemTypeTag = (type) => {
  const map = {
    base: '',
    commission: 'success',
    team_share: 'success',
    fund_pool: 'warning',
    referral: 'warning',
    deduction: 'danger',
    other: 'info',
  }
  return map[type] || 'info'
}

// ==================== 操作 ====================
const handleBack = () => {
  router.push('/salary')
}

const handleConfirm = () => {
  ElMessageBox.confirm(
    '确定要审核确认该工资记录吗？确认后将进入待发放状态。',
    '审核确认',
    {
      confirmButtonText: '确定确认',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await confirmSalary(salaryId)
      ElMessage.success('审核确认成功')
      fetchDetail()
    } catch (error) {
      console.error('审核确认失败:', error)
    }
  }).catch(() => {})
}

// ==================== 发放 ====================
const payDialogVisible = ref(false)
const submitLoading = ref(false)
const payFormRef = ref(null)

const payForm = reactive({
  pay_method: 'bank_transfer',
  remark: '',
})

const payFormRules = {
  pay_method: [{ required: true, message: '请选择发放方式', trigger: 'change' }],
}

const handlePay = () => {
  payForm.pay_method = 'bank_transfer'
  payForm.remark = ''
  payDialogVisible.value = true
}

const handlePaySubmit = async () => {
  const valid = await payFormRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    await paySalary(salaryId, {
      pay_method: payForm.pay_method,
      remark: payForm.remark,
    })
    ElMessage.success('发放成功')
    payDialogVisible.value = false
    fetchDetail()
  } catch (error) {
    console.error('发放工资失败:', error)
  } finally {
    submitLoading.value = false
  }
}

// ==================== 导出 ====================
const handleExport = async () => {
  try {
    // 文件下载需要设置 responseType: 'blob'
    const res = await exportSalarySlip(salaryId)
    const blob = new Blob([res], { type: 'text/csv;charset=utf-8' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `工资条_${salaryData.value.employee?.real_name || ''}_${salaryData.value.salary_month || ''}.csv`
    link.click()
    window.URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch (error) {
    console.error('导出工资条失败:', error)
  }
}

// ==================== 初始化 ====================
onMounted(() => {
  fetchDetail()
})
</script>

<style lang="scss" scoped>
.salary-detail {
  .page-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 16px;
  }

  .page-title {
    font-size: 18px;
    font-weight: 600;
    color: #303133;
  }

  .info-card {
    margin-bottom: 16px;
  }

  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;

    .title {
      font-size: 16px;
      font-weight: 600;
      color: #303133;
    }

    .header-actions {
      display: flex;
      gap: 8px;
    }
  }

  .amount {
    color: #303133;
    font-weight: 500;
  }

  .commission {
    color: #67c23a;
    font-weight: 500;
  }

  .deduction {
    color: #f56c6c;
    font-weight: 500;
  }

  .gross {
    color: #e6a23c;
    font-weight: 600;
  }

  .net {
    color: #409eff;
    font-weight: 700;
  }

  .income {
    color: #67c23a;
    font-weight: 500;
  }
}
</style>
