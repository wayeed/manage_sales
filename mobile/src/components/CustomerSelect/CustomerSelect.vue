<template>
  <view class="cs" v-if="visible">
    <view class="cs-mask" @tap="close"></view>
    <view class="cs-sheet">
      <!-- 顶部拖拽指示条 -->
      <view class="cs-handle-bar">
        <view class="cs-handle"></view>
      </view>

      <!-- 标题栏 -->
      <view class="cs-header">
        <text class="cs-title">选择客户</text>
        <view class="cs-close" @tap="close">
          <text class="cs-close-icon">✕</text>
        </view>
      </view>

      <!-- 搜索栏 -->
      <view class="cs-search">
        <view class="cs-search-box">
          <text class="cs-search-icon">⌕</text>
          <input
            v-model="keyword"
            class="cs-search-input"
            placeholder="搜索手机号 / 客户姓名"
            placeholder-class="cs-search-ph"
            confirm-type="search"
            @confirm="handleSearch"
          />
          <view v-if="keyword" class="cs-search-clear" @tap="keyword = ''; handleSearch()">
            <text class="cs-clear-icon">✕</text>
          </view>
        </view>
      </view>

      <!-- 客户列表 -->
      <scroll-view class="cs-list" scroll-y @scrolltolower="loadMore">
        <!-- 加载中 -->
        <view v-if="loading && list.length === 0" class="cs-state">
          <view class="cs-spinner"></view>
          <text class="cs-state-text">加载中</text>
        </view>

        <!-- 空状态 -->
        <view v-else-if="list.length === 0" class="cs-state">
          <text class="cs-state-icon">👤</text>
          <text class="cs-state-text">暂无匹配客户</text>
          <text class="cs-state-hint">试试其他关键词或新增客户</text>
        </view>

        <!-- 客户条目 -->
        <view v-else class="cs-items">
          <view
            v-for="item in list"
            :key="item.id"
            class="cs-item"
            :class="{ 'cs-item--disabled': !canSelect(item) }"
          >
            <!-- 客户信息 -->
            <view class="cs-item-body" @tap="canSelect(item) && selectCustomer(item)">
              <text class="cs-item-name">{{ item.customer_name || item.name || '未命名' }}</text>
              <view class="cs-item-meta">
                <text class="cs-item-phone">{{ item.phone || '--' }}</text>
                <text class="cs-item-salesman">业务员: {{ item.salesman_name || '无' }}</text>
              </view>
            </view>

            <!-- 操作按钮 -->
            <view class="cs-item-actions">
              <view
                v-if="canSelect(item)"
                class="cs-btn cs-btn--primary"
                @tap="selectCustomer(item)"
              >
                <text class="cs-btn-text">选择</text>
              </view>
              <view
                v-else
                class="cs-btn cs-btn--warning"
                @tap="applyFollowUp(item)"
              >
                <text class="cs-btn-text">申请跟进</text>
              </view>
            </view>
          </view>

          <!-- 加载更多 -->
          <view v-if="loading && list.length > 0" class="cs-more">
            <view class="cs-spinner cs-spinner--sm"></view>
            <text class="cs-more-text">加载更多</text>
          </view>
          <view v-else-if="!hasMore && list.length > 0" class="cs-more">
            <text class="cs-more-text">已加载全部</text>
          </view>
        </view>
      </scroll-view>

      <!-- 底部操作栏 -->
      <view class="cs-footer">
        <button class="cs-btn cs-btn--add" @tap="showAddCustomer">
          <text class="cs-btn-icon">+</text>
          <text class="cs-btn-text">新增客户</text>
        </button>
      </view>
    </view>

    <!-- 新增客户弹窗 -->
    <view v-if="showAddForm" class="cs-add-form">
      <view class="cs-add-mask" @tap="showAddForm = false"></view>
      <view class="cs-add-sheet">
        <view class="cs-add-header">
          <text class="cs-add-title">新增客户</text>
          <view class="cs-close" @tap="showAddForm = false">
            <text class="cs-close-icon">✕</text>
          </view>
        </view>
        <view class="cs-add-body">
          <view class="cs-add-field">
            <text class="cs-add-label">客户姓名 <text class="required">*</text></text>
            <input v-model="newCustomer.name" class="cs-add-input" placeholder="请输入客户姓名" />
          </view>
          <view class="cs-add-field">
            <text class="cs-add-label">手机号 <text class="required">*</text></text>
            <input v-model="newCustomer.phone" class="cs-add-input" type="number" maxlength="11" placeholder="请输入手机号" />
          </view>
          <view class="cs-add-field">
            <text class="cs-add-label">地址</text>
            <input v-model="newCustomer.address" class="cs-add-input" placeholder="请输入地址（选填）" />
          </view>
          <view class="cs-add-field">
            <text class="cs-add-label">客户来源</text>
            <view class="cs-add-radio-group">
              <view
                v-for="item in sourceTypeOptions"
                :key="item.value"
                class="cs-add-radio"
                :class="{ 'cs-add-radio--active': newCustomer.source_type === item.value }"
                @tap="newCustomer.source_type = item.value"
              >
                <view class="cs-add-radio-dot">
                  <view v-if="newCustomer.source_type === item.value" class="cs-add-radio-inner"></view>
                </view>
                <text class="cs-add-radio-text">{{ item.label }}</text>
              </view>
            </view>
          </view>
        </view>
        <view class="cs-add-footer">
          <button class="cs-btn cs-btn--cancel" @tap="showAddForm = false">取消</button>
          <button class="cs-btn cs-btn--confirm" @tap="handleAddCustomer">确定</button>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { ref, watch } from 'vue'
