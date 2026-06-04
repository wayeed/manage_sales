<template>
  <view class="gs" v-if="visible">
    <view class="gs-mask" @tap="close"></view>
    <view class="gs-sheet">
      <!-- 顶部拖拽指示条 -->
      <view class="gs-handle-bar">
        <view class="gs-handle"></view>
      </view>

      <!-- 标题栏 -->
      <view class="gs-header">
        <text class="gs-title">选择礼品</text>
        <view class="gs-close" @tap="close">
          <text class="gs-close-icon">✕</text>
        </view>
      </view>

      <!-- 搜索栏 -->
      <view class="gs-search">
        <view class="gs-search-box">
          <text class="gs-search-icon">⌕</text>
          <input
            v-model="keyword"
            class="gs-search-input"
            placeholder="搜索礼品名称 / 编码"
            placeholder-class="gs-search-ph"
            confirm-type="search"
            @confirm="handleSearch"
          />
          <view v-if="keyword" class="gs-search-clear" @tap="keyword = ''; handleSearch()">
            <text class="gs-clear-icon">✕</text>
          </view>
        </view>
      </view>

      <!-- 礼品列表 -->
      <scroll-view class="gs-list" scroll-y @scrolltolower="loadMore">
        <!-- 加载中 -->
        <view v-if="loading && list.length === 0" class="gs-state">
          <view class="gs-spinner"></view>
          <text class="gs-state-text">加载中</text>
        </view>

        <!-- 空状态 -->
        <view v-else-if="list.length === 0" class="gs-state">
          <text class="gs-state-icon">🎁</text>
          <text class="gs-state-text">暂无匹配礼品</text>
          <text class="gs-state-hint">试试其他关键词</text>
        </view>

        <!-- 礼品条目 -->
        <view v-else class="gs-items">
          <view
            v-for="item in list"
            :key="item.id"
            class="gs-item"
            :class="{
              'gs-item--active': selectedId === item.id,
              'gs-item--disabled': isDisabled(item)
            }"
            @tap="selectGift(item)"
          >
            <!-- 礼品图片 -->
            <view class="gs-item-thumb">
              <image
                v-if="item.gift_image"
                :src="item.gift_image"
                class="gs-thumb-img"
                mode="aspectFill"
              />
              <view v-else class="gs-thumb-placeholder">
                <text class="gs-thumb-text">礼</text>
              </view>
            </view>

            <!-- 礼品信息 -->
            <view class="gs-item-body">
              <text class="gs-item-name">{{ item.gift_name || '未命名礼品' }}</text>
              <view class="gs-item-meta">
                <text class="gs-item-code">{{ item.gift_code }}</text>
                <text class="gs-item-dot">·</text>
                <text v-if="isOutOfStock(item)" class="gs-item-stock gs-item-stock--out">库存不足</text>
                <text v-else class="gs-item-stock" :class="{ 'gs-item-stock--low': item.stock_quantity <= item.warning_stock }">
                  库存 {{ item.stock_quantity || 0 }}
                </text>
              </view>
              <view class="gs-item-price">
                <text class="gs-price-label">成本价</text>
                <text class="gs-price-value">¥{{ formatPrice(item.cost_price) }}</text>
              </view>
            </view>

            <!-- 选中标记 -->
            <view v-if="selectedId === item.id" class="gs-item-check">
              <text class="gs-check-icon">✓</text>
            </view>
          </view>
        </view>

        <!-- 加载更多 -->
        <view v-if="loading && list.length > 0" class="gs-loading-more">
          <view class="gs-spinner gs-spinner--small"></view>
          <text class="gs-loading-text">加载中...</text>
        </view>
        <view v-else-if="!hasMore && list.length > 0" class="gs-no-more">
          <text class="gs-no-more-text">没有更多了</text>
        </view>
      </scroll-view>

      <!-- 底部按钮 -->
      <view class="gs-footer">
        <view class="gs-cancel" @tap="close">
          <text class="gs-cancel-text">取消</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { ref, watch, onMounted } from 'vue'
import { getGiftList } from '../../api/gift'

