import { get, post, put } from './request'

export const getUserProfile = () => get('/users/me')
export const updateProfile = (data) => put('/users/profile', data)
export const changePassword = (data) => post('/users/change-password', data)
