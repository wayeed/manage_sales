import { get, post, put, del } from './request'

// 获取APP版本列表
export const getAppVersionList = (params) => get('/app-versions', params)

// 获取APP版本详情
export const getAppVersionById = (id) => get(`/app-versions/${id}`)

// 创建APP版本
export const createAppVersion = (data) => post('/app-versions', data)

// 更新APP版本
export const updateAppVersion = (id, data) => put(`/app-versions/${id}`, data)

// 删除APP版本
export const deleteAppVersion = (id) => del(`/app-versions/${id}`)

// 获取最新版本（APP端调用）
export const getLatestAppVersion = (platform = 'android') => get('/app-versions/latest', { platform })
