import { get, post, put, del } from './request'

export const getOrders = (params) => get('/orders', params)
export const getOrderDetail = (id) => get(`/orders/${id}`)
export const createOrder = (data) => post('/orders', data)
export const updateOrder = (id, data) => put(`/orders/${id}`, data)
export const deleteOrder = (id) => del(`/orders/${id}`)
export const approveOrder = (id, data) => post(`/orders/${id}/approve`, data)
export const getOrderCommission = (id) => get(`/orders/${id}/commission`)

// 获取客户最新草稿订单
export const getCustomerDraft = (customerId) => get('/orders/customer-draft', { customer_id: customerId })
