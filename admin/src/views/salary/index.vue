<template>
  <div class="salary-manage">
    <!-- 搜索栏 -->
    <el-card class="search-card" shadow="never">
      <el-form :model="searchForm" inline>
        <el-form-item label="月份">
          <el-date-picker
            v-model="searchForm.month"
            type="month"
            placeholder="请选择月份"
            value-format="YYYY-MM"
            style="width: 160px"
          />
        </el-form-item>
        <el-form-item label="员工">
          <el-input
            v-model="searchForm.keyword"
            placeholder="员工姓名"
            clearable
            style="width: 160px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select
            v-model="searchForm.status"
            placeholder="全部状态"
            clearable
            style="width: 140px"
          >
            <el-option label="草稿" :value="1" />
            <el-option label="已确认" :value="2" />
            <el-option label="已发放" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="Search" @click="handleSearch">搜索</el-button>
          <el-button icon="Refresh" @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 数据表格 -->
    <el-card class="table-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">工资列表</span>
          <div class="header-actions">
            <el-button type="success" icon="DocumentAdd" @click="handleGenerate">
              生成工资
            </el-button>
            <el-button
              v-if="selectedRows.length > 0"
              type="warning"
              @click="handleBatchConfirm"
            >
              批量审核 ({{ selectedRows.length }})
            </el-button>
            <el-button
              v-if="selectedRows.length > 0"
              type="primary"
              @click="handleBatchPay"
            >
              批量发放 ({{ selectedRows.length }})
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="salaryList"
        border
        stripe
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="50" align="center" />
        <el-table-column label="员工" width="120" fixed="left">
          <template #default="{ row }">
            {{ row.employee?.real_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="salary_month" label="月份" width="100" align="center" />
        <el-table-column label="基本工资" width="120" align="right">
          <template #default="{ row }">
            {{ formatCurrency(row.base_salary) }}
          </template>
        </el-table-column>
        <el-table-column label="销售提成" width="120" align="right">
          <template #default="{ row }">
            <span class="commission">{{ formatCurrency(row.sales_commission) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="团队分润" width="120" align="right">
          <template #default="{ row }">
            <span class="commission">{{ formatCurrency(row.team_commission) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="基金池奖励" width="120" align="right">
          <template #default="{ row }">
            <span class="commission">{{ formatCurrency(row.fund_reward) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="老带新奖励" width="120" align="right">
          <template #default="{ row }">
            <span class="commission">{{ formatCurrency(row.referral_reward) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="扣款" width="110" align="right">
          <template #default="{ row }">
            <span class="deduction">{{ formatCurrency(row.deduction) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="应发总额" width="130" align="right">
          <template #default="{ row }">
            <span class="gross">{{ formatCurrency(row.gross_salary) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="实发总额" width="130" align="right">
          <template #default="{ row }">
            <span class="net">{{ formatCurrency(row.net_salary) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center" fixed="right">
          <template #default="{ row }">
            <el-tag :type="getSalaryStatusTag(row.status)" size="small">
              {{ getSalaryStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleDetail(row)">
              查看详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.page_size"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @size-change="fetchList"
          @current-change="fetchList"
        />
      </div>
    </el-card>

    <!-- 生成工资弹窗 -->
    
<el-dialog v-dialog-drag
      v-model="generateDialogVisible"
      title="生成工资"
      width="420px"
      destroy-on-close
    >
      <el-form
        ref="generateFormRef"
        :model="generateForm"
        :rules="generateFormRules"
        label-width="80px"
      >
        <el-form-item label="月份" prop="month">
          <el-date-picker
            v-model="generateForm.month"
            type="month"
            placeholder="请选择月份"
            value-format="YYYY-MM"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="generateDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleGenerateSubmit">
          确认生成
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getSalaryList, generateSalary, confirmSalary, paySalary } from '@/api/salary'
import { formatCurrency } from '@/utils/format'

const router = useRouter()

// ==================== 搜索与列表 ====================
const loading = ref(false)
const salaryList = ref([])
const selectedRows = ref([])

const searchForm = reactive({
  month: '',
  keyword: '',
  status: '',
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
})

const fetchList = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.page_size,
      ...getSearchParams(),
    }
    const res = await getSalaryList(params)
    salaryList.value = res.data?.list || res.data || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取工资列表失败:', error)
  } finally {
    loading.value = false
  }
}

const getSearchParams = () => {
  const params = {}
  if (searchForm.month) params.month = searchForm.month
  if (searchForm.keyword) params.keyword = searchForm.keyword
  if (searchForm.status) params.status = searchForm.status
  return params
}

const handleSearch = () => {
  pagination.page = 1
  fetchList()
}

const handleReset = () => {
  searchForm.month = ''
  searchForm.keyword = ''
  searchForm.status = ''
  pagination.page = 1
  fetchList()
}

const handleSelectionChange = (rows) => {
  selectedRows.value = rows
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

// ==================== 操作 ====================
const handleDetail = (row) => {
  router.push(`/salary/detail/${row.id}`)
}

// ==================== 生成工资 ====================
const generateDialogVisible = ref(false)
const submitLoading = ref(false)
const generateFormRef = ref(null)

const generateForm = reactive({
  month: '',
})

const generateFormRules = {
  month: [{ required: true, message: '请选择月份', trigger: 'change' }],
}

const handleGenerate = () => {
  generateForm.month = ''
  generateDialogVisible.value = true
}

const handleGenerateSubmit = async () => {
  const valid = await generateFormRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    await generateSalary({ month: generateForm.month })
    ElMessage.success('工资生成成功')
    generateDialogVisible.value = false
    fetchList()
  } catch (error) {
    console.error('生成工资失败:', error)
  } finally {
    submitLoading.value = false
  }
}

// ==================== 批量审核 ====================
const handleBatchConfirm = () => {
  const draftRows = selectedRows.value.filter((r) => r.status === 1)
  if (draftRows.length === 0) {
    ElMessage.warning('请选择草稿状态的工资记录')
    return
  }

  ElMessageBox.confirm(
    `确定要审核 ${draftRows.length} 条草稿工资记录吗？`,
    '批量审核确认',
    {
      confirmButtonText: '确定审核',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      const promises = draftRows.map((row) => confirmSalary(row.id))
      await Promise.all(promises)
      ElMessage.success('批量审核成功')
      fetchList()
    } catch (error) {
      console.error('批量审核失败:', error)
    }
  }).catch(() => {})
}

// ==================== 批量发放 ====================
const handleBatchPay = () => {
  const confirmedRows = selectedRows.value.filter((r) => r.status === 2)
  if (confirmedRows.length === 0) {
    ElMessage.warning('请选择已确认状态的工资记录')
    return
  }

  ElMessageBox.confirm(
    `确定要发放 ${confirmedRows.length} 条已确认工资记录吗？`,
    '批量发放确认',
    {
      confirmButtonText: '确定发放',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      const promises = confirmedRows.map((row) =>
        paySalary(row.id, { pay_method: 'bank_transfer', remark: '批量发放' })
      )
      await Promise.all(promises)
      ElMessage.success('批量发放成功')
      fetchList()
    } catch (error) {
      console.error('批量发放失败:', error)
    }
  }).catch(() => {})
}

// ==================== 初始化 ====================
onMounted(() => {
  fetchList()
})
</script>

<style lang="scss" scoped>
.salary-manage {
  .search-card {
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

  .table-card {
    .pagination-wrapper {
      display: flex;
      justify-content: flex-end;
      margin-top: 16px;
    }
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
}
</style>
