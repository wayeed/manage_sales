import { get, post } from './request'

/**
 * 获取客户列表
 * @param {Object} params - { keyword, page, page_size }
 */
export const getCustomerList = (params) => get('/customers', params)

/**
 * 获取客户详情
 * @param {number} id - 客户ID
 */
export const getCustomerDetail = (id) => get(`/customers/${id}`)

/**
 * 新增客户
 * @param {Object} data - { customer_name, phone, address }
 */
export const createCustomer = (data) => post('/customers', data)

/**
 * 更新客户
 * @param {number} id - 客户ID
 * @param {Object} data - 更新数据
 */
export const updateCustomer = (id, data) => post(`/customers/${id}`, data)
