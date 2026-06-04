<template>
  <view class="follow-up-page">
    <!-- 搜索栏 -->
    <view class="search-bar">
      <view class="search-input-wrap">
        <text class="search-icon">🔍</text>
        <input
          v-model="keyword"
          class="search-input"
          placeholder="搜索客户姓名/手机号"
          confirm-type="search"
          @confirm="handleSearch"
        />
        <text v-if="keyword" class="search-clear" @tap="keyword = ''; loadData()">✕</text>
      </view>
    </view>

    <!-- 客户列表 -->
    <scroll-view
      scroll-y
      class="customer-list"
      @scrolltolower="loadMore"
    >
      <view
        v-for="customer in customerList"
        :key="customer.id"
        class="customer-card"
        @tap="goToDetail(customer)"
      >
        <view class="cc-main">
          <view class="cc-left">
            <view class="cc-row">
              <text class="cc-name">{{ customer.customer_name }}</text>
              <text class="cc-phone">{{ customer.phone || '-' }}</text>
            </view>
            <view class="cc-row">
              <text class="cc-address">{{ customer.address || '暂无地址' }}</text>
            </view>
            <view class="cc-row">
              <text class="cc-first-visit">首次进店：{{ formatDate(customer.created_at) }}</text>
            </view>
          </view>
          <view v-if="customer.has_draft" class="cc-right">
            <view class="cc-badge">
              <text class="cc-badge-text">待跟进</text>
            </view>
            <text class="cc-draft-count">{{ customer.draft_items }}件商品</text>
          </view>
        </view>
      </view>

      <!-- 空状态 -->
      <view v-if="!loading && customerList.length === 0" class="empty-state">
        <text class="empty-text">暂无客户数据</text>
      </view>

      <!-- 加载更多 -->
      <view v-if="loading" class="loading-more">
        <text class="loading-text">加载中...</text>
      </view>
      <view v-else-if="noMore && customerList.length > 0" class="loading-more">
        <text class="loading-text">没有更多了</text>
      </view>
    </scroll-view>
    <CustomTabBar :current="2" />
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import CustomTabBar from '../../components/CustomTabBar.vue'
import { getCustomersWithDraftStatus } from '../../api/customer'

const keyword = ref('')
const customerList = ref([])
const loading = ref(false)
const noMore = ref(false)
const page = ref(1)
const pageSize = 20

const loadData = async () => {
  if (loading.value) return
  loading.value = true
  try {
    const res = await getCustomersWithDraftStatus({
      keyword: keyword.value,
      page: page.value,
      page_size: pageSize
    })
    const list = res.data?.list || res.data || []
    const total = res.data?.total || 0

    if (page.value === 1) {
      customerList.value = list
    } else {
      customerList.value = [...customerList.value, ...list]
    }
    noMore.value = customerList.value.length >= total
  } catch (e) {
    console.error('加载客户列表失败:', e)
  } finally {
    loading.value = false
  }
}

const loadMore = () => {
  if (noMore.value || loading.value) return
  page.value++
  loadData()
}

const handleSearch = () => {
  page.value = 1
  noMore.value = false
  loadData()
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

const goToDetail = (customer) => {
  uni.navigateTo({
    url: `/pages/customer/detail?id=${customer.id}&name=${encodeURIComponent(customer.customer_name)}&phone=${customer.phone || ''}&address=${encodeURIComponent(customer.address || '')}&has_draft=${customer.has_draft ? 1 : 0}&draft_id=${customer.draft_id || ''}`
  })
}

// 页面显示时刷新数据
onShow(() => {
  page.value = 1
  noMore.value = false
  loadData()
})
</script>

<style lang="scss" scoped>
.follow-up-page {
  height: 100vh;
  background: #f5f5f5;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-sizing: border-box;
}

.search-bar {
  padding: 16rpx 24rpx;
  background: #fff;
  flex-shrink: 0;
  z-index: 10;
}

.search-input-wrap {
  display: flex;
  align-items: center;
  background: #f5f5f5;
  border-radius: 12rpx;
  padding: 0 20rpx;
  height: 72rpx;
}

.search-icon {
  font-size: 28rpx;
  margin-right: 12rpx;
}

.search-input {
  flex: 1;
  height: 72rpx;
  font-size: 28rpx;
  color: #333;
}

.search-clear {
  font-size: 28rpx;
  color: #999;
  padding: 8rpx;
}

.customer-list {
  flex: 1;
  padding: 16rpx 24rpx;
  width: 100%;
  box-sizing: border-box;
  overflow-x: hidden;
}

.customer-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
  &:active { background: #fafafa; }
}

.cc-main {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16rpx;
}

.cc-left {
  flex: 1;
  min-width: 0;
}

.cc-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-bottom: 8rpx;
  &:last-child { margin-bottom: 0; }
}

.cc-name {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 200rpx;
}

.cc-phone {
  font-size: 26rpx;
  color: #666;
  flex-shrink: 0;
}

.cc-address {
  font-size: 24rpx;
  color: #999;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.cc-first-visit {
  font-size: 22rpx;
  color: #bbb;
}

.cc-right {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8rpx;
  flex-shrink: 0;
}

.cc-badge {
  background: linear-gradient(135deg, #ff6b35, #ff4d4f);
  padding: 6rpx 20rpx;
  border-radius: 24rpx;
}

.cc-badge-text {
  font-size: 22rpx;
  color: #fff;
  font-weight: 600;
}

.cc-draft-count {
  font-size: 20rpx;
  color: #ff6b35;
}

.empty-state {
  padding: 120rpx 0;
  text-align: center;
}

.empty-text {
  font-size: 28rpx;
  color: #999;
}

.loading-more {
  padding: 24rpx 0;
  text-align: center;
}

.loading-text {
  font-size: 24rpx;
  color: #999;
}
</style>
