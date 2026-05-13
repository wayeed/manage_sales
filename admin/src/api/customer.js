import { get, post, put, del } from './request'

export const getCustomerList = (params) => get('/customers', params)
export const getCustomerDetail = (id) => get(`/customers/${id}`)
export const createCustomer = (data) => post('/customers', data)
export const updateCustomer = (id, data) => put(`/customers/${id}`, data)
export const deleteCustomer = (id) => del(`/customers/${id}`)
export const getFollowUps = (customerId) => get(`/customers/${customerId}/follow-ups`)
export const addFollowUp = (customerId, data) => post(`/customers/${customerId}/follow-ups`, data)
