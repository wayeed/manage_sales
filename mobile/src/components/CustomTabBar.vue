<template>
  <view class="custom-tabbar safe-bottom">
    <view
      v-for="(item, index) in tabList"
      :key="index"
      class="tabbar-item"
      :class="{ active: currentIndex === index }"
      @tap="switchTab(index)"
    >
      <view class="tabbar-icon">
        <!-- 首页图标 -->
        <view v-if="item.key === 'home'" class="icon-home">
          <view class="icon-home-roof" :class="{ active: currentIndex === index }"></view>
          <view class="icon-home-body" :class="{ active: currentIndex === index }"></view>
        </view>
        <!-- 订单图标 -->
        <view v-else-if="item.key === 'order'" class="icon-order">
          <view class="icon-order-rect" :class="{ active: currentIndex === index }"></view>
          <view class="icon-order-lines">
            <view class="icon-order-line" :class="{ active: currentIndex === index }"></view>
            <view class="icon-order-line" :class="{ active: currentIndex === index }"></view>
            <view class="icon-order-line" :class="{ active: currentIndex === index }"></view>
          </view>
        </view>
        <!-- 业绩图标 -->
        <view v-else-if="item.key === 'chart'" class="icon-chart">
          <view class="icon-chart-bar" :class="{ active: currentIndex === index }" style="height: 20rpx;"></view>
          <view class="icon-chart-bar" :class="{ active: currentIndex === index }" style="height: 32rpx;"></view>
          <view class="icon-chart-bar" :class="{ active: currentIndex === index }" style="height: 26rpx;"></view>
          <view class="icon-chart-bar" :class="{ active: currentIndex === index }" style="height: 36rpx;"></view>
        </view>
        <!-- 客户跟进图标（中间突出） -->
        <view v-else-if="item.key === 'customer'" class="icon-customer" :class="{ active: currentIndex === index }">
          <view class="icon-customer-head"></view>
          <view class="icon-customer-body"></view>
        </view>
        <!-- 我的图标 -->
        <view v-else-if="item.key === 'user'" class="icon-user">
          <view class="icon-user-head" :class="{ active: currentIndex === index }"></view>
          <view class="icon-user-body" :class="{ active: currentIndex === index }"></view>
        </view>
      </view>
      <text class="tabbar-text">{{ item.text }}</text>
    </view>
  </view>
</template>

<script>
export default {
  name: 'CustomTabBar',
  props: {
    current: {
      type: Number,
      default: 0
    }
  },
  data() {
    return {
      tabList: [
        { key: 'home', text: '首页', path: '/pages/index/index' },
        { key: 'order', text: '订单', path: '/pages/orders/index' },
        { key: 'customer', text: '客户跟进', path: '/pages/customer/follow-up' },
        { key: 'chart', text: '业绩', path: '/pages/performance/index' },
        { key: 'user', text: '我的', path: '/pages/profile/index' }
      ]
    }
  },
  computed: {
    currentIndex() {
      return this.current
    }
  },
  methods: {
    switchTab(index) {
      if (index === this.currentIndex) return
      uni.reLaunch({
        url: this.tabList[index].path
      })
    }
  }
}
</script>

<style lang="scss" scoped>
.custom-tabbar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  justify-content: space-around;
  height: 100rpx;
  background-color: #ffffff;
  border-top: 1rpx solid #eeeeee;
  z-index: 999;
}

.tabbar-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  height: 100%;
  position: relative;

  &.active .tabbar-text {
    color: #1890ff;
    font-weight: 600;
    font-size: 22rpx;
  }

  // 选中指示条
  &.active::after {
    content: '';
    position: absolute;
    top: 0;
    left: 50%;
    transform: translateX(-50%);
    width: 48rpx;
    height: 6rpx;
    background: #1890ff;
    border-radius: 0 0 4rpx 4rpx;
  }
}

.tabbar-icon {
  width: 48rpx;
  height: 48rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 4rpx;
}

.tabbar-text {
  font-size: 20rpx;
  color: #999999;
  line-height: 1;
}

/* ===== 首页图标 ===== */
.icon-home {
  position: relative;
  width: 40rpx;
  height: 40rpx;
}

