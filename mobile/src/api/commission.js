import { get, post } from './request'

/**
 * 预估提成
 * @param {Object} data - { items: [{ list_price, sale_price, quantity }], is_peer_order }
 */
export const estimateCommission = (data) => post('/commissions/estimate', data)

/**
 * 获取提成列表
 * @param {Object} params - { page, page_size, period_value }
 */
export const getCommissionList = (params) => get('/commissions', params)
