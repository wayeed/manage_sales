import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from '@/utils/auth'
import { useUserStore } from '@/store/user'
import { usePermissionStore } from '@/store/permission'
import { constantRoutes } from '@/store/permission'

// 创建路由实例
const router = createRouter({
  history: createWebHistory(),
  routes: constantRoutes,
  scrollBehavior: () => ({ top: 0 }),
})

// 白名单路由（不需要登录即可访问）
const whiteList = ['/login']

// 路由守卫
router.beforeEach(async (to, from, next) => {
  // 设置页面标题
  document.title = to.meta?.title
    ? `${to.meta.title} - ${import.meta.env.VITE_APP_TITLE}`
    : import.meta.env.VITE_APP_TITLE

  const token = getToken()

  if (token) {
    if (to.path === '/login') {
      // 已登录状态访问登录页，重定向到首页
      next({ path: '/' })
    } else {
      const userStore = useUserStore()
      if (userStore.roles.length > 0) {
        // 已获取用户信息，直接放行
        next()
      } else {
        try {
          // 获取用户信息
          await userStore.getUserInfo()
          // 根据角色生成路由
          const permissionStore = usePermissionStore()
          const accessRoutes = permissionStore.generateRoutes(userStore.roles)
          // 动态添加路由
          accessRoutes.forEach((route) => {
            router.addRoute(route)
          })
          // 使用 replace 确保导航记录正确
          next({ ...to, replace: true })
        } catch (error) {
          // 获取用户信息失败，清除 token 并跳转登录页
          userStore.resetState()
          next(`/login?redirect=${to.path}`)
        }
      }
    }
  } else {
    // 未登录
    if (whiteList.includes(to.path)) {
      next()
    } else {
      next(`/login?redirect=${to.path}`)
    }
  }
})

export default router
