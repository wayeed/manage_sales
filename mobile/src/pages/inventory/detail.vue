<template>
  <view class="detail-page">
    <view v-if="item" class="detail-content">
      <!-- 商品基本信息 -->
      <view class="info-section card">
        <text class="section-title">商品信息</text>
        <view class="info-row">
          <text class="info-label">商品名称</text>
          <text class="info-value">{{ skuName || '--' }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">SKU编号</text>
          <text class="info-value">{{ skuCode || '--' }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">所属仓库</text>
          <text class="info-value">{{ warehouseName || '--' }}</text>
        </view>
      </view>

      <!-- 库存信息 -->
      <view class="info-section card">
        <text class="section-title">库存信息</text>
        <view class="stock-grid">
          <view class="stock-item">
            <text class="stock-value" :class="getStockValueClass(item.stock_quantity)">{{ item.stock_quantity || 0 }}</text>
            <text class="stock-label">总库存</text>
          </view>
          <view class="stock-item">
            <text class="stock-value text-success">{{ item.available_quantity || 0 }}</text>
            <text class="stock-label">可用库存</text>
          </view>
          <view class="stock-item">
            <text class="stock-value text-warning">{{ item.locked_quantity || 0 }}</text>
            <text class="stock-label">锁定库存</text>
          </view>
        </view>
        <view class="info-row mt-4">
          <text class="info-label">库存状态</text>
          <view class="tag" :class="getStockClass(item.available_quantity)">
            <text>{{ getStockText(item.available_quantity) }}</text>
          </view>
        </view>
        <view class="info-row">
          <text class="info-label">预警值</text>
          <text class="info-value">{{ item.warning_stock || 10 }}</text>
        </view>
      </view>

      <!-- 商品规格 -->
      <view v-if="skuAttrs.length > 0" class="info-section card">
        <text class="section-title">商品规格</text>
        <view v-for="(attr, idx) in skuAttrs" :key="idx" class="info-row">
          <text class="info-label">{{ attr.name }}</text>
          <text class="info-value">{{ attr.value }}</text>
        </view>
      </view>
    </view>

    <!-- 加载中 -->
    <view v-else class="loading-state">
      <text class="loading-text">加载中...</text>
    </view>
  </view>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { get } from '../../api/request'

export default {
  setup() {
    const item = ref(null)

    // 计算属性：SKU名称
    const skuName = computed(() => {
      if (!item.value?.sku) return '--'
      return item.value.sku.sku_name || item.value.sku.product?.product_name || '--'
    })

    // 计算属性：SKU编号
    const skuCode = computed(() => {
      return item.value?.sku?.sku_code || '--'
    })

    // 计算属性：仓库名称
    const warehouseName = computed(() => {
      return item.value?.warehouse?.warehouse_name || '--'
    })

    // 计算属性：SKU属性列表
    const skuAttrs = computed(() => {
      const attrs = item.value?.sku?.attributes
      if (!attrs) return []
      try {
        const obj = typeof attrs === 'string' ? JSON.parse(attrs) : attrs
        return Object.entries(obj).map(([name, value]) => ({ name, value }))
      } catch {
        return []
      }
    })

    const getStockText = (stock) => {
      if (stock === undefined || stock === null) return '未知'
      if (stock <= 0) return '缺货'
      const warningStock = item.value?.warning_stock || 10
      if (stock <= warningStock) return '库存不足'
      return '有货'
    }

    const getStockClass = (stock) => {
      if (stock === undefined || stock === null) return 'tag-warning'
      if (stock <= 0) return 'tag-danger'
      const warningStock = item.value?.warning_stock || 10
      if (stock <= warningStock) return 'tag-warning'
      return 'tag-success'
    }

    const getStockValueClass = (stock) => {
      if (stock === undefined || stock === null) return ''
      if (stock <= 0) return 'text-danger'
      const warningStock = item.value?.warning_stock || 10
      if (stock <= warningStock) return 'text-warning'
      return 'text-primary'
    }

    onMounted(() => {
      const pages = getCurrentPages()
      const currentPage = pages[pages.length - 1]
      const id = currentPage.$page?.options?.id || currentPage.options?.id
      if (id) {
        loadDetail(id)
      }
    })

    const loadDetail = async (id) => {
      try {
        const res = await get(`/inventory/${id}`)
        item.value = res.data
      } catch (e) {
        console.error('加载库存详情失败:', e)
      }
    }

    return {
      item,
      skuName,
      skuCode,
      warehouseName,
      skuAttrs,
      getStockText,
      getStockClass,
      getStockValueClass
    }
  }
}
</script>

<style lang="scss" scoped>
.detail-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding: 24rpx;
}

.detail-content {
  padding-bottom: 40rpx;
}

.card {
  background-color: #ffffff;
  border-radius: 16rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
}

.info-section {
  padding: 32rpx;
  margin-bottom: 24rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 24rpx;
  padding-bottom: 16rpx;
  border-bottom: 2rpx solid #f0f0f0;
}

.info-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16rpx 0;
}

.info-label {
  font-size: 28rpx;
  color: #666666;
}

.info-value {
  font-size: 28rpx;
  color: #1a1a1a;
  font-weight: 500;
}

// 库存网格展示
.stock-grid {
  display: flex;
  justify-content: space-around;
  padding: 24rpx 0;
  border-bottom: 2rpx solid #f5f5f5;
}

.stock-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
}

.stock-value {
  font-size: 48rpx;
  font-weight: 700;
  color: #1a1a1a;
  margin-bottom: 8rpx;
}

.stock-label {
  font-size: 24rpx;
  color: #999999;
}

// 标签样式
.tag {
  padding: 8rpx 20rpx;
  border-radius: 8rpx;
  font-size: 24rpx;

  text {
    font-weight: 500;
  }
}

.tag-success {
  background-color: #e6f7ed;
  color: #52c41a;
}

.tag-warning {
  background-color: #fff7e6;
  color: #fa8c16;
}

.tag-danger {
  background-color: #fff1f0;
  color: #f5222d;
}

// 文本颜色
.text-primary {
  color: #1890ff;
}

.text-success {
  color: #52c41a;
}

.text-warning {
  color: #fa8c16;
}

.text-danger {
  color: #f5222d;
}

// 间距
.mt-4 {
  margin-top: 32rpx;
}

// 加载状态
.loading-state {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 60vh;
}

.loading-text {
  font-size: 28rpx;
  color: #999999;
}
</style>
