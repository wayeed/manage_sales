import { get, post } from './request'

/**
 * 申请跟进
 * @param {Object} data - { customer_id, remark }
 */
export const applyFollowUp = (data) => post('/follow-up-approvals', data)

/**
 * 查询审批状态
 * @param {number} id - 申请ID
 */
export const getApprovalStatus = (id) => get(`/follow-up-approvals/${id}`)

/**
 * 查询我的申请列表
 * @param {Object} params - { page, page_size }
 */
export const getMyApplications = (params) => get('/follow-up-approvals/my', params)

/**
 * 查询待我审批的列表
 * @param {Object} params - { page, page_size }
 */
export const getPendingApprovals = (params) => get('/follow-up-approvals/pending', params)

/**
 * 审批通过
 * @param {number} id - 申请ID
 */
export const approveFollowUp = (id) => post(`/follow-up-approvals/${id}/approve`)

/**
 * 审批拒绝
 * @param {number} id - 申请ID
 * @param {string} reason - 拒绝原因
 */
export const rejectFollowUp = (id, reason) => post(`/follow-up-approvals/${id}/reject`, { reason })

/**
 * 撤回申请
 * @param {number} id - 申请ID
 */
export const cancelFollowUp = (id) => post(`/follow-up-approvals/${id}/cancel`)
