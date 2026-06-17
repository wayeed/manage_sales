import { defineStore } from 'pinia'
import { ref } from 'vue'
import { login as loginApi, logout as logoutApi, getUserInfo } from '../api/auth'

export const useUserStore = defineStore('user', () => {
  const token = ref(uni.getStorageSync('token') || '')
  const userInfo = ref(null)

  const setToken = (val) => {
    token.value = val
    uni.setStorageSync('token', val)
  }

  const login = async (data) => {
    const res = await loginApi(data)
    // 从登录响应中提取token和csrf_token
    const respData = res.data || {}
    setToken(respData.token)
    // 保存CSRF token到storage
    if (respData.csrf_token) {
      uni.setStorageSync('csrf_token', respData.csrf_token)
    }
    return res
  }

  const fetchUserInfo = async () => {
    const res = await getUserInfo()
    userInfo.value = res.data
    return res.data
  }

  const logout = async () => {
    try {
      await logoutApi()
    } catch (e) {
      // ignore
    }
    token.value = ''
    userInfo.value = null
    uni.removeStorageSync('token')
    uni.reLaunch({ url: '/pages/login/index' })
  }

  const isLoggedIn = () => !!token.value

  return {
    token,
    userInfo,
    setToken,
    login,
    fetchUserInfo,
    logout,
    isLoggedIn
  }
})
