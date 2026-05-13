import { get } from './request'

/**
 * 获取SKU列表（支持搜索）
 * @param {Object} params - { keyword, page, page_size }
 */
export const getSkuList = (params) => get('/skus', params)

/**
 * 获取带库存的SKU列表（用于订单选商品）
 * @param {Object} params - { store_id, keyword, page, page_size }
 */
export const getSkuListWithStock = (params) => get('/skus/with-stock', params)

/**
 * 获取商品SKU详情
 * @param {number} id - SKU ID
 */
export const getSkuDetail = (id) => get(`/skus/${id}`)
