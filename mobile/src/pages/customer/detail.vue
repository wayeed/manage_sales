<template>
  <view class="detail-page">
    <!-- 客户信息 -->
    <view class="section card">
      <view class="customer-main">
        <view class="cm-left">
          <text class="cm-name">{{ customerName }}</text>
          <text class="cm-phone">{{ customerPhone || '-' }}</text>
        </view>
        <view class="cm-edit-btn" @tap="openEditModal">
          <text class="cm-edit-text">编辑</text>
        </view>
      </view>
      <view v-if="customerAddress" class="cm-address">
        <text class="cm-address-label">地址</text>
        <text class="cm-address-value">{{ customerAddress }}</text>
      </view>
      <view class="cm-source">
        <text class="cm-source-label">来源</text>
        <text class="cm-source-value">{{ getSourceTypeLabel(customerSourceType) }}</text>
      </view>
    </view>

    <!-- 编辑客户弹窗 -->
    <view v-if="showEditModal" class="edit-modal-overlay" @tap="closeEditModal">
      <view class="edit-modal" @tap.stop>
        <view class="edit-modal-header">
          <text class="edit-modal-title">编辑客户</text>
          <view class="edit-modal-close" @tap="closeEditModal">
            <text class="edit-modal-close-text">×</text>
          </view>
        </view>
        <view class="edit-modal-body">
          <view class="edit-field">
            <text class="edit-label">客户姓名 <text class="edit-required">*</text></text>
            <input v-model="editForm.customer_name" class="edit-input" placeholder="请输入客户姓名" />
          </view>
          <view class="edit-field">
            <text class="edit-label">手机号 <text class="edit-required">*</text></text>
            <input v-model="editForm.phone" class="edit-input" placeholder="请输入手机号" maxlength="11" />
            <text v-if="customerPhone && editForm.phone !== customerPhone" class="edit-tip">修改后将记录原手机号</text>
          </view>
          <view class="edit-field">
            <text class="edit-label">地址</text>
            <input v-model="editForm.address" class="edit-input" placeholder="请输入地址" />
          </view>
          <view class="edit-field">
            <text class="edit-label">客户来源</text>
            <text class="edit-source-display">{{ getSourceTypeLabel(customerSourceType) }}</text>
          </view>
        </view>
        <view class="edit-modal-footer">
          <view class="edit-cancel-btn" @tap="closeEditModal">
            <text class="edit-cancel-text">取消</text>
          </view>
          <view class="edit-submit-btn" @tap="submitEdit">
            <text class="edit-submit-text">保存</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 未提交订单 -->
    <view v-if="hasDraft" class="section card">
      <view class="section-header">
        <text class="section-title">未提交订单</text>
        <view class="go-order-btn" @tap="goToOrder">
          <text class="go-order-text">去下单</text>
          <text class="go-order-arrow">›</text>
        </view>
      </view>
      <view v-if="draftLoading" class="draft-loading">
        <text class="draft-loading-text">加载中...</text>
      </view>
      <view v-else-if="draftItems.length > 0" class="draft-list">
        <view v-for="(item, index) in draftItems" :key="index" class="draft-item">
          <text class="di-name">{{ item.product_name || item.sku_name }}</text>
          <text class="di-qty">x{{ item.quantity }}</text>
          <text class="di-price">¥{{ item.sale_price || item.list_price || '0' }}</text>
        </view>
        <view v-if="draftGifts.length > 0" class="draft-gifts">
          <text class="dg-label">赠品：</text>
          <text v-for="(g, gi) in draftGifts" :key="gi" class="dg-item">{{ g.gift_name }} x{{ g.quantity }}</text>
        </view>
      </view>
      <view v-else class="draft-empty">
        <text class="draft-empty-text">暂无商品信息</text>
      </view>
    </view>

    <!-- 跟进记录 -->
    <view class="section card">
      <view class="section-header">
        <text class="section-title">跟进记录</text>
        <view class="add-follow-btn" @tap="showFollowUpInput = true">
          <text class="add-follow-text">+ 添加</text>
        </view>
      </view>

      <!-- 添加跟进记录 -->
      <view v-if="showFollowUpInput" class="follow-input-wrap">
        <textarea
          v-model="followContent"
          class="follow-textarea"
          placeholder="输入跟进内容..."
          :maxlength="500"
        />
        <view class="follow-input-actions">
          <view class="follow-cancel-btn" @tap="showFollowUpInput = false; followContent = ''">
            <text class="follow-cancel-text">取消</text>
          </view>
          <view class="follow-submit-btn" @tap="submitFollowUp">
            <text class="follow-submit-text">提交</text>
          </view>
        </view>
      </view>

      <!-- 跟进记录列表 -->
      <view v-if="followUps.length > 0" class="follow-list">
        <view v-for="item in followUps" :key="item.id" class="follow-item">
          <view class="fi-top">
            <text class="fi-follower">{{ item.follower?.real_name || item.follower?.username || '我' }}</text>
            <text class="fi-time">{{ formatTime(item.created_at) }}</text>
          </view>
          <text class="fi-content">{{ item.content }}</text>
        </view>
      </view>
      <view v-else class="follow-empty">
        <text class="follow-empty-text">暂无跟进记录</text>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, onMounted, reactive } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getCustomerDraft } from '../../api/order'
