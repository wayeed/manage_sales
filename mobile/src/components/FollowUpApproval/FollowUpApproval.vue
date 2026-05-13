<template>
  <view class="fa" v-if="visible">
    <view class="fa-mask"></view>
    <view class="fa-sheet">
      <!-- 顶部拖拽指示条 -->
      <view class="fa-handle-bar">
        <view class="fa-handle"></view>
      </view>

      <!-- 标题栏 -->
      <view class="fa-header">
        <text class="fa-title">申请跟进</text>
      </view>

      <!-- 内容区 -->
      <view class="fa-body">
        <!-- 客户信息 -->
        <view class="fa-customer">
          <text class="fa-customer-name">{{ customer?.customer_name || customer?.name }}</text>
          <text class="fa-customer-phone">{{ customer?.phone }}</text>
        </view>

        <!-- 审批状态 -->
        <view v-if="status === 'pending'" class="fa-status fa-status--pending">
          <view class="fa-spinner"></view>
          <text class="fa-status-text">等待审批中...</text>
          <text class="fa-status-hint">审批人：{{ approverName || '系统分配中' }}</text>
        </view>

        <view v-else-if="status === 'approved'" class="fa-status fa-status--approved">
          <text class="fa-status-icon">✓</text>
          <text class="fa-status-text">审批通过</text>
          <text class="fa-status-hint">客户已转移至您名下</text>
        </view>

        <view v-else-if="status === 'rejected'" class="fa-status fa-status--rejected">
          <text class="fa-status-icon">✕</text>
          <text class="fa-status-text">审批被拒绝</text>
          <text class="fa-status-hint">{{ rejectReason || '请联系审批人了解详情' }}</text>
        </view>
      </view>

      <!-- 底部操作栏 -->
      <view class="fa-footer">
        <button v-if="status === 'pending'" class="fa-btn fa-btn--cancel" @tap="handleCancel">
          取消申请
        </button>
        <button v-else class="fa-btn fa-btn--confirm" @tap="handleConfirm">
          {{ status === 'approved' ? '选择客户' : '关闭' }}
        </button>
      </view>
    </view>
  </view>
</template>

<script>
import { ref, watch, onUnmounted } from 'vue'
import { applyFollowUp, getApprovalStatus, cancelFollowUp } from '../../api/approval'
import { getCustomerDetail } from '../../api/customer'

export default {
  name: 'FollowUpApproval',
  props: {
    visible: {
      type: Boolean,
      default: false
    },
    customer: {
      type: Object,
      default: null
    }
  },
  emits: ['close', 'approved', 'rejected'],
  setup(props, { emit }) {
    const status = ref('pending') // pending, approved, rejected
    const approvalId = ref(null)
    const approverName = ref('')
    const rejectReason = ref('')
    let pollingTimer = null

    watch(() => props.visible, async (val) => {
      if (val && props.customer) {
        status.value = 'pending'
        rejectReason.value = ''
        await createApproval()
        startPolling()
      } else {
        stopPolling()
      }
    })

    const createApproval = async () => {
      try {
        const res = await applyFollowUp({
          customer_id: props.customer.id,
          remark: '申请跟进该客户'
        })
        approvalId.value = res.data?.id
        if (res.data?.approver) {
          approverName.value = res.data.approver.real_name || res.data.approver.username
        }
      } catch (e) {
        console.error('创建申请失败:', e)
        uni.showToast({ title: e.message || '申请失败', icon: 'none' })
        emit('close')
      }
    }

    const startPolling = () => {
      pollingTimer = setInterval(async () => {
        if (!approvalId.value) return
        try {
          const res = await getApprovalStatus(approvalId.value)
          const approval = res.data
          if (approval.status === 1) {
            // 审批通过
            status.value = 'approved'
            stopPolling()
            uni.showToast({ title: '审批通过', icon: 'success' })
          } else if (approval.status === 2) {
            // 审批拒绝
            status.value = 'rejected'
            rejectReason.value = approval.reject_reason
            stopPolling()
            emit('rejected', { customer: props.customer, reason: approval.reject_reason })
          }
        } catch (e) {
          console.error('轮询审批状态失败:', e)
        }
      }, 5000) // 每5秒轮询一次
    }

    const stopPolling = () => {
      if (pollingTimer) {
        clearInterval(pollingTimer)
        pollingTimer = null
      }
    }

    const close = () => {
      stopPolling()
      emit('close')
    }

    const handleCancel = async () => {
      if (!approvalId.value) {
        close()
        return
      }
      try {
        await cancelFollowUp(approvalId.value)
        uni.showToast({ title: '已撤回申请', icon: 'success' })
      } catch (e) {
        console.error('撤回申请失败:', e)
      }
      close()
    }

    const handleConfirm = async () => {
      if (status.value === 'approved') {
        // 审批通过后重新获取客户数据，确保业务员信息是最新的
        try {
          const res = await getCustomerDetail(props.customer.id)
          const updatedCustomer = res.data
          // 更新业务员信息
          emit('approved', {
            ...props.customer,
            ...updatedCustomer,
            created_by: updatedCustomer.created_by,
            salesman_name: updatedCustomer.salesman?.real_name || updatedCustomer.salesman?.username || ''
          })
        } catch (e) {
          console.error('获取客户详情失败:', e)
          // 降级使用原数据
          emit('approved', props.customer)
        }
      }
      close()
    }

    onUnmounted(() => {
      stopPolling()
    })

    return {
      status,
      approverName,
      rejectReason,
      close,
      handleCancel,
      handleConfirm
    }
  }
}
</script>

