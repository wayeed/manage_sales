import { post, get } from './request'

/**
 * 用户登录
 * @param {Object} data - { phone, password }
 */
export function login(data) {
  return post('/login', data)
}

/**
 * 用户登出
 */
export function logout() {
  return post('/logout')
}

/**
 * 获取当前用户信息
 */
export function getUserInfo() {
  return get('/users/me')
}
