package handler

import (
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// RejectApprovalRequest 审批拒绝请求
type RejectApprovalRequest struct {
	Reason string `json:"reason" example:"不符合要求"`
}

// FollowUpApprovalHandler 申请跟进审批处理器
type FollowUpApprovalHandler struct {
	approvalService *service.FollowUpApprovalService
}

// NewFollowUpApprovalHandler 创建实例
func NewFollowUpApprovalHandler(approvalService *service.FollowUpApprovalService) *FollowUpApprovalHandler {
	return &FollowUpApprovalHandler{approvalService: approvalService}
}

// Create 创建申请
// POST /api/follow-up-approvals
func (h *FollowUpApprovalHandler) Create(c *gin.Context) {
	uid := GetUserID(c)
	sid := GetStoreID(c)

	var req service.CreateFollowUpApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	approval, err := h.approvalService.Create(&req, uid, sid)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "创建申请失败: "+err.Error())
		return
	}

	Success(c, approval)
}

// GetStatus 查询审批状态
// GET /api/follow-up-approvals/:id
func (h *FollowUpApprovalHandler) GetStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的申请ID")
		return
	}

	approval, err := h.approvalService.GetStatus(id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询失败")
		return
	}

	Success(c, approval)
}

// Approve 审批通过
// POST /api/follow-up-approvals/:id/approve
func (h *FollowUpApprovalHandler) Approve(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的申请ID")
		return
	}

	uid := GetUserID(c)

	if err := h.approvalService.Approve(id, uid); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "审批失败: "+err.Error())
		return
	}

	Success(c, nil)
}

// Reject 审批拒绝
// POST /api/follow-up-approvals/:id/reject
func (h *FollowUpApprovalHandler) Reject(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的申请ID")
		return
	}

	var req RejectApprovalRequest
	c.ShouldBindJSON(&req)

	uid := GetUserID(c)

	if err := h.approvalService.Reject(id, uid, req.Reason); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "审批失败: "+err.Error())
		return
	}

	Success(c, nil)
}

// Cancel 撤回申请
// POST /api/follow-up-approvals/:id/cancel
func (h *FollowUpApprovalHandler) Cancel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的申请ID")
		return
	}

	uid := GetUserID(c)

	if err := h.approvalService.Cancel(id, uid); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "撤回失败: "+err.Error())
		return
	}

	Success(c, nil)
}

// ListMyApplications 查询我的申请列表
// GET /api/follow-up-approvals/my
func (h *FollowUpApprovalHandler) ListMyApplications(c *gin.Context) {
	uid := GetUserID(c)

	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	list, total, err := h.approvalService.ListByApplicant(uid, page, pageSize)
	if err != nil {
		Error(c, 500, "查询失败")
		return
	}

	PageResponse(c, list, total, page, pageSize)
}

// ListPendingApprovals 查询待我审批的列表
// GET /api/follow-up-approvals/pending
func (h *FollowUpApprovalHandler) ListPendingApprovals(c *gin.Context) {
	uid := GetUserID(c)

	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	list, total, err := h.approvalService.ListByApprover(uid, page, pageSize)
	if err != nil {
		Error(c, 500, "查询失败")
		return
	}

	PageResponse(c, list, total, page, pageSize)
}
