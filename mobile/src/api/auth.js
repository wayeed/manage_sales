import { post, get } from './request'

export const login = (data) => post('/login', data)
export const logout = () => post('/logout')
export const getUserInfo = () => get('/users/me')
