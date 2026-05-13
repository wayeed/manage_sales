import { get } from './request'

export const getOverview = () => get('/dashboard/overview')
export const getPerformance = (params) => get('/performance/overview', params)
export const getCommissions = (params) => get('/commissions', params)
