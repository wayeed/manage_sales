import { get, post } from './request'

export const createOutboundRequest = (data) => post('/outbound-requests', data)
export const getOutboundRequestByOrder = (orderId) => get(`/outbound-requests/order/${orderId}`)
export const supervisorApprove = (id, data) => post(`/outbound-requests/${id}/supervisor-approve`, data)
export const financeApprove = (id, data) => post(`/outbound-requests/${id}/finance-approve`, data)
export const rejectOutboundRequest = (id, data) => post(`/outbound-requests/${id}/reject`, data)
export const getPendingOutboundRequests = (params) => get('/outbound-requests/pending', params)
