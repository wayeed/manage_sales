<template>
  <view class="home-page">
    <!-- 顶部问候 -->
    <view class="home-header">
      <view class="greeting">
        <text class="greeting-text">{{ greetingText }}，{{ userName }}</text>
        <text class="greeting-date">{{ currentDate }}</text>
      </view>
      <!-- 审批提醒入口 -->
      <view class="header-actions" @tap="goToApproval">
        <view class="approval-icon">
          <text class="approval-icon-text">审</text>
          <view v-if="pendingCount > 0" class="approval-badge">{{ pendingCount > 99 ? '99+' : pendingCount }}</view>
        </view>
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
            <view class="order-meta">
              <text class="order-meta-item">{{ order.salesman_name || order.salesman?.real_name || '--' }}</text>
              <text class="order-meta-divider">·</text>
              <text class="order-meta-item">{{ formatOrderTime(order.created_at) }}</text>
            </view>
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
import { onPullDownRefresh } from '@dcloudio/uni-app'
import { useUserStore } from '../../store/user'
import { getOverview } from '../../api/performance'
import { getOrders } from '../../api/order'
import { getPendingApprovals } from '../../api/approval'
import { getPendingOutboundRequests } from '../../api/outbound-request'
import CustomTabBar from '../../components/CustomTabBar.vue'

export default {
  components: { CustomTabBar },
  setup() {
    const userStore = useUserStore()
    const overview = ref({})
    const recentOrders = ref([])
    const pendingCount = ref(0)

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

    // 格式化订单时间（MM-DD HH:mm）
    const formatOrderTime = (dateStr) => {
      if (!dateStr) return '--'
      const date = new Date(dateStr)
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      const hour = String(date.getHours()).padStart(2, '0')
      const min = String(date.getMinutes()).padStart(2, '0')
      return `${month}-${day} ${hour}:${min}`
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
        uni.reLaunch({ url })
      } else {
        uni.navigateTo({ url })
      }
    }

    // 跳转到审批管理页面
    const goToApproval = () => {
      uni.navigateTo({ url: '/pages/approval/pending' })
    }

    // 查询待审批数量
    const fetchPendingCount = async () => {
      try {
        const res = await getPendingApprovals({ page: 1, page_size: 1 })
        let total = res.data?.total || 0

        // 同时查询待审批的出库申请数量
        try {
          const outboundRes = await getPendingOutboundRequests({ page: 1, page_size: 1 })
          total += outboundRes.data?.total || 0
        } catch (e) {
          // 忽略出库审批查询错误
        }

        pendingCount.value = total
      } catch (e) {
        pendingCount.value = 0
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
    const handlePullDownRefresh = async () => {
      await loadData()
      await fetchPendingCount()
      uni.stopPullDownRefresh()
    }

    onPullDownRefresh(handlePullDownRefresh)

    onMounted(() => {
      if (userStore.isLoggedIn()) {
        userStore.fetchUserInfo().catch(() => {})
        loadData()
        fetchPendingCount()
      }
    })

    return {
      overview,
      recentOrders,
      pendingCount,
      userName,
      greetingText,
      currentDate,
      getStatusText,
      getStatusClass,
      formatOrderTime,
      goTo,
      goToApproval,
      handleScan
    }
  }
}
</script>

<style lang="scss" scoped>
.home-page {
  min-height: 100vh;
  background-color: #f5f5f5;
}

/* 卡片基础样式 */
.card {
  background-color: #ffffff;
  border-radius: 16rpx;
  padding: 24rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
}

.home-header {
  background: linear-gradient(135deg, #1890ff 0%, #096dd9 100%);
  padding: 40rpx 30rpx 60rpx;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.greeting {
  display: flex;
  flex-direction: column;
}

/* 头部操作区 */
.header-actions {
  display: flex;
  align-items: center;
}

/* 审批图标 */
.approval-icon {
  position: relative;
  width: 64rpx;
  height: 64rpx;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;

  &:active {
    background: rgba(255, 255, 255, 0.3);
  }
}

.approval-icon-text {
  font-size: 28rpx;
  color: #ffffff;
  font-weight: 500;
}

/* 审批角标 */
.approval-badge {
  position: absolute;
  top: -4rpx;
  right: -4rpx;
  min-width: 32rpx;
  height: 32rpx;
  padding: 0 8rpx;
  background: #ff4d4f;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20rpx;
  color: #ffffff;
  font-weight: bold;
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

.order-meta {
  display: flex;
  align-items: center;
  margin-top: 6rpx;
}

.order-meta-item {
  font-size: 22rpx;
  color: #999999;
}

.order-meta-divider {
  font-size: 22rpx;
  color: #cccccc;
  margin: 0 10rpx;
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
