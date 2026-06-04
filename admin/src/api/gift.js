import { get, post, put, del } from './request'

export const getGiftList = (params) => get('/gifts', params)
export const createGift = (data) => post('/gifts', data)
export const updateGift = (id, data) => put(`/gifts/${id}`, data)
export const deleteGift = (id) => del(`/gifts/${id}`)
export const addGiftStock = (id, data) => post(`/gifts/${id}/stock`, data)
