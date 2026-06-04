<template>
  <view class="profile-page">
    <!-- 头部 -->
    <view class="profile-header">
      <view class="avatar-wrap">
        <image
          class="avatar"
          :src="userInfo?.avatar || '/static/avatar-default.png'"
          mode="aspectFill"
        />
      </view>
      <text class="user-name">{{ userInfo?.real_name || '--' }}</text>
      <text class="user-role">{{ roleNames }}</text>
    </view>

    <!-- 基本信息 -->
    <view class="section">
      <text class="section-title">基本信息</text>
      <view class="info-list">
        <view class="info-item">
          <text class="info-label">用户名</text>
          <text class="info-value">{{ userInfo?.username || '--' }}</text>
        </view>
        <view class="info-item">
          <text class="info-label">工号</text>
          <text class="info-value">{{ userInfo?.employee_no || '--' }}</text>
        </view>
        <view class="info-item">
          <text class="info-label">姓名</text>
          <text class="info-value">{{ userInfo?.real_name || '--' }}</text>
        </view>
        <view class="info-item">
          <text class="info-label">手机号</text>
          <text class="info-value">{{ userInfo?.phone || '--' }}</text>
        </view>
        <view class="info-item">
          <text class="info-label">所属门店</text>
          <text class="info-value">{{ userInfo?.store_name || '--' }}</text>
        </view>
        <view class="info-item">
          <text class="info-label">入职日期</text>
          <text class="info-value">{{ formatDate(userInfo?.entry_date) }}</text>
        </view>
      </view>
    </view>

    <!-- 工作信息 -->
    <view class="section">
      <text class="section-title">工作信息</text>
      <view class="info-list">
        <view class="info-item">
          <text class="info-label">我的角色</text>
          <text class="info-value">{{ roleNames || '--' }}</text>
        </view>
        <view class="info-item">
          <text class="info-label">我的上级</text>
          <text class="info-value">{{ userInfo?.parent_name || '--' }}</text>
        </view>
        <view class="info-item">
          <text class="info-label">业务员等级</text>
          <text class="info-value">{{ levelText }}</text>
        </view>
      </view>
    </view>

    <!-- 银行信息 -->
    <view class="section">
      <text class="section-title">银行信息</text>
      <view class="info-list">
        <view class="info-item">
          <text class="info-label">开户银行</text>
          <text class="info-value">{{ userInfo?.bank_name || '--' }}</text>
        </view>
        <view class="info-item">
          <text class="info-label">银行卡号</text>
          <text class="info-value">{{ maskBankCard(userInfo?.bank_account) }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { computed, onMounted } from 'vue'
import { useUserStore } from '../../store/user'

export default {
  setup() {
    const userStore = useUserStore()
    const userInfo = computed(() => userStore.userInfo)

    // 角色名称列表
    const roleNames = computed(() => {
      const roles = userInfo.value?.roles || []
      if (roles.length === 0) return '--'
      return roles.map(r => r.role_name).join('、')
    })

    // 业务员等级文字
    const levelText = computed(() => {
      const map = { 1: '初级业务员', 2: '中级业务员', 3: '高级业务员' }
      return map[userInfo.value?.level] || '--'
    })

    // 格式化日期
    const formatDate = (dateStr) => {
      if (!dateStr) return '--'
      const d = new Date(dateStr)
      const y = d.getFullYear()
      const m = String(d.getMonth() + 1).padStart(2, '0')
      const day = String(d.getDate()).padStart(2, '0')
      return `${y}-${m}-${day}`
    }

    // 银行卡号脱敏
    const maskBankCard = (cardNo) => {
      if (!cardNo) return '--'
      if (cardNo.length <= 8) return cardNo
      return cardNo.slice(0, 4) + ' **** **** ' + cardNo.slice(-4)
    }

    onMounted(() => {
      userStore.fetchUserInfo()
    })

    return {
      userInfo,
      roleNames,
      levelText,
      formatDate,
      maskBankCard
    }
  }
}
</script>

<style lang="scss" scoped>
.profile-page {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.profile-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 60rpx 24rpx 40rpx;
  background: linear-gradient(135deg, #1890ff 0%, #096dd9 100%);
}

.avatar-wrap {
  width: 128rpx;
  height: 128rpx;
  border-radius: 50%;
  overflow: hidden;
  border: 4rpx solid rgba(255, 255, 255, 0.4);
}

.avatar {
  width: 100%;
  height: 100%;
}

.user-name {
  margin-top: 20rpx;
  font-size: 36rpx;
  font-weight: 600;
  color: #ffffff;
}

.user-role {
  margin-top: 8rpx;
  font-size: 24rpx;
  color: rgba(255, 255, 255, 0.8);
}

.section {
  margin: 24rpx;
  background-color: #ffffff;
  border-radius: 16rpx;
  padding: 24rpx;
}

.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #333333;
  margin-bottom: 16rpx;
  padding-left: 16rpx;
  border-left: 6rpx solid #1890ff;
}

.info-list {
  padding: 0 8rpx;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24rpx 0;
  border-bottom: 1rpx solid #f0f0f0;

  &:last-child {
    border-bottom: none;
  }
}

.info-label {
  font-size: 28rpx;
  color: #999999;
  flex-shrink: 0;
}

.info-value {
  font-size: 28rpx;
  color: #333333;
  text-align: right;
  word-break: break-all;
}
</style>
