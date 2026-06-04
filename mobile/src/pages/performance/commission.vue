<template>
  <view class="commission-page">
    <!-- 月份选择 -->
    <view class="month-selector card">
      <view class="month-arrow" @tap="changeMonth(-1)">
        <text class="arrow-text"><</text>
      </view>
      <text class="month-text">{{ currentMonth }}</text>
      <view class="month-arrow" @tap="changeMonth(1)">
        <text class="arrow-text">></text>
      </view>
    </view>

    <!-- 提成汇总 -->
    <view class="summary-card card">
      <view class="summary-row">
        <text class="summary-label">销售提成</text>
        <text class="summary-value">{{ formatAmount(summary.salesCommission) }}元</text>
      </view>
      <view class="summary-row">
        <text class="summary-label">团队分润</text>
        <text class="summary-value">{{ formatAmount(summary.teamShare) }}元</text>
      </view>
      <view class="summary-row">
        <text class="summary-label">基金池奖励</text>
        <text class="summary-value">{{ formatAmount(summary.fundReward) }}元</text>
      </view>
      <view class="summary-row">
        <text class="summary-label">老带新奖励</text>
        <text class="summary-value">{{ formatAmount(summary.mentorReward) }}元</text>
      </view>
      <view class="divider"></view>
      <view class="summary-row summary-total">
        <text class="summary-label summary-total-label">提成合计</text>
        <text class="summary-value summary-total-value">{{ formatAmount(summary.totalCommission) }}元</text>
      </view>
    </view>

    <!-- 提成明细列表 -->
    <view class="section card">
      <text class="section-title">提成明细</text>
      <view v-if="commissionList.length > 0">
        <view
          v-for="(item, index) in commissionList"
          :key="item.id || index"
          class="commission-item"
        >
          <view class="commission-main">
            <view class="commission-header">
              <text class="commission-order">{{ getOrderNo(item) }}</text>
              <text class="commission-amount">+{{ formatAmount(item.amount) }}元</text>
            </view>
            <view class="commission-meta">
              <text class="commission-type">{{ getCommissionTypeName(item.commission_type) }}</text>
              <text class="commission-dot">·</text>
              <text class="commission-time">{{ formatDateTime(item.created_at) }}</text>
              <text class="commission-dot">·</text>
              <view class="commission-status-tag" :class="getStatusClass(item.status)">
                <text>{{ getStatusName(item.status) }}</text>
              </view>
            </view>
            <view v-if="item.remark" class="commission-remark">
              <text>备注：{{ item.remark }}</text>
            </view>
          </view>
        </view>
      </view>
      <view v-else class="empty-state">
        <text class="empty-state__text">暂无提成记录</text>
      </view>

      <!-- 加载更多 -->
      <view v-if="loading" class="loading-more">
        <text class="loading-text">加载中...</text>
      </view>
      <view v-else-if="!hasMore && commissionList.length > 0" class="loading-more">
        <text class="loading-text">没有更多了</text>
      </view>
    </view>

    <!-- 底部占位 -->
    <view style="height: 60rpx;"></view>
  </view>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { getCommissions } from '../../api/performance'

