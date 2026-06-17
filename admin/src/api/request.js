import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getToken, removeToken } from '@/utils/auth'
import router from '@/router'

// 创建 axios 实例
const service = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true,
})

// 从cookie中获取CSRF token
const getCsrfToken = () => {
  const cookies = document.cookie.split(';')
  for (const cookie of cookies) {
    const [name, value] = cookie.trim().split('=')
    if (name === 'csrf_token') {
      return value
    }
  }
  return ''
}

// 初始化CSRF token
export const initCsrfToken = async () => {
  try {
    // service已配置baseURL包含/api，所以这里只需/csrf-token
    await service.get('/csrf-token')
  } catch (error) {
    console.error('获取CSRF token失败:', error)
  }
}

// 请求拦截器
service.interceptors.request.use(
  (config) => {
    const token = getToken()
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    
    // 非GET请求需要添加CSRF token
    if (config.method && config.method.toLowerCase() !== 'get') {
      const csrfToken = getCsrfToken()
      console.log('[CSRF] 请求方法:', config.method, 'URL:', config.url, 'token:', csrfToken ? '已获取' : '未获取')
      if (csrfToken) {
        config.headers['X-CSRF-Token'] = csrfToken
      } else {
        console.warn('[CSRF] 警告：CSRF token为空，尝试重新获取')
        // 尝试重新获取CSRF token
        initCsrfToken().then(() => {
          const newToken = getCsrfToken()
          console.log('[CSRF] 重新获取token:', newToken ? '成功' : '失败')
        })
      }
    }
    return config
  },
  (error) => {
    console.error('请求错误:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
service.interceptors.response.use(
  (response) => {
    const res = response.data
    // 如果返回的状态码不是 200 或 0（成功），说明接口有问题
    if (res.code !== undefined && res.code !== null && res.code !== 200 && res.code !== 0) {
      ElMessage({
        message: res.message || '请求失败',
        type: 'error',
        duration: 3000,
      })

      // 401: Token 过期或未登录
      if (res.code === 401) {
        ElMessageBox.confirm('登录已过期，请重新登录', '提示', {
          confirmButtonText: '重新登录',
          cancelButtonText: '取消',
          type: 'warning',
        }).then(() => {
          removeToken()
          router.push('/login')
        })
      }
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    return res
  },
  (error) => {
    const { response } = error
    if (response) {
      switch (response.status) {
        case 401:
          ElMessage.error('登录已过期，请重新登录')
          removeToken()
          router.push('/login')
          break
        case 403:
          ElMessage.error('没有权限访问该资源')
          break
        case 404:
          ElMessage.error('请求的资源不存在')
          break
        case 500:
          ElMessage.error('服务器内部错误，请稍后重试')
          break
        default:
          ElMessage.error(response.data?.message || `请求失败(${response.status})`)
      }
    } else if (error.code === 'ECONNABORTED') {
      ElMessage.error('请求超时，请稍后重试')
    } else {
      ElMessage.error('网络连接异常，请检查网络')
    }
    return Promise.reject(error)
  }
)

// 导出请求方法
export const get = (url, params, config = {}) => {
  return service({
    method: 'get',
    url,
    params,
    ...config,
  })
}

export const post = (url, data, config = {}) => {
  return service({
    method: 'post',
    url,
    data,
    ...config,
  })
}

export const put = (url, data, config = {}) => {
  return service({
    method: 'put',
    url,
    data,
    ...config,
  })
}

export const del = (url, params, config = {}) => {
  return service({
    method: 'delete',
    url,
    params,
    ...config,
  })
}

export default service
