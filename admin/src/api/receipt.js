import { get, post, put } from './request'

export const getReceiptList = (params) => get('/receipts', params)
export const getReceiptDetail = (id) => get(`/receipts/${id}`)
export const createReceipt = (data) => post('/receipts', data)
export const approveReceipt = (id) => put(`/receipts/${id}/approve`)
export const receiveReceipt = (id, data) => put(`/receipts/${id}/receive`, data)
export const cancelReceipt = (id) => put(`/receipts/${id}/cancel`)