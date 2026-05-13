<template>
  <view class="login-page">
    <view class="login-header">
      <view class="login-logo">
        <view class="logo-icon">
          <view class="logo-chair"></view>
        </view>
      </view>
      <text class="login-title">家具销售提成</text>
      <text class="login-subtitle">员工端管理系统</text>
    </view>

    <view class="login-form">
      <view class="form-item">
        <view class="form-label">用户名</view>
        <input
          v-model="form.username"
          class="input-field"
          type="text"
          placeholder="请输入用户名"
          placeholder-class="input-placeholder"
        />
      </view>

      <view class="form-item">
        <view class="form-label">密码</view>
        <input
          v-model="form.password"
          class="input-field"
          :password="!showPassword"
          placeholder="请输入密码"
          placeholder-class="input-placeholder"
        />
        <view class="password-toggle" @tap="showPassword = !showPassword">
          <text class="toggle-text">{{ showPassword ? '隐藏' : '显示' }}</text>
        </view>
      </view>

      <button class="btn-primary login-btn" :class="{ disabled: loading }" @tap="handleLogin">
        {{ loading ? '登录中...' : '登 录' }}
      </button>
    </view>

    <view class="login-footer">
      <text class="footer-text">v1.0.0</text>
    </view>
  </view>
</template>

<script>
import { useUserStore } from '../../store/user'

export default {
  data() {
    return {
      form: {
        username: '',
        password: ''
      },
      showPassword: false,
      loading: false
    }
  },
  methods: {
    async handleLogin() {
      if (!this.form.username) {
        return uni.showToast({ title: '请输入用户名', icon: 'none' })
      }
      if (!this.form.password) {
        return uni.showToast({ title: '请输入密码', icon: 'none' })
      }

      this.loading = true
      try {
        const userStore = useUserStore()
        await userStore.login(this.form)
        await userStore.fetchUserInfo()
        uni.showToast({ title: '登录成功', icon: 'success' })
        setTimeout(() => {
          uni.switchTab({ url: '/pages/index/index' })
        }, 500)
      } catch (err) {
        console.error('登录失败:', err)
      } finally {
        this.loading = false
      }
    }
  }
}
</script>

<style lang="scss" scoped>
.login-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #1890ff 0%, #096dd9 100%);
  display: flex;
  flex-direction: column;
  padding: 0 60rpx;
}

.login-header {
  padding-top: 180rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 80rpx;
}

.login-logo {
  margin-bottom: 30rpx;
}

.logo-icon {
  width: 120rpx;
  height: 120rpx;
  background-color: rgba(255, 255, 255, 0.2);
  border-radius: 30rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-chair {
  width: 60rpx;
  height: 60rpx;
  border: 6rpx solid #ffffff;
  border-radius: 8rpx 8rpx 0 0;
  position: relative;

  &::after {
    content: '';
    position: absolute;
    bottom: -16rpx;
    left: 50%;
    transform: translateX(-50%);
    width: 70rpx;
    height: 6rpx;
    background-color: #ffffff;
    border-radius: 3rpx;
  }
}

.login-title {
  font-size: 44rpx;
  font-weight: bold;
  color: #ffffff;
  margin-bottom: 12rpx;
}

.login-subtitle {
  font-size: 28rpx;
  color: rgba(255, 255, 255, 0.8);
}

.login-form {
  background-color: #ffffff;
  border-radius: 20rpx;
  padding: 50rpx 40rpx;
  box-shadow: 0 8rpx 30rpx rgba(0, 0, 0, 0.1);
}

.form-item {
  margin-bottom: 36rpx;
  position: relative;
}

.form-label {
  font-size: 28rpx;
  color: #333333;
  font-weight: 500;
  margin-bottom: 16rpx;
}

.input-field {
  width: 100%;
  height: 88rpx;
  background-color: #f5f5f5;
  border: 2rpx solid #eeeeee;
  border-radius: 12rpx;
  padding: 0 24rpx;
  font-size: 28rpx;
  color: #333333;

  &:focus {
    border-color: #1890ff;
    background-color: #ffffff;
  }
}

.input-placeholder {
  color: #cccccc;
}

.password-toggle {
  position: absolute;
  right: 24rpx;
  bottom: 24rpx;
}

.toggle-text {
  font-size: 24rpx;
  color: #1890ff;
}

.login-btn {
  margin-top: 20rpx;
  width: 100%;
  height: 88rpx;
  line-height: 88rpx;
  font-size: 32rpx;
  border-radius: 12rpx;
  background-color: #1890ff;
  color: #ffffff;
  border: none;

  &:active {
    background-color: #096dd9;
  }

  &.disabled {
    opacity: 0.6;
  }
}

.login-footer {
  margin-top: auto;
  padding: 40rpx 0;
  display: flex;
  justify-content: center;
}

.footer-text {
  font-size: 24rpx;
  color: rgba(255, 255, 255, 0.6);
}
</style>
