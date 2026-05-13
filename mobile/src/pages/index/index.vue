<template>
  <view class="home-page">
    <!-- 顶部问候 -->
    <view class="home-header">
      <view class="greeting">
        <text class="greeting-text">{{ greetingText }}，{{ userName }}</text>
        <text class="greeting-date">{{ currentDate }}</text>
      </view>
    </view>

    <!-- 数据概览卡片 2x2 -->
    <view class="overview-grid">
      <view class="overview-card" @tap="goTo('/pages/performance/index')">
        <text class="card-value">{{ overview.totalSales || '0.00' }}</text>
        <text class="card-label">本月销售额(元)</text>
      </view>
      <view class="overview-card" @tap="goTo('/pages/orders/index')">
        <text class="card-value">{{ overview.totalOrders || 0 }}</text>
        <text class="card-label">本月订单数</text>
      </view>
      <view class="overview-card" @tap="goTo('/pages/performance/index')">
        <text class="card-value card-value--green">{{ overview.totalProfit || '0.00' }}</text>
        <text class="card-label">本月利润(元)</text>
      </view>
      <view class="overview-card" @tap="goTo('/pages/orders/index?status=PENDING_APPROVAL')">
        <text class="card-value card-value--orange">{{ overview.pendingApproval || 0 }}</text>
        <text class="card-label">待审批订单</text>
      </view>
    </view>

    <!-- 快捷操作 -->
    <view class="section card">
      <text class="section-title">快捷操作</text>
      <view class="quick-actions">
        <view class="action-item" @tap="goTo('/pages/orders/create')">
          <view class="action-icon action-icon--blue">
            <text class="icon-text">+</text>
          </view>
          <text class="action-label">新建订单</text>
        </view>
        <view class="action-item" @tap="goTo('/pages/inventory/index')">
          <view class="action-icon action-icon--green">
            <text class="icon-text">查</text>
          </view>
          <text class="action-label">库存查询</text>
        </view>
        <view class="action-item" @tap="goTo('/pages/performance/index')">
          <view class="action-icon action-icon--orange">
            <text class="icon-text">提</text>
          </view>
          <text class="action-label">业绩查看</text>
        </view>
        <view class="action-item" @tap="handleScan">
          <view class="action-icon action-icon--purple">
            <text class="icon-text">扫</text>
          </view>
          <text class="action-label">扫码</text>
        </view>
      </view>
    </view>

    <!-- 最近订单动态 -->
    <view class="section card">
      <view class="section-header">
        <text class="section-title">最近订单动态</text>
        <text class="section-more" @tap="goTo('/pages/orders/index')">查看全部 ></text>
      </view>
      <view v-if="recentOrders.length > 0">
        <view
          v-for="(order, index) in recentOrders"
          :key="order.id || index"
          class="order-item"
          @tap="goTo('/pages/orders/detail?id=' + order.id)"
        >
          <view class="order-info">
            <view class="order-top">
              <text class="order-no">{{ order.order_no || '--' }}</text>
              <view class="tag" :class="getStatusClass(order.order_status)">
                <text>{{ getStatusText(order.order_status) }}</text>
              </view>
            </view>
            <text class="order-customer">{{ order.customer_name || '客户' }}</text>
          </view>
          <view class="order-right">
            <text class="order-amount">{{ order.final_amount || '--' }}元</text>
          </view>
        </view>
      </view>
      <view v-else class="empty-state">
        <text class="empty-state__text">暂无订单数据</text>
      </view>
    </view>

    <!-- 底部占位 -->
    <view style="height: 120rpx;"></view>

    <CustomTabBar :current="0" />
  </view>
</template>

<script>
import { ref, onMounted, computed } from 'vue'
import { useUserStore } from '../../store/user'
import { getOverview } from '../../api/performance'
import { getOrders } from '../../api/order'
import CustomTabBar from '../../components/CustomTabBar.vue'

