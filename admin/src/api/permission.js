import { get } from './request'

/**
 * 获取权限列表
 */
export const getPermissionList = () => get('/permissions')

/**
 * 获取权限树
 */
export const getPermissionTree = () => get('/permissions/tree')
