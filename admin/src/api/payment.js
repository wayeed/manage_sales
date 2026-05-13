import { get, post } from './request'

export const getPaymentList = (params) => get('/payments', params)
export const createPayment = (data) => post('/payments', data)
export const approvePayment = (id, data) => post(`/payments/${id}/approve`, data)
