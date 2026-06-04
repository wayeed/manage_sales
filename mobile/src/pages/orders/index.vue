<template>
  <view class="orders-page">
    <!-- 一级标签 -->
    <view class="tabs-container">
      <view class="tabs">
        <view
          v-for="(tab, index) in primaryTabs"
          :key="index"
          class="tab-item"
          :class="{ active: currentTab === index, 'tab-highlight': tab.highlight }"
          @tap="switchTab(index)"
        >
          <text class="tab-text">{{ tab.text }}</text>
          <view v-if="tab.count > 0" class="tab-badge">{{ tab.count > 99 ? '99+' : tab.count }}</view>
        </view>
      </view>
      <!-- 筛选按钮 -->
      <view class="filter-btn" :class="{ 'filter-active': hasActiveFilter }" @tap="openFilter">
        <text class="filter-icon">筛选</text>
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
            <text class="order-label">业务员</text>
            <text class="order-value">{{ order.salesman_name || order.salesman?.real_name || '--' }}</text>
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
        <!-- 底部状态行 -->
        <view class="order-footer">
          <view class="order-status-row">
            <view class="status-item" :class="getPaymentStatusClass(order.payment_status)">
              <text class="status-icon">💰</text>
              <text class="status-text">{{ getPaymentStatusText(order.payment_status) }}</text>
            </view>
            <view class="status-item" :class="getDeliveryStatusClass(order.delivery_status)">
              <text class="status-icon">🚚</text>
              <text class="status-text">{{ getDeliveryStatusText(order.delivery_status) }}</text>
            </view>
          </view>
          <text class="order-time">{{ formatTime(order.created_at) }}</text>
        </view>
      </view>

      <!-- 空状态 -->
      <view v-if="!loading && orderList.length === 0" class="empty-state">
        <text class="empty-state__icon">📦</text>
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

    <!-- 筛选抽屉 -->
    <view v-if="showFilter" class="filter-drawer">
      <view class="filter-mask" @tap="closeFilter"></view>
      <view class="filter-content">
        <view class="filter-header">
          <text class="filter-title">筛选条件</text>
          <view class="filter-close" @tap="closeFilter">
            <text class="close-icon">✕</text>
          </view>
        </view>

        <!-- 订单状态 -->
        <view class="filter-section">
          <text class="filter-section-title">订单状态</text>
          <view class="filter-options">
            <view
              v-for="(item, idx) in orderStatusOptions"
              :key="idx"
              class="filter-option"
              :class="{ selected: tempFilters.orderStatus === item.value }"
              @tap="tempFilters.orderStatus = item.value"
            >
              <text>{{ item.label }}</text>
            </view>
          </view>
        </view>

        <!-- 回款状态 -->
        <view class="filter-section">
          <text class="filter-section-title">回款状态</text>
          <view class="filter-options">
            <view
              v-for="(item, idx) in paymentStatusOptions"
              :key="idx"
              class="filter-option"
              :class="{ selected: tempFilters.paymentStatus === item.value }"
              @tap="tempFilters.paymentStatus = item.value"
            >
              <text>{{ item.label }}</text>
            </view>
          </view>
        </view>

        <!-- 配送状态 -->
        <view class="filter-section">
          <text class="filter-section-title">配送状态</text>
          <view class="filter-options">
            <view
              v-for="(item, idx) in deliveryStatusOptions"
              :key="idx"
              class="filter-option"
              :class="{ selected: tempFilters.deliveryStatus === item.value }"
              @tap="tempFilters.deliveryStatus = item.value"
            >
              <text>{{ item.label }}</text>
            </view>
          </view>
        </view>

        <!-- 底部按钮 -->
        <view class="filter-footer">
          <view class="filter-btn-reset" @tap="resetFilter">
            <text>重置</text>
          </view>
          <view class="filter-btn-confirm" @tap="applyFilter">
            <text>确定</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 底部占位 -->
    <view style="height: 120rpx;"></view>

    <CustomTabBar :current="1" />
  </view>
</template>

<script>
import { ref, reactive, computed, onMounted } from 'vue'
import { useUserStore } from '../../store/user'
import { getOrders } from '../../api/order'
import CustomTabBar from '../../components/CustomTabBar.vue'

