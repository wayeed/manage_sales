<template>
  <view class="ps" v-if="visible">
    <view class="ps-mask" @tap="close"></view>
    <view class="ps-sheet">
      <!-- 顶部拖拽指示条 -->
      <view class="ps-handle-bar">
        <view class="ps-handle"></view>
      </view>

      <!-- 标题栏 -->
      <view class="ps-header">
        <text class="ps-title">选择商品</text>
        <view class="ps-close" @tap="close">
          <text class="ps-close-icon">✕</text>
        </view>
      </view>

      <!-- 搜索栏 -->
      <view class="ps-search">
        <view class="ps-search-box">
          <text class="ps-search-icon">⌕</text>
          <input
            v-model="keyword"
            class="ps-search-input"
            placeholder="搜索商品名称 / SKU编码"
            placeholder-class="ps-search-ph"
            confirm-type="search"
            @confirm="handleSearch"
          />
          <view v-if="keyword" class="ps-search-clear" @tap="keyword = ''; handleSearch()">
            <text class="ps-clear-icon">✕</text>
          </view>
        </view>
      </view>

      <!-- 商品列表 -->
      <scroll-view class="ps-list" scroll-y @scrolltolower="loadMore">
        <!-- 加载中 -->
        <view v-if="loading && list.length === 0" class="ps-state">
          <view class="ps-spinner"></view>
          <text class="ps-state-text">加载中</text>
        </view>

        <!-- 空状态 -->
        <view v-else-if="list.length === 0" class="ps-state">
          <text class="ps-state-icon">📦</text>
          <text class="ps-state-text">暂无匹配商品</text>
          <text class="ps-state-hint">试试其他关键词</text>
        </view>

        <!-- 商品条目 -->
        <view v-else class="ps-items">
          <view
            v-for="item in list"
            :key="item.id"
            class="ps-item"
            :class="{
              'ps-item--active': selectedId === item.id,
              'ps-item--disabled': isDisabled(item)
            }"
            @tap="selectProduct(item)"
          >
            <!-- 商品图片 -->
            <view class="ps-item-thumb">
              <image
                v-if="item.product?.product_image"
                :src="item.product.product_image"
                class="ps-thumb-img"
                mode="aspectFill"
              />
              <view v-else class="ps-thumb-placeholder">
                <text class="ps-thumb-text">品</text>
              </view>
            </view>

            <!-- 商品信息 -->
            <view class="ps-item-body">
              <text class="ps-item-name">{{ item.sku_name || item.product?.product_name || '未命名' }}</text>
              <view class="ps-item-meta">
                <text class="ps-item-sku">{{ item.sku_code }}</text>
                <text v-if="item.barcode" class="ps-item-barcode">{{ item.barcode }}</text>
              </view>
              <view class="ps-item-bottom">
                <text class="ps-item-price">¥{{ item.product?.list_price || '0.00' }}</text>
                <text class="ps-item-stock" :class="{ 'ps-item-stock--empty': (item.available_stock || 0) <= 0 }">
                  库存: {{ item.available_stock || 0 }}
                </text>
              </view>
            </view>

            <!-- 选中/禁用指示 -->
            <view v-if="isDisabled(item)" class="ps-item-badge">
              <text class="ps-badge-text">已选</text>
            </view>
            <view v-else class="ps-item-check" :class="{ 'ps-item-check--on': selectedId === item.id }">
              <text v-if="selectedId === item.id" class="ps-check-icon">✓</text>
            </view>
          </view>

          <!-- 加载更多 -->
          <view v-if="loading && list.length > 0" class="ps-more">
            <view class="ps-spinner ps-spinner--sm"></view>
            <text class="ps-more-text">加载更多</text>
          </view>
          <view v-else-if="!hasMore && list.length > 0" class="ps-more">
            <text class="ps-more-text">已加载全部</text>
          </view>
        </view>
      </scroll-view>

      <!-- 底部操作栏 -->
      <view class="ps-footer">
        <view v-if="selectedProduct" class="ps-footer-info">
          <text class="ps-footer-label">已选</text>
          <text class="ps-footer-name">{{ selectedProduct.sku_name || selectedProduct.product?.product_name }}</text>
          <text class="ps-footer-stock">库存: {{ selectedProduct.available_stock || 0 }}</text>
        </view>
        <view v-else class="ps-footer-info">
          <text class="ps-footer-hint">请选择一个商品</text>
        </view>
        <button
          class="ps-btn"
          :class="{ 'ps-btn--off': !selectedId }"
          @tap="confirm"
        >确定选择</button>
      </view>
    </view>
  </view>
</template>

