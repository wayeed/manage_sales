<template>
  <view class="inventory-page">
    <!-- 搜索栏 -->
    <view class="search-bar">
      <view class="search-input">
        <text class="search-icon">&#x1F50D;</text>
        <input
          v-model="keyword"
          class="search-field"
          placeholder="搜索商品名称/SKU"
          placeholder-class="search-placeholder"
          confirm-type="search"
          @confirm="handleSearch"
        />
        <text v-if="keyword" class="search-clear" @tap="clearSearch">x</text>
      </view>
    </view>

    <!-- 库存列表 -->
    <view class="inventory-list">
      <view
        v-for="(item, index) in inventoryList"
        :key="item.id || index"
        class="inventory-card card"
        :class="{ 'inventory-card--warning': isLowStock(item) }"
        @tap="goDetail(item.id)"
      >
        <view class="inventory-header">
          <text class="inventory-name">{{ item.name || '--' }}</text>
          <view class="tag" :class="getStockClass(item)">
            <text>{{ getStockText(item) }}</text>
          </view>
        </view>
        <view class="inventory-body">
          <view class="inventory-row">
            <text class="inventory-label">SKU</text>
            <text class="inventory-value">{{ item.sku || '--' }}</text>
          </view>
          <view class="inventory-row">
            <text class="inventory-label">总库存</text>
            <text class="inventory-value">{{ item.totalStock || 0 }}</text>
          </view>
          <view class="inventory-row">
            <text class="inventory-label">可用库存</text>
            <text class="inventory-value" :class="{ 'inventory-value--danger': isLowStock(item) }">
              {{ item.availableStock || 0 }}
            </text>
          </view>
        </view>
      </view>

      <!-- 空状态 -->
      <view v-if="!loading && inventoryList.length === 0" class="empty-state">
        <text class="empty-state__icon">&#x1F4E6;</text>
        <text class="empty-state__text">暂无库存数据</text>
      </view>

      <!-- 加载更多 -->
      <view v-if="loading" class="loading-more">
        <text class="loading-text">加载中...</text>
      </view>
      <view v-else-if="!hasMore && inventoryList.length > 0" class="loading-more">
        <text class="loading-text">没有更多了</text>
      </view>
    </view>

    <!-- 底部占位 -->
    <view style="height: 120rpx;"></view>

    <CustomTabBar :current="3" />
  </view>
</template>

<script>
import { ref } from 'vue'
import { getStockList } from '../../api/inventory'
import CustomTabBar from '../../components/CustomTabBar.vue'

export default {
  components: { CustomTabBar },
  data() {
    return {
      keyword: '',
      inventoryList: [],
      loading: false,
      page: 1,
      hasMore: true
    }
  },
  methods: {
    clearSearch() {
      this.keyword = ''
      this.page = 1
      this.inventoryList = []
      this.hasMore = true
      this.loadInventory()
    },

    handleSearch() {
      this.page = 1
      this.inventoryList = []
      this.hasMore = true
      this.loadInventory()
    },

    async loadInventory() {
      if (this.loading) return
      this.loading = true
      try {
        const params = {
          page: this.page,
          page_size: 10,
          keyword: this.keyword
        }
        const res = await getStockList(params)
        const list = res.data?.records || res.data?.list || []
        // 后端字段映射到前端显示字段
        const mappedList = list.map(item => ({
          id: item.id,
          name: item.sku?.sku_name || item.sku?.product?.product_name || '--',
          sku: item.sku?.sku_code || '--',
          totalStock: item.stock_quantity || 0,
          availableStock: item.available_quantity || 0,
          lockedQuantity: item.locked_quantity || 0,
          warningStock: item.warning_stock || 10,
          warehouseName: item.warehouse?.warehouse_name || '--'
        }))
        if (this.page === 1) {
          this.inventoryList = mappedList
        } else {
          this.inventoryList = [...this.inventoryList, ...mappedList]
        }
        const total = res.data?.total || 0
        this.hasMore = this.inventoryList.length < total
      } catch (e) {
        console.error('加载库存失败:', e)
        this.inventoryList = []
      } finally {
        this.loading = false
      }
    },

    // 库存预警判断：可用库存 <= 预警值
    isLowStock(item) {
      const available = item.availableStock || 0
      const warning = item.warningStock || 10
      return available <= warning
    },

    getStockText(item) {
      if (this.isLowStock(item)) {
        const available = item.availableStock || 0
        if (available <= 0) return '缺货'
        return '库存预警'
      }
      return '正常'
    },

    getStockClass(item) {
      if (this.isLowStock(item)) {
        const available = item.availableStock || 0
        if (available <= 0) return 'tag-danger'
        return 'tag-warning'
      }
      return 'tag-success'
    },

    goDetail(id) {
      uni.navigateTo({ url: `/pages/inventory/detail?id=${id}` })
    }
  },
  onReachBottom() {
    if (this.hasMore && !this.loading) {
      this.page++
      this.loadInventory()
    }
  },
  onPullDownRefresh() {
    this.page = 1
    this.inventoryList = []
    this.hasMore = true
    this.loadInventory().then(() => {
      uni.stopPullDownRefresh()
    })
  },
  onShow() {
    this.page = 1
    this.inventoryList = []
    this.hasMore = true
    this.loadInventory()
  }
}
</script>

<style lang="scss" scoped>
.inventory-page {
  min-height: 100vh;
  background-color: #f5f5f5;
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

.inventory-list {
  padding: 24rpx;
}

.inventory-card {
  &:active {
    opacity: 0.9;
  }

  &--warning {
    border-left: 6rpx solid #ff4d4f;
  }
}

.inventory-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16rpx;
}

.inventory-name {
  font-size: 30rpx;
  font-weight: 500;
  color: #333333;
}

.inventory-body {
  margin-bottom: 8rpx;
}

.inventory-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8rpx 0;
}

.inventory-label {
  font-size: 26rpx;
  color: #999999;
}

.inventory-value {
  font-size: 28rpx;
  color: #333333;

  &--danger {
    color: #ff4d4f;
    font-weight: bold;
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
