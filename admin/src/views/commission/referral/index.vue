<template>
  <div class="referral-manage">
    <!-- 引荐关系列表 -->
    <el-card class="table-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">老带新管理</span>
          <span class="subtitle">系统奖励比例：{{ systemRewardRate }}</span>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="referralList"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column label="引荐人" width="140">
          <template #default="{ row }">
            {{ row.referrer?.real_name || row.referrer?.username || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="被引荐人" width="140">
          <template #default="{ row }">
            {{ row.referred?.real_name || row.referred?.username || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getReferralStatusTag(row.status)" size="small">
              {{ getReferralStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="奖励比例" width="120" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.status === 1" type="warning" size="small">{{ systemRewardRate }}</el-tag>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ row.created_at ? row.created_at.substring(0, 19).replace('T', ' ') : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="终止时间" width="180">
          <template #default="{ row }">
            {{ row.ended_at ? row.ended_at.substring(0, 19).replace('T', ' ') : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="终止原因" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.ended_reason || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" align="center" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 1"
              type="danger" link size="small" @click="handleTerminate(row)"
            >
              终止关系
            </el-button>
            <span v-else class="text-muted">-</span>
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
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { get, post } from '@/api/request'
import { getConfigList } from '@/api/config'

// ==================== 系统配置 ====================
const systemConfig = ref({})
const systemRewardRate = computed(() => {
  const rate = systemConfig.value['referral_reward_rate']
  if (rate) {
    const val = Number(rate)
    return val < 1 ? `${(val * 100).toFixed(1)}%` : `${val}%`
  }
  return '-'
})

const fetchSystemConfig = async () => {
  try {
    const res = await getConfigList()
    const configs = res.data || []
    const configMap = {}
    configs.forEach(c => { configMap[c.config_key] = c.config_value })
    systemConfig.value = configMap
  } catch (e) {
    console.error('获取系统配置失败:', e)
  }
}

// ==================== 列表 ====================
const loading = ref(false)
const referralList = ref([])

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
    }
    const res = await get('/referrals', params)
    referralList.value = res.data?.list || res.data || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取引荐关系列表失败:', error)
  } finally {
    loading.value = false
  }
}

// ==================== 状态映射 ====================
const getReferralStatusLabel = (status) => {
  const map = { 1: '生效中', 0: '已终止' }
  return map[status] ?? '未知'
}

const getReferralStatusTag = (status) => {
  return status === 1 ? 'success' : 'info'
}

// ==================== 终止关系 ====================
const handleTerminate = (row) => {
  const referrerName = row.referrer?.real_name || row.referrer?.username || '-'
  const referredName = row.referred?.real_name || row.referred?.username || '-'
  ElMessageBox.confirm(
    `确定要终止 "${referrerName}" 与 "${referredName}" 的引荐关系吗？`,
    '终止确认',
    {
      confirmButtonText: '确定终止',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await post(`/referrals/${row.id}/terminate`)
      ElMessage.success('终止成功')
      fetchList()
    } catch (error) {
      console.error('终止引荐关系失败:', error)
    }
  }).catch(() => {})
}

// ==================== 初始化 ====================
onMounted(() => {
  fetchSystemConfig()
  fetchList()
})
</script>

<style lang="scss" scoped>
.referral-manage {
  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;

    .title {
      font-size: 16px;
      font-weight: 600;
      color: #303133;
    }

    .subtitle {
      font-size: 13px;
      color: #909399;
    }
  }

  .table-card {
    .pagination-wrapper {
      display: flex;
      justify-content: flex-end;
      margin-top: 16px;
    }
  }

  .text-muted {
    color: #c0c4cc;
  }
}
</style>
