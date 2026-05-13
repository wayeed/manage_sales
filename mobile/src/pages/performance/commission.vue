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
        <text class="summary-value">{{ summary.salesCommission || '0.00' }}元</text>
      </view>
      <view class="summary-row">
        <text class="summary-label">团队分润</text>
        <text class="summary-value">{{ summary.teamShare || '0.00' }}元</text>
      </view>
      <view class="summary-row">
        <text class="summary-label">基金池奖励</text>
        <text class="summary-value">{{ summary.fundReward || '0.00' }}元</text>
      </view>
      <view class="summary-row">
        <text class="summary-label">老带新奖励</text>
        <text class="summary-value">{{ summary.mentorReward || '0.00' }}元</text>
      </view>
      <view class="divider"></view>
      <view class="summary-row summary-total">
        <text class="summary-label summary-total-label">提成合计</text>
        <text class="summary-value summary-total-value">{{ summary.totalCommission || '0.00' }}元</text>
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
          <view class="commission-left">
            <text class="commission-order">{{ item.orderNo || '--' }}</text>
            <text class="commission-type">{{ item.typeName || '销售提成' }}</text>
            <text class="commission-date">{{ item.createTime || '' }}</text>
          </view>
          <view class="commission-right">
            <text class="commission-amount">+{{ item.amount || '0.00' }}</text>
            <view class="tag" :class="getCommissionStatusClass(item.status)">
              <text>{{ getCommissionStatusText(item.status) }}</text>
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

    const getCommissionStatusText = (status) => {
      const map = {
        'PENDING': '待发放',
        'PAID': '已发放',
        'CANCELLED': '已取消'
      }
      return map[status] || status || '未知'
    }

    const getCommissionStatusClass = (status) => {
      const map = {
        'PENDING': 'tag-warning',
        'PAID': 'tag-success',
        'CANCELLED': 'tag-danger'
      }
      return map[status] || ''
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
        commissionList.value = []
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
      getCommissionStatusText,
      getCommissionStatusClass
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
  font-weight: bold;
  color: #333333;
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
  font-weight: bold;
  color: #333333;
}

.summary-total-value {
  font-size: 32rpx;
  font-weight: bold;
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
  font-weight: bold;
  color: #333333;
  margin-bottom: 20rpx;
}

.commission-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #eeeeee;

  &:last-child {
    border-bottom: none;
  }
}

.commission-left {
  display: flex;
  flex-direction: column;
}

.commission-order {
  font-size: 26rpx;
  color: #999999;
  margin-bottom: 6rpx;
}

.commission-type {
  font-size: 26rpx;
  color: #333333;
  margin-bottom: 4rpx;
}

.commission-date {
  font-size: 22rpx;
  color: #cccccc;
}

.commission-right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 8rpx;
}

.commission-amount {
  font-size: 30rpx;
  font-weight: bold;
  color: #ff4d4f;
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
