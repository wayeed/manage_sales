<template>
  <view class="login-page">
    <view class="login-header">
      <view class="login-logo">
        <view class="logo-icon">
          <view class="logo-chair"></view>
        </view>
      </view>
      <text class="login-title">家具销售管理系统</text>
      <text class="login-subtitle">员工端</text>
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

    <!-- APP下载引导模块 - 仅H5模式显示 -->
    <view class="download-banner" v-if="isH5">
      <view class="download-icon-wrapper">
        <view class="download-icon"></view>
      </view>
      <view class="download-content">
        <view class="download-title">下载官方APP</view>
        <view class="download-desc">体验更流畅的操作体验，随时随地管理订单、查看业绩</view>
        <view class="download-badge">
          <view class="badge-item">
          <view class="badge-dot"></view>
            极速响应</view>
          <view class="badge-item">
            <view class="badge-dot"></view>
            实时数据</view>
        </view>
        <view class="download-btn-wrapper" @tap="handleDownload">立即下载</view>
      </view>
    </view>

    <view class="login-footer">
      <text class="footer-text">v1.0.0</text>
    </view>
  </view>
</template>

<script>
import { useUserStore } from '../../store/user'
import { getLatestVersion } from '../../api/app-version'
import { APP_CONFIG } from '../../api/config'

// 默认下载链接（fallback）
const DEFAULT_DOWNLOAD_URL = APP_CONFIG.DEFAULT_DOWNLOAD_URL