export default {
  setup() {
    const year = ref(new Date().getFullYear())
    const month = ref(new Date().getMonth() + 1)
    const summary = ref({})
    const commissionList = ref([])
    const loading = ref(false)
    const page = ref(1)
    const hasMore = ref(true)

    const currentMonth = computed(() => {
      return `${year.value}年${String(month.value).padStart(2, '0')}月`
    })

    const changeMonth = (delta) => {
      month.value += delta
      if (month.value > 12) {
        month.value = 1
        year.value++
      } else if (month.value < 1) {
        month.value = 12
        year.value--
      }
      page.value = 1
      commissionList.value = []
      hasMore.value = true
      loadData()
    }

    // 格式化金额
    const formatAmount = (val) => {
      if (val === undefined || val === null || val === '') return '0.00'
      const num = parseFloat(val)
      if (isNaN(num)) return '0.00'
      return num.toFixed(2)
    }

    // 格式化日期时间
    const formatDateTime = (dateStr) => {
      if (!dateStr) return '--'
      const date = new Date(dateStr)
      if (isNaN(date.getTime())) return dateStr
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      const hours = String(date.getHours()).padStart(2, '0')
      const minutes = String(date.getMinutes()).padStart(2, '0')
      return `${month}-${day} ${hours}:${minutes}`
    }

    // 获取订单号
    const getOrderNo = (item) => {
      return item.order?.order_no || item.order_id || '--'
    }

    // 提成类型名称映射
    const getCommissionTypeName = (type) => {
      const map = {
        1: '业务员提成',
        2: '同行分成',
        3: '主管团队分润',
        4: '店长团队分润',
        5: '基金池奖励',
        6: '老带新奖励'
      }
      return map[type] || '销售提成'
    }

    // 提成状态名称映射
    const getStatusName = (status) => {
      const map = {
        0: '待回款',
        1: '可发放',
        2: '已发放'
      }
      return map[status] || '未知'
    }

    // 提成状态样式类
    const getStatusClass = (status) => {
      const map = {
        0: 'status-pending',
        1: 'status-ready',
        2: 'status-paid'
      }
      return map[status] || 'status-pending'
    }

    const loadData = async () => {
      if (loading.value) return
      loading.value = true
      try {
        const res = await getCommissions({
          period_value: `${year.value}-${String(month.value).padStart(2, '0')}`,
          page: page.value,
          page_size: 10
        })
        summary.value = res.data?.summary || {}
        const list = res.data?.records || res.data?.list || []
        if (page.value === 1) {
          commissionList.value = list
        } else {
          commissionList.value = [...commissionList.value, ...list]
        }
        const total = res.data?.total || 0
        hasMore.value = commissionList.value.length < total
      } catch (e) {
        console.error('加载提成明细失败:', e)
        if (page.value === 1) {
          commissionList.value = []
        }
      } finally {
        loading.value = false
      }
    }

    onMounted(() => {
      loadData()
    })

    return {
      currentMonth,
      summary,
      commissionList,
      loading,
      hasMore,
      changeMonth,
      formatAmount,
      formatDateTime,
      getOrderNo,
      getCommissionTypeName,
      getStatusName,
      getStatusClass
    }
  },
  onReachBottom() {
    if (this.hasMore && !this.loading) {
      this.page++
      this.loadData()
    }
  }
}
</script>

<style lang="scss" scoped>
.commission-page {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.card {
  background-color: #ffffff;
  border-radius: 16rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
}

.month-selector {
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 24rpx;
  padding: 24rpx;
}

.month-arrow {
  width: 60rpx;
  height: 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;

  &:active {
    opacity: 0.6;
  }
}

.arrow-text {
  font-size: 32rpx;
  color: #666666;
}

.month-text {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin: 0 40rpx;
}

.summary-card {
  margin: 0 24rpx 24rpx;
  padding: 24rpx;
}

.summary-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14rpx 0;
}

.summary-label {
  font-size: 28rpx;
  color: #666666;
}

.summary-value {
  font-size: 28rpx;
  font-weight: 500;
  color: #333333;
}

.summary-total {
  padding-top: 16rpx;
}

.summary-total-label {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.summary-total-value {
  font-size: 32rpx;
  font-weight: 700;
  color: #ff4d4f;
}

.divider {
  height: 1rpx;
  background-color: #eeeeee;
  margin: 12rpx 0;
}

.section {
  margin: 0 24rpx 24rpx;
  padding: 24rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 20rpx;
}

// 提成明细列表 - 优化布局
.commission-item {
  padding: 24rpx 0;
  border-bottom: 1rpx solid #f0f0f0;

  &:last-child {
    border-bottom: none;
  }
}

.commission-main {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}

.commission-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.commission-order {
  font-size: 28rpx;
  color: #333333;
  font-weight: 500;
}

.commission-amount {
  font-size: 30rpx;
  font-weight: 700;
  color: #ff4d4f;
}

.commission-meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8rpx;
}

.commission-type {
  font-size: 24rpx;
  color: #666666;
}

.commission-dot {
  font-size: 24rpx;
  color: #cccccc;
}

.commission-time {
  font-size: 24rpx;
  color: #999999;
}

.commission-remark {
  padding: 12rpx 16rpx;
  background-color: #f5f5f5;
  border-radius: 8rpx;
  margin-top: 4rpx;

  text {
    font-size: 24rpx;
    color: #999999;
  }
}

// 状态标签
.commission-status-tag {
  padding: 4rpx 12rpx;
  border-radius: 6rpx;
  font-size: 22rpx;

  text {
    font-weight: 500;
  }
}

.status-pending {
  background-color: #fff7e6;
  color: #fa8c16;
}

.status-ready {
  background-color: #e6f7ff;
  color: #1890ff;
}

.status-paid {
  background-color: #f6ffed;
  color: #52c41a;
}

// 空状态
.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 60rpx 0;

  &__text {
    font-size: 28rpx;
    color: #999999;
  }
}

.loading-more {
  display: flex;
  justify-content: center;
  padding: 30rpx 0;
}

.loading-text {
  font-size: 24rpx;
  color: #999999;
}
</style>
