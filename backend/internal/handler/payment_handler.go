package handler

import (
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// ApprovePaymentRequest 审核回款请求
type ApprovePaymentRequest struct {
	Approved     bool   `json:"approved" example:"true"`
	RejectReason string `json:"reject_reason" example:"驳回原因"`
}

// PaymentHandler 回款处理器
type PaymentHandler struct {
	paymentService *service.PaymentService
}

// NewPaymentHandler 创建回款处理器实例
func NewPaymentHandler(paymentService *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

// CreatePayment 录入回款
// POST /api/payments
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req service.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误: "+err.Error())
		return
	}

	createdBy := GetUserID(c)
	roleCodes := GetRoleCodes(c)

	if err := h.paymentService.CreatePayment(&req, createdBy, roleCodes); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "录入回款失败")
		return
	}

	Success(c, nil)
}

// List 获取回款列表
// GET /api/payments?order_id=1&status=0&page=1&page_size=10
func (h *PaymentHandler) List(c *gin.Context) {
	var req service.ListPaymentRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	result, err := h.paymentService.List(&req)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询回款列表失败")
		return
	}

	PageResponse(c, result.List, result.Total, result.Page, result.PageSize)
}

// ApprovePayment 审核回款
// POST /api/payments/:id/approve
func (h *PaymentHandler) ApprovePayment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的回款ID")
		return
	}

	var req ApprovePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	approvedBy := GetUserID(c)

	if err := h.paymentService.ApprovePayment(id, approvedBy, req.Approved, req.RejectReason); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "审核回款失败")
		return
	}

	Success(c, nil)
}
