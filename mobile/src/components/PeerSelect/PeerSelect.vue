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
        <text class="ps-title">选择同行</text>
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
            placeholder="搜索同行姓名 / 手机号"
            placeholder-class="ps-search-ph"
            confirm-type="search"
            @confirm="handleSearch"
          />
        </view>
      </view>

      <!-- 同行列表 -->
      <scroll-view class="ps-list" scroll-y @scrolltolower="loadMore">
        <view v-if="loading && list.length === 0" class="ps-state">
          <view class="ps-spinner"></view>
          <text class="ps-state-text">加载中</text>
        </view>
        <view v-else-if="list.length === 0" class="ps-state">
          <text class="ps-state-icon">👥</text>
          <text class="ps-state-text">暂无同行信息</text>
          <text class="ps-state-hint" @tap="showAddDialog">点击手动添加同行</text>
        </view>
        <view v-else class="ps-items">
          <view
            v-for="item in list"
            :key="item.id"
            class="ps-item"
            :class="{ 'ps-item--active': selectedId === item.id }"
            @tap="selectPeer(item)"
          >
            <view class="ps-item-body">
              <text class="ps-item-name">{{ item.peer_name }}</text>
              <text class="ps-item-phone">{{ item.phone || '--' }}</text>
            </view>
            <view class="ps-item-check" :class="{ 'ps-item-check--on': selectedId === item.id }">
              <text v-if="selectedId === item.id" class="ps-check-icon">✓</text>
            </view>
          </view>
        </view>
      </scroll-view>

      <!-- 底部操作栏 -->
      <view class="ps-footer">
        <button class="ps-btn ps-btn--add" @tap="showAddDialog">手动添加</button>
        <button class="ps-btn" :class="{ 'ps-btn--off': !selectedId }" @tap="confirm">
          确定选择
        </button>
      </view>
    </view>

    <!-- 手动添加弹窗 -->
    <view v-if="addDialogVisible" class="add-dialog">
      <view class="add-mask" @tap="addDialogVisible = false"></view>
      <view class="add-content">
        <view class="add-header">
          <text class="add-title">添加同行</text>
        </view>
        <view class="add-body">
          <view class="add-field">
            <text class="add-label">姓名 *</text>
            <input v-model="addForm.peer_name" class="add-input" placeholder="请输入同行姓名" />
          </view>
          <view class="add-field">
            <text class="add-label">手机号</text>
            <input v-model="addForm.phone" class="add-input" type="tel" placeholder="请输入手机号" />
          </view>
        </view>
        <view class="add-footer">
          <button class="add-btn add-btn--cancel" @tap="addDialogVisible = false">取消</button>
          <button class="add-btn add-btn--confirm" :disabled="addLoading" @tap="handleAddPeer">
            {{ addLoading ? '保存中...' : '保存' }}
          </button>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { ref, reactive, watch } from 'vue'
import { getPeerList, createPeer } from '../../api/peer'

export default {
  name: 'PeerSelect',
  props: {
    visible: {
      type: Boolean,
      default: false
    }
  },
  emits: ['close', 'select'],
  setup(props, { emit }) {
    const keyword = ref('')
    const list = ref([])
    const loading = ref(false)
    const selectedId = ref(null)
    const selectedPeer = ref(null)
    const page = ref(1)
    const hasMore = ref(true)

    // 手动添加
    const addDialogVisible = ref(false)
    const addLoading = ref(false)
    const addForm = reactive({
      peer_name: '',
      phone: ''
    })

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
      selectedPeer.value = null
      page.value = 1
      hasMore.value = true
    }

    const fetchList = async () => {
      if (loading.value || !hasMore.value) return
      loading.value = true
      try {
        const res = await getPeerList({
          keyword: keyword.value,
          page: page.value,
          page_size: 20
        })
        const data = res.data?.list || res.data?.records || []
        if (page.value === 1) {
          list.value = data
        } else {
          list.value = [...list.value, ...data]
        }
        hasMore.value = data.length === 20
      } catch (e) {
        console.error('获取同行列表失败:', e)
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

    const selectPeer = (item) => {
      selectedId.value = item.id
      selectedPeer.value = item
    }

    const close = () => emit('close')

    const confirm = () => {
      if (!selectedPeer.value) return
      emit('select', selectedPeer.value)
      close()
    }

    // 显示添加弹窗
    const showAddDialog = () => {
      addForm.peer_name = keyword.value || ''
      addForm.phone = ''
      addDialogVisible.value = true
    }

    // 添加同行
    const handleAddPeer = async () => {
      if (!addForm.peer_name.trim()) {
        uni.showToast({ title: '请输入同行姓名', icon: 'none' })
        return
      }

      // 去重检查
      const exists = list.value.find(
        item => item.peer_name === addForm.peer_name.trim()
      )
      if (exists) {
        uni.showToast({ title: '该同行已存在', icon: 'none' })
        selectedId.value = exists.id
        selectedPeer.value = exists
        addDialogVisible.value = false
        return
      }

      addLoading.value = true
      try {
        const res = await createPeer({
          peer_name: addForm.peer_name.trim(),
          phone: addForm.phone.trim() || undefined
        })
        uni.showToast({ title: '添加成功', icon: 'success' })
        addDialogVisible.value = false

        // 刷新列表并选中新添加的同行
        page.value = 1
        hasMore.value = true
        await fetchList()

        // 选中新添加的同行
        if (res.data?.id) {
          selectedId.value = res.data.id
          selectedPeer.value = res.data
        }
      } catch (e) {
        console.error('添加同行失败:', e)
        uni.showToast({ title: '添加失败', icon: 'none' })
      } finally {
        addLoading.value = false
      }
    }

    return {
      keyword,
      list,
      loading,
      selectedId,
      addDialogVisible,
      addLoading,
      addForm,
      handleSearch,
      loadMore,
      selectPeer,
      close,
      confirm,
      showAddDialog,
      handleAddPeer
    }
  }
}
</script>

<style lang="scss" scoped>
/* 弹窗容器 */
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

.ps-list {
  flex: 1;
  min-height: 0;
}

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
  font-size: 26rpx;
  color: #1890ff;
  margin-top: 16rpx;
}

