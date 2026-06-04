import { get, post, put } from './request'

/**
 * 获取客户列表
 * @param {Object} params - { keyword, page, page_size }
 */
export const getCustomerList = (params) => get('/customers', params)

/**
 * 获取客户列表（含草稿状态）
 * @param {Object} params - { keyword, page, page_size }
 */
export const getCustomersWithDraftStatus = (params) => get('/customers/with-draft-status', params)

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
export const updateCustomer = (id, data) => put(`/customers/${id}`, data)

/**
 * 获取客户跟进记录
 * @param {number} customerId - 客户ID
 */
export const getCustomerFollowUps = (customerId) => get(`/customers/${customerId}/follow-ups`)

/**
 * 添加跟进记录
 * @param {number} customerId - 客户ID
 * @param {Object} data - { content, follow_type, next_follow_date }
 */
export const addCustomerFollowUp = (customerId, data) => post(`/customers/${customerId}/follow-ups`, data)
