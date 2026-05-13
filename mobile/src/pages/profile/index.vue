<template>
  <view class="profile-page">
    <!-- 用户信息头部 -->
    <view class="profile-header">
      <view class="user-info">
        <view class="avatar">
          <text class="avatar-text">{{ avatarText }}</text>
        </view>
        <view class="user-detail">
          <text class="user-name">{{ userInfo.name || userInfo.username || '未登录' }}</text>
          <text class="user-role">{{ userInfo.roleName || '销售人员' }}</text>
          <text class="user-code">工号: {{ userInfo.employeeNo || '--' }}</text>
        </view>
      </view>
    </view>

    <!-- 月度数据概览 -->
    <view class="month-stats card">
      <view class="stat-item">
        <text class="stat-num">{{ monthData.totalSales || '0' }}</text>
        <text class="stat-label">本月销售额(元)</text>
      </view>
      <view class="stat-divider"></view>
      <view class="stat-item">
        <text class="stat-num stat-num--blue">{{ monthData.totalCommission || '0' }}</text>
        <text class="stat-label">本月提成(元)</text>
      </view>
    </view>

    <!-- 功能菜单 -->
    <view class="menu-section card">
      <view class="menu-item" @tap="goTo('/pages/profile/edit')">
        <view class="menu-icon menu-icon--blue">
          <text class="menu-icon-text">P</text>
        </view>
        <text class="menu-text">编辑资料</text>
        <text class="menu-arrow">></text>
      </view>
      <view class="menu-item" @tap="goTo('/pages/profile/password')">
        <view class="menu-icon menu-icon--orange">
          <text class="menu-icon-text">K</text>
        </view>
        <text class="menu-text">修改密码</text>
        <text class="menu-arrow">></text>
      </view>
      <view class="menu-item" @tap="goTo('/pages/approval/pending')">
        <view class="menu-icon menu-icon--purple">
          <text class="menu-icon-text">S</text>
        </view>
        <text class="menu-text">审批管理</text>
        <text class="menu-arrow">></text>
      </view>
      <view class="menu-item" @tap="goTo('/pages/approval/my')">
        <view class="menu-icon menu-icon--cyan">
          <text class="menu-icon-text">M</text>
        </view>
        <text class="menu-text">我的申请</text>
        <text class="menu-arrow">></text>
      </view>
      <view class="menu-item" @tap="showAbout">
        <view class="menu-icon menu-icon--green">
          <text class="menu-icon-text">A</text>
        </view>
        <text class="menu-text">关于系统</text>
        <text class="menu-arrow">></text>
      </view>
      <view class="menu-item" @tap="handleLogout">
        <view class="menu-icon menu-icon--red">
          <text class="menu-icon-text">Q</text>
        </view>
        <text class="menu-text menu-text--danger">退出登录</text>
        <text class="menu-arrow">></text>
      </view>
    </view>

    <!-- 底部占位 -->
    <view style="height: 120rpx;"></view>

    <CustomTabBar :current="4" />
  </view>
</template>

<script>
import { computed, ref, onMounted } from 'vue'
import { useUserStore } from '../../store/user'
import { getOverview } from '../../api/performance'
import CustomTabBar from '../../components/CustomTabBar.vue'

export default {
  components: { CustomTabBar },
  setup() {
    const userStore = useUserStore()
    const monthData = ref({})

    const userInfo = computed(() => {
      return userStore.userInfo || {}
    })

    const avatarText = computed(() => {
      const name = userStore.userInfo?.name || userStore.userInfo?.username || '?'
      return name.charAt(0).toUpperCase()
    })

    const goTo = (url) => {
      uni.navigateTo({ url })
    }

    const showAbout = () => {
      uni.showModal({
        title: '关于系统',
        content: '家具销售提成管理系统 v1.0.0',
        showCancel: false
      })
    }

    const handleLogout = () => {
      uni.showModal({
        title: '提示',
        content: '确定要退出登录吗？',
        success: (res) => {
          if (res.confirm) {
            userStore.logout()
          }
        }
      })
    }

    const loadMonthData = async () => {
      try {
        const res = await getOverview()
        monthData.value = res.data || {}
      } catch (e) {
        monthData.value = { totalSales: '0', totalCommission: '0' }
      }
    }

    onMounted(() => {
      if (userStore.isLoggedIn()) {
        userStore.fetchUserInfo().catch(() => {})
        loadMonthData()
      }
    })

    return { userInfo, avatarText, monthData, goTo, showAbout, handleLogout }
  }
}
</script>

<style lang="scss" scoped>
.profile-page {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.profile-header {
  background: linear-gradient(135deg, #1890ff 0%, #096dd9 100%);
  padding: 60rpx 30rpx 80rpx;
}

.user-info {
  display: flex;
  align-items: center;
}

.avatar {
  width: 120rpx;
  height: 120rpx;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 24rpx;
}

.avatar-text {
  font-size: 48rpx;
  font-weight: bold;
  color: #ffffff;
}

.user-detail {
  display: flex;
  flex-direction: column;
}

.user-name {
  font-size: 36rpx;
  font-weight: bold;
  color: #ffffff;
  margin-bottom: 8rpx;
}

.user-role {
  font-size: 24rpx;
  color: rgba(255, 255, 255, 0.8);
  margin-bottom: 4rpx;
}

.user-code {
  font-size: 22rpx;
  color: rgba(255, 255, 255, 0.6);
}

/* 月度数据 */
.month-stats {
  display: flex;
  align-items: center;
  margin: -40rpx 24rpx 24rpx;
  padding: 30rpx 0;
}

.stat-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-num {
  font-size: 32rpx;
  font-weight: bold;
  color: #333333;
  margin-bottom: 8rpx;

  &--blue {
    color: #1890ff;
  }
}

.stat-label {
  font-size: 22rpx;
  color: #999999;
}

.stat-divider {
  width: 1rpx;
  height: 60rpx;
  background-color: #eeeeee;
}

/* 功能菜单 */
.menu-section {
  margin: 0 24rpx 24rpx;
  padding: 0 24rpx;
}

.menu-item {
  display: flex;
  align-items: center;
  padding: 28rpx 0;
  border-bottom: 1rpx solid #eeeeee;

  &:last-child {
    border-bottom: none;
  }

  &:active {
    opacity: 0.8;
  }
}

.menu-icon {
  width: 56rpx;
  height: 56rpx;
  border-radius: 14rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 20rpx;

  &--blue {
    background-color: #e6f7ff;
  }

  &--orange {
    background-color: #fff7e6;
  }

  &--green {
    background-color: #f6ffed;
  }

  &--red {
    background-color: #fff2f0;
  }

  &--purple {
    background-color: #f9f0ff;
  }

  &--cyan {
    background-color: #e6fffb;
  }
}

.menu-icon-text {
  font-size: 28rpx;
  font-weight: bold;

  .menu-icon--blue & {
    color: #1890ff;
  }

  .menu-icon--orange & {
    color: #faad14;
  }

  .menu-icon--green & {
    color: #52c41a;
  }

  .menu-icon--red & {
    color: #ff4d4f;
  }

  .menu-icon--purple & {
    color: #722ed1;
  }

  .menu-icon--cyan & {
    color: #13c2c2;
  }
}

.menu-text {
  flex: 1;
  font-size: 28rpx;
  color: #333333;

  &--danger {
    color: #ff4d4f;
  }
}

.menu-arrow {
  font-size: 24rpx;
  color: #cccccc;
}
</style>