import { getCustomerFollowUps, addCustomerFollowUp, updateCustomer, getCustomerDetail } from '../../api/customer'

const customerId = ref(null)
const customerName = ref('')
const customerPhone = ref('')
const customerAddress = ref('')
const customerSourceType = ref(0)
const hasDraft = ref(false)
const draftId = ref(null)
const draftItems = ref([])
const draftGifts = ref([])
const draftLoading = ref(false)
const followUps = ref([])
const showFollowUpInput = ref(false)
const followContent = ref('')

// 编辑相关
const showEditModal = ref(false)
const editForm = reactive({
  customer_name: '',
  phone: '',
  address: ''
})

const getSourceTypeLabel = (type) => {
  const map = { 0: '自然进店', 1: '主动邀约', 2: '同行带单' }
  return map[type] ?? '-'
}

const openEditModal = () => {
  editForm.customer_name = customerName.value
  editForm.phone = customerPhone.value
  editForm.address = customerAddress.value
  showEditModal.value = true
}

const closeEditModal = () => {
  showEditModal.value = false
}

const submitEdit = async () => {
  if (!editForm.customer_name.trim()) {
    uni.showToast({ title: '请输入客户姓名', icon: 'none' })
    return
  }
  if (!editForm.phone.trim()) {
    uni.showToast({ title: '请输入手机号', icon: 'none' })
    return
  }
  if (!/^1[3-9]\d{9}$/.test(editForm.phone)) {
    uni.showToast({ title: '请输入正确的手机号', icon: 'none' })
    return
  }
  try {
    await updateCustomer(customerId.value, {
      customer_name: editForm.customer_name,
      phone: editForm.phone,
      address: editForm.address
    })
    uni.showToast({ title: '保存成功', icon: 'success' })
    // 更新本地数据
    customerName.value = editForm.customer_name
    customerPhone.value = editForm.phone
    customerAddress.value = editForm.address
    showEditModal.value = false
  } catch (e) {
    uni.showToast({ title: '保存失败', icon: 'none' })
  }
}

const formatTime = (time) => {
  if (!time) return ''
  return time.replace('T', ' ').substring(0, 16)
}

const loadDraft = async () => {
  if (!customerId.value) return
  draftLoading.value = true
  try {
    const res = await getCustomerDraft(customerId.value)
    if (res.data) {
      draftItems.value = res.data.items || []
      draftGifts.value = res.data.gifts || []
      draftId.value = res.data.order?.id || res.data.id
      hasDraft.value = true
    } else {
      hasDraft.value = false
    }
  } catch (e) {
    console.error('加载草稿失败:', e)
    hasDraft.value = false
  } finally {
    draftLoading.value = false
  }
}

const loadFollowUps = async () => {
  if (!customerId.value) return
  try {
    const res = await getCustomerFollowUps(customerId.value)
    followUps.value = res.data || []
  } catch (e) {
    console.error('加载跟进记录失败:', e)
  }
}

const submitFollowUp = async () => {
  if (!followContent.value.trim()) {
    uni.showToast({ title: '请输入跟进内容', icon: 'none' })
    return
  }
  try {
    await addCustomerFollowUp(customerId.value, {
      content: followContent.value.trim(),
      follow_type: 1
    })
    uni.showToast({ title: '添加成功', icon: 'success' })
    followContent.value = ''
    showFollowUpInput.value = false
    loadFollowUps()
  } catch (e) {
    uni.showToast({ title: '添加失败', icon: 'none' })
  }
}

