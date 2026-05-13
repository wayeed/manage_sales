import { get, post, put, del } from './request'

export const getWarehouseList = (params) => get('/warehouses', params)
export const createWarehouse = (data) => post('/warehouses', data)
export const updateWarehouse = (id, data) => put(`/warehouses/${id}`, data)
export const deleteWarehouse = (id) => del(`/warehouses/${id}`)
