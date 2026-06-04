<template>
  <div class="settlement-manage">
    <!-- 月度结算 -->
    <el-card class="settle-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">月度结算</span>
          <el-tag type="info" size="small">自动结算：每月1日凌晨2点</el-tag>
        </div>
      </template>

      <el-form :inline="true" :model="settleForm" class="settle-form">
        <el-form-item label="结算月份">
          <el-date-picker
            v-model="settleForm.period_value"
            type="month"
            placeholder="选择月份"
            format="YYYY-MM"
            value-format="YYYY-MM"
          />
        </el-form-item>
        <el-form-item label="结算类型">
          <el-select v-model="settleForm.settle_type" style="width: 160px">
            <el-option label="全部" value="all" />
            <el-option label="基金池结算" value="fund" />
            <el-option label="固定提成计算" value="fixed" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="settleLoading" @click="handleSettle">
            执行结算
          </el-button>
        </el-form-item>
      </el-form>

      <!-- 结算结果 -->
      <el-alert
        v-if="settleResult"
        :title="`结算完成：${settleResult.period_value}`"
        type="success"
        :closable="false"
        show-icon
        style="margin-top: 12px"
      >
        <template #default>
          <div v-if="settleResult.fund_pool_result" style="margin-top: 4px">
            <strong>基金池：</strong>{{ settleResult.fund_pool_result }}
          </div>
          <div v-if="settleResult.fixed_commission_result" style="margin-top: 4px">
            <strong>固定提成：</strong>{{ settleResult.fixed_commission_result }}
          </div>
        </template>
      </el-alert>
    </el-card>

    <!-- 订单提成重新计算 -->
    <el-card class="settle-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">订单提成重新计算</span>
          <span class="subtitle">删除订单已有提成记录后重新计算</span>
        </div>
      </template>

      <el-form :inline="true" :model="recalcForm" class="settle-form">
        <el-form-item label="订单ID">
          <el-input
            v-model="recalcForm.order_id"
            placeholder="请输入订单ID"
            style="width: 200px"
            clearable
          />
        </el-form-item>
        <el-form-item>
          <el-button type="warning" :loading="recalcLoading" @click="handleRecalculate">
            重新计算
          </el-button>
        </el-form-item>
      </el-form>

      <el-alert
        v-if="recalcResult"
        title="重新计算完成"
        type="success"
        :closable="false"
        show-icon
        style="margin-top: 12px"
      />
    </el-card>

    <!-- 结算说明 -->
    <el-card class="settle-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">结算说明</span>
        </div>
      </template>
      <el-descriptions :column="1" border size="small">
        <el-descriptions-item label="订单提成">
          订单回款完成后自动实时计算（commission_type=1~6）
        </el-descriptions-item>
        <el-descriptions-item label="基金池结算">
          汇总各门店月度基金池提成，按员工利润比例分配份额
        </el-descriptions-item>
        <el-descriptions-item label="固定提成">
          按业务员当月回款总额的固定比例计算（默认5%）
        </el-descriptions-item>
        <el-descriptions-item label="月度工资">
          结算完成后，需在"工资管理"中手动生成月度工资
        </el-descriptions-item>
        <el-descriptions-item label="重复结算">
          支持重复结算同一月份，已有数据会被更新而非重复创建
        </el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { settleMonthly, recalculateOrder } from '@/api/commission'

// ==================== 月度结算 ====================
const settleLoading = ref(false)
const settleResult = ref(null)

const settleForm = reactive({
  period_value: '',
  settle_type: 'all',
})

const handleSettle = () => {
  if (!settleForm.period_value) {
    ElMessage.warning('请选择结算月份')
    return
  }

  const typeLabels = { all: '全部（基金池+固定提成）', fund: '基金池结算', fixed: '固定提成计算' }
  const label = typeLabels[settleForm.settle_type] || '全部'

  ElMessageBox.confirm(
    `确定要对 ${settleForm.period_value} 执行${label}吗？`,
    '确认结算',
    {
      confirmButtonText: '确定执行',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    settleLoading.value = true
    settleResult.value = null
    try {
      const res = await settleMonthly({
        period_value: settleForm.period_value,
        settle_type: settleForm.settle_type,
      })
      settleResult.value = res.data
      ElMessage.success('结算完成')
    } catch (error) {
      console.error('结算失败:', error)
    } finally {
      settleLoading.value = false
    }
  }).catch(() => {})
}

// ==================== 订单提成重新计算 ====================
const recalcLoading = ref(false)
const recalcResult = ref(false)

const recalcForm = reactive({
  order_id: '',
})

const handleRecalculate = () => {
  const orderId = Number(recalcForm.order_id)
  if (!orderId || orderId <= 0) {
    ElMessage.warning('请输入有效的订单ID')
    return
  }

  ElMessageBox.confirm(
    `确定要重新计算订单 ${orderId} 的提成吗？已有提成记录将被删除后重新计算。`,
    '确认重新计算',
    {
      confirmButtonText: '确定执行',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    recalcLoading.value = true
    recalcResult.value = false
    try {
      await recalculateOrder(orderId)
      recalcResult.value = true
      ElMessage.success('重新计算完成')
    } catch (error) {
      console.error('重新计算失败:', error)
    } finally {
      recalcLoading.value = false
    }
  }).catch(() => {})
}
</script>

<style lang="scss" scoped>
.settlement-manage {
  .settle-card {
    margin-bottom: 16px;

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

    .settle-form {
      margin-bottom: 0;
    }
  }
}
</style>
