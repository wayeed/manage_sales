import { get, post } from './request'

export const getSalaryList = (params) => get('/salaries', params)
export const getSalaryDetail = (id) => get(`/salaries/${id}`)
export const generateSalary = (data) => post('/salaries/generate', data)
export const confirmSalary = (id) => post(`/salaries/${id}/confirm`)
export const paySalary = (id, data) => post(`/salaries/${id}/pay`, data)
export const exportSalarySlip = (id) => get(`/salaries/${id}/export`, {}, { responseType: 'blob' })