import { getCustomerList, createCustomer } from '../../api/customer'
import { useUserStore } from '../../store/user'

export default {
  name: 'CustomerSelect',
  props: {
    visible: {
      type: Boolean,
      default: false
    }
  },
  emits: ['close', 'select', 'applyFollowUp'],
  setup(props, { emit }) {
    const keyword = ref('')
    const list = ref([])
    const loading = ref(false)
    const selectedCustomer = ref(null)
    const page = ref(1)
    const pageSize = 20
    const hasMore = ref(true)
    const userStore = useUserStore()

    // 新增客户
    const showAddForm = ref(false)
    const newCustomer = ref({
      name: '',
      phone: '',
      address: '',
      source_type: 0
    })
    const sourceTypeOptions = [
      { value: 0, label: '自然进店' },
      { value: 1, label: '主动邀约' },
      { value: 2, label: '同行带单' }
    ]

    watch(() => props.visible, (val) => {
      if (val) {
        reset()
        fetchList()
      }
    })

    const reset = () => {
      keyword.value = ''
      list.value = []
      selectedCustomer.value = null
      page.value = 1
      hasMore.value = true
    }

    const fetchList = async () => {
      if (loading.value && !hasMore.value) return
      loading.value = true
      try {
        const params = {
          page: page.value,
          page_size: pageSize
        }
        // 无关键字时，传递当前用户ID过滤自己的客户
        if (!keyword.value) {
          params.salesman_id = userStore.userInfo?.id
        } else {
          // 有关键字时，传递关键字用于精确匹配
          params.keyword = keyword.value
        }
        const res = await getCustomerList(params)
        const data = res.data?.list || res.data?.records || []
        if (page.value === 1) {
          list.value = data
        } else {
          list.value = [...list.value, ...data]
        }
        hasMore.value = data.length === pageSize
      } catch (e) {
        console.error('获取客户列表失败:', e)
        uni.showToast({ title: '获取客户列表失败', icon: 'none' })
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

    // 判断是否可以选择该客户
    const canSelect = (customer) => {
      const currentUserId = userStore.userInfo?.id
      if (!currentUserId) return false
      // 比较时转换为字符串避免类型不匹配
      const cCreatedBy = customer.created_by ? String(customer.created_by) : ''
      const cSalesmanId = customer.salesman_id ? String(customer.salesman_id) : ''
      const curId = String(currentUserId)
      return cCreatedBy === curId || cSalesmanId === curId
    }

    const selectCustomer = (customer) => {
      selectedCustomer.value = customer
      emit('select', customer)
      close()
    }

    const applyFollowUp = (customer) => {
      emit('applyFollowUp', customer)
    }

    const close = () => {
      emit('close')
    }

    const showAddCustomer = () => {
      newCustomer.value = { name: '', phone: '', address: '' }
      showAddForm.value = true
    }

    const handleAddCustomer = async () => {
      if (!newCustomer.value.name) {
        uni.showToast({ title: '请输入客户姓名', icon: 'none' })
        return
      }
      if (!newCustomer.value.phone || !/^1\d{10}$/.test(newCustomer.value.phone)) {
        uni.showToast({ title: '请输入正确的手机号', icon: 'none' })
        return
      }

      try {
        const res = await createCustomer({
          customer_name: newCustomer.value.name,
          phone: newCustomer.value.phone,
          address: newCustomer.value.address,
          source_type: newCustomer.value.source_type
        })
        uni.showToast({ title: '新增成功', icon: 'success' })
        showAddForm.value = false
        // 刷新列表
        page.value = 1
        hasMore.value = true
        fetchList()
      } catch (e) {
        console.error('新增客户失败:', e)
        uni.showToast({ title: '新增客户失败', icon: 'none' })
      }
    }

    return {
      keyword,
      list,
      loading,
      selectedCustomer,
      showAddForm,
      newCustomer,
      sourceTypeOptions,
      handleSearch,
      loadMore,
      canSelect,
      selectCustomer,
      applyFollowUp,
      close,
      showAddCustomer,
      handleAddCustomer
    }
  }
}
</script>

<style lang="scss" scoped>
/* 弹窗容器 */
.cs {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 1000;
}

.cs-mask {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
}

.cs-sheet {
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

/* 拖拽指示条 */
.cs-handle-bar {
  display: flex;
  justify-content: center;
  padding: 18rpx 0 0;
}

.cs-handle {
  width: 64rpx;
  height: 8rpx;
  border-radius: 4rpx;
  background: #ddd;
}

/* 标题栏 */
.cs-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20rpx 32rpx 16rpx;
}

.cs-title {
  font-size: 34rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.cs-close {
  width: 64rpx;
  height: 64rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: #f5f5f5;
}

.cs-close-icon {
  font-size: 28rpx;
  color: #888;
}

/* 搜索栏 */
.cs-search {
  padding: 8rpx 32rpx 16rpx;
}

.cs-search-box {
  display: flex;
  align-items: center;
  height: 80rpx;
  background: #f7f7f7;
  border-radius: 16rpx;
  padding: 0 24rpx;
  gap: 12rpx;
}

.cs-search-icon {
  font-size: 32rpx;
  color: #bbb;
  flex-shrink: 0;
}

.cs-search-input {
  flex: 1;
  height: 80rpx;
  font-size: 28rpx;
  color: #333;
}

.cs-search-ph {
  color: #bbb;
}

.cs-search-clear {
  width: 40rpx;
  height: 40rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #ddd;
  border-radius: 50%;
  flex-shrink: 0;
}

.cs-clear-icon {
  font-size: 20rpx;
  color: #fff;
}

/* 列表区域 */
.cs-list {
  flex: 1;
  min-height: 0;
}

/* 状态 */
.cs-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80rpx 0;
  gap: 16rpx;
}

.cs-state-icon {
  font-size: 64rpx;
}

.cs-state-text {
  font-size: 28rpx;
  color: #999;
}

.cs-state-hint {
  font-size: 24rpx;
  color: #ccc;
}

/* 加载动画 */
.cs-spinner {
  width: 40rpx;
  height: 40rpx;
  border: 4rpx solid #e0e0e0;
  border-top-color: #1890ff;
  border-radius: 50%;
  animation: cs-spin 0.6s linear infinite;
}

.cs-spinner--sm {
  width: 28rpx;
  height: 28rpx;
  border-width: 3rpx;
}

@keyframes cs-spin {
  to { transform: rotate(360deg); }
}

/* 客户条目 */
.cs-items {
  padding: 0 24rpx;
}

.cs-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24rpx 16rpx;
  margin-bottom: 2rpx;
  border-radius: 16rpx;
  background: #fafafa;

  &--disabled {
    opacity: 0.7;
  }
}

.cs-item-body {
  flex: 1;
  min-width: 0;
}

.cs-item-name {
  font-size: 30rpx;
  font-weight: 500;
  color: #1a1a1a;
  display: block;
  margin-bottom: 8rpx;
}

.cs-item-meta {
  display: flex;
  flex-direction: column;
  gap: 4rpx;
}

.cs-item-phone {
  font-size: 26rpx;
  color: #666;
}

.cs-item-salesman {
  font-size: 24rpx;
  color: #999;
}

.cs-item-actions {
  flex-shrink: 0;
  margin-left: 16rpx;
}

/* 按钮样式 */
.cs-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 24rpx;
  height: 64rpx;
  border-radius: 12rpx;
  gap: 8rpx;

  &--primary {
    background: #1890ff;
  }

  &--warning {
    background: #ff9500;
  }

  &--add {
    flex: 1;
    background: #f5f5f5;
    border: 2rpx dashed #ddd;
  }

  &--cancel {
    flex: 1;
    background: #f5f5f5;
  }

  &--confirm {
    flex: 1;
    background: #1890ff;
  }
}

