import { get, post } from './request'

export const getOrders = (params) => get('/orders', params)
export const getOrderDetail = (id) => get(`/orders/${id}`)
export const createOrder = (data) => post('/orders', data)
export const approveOrder = (id, data) => post(`/orders/${id}/approve`, data)
