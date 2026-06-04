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
        <!-- 出库审批 -->
        <template v-if="currentTab === 3">
          <view class="approval-header">
            <text class="approval-type">出库申请</text>
            <view class="tag" :class="getOutboundStatusClass(item.status)">
              <text>{{ getOutboundStatusText(item.status) }}</text>
            </view>
          </view>
          <view class="approval-body">
            <view class="approval-row">
              <text class="approval-label">订单号</text>
              <text class="approval-value">{{ item.order?.order_no || '--' }}</text>
            </view>
            <view class="approval-row">
              <text class="approval-label">客户</text>
              <text class="approval-value">{{ item.order?.customer_name || '--' }}</text>
            </view>
            <view class="approval-row">
              <text class="approval-label">申请人</text>
              <text class="approval-value">{{ item.applicant_name || '--' }}</text>
            </view>
            <view class="approval-row">
              <text class="approval-label">尾款金额</text>
              <text class="approval-value text-danger">{{ item.remaining_amount || '0.00' }}元 ({{ formatRate(item.remaining_rate) }}%)</text>
            </view>
            <view class="approval-row">
              <text class="approval-label">申请时间</text>
              <text class="approval-value">{{ formatTime(item.created_at) }}</text>
            </view>
            <view v-if="item.remark" class="approval-row">
              <text class="approval-label">备注</text>
              <text class="approval-value">{{ item.remark }}</text>
            </view>
            <view v-if="item.supervisor_remark" class="approval-row">
              <text class="approval-label">主管备注</text>
              <text class="approval-value">{{ item.supervisor_remark }}</text>
            </view>
            <view v-if="item.finance_remark" class="approval-row">
              <text class="approval-label">财务备注</text>
              <text class="approval-value">{{ item.finance_remark }}</text>
            </view>
          </view>
          <!-- 出库审批操作按钮 -->
          <!-- 主管/店长只能操作status=1，财务只能操作status=2 -->
          <view v-if="(isSupervisor && item.status === 1) || (isFinance && item.status === 2)" class="approval-actions">
            <button class="btn btn-reject" @tap="handleOutboundReject(item)">拒绝</button>
            <button class="btn btn-approve" @tap="handleOutboundApprove(item)">同意</button>
          </view>
        </template>

        <!-- 其他审批 -->
        <template v-else>
          <view class="approval-header">
            <text class="approval-type">{{ getApprovalTypeText(item.approval_type) }}</text>
            <view class="tag" :class="getStatusClass(item.status)">
              <text>{{ getStatusText(item.status) }}</text>
            </view>
          </view>

          <view class="approval-body">
            <!-- 打印审批显示订单信息，跟进申请显示客户信息 -->
            <template v-if="item.approval_type === 2">
              <view class="approval-row">
                <text class="approval-label">订单号</text>
                <text class="approval-value">{{ item.order?.order_no || '--' }}</text>
              </view>
              <view class="approval-row">
                <text class="approval-label">客户</text>
                <text class="approval-value">{{ item.order?.customer_name || '--' }}</text>
              </view>
              <view class="approval-row">
                <text class="approval-label">手机号</text>
                <text class="approval-value">{{ item.order?.customer_phone || '--' }}</text>
              </view>
            </template>
            <template v-else>
              <view class="approval-row">
                <text class="approval-label">客户</text>
                <text class="approval-value">{{ item.customer?.customer_name || '--' }}</text>
              </view>
              <view class="approval-row">
                <text class="approval-label">手机号</text>
                <text class="approval-value">{{ item.customer?.phone || '--' }}</text>
              </view>
            </template>
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
        </template>
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

    <!-- 出库审批拒绝弹窗 -->
    <uni-popup ref="outboundRejectPopup" type="dialog">
      <uni-popup-dialog
        mode="input"
        title="拒绝原因"
        placeholder="请输入拒绝原因"
        :value="outboundRejectReason"
        @confirm="confirmOutboundReject"
        @close="cancelOutboundReject"
      />
    </uni-popup>

    <CustomTabBar :current="4" />
  </view>
</template>

<script>
import { ref, computed } from 'vue'
import { getPendingApprovals, getProcessedApprovals, approveFollowUp, rejectFollowUp } from '../../api/approval'
import { getPendingOutboundRequests, getOutboundRequestByOrder, supervisorApprove, financeApprove, rejectOutboundRequest } from '../../api/outbound-request'
import { useUserStore } from '../../store/user'
import CustomTabBar from '../../components/CustomTabBar.vue'