export default {
  props: {
    visible: {
      type: Boolean,
      default: false
    },
    excludeIds: {
      type: Array,
      default: () => []
    }
  },
  emits: ['close', 'select'],
  setup(props, { emit }) {
    const keyword = ref('')
    const list = ref([])
    const loading = ref(false)
    const page = ref(1)
    const hasMore = ref(true)
    const selectedId = ref(null)
    
    // 监听 visible 变化
    watch(() => props.visible, (val) => {
      if (val) {
        keyword.value = ''
        selectedId.value = null
        loadData(true)
      }
    })

    // 格式化价格
    const formatPrice = (price) => {
      if (!price && price !== 0) return '0.00'
      const num = parseFloat(price)
      if (isNaN(num)) return '0.00'
      return num.toFixed(2)
    }

    // 是否禁用（已选择或库存不足）
    const isDisabled = (item) => {
      // 已选择的礼品
      if (props.excludeIds.includes(item.id)) {
        return true
      }
      // 库存为0或不足的礼品
      if (!item.stock_quantity || item.stock_quantity <= 0) {
        return true
      }
      return false
    }

    // 检查是否库存不足
    const isOutOfStock = (item) => {
      return !item.stock_quantity || item.stock_quantity <= 0
    }

    // 加载数据
    const loadData = async (reset = false) => {
      if (loading.value) return
      if (!reset && !hasMore.value) return

      loading.value = true
      if (reset) {
        page.value = 1
        list.value = []
        hasMore.value = true
      }

      try {
        const res = await getGiftList({
          page: page.value,
          page_size: 10,
          keyword: keyword.value,
          status: 1  // 只查询启用状态的礼品
        })

        const records = res.data?.records || res.data?.list || []
        const total = res.data?.total || 0

        if (reset) {
          list.value = records
        } else {
          list.value = [...list.value, ...records]
        }

        hasMore.value = list.value.length < total
        if (records.length > 0) {
          page.value++
        }
      } catch (e) {
        console.error('加载礼品列表失败:', e)
        uni.showToast({ title: '加载失败', icon: 'none' })
      } finally {
        loading.value = false
      }
    }

    // 搜索
    const handleSearch = () => {
      loadData(true)
    }

    // 加载更多
    const loadMore = () => {
      loadData(false)
    }

    // 选择礼品
    const selectGift = (item) => {
      if (isOutOfStock(item)) {
        uni.showToast({ title: '库存不足，无法选择', icon: 'none' })
        return
      }
      if (isDisabled(item)) {
        uni.showToast({ title: '该礼品已选择', icon: 'none' })
        return
      }
      selectedId.value = item.id
      emit('select', item)
    }

    // 关闭
    const close = () => {
      selectedId.value = null
      emit('close')
    }

    return {
      keyword,
      list,
      loading,
      hasMore,
      selectedId,
      formatPrice,
      isDisabled,
      isOutOfStock,
      handleSearch,
      loadMore,
      selectGift,
      close
    }
  }
}
</script>

<style lang="scss" scoped>
.gs {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 1000;
}

.gs-mask {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
}

.gs-sheet {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 70vh;
  background-color: #ffffff;
  border-radius: 24rpx 24rpx 0 0;
  display: flex;
  flex-direction: column;
}

/* 拖拽指示条 */
.gs-handle-bar {
  display: flex;
  justify-content: center;
  padding: 12rpx 0;
}

.gs-handle {
  width: 36rpx;
  height: 4rpx;
  background-color: #e5e5e5;
  border-radius: 2rpx;
}

/* 标题栏 */
.gs-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16rpx 24rpx;
  border-bottom: 1rpx solid #f5f5f5;
}

