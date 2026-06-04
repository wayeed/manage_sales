import { get } from './request'

/**
 * 获取礼品列表
 * @param {Object} params - 查询参数
 * @param {number} params.page - 页码
 * @param {number} params.page_size - 每页数量
 * @param {string} params.keyword - 搜索关键词
 * @param {number} params.status - 状态筛选
 * @returns {Promise}
 */
export const getGiftList = (params) => get('/gifts', params)

/**
 * 获取礼品详情
 * @param {number} id - 礼品ID
 * @returns {Promise}
 */
export const getGiftDetail = (id) => get(`/gifts/${id}`)