.ps-spinner {
  width: 40rpx;
  height: 40rpx;
  border: 4rpx solid #e0e0e0;
  border-top-color: #1890ff;
  border-radius: 50%;
  animation: ps-spin 0.6s linear infinite;
}

@keyframes ps-spin {
  to { transform: rotate(360deg); }
}

.ps-items {
  padding: 0 24rpx;
}

.ps-item {
  display: flex;
  align-items: center;
  padding: 24rpx 16rpx;
  margin-bottom: 2rpx;
  border-radius: 16rpx;
  background: #fafafa;

  &--active {
    background: #e6f7ff;
  }
}

.ps-item-body {
  flex: 1;
  min-width: 0;
}

.ps-item-name {
  font-size: 30rpx;
  font-weight: 500;
  color: #1a1a1a;
  display: block;
  margin-bottom: 8rpx;
}

.ps-item-phone {
  font-size: 26rpx;
  color: #666;
}

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

  &--on {
    background: #1890ff;
    border-color: #1890ff;
  }
}

.ps-check-icon {
  font-size: 26rpx;
  color: #fff;
}

.ps-footer {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 20rpx 32rpx;
  padding-bottom: calc(20rpx + env(safe-area-inset-bottom));
  border-top: 1rpx solid #f0f0f0;
  background: #fff;
}

.ps-btn {
  flex: 1;
  height: 88rpx;
  border-radius: 16rpx;
  font-size: 30rpx;
  font-weight: 500;
  background: #1890ff;
  color: #fff;
  text-align: center;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;

  &--off {
    opacity: 0.35;
  }

  &--add {
    flex: 0 0 180rpx;
    background: #f5f5f5;
    color: #666;
  }
}

/* 手动添加弹窗 */
.add-dialog {
  position: fixed;
  inset: 0;
  z-index: 1001;
}

.add-mask {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
}

.add-content {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  width: 600rpx;
  background: #fff;
  border-radius: 24rpx;
  overflow: hidden;
}

.add-header {
  padding: 32rpx;
  text-align: center;
  border-bottom: 1rpx solid #f0f0f0;
}

.add-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.add-body {
  padding: 32rpx;
}

.add-field {
  margin-bottom: 24rpx;
}

.add-label {
  display: block;
  font-size: 26rpx;
  color: #666;
  margin-bottom: 12rpx;
}

.add-input {
  width: 100%;
  height: 80rpx;
  padding: 0 24rpx;
  background: #f7f7f7;
  border-radius: 12rpx;
  font-size: 28rpx;
  color: #333;
  box-sizing: border-box;
}

.add-footer {
  display: flex;
  border-top: 1rpx solid #f0f0f0;
}

.add-btn {
  flex: 1;
  height: 96rpx;
  line-height: 96rpx;
  text-align: center;
  font-size: 30rpx;
  background: none;
  border: none;
  border-radius: 0;

  &--cancel {
    color: #666;
    border-right: 1rpx solid #f0f0f0;
  }

  &--confirm {
    color: #1890ff;
    font-weight: 500;
  }

  &[disabled] {
    color: #bbb;
  }
}
</style>
