<template>
  <view class="performance-page">
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

    <!-- 4个统计卡片 -->
    <view class="stats-grid">
      <view class="stat-card card">
        <text class="stat-value">{{ stats.totalSales || '0.00' }}</text>
        <text class="stat-label">本月销售额(元)</text>
      </view>
      <view class="stat-card card">
        <text class="stat-value stat-value--green">{{ stats.totalProfit || '0.00' }}</text>
        <text class="stat-label">本月利润(元)</text>
      </view>
      <view class="stat-card card">
        <text class="stat-value stat-value--blue">{{ stats.totalCommission || '0.00' }}</text>
        <text class="stat-label">本月提成(元)</text>
      </view>
      <view class="stat-card card">
        <text class="stat-value stat-value--orange">{{ stats.commissionRank || '--' }}</text>
        <text class="stat-label">提成排名</text>
      </view>
    </view>

    <!-- 提成构成 -->
    <view class="section card">
      <text class="section-title">提成构成</text>
      <view class="chart-area">
        <view class="chart-bar-item" v-for="(item, index) in commissionBreakdown" :key="index">
          <view class="chart-bar-header">
            <text class="chart-bar-label">{{ item.label }}</text>
            <text class="chart-bar-value">{{ item.value }}元</text>
          </view>
          <view class="chart-bar-track">
            <view
              class="chart-bar-fill"
              :style="{ width: item.percent + '%', backgroundColor: item.color }"
            ></view>
          </view>
        </view>
      </view>
    </view>

    <!-- 最近提成记录 -->
    <view class="section card">
      <view class="section-header">
        <text class="section-title">最近提成记录</text>
        <text class="section-more" @tap="goTo('/pages/performance/commission')">查看全部 ></text>
      </view>
      <view v-if="recentCommissions.length > 0">
        <view
          v-for="(item, index) in recentCommissions"
          :key="item.id || index"
          class="commission-item"
        >
          <view class="commission-left">
            <text class="commission-order">{{ item.orderNo || '--' }}</text>
            <text class="commission-type">{{ item.typeName || '销售提成' }}</text>
          </view>
          <text class="commission-amount">+{{ item.amount || '0.00' }}元</text>
        </view>
      </view>
      <view v-else class="empty-state">
        <text class="empty-state__text">暂无提成记录</text>
      </view>
    </view>

    <!-- 底部占位 -->
    <view style="height: 120rpx;"></view>

    <CustomTabBar :current="2" />
  </view>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { getPerformance, getCommissions } from '../../api/performance'
import CustomTabBar from '../../components/CustomTabBar.vue'

export default {
  components: { CustomTabBar },
  setup() {
    const year = ref(new Date().getFullYear())
    const month = ref(new Date().getMonth() + 1)
    const stats = ref({})
    const recentCommissions = ref([])

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
      loadData()
    }

    const goTo = (url) => {
      uni.navigateTo({ url })
    }

    // 提成构成数据
    const commissionBreakdown = computed(() => {
      const items = [
        { label: '销售提成', value: stats.value.salesCommission || 0, color: '#1890ff' },
        { label: '团队分润', value: stats.value.teamShare || 0, color: '#52c41a' },
        { label: '基金池奖励', value: stats.value.fundReward || 0, color: '#faad14' },
        { label: '老带新奖励', value: stats.value.mentorReward || 0, color: '#722ed1' }
      ]
      const maxVal = Math.max(...items.map(i => i.value), 1)
      return items.map(item => ({
        ...item,
        percent: Math.round((item.value / maxVal) * 100)
      }))
    })

    const loadData = async () => {
      try {
        const res = await getPerformance({
          year: year.value,
          month: month.value
        })
        stats.value = res.data || {}
      } catch (e) {
        stats.value = {
          totalSales: '0.00',
          totalProfit: '0.00',
          totalCommission: '0.00',
          commissionRank: '--'
        }
      }

      try {
        const res = await getCommissions({
          period_value: `${year.value}-${String(month.value).padStart(2, '0')}`,
          page: 1,
          page_size: 5
        })
        recentCommissions.value = res.data?.records || res.data?.list || []
      } catch (e) {
        recentCommissions.value = []
      }
    }

    onMounted(() => {
      loadData()
    })

    return {
      year,
      month,
      currentMonth,
      stats,
      recentCommissions,
      commissionBreakdown,
      changeMonth,
      goTo
    }
  }
}
</script>

<style lang="scss" scoped>
.performance-page {
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

/* 2x2 统计卡片 */
.stats-grid {
  display: flex;
  flex-wrap: wrap;
  padding: 0 24rpx;
  gap: 20rpx;
}

.stat-card {
  width: calc(50% - 10rpx);
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 30rpx 16rpx;
}

.stat-value {
  font-size: 36rpx;
  font-weight: bold;
  color: #333333;
  margin-bottom: 8rpx;

  &--green {
    color: #52c41a;
  }

  &--blue {
    color: #1890ff;
  }

  &--orange {
    color: #faad14;
  }
}

.stat-label {
  font-size: 24rpx;
  color: #999999;
}

.section {
  margin: 24rpx;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333333;
}

.section-more {
  font-size: 24rpx;
  color: #999999;
}

/* 提成构成进度条 */
.chart-area {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.chart-bar-item {
  display: flex;
  flex-direction: column;
}

.chart-bar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8rpx;
}

.chart-bar-label {
  font-size: 26rpx;
  color: #666666;
}

.chart-bar-value {
  font-size: 26rpx;
  color: #333333;
  font-weight: 500;
}

.chart-bar-track {
  height: 16rpx;
  background-color: #f5f5f5;
  border-radius: 8rpx;
  overflow: hidden;
}

.chart-bar-fill {
  height: 100%;
  border-radius: 8rpx;
  transition: width 0.3s ease;
}

/* 提成记录列表 */
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
  font-size: 24rpx;
  color: #666666;
}

.commission-amount {
  font-size: 30rpx;
  font-weight: bold;
  color: #ff4d4f;
}
</style>
