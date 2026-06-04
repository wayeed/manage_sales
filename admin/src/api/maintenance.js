import { get, post, del } from './request'

// 获取可清除的数据表列表
export const getDataTables = () => get('/maintenance/data-tables')

// 检查是否有10分钟内备份
export const checkRecentBackup = () => post('/maintenance/check-recent-backup')

// 获取备份列表
export const getBackupList = (params) => get('/backups', params)

// 创建备份
export const createBackup = (data) => post('/backups', data)

// 删除备份
export const deleteBackup = (id) => del(`/backups/${id}`)

// 还原备份
export const restoreBackup = (id) => post(`/backups/${id}/restore`)

// 清除业务数据
export const clearData = (data) => post('/maintenance/clear-data', data)