.cs-btn-text {
  font-size: 26rpx;
  color: #fff;

  .cs-btn--add &,
  .cs-btn--cancel & {
    color: #666;
  }
}

.cs-btn-icon {
  font-size: 32rpx;
  color: #1890ff;
}

/* 加载更多 */
.cs-more {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  padding: 32rpx 0;
}

.cs-more-text {
  font-size: 24rpx;
  color: #ccc;
}

/* 底部操作栏 */
.cs-footer {
  display: flex;
  align-items: center;
  padding: 20rpx 32rpx;
  padding-bottom: calc(20rpx + env(safe-area-inset-bottom));
  border-top: 1rpx solid #f0f0f0;
  background: #fff;
}

/* 新增客户弹窗 */
.cs-add-form {
  position: absolute;
  inset: 0;
  z-index: 100;
}

.cs-add-mask {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
}

.cs-add-sheet {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  background: #fff;
  border-radius: 28rpx 28rpx 0 0;
  padding-bottom: env(safe-area-inset-bottom);
}

.cs-add-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24rpx 32rpx;
  border-bottom: 1rpx solid #f0f0f0;
}

.cs-add-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.cs-add-body {
  padding: 24rpx 32rpx;
}

.cs-add-field {
  margin-bottom: 24rpx;
}

