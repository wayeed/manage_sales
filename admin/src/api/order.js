import { get, post, put, del } from './request'

export const getOrderList = (params) => get('/orders', params)
export const getOrderDetail = (id) => get(`/orders/${id}`)
export const createOrder = (data) => post('/orders', data)
export const approveOrder = (id, data) => post(`/orders/${id}/approve`, data)
export const cancelOrder = (id) => post(`/orders/${id}/cancel`)
export const returnOrder = (id, data) => post(`/orders/${id}/return`, data)
export const getOrderFeed = (params) => get('/orders/feed', params)
export const exportOrders = (params) => get('/orders/export', params)
export const createPrintApproval = (id) => post(`/orders/${id}/print-approval`)
export const getPrintApprovalStatus = (id) => get(`/orders/${id}/print-approval`)
// 从订单生成采购单
export const generatePurchaseFromOrder = (id, data) => post(`/orders/${id}/generate-purchase`, data)
