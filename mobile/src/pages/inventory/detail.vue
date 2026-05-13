<template>
  <view class="detail-page">
    <view v-if="item" class="detail-content">
      <!-- 商品基本信息 -->
      <view class="info-section card">
        <text class="section-title">商品信息</text>
        <view class="info-row">
          <text class="info-label">商品名称</text>
          <text class="info-value">{{ item.name || '--' }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">商品编号</text>
          <text class="info-value">{{ item.code || '--' }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">商品分类</text>
          <text class="info-value">{{ item.category || '--' }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">规格型号</text>
          <text class="info-value">{{ item.spec || '--' }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">单价</text>
          <text class="info-value text-primary text-bold">{{ item.price || '--' }}元</text>
        </view>
      </view>

      <!-- 库存信息 -->
      <view class="info-section card">
        <text class="section-title">库存信息</text>
        <view class="info-row">
          <text class="info-label">当前库存</text>
          <text class="info-value" :class="{ 'text-danger': (item.stock || 0) <= 10 }">
            {{ item.stock || 0 }}
          </text>
        </view>
        <view class="info-row">
          <text class="info-label">库存状态</text>
          <view class="tag" :class="getStockClass(item.stock)">
            <text>{{ getStockText(item.stock) }}</text>
          </view>
        </view>
        <view class="info-row">
          <text class="info-label">仓库位置</text>
          <text class="info-value">{{ item.warehouse || '--' }}</text>
        </view>
      </view>

      <!-- 商品描述 -->
      <view v-if="item.description" class="info-section card">
        <text class="section-title">商品描述</text>
        <text class="desc-text">{{ item.description }}</text>
      </view>
    </view>

    <!-- 加载中 -->
    <view v-else class="loading-state">
      <text class="loading-text">加载中...</text>
    </view>
  </view>
</template>

<script>
import { ref, onMounted } from 'vue'
import { get } from '../../api/request'

export default {
  setup() {
    const item = ref(null)

    const getStockText = (stock) => {
      if (stock === undefined || stock === null) return '未知'
      if (stock <= 0) return '缺货'
      if (stock <= 10) return '库存不足'
      return '有货'
    }

    const getStockClass = (stock) => {
      if (stock === undefined || stock === null) return 'tag-warning'
      if (stock <= 0) return 'tag-danger'
      if (stock <= 10) return 'tag-warning'
      return 'tag-success'
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

    return { item, getStockText, getStockClass }
  }
}
</script>

<style lang="scss" scoped>
.detail-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding: 24rpx;
}

.info-section {
  padding: 30rpx;
  margin-bottom: 24rpx;
}

.section-title {
  font-size: 30rpx;
  font-weight: bold;
  color: #333333;
  margin-bottom: 20rpx;
}

.info-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12rpx 0;
}

.info-label {
  font-size: 26rpx;
  color: #999999;
}

.info-value {
  font-size: 28rpx;
  color: #333333;
}

.desc-text {
  font-size: 26rpx;
  color: #666666;
  line-height: 1.6;
}

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