.cs-add-label {
  display: block;
  font-size: 28rpx;
  color: #333;
  margin-bottom: 12rpx;

  .required {
    color: #ff4d4f;
  }
}

.cs-add-input {
  width: 100%;
  height: 88rpx;
  background: #f7f7f7;
  border-radius: 12rpx;
  padding: 0 24rpx;
  font-size: 28rpx;
  color: #333;
}

/* 来源类型单选 */
.cs-add-radio-group {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}

.cs-add-radio {
  display: flex;
  align-items: center;
  gap: 10rpx;
  padding: 16rpx 24rpx;
  background: #f7f7f7;
  border-radius: 12rpx;
  border: 2rpx solid transparent;
  transition: all 0.2s;

  &--active {
    background: #e6f4ff;
    border-color: #1890ff;
  }
}

.cs-add-radio-dot {
  width: 36rpx;
  height: 36rpx;
  border-radius: 50%;
  border: 2rpx solid #ccc;
  display: flex;
  align-items: center;
  justify-content: center;

  .cs-add-radio--active & {
    border-color: #1890ff;
  }
}

.cs-add-radio-inner {
  width: 20rpx;
  height: 20rpx;
  border-radius: 50%;
  background: #1890ff;
}

.cs-add-radio-text {
  font-size: 28rpx;
  color: #333;

  .cs-add-radio--active & {
    color: #1890ff;
    font-weight: 500;
  }
}

.cs-add-footer {
  display: flex;
  gap: 24rpx;
  padding: 16rpx 32rpx 32rpx;
}
</style>
