import { get, post, put, del } from './request'

export const getSupplierList = (params) => get('/suppliers', params)
export const createSupplier = (data) => post('/suppliers', data)
export const updateSupplier = (id, data) => put(`/suppliers/${id}`, data)
export const deleteSupplier = (id) => del(`/suppliers/${id}`)
