import { get, post, put } from './request'

export const getPurchaseList = (params) => get('/purchases', params)
export const getPurchaseDetail = (id) => get(`/purchases/${id}`)
export const createPurchase = (data) => post('/purchases', data)
export const updatePurchase = (id, data) => put(`/purchases/${id}`, data)
export const approvePurchase = (id) => put(`/purchases/${id}/approve`)
export const confirmReceipt = (id, data) => put(`/purchases/${id}/receipt`, data)
export const cancelPurchase = (id) => put(`/purchases/${id}/cancel`)
export const getMergeablePurchases = (params) => get('/purchases/mergeable', params)
export const appendPurchaseItems = (id, data) => post(`/purchases/${id}/items`, data)