.icon-home-roof {
  position: absolute;
  top: 4rpx;
  left: 50%;
  transform: translateX(-50%);
  width: 0;
  height: 0;
  border-left: 22rpx solid transparent;
  border-right: 22rpx solid transparent;
  border-bottom: 16rpx solid #999999;

  &.active {
    border-bottom-color: #1890ff;
  }
}

.icon-home-body {
  position: absolute;
  bottom: 2rpx;
  left: 50%;
  transform: translateX(-50%);
  width: 28rpx;
  height: 18rpx;
  border: 3rpx solid #999999;
  border-top: none;
  border-radius: 0 0 4rpx 4rpx;

  &.active {
    border-color: #1890ff;
  }
}

/* ===== 订单图标 ===== */
.icon-order {
  position: relative;
  width: 36rpx;
  height: 40rpx;
}

.icon-order-rect {
  width: 36rpx;
  height: 40rpx;
  border: 3rpx solid #999999;
  border-radius: 4rpx;

  &.active {
    border-color: #1890ff;
  }
}

.icon-order-lines {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  display: flex;
  flex-direction: column;
  gap: 4rpx;
}

.icon-order-line {
  width: 20rpx;
  height: 3rpx;
  background-color: #999999;
  border-radius: 2rpx;

  &.active {
    background-color: #1890ff;
  }
}

/* ===== 业绩图标 ===== */
.icon-chart {
  display: flex;
  align-items: flex-end;
  gap: 4rpx;
  width: 40rpx;
  height: 40rpx;
}

.icon-chart-bar {
  flex: 1;
  background-color: #999999;
  border-radius: 2rpx 2rpx 0 0;
  min-height: 12rpx;

  &.active {
    background-color: #1890ff;
  }
}

/* ===== 客户跟进图标（中间突出） ===== */
.tabbar-item:nth-child(3) {
  margin-top: -20rpx;
}

.tabbar-item:nth-child(3) .tabbar-icon {
  width: 96rpx;
  height: 96rpx;
  background: #e8f4ff;
  border-radius: 50%;
  box-shadow: 0 4rpx 16rpx rgba(24, 144, 255, 0.2);
  transition: all 0.2s ease;
}

.tabbar-item:nth-child(3).active .tabbar-icon {
  background: #1890ff;
  box-shadow: 0 6rpx 24rpx rgba(24, 144, 255, 0.5);
}

.tabbar-item:nth-child(3) .tabbar-text {
  color: #999;
  font-weight: 400;
  transition: all 0.2s ease;
}

.tabbar-item:nth-child(3).active .tabbar-text {
  color: #1890ff;
  font-weight: 600;
  font-size: 22rpx;
}

// 中间按钮不需要顶部指示条（已有圆形背景区分）
.tabbar-item:nth-child(3).active::after {
  display: none;
}

.icon-customer {
  position: relative;
  width: 36rpx;
  height: 40rpx;
}

.icon-customer-head {
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 16rpx;
  height: 16rpx;
  border: 3rpx solid #1890ff;
  border-radius: 50%;
  transition: border-color 0.2s ease;
}

.icon-customer-body {
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 30rpx;
  height: 18rpx;
  border: 3rpx solid #1890ff;
  border-top: none;
  border-radius: 0 0 16rpx 16rpx;
  transition: border-color 0.2s ease;
}

.tabbar-item:nth-child(3).active .icon-customer-head,
.tabbar-item:nth-child(3).active .icon-customer-body {
  border-color: #fff;
}

/* ===== 我的图标 ===== */
.icon-user {
  position: relative;
  width: 36rpx;
  height: 40rpx;
}

.icon-user-head {
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 16rpx;
  height: 16rpx;
  border: 3rpx solid #999999;
  border-radius: 50%;

  &.active {
    border-color: #1890ff;
  }
}

.icon-user-body {
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 30rpx;
  height: 18rpx;
  border: 3rpx solid #999999;
  border-top: none;
  border-radius: 0 0 16rpx 16rpx;

  &.active {
    border-color: #1890ff;
  }
}
</style>