export default {
  components: { CustomTabBar },
  setup() {
    const userStore = useUserStore()

    // 一级标签（精简为4个）
    const primaryTabs = ref([
      { text: '全部', status: null, highlight: false, count: 0 },
      { text: '待处理', status: [0], highlight: true, count: 0 },
      { text: '进行中', status: [1], highlight: false, count: 0 },
      { text: '已结束', status: [2, 3, 4], highlight: false, count: 0 }
    ])
    const currentTab = ref(0)

    // 筛选相关
    const showFilter = ref(false)
    const filters = reactive({
      orderStatus: null,
      paymentStatus: null,
      deliveryStatus: null
    })
    const tempFilters = reactive({
      orderStatus: null,
      paymentStatus: null,
      deliveryStatus: null
    })

    // 筛选选项
    const orderStatusOptions = [
      { label: '全部', value: null },
      { label: '待审批', value: 0 },
      { label: '已生效', value: 1 },
      { label: '已驳回', value: 2 },
      { label: '已取消', value: 3 },
      { label: '已退货', value: 4 }
    ]
    const paymentStatusOptions = [
      { label: '全部', value: null },
      { label: '未回款', value: 0 },
      { label: '部分回款', value: 1 },
      { label: '已回款', value: 2 }
    ]
    const deliveryStatusOptions = [
      { label: '全部', value: null },
      { label: '未配送', value: 0 },
      { label: '配送中', value: 1 },
      { label: '已配送', value: 2 }
    ]

    // 是否有激活的筛选条件
    const hasActiveFilter = computed(() => {
      return filters.orderStatus !== null || 
             filters.paymentStatus !== null || 
             filters.deliveryStatus !== null
    })

    // 订单列表
    const orderList = ref([])
    const loading = ref(false)
    const page = ref(1)
    const total = ref(0)
    const hasMore = ref(true)
    const keyword = ref('')

    // 切换标签
    const switchTab = (index) => {
      currentTab.value = index
      page.value = 1
      orderList.value = []
      hasMore.value = true
      loadOrders()
    }

    // 打开筛选抽屉
    const openFilter = () => {
      tempFilters.orderStatus = filters.orderStatus
      tempFilters.paymentStatus = filters.paymentStatus
      tempFilters.deliveryStatus = filters.deliveryStatus
      showFilter.value = true
    }

    // 关闭筛选抽屉
    const closeFilter = () => {
      showFilter.value = false
    }

    // 重置筛选
    const resetFilter = () => {
      tempFilters.orderStatus = null
      tempFilters.paymentStatus = null
      tempFilters.deliveryStatus = null
    }

    // 应用筛选
    const applyFilter = () => {
      filters.orderStatus = tempFilters.orderStatus
      filters.paymentStatus = tempFilters.paymentStatus
      filters.deliveryStatus = tempFilters.deliveryStatus
      showFilter.value = false
      page.value = 1
      orderList.value = []
      hasMore.value = true
      loadOrders()
    }

    // 格式化时间
    const formatTime = (timeStr) => {
      if (!timeStr) return '--'
      const date = new Date(timeStr)
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      const hour = String(date.getHours()).padStart(2, '0')
      const minute = String(date.getMinutes()).padStart(2, '0')
      const second = String(date.getSeconds()).padStart(2, '0')
      return `${year}-${month}-${day} ${hour}:${minute}:${second}`
    }

    // 加载订单
    const loadOrders = async () => {
      if (loading.value) return
      loading.value = true
      try {
        const params = {
          page: page.value,
          page_size: 10,
          keyword: keyword.value
          // 移除 salesman_id，由后端根据用户角色自动判断
        }

        // 一级标签状态筛选
        const tabStatus = primaryTabs.value[currentTab.value].status
        if (tabStatus && tabStatus.length > 0) {
          if (filters.orderStatus !== null) {
            // 如果抽屉中也选了订单状态，优先使用抽屉的
            params.order_status = filters.orderStatus
          } else if (tabStatus.length === 1) {
            params.order_status = tabStatus[0]
          } else {
            // 多个状态用逗号分隔（后端需支持）
            params.order_status = tabStatus.join(',')
          }
        } else if (filters.orderStatus !== null) {
          params.order_status = filters.orderStatus
        }

        // 回款状态筛选
        if (filters.paymentStatus !== null) {
          params.payment_status = filters.paymentStatus
        }

        // 配送状态筛选
        if (filters.deliveryStatus !== null) {
          params.delivery_status = filters.deliveryStatus
        }

        const res = await getOrders(params)
        const list = res.data?.records || res.data?.list || []
        if (page.value === 1) {
          orderList.value = list
        } else {
          orderList.value = [...orderList.value, ...list]
        }
        total.value = res.data?.total || 0
        hasMore.value = orderList.value.length < total.value
      } catch (e) {
        console.error('加载订单失败:', e)
      } finally {
        loading.value = false
      }
    }

    // 状态文本映射
    const getStatusText = (status) => {
      const map = { 0: '待审批', 1: '已生效', 2: '已驳回', 3: '已取消', 4: '已退货' }
      return map[status] || '未知'
    }

    const getStatusClass = (status) => {
      const map = { 0: 'tag-warning', 1: 'tag-success', 2: 'tag-danger', 3: 'tag-default', 4: 'tag-purple' }
      return map[status] || ''
    }

    const getPaymentStatusText = (status) => {
      const map = { 0: '未回款', 1: '部分回款', 2: '已回款' }
      return map[status] || '未回款'
    }

    const getPaymentStatusClass = (status) => {
      const map = { 0: 'status-danger', 1: 'status-warning', 2: 'status-success' }
      return map[status] || 'status-default'
    }

    const getDeliveryStatusText = (status) => {
      const map = { 0: '未配送', 1: '配送中', 2: '已配送' }
      return map[status] || '未配送'
    }

    const getDeliveryStatusClass = (status) => {
      const map = { 0: 'status-default', 1: 'status-primary', 2: 'status-success' }
      return map[status] || 'status-default'
    }

    // 页面跳转
    const goDetail = (id) => {
      uni.navigateTo({ url: `/pages/orders/detail?id=${id}` })
    }

    const goCreate = () => {
      uni.navigateTo({ url: '/pages/orders/create' })
    }

    // 生命周期
    onMounted(() => {
      loadOrders()
    })

    // 返回所有需要的数据和方法
    return {
      primaryTabs,
      currentTab,
      showFilter,
      filters,
      tempFilters,
      orderStatusOptions,
      paymentStatusOptions,
      deliveryStatusOptions,
      hasActiveFilter,
      orderList,
      loading,
      hasMore,
      keyword,
      switchTab,
      openFilter,
      closeFilter,
      resetFilter,
      applyFilter,
      loadOrders,
      getStatusText,
      getStatusClass,
      getPaymentStatusText,
      getPaymentStatusClass,
      getDeliveryStatusText,
      getDeliveryStatusClass,
      formatTime,
      goDetail,
      goCreate
    }
  },
  onReachBottom() {
    if (this.hasMore && !this.loading) {
      this.page = this.page + 1
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

/* 标签栏 */
.tabs-container {
  display: flex;
  align-items: center;
  background-color: #ffffff;
  border-bottom: 1rpx solid #eeeeee;
  padding-right: 16rpx;
}

.tabs {
  flex: 1;
  display: flex;
}

.tab-item {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24rpx 0;
  position: relative;
}

.tab-text {
  font-size: 28rpx;
  color: #666666;
}

/* 统一选中标签样式 */
.tab-item.active {
  background-color: #e6f7ff;
}

.tab-item.active .tab-text {
  color: #1890ff;
  font-weight: 600;
}

/* 高亮标签（待处理）选中时特殊样式 */
.tab-item.tab-highlight.active {
  background-color: #fff1f0;
}

.tab-item.tab-highlight.active .tab-text {
  color: #ff4d4f;
}

.tab-badge {
  position: absolute;
  top: 12rpx;
  right: 8rpx;
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
}

/* 筛选按钮 */
.filter-btn {
  padding: 12rpx 24rpx;
  background: #f5f5f5;
  border-radius: 24rpx;
  display: flex;
  align-items: center;
}

.filter-icon {
  font-size: 24rpx;
  color: #666666;
}

.filter-btn.filter-active {
  background: #e6f7ff;
}

.filter-btn.filter-active .filter-icon {
  color: #1890ff;
}

/* 订单列表 */
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

/* 底部状态行 */
.order-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 16rpx;
  border-top: 1rpx solid #eeeeee;
}

.order-status-row {
  display: flex;
  align-items: center;
  gap: 24rpx;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 4rpx;
}

.status-icon {
  font-size: 24rpx;
}

.status-text {
  font-size: 22rpx;
}

.status-danger .status-text { color: #ff4d4f; }
.status-warning .status-text { color: #fa8c16; }
.status-success .status-text { color: #52c41a; }
.status-primary .status-text { color: #1890ff; }
.status-default .status-text { color: #999999; }

.order-time {
  font-size: 22rpx;
  color: #cccccc;
}

/* 筛选抽屉 */
.filter-drawer {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 1000;
}

.filter-mask {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
}

.filter-content {
  position: absolute;
  right: 0;
  top: 0;
  bottom: 0;
  width: 600rpx;
  background: #ffffff;
  display: flex;
  flex-direction: column;
}

.filter-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 32rpx;
  border-bottom: 1rpx solid #f0f0f0;
}

.filter-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #333333;
}

.filter-close {
  padding: 8rpx;
}

.close-icon {
  font-size: 32rpx;
  color: #999999;
}

.filter-section {
  padding: 24rpx 32rpx;
  border-bottom: 1rpx solid #f0f0f0;
}

.filter-section-title {
  font-size: 28rpx;
  color: #333333;
  font-weight: 500;
  margin-bottom: 20rpx;
}

.filter-options {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}

.filter-option {
  padding: 12rpx 24rpx;
  background: #f5f5f5;
  border-radius: 8rpx;
  font-size: 26rpx;
  color: #666666;
}

.filter-option.selected {
  background: #e6f7ff;
  color: #1890ff;
}

.filter-footer {
  display: flex;
  gap: 24rpx;
  padding: 32rpx;
  margin-top: auto;
}

.filter-btn-reset,
.filter-btn-confirm {
  flex: 1;
  height: 80rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12rpx;
  font-size: 28rpx;
}

.filter-btn-reset {
  background: #f5f5f5;
  color: #666666;
}

.filter-btn-confirm {
  background: #1890ff;
  color: #ffffff;
}

/* 加载状态 */
.loading-more {
  display: flex;
  justify-content: center;
  padding: 30rpx 0;
}

.loading-text {
  font-size: 24rpx;
  color: #999999;
}

/* 新建按钮 */
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

/* 标签样式 */
.tag-default {
  background-color: #f5f5f5;
  color: #999999;
}

.tag-purple {
  background-color: #f9f0ff;
  color: #722ed1;
}
</style>
