<template>
  <view class="approval-page">
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

    <!-- 审批列表 -->
    <view class="approval-list">
      <view
        v-for="(item, index) in approvalList"
        :key="item.id || index"
        class="approval-card card"
      >
        <view class="approval-header">
          <text class="approval-type">跟进申请</text>
          <view class="tag" :class="getStatusClass(item.status)">
            <text>{{ getStatusText(item.status) }}</text>
          </view>
        </view>

        <view class="approval-body">
          <view class="approval-row">
            <text class="approval-label">客户</text>
            <text class="approval-value">{{ item.customer?.customer_name || '--' }}</text>
          </view>
          <view class="approval-row">
            <text class="approval-label">手机号</text>
            <text class="approval-value">{{ item.customer?.phone || '--' }}</text>
          </view>
          <view class="approval-row">
            <text class="approval-label">申请人</text>
            <text class="approval-value">{{ item.applicant?.real_name || item.applicant?.username || '--' }}</text>
          </view>
          <view class="approval-row">
            <text class="approval-label">申请时间</text>
            <text class="approval-value">{{ formatTime(item.created_at) }}</text>
          </view>
          <view v-if="item.remark" class="approval-row">
            <text class="approval-label">备注</text>
            <text class="approval-value">{{ item.remark }}</text>
          </view>
          <view v-if="item.status === 2 && item.reject_reason" class="approval-row">
            <text class="approval-label">拒绝原因</text>
            <text class="approval-value reject-reason">{{ item.reject_reason }}</text>
          </view>
        </view>

        <!-- 操作按钮 -->
        <view v-if="item.status === 0" class="approval-actions">
          <button class="btn btn-reject" @tap="handleReject(item)">拒绝</button>
          <button class="btn btn-approve" @tap="handleApprove(item)">同意</button>
        </view>
      </view>

      <!-- 空状态 -->
      <view v-if="!loading && approvalList.length === 0" class="empty-state">
        <text class="empty-state__icon">&#x1F4E6;</text>
        <text class="empty-state__text">暂无{{ tabs[currentTab].text }}记录</text>
      </view>

      <!-- 加载更多 -->
      <view v-if="loading" class="loading-more">
        <text class="loading-text">加载中...</text>
      </view>
      <view v-else-if="!hasMore && approvalList.length > 0" class="loading-more">
        <text class="loading-text">没有更多了</text>
      </view>
    </view>

    <!-- 拒绝原因弹窗 -->
    <uni-popup ref="rejectPopup" type="dialog">
      <uni-popup-dialog
        mode="input"
        title="拒绝原因"
        placeholder="请输入拒绝原因"
        :value="rejectReason"
        @confirm="confirmReject"
        @close="cancelReject"
      />
    </uni-popup>

    <CustomTabBar :current="4" />
  </view>
</template>

<script>
import { ref } from 'vue'
import { getPendingApprovals, getMyApplications, approveFollowUp, rejectFollowUp } from '../../api/approval'
import CustomTabBar from '../../components/CustomTabBar.vue'

