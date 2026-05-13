import { defineStore } from 'pinia'
import { getToken, setToken, removeToken } from '@/utils/auth'
import { login as loginApi, logout as logoutApi, getUserInfo as getUserInfoApi } from '@/api/auth'
import router from '@/router'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: getToken() || '',
    userInfo: {},
    roles: [],
    permissions: [],
  }),

  getters: {
    isLoggedIn: (state) => !!state.token,
    userName: (state) => state.userInfo?.real_name || state.userInfo?.username || '',
    userPhone: (state) => state.userInfo?.phone || '',
    userRoles: (state) => state.userInfo?.roles || [],
    hasRole: (state) => {
      return (role) => state.roles.includes(role)
    },
    hasPermission: (state) => {
      return (permission) => state.permissions.includes(permission)
    },
  },

  actions: {
    /**
     * 用户登录
     */
    async login(loginForm) {
      try {
        const res = await loginApi(loginForm)
        const { token } = res.data
        this.token = token
        setToken(token)
        return res
      } catch (error) {
        return Promise.reject(error)
      }
    },

    /**
     * 获取用户信息、角色、权限
     */
    async getUserInfo() {
      try {
        const res = await getUserInfoApi()
        // 后端返回的数据结构：{ user: {...}, roles: [...], permissions: [...] }
        const responseData = res.data || {}
        const user = responseData.user || responseData
        const roles = responseData.roles || user.roles || []
        const permissions = responseData.permissions || user.permissions || []
        this.userInfo = { ...user, roles }
        // 后端 roles 是对象数组，提取 role_code 或 role_name 作为字符串数组
        this.roles = Array.isArray(roles) ? roles.map(r => r.role_code || r.role_name || r.code || r.name || String(r.id)) : []
        this.permissions = Array.isArray(permissions) ? permissions.map(p => p.permission_code || p.code || p.name || String(p.id)) : []
        return res
      } catch (error) {
        return Promise.reject(error)
      }
    },

    /**
     * 用户登出
     */
    async logout() {
      try {
        await logoutApi()
      } catch (error) {
        // 即使接口报错也要清理本地状态
      } finally {
        this.resetState()
        router.push('/login')
      }
    },

    /**
     * 重置状态
     */
    resetState() {
      this.token = ''
      this.userInfo = {}
      this.roles = []
      this.permissions = []
      removeToken()
    },
  },
})
