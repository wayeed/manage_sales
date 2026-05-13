import { get, post, put, del } from './request'

export const getPeerList = (params) => get('/peers', params)
export const createPeer = (data) => post('/peers', data)
export const updatePeer = (id, data) => put(`/peers/${id}`, data)
export const deletePeer = (id) => del(`/peers/${id}`)
