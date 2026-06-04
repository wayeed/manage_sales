import { get } from './request'

// 正向穿透：订单 → 源头
export const forwardTrace = (orderNo) => get('/inventory/trace/forward', { order_no: orderNo })

// 反向穿透：批次 → 去向
export const backwardTrace = (batchNo) => get('/inventory/trace/backward', { batch_no: batchNo })

// SKU库存全景
export const skuBatchTrace = (skuId) => get('/inventory/trace/sku', { sku_id: skuId })
