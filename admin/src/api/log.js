import { get } from './request'

export const getOperationLogs = (params) => get('/operation-logs', params)
