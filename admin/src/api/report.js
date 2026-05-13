import { get } from './request'

export const getSalesSummary = (params) => get('/reports/sales/summary', params)
export const getSalesTrend = (params) => get('/reports/sales/trend', params)
export const getPerformanceRanking = (params) => get('/reports/sales/ranking', params)
export const getProfitAnalysis = (params) => get('/reports/profit/analysis', params)
export const getPaymentAnalysis = (params) => get('/reports/payment/analysis', params)
export const getInventoryAnalysis = (params) => get('/reports/inventory/analysis', params)
export const getCommissionAnalysis = (params) => get('/reports/commission/analysis', params)