export default {
  components: { CustomTabBar },
  setup() {
    const userStore = useUserStore()
    const roles = computed(() => userStore.userInfo?.roles || [])
    const roleCodes = computed(() => roles.value.map(r => r.role_code || ''))
    // 是否是财务角色
    const isFinance = computed(() => roleCodes.value.includes('FINANCE'))
    // 是否是主管/店长角色
    const isSupervisor = computed(() =>
      roleCodes.value.includes('BOSS') ||
      roleCodes.value.includes('SUPERVISOR') ||
      roleCodes.value.includes('STORE_MANAGER')
    )
    return { userStore, roles, roleCodes, isFinance, isSupervisor }
  },
  data() {
    return {
      tabs: [
        { text: '待审批', status: 0 },
        { text: '已通过', status: 1 },
        { text: '已拒绝', status: 2 },
        { text: '出库审批', status: 3 }
      ],
      currentTab: 0,
      approvalList: [],
      loading: false,
      page: 1,
      total: 0,
      hasMore: true,
      rejectReason: '',
      currentRejectItem: null,
      outboundRejectReason: '',
      currentOutboundRejectItem: null
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
        // 出库审批单独处理
        if (this.currentTab === 3) {
          await this.loadOutboundApprovals()
          return
        }

        const status = this.tabs[this.currentTab].status
        let res
        if (status === 0) {
          // 待审批 - 查询待我审批的列表
          res = await getPendingApprovals({ page: this.page, page_size: 10 })
        } else {
          // 已通过/已拒绝 - 查询我已处理的审批列表
          res = await getProcessedApprovals({ status: status, page: this.page, page_size: 10 })
        }
        const list = res.data?.list || res.data?.records || []
        if (this.page === 1) {
          this.approvalList = list
        } else {
          this.approvalList = [...this.approvalList, ...list]
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

    // 加载出库审批列表
    async loadOutboundApprovals() {
      try {
        const res = await getPendingOutboundRequests({ page: this.page, page_size: 10 })
        const list = res.data?.list || res.data?.records || []
        if (this.page === 1) {
          this.approvalList = list
        } else {
          this.approvalList = [...this.approvalList, ...list]
        }
        this.total = res.data?.total || 0
        this.hasMore = this.approvalList.length < this.total
      } catch (e) {
        console.error('加载出库审批列表失败:', e)
        uni.showToast({ title: '加载失败', icon: 'none' })
      } finally {
        this.loading = false
      }
    },

    formatTime(time) {
      if (!time) return '--'
      return time.replace('T', ' ').substring(0, 16)
    },

    formatRate(rate) {
      if (!rate && rate !== 0) return '0.00'
      const num = parseFloat(rate)
      if (isNaN(num)) return '0.00'
      return num.toFixed(2)
    },

    getStatusText(status) {
      const map = { 0: '待审批', 1: '已通过', 2: '已拒绝' }
      return map[status] || '未知'
    },

    getStatusClass(status) {
      const map = { 0: 'tag-warning', 1: 'tag-success', 2: 'tag-danger' }
      return map[status] || ''
    },

    // 出库申请状态
    getOutboundStatusText(status) {
      const map = {
        1: '待主管审批',
        2: '待财务审批',
        3: '已驳回',
        4: '已通过'
      }
      return map[status] || '未知'
    },

    getOutboundStatusClass(status) {
      const map = {
        1: 'tag-warning',
        2: 'tag-warning',
        3: 'tag-danger',
        4: 'tag-success'
      }
      return map[status] || ''
    },

    getApprovalTypeText(type) {
      const map = { 1: '跟进申请', 2: '打印审批' }
      return map[type] || '未知'
    },

    async handleApprove(item) {
      const applicantName = item.applicant?.real_name || item.applicant?.username || '该用户'
      let content = ''
      if (item.approval_type === 2) {
        content = `同意 ${applicantName} 的送货单打印申请？`
      } else {
        content = `同意 ${applicantName} 跟进客户 ${item.customer?.customer_name || ''}？`
      }
      uni.showModal({
        title: '确认通过',
        content: content,
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
    },

    // 出库审批 - 同意
    async handleOutboundApprove(item) {
      const content = `确认通过该出库申请？`
      uni.showModal({
        title: '确认通过',
        content: content,
        success: async (res) => {
          if (res.confirm) {
            try {
              uni.showLoading({ title: '处理中...' })
              // 根据用户角色调用对应的审批接口
              if (this.isFinance) {
                // 财务调用财务审批接口
                await financeApprove(item.id, {})
              } else if (this.isSupervisor) {
                // 主管/店长调用主管审批接口
                await supervisorApprove(item.id, {})
              }
              uni.hideLoading()
              uni.showToast({ title: '审批通过', icon: 'success' })
              this.page = 1
              this.loadOutboundApprovals()
            } catch (e) {
              uni.hideLoading()
              console.error('出库审批失败:', e)
              uni.showToast({ title: e.message || '审批失败', icon: 'none' })
            }
          }
        }
      })
    },

    // 出库审批 - 拒绝
    handleOutboundReject(item) {
      this.currentOutboundRejectItem = item
      this.outboundRejectReason = ''
      this.$refs.outboundRejectPopup.open()
    },

    async confirmOutboundReject(value) {
      if (!value || !value.trim()) {
        uni.showToast({ title: '请输入拒绝原因', icon: 'none' })
        return
      }
      try {
        uni.showLoading({ title: '处理中...' })
        await rejectOutboundRequest(this.currentOutboundRejectItem.id, { reason: value.trim() })
        uni.hideLoading()
        uni.showToast({ title: '已拒绝', icon: 'success' })
        this.$refs.outboundRejectPopup.close()
        this.page = 1
        this.loadOutboundApprovals()
      } catch (e) {
        uni.hideLoading()
        console.error('出库拒绝失败:', e)
        uni.showToast({ title: e.message || '拒绝失败', icon: 'none' })
      }
    },

    cancelOutboundReject() {
      this.currentOutboundRejectItem = null
      this.outboundRejectReason = ''
    }
  },
  onReachBottom() {
    if (this.hasMore && !this.loading) {
      this.page++
      if (this.currentTab === 3) {
        this.loadOutboundApprovals()
      } else {
        this.loadApprovals()
      }
    }
  },
  onPullDownRefresh() {
    this.page = 1
    this.approvalList = []
    this.hasMore = true
    if (this.currentTab === 3) {
      this.loadOutboundApprovals().then(() => {
        uni.stopPullDownRefresh()
      })
    } else {
      this.loadApprovals().then(() => {
        uni.stopPullDownRefresh()
      })
    }
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

/* 文字颜色 */
.text-danger {
  color: #ff4d4f;
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
