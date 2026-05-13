import { get, post, put, del } from './request'

export const getTransferList = (params) => get('/transfers', params)
export const createTransfer = (data) => post('/transfers', data)
export const approveTransfer = (id) => put(`/transfers/${id}/approve`)
export const confirmOut = (id) => put(`/transfers/${id}/out`)
export const confirmIn = (id) => put(`/transfers/${id}/in`)
export const cancelTransfer = (id) => put(`/transfers/${id}/cancel`)
