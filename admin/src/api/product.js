import { get, post, put, del, default as service } from './request'

export const getProductList = (params) => get('/products', params)
export const getProductDetail = (id) => get(`/products/${id}`)
export const createProduct = (data) => post('/products', data)
export const updateProduct = (id, data) => put(`/products/${id}`, data)
export const deleteProduct = (id) => del(`/products/${id}`)
export const updateProductStatus = (id, status) => put(`/products/${id}/status`, { status })
export const getSkuList = (productId) => get(`/products/${productId}/skus`)
export const getAllSkuList = (params) => get('/skus', params)
export const createSku = (productId, data) => post(`/products/${productId}/skus`, data)
export const updateSku = (id, data) => put(`/skus/${id}`, data)
export const deleteSku = (id) => del(`/skus/${id}`)

// 批量导入商品
export const importProducts = (file) => {
  const formData = new FormData()
  formData.append('file', file)
  return service({
    method: 'post',
    url: '/products/import',
    data: formData,
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 60000, // 导入可能较慢，超时设为60秒
  })
}

// 下载导入模板
export const downloadImportTemplate = () => {
  return service({
    method: 'get',
    url: '/products/import-template',
    responseType: 'blob',
  })
}