export default {
  components: { CustomTabBar },
  data() {
    return {
      tabs: [
        { text: '待审批', status: 0 },
        { text: '已通过', status: 1 },
        { text: '已拒绝', status: 2 }
      ],
      currentTab: 0,
      approvalList: [],
      loading: false,
      page: 1,
      total: 0,
      hasMore: true,
      rejectReason: '',
      currentRejectItem: null
    }
  },
  methods: {
    switchTab(index) {
      this.currentTab = index
      this.page = 1
      this.approvalList = []
      this.hasMore = true
      this.loadApprovals()
    },

    async loadApprovals() {
      if (this.loading) return
      this.loading = true
      try {
        const status = this.tabs[this.currentTab].status
        let res
        if (status === 0) {
          // 待审批 - 查询待我审批的列表
          res = await getPendingApprovals({ page: this.page, page_size: 10 })
        } else {
          // 已通过/已拒绝 - 查询我的申请列表中对应状态的
          res = await getMyApplications({ page: this.page, page_size: 10 })
        }
        const list = res.data?.list || res.data?.records || []
        // 过滤状态（后端可能返回所有状态）
        const filteredList = status === 0 ? list : list.filter(item => item.status === status)
        if (this.page === 1) {
          this.approvalList = filteredList
        } else {
          this.approvalList = [...this.approvalList, ...filteredList]
        }
        this.total = res.data?.total || 0
        this.hasMore = this.approvalList.length < this.total
      } catch (e) {
        console.error('加载审批列表失败:', e)
        uni.showToast({ title: '加载失败', icon: 'none' })
      } finally {
        this.loading = false
      }
    },

    formatTime(time) {
      if (!time) return '--'
      return time.replace('T', ' ').substring(0, 16)
    },

    getStatusText(status) {
      const map = { 0: '待审批', 1: '已通过', 2: '已拒绝' }
      return map[status] || '未知'
    },

    getStatusClass(status) {
      const map = { 0: 'tag-warning', 1: 'tag-success', 2: 'tag-danger' }
      return map[status] || ''
    },

    async handleApprove(item) {
      uni.showModal({
        title: '确认通过',
        content: `同意 ${item.applicant?.real_name || item.applicant?.username || '该用户'} 跟进客户 ${item.customer?.customer_name || ''}？`,
        success: async (res) => {
          if (res.confirm) {
            try {
              uni.showLoading({ title: '处理中...' })
              await approveFollowUp(item.id)
              uni.hideLoading()
              uni.showToast({ title: '审批通过', icon: 'success' })
              // 刷新列表
              this.page = 1
              this.loadApprovals()
            } catch (e) {
              uni.hideLoading()
              console.error('审批失败:', e)
              uni.showToast({ title: '审批失败', icon: 'none' })
            }
          }
        }
      })
    },

    handleReject(item) {
      this.currentRejectItem = item
      this.rejectReason = ''
      this.$refs.rejectPopup.open()
    },

    async confirmReject(value) {
      if (!value || !value.trim()) {
        uni.showToast({ title: '请输入拒绝原因', icon: 'none' })
        return
      }
      try {
        uni.showLoading({ title: '处理中...' })
        await rejectFollowUp(this.currentRejectItem.id, value.trim())
        uni.hideLoading()
        uni.showToast({ title: '已拒绝', icon: 'success' })
        this.$refs.rejectPopup.close()
        // 刷新列表
        this.page = 1
        this.loadApprovals()
      } catch (e) {
        uni.hideLoading()
        console.error('拒绝失败:', e)
        uni.showToast({ title: '拒绝失败', icon: 'none' })
      }
    },

    cancelReject() {
      this.currentRejectItem = null
      this.rejectReason = ''
    }
  },
  onReachBottom() {
    if (this.hasMore && !this.loading) {
      this.page++
      this.loadApprovals()
    }
  },
  onPullDownRefresh() {
    this.page = 1
    this.approvalList = []
    this.hasMore = true
    this.loadApprovals().then(() => {
      uni.stopPullDownRefresh()
    })
  },
  onShow() {
    this.page = 1
    this.approvalList = []
    this.hasMore = true
    this.loadApprovals()
  }
}
</script>

<style lang="scss" scoped>
.approval-page {
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
  padding: 24rpx 32rpx 20rpx;
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

.approval-list {
  padding: 24rpx;
}

.approval-card {
  margin-bottom: 24rpx;
}

.approval-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20rpx;
}

.approval-type {
  font-size: 28rpx;
  font-weight: 500;
  color: #333333;
}

.approval-body {
  margin-bottom: 20rpx;
}

.approval-row {
  display: flex;
  align-items: flex-start;
  padding: 8rpx 0;
}

.approval-label {
  font-size: 26rpx;
  color: #999999;
  width: 140rpx;
  flex-shrink: 0;
}

.approval-value {
  flex: 1;
  font-size: 28rpx;
  color: #333333;
  word-break: break-all;
}

.reject-reason {
  color: #ff4d4f;
}

.approval-actions {
  display: flex;
  justify-content: flex-end;
  gap: 20rpx;
  padding-top: 20rpx;
  border-top: 1rpx solid #eeeeee;
}

.btn {
  min-width: 140rpx;
  height: 64rpx;
  line-height: 64rpx;
  font-size: 28rpx;
  border-radius: 32rpx;
  margin: 0;
  padding: 0 32rpx;
}

.btn-reject {
  background-color: #ffffff;
  color: #666666;
  border: 1rpx solid #d9d9d9;
}

.btn-approve {
  background-color: #1890ff;
  color: #ffffff;
  border: none;
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
