<script>
import { checkUpdate } from './utils/update'

export default {
  onLaunch() {
    console.log('App Launch')

    // 登录白名单（无需登录即可访问的页面）
    const whiteList = ['/pages/login/index']

    // 检查是否已登录
    const isLoggedIn = () => !!uni.getStorageSync('token')

    // 保存原始路由方法
    const originalNavigateTo = uni.navigateTo
    const originalSwitchTab = uni.switchTab
    const originalReLaunch = uni.reLaunch
    const originalRedirectTo = uni.redirectTo

    // 拦截 navigateTo
    uni.navigateTo = function(options) {
      const url = (options.url || '').split('?')[0]
      if (!isLoggedIn() && !whiteList.includes(url)) {
        originalReLaunch({ url: '/pages/login/index' })
        return
      }
      originalNavigateTo(options)
    }

    // 拦截 switchTab（tabBar 页面切换）
    uni.switchTab = function(options) {
      const url = (options.url || '').split('?')[0]
      if (!isLoggedIn() && !whiteList.includes(url)) {
        originalReLaunch({ url: '/pages/login/index' })
        return
      }
      originalSwitchTab(options)
    }

    // 拦截 reLaunch
    uni.reLaunch = function(options) {
      const url = (options.url || '').split('?')[0]
      if (!isLoggedIn() && !whiteList.includes(url)) {
        originalReLaunch({ url: '/pages/login/index' })
        return
      }
      originalReLaunch(options)
    }

    // 拦截 redirectTo
    uni.redirectTo = function(options) {
      const url = (options.url || '').split('?')[0]
      if (!isLoggedIn() && !whiteList.includes(url)) {
        originalReLaunch({ url: '/pages/login/index' })
        return
      }
      originalRedirectTo(options)
    }

    // 启动时检查：未登录则跳转登录页
    if (!isLoggedIn()) {
      // 延迟执行，确保页面栈已初始化
      setTimeout(() => {
        const pages = getCurrentPages()
        const currentPage = pages.length > 0 ? pages[pages.length - 1] : null
        if (!currentPage || currentPage.route !== 'pages/login/index') {
          uni.reLaunch({ url: '/pages/login/index' })
        }
      }, 0)
    }

    // APP端启动时自动检查更新（静默模式，仅在有新版本时提示）
    // #ifdef APP-PLUS
    setTimeout(() => {
      checkUpdate({ silent: true })
    }, 3000) // 延迟3秒，避免影响启动速度
    // #endif
  }
}
</script>

<style>
@import './styles/common.scss';
</style>
