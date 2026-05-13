import { get, post, put, del } from './request'

/**
 * 获取用户列表
 * @param {Object} params - { page, page_size, keyword, role_id, status }
 */
export const getUserList = (params) => get('/users', params)

/**
 * 获取用户详情
 * @param {number} id
 */
export const getUserDetail = (id) => get(`/users/${id}`)

/**
 * 创建用户
 * @param {Object} data - { name, phone, password, role_ids, status }
 */
export const createUser = (data) => post('/users', data)

/**
 * 更新用户
 * @param {number} id
 * @param {Object} data - { name, phone, role_ids, status }
 */
export const updateUser = (id, data) => put(`/users/${id}`, data)

/**
 * 删除用户
 * @param {number} id
 */
export const deleteUser = (id) => del(`/users/${id}`)

/**
 * 重置密码
 * @param {number} id
 */
export const resetPassword = (id) => post(`/users/${id}/reset-password`)

/**
 * 分配角色
 * @param {number} id
 * @param {Array} roleIds
 */
export const assignRole = (id, roleIds) => post(`/users/${id}/roles`, { role_ids: roleIds })

/**
 * 更新用户状态（启用/禁用）
 * @param {number} id
 * @param {number} status - 0: 禁用, 1: 启用
 */
export const updateUserStatus = (id, status) => put(`/users/${id}/status`, { status })