export default {
  components: { CustomTabBar },
  setup() {
    const userStore = useUserStore()
    const overview = ref({})
    const recentOrders = ref([])

    const userName = computed(() => {
      return userStore.userInfo?.name || userStore.userInfo?.username || '同事'
    })

    const greetingText = computed(() => {
      const hour = new Date().getHours()
      if (hour < 12) return '早上好'
      if (hour < 18) return '下午好'
      return '晚上好'
    })

    const currentDate = computed(() => {
      const d = new Date()
      const weekDays = ['日', '一', '二', '三', '四', '五', '六']
      return `${d.getMonth() + 1}月${d.getDate()}日 星期${weekDays[d.getDay()]}`
    })

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

    const getStatusClass = (status) => {
      const map = {
        0: 'tag-warning',
        1: 'tag-success',
        2: 'tag-danger',
        3: 'tag-default',
        4: 'tag-purple'
      }
      return map[status] || ''
    }

    const goTo = (url) => {
      // tabbar 页面使用 switchTab，其他页面使用 navigateTo
      const tabbarPages = [
        '/pages/index/index',
        '/pages/orders/index',
        '/pages/performance/index',
        '/pages/inventory/index',
        '/pages/profile/index'
      ]
      const isTabbar = tabbarPages.some(path => url.startsWith(path))
      if (isTabbar) {
        uni.switchTab({ url })
      } else {
        uni.navigateTo({ url })
      }
    }

    const handleScan = () => {
      uni.scanCode({
        success: (res) => {
          uni.showToast({ title: '扫码结果: ' + res.result, icon: 'none' })
        },
        fail: () => {
          uni.showToast({ title: '扫码取消', icon: 'none' })
        }
      })
    }

    const loadData = async () => {
      try {
        const res = await getOverview()
        overview.value = res.data || {}
      } catch (e) {
        overview.value = {
          totalSales: '0.00',
          totalOrders: 0,
          totalProfit: '0.00',
          pendingApproval: 0
        }
      }

      try {
        const res = await getOrders({ page: 1, page_size: 5 })
        recentOrders.value = res.data?.records || res.data?.list || []
      } catch (e) {
        recentOrders.value = []
      }
    }

    // 下拉刷新
    const onPullDownRefresh = async () => {
      await loadData()
      uni.stopPullDownRefresh()
    }

    onMounted(() => {
      if (userStore.isLoggedIn()) {
        userStore.fetchUserInfo().catch(() => {})
        loadData()
      }
    })

    return {
      overview,
      recentOrders,
      userName,
      greetingText,
      currentDate,
      getStatusText,
      getStatusClass,
      goTo,
      handleScan,
      onPullDownRefresh
    }
  }
}
</script>

<style lang="scss" scoped>
.home-page {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.home-header {
  background: linear-gradient(135deg, #1890ff 0%, #096dd9 100%);
  padding: 40rpx 30rpx 60rpx;
}

.greeting {
  display: flex;
  flex-direction: column;
}

.greeting-text {
  font-size: 36rpx;
  font-weight: bold;
  color: #ffffff;
  margin-bottom: 8rpx;
}

.greeting-date {
  font-size: 24rpx;
  color: rgba(255, 255, 255, 0.8);
}

/* 2x2 数据卡片 */
.overview-grid {
  display: flex;
  flex-wrap: wrap;
  margin: -30rpx 24rpx 0;
  gap: 20rpx;
}

.overview-card {
  width: calc(50% - 10rpx);
  background-color: #ffffff;
  border-radius: 16rpx;
  padding: 24rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);

  &:active {
    opacity: 0.8;
  }
}

.card-value {
  font-size: 32rpx;
  font-weight: bold;
  color: #333333;
  margin-bottom: 8rpx;

  &--green {
    color: #52c41a;
  }

  &--orange {
    color: #faad14;
  }
}

.card-label {
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

/* 快捷操作 */
.quick-actions {
  display: flex;
  justify-content: space-around;
}

.action-item {
  display: flex;
  flex-direction: column;
  align-items: center;

  &:active {
    opacity: 0.8;
  }
}

.action-icon {
  width: 88rpx;
  height: 88rpx;
  border-radius: 20rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 12rpx;

  &--blue {
    background-color: #e6f7ff;
  }

  &--green {
    background-color: #f6ffed;
  }

  &--orange {
    background-color: #fff7e6;
  }

  &--purple {
    background-color: #f9f0ff;
  }
}

.icon-text {
  font-size: 32rpx;
  font-weight: bold;
  color: #1890ff;

  .action-icon--green & {
    color: #52c41a;
  }

  .action-icon--orange & {
    color: #faad14;
  }

  .action-icon--purple & {
    color: #722ed1;
  }
}

.action-label {
  font-size: 24rpx;
  color: #666666;
}

/* 订单动态列表 */
.order-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #eeeeee;

  &:last-child {
    border-bottom: none;
  }

  &:active {
    opacity: 0.8;
  }
}

.order-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.order-top {
  display: flex;
  align-items: center;
  margin-bottom: 6rpx;
}

.order-no {
  font-size: 26rpx;
  color: #999999;
  margin-right: 12rpx;
}

.order-customer {
  font-size: 28rpx;
  color: #333333;
  font-weight: 500;
}

.order-right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
}

.order-amount {
  font-size: 28rpx;
  color: #333333;
  font-weight: 500;
}

/* 标签扩展 */
.tag-default {
  background-color: #f5f5f5;
  color: #999999;
}

.tag-purple {
  background-color: #f9f0ff;
  color: #722ed1;
}
</style>
