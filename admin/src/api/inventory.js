import { get, post, put, del } from './request'

export const getStockList = (params) => get('/inventory/stocks', params)
export const getStockAlerts = (params) => get('/stock-alerts', params)
export const checkStockAlerts = () => post('/stock-alerts/check')
export const handleAlert = (id, data) => post(`/stock-alerts/${id}/handle`, data)
export const getInventoryTransactions = (params) => get('/inventory/transactions', params)
export const getWarehouseList = (params) => get('/warehouses', params)

// 库存盘点
export const getStocktakeList = (params) => get('/stocktakes', params)
export const getStocktakeDetail = (id) => get(`/stocktakes/${id}`)
export const createStocktake = (data) => post('/stocktakes', data)
export const updateStocktake = (id, data) => put(`/stocktakes/${id}`, data)
export const submitStocktake = (id) => post(`/stocktakes/${id}/submit`)
export const approveStocktake = (id, data) => post(`/stocktakes/${id}/approve`, data)
export const deleteStocktake = (id) => del(`/stocktakes/${id}`)
