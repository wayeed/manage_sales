<template>
  <view class="edit-page">
    <view class="form-section card">
      <text class="section-title">基本信息</text>
      <view class="form-item">
        <text class="form-label">姓名 <text class="required">*</text></text>
        <input v-model="form.name" class="input-field" placeholder="请输入姓名" />
      </view>
      <view class="form-item">
        <text class="form-label">手机号 <text class="required">*</text></text>
        <input v-model="form.phone" class="input-field" type="number" maxlength="11" placeholder="请输入手机号" />
      </view>
      <view class="form-item">
        <text class="form-label">邮箱</text>
        <input v-model="form.email" class="input-field" placeholder="请输入邮箱" />
      </view>
    </view>

    <view class="submit-section">
      <button class="btn-primary submit-btn" :class="{ disabled: submitting }" @tap="handleSubmit">
        {{ submitting ? '保存中...' : '保存修改' }}
      </button>
    </view>
  </view>
</template>

<script>
import { ref, onMounted } from 'vue'
import { useUserStore } from '../../store/user'
import { updateProfile } from '../../api/user'

export default {
  setup() {
    const userStore = useUserStore()
    const form = ref({
      name: '',
      phone: '',
      email: ''
    })
    const submitting = ref(false)

    onMounted(() => {
      if (userStore.userInfo) {
        form.value.name = userStore.userInfo.name || ''
        form.value.phone = userStore.userInfo.phone || ''
        form.value.email = userStore.userInfo.email || ''
      }
    })

    const validate = () => {
      if (!form.value.name) {
        uni.showToast({ title: '请输入姓名', icon: 'none' })
        return false
      }
      if (!form.value.phone) {
        uni.showToast({ title: '请输入手机号', icon: 'none' })
        return false
      }
      if (!/^1\d{10}$/.test(form.value.phone)) {
        uni.showToast({ title: '请输入正确的手机号', icon: 'none' })
        return false
      }
      if (form.value.email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.value.email)) {
        uni.showToast({ title: '请输入正确的邮箱', icon: 'none' })
        return false
      }
      return true
    }

    const handleSubmit = async () => {
      if (submitting.value) return
      if (!validate()) return

      submitting.value = true
      try {
        await updateProfile(form.value)
        await userStore.fetchUserInfo()
        uni.showToast({ title: '保存成功', icon: 'success' })
        setTimeout(() => {
          uni.navigateBack()
        }, 1500)
      } catch (e) {
        console.error('保存失败:', e)
      } finally {
        submitting.value = false
      }
    }

    return { form, submitting, handleSubmit }
  }
}
</script>

<style lang="scss" scoped>
.edit-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding: 24rpx;
}

.form-section {
  padding: 24rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333333;
  margin-bottom: 24rpx;
}

.form-item {
  margin-bottom: 24rpx;

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