<script>
import { ref, watch } from 'vue'
import { getSkuListWithStock } from '../../api/product'
import { useUserStore } from '../../store/user'

export default {
  name: 'ProductSelect',
  props: {
    visible: {
      type: Boolean,
      default: false
    },
    // 已选SKU ID列表（用于禁止重复选择）
    excludeIds: {
      type: Array,
      default: () => []
    }
  },
  emits: ['close', 'select'],
  setup(props, { emit }) {
    const userStore = useUserStore()
    const keyword = ref('')
    const list = ref([])
    const loading = ref(false)
    const selectedId = ref(null)
    const selectedProduct = ref(null)
    const page = ref(1)
    const pageSize = 20
    const hasMore = ref(true)

    watch(() => props.visible, (val) => {
      if (val) {
        reset()
        fetchList()
      }
    })

    const reset = () => {
      keyword.value = ''
      list.value = []
      selectedId.value = null
      selectedProduct.value = null
      page.value = 1
      hasMore.value = true
    }

    const fetchList = async () => {
      if (loading.value || !hasMore.value) return
      loading.value = true
      try {
        const storeId = userStore.userInfo?.store_id || 0
        const res = await getSkuListWithStock({
          store_id: storeId,
          keyword: keyword.value,
          page: page.value,
          page_size: pageSize
        })
        let data = res.data?.list || res.data || []
        // 过滤无库存商品
        data = data.filter(item => (item.available_stock || 0) > 0)
        if (page.value === 1) {
          list.value = data
        } else {
          list.value = [...list.value, ...data]
        }
        hasMore.value = data.length === pageSize
      } catch (e) {
        console.error('获取商品列表失败:', e)
        uni.showToast({ title: '获取商品列表失败', icon: 'none' })
      } finally {
        loading.value = false
      }
    }

    const handleSearch = () => {
      page.value = 1
      hasMore.value = true
      fetchList()
    }

    const loadMore = () => {
      if (!loading.value && hasMore.value) {
        page.value++
        fetchList()
      }
    }

    const isDisabled = (item) => {
      return props.excludeIds.includes(item.id)
    }

    const selectProduct = (item) => {
      if (isDisabled(item)) {
        uni.showToast({ title: '该商品已添加', icon: 'none' })
        return
      }
      selectedId.value = item.id
      selectedProduct.value = item
    }

    const close = () => {
      emit('close')
    }

    const confirm = () => {
      if (!selectedProduct.value) return
      emit('select', selectedProduct.value)
      close()
    }

    return {
      keyword,
      list,
      loading,
      selectedId,
      selectedProduct,
      handleSearch,
      loadMore,
      selectProduct,
      isDisabled,
      close,
      confirm
    }
  }
}
</script>

<style lang="scss" scoped>
/* ===== 弹窗容器 ===== */
.ps {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 1000;
}

.ps-mask {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
}

.ps-sheet {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 75vh;
  background: #fff;
  border-radius: 28rpx 28rpx 0 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* ===== 拖拽指示条 ===== */
.ps-handle-bar {
  display: flex;
  justify-content: center;
  padding: 18rpx 0 0;
}

.ps-handle {
  width: 64rpx;
  height: 8rpx;
  border-radius: 4rpx;
  background: #ddd;
}

/* ===== 标题栏 ===== */
.ps-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20rpx 32rpx 16rpx;
}

.ps-title {
  font-size: 34rpx;
  font-weight: 600;
  color: #1a1a1a;
  letter-spacing: 0.5rpx;
}

.ps-close {
  width: 64rpx;
  height: 64rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: #f5f5f5;
}

.ps-close-icon {
  font-size: 28rpx;
  color: #888;
}

/* ===== 搜索栏 ===== */
.ps-search {
  padding: 8rpx 32rpx 16rpx;
}

.ps-search-box {
  display: flex;
  align-items: center;
  height: 80rpx;
  background: #f7f7f7;
  border-radius: 16rpx;
  padding: 0 24rpx;
  gap: 12rpx;
}

.ps-search-icon {
  font-size: 32rpx;
  color: #bbb;
  flex-shrink: 0;
}

.ps-search-input {
  flex: 1;
  height: 80rpx;
  font-size: 28rpx;
  color: #333;
}

.ps-search-ph {
  color: #bbb;
}

.ps-search-clear {
  width: 40rpx;
  height: 40rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #ddd;
  border-radius: 50%;
  flex-shrink: 0;
}

.ps-clear-icon {
  font-size: 20rpx;
  color: #fff;
}

/* ===== 列表区域 ===== */
.ps-list {
  flex: 1;
  min-height: 0;
}

/* ===== 状态 ===== */
.ps-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80rpx 0;
  gap: 16rpx;
}

