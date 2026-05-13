import { get, post, put, del } from './request'

/**
 * 获取角色列表
 */
export const getRoleList = () => get('/roles')

/**
 * 获取角色详情
 * @param {number} id
 */
export const getRoleDetail = (id) => get(`/roles/${id}`)

/**
 * 创建角色
 * @param {Object} data - { name, code, type, description }
 */
export const createRole = (data) => post('/roles', data)

/**
 * 更新角色
 * @param {number} id
 * @param {Object} data - { name, code, type, description }
 */
export const updateRole = (id, data) => put(`/roles/${id}`, data)

/**
 * 删除角色
 * @param {number} id
 */
export const deleteRole = (id) => del(`/roles/${id}`)

/**
 * 分配权限
 * @param {number} id
 * @param {Array} permIds
 */
export const assignPermissions = (id, permIds) => post(`/roles/${id}/permissions`, { permission_ids: permIds })

/**
 * 获取角色权限列表
 * @param {number} id
 */
export const getRolePermissions = (id) => get(`/roles/${id}/permissions`)