.gs-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.gs-close {
  width: 64rpx;
  height: 64rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.gs-close-icon {
  font-size: 32rpx;
  color: #999999;
}

/* 搜索栏 */
.gs-search {
  padding: 20rpx 24rpx;
  border-bottom: 1rpx solid #f5f5f5;
}

.gs-search-box {
  display: flex;
  align-items: center;
  height: 72rpx;
  background-color: #f5f5f5;
  border-radius: 36rpx;
  padding: 0 24rpx;
}

.gs-search-icon {
  font-size: 32rpx;
  color: #999999;
  margin-right: 12rpx;
}

.gs-search-input {
  flex: 1;
  height: 72rpx;
  font-size: 28rpx;
  color: #333333;
}

.gs-search-ph {
  color: #999999;
}

.gs-search-clear {
  width: 48rpx;
  height: 48rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.gs-clear-icon {
  font-size: 24rpx;
  color: #999999;
}

/* 列表区域 */
.gs-list {
  flex: 1;
  overflow-y: auto;
}

/* 状态提示 */
.gs-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 120rpx 40rpx;
}

.gs-spinner {
  width: 48rpx;
  height: 48rpx;
  border: 4rpx solid #f3f3f3;
  border-top-color: #1890ff;
  border-radius: 50%;
  animation: gs-spin 1s linear infinite;
}

.gs-spinner--small {
  width: 32rpx;
  height: 32rpx;
  border-width: 3rpx;
}

@keyframes gs-spin {
  to {
    transform: rotate(360deg);
  }
}

.gs-state-icon {
  font-size: 80rpx;
  margin-bottom: 20rpx;
}

.gs-state-text {
  font-size: 28rpx;
  color: #666666;
  margin-bottom: 8rpx;
}

.gs-state-hint {
  font-size: 24rpx;
  color: #999999;
}

/* 礼品条目 */
.gs-items {
  padding: 8rpx 0;
}

.gs-item {
  display: flex;
  align-items: center;
  padding: 24rpx;
  margin: 0 24rpx;
  border-radius: 12rpx;
  background-color: #ffffff;
  margin-bottom: 16rpx;
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.04);

  &:active {
    background-color: #f9f9f9;
  }

  &--active {
    background-color: #e6f7ff;
    border: 2rpx solid #1890ff;
  }

  &--disabled {
    opacity: 0.5;
    pointer-events: none;
  }
}

.gs-item-thumb {
  width: 120rpx;
  height: 120rpx;
  border-radius: 12rpx;
  overflow: hidden;
  margin-right: 24rpx;
  background-color: #f5f5f5;
  flex-shrink: 0;
}

.gs-thumb-img {
  width: 100%;
  height: 100%;
}

.gs-thumb-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.gs-thumb-text {
  font-size: 48rpx;
  color: #cccccc;
}

.gs-item-body {
  flex: 1;
  min-width: 0;
}

.gs-item-name {
  font-size: 30rpx;
  font-weight: 500;
  color: #1a1a1a;
  margin-bottom: 8rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.gs-item-meta {
  display: flex;
  align-items: center;
  margin-bottom: 8rpx;
}

.gs-item-code {
  font-size: 24rpx;
  color: #999999;
}

.gs-item-dot {
  font-size: 24rpx;
  color: #cccccc;
  margin: 0 8rpx;
}

.gs-item-stock {
  font-size: 24rpx;
  color: #52c41a;

  &--low {
    color: #fa8c16;
  }

  &--out {
    color: #ff4d4f;
  }
}

.gs-item-price {
  display: flex;
  align-items: center;
}

.gs-price-label {
  font-size: 24rpx;
  color: #999999;
  margin-right: 8rpx;
}

.gs-price-value {
  font-size: 28rpx;
  font-weight: 600;
  color: #ff4d4f;
}

.gs-item-check {
  width: 48rpx;
  height: 48rpx;
  border-radius: 50%;
  background-color: #1890ff;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-left: 16rpx;
}

.gs-check-icon {
  font-size: 28rpx;
  color: #ffffff;
}

/* 加载更多 */
.gs-loading-more,
.gs-no-more {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 30rpx 0;
}

.gs-loading-text,
.gs-no-more-text {
  font-size: 24rpx;
  color: #999999;
  margin-left: 12rpx;
}

/* 底部按钮 */
.gs-footer {
  padding: 20rpx 24rpx;
  padding-bottom: calc(20rpx + env(safe-area-inset-bottom));
  border-top: 1rpx solid #f5f5f5;
}

.gs-cancel {
  height: 88rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f5f5f5;
  border-radius: 44rpx;

  &:active {
    background-color: #e8e8e8;
  }
}

.gs-cancel-text {
  font-size: 30rpx;
  color: #666666;
}
</style>
