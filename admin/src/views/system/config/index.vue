<template>
  <div class="system-config">
    <!-- 配置项分组展示 -->
    <el-card
      v-for="group in groupedConfigs"
      :key="group.key"
      class="config-group-card"
      shadow="never"
    >
      <template #header>
        <div class="card-header">
          <span class="title">{{ group.label }}</span>
          <el-tag :type="group.tagType" size="small">{{ group.configs.length }} 项</el-tag>
        </div>
      </template>

      <el-table
        :data="group.configs"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column prop="config_key" label="配置键" width="240" show-overflow-tooltip />
        <el-table-column label="当前值" width="160" align="center">
          <template #default="{ row }">
            <span class="value-text">{{ formatConfigValue(row) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="说明" min-width="240" show-overflow-tooltip />
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
        <el-form-item label="配置键">
          <el-input :model-value="currentConfig?.config_key" disabled />
        </el-form-item>
        <el-form-item label="配置值" prop="value">
          <el-input-number
            v-if="isRateConfig(currentConfig?.config_key)"
            v-model="editForm.value"
            :min="0"
            :max="10"
            :precision="4"
            :step="0.1"
            controls-position="right"
            style="width: 100%"
          />
          <el-input-number
            v-else-if="isPercentConfig(currentConfig?.config_key)"
            v-model="editForm.value"
            :min="0"
            :max="100"
            :precision="2"
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
        <el-form-item v-if="isRateConfig(currentConfig?.config_key)" label="">
          <span style="color: #909399; font-size: 12px">单位：倍率（如 1.2 表示乘以1.2倍）</span>
        </el-form-item>
        <el-form-item v-else-if="isPercentConfig(currentConfig?.config_key)" label="">
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
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getConfigList, updateConfig } from '@/api/config'

// ==================== 配置列表 ====================
const loading = ref(false)
const configList = ref([])

const percentKeys = [
  // 用户等级提成比例
  'commission_rate_level1_single',
  'commission_rate_level1_multi',
  'commission_rate_level2_single',
  'commission_rate_level2_multi',
  'commission_rate_level3_single',
  'commission_rate_level3_multi',
  // 同行提成比例
  'commission_rate_peer_single',
  'commission_rate_peer_multi',
  'commission_rate_peer_special',
  // 团队分润
  'fund_pool_extract_rate',
  'team_share_rate_manager',
  'team_share_rate_store',
  'referral_reward_rate',
  // 固定提成
  'fixed_commission_rate',
  // 试用期提成
  'probation_commission_rate',
]

const isPercentConfig = (key) => percentKeys.includes(key)

const isRateConfig = (key) => ['cost_rate', 'min_discount_rate'].includes(key)

// 格式化显示：如果是小数(<1)则乘以100显示为百分比
const formatConfigValue = (row) => {
  if (isRateConfig(row.config_key)) {
    return `${row.config_value} 倍`
  }
  if (isPercentConfig(row.config_key)) {
    const val = Number(row.config_value)
    if (val < 1) {
      return `${(val * 100).toFixed(1)}%`
    }
    return `${val}%`
  }
  return row.config_value
}

// 编辑时：将小数转为百分比数值显示
const toEditValue = (value) => {
  const val = Number(value)
  if (val < 1) {
    return Math.round(val * 10000) / 100 // 保留2位小数
  }
  return val
}

// 保存时：将百分比数值转为小数存储
const toSaveValue = (value) => {
  const val = Number(value)
  if (val > 1) {
    return (val / 100).toFixed(4) // 转为小数，保留4位精度
  }
  return val
}

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

// ==================== 分组 ====================
const groupMap = {
  commission: { label: '提成配置', tagType: '' },
  product: { label: '商品配置', tagType: 'success' },
  system: { label: '系统配置', tagType: 'warning' },
}

const groupedConfigs = computed(() => {
  const groups = {}
  configList.value.forEach((item) => {
    const groupKey = item.config_type || 'system'
    if (!groups[groupKey]) {
      groups[groupKey] = {
        key: groupKey,
        label: groupMap[groupKey]?.label || groupKey,
        tagType: groupMap[groupKey]?.tagType || 'info',
        configs: [],
      }
    }
    groups[groupKey].configs.push(item)
  })
  return Object.values(groups)
})

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
  let value = row.config_value
  if (isRateConfig(row.config_key)) {
    value = Number(row.config_value)
  } else if (isPercentConfig(row.config_key)) {
    // 百分比配置：编辑时将小数转为百分比数值显示
    value = toEditValue(row.config_value)
  }
  editForm.value = { value }
  editDialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await editFormRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    let saveValue = editForm.value.value
    if (isRateConfig(currentConfig.value.config_key)) {
      saveValue = String(editForm.value.value)
    } else if (isPercentConfig(currentConfig.value.config_key)) {
      // 百分比配置：保存时将百分比数值转为小数存储
      saveValue = toSaveValue(editForm.value.value)
    }
    await updateConfig(currentConfig.value.config_key, {
      config_value: String(saveValue),
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
.system-config {
  .config-group-card {
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
  }

  .value-text {
    font-size: 15px;
    font-weight: 600;
    color: #409eff;
  }
}
</style>
