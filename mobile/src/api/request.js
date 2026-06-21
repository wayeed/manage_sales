// 后端API基础地址
// 从配置文件读取 API 配置
import { API_CONFIG } from './config'

const BASE_URL = API_CONFIG.BASE_URL

export { BASE_URL }

const request = (options) => {
  return new Promise((resolve, reject) => {
    const token = uni.getStorageSync('token')
    const csrfToken = uni.getStorageSync('csrf_token')
    const headers = {
      'Content-Type': 'application/json',
      'Authorization': token ? `Bearer ${token}` : ''
    }
    // 非GET请求携带CSRF token
    if ((options.method || 'GET') !== 'GET' && csrfToken) {
      headers['X-CSRF-Token'] = csrfToken
    }
    uni.request({
      url: BASE_URL + options.url,
      method: options.method || 'GET',
      data: options.data || {},
      header: headers,
      success: (res) => {
        // 处理 HTTP 状态码
        if (res.statusCode === 401) {
          uni.removeStorageSync('token')
          uni.reLaunch({ url: '/pages/login/index' })
          reject({ code: 401, message: '登录已过期，请重新登录' })
          return
        }

        if (res.statusCode === 403) {
          uni.showToast({ title: '权限不足，无法执行此操作', icon: 'none' })
          reject({ code: 403, message: '权限不足' })
          return
        }

        if (res.statusCode !== 200) {
          uni.showToast({ title: `请求失败(${res.statusCode})`, icon: 'none' })
          reject({ code: res.statusCode, message: '请求失败' })
          return
        }

        // 处理业务状态码
        if (res.data.code === 200) {
          resolve(res.data)
        } else if (res.data.code === 401) {
          uni.removeStorageSync('token')
          uni.reLaunch({ url: '/pages/login/index' })
          reject(res.data)
        } else {
          uni.showToast({ title: res.data.message || '请求失败', icon: 'none' })
          reject(res.data)
        }
      },
      fail: (err) => {
        uni.showToast({ title: '网络连接失败', icon: 'none' })
        reject(err)
      }
    })
  })
}

export const get = (url, data) => request({ url, method: 'GET', data })
export const post = (url, data) => request({ url, method: 'POST', data })
export const put = (url, data) => request({ url, method: 'PUT', data })
export const del = (url, data) => request({ url, method: 'DELETE', data })

export default request
