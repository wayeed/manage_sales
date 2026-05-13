import { get, post } from './request'

export const getCommissionList = (params) => get('/commissions', params)
export const getCommissionSummary = (params) => get('/commissions/summary', params)
export const manualAdjust = (id, data) => post(`/commissions/${id}/adjust`, data)
export const triggerCalculate = (orderId) => post('/commissions/calculate', { order_id: orderId })
