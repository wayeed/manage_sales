<template>
  <view class="password-page">
    <view class="form-section card">
      <text class="section-title">修改密码</text>
      <view class="form-item">
        <text class="form-label">旧密码 <text class="required">*</text></text>
        <input
          v-model="form.oldPassword"
          class="input-field"
          :password="!showOld"
          placeholder="请输入旧密码"
        />
        <text class="eye-btn" @tap="showOld = !showOld">{{ showOld ? '隐藏' : '显示' }}</text>
      </view>
      <view class="form-item">
        <text class="form-label">新密码 <text class="required">*</text></text>
        <input
          v-model="form.newPassword"
          class="input-field"
          :password="!showNew"
          placeholder="请输入新密码(6-20位)"
          @input="checkStrength"
        />
        <text class="eye-btn" @tap="showNew = !showNew">{{ showNew ? '隐藏' : '显示' }}</text>
      </view>
      <!-- 密码强度 -->
      <view v-if="form.newPassword" class="strength-area">
        <view class="strength-bars">
          <view class="strength-bar" :class="{ active: strength >= 1, weak: strength === 1, medium: strength === 2, strong: strength >= 3 }"></view>
          <view class="strength-bar" :class="{ active: strength >= 2, medium: strength === 2, strong: strength >= 3 }"></view>
          <view class="strength-bar" :class="{ active: strength >= 3, strong: strength >= 3 }"></view>
        </view>
        <text class="strength-text" :class="strengthClass">{{ strengthText }}</text>
      </view>
      <view class="form-item">
        <text class="form-label">确认密码 <text class="required">*</text></text>
        <input
          v-model="form.confirmPassword"
          class="input-field"
          :password="!showConfirm"
          placeholder="请再次输入新密码"
        />
        <text class="eye-btn" @tap="showConfirm = !showConfirm">{{ showConfirm ? '隐藏' : '显示' }}</text>
      </view>
    </view>

    <view class="tips-section card">
      <text class="tips-title">密码要求</text>
      <text class="tips-text">- 密码长度为6-20位字符</text>
      <text class="tips-text">- 建议包含字母和数字</text>
      <text class="tips-text">- 不能与原密码相同</text>
    </view>

    <view class="submit-section">
      <button class="btn-primary submit-btn" :class="{ disabled: submitting }" @tap="handleSubmit">
        {{ submitting ? '提交中...' : '确认修改' }}
      </button>
    </view>
  </view>
</template>

<script>
import { ref, computed } from 'vue'
import { changePassword } from '../../api/user'
import { useUserStore } from '../../store/user'

export default {
  setup() {
    const userStore = useUserStore()
    const form = ref({
      oldPassword: '',
      newPassword: '',
      confirmPassword: ''
    })
    const submitting = ref(false)
    const showOld = ref(false)
    const showNew = ref(false)
    const showConfirm = ref(false)
    const strength = ref(0)

    const strengthText = computed(() => {
      const map = { 0: '', 1: '弱', 2: '中', 3: '强' }
      return map[strength.value] || ''
    })

    const strengthClass = computed(() => {
      const map = { 0: '', 1: 'strength-weak', 2: 'strength-medium', 3: 'strength-strong' }
      return map[strength.value] || ''
    })

    const checkStrength = () => {
      const pwd = form.value.newPassword
      let level = 0
      if (pwd.length >= 6) level++
      if (/[a-zA-Z]/.test(pwd) && /[0-9]/.test(pwd)) level++
      if (/[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(pwd)) level++
      strength.value = level
    }

    const validate = () => {
      if (!form.value.oldPassword) {
        uni.showToast({ title: '请输入旧密码', icon: 'none' })
        return false
      }
      if (!form.value.newPassword) {
        uni.showToast({ title: '请输入新密码', icon: 'none' })
        return false
      }
      if (form.value.newPassword.length < 6 || form.value.newPassword.length > 20) {
        uni.showToast({ title: '密码长度为6-20位', icon: 'none' })
        return false
      }
      if (form.value.newPassword !== form.value.confirmPassword) {
        uni.showToast({ title: '两次密码输入不一致', icon: 'none' })
        return false
      }
      if (form.value.oldPassword === form.value.newPassword) {
        uni.showToast({ title: '新密码不能与旧密码相同', icon: 'none' })
        return false
      }
      return true
    }

    const handleSubmit = async () => {
      if (submitting.value) return
      if (!validate()) return

      submitting.value = true
      try {
        await changePassword({
          oldPassword: form.value.oldPassword,
          newPassword: form.value.newPassword
        })
        uni.showToast({ title: '修改成功，请重新登录', icon: 'success' })
        setTimeout(() => {
          userStore.logout()
        }, 1500)
      } catch (e) {
        console.error('修改密码失败:', e)
      } finally {
        submitting.value = false
      }
    }

    return {
      form,
      submitting,
      showOld,
      showNew,
      showConfirm,
      strength,
      strengthText,
      strengthClass,
      checkStrength,
      handleSubmit
    }
  }
}
</script>

<style lang="scss" scoped>
.password-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding: 24rpx;
}

.form-section {
  padding: 24rpx;
  margin-bottom: 24rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333333;
  margin-bottom: 24rpx;
}

.form-item {
  margin-bottom: 24rpx;
  position: relative;

  &:last-child {
    margin-bottom: 0;
  }
}

.form-label {
  font-size: 28rpx;
  color: #666666;
  margin-bottom: 12rpx;
  display: block;
}

.required {
  color: #ff4d4f;
}

.input-field {
  width: 100%;
  height: 80rpx;
  background-color: #f5f5f5;
  border: 2rpx solid #eeeeee;
  border-radius: 12rpx;
  padding: 0 24rpx;
  font-size: 28rpx;
  color: #333333;
}

.eye-btn {
  position: absolute;
  right: 24rpx;
  bottom: 24rpx;
  font-size: 24rpx;
  color: #1890ff;
  z-index: 1;
}

/* 密码强度 */
.strength-area {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24rpx;
}

.strength-bars {
  display: flex;
  gap: 8rpx;
  flex: 1;
}

.strength-bar {
  height: 8rpx;
  flex: 1;
  background-color: #eeeeee;
  border-radius: 4rpx;
  transition: background-color 0.3s;

  &.active.weak {
    background-color: #ff4d4f;
  }

  &.active.medium {
    background-color: #faad14;
  }

  &.active.strong {
    background-color: #52c41a;
  }
}

.strength-text {
  font-size: 24rpx;
  margin-left: 16rpx;

  &.strength-weak {
    color: #ff4d4f;
  }

  &.strength-medium {
    color: #faad14;
  }

  &.strength-strong {
    color: #52c41a;
  }
}

/* 提示 */
.tips-section {
  padding: 24rpx;
  margin-bottom: 24rpx;
}

.tips-title {
  font-size: 26rpx;
  font-weight: 500;
  color: #333333;
  margin-bottom: 16rpx;
  display: block;
}

.tips-text {
  font-size: 24rpx;
  color: #999999;
  line-height: 1.8;
  display: block;
}

/* 提交 */
.submit-section {
  padding: 30rpx 0 60rpx;
}

.submit-btn {
  width: 100%;
  height: 88rpx;
  line-height: 88rpx;
  font-size: 32rpx;
  border-radius: 12rpx;

  &.disabled {
    opacity: 0.6;
  }
}
</style>