.ps-state-icon {
  font-size: 64rpx;
}

.ps-state-text {
  font-size: 28rpx;
  color: #999;
}

.ps-state-hint {
  font-size: 24rpx;
  color: #ccc;
}

/* ===== 加载动画 ===== */
.ps-spinner {
  width: 40rpx;
  height: 40rpx;
  border: 4rpx solid #e0e0e0;
  border-top-color: #1890ff;
  border-radius: 50%;
  animation: ps-spin 0.6s linear infinite;
}

.ps-spinner--sm {
  width: 28rpx;
  height: 28rpx;
  border-width: 3rpx;
}

@keyframes ps-spin {
  to { transform: rotate(360deg); }
}

/* ===== 商品条目 ===== */
.ps-items {
  padding: 0 24rpx;
}

.ps-item {
  display: flex;
  align-items: center;
  padding: 24rpx 16rpx;
  margin-bottom: 2rpx;
  border-radius: 16rpx;
  transition: background 0.15s;

  &--active {
    background: #f0f7ff;
  }

  &--disabled {
    opacity: 0.45;
    pointer-events: none;
  }
}

/* 缩略图 */
.ps-item-thumb {
  width: 120rpx;
  height: 120rpx;
  border-radius: 12rpx;
  overflow: hidden;
  flex-shrink: 0;
  margin-right: 24rpx;
  background: #f5f5f5;
}

.ps-thumb-img {
  width: 100%;
  height: 100%;
}

.ps-thumb-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #e8e8e8, #f5f5f5);
}

.ps-thumb-text {
  font-size: 36rpx;
  color: #ccc;
  font-weight: 500;
}

/* 信息区 */
.ps-item-body {
  flex: 1;
  min-width: 0;
}

.ps-item-name {
  font-size: 30rpx;
  font-weight: 500;
  color: #1a1a1a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: block;
  margin-bottom: 8rpx;
}

.ps-item-meta {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-bottom: 10rpx;
}

.ps-item-sku {
  font-size: 24rpx;
  color: #999;
  background: #f5f5f5;
  padding: 2rpx 12rpx;
  border-radius: 6rpx;
}

.ps-item-barcode {
  font-size: 24rpx;
  color: #bbb;
}

.ps-item-bottom {
  display: flex;
  align-items: baseline;
  gap: 16rpx;
}

.ps-item-price {
  font-size: 32rpx;
  font-weight: 600;
  color: #e8453c;
}

.ps-item-stock {
  font-size: 24rpx;
  color: #52c41a;
  font-weight: 500;

  &--empty {
    color: #ff4d4f;
  }
}

/* 已选标记 */
.ps-item-badge {
  flex-shrink: 0;
  margin-left: 16rpx;
  padding: 4rpx 16rpx;
  border-radius: 8rpx;
  background: #f0f0f0;
}

.ps-badge-text {
  font-size: 22rpx;
  color: #999;
}

/* 选中勾 */
.ps-item-check {
  width: 48rpx;
  height: 48rpx;
  border-radius: 50%;
  border: 3rpx solid #ddd;
  flex-shrink: 0;
  margin-left: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;

  &--on {
    background: #1890ff;
    border-color: #1890ff;
  }
}

.ps-check-icon {
  font-size: 26rpx;
  color: #fff;
  font-weight: 600;
}

/* ===== 加载更多 ===== */
.ps-more {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  padding: 32rpx 0;
}

.ps-more-text {
  font-size: 24rpx;
  color: #ccc;
}

/* ===== 底部操作栏 ===== */
.ps-footer {
  display: flex;
  align-items: center;
  padding: 20rpx 32rpx;
  padding-bottom: calc(20rpx + env(safe-area-inset-bottom));
  gap: 24rpx;
  border-top: 1rpx solid #f0f0f0;
  background: #fff;
}

.ps-footer-info {
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

.ps-footer-label {
  font-size: 22rpx;
  color: #999;
  margin-right: 8rpx;
}

.ps-footer-name {
  font-size: 26rpx;
  color: #333;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ps-footer-stock {
  font-size: 22rpx;
  color: #52c41a;
  margin-left: 12rpx;
}

.ps-footer-hint {
  font-size: 26rpx;
  color: #ccc;
}

.ps-btn {
  flex-shrink: 0;
  height: 88rpx;
  line-height: 88rpx;
  padding: 0 56rpx;
  border-radius: 16rpx;
  font-size: 30rpx;
  font-weight: 500;
  background: #1890ff;
  color: #fff;
  letter-spacing: 1rpx;

  &--off {
    opacity: 0.35;
  }
}
</style>
