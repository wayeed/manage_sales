<template>
  <view class="detail-page">
    <view v-if="order" class="detail-content">
      <!-- 订单状态卡片 -->
      <view class="status-banner" :class="'status-bg--' + getStatusClassName(order.order_status)">
        <view class="status-tag-large">
          <text>{{ getStatusText(order.order_status) }}</text>
        </view>
        <text class="status-order-no">订单号: {{ order.order_no || '--' }}</text>
      </view>

      <!-- 基本信息 -->
      <view class="info-section card">
        <text class="section-title">基本信息</text>
        <view class="info-row">
          <text class="info-label">订单号</text>
          <text class="info-value">{{ order.order_no || '--' }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">创建时间</text>
          <text class="info-value">{{ order.created_at || '--' }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">业务员</text>
          <text class="info-value">{{ order.salesman?.real_name || '--' }}</text>
        </view>
      </view>

      <!-- 客户信息 -->
      <view class="info-section card">
        <text class="section-title">客户信息</text>
        <view class="info-row">
          <text class="info-label">客户姓名</text>
          <text class="info-value">{{ order.customer_name || '--' }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">联系电话</text>
          <text class="info-value">{{ order.customer_phone || '--' }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">收货地址</text>
          <text class="info-value">{{ order.customer_address || '--' }}</text>
        </view>
      </view>

      <!-- 金额信息 -->
      <view class="info-section card">
        <text class="section-title">金额信息</text>
        <view class="info-row">
          <text class="info-label">挂牌总价</text>
          <text class="info-value">{{ order.total_list_price || '0.00' }}元</text>
        </view>
        <view class="info-row">
          <text class="info-label">销售总价</text>
          <text class="info-value">{{ order.total_sale_price || '0.00' }}元</text>
        </view>
        <view class="info-row">
          <text class="info-label">折扣金额</text>
          <text class="info-value info-discount">{{ order.discount_amount || '0.00' }}元</text>
        </view>
        <view class="info-row info-row--highlight">
          <text class="info-label info-label--bold">成交价</text>
          <text class="info-value info-value--red">{{ order.final_amount || '0.00' }}元</text>
        </view>
        <view class="info-row">
          <text class="info-label">成本</text>
          <text class="info-value">{{ order.total_cost || '0.00' }}元</text>
        </view>
        <view class="info-row">
          <text class="info-label">利润</text>
          <text class="info-value info-profit">{{ order.actual_profit || '0.00' }}元</text>
        </view>
      </view>

      <!-- 商品明细 -->
      <view class="info-section card">
        <text class="section-title">商品明细</text>
        <view v-if="items && items.length > 0">
          <view v-for="(item, index) in items" :key="index" class="product-detail">
            <view class="product-detail-header">
              <text class="product-detail-name">{{ item.product_name || '--' }}</text>
              <text class="product-detail-sku">SKU: {{ item.sku_name || '--' }}</text>
            </view>
            <view class="product-detail-row">
              <text class="product-detail-label">数量</text>
              <text class="product-detail-value">{{ item.quantity || 0 }}</text>
            </view>
            <view class="product-detail-row">
              <text class="product-detail-label">单价</text>
              <text class="product-detail-value">{{ item.sale_price || '0.00' }}元</text>
            </view>
            <view class="product-detail-row">
              <text class="product-detail-label">成本</text>
              <text class="product-detail-value">{{ item.unit_cost || '0.00' }}元</text>
            </view>
          </view>
        </view>
        <view v-else class="empty-hint">
          <text class="empty-hint-text">暂无商品明细</text>
        </view>
      </view>

      <!-- 礼品信息 -->
      <view v-if="gifts && gifts.length > 0" class="info-section card">
        <text class="section-title">礼品信息</text>
        <view v-for="(gift, index) in gifts" :key="index" class="info-row">
          <text class="info-label">{{ gift.name || '礼品' + (index + 1) }}</text>
          <text class="info-value">x{{ gift.quantity || 0 }}</text>
        </view>
      </view>

      <!-- 回款记录 -->
      <view class="info-section card">
        <text class="section-title">回款记录</text>
        <view v-if="payments && payments.length > 0">
          <view v-for="(payment, index) in payments" :key="index" class="payment-item">
            <view class="payment-left">
              <text class="payment-amount">+{{ payment.amount || '0.00' }}元</text>
              <text class="payment-time">{{ payment.created_at || '' }}</text>
            </view>
            <view class="tag" :class="payment.status === 1 ? 'tag-success' : 'tag-warning'">
              <text>{{ payment.status === 1 ? '已确认' : '待确认' }}</text>
            </view>
          </view>
        </view>
        <view v-else class="empty-hint">
          <text class="empty-hint-text">暂无回款记录</text>
        </view>
      </view>

      <!-- 底部占位（为操作按钮留空间） -->
      <view v-if="showActions" style="height: 140rpx;"></view>
    </view>

    <!-- 加载中 -->
    <view v-else class="loading-state">
      <text class="loading-text">加载中...</text>
    </view>

    <!-- 底部操作按钮 -->
    <view v-if="showActions && order" class="bottom-actions safe-bottom">
      <!-- 待审批 -->
      <template v-if="order.order_status === 0">
        <button class="action-btn action-btn--reject" @tap="showRejectDialog">驳回</button>
        <button class="action-btn action-btn--approve" @tap="handleApprove">审核通过</button>
      </template>
      <!-- 已生效 -->
      <template v-if="order.order_status === 1">
        <button class="action-btn action-btn--payment" @tap="showPaymentDialog">录入回款</button>
      </template>
    </view>

    <!-- 审核驳回弹窗 -->
    <view v-if="rejectDialogVisible" class="modal-mask" @tap="rejectDialogVisible = false">
      <view class="modal-content" @tap.stop>
        <text class="modal-title">驳回原因</text>
        <textarea
          v-model="rejectReason"
          class="modal-textarea"
          placeholder="请输入驳回原因"
          maxlength="200"
        />
        <view class="modal-actions">
          <button class="modal-btn modal-btn--cancel" @tap="rejectDialogVisible = false">取消</button>
          <button class="modal-btn modal-btn--confirm" @tap="handleReject">确认驳回</button>
        </view>
      </view>
    </view>

    <!-- 回款录入弹窗 -->
    <view v-if="paymentDialogVisible" class="modal-mask" @tap="paymentDialogVisible = false">
      <view class="modal-content" @tap.stop>
        <text class="modal-title">录入回款</text>
        <view class="modal-form-item">
          <text class="modal-form-label">回款金额(元)</text>
          <input v-model="paymentAmount" class="input-field" type="digit" placeholder="请输入回款金额" />
        </view>
        <view class="modal-form-item">
          <text class="modal-form-label">备注</text>
          <input v-model="paymentRemark" class="input-field" placeholder="请输入备注(选填)" />
        </view>
        <view class="modal-actions">
          <button class="modal-btn modal-btn--cancel" @tap="paymentDialogVisible = false">取消</button>
          <button class="modal-btn modal-btn--confirm" @tap="handlePayment">确认录入</button>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { getOrderDetail, approveOrder } from '../../api/order'
import { post } from '../../api/request'

export default {
  setup() {
    const order = ref(null)
    const items = ref([])
    const gifts = ref([])
    const payments = ref([])
    const rejectDialogVisible = ref(false)
    const rejectReason = ref('')
    const paymentDialogVisible = ref(false)
    const paymentAmount = ref('')
    const paymentRemark = ref('')
    const submitting = ref(false)

    const getStatusText = (status) => {
      const map = {
        0: '待审批',
        1: '已生效',
        2: '已驳回',
        3: '已取消',
        4: '已退货'
      }
      return map[status] || status || '未知'
    }

    const getStatusClassName = (status) => {
      const map = {
        0: 'pending_approval',
        1: 'effective',
        2: 'rejected',
        3: 'cancelled',
        4: 'returned'
      }
      return map[status] || ''
    }

    // 是否显示操作按钮
    const showActions = computed(() => {
      if (!order.value) return false
      return order.value.order_status === 0 || order.value.order_status === 1
    })

    // 审核通过
    const handleApprove = async () => {
      if (submitting.value) return
      uni.showModal({
        title: '确认',
        content: '确定审核通过该订单吗？',
        success: async (res) => {
          if (res.confirm) {
            submitting.value = true
            try {
              await approveOrder(order.value.id, { approved: true })
              uni.showToast({ title: '审核通过', icon: 'success' })
              loadDetail(order.value.id)
            } catch (e) {
              console.error('审核失败:', e)
            } finally {
              submitting.value = false
            }
          }
        }
      })
    }

    // 显示驳回弹窗
    const showRejectDialog = () => {
      rejectReason.value = ''
      rejectDialogVisible.value = true
    }

    // 驳回
    const handleReject = async () => {
      if (!rejectReason.value) {
        uni.showToast({ title: '请输入驳回原因', icon: 'none' })
        return
      }
      if (submitting.value) return
      submitting.value = true
      try {
        await approveOrder(order.value.id, { approved: false, reason: rejectReason.value })
        uni.showToast({ title: '已驳回', icon: 'success' })
        rejectDialogVisible.value = false
        loadDetail(order.value.id)
      } catch (e) {
        console.error('驳回失败:', e)
      } finally {
        submitting.value = false
      }
    }

    // 显示回款弹窗
    const showPaymentDialog = () => {
      paymentAmount.value = ''
      paymentRemark.value = ''
      paymentDialogVisible.value = true
    }

    // 录入回款
    const handlePayment = async () => {
      if (!paymentAmount.value || Number(paymentAmount.value) <= 0) {
        uni.showToast({ title: '请输入有效金额', icon: 'none' })
        return
      }
      if (submitting.value) return
      submitting.value = true
      try {
        await post('/payments', {
          order_id: order.value.id,
          amount: paymentAmount.value,
          remark: paymentRemark.value || undefined
        })
        uni.showToast({ title: '录入成功', icon: 'success' })
        paymentDialogVisible.value = false
        loadDetail(order.value.id)
      } catch (e) {
        console.error('录入回款失败:', e)
      } finally {
        submitting.value = false
      }
    }

    onMounted(() => {
      const pages = getCurrentPages()
      const currentPage = pages[pages.length - 1]
      const id = currentPage.$page?.options?.id || currentPage.options?.id
      if (id) {
        loadDetail(id)
      }
    })

    const loadDetail = async (id) => {
      try {
        const res = await getOrderDetail(id)
        // 后端返回结构: { order, items, gifts, payments }
        order.value = res.data?.order || null
        items.value = res.data?.items || []
        gifts.value = res.data?.gifts || []
        payments.value = res.data?.payments || []
      } catch (e) {
        console.error('加载订单详情失败:', e)
      }
    }

    return {
      order,
      items,
      gifts,
      payments,
      showActions,
      rejectDialogVisible,
      rejectReason,
      paymentDialogVisible,
      paymentAmount,
      paymentRemark,
      getStatusText,
      getStatusClassName,
      handleApprove,
      showRejectDialog,
      handleReject,
      showPaymentDialog,
      handlePayment
    }
  }
}
</script>

<style lang="scss" scoped>
.detail-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding-bottom: 0;
}

/* 状态横幅 */
.status-banner {
  padding: 40rpx 30rpx;
  display: flex;
  flex-direction: column;
  align-items: center;

  &.status-bg--pending_approval {
    background: linear-gradient(135deg, #faad14 0%, #d48806 100%);
  }

  &.status-bg--effective {
    background: linear-gradient(135deg, #52c41a 0%, #389e0d 100%);
  }

  &.status-bg--rejected {
    background: linear-gradient(135deg, #ff4d4f 0%, #cf1322 100%);
  }

  &.status-bg--cancelled {
    background: linear-gradient(135deg, #999999 0%, #666666 100%);
  }

  &.status-bg--returned {
    background: linear-gradient(135deg, #722ed1 0%, #531dab 100%);
  }
}

.status-tag-large {
  font-size: 36rpx;
  font-weight: bold;
  color: #ffffff;
  margin-bottom: 12rpx;
}

.status-order-no {
  font-size: 24rpx;
  color: rgba(255, 255, 255, 0.8);
}

/* 信息区域 */
.info-section {
  margin: 20rpx 24rpx;
  padding: 24rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333333;
  margin-bottom: 20rpx;
}

.info-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12rpx 0;
  border-bottom: 1rpx solid #f5f5f5;

  &:last-child {
    border-bottom: none;
  }
}

.info-row--highlight {
  padding: 16rpx 0;
  background-color: #fffbe6;
  margin: 8rpx -24rpx 8rpx;
  padding-left: 24rpx;
  padding-right: 24rpx;
  border-bottom: none;
}

.info-label {
  font-size: 28rpx;
  color: #999999;

  &--bold {
    color: #333333;
    font-weight: bold;
  }
}

.info-value {
  font-size: 28rpx;
  color: #333333;
  max-width: 400rpx;
  text-align: right;
}

.info-discount {
  color: #ff4d4f;
}

.info-value--red {
  font-size: 32rpx;
  font-weight: bold;
  color: #ff4d4f;
}

.info-profit {
  color: #52c41a;
  font-weight: 500;
}

/* 商品明细 */
.product-detail {
  padding: 16rpx 0;
  border-bottom: 1rpx solid #f5f5f5;

  &:last-child {
    border-bottom: none;
  }
}

.product-detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12rpx;
}

.product-detail-name {
  font-size: 28rpx;
  font-weight: 500;
  color: #333333;
}

.product-detail-sku {
  font-size: 24rpx;
  color: #999999;
}

.product-detail-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4rpx 0;
}

.product-detail-label {
  font-size: 26rpx;
  color: #999999;
}

.product-detail-value {
  font-size: 26rpx;
  color: #333333;
}

.empty-hint {
  padding: 30rpx 0;
  text-align: center;
}

.empty-hint-text {
  font-size: 26rpx;
  color: #cccccc;
}

/* 回款记录 */
.payment-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16rpx 0;
  border-bottom: 1rpx solid #f5f5f5;

  &:last-child {
    border-bottom: none;
  }
}

.payment-left {
  display: flex;
  flex-direction: column;
}

.payment-amount {
  font-size: 28rpx;
  font-weight: 500;
  color: #52c41a;
  margin-bottom: 4rpx;
}

.payment-time {
  font-size: 22rpx;
  color: #cccccc;
}

/* 底部操作按钮 */
.bottom-actions {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  gap: 20rpx;
  padding: 20rpx 24rpx;
  background-color: #ffffff;
  box-shadow: 0 -2rpx 12rpx rgba(0, 0, 0, 0.06);
  z-index: 100;
}

.action-btn {
  flex: 1;
  height: 80rpx;
  line-height: 80rpx;
  text-align: center;
  font-size: 28rpx;
  border-radius: 12rpx;
  border: none;

  &--reject {
    background-color: #ffffff;
    color: #ff4d4f;
    border: 2rpx solid #ff4d4f;

    &:active {
      background-color: #fff2f0;
    }
  }

  &--approve {
    background-color: #52c41a;
    color: #ffffff;

    &:active {
      background-color: #389e0d;
    }
  }

  &--payment {
    background-color: #1890ff;
    color: #ffffff;

    &:active {
      background-color: #096dd9;
    }
  }
}

/* 弹窗 */
.modal-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  width: 600rpx;
  background-color: #ffffff;
  border-radius: 16rpx;
  padding: 40rpx;
}

.modal-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333333;
  margin-bottom: 24rpx;
  display: block;
}

.modal-textarea {
  width: 100%;
  height: 200rpx;
  background-color: #f5f5f5;
  border: 2rpx solid #eeeeee;
  border-radius: 12rpx;
  padding: 20rpx;
  font-size: 28rpx;
  color: #333333;
  margin-bottom: 24rpx;
}

.modal-form-item {
  margin-bottom: 20rpx;
}

.modal-form-label {
  font-size: 26rpx;
  color: #666666;
  margin-bottom: 12rpx;
  display: block;
}

.input-field {
  width: 100%;
  height: 80rpx;
  background-color: #f5f5f5;
  border: 2rpx solid #eeeeee;
  border-radius: 12rpx;
  padding: 0 24rpx;
  font-size: 28rpx;
  color: #333333;
}

.modal-actions {
  display: flex;
  gap: 20rpx;
  margin-top: 30rpx;
}

.modal-btn {
  flex: 1;
  height: 76rpx;
  line-height: 76rpx;
  text-align: center;
  font-size: 28rpx;
  border-radius: 12rpx;
  border: none;

  &--cancel {
    background-color: #f5f5f5;
    color: #666666;

    &:active {
      background-color: #eeeeee;
    }
  }

  &--confirm {
    background-color: #1890ff;
    color: #ffffff;

    &:active {
      background-color: #096dd9;
    }
  }
}

.loading-state {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 60vh;
}

.loading-text {
  font-size: 28rpx;
  color: #999999;
}
</style>
