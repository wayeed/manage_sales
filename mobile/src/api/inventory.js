import { get } from './request'

export const getStockList = (params) => get('/inventory/stocks', params)
