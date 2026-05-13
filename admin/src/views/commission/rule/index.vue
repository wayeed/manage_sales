<template>
  <div class="commission-rule">
    <!-- 配置项表格 -->
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">提成规则配置</span>
          <el-button type="primary" icon="Refresh" @click="fetchConfigList">刷新</el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="configList"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column prop="config_key" label="配置键" width="220" show-overflow-tooltip />
        <el-table-column label="当前值" width="160" align="center">
          <template #default="{ row }">
            <span class="value-text">{{ formatConfigValue(row) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="说明" min-width="240" show-overflow-tooltip />
        <el-table-column prop="config_type" label="分组" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="getGroupTag(row.config_type)" size="small">
              {{ getGroupLabel(row.config_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">
              编辑
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 编辑弹窗 -->
    
<el-dialog v-dialog-drag
      v-model="editDialogVisible"
      title="编辑配置"
      width="480px"
      destroy-on-close
    >
      <el-form
        ref="editFormRef"
        :model="editForm"
        :rules="editFormRules"
        label-width="100px"
      >
        <el-form-item label="配置项">
          <el-input :model-value="currentConfig?.config_key" disabled />
        </el-form-item>
        <el-form-item label="配置值" prop="value">
          <el-input-number
            v-if="isPercentConfig(currentConfig?.config_key)"
            v-model="editForm.value"
            :min="0"
            :max="100"
            :precision="1"
            :step="0.5"
            controls-position="right"
            style="width: 100%"
          />
          <el-input
            v-else
            v-model="editForm.value"
            placeholder="请输入配置值"
          />
        </el-form-item>
        <el-form-item v-if="isPercentConfig(currentConfig?.config_key)" label="">
          <span style="color: #909399; font-size: 12px">单位：百分比（%）</span>
        </el-form-item>
        <el-form-item label="说明">
          <el-input :model-value="currentConfig?.remark" disabled type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getConfigList, updateConfig } from '@/api/config'

// ==================== 配置列表 ====================
const loading = ref(false)
const configList = ref([])

const percentKeys = [
  'single_commission_rate',
  'multi_commission_rate',
  'special_commission_rate',
  'peer_single_rate',
  'peer_multi_rate',
  'peer_special_rate',
  'fund_pool_rate',
  'manager_team_rate',
  'store_manager_team_rate',
  'referral_reward_rate',
]

const isPercentConfig = (key) => percentKeys.includes(key)

const fetchConfigList = async () => {
  loading.value = true
  try {
    const res = await getConfigList()
    configList.value = res.data || []
  } catch (error) {
    console.error('获取配置列表失败:', error)
  } finally {
    loading.value = false
  }
}

const formatConfigValue = (row) => {
  if (isPercentConfig(row.config_key)) {
    return `${row.config_value}%`
  }
  return row.config_value
}

const getGroupLabel = (group) => {
  const map = {
    commission: '提成配置',
    fund_pool: '基金池配置',
    system: '系统配置',
  }
  return map[group] || group || '其他'
}

const getGroupTag = (group) => {
  const map = {
    commission: '',
    fund_pool: 'success',
    system: 'warning',
  }
  return map[group] || 'info'
}

// ==================== 编辑配置 ====================
const editDialogVisible = ref(false)
const submitLoading = ref(false)
const editFormRef = ref(null)
const currentConfig = ref(null)

const editForm = ref({
  value: '',
})

const editFormRules = {
  value: [{ required: true, message: '请输入配置值', trigger: 'blur' }],
}

const handleEdit = (row) => {
  currentConfig.value = row
  editForm.value = {
    value: isPercentConfig(row.config_key) ? Number(row.config_value) : row.config_value,
  }
  editDialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await editFormRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    await updateConfig(currentConfig.value.config_key, {
      config_value: editForm.value.value,
    })
    ElMessage.success('配置更新成功')
    editDialogVisible.value = false
    fetchConfigList()
  } catch (error) {
    console.error('更新配置失败:', error)
  } finally {
    submitLoading.value = false
  }
}

// ==================== 初始化 ====================
onMounted(() => {
  fetchConfigList()
})
</script>

<style lang="scss" scoped>
.commission-rule {
  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;

    .title {
      font-size: 16px;
      font-weight: 600;
      color: #303133;
    }
  }

  .value-text {
    font-size: 16px;
    font-weight: 600;
    color: #409eff;
  }
}
</style>
