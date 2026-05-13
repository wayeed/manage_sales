import { get, post, put, del } from './request'

export const getOrderList = (params) => get('/orders', params)
export const getOrderDetail = (id) => get(`/orders/${id}`)
export const createOrder = (data) => post('/orders', data)
export const approveOrder = (id, data) => post(`/orders/${id}/approve`, data)
export const cancelOrder = (id) => post(`/orders/${id}/cancel`)
export const returnOrder = (id, data) => post(`/orders/${id}/return`, data)
export const getOrderFeed = (params) => get('/orders/feed', params)
export const exportOrders = (params) => get('/orders/export', params)
