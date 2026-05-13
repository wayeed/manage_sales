import { get, post } from './request'

export const getFundPoolList = (params) => get('/fund-pools', params)
export const getFundPoolShares = (id) => get(`/fund-pools/${id}/shares`)
export const settleFundPool = (data) => post('/fund-pools/settle', data)
