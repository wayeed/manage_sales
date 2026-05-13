<template>
  <view class="orders-page">
    <!-- 状态筛选 Tab -->
    <scroll-view scroll-x class="tabs-scroll">
      <view class="tabs">
        <view
          v-for="(tab, index) in tabs"
          :key="index"
          class="tab-item"
          :class="{ active: currentTab === index }"
          @tap="switchTab(index)"
        >
          <text class="tab-text">{{ tab.text }}</text>
          <view v-if="currentTab === index" class="tab-line"></view>
        </view>
      </view>
    </scroll-view>

    <!-- 搜索栏 -->
    <view class="search-bar">
      <view class="search-input">
        <text class="search-icon">&#x1F50D;</text>
        <input
          v-model="keyword"
          class="search-field"
          placeholder="搜索订单号/客户名"
          placeholder-class="search-placeholder"
          confirm-type="search"
          @confirm="handleSearch"
        />
        <text v-if="keyword" class="search-clear" @tap="clearSearch">x</text>
      </view>
    </view>

    <!-- 订单列表 -->
    <view class="order-list">
      <view
        v-for="(order, index) in orderList"
        :key="order.id || index"
        class="order-card card"
        @tap="goDetail(order.id)"
      >
        <view class="order-header">
          <text class="order-no">{{ order.order_no || '--' }}</text>
          <view class="tag" :class="getStatusClass(order.order_status)">
            <text>{{ getStatusText(order.order_status) }}</text>
          </view>
        </view>
        <view class="order-body">
          <view class="order-row">
            <text class="order-label">客户</text>
            <text class="order-value">{{ order.customer_name || '--' }}</text>
          </view>
          <view class="order-row">
            <text class="order-label">商品数量</text>
            <text class="order-value">{{ order.sku_count || order.total_quantity || 0 }}件</text>
          </view>
          <view class="order-row">
            <text class="order-label">金额</text>
            <text class="order-value order-amount">{{ order.final_amount || '0.00' }}元</text>
          </view>
        </view>
        <view class="order-footer">
          <text class="order-time">{{ order.created_at || '' }}</text>
          <text class="order-arrow">></text>
        </view>
      </view>

      <!-- 空状态 -->
      <view v-if="!loading && orderList.length === 0" class="empty-state">
        <text class="empty-state__icon">&#x1F4E6;</text>
        <text class="empty-state__text">暂无订单</text>
      </view>

      <!-- 加载更多 -->
      <view v-if="loading" class="loading-more">
        <text class="loading-text">加载中...</text>
      </view>
      <view v-else-if="!hasMore && orderList.length > 0" class="loading-more">
        <text class="loading-text">没有更多了</text>
      </view>
    </view>

    <!-- 新建按钮 -->
    <view class="fab-btn" @tap="goCreate">
      <text class="fab-icon">+</text>
    </view>

    <!-- 底部占位 -->
    <view style="height: 120rpx;"></view>

    <CustomTabBar :current="1" />
  </view>
</template>

<script>
import { ref, onMounted } from 'vue'
import { getOrders } from '../../api/order'
import CustomTabBar from '../../components/CustomTabBar.vue'

