<template>
  <view class="ce" v-if="visible">
    <view class="ce-mask" @tap="close"></view>
    <view class="ce-sheet">
      <!-- 顶部拖拽指示条 -->
      <view class="ce-handle-bar">
        <view class="ce-handle"></view>
      </view>

      <!-- 标题栏 -->
      <view class="ce-header">
        <text class="ce-title">预估提成</text>
        <view class="ce-close" @tap="close">
          <text class="ce-close-icon">✕</text>
        </view>
      </view>

      <!-- 内容区 -->
      <view class="ce-body">
        <view class="ce-row">
          <text class="ce-label">预估利润</text>
          <text class="ce-value">¥{{ formatAmount(estimate.estimated_profit) }}</text>
        </view>
        <view class="ce-row">
          <text class="ce-label">提成比例</text>
          <text class="ce-value">{{ formatRate(estimate.commission_rate) }}%</text>
        </view>
        <view class="ce-row ce-row--highlight">
          <text class="ce-label">预估提成</text>
          <text class="ce-value ce-value--primary">¥{{ formatAmount(estimate.estimated_commission) }}</text>
        </view>

        <!-- 品类提示 -->
        <view v-if="estimate.show_category_hint" class="ce-hint">
          <text class="ce-hint-text">{{ estimate.category_hint }}</text>
        </view>
      </view>

      <!-- 底部操作栏 -->
      <view class="ce-footer">
        <button class="ce-btn" @tap="close">确定</button>
      </view>
    </view>
  </view>
</template>

<script>
import { ref, watch } from 'vue'
import { estimateCommission } from '../../api/commission'

export default {
  name: 'CommissionEstimate',
  props: {
    visible: {
      type: Boolean,
      default: false
    },
    items: {
      type: Array,
      default: () => []
    },
    isPeerOrder: {
      type: Boolean,
      default: false
    }
  },
  emits: ['close'],
  setup(props, { emit }) {
    const estimate = ref({
      estimated_profit: 0,
      commission_rate: 0,
      estimated_commission: 0,
      category_count: 0,
      show_category_hint: false,
      category_hint: ''
    })

    watch(() => props.visible, async (val) => {
      if (val && props.items.length > 0) {
        await fetchEstimate()
      }
    })

    const fetchEstimate = async () => {
      try {
        const res = await estimateCommission({
          items: props.items.map(item => ({
            list_price: parseFloat(item.listPrice) || 0,
            sale_price: parseFloat(item.salePrice) || 0,
            cost_price: parseFloat(item.costPrice) || 0,
            quantity: parseInt(item.quantity) || 1
          })),
          is_peer_order: props.isPeerOrder ? 1 : 0
        })
        estimate.value = res.data || {}
      } catch (e) {
        console.error('预估提成失败:', e)
      }
    }

    const formatAmount = (val) => {
      const num = parseFloat(val) || 0
      return num.toFixed(2)
    }

    const formatRate = (val) => {
      const num = parseFloat(val) || 0
      return (num * 100).toFixed(1)
    }

    const close = () => {
      emit('close')
    }

    return {
      estimate,
      formatAmount,
      formatRate,
      close
    }
  }
}
</script>

<style lang="scss" scoped>
/* 弹窗容器 */
.ce {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 1000;
}

.ce-mask {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
}

.ce-sheet {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  background: #fff;
  border-radius: 28rpx 28rpx 0 0;
  padding-bottom: env(safe-area-inset-bottom);
}

/* 拖拽指示条 */
.ce-handle-bar {
  display: flex;
  justify-content: center;
  padding: 18rpx 0 0;
}

.ce-handle {
  width: 64rpx;
  height: 8rpx;
  border-radius: 4rpx;
  background: #ddd;
}

/* 标题栏 */
.ce-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20rpx 32rpx 16rpx;
}

.ce-title {
  font-size: 34rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.ce-close {
  width: 64rpx;
  height: 64rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: #f5f5f5;
}

.ce-close-icon {
  font-size: 28rpx;
  color: #888;
}

/* 内容区 */
.ce-body {
  padding: 16rpx 32rpx 32rpx;
}

.ce-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24rpx 0;
  border-bottom: 1rpx solid #f0f0f0;

  &--highlight {
    background: #f0f7ff;
    margin: 16rpx -24rpx;
    padding: 24rpx;
    border-radius: 12rpx;
    border-bottom: none;
  }
}

.ce-label {
  font-size: 28rpx;
  color: #666;
}

.ce-value {
  font-size: 32rpx;
  font-weight: 500;
  color: #1a1a1a;

  &--primary {
    color: #1890ff;
  }
}

/* 品类提示 */
.ce-hint {
  margin-top: 24rpx;
  padding: 20rpx 24rpx;
  background: #fff2f0;
  border-radius: 12rpx;
  border-left: 6rpx solid #ff4d4f;
}

.ce-hint-text {
  font-size: 26rpx;
  color: #ff4d4f;
  font-weight: 600;
}

/* 底部操作栏 */
.ce-footer {
  padding: 16rpx 32rpx 32rpx;
}

.ce-btn {
  width: 100%;
  height: 88rpx;
  border-radius: 16rpx;
  font-size: 30rpx;
  font-weight: 500;
  background: #1890ff;
  color: #fff;
}
</style>