const goToOrder = () => {
  uni.navigateTo({
    url: `/pages/orders/create?customer_id=${customerId.value}&load_draft=1`
  })
}

onLoad((options) => {
  customerId.value = options.id
  customerName.value = decodeURIComponent(options.name || '')
  customerPhone.value = decodeURIComponent(options.phone || '')
  customerAddress.value = decodeURIComponent(options.address || '')
  customerSourceType.value = parseInt(options.source_type || '0')

  // 无论是否有草稿标记，都尝试加载草稿
  loadDraft()
  loadFollowUps()
})
</script>

<style lang="scss" scoped>
.detail-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 16rpx 24rpx;
  padding-bottom: 40rpx;
}

.section {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
}

.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
}

/* 客户信息 */
.customer-main {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.cm-left {
  display: flex;
  align-items: baseline;
  gap: 16rpx;
}

.cm-name {
  font-size: 36rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.cm-phone {
  font-size: 28rpx;
  color: #666;
}

.cm-address {
  display: flex;
  align-items: flex-start;
  gap: 12rpx;
  margin-top: 16rpx;
}

.cm-address-label {
  font-size: 24rpx;
  color: #999;
  flex-shrink: 0;
  width: 60rpx;
}

.cm-address-value {
  font-size: 24rpx;
  color: #666;
  line-height: 1.5;
}

/* 去下单按钮 */
.go-order-btn {
  display: flex;
  align-items: center;
  background: #1890ff;
  padding: 10rpx 24rpx;
  border-radius: 8rpx;
  &:active { opacity: 0.8; }
}

.go-order-text {
  font-size: 26rpx;
  color: #fff;
  font-weight: 500;
}

.go-order-arrow {
  font-size: 32rpx;
  color: #fff;
  margin-left: 4rpx;
}

/* 草稿列表 */
.draft-list {
  background: #f7f9fc;
  border-radius: 12rpx;
  padding: 16rpx;
}

.draft-item {
  display: flex;
  align-items: center;
  padding: 12rpx 0;
  border-bottom: 1rpx solid #eee;
  &:last-child { border-bottom: none; }
}

.di-name {
  flex: 1;
  font-size: 26rpx;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-right: 16rpx;
}

.di-qty {
  font-size: 24rpx;
  color: #999;
  flex-shrink: 0;
  margin-right: 16rpx;
}

.di-price {
  font-size: 26rpx;
  color: #e8453c;
  font-weight: 600;
  flex-shrink: 0;
}

.draft-gifts {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8rpx;
  padding-top: 12rpx;
  border-top: 1rpx dashed #eee;
  margin-top: 8rpx;
}

.dg-label {
  font-size: 24rpx;
  color: #faad14;
  font-weight: 500;
}

.dg-item {
  font-size: 22rpx;
  color: #999;
  background: #fff8e6;
  padding: 2rpx 12rpx;
  border-radius: 6rpx;
}

.draft-loading,
.draft-empty,
.follow-empty {
  padding: 32rpx 0;
  text-align: center;
}

.draft-loading-text,
.draft-empty-text,
.follow-empty-text {
  font-size: 26rpx;
  color: #999;
}

/* 添加跟进 */
.add-follow-btn {
  padding: 8rpx 20rpx;
  background: #f0f7ff;
  border-radius: 8rpx;
  &:active { background: #e0efff; }
}

.add-follow-text {
  font-size: 24rpx;
  color: #1890ff;
  font-weight: 500;
}

.follow-input-wrap {
  background: #f7f9fc;
  border-radius: 12rpx;
  padding: 16rpx;
  margin-bottom: 16rpx;
}

.follow-textarea {
  width: 100%;
  height: 160rpx;
  background: #fff;
  border: 2rpx solid #eee;
  border-radius: 10rpx;
  padding: 16rpx;
  font-size: 28rpx;
  color: #333;
  box-sizing: border-box;
}

.follow-input-actions {
  display: flex;
  justify-content: flex-end;
  gap: 16rpx;
  margin-top: 16rpx;
}

.follow-cancel-btn {
  padding: 12rpx 32rpx;
  border-radius: 8rpx;
  background: #f5f5f5;
  &:active { background: #eee; }
}

.follow-cancel-text {
  font-size: 26rpx;
  color: #666;
}

.follow-submit-btn {
  padding: 12rpx 32rpx;
  border-radius: 8rpx;
  background: #1890ff;
  &:active { opacity: 0.8; }
}

.follow-submit-text {
  font-size: 26rpx;
  color: #fff;
  font-weight: 500;
}

/* 跟进记录列表 */
.follow-list {
  margin-top: 8rpx;
}

.follow-item {
  padding: 20rpx 0;
  border-bottom: 1rpx solid #f0f0f0;
  &:last-child { border-bottom: none; }
}

.fi-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8rpx;
}

.fi-follower {
  font-size: 26rpx;
  color: #1890ff;
  font-weight: 500;
}

.fi-time {
  font-size: 22rpx;
  color: #bbb;
}

.fi-content {
  font-size: 28rpx;
  color: #333;
  line-height: 1.6;
}

/* 编辑按钮 */
.cm-edit-btn {
  padding: 8rpx 20rpx;
  background: #f0f7ff;
  border-radius: 8rpx;
  &:active { background: #e0efff; }
}

.cm-edit-text {
  font-size: 24rpx;
  color: #1890ff;
  font-weight: 500;
}

/* 客户来源 */
.cm-source {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-top: 12rpx;
}

.cm-source-label {
  font-size: 24rpx;
  color: #999;
  flex-shrink: 0;
  width: 60rpx;
}

.cm-source-value {
  font-size: 24rpx;
  color: #666;
}

/* 编辑弹窗 */
.edit-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.edit-modal {
  width: 640rpx;
  background: #fff;
  border-radius: 16rpx;
  overflow: hidden;
}

.edit-modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24rpx 32rpx;
  border-bottom: 1rpx solid #f0f0f0;
}

.edit-modal-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.edit-modal-close {
  padding: 8rpx;
  &:active { opacity: 0.6; }
}

.edit-modal-close-text {
  font-size: 40rpx;
  color: #999;
  line-height: 1;
}

.edit-modal-body {
  padding: 24rpx 32rpx;
}

.edit-field {
  margin-bottom: 24rpx;
  &:last-child { margin-bottom: 0; }
}

.edit-label {
  display: block;
  font-size: 28rpx;
  color: #333;
  margin-bottom: 12rpx;
}

.edit-required {
  color: #e8453c;
}

.edit-input {
  width: 100%;
  height: 80rpx;
  background: #f7f7f7;
  border-radius: 10rpx;
  padding: 0 20rpx;
  font-size: 28rpx;
  color: #333;
  box-sizing: border-box;
}

.edit-tip {
  display: block;
  font-size: 22rpx;
  color: #faad14;
  margin-top: 8rpx;
}

.edit-source-display {
  display: block;
  font-size: 28rpx;
  color: #666;
  padding: 20rpx 0;
}

.edit-radio-group {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}

.edit-radio {
  display: flex;
  align-items: center;
  gap: 10rpx;
  padding: 16rpx 24rpx;
  background: #f7f7f7;
  border-radius: 10rpx;
  border: 2rpx solid transparent;

  &--active {
    background: #e6f4ff;
    border-color: #1890ff;
  }
}

.edit-radio-dot {
  width: 36rpx;
  height: 36rpx;
  border-radius: 50%;
  border: 2rpx solid #ccc;
  display: flex;
  align-items: center;
  justify-content: center;

  .edit-radio--active & {
    border-color: #1890ff;
  }
}

.edit-radio-inner {
  width: 20rpx;
  height: 20rpx;
  border-radius: 50%;
  background: #1890ff;
}

.edit-radio-text {
  font-size: 28rpx;
  color: #333;

  .edit-radio--active & {
    color: #1890ff;
    font-weight: 500;
  }
}

.edit-modal-footer {
  display: flex;
  gap: 20rpx;
  padding: 24rpx 32rpx;
  border-top: 1rpx solid #f0f0f0;
}

.edit-cancel-btn {
  flex: 1;
  padding: 20rpx 0;
  background: #f5f5f5;
  border-radius: 10rpx;
  text-align: center;
  &:active { background: #eee; }
}

.edit-cancel-text {
  font-size: 28rpx;
  color: #666;
}

.edit-submit-btn {
  flex: 1;
  padding: 20rpx 0;
  background: #1890ff;
  border-radius: 10rpx;
  text-align: center;
  &:active { opacity: 0.8; }
}

.edit-submit-text {
  font-size: 28rpx;
  color: #fff;
  font-weight: 500;
}
</style>