export default {
  data() {
    return {
      form: {
        username: '',
        password: ''
      },
      showPassword: false,
      loading: false,
      isH5: false,
      downloadUrl: DEFAULT_DOWNLOAD_URL,
      downloading: false,
      versionInfo: null
    }
  },
  created() {
    // 判断是否H5环境
    this.isH5 = process.env.UNI_PLATFORM === 'h5'
    
    // 如果是H5环境，获取最新版本信息
    if (this.isH5) {
      this.fetchLatestVersion()
    }
  },
  methods: {
    /**
     * 获取最新APP版本信息
     */
    async fetchLatestVersion() {
      try {
        // 从缓存中获取（缓存有效期24小时）
        const cachedVersion = uni.getStorageSync('app_version_cache')
        const cacheTime = uni.getStorageSync('app_version_cache_time')
        const now = Date.now()
        
        // 如果缓存存在且未过期（24小时内），使用缓存
        if (cachedVersion && cacheTime && (now - cacheTime < 24 * 60 * 60 * 1000)) {
          const cachedUrl = this.validateDownloadUrl(cachedVersion.download_url)
          this.downloadUrl = cachedUrl || DEFAULT_DOWNLOAD_URL
          this.versionInfo = cachedVersion
          return
        }
        
        // 调用API获取最新版本信息
        const res = await getLatestVersion('android')
        if (res.code === 200 && res.data) {
          const versionData = res.data
          
          // update_type: 0-整包更新, 1-热更新
          // 只处理整包更新
          if (versionData.update_type === 0) {
            // 验证下载链接格式
            const downloadUrl = this.validateDownloadUrl(versionData.download_url)
            
            if (downloadUrl) {
              this.downloadUrl = downloadUrl
              this.versionInfo = versionData
              
              // 缓存到本地
              uni.setStorageSync('app_version_cache', versionData)
              uni.setStorageSync('app_version_cache_time', now)
              
              console.log('获取最新版本成功:', versionData.version_name, versionData.version_code)
              return
            } else {
              console.warn('最新版本下载链接无效:', versionData.download_url)
            }
          } else {
            console.log('最新版本为热更新类型，使用默认下载链接')
          }
        }
        
        // 如果没有有效数据，使用默认下载链接
        this.downloadUrl = DEFAULT_DOWNLOAD_URL
      } catch (err) {
        console.error('获取最新版本失败:', err)
        // 失败时使用默认下载链接
        this.downloadUrl = DEFAULT_DOWNLOAD_URL
      }
    },
    
    /**
     * 验证下载链接格式
     * @param {string} url - 下载链接
     * @returns {string|null} - 验证通过返回链接，否则返回null
     */
    validateDownloadUrl(url) {
      if (!url || typeof url !== 'string') {
        return null
      }
      
      // 去除首尾空格
      url = url.trim()
      
      // 验证是否为有效的URL格式
      try {
        const urlObj = new URL(url)
        
        // 验证协议
        if (!['http:', 'https:'].includes(urlObj.protocol)) {
          return null
        }
        
        // 验证是否包含文件名（至少有一个路径段）
        if (urlObj.pathname === '/' || !urlObj.pathname.includes('.')) {
          console.warn('下载链接不包含有效文件名')
          return null
        }
        
        return url
      } catch (err) {
        console.error('下载链接格式无效:', err)
        return null
      }
    },
    
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
          uni.reLaunch({ url: '/pages/index/index' })
        }, 500)
      } catch (err) {
        console.error('登录失败:', err)
      } finally {
        this.loading = false
      }
    },
    
    async handleDownload() {
      // 防止重复点击
      if (this.downloading) return
      
      this.downloading = true
      
      try {
        // 显示版本信息提示
        let toastMessage = '正在跳转下载页面'
        if (this.versionInfo) {
          toastMessage = `正在下载 v${this.versionInfo.version_name}`
        }
        
        uni.showToast({ 
          title: toastMessage, 
          icon: 'none',
          duration: 1500
        })
        
        // 延迟跳转，让用户看到提示
        await new Promise(resolve => setTimeout(resolve, 800))
        
        // 跳转到下载链接
        if (this.downloadUrl) {
          window.location.href = this.downloadUrl
        } else {
          // 如果动态链接获取失败，使用默认链接
          window.location.href = DEFAULT_DOWNLOAD_URL
        }
      } catch (err) {
        console.error('下载跳转失败:', err)
        // 出错时使用默认链接
        uni.showToast({ title: '下载跳转失败，请稍后重试', icon: 'none' })
      } finally {
        // 延迟释放下载状态，防止用户快速重复点击
        setTimeout(() => {
          this.downloading = false
        }, 2000)
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
  font-size: 32rpx;
  border-radius: 12rpx;
  background-color: #1890ff;
  color: #ffffff;
  border: none;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;

  &:active {
    background-color: #096dd9;
  }

  &.disabled {
    opacity: 0.6;
  }
}

/* APP下载引导模块样式 */
.download-banner {
  background: linear-gradient(135deg, rgba(255, 249, 230, 0.98) 0%, rgba(255, 245, 204, 0.98) 100%);
  border: 3rpx dashed #ffc53d;
  border-radius: 16rpx;
  padding: 30rpx;
  margin-top: 30rpx;
  display: flex;
  align-items: center;
  gap: 24rpx;
}

.download-icon-wrapper {
  width: 96rpx;
  height: 96rpx;
  background: linear-gradient(135deg, #ffc53d 0%, #ffa940 100%);
  border-radius: 24rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 8rpx 24rpx rgba(255, 197, 61, 0.4);
}

.download-icon {
  width: 52rpx;
  height: 52rpx;
  border: 4rpx solid #fff;
  border-radius: 8rpx 8rpx 0 0;
  position: relative;

  &::before {
    content: '';
    position: absolute;
    bottom: -20rpx;
    left: 50%;
    transform: translateX(-50%);
    width: 60rpx;
    height: 6rpx;
    background-color: #fff;
    border-radius: 3rpx;
  }

  &::after {
    content: '';
    position: absolute;
    top: 10rpx;
    left: 50%;
    transform: translateX(-50%);
    width: 16rpx;
    height: 20rpx;
    border-left: 4rpx solid #fff;
    border-bottom: 4rpx solid #fff;
  }
}

.download-content {
  flex: 1;
  min-width: 0;
}

.download-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #d46b08;
  margin-bottom: 8rpx;
}

.download-desc {
  font-size: 24rpx;
  color: #fa8c16;
  margin-bottom: 16rpx;
  line-height: 1.5;
}

.download-badge {
  display: flex;
  gap: 12rpx;
  margin-bottom: 16rpx;
  flex-wrap: wrap;
}

.badge-item {
  display: inline-flex;
  align-items: center;
  gap: 6rpx;
  background: rgba(255, 255, 255, 0.9);
  padding: 6rpx 16rpx;
  border-radius: 20rpx;
  font-size: 20rpx;
  color: #d46b08;
  font-weight: 500;
}

.badge-dot {
  width: 8rpx;
  height: 8rpx;
  background-color: #ffc53d;
  border-radius: 50%;
}

.download-btn-wrapper {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #ffc53d 0%, #ffa940 100%);
  color: #fff;
  padding: 14rpx 36rpx;
  border-radius: 40rpx;
  font-size: 26rpx;
  font-weight: 600;
  box-shadow: 0 6rpx 16rpx rgba(255, 197, 61, 0.35);
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
