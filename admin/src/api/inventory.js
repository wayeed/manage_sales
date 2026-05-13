import { get, post, put } from './request'

export const getStockList = (params) => get('/inventory/stocks', params)
export const getStockAlerts = (params) => get('/stock-alerts', params)
export const handleAlert = (id, data) => post(`/stock-alerts/${id}/handle`, data)
export const getInventoryTransactions = (params) => get('/inventory/transactions', params)
