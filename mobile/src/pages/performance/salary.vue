<template>
  <view class="salary-page">
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

    <!-- 工资概览卡片 -->
    <view class="salary-overview">
      <view class="salary-item">
        <text class="salary-label">应发工资</text>
        <view class="salary-amount">
          <text class="salary-symbol">¥</text>
          <text class="salary-value">{{ salary.grossSalary || '0.00' }}</text>
        </view>
      </view>
      <view class="salary-divider"></view>
      <view class="salary-item">
        <text class="salary-label">实发工资</text>
        <view class="salary-amount">
          <text class="salary-symbol">¥</text>
          <text class="salary-value salary-value--highlight">{{ salary.netSalary || '0.00' }}</text>
        </view>
      </view>
    </view>

    <!-- 工资明细 -->
    <view class="detail-section card">
      <text class="section-title">收入明细</text>
      <view class="detail-row">
        <text class="detail-label">基本工资</text>
        <text class="detail-value">{{ salary.baseSalary || '0.00' }}元</text>
      </view>
      <view class="detail-row">
        <text class="detail-label">销售提成</text>
        <text class="detail-value detail-value--blue">{{ salary.salesCommission || '0.00' }}元</text>
      </view>
      <view class="detail-row">
        <text class="detail-label">团队分润</text>
        <text class="detail-value">{{ salary.teamShare || '0.00' }}元</text>
      </view>
      <view class="detail-row">
        <text class="detail-label">基金池奖励</text>
        <text class="detail-value">{{ salary.fundReward || '0.00' }}元</text>
      </view>
      <view class="detail-row">
        <text class="detail-label">老带新奖励</text>
        <text class="detail-value">{{ salary.mentorReward || '0.00' }}元</text>
      </view>
      <view class="detail-row">
        <text class="detail-label">奖金</text>
        <text class="detail-value detail-value--green">{{ salary.bonus || '0.00' }}元</text>
      </view>
    </view>

    <view class="detail-section card">
      <text class="section-title">扣除明细</text>
      <view class="detail-row">
        <text class="detail-label">社保</text>
        <text class="detail-value detail-value--red">-{{ salary.socialInsurance || '0.00' }}元</text>
      </view>
      <view class="detail-row">
        <text class="detail-label">公积金</text>
        <text class="detail-value detail-value--red">-{{ salary.housingFund || '0.00' }}元</text>
      </view>
      <view class="detail-row">
        <text class="detail-label">个人所得税</text>
        <text class="detail-value detail-value--red">-{{ salary.tax || '0.00' }}元</text>
      </view>
      <view class="detail-row">
        <text class="detail-label">其他扣款</text>
        <text class="detail-value detail-value--red">-{{ salary.otherDeduction || '0.00' }}元</text>
      </view>
    </view>

    <!-- 底部占位 -->
    <view style="height: 60rpx;"></view>
  </view>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { getPerformance } from '../../api/performance'

export default {
  setup() {
    const year = ref(new Date().getFullYear())
    const month = ref(new Date().getMonth() + 1)
    const salary = ref({})

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

    const loadData = async () => {
      try {
        const res = await getPerformance({
          year: year.value,
          month: month.value,
          type: 'salary'
        })
        salary.value = res.data?.salary || {}
      } catch (e) {
        salary.value = {}
      }
    }

    onMounted(() => {
      loadData()
    })

    return { currentMonth, salary, changeMonth }
  }
}
</script>

<style lang="scss" scoped>
.salary-page {
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

/* 工资概览 */
.salary-overview {
  margin: 0 24rpx 24rpx;
  padding: 36rpx 30rpx;
  background: linear-gradient(135deg, #1890ff 0%, #096dd9 100%);
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: space-around;
}

.salary-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.salary-label {
  font-size: 24rpx;
  color: rgba(255, 255, 255, 0.8);
  margin-bottom: 12rpx;
}

.salary-amount {
  display: flex;
  align-items: baseline;
}

.salary-symbol {
  font-size: 28rpx;
  color: #ffffff;
  margin-right: 4rpx;
}

.salary-value {
  font-size: 40rpx;
  font-weight: bold;
  color: #ffffff;

  &--highlight {
    font-size: 48rpx;
  }
}

.salary-divider {
  width: 1rpx;
  height: 80rpx;
  background-color: rgba(255, 255, 255, 0.3);
}

/* 工资明细 */
.detail-section {
  margin: 0 24rpx 24rpx;
  padding: 24rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333333;
  margin-bottom: 20rpx;
}

.detail-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14rpx 0;
  border-bottom: 1rpx solid #f5f5f5;

  &:last-child {
    border-bottom: none;
  }
}

.detail-label {
  font-size: 28rpx;
  color: #666666;
}

.detail-value {
  font-size: 28rpx;
  color: #333333;
  font-weight: 500;

  &--blue {
    color: #1890ff;
  }

  &--green {
    color: #52c41a;
  }

  &--red {
    color: #ff4d4f;
  }
}
</style>
