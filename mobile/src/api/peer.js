import { get } from './request'

/**
 * 获取同行列表
 * @param {Object} params - { keyword, page, page_size }
 */
export const getPeerList = (params) => get('/peers', params)

/**
 * 获取同行详情
 * @param {number} id - 同行ID
 */
export const getPeerDetail = (id) => get(`/peers/${id}`)
