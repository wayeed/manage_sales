import { get, post, put } from './request'

/**
 * 创建送货出库记录
 * @param {Object} data - { order_id, warehouse_id, print_mode, ... }
 * @returns {Promise}
 */
export const createDelivery = (data) => post('/deliveries', data)

/**
 * 确认送达
 * @param {number} orderId - 订单ID
 * @returns {Promise}
 */
export const confirmDelivery = (orderId) => put(`/deliveries/${orderId}/confirm`, {})

/**
 * 获取送货列表
 * @param {Object} params - 查询参数
 * @returns {Promise}
 */
export const getDeliveryList = (params) => get('/deliveries', params)

/**
 * 获取送货详情
 * @param {number} id - 送货记录ID
 * @returns {Promise}
 */
export const getDeliveryDetail = (id) => get(`/deliveries/${id}`)

/**
 * 作废送货记录
 * @param {number} id - 送货记录ID
 * @param {string} remark - 作废原因
 * @returns {Promise}
 */
export const cancelDelivery = (id, remark) => put(`/deliveries/${id}/cancel`, { remark })

/**
 * 获取待送货订单列表
 * @param {Object} params - 查询参数
 * @returns {Promise}
 */
export const getPendingDeliveryOrders = (params) => get('/deliveries/orders/pending', params)

/**
 * 获取订单库存状态（含锁定批次信息）
 * @param {number} orderId - 订单ID
 * @param {number} warehouseId - 仓库ID
 * @returns {Promise}
 */
export const getOrderStockStatus = (orderId, warehouseId) => get(`/deliveries/stock-status`, { order_id: orderId, warehouse_id: warehouseId })
export const printDelivery = (orderId) => post(`/deliveries/${orderId}/print`)