export default {
  components: { CustomTabBar },
  data() {
    return {
      tabs: [
        { text: '全部', status: '' },
        { text: '待审批', status: 0 },
        { text: '已生效', status: 1 },
        { text: '已驳回', status: 2 },
        { text: '已取消', status: 3 },
        { text: '已退货', status: 4 }
      ],
      currentTab: 0,
      keyword: '',
      orderList: [],
      loading: false,
      page: 1,
      total: 0,
      hasMore: true
    }
  },
  methods: {
    switchTab(index) {
      this.currentTab = index
      this.page = 1
      this.orderList = []
      this.hasMore = true
      this.loadOrders()
    },

    clearSearch() {
      this.keyword = ''
      this.page = 1
      this.orderList = []
      this.hasMore = true
      this.loadOrders()
    },

    handleSearch() {
      this.page = 1
      this.orderList = []
      this.hasMore = true
      this.loadOrders()
    },

    async loadOrders() {
      if (this.loading) return
      this.loading = true
      try {
        const params = {
          page: this.page,
          size: 10,
          keyword: this.keyword
        }
        const status = this.tabs[this.currentTab].status
        if (status) {
          params.order_status = status
        }
        const res = await getOrders(params)
        const list = res.data?.records || res.data?.list || []
        if (this.page === 1) {
          this.orderList = list
        } else {
          this.orderList = [...this.orderList, ...list]
        }
        this.total = res.data?.total || 0
        this.hasMore = this.orderList.length < this.total
      } catch (e) {
        console.error('加载订单失败:', e)
      } finally {
        this.loading = false
      }
    },

    getStatusText(status) {
      const map = {
        0: '待审批',
        1: '已生效',
        2: '已驳回',
        3: '已取消',
        4: '已退货'
      }
      return map[status] || status || '未知'
    },

    getStatusClass(status) {
      const map = {
        0: 'tag-warning',
        1: 'tag-success',
        2: 'tag-danger',
        3: 'tag-default',
        4: 'tag-purple'
      }
      return map[status] || ''
    },

    goDetail(id) {
      uni.navigateTo({ url: `/pages/orders/detail?id=${id}` })
    },

    goCreate() {
      uni.navigateTo({ url: '/pages/orders/create' })
    }
  },
  onReachBottom() {
    if (this.hasMore && !this.loading) {
      this.page++
      this.loadOrders()
    }
  },
  onPullDownRefresh() {
    this.page = 1
    this.orderList = []
    this.hasMore = true
    this.loadOrders().then(() => {
      uni.stopPullDownRefresh()
    })
  },
  onShow() {
    this.page = 1
    this.orderList = []
    this.hasMore = true
    this.loadOrders()
  }
}
</script>

<style lang="scss" scoped>
.orders-page {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.tabs-scroll {
  white-space: nowrap;
  background-color: #ffffff;
  border-bottom: 1rpx solid #eeeeee;
}

.tabs {
  display: inline-flex;
  padding: 0 10rpx;
}

.tab-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24rpx 24rpx 20rpx;
  position: relative;
}

.tab-text {
  font-size: 28rpx;
  color: #666666;
  white-space: nowrap;
}

.tab-item.active .tab-text {
  color: #1890ff;
  font-weight: 500;
}

.tab-line {
  position: absolute;
  bottom: 0;
  width: 48rpx;
  height: 4rpx;
  background-color: #1890ff;
  border-radius: 2rpx;
}

.search-bar {
  padding: 16rpx 24rpx;
  background-color: #ffffff;
}

.search-input {
  display: flex;
  align-items: center;
  background-color: #f5f5f5;
  border-radius: 32rpx;
  padding: 12rpx 24rpx;
}

.search-icon {
  font-size: 28rpx;
  margin-right: 12rpx;
}

.search-field {
  flex: 1;
  font-size: 26rpx;
  height: 48rpx;
}

.search-clear {
  font-size: 28rpx;
  color: #cccccc;
  padding: 0 8rpx;
}

.search-placeholder {
  color: #cccccc;
  font-size: 26rpx;
}

.order-list {
  padding: 24rpx;
}

.order-card {
  &:active {
    opacity: 0.9;
  }
}

.order-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20rpx;
}

.order-no {
  font-size: 26rpx;
  color: #999999;
}

.order-body {
  margin-bottom: 16rpx;
}

.order-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8rpx 0;
}

.order-label {
  font-size: 26rpx;
  color: #999999;
  width: 120rpx;
}

.order-value {
  flex: 1;
  font-size: 28rpx;
  color: #333333;
  text-align: right;
}

.order-amount {
  color: #ff4d4f;
  font-weight: 500;
}

.order-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 16rpx;
  border-top: 1rpx solid #eeeeee;
}

.order-time {
  font-size: 22rpx;
  color: #cccccc;
}

.order-arrow {
  font-size: 24rpx;
  color: #cccccc;
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

.fab-btn {
  position: fixed;
  right: 40rpx;
  bottom: 180rpx;
  width: 100rpx;
  height: 100rpx;
  border-radius: 50%;
  background-color: #1890ff;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4rpx 16rpx rgba(24, 144, 255, 0.4);
  z-index: 100;

  &:active {
    background-color: #096dd9;
  }
}

.fab-icon {
  font-size: 48rpx;
  color: #ffffff;
  line-height: 1;
  margin-top: -4rpx;
}

.tag-default {
  background-color: #f5f5f5;
  color: #999999;
}

.tag-purple {
  background-color: #f9f0ff;
  color: #722ed1;
}
</style>
