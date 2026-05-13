import { get, post, put } from './request'

export const getConfigList = () => get('/configs')
export const getConfig = (key) => get(`/configs/${key}`)
export const updateConfig = (key, data) => put(`/configs/${key}`, data)
export const batchUpdateConfig = (data) => post('/configs/batch', data)