<style lang="scss" scoped>
/* 弹窗容器 */
.fa {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 1001;
}

.fa-mask {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
}

.fa-sheet {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  background: #fff;
  border-radius: 28rpx 28rpx 0 0;
  padding-bottom: env(safe-area-inset-bottom);
}

/* 拖拽指示条 */
.fa-handle-bar {
  display: flex;
  justify-content: center;
  padding: 18rpx 0 0;
}

.fa-handle {
  width: 64rpx;
  height: 8rpx;
  border-radius: 4rpx;
  background: #ddd;
}

/* 标题栏 */
.fa-header {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24rpx 32rpx;
  border-bottom: 1rpx solid #f0f0f0;
}

.fa-title {
  font-size: 34rpx;
  font-weight: 600;
  color: #1a1a1a;
}

/* 内容区 */
.fa-body {
  padding: 32rpx;
}

/* 客户信息 */
.fa-customer {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8rpx;
  padding: 24rpx;
  background: #f7f7f7;
  border-radius: 16rpx;
  margin-bottom: 32rpx;
}

.fa-customer-name {
  font-size: 32rpx;
  font-weight: 500;
  color: #1a1a1a;
}

.fa-customer-phone {
  font-size: 26rpx;
  color: #666;
}

/* 审批状态 */
.fa-status {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16rpx;
  padding: 48rpx 0;
}

.fa-status-icon {
  width: 80rpx;
  height: 80rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 40rpx;
  color: #fff;

  .fa-status--approved & {
    background: #52c41a;
  }

  .fa-status--rejected & {
    background: #ff4d4f;
  }
}

.fa-status-text {
  font-size: 32rpx;
  font-weight: 500;
  color: #1a1a1a;
}

.fa-status-hint {
  font-size: 26rpx;
  color: #999;
}

/* 加载动画 */
.fa-spinner {
  width: 64rpx;
  height: 64rpx;
  border: 6rpx solid #e0e0e0;
  border-top-color: #1890ff;
  border-radius: 50%;
  animation: fa-spin 0.6s linear infinite;
}

@keyframes fa-spin {
  to { transform: rotate(360deg); }
}

/* 底部操作栏 */
.fa-footer {
  padding: 16rpx 32rpx 32rpx;
}

.fa-btn {
  width: 100%;
  height: 88rpx;
  border-radius: 16rpx;
  font-size: 30rpx;
  font-weight: 500;

  &--cancel {
    background: #f5f5f5;
    color: #666;
  }

  &--confirm {
    background: #1890ff;
    color: #fff;
  }
}
</style>
