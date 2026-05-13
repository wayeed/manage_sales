import { get, post, put, del } from './request'

export const getStoreList = () => get('/stores')
export const createStore = (data) => post('/stores', data)
export const updateStore = (id, data) => put(`/stores/${id}`, data)
export const deleteStore = (id) => del(`/stores/${id}`)
