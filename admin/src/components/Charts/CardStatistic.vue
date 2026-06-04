<template>
  <el-card shadow="hover" class="card-statistic">
    <div class="card-content">
      <div class="card-icon" v-if="icon">
        <el-icon :size="32"><component :is="icon" /></el-icon>
      </div>
      <div class="card-info">
        <div class="card-title">{{ title }}</div>
        <div class="card-value">
          <span class="value-text">{{ displayValue }}</span>
          <span v-if="suffix" class="value-suffix">{{ suffix }}</span>
        </div>
      </div>
      <div v-if="trend" class="card-trend" :class="trendClass">
        <el-icon v-if="trend === 'up'" :size="16"><Top /></el-icon>
        <el-icon v-else :size="16"><Bottom /></el-icon>
        <span class="trend-value">{{ trendValue }}</span>
      </div>
    </div>
  </el-card>
</template>

<script setup>
import { computed } from 'vue'
import { Top, Bottom } from '@element-plus/icons-vue'

const props = defineProps({
  title: { type: String, default: '' },
  value: { type: [Number, String], default: '' },
  suffix: { type: String, default: '' },
  icon: { type: String, default: '' },
  decimals: { type: Number, default: 2 },
  trend: { type: String, default: '', validator: (v) => ['', 'up', 'down'].includes(v) },
  trendValue: { type: String, default: '' },
})

const displayValue = computed(() => {
  const num = Number(props.value)
  if (isNaN(num)) return props.value
  return num.toLocaleString('zh-CN', {
    minimumFractionDigits: 0,
    maximumFractionDigits: props.decimals,
  })
})

const trendClass = computed(() => {
  if (props.trend === 'up') return 'trend-up'
  if (props.trend === 'down') return 'trend-down'
  return ''
})
</script>

<style lang="scss" scoped>
.card-statistic {
  height: 100%;
  border-radius: 8px;

  :deep(.el-card__body) {
    padding: 20px;
  }

  .card-content {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .card-icon {
    flex-shrink: 0;
    width: 56px;
    height: 56px;
    border-radius: 12px;
    background: linear-gradient(135deg, #1890ff 0%, #36cfc9 100%);
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
  }

  .card-info {
    flex: 1;
    min-width: 0;
  }

  .card-title {
    font-size: 14px;
    color: #909399;
    margin-bottom: 8px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .card-value {
    display: flex;
    align-items: baseline;
    gap: 4px;
  }

  .value-text {
    font-size: 28px;
    font-weight: 700;
    color: #303133;
    line-height: 1.2;
  }

  .value-suffix {
    font-size: 14px;
    color: #909399;
  }

  .card-trend {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 4px 10px;
    border-radius: 4px;
    font-size: 13px;
    font-weight: 500;
    flex-shrink: 0;

    &.trend-up {
      color: #52c41a;
      background-color: #f6ffed;
    }

    &.trend-down {
      color: #f5222d;
      background-color: #fff1f0;
    }

    .trend-value {
      font-size: 13px;
    }
  }
}
</style>
