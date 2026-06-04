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
        <text class="stat-value">{{ formatAmount(stats.totalSales) }}</text>
        <text class="stat-label">本月销售额(元)</text>
      </view>
      <view class="stat-card card">
        <text class="stat-value stat-value--green">{{ formatAmount(stats.totalProfit) }}</text>
        <text class="stat-label">本月利润(元)</text>
      </view>
      <view class="stat-card card">
        <text class="stat-value stat-value--blue">{{ formatAmount(stats.totalCommission) }}</text>
        <text class="stat-label">本月提成(元)</text>
      </view>
      <view class="stat-card card">
        <text class="stat-value stat-value--orange">{{ formatRank(stats.commissionRank) }}</text>
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
            <text class="chart-bar-value">{{ formatAmount(item.value) }}元</text>
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
          <view class="commission-main">
            <view class="commission-header">
              <text class="commission-order">{{ getOrderNo(item) }}</text>
              <text class="commission-amount">+{{ formatAmount(item.amount) }}元</text>
            </view>
            <view class="commission-meta">
              <text class="commission-type">{{ getCommissionTypeName(item.commission_type) }}</text>
              <text class="commission-dot">·</text>
              <text class="commission-time">{{ formatDate(item.created_at) }}</text>
              <text class="commission-dot">·</text>
              <view class="commission-status-tag" :class="getStatusClass(item.status)">
                <text>{{ getStatusName(item.status) }}</text>
              </view>
            </view>
          </view>
        </view>
      </view>
      <view v-else class="empty-state">
        <text class="empty-state__text">暂无提成记录</text>
      </view>
    </view>

    <!-- 底部占位 -->
    <view style="height: 120rpx;"></view>

    <CustomTabBar :current="3" />
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

    // 格式化金额
    const formatAmount = (val) => {
      if (val === undefined || val === null || val === '') return '0.00'
      const num = parseFloat(val)
      if (isNaN(num)) return '0.00'
      return num.toFixed(2)
    }

    // 格式化排名
    const formatRank = (rank) => {
      if (!rank || rank <= 0) return '--'
      return rank
    }

    // 格式化日期
    const formatDate = (dateStr) => {
      if (!dateStr) return '--'
      const date = new Date(dateStr)
      if (isNaN(date.getTime())) return dateStr
      return `${date.getMonth() + 1}/${date.getDate()}`
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

    // 提成构成数据
    const commissionBreakdown = computed(() => {
      const items = [
        { label: '销售提成', value: stats.value.salesCommission || 0, color: '#1890ff' },
        { label: '团队分润', value: stats.value.teamShare || 0, color: '#52c41a' },
        { label: '基金池奖励', value: stats.value.fundReward || 0, color: '#faad14' },
        { label: '老带新奖励', value: stats.value.mentorReward || 0, color: '#722ed1' }
      ]
      const maxVal = Math.max(...items.map(i => parseFloat(i.value) || 0), 1)
      return items.map(item => ({
        ...item,
        percent: Math.round(((parseFloat(item.value) || 0) / maxVal) * 100)
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
        console.error('加载业绩数据失败:', e)
        stats.value = {
          totalSales: '0.00',
          totalProfit: '0.00',
          totalCommission: '0.00',
          commissionRank: 0,
          salesCommission: '0.00',
          teamShare: '0.00',
          fundReward: '0.00',
          mentorReward: '0.00'
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
        console.error('加载提成记录失败:', e)
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
      goTo,
      formatAmount,
      formatRank,
      formatDate,
      getOrderNo,
      getCommissionTypeName,
      getStatusName,
      getStatusClass
    }
  }
}
</script>

<style lang="scss" scoped>
.performance-page {
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
  font-weight: 700;
  color: #1a1a1a;
  margin-bottom: 8rpx;

  &--green {
    color: #52c41a;
  }

  &--blue {
    color: #1890ff;
  }

  &--orange {
    color: #fa8c16;
  }
}

.stat-label {
  font-size: 24rpx;
  color: #999999;
}

.section {
  margin: 24rpx;
  padding: 32rpx;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.section-more {
  font-size: 24rpx;
  color: #999999;

  &:active {
    color: #1890ff;
  }
}

/* 提成构成进度条 */
.chart-area {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.chart-bar-item {
  display: flex;
  flex-direction: column;
}

.chart-bar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12rpx;
}

.chart-bar-label {
  font-size: 26rpx;
  color: #666666;
}

.chart-bar-value {
  font-size: 26rpx;
  color: #1a1a1a;
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

/* 提成记录列表 - 优化布局 */
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

/* 状态标签 */
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

/* 空状态 */
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
</style>
