package handler

import (
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// OutboundRequestHandler 出库申请处理器
type OutboundRequestHandler struct {
	outboundRequestService *service.OutboundRequestService
}

// NewOutboundRequestHandler 创建出库申请处理器
func NewOutboundRequestHandler(outboundRequestService *service.OutboundRequestService) *OutboundRequestHandler {
	return &OutboundRequestHandler{
		outboundRequestService: outboundRequestService,
	}
}

// createRequestReq 创建出库申请请求
type createRequestReq struct {
	OrderID int64 `json:"order_id" binding:"required"`
}

// approveReq 审批请求
type approveReq struct {
	Remark string `json:"remark"`
}

// CreateRequest 创建出库申请
// POST /outbound-requests
func (h *OutboundRequestHandler) CreateRequest(c *gin.Context) {
	var req createRequestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误: "+err.Error())
		return
	}

	// 从JWT获取applicantID
	userID, exists := c.Get("user_id")
	if !exists {
		Error(c, 401, "未登录")
		return
	}
	applicantID := int64(userID.(int64))

	result, err := h.outboundRequestService.CreateRequest(req.OrderID, applicantID)
	if err != nil {
		handleAppError(c, err)
		return
	}

	Success(c, result)
}

// GetByOrderID 根据订单ID查询出库申请
// GET /outbound-requests/order/:orderID
func (h *OutboundRequestHandler) GetByOrderID(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("orderID"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的订单ID")
		return
	}

	result, err := h.outboundRequestService.GetByOrderID(orderID)
	if err != nil {
		handleAppError(c, err)
		return
	}

	Success(c, result)
}

// SupervisorApprove 主管审批
// POST /outbound-requests/:id/supervisor-approve
func (h *OutboundRequestHandler) SupervisorApprove(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的ID")
		return
	}

	var req approveReq
	c.ShouldBindJSON(&req) // remark 可选

	userID, exists := c.Get("user_id")
	if !exists {
		Error(c, 401, "未登录")
		return
	}
	supervisorID := int64(userID.(int64))

	if err := h.outboundRequestService.SupervisorApprove(id, supervisorID, req.Remark); err != nil {
		handleAppError(c, err)
		return
	}

	Success(c, nil)
}

// FinanceApprove 财务审批
// POST /outbound-requests/:id/finance-approve
func (h *OutboundRequestHandler) FinanceApprove(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的ID")
		return
	}

	var req approveReq
	c.ShouldBindJSON(&req) // remark 可选

	userID, exists := c.Get("user_id")
	if !exists {
		Error(c, 401, "未登录")
		return
	}
	financeID := int64(userID.(int64))

	if err := h.outboundRequestService.FinanceApprove(id, financeID, req.Remark); err != nil {
		handleAppError(c, err)
		return
	}

	Success(c, nil)
}

// Reject 拒绝出库申请
// POST /outbound-requests/:id/reject
func (h *OutboundRequestHandler) Reject(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的ID")
		return
	}

	var req approveReq
	c.ShouldBindJSON(&req) // remark 可选

	userID, exists := c.Get("user_id")
	if !exists {
		Error(c, 401, "未登录")
		return
	}
	approverID := int64(userID.(int64))

	if err := h.outboundRequestService.Reject(id, approverID, req.Remark); err != nil {
		handleAppError(c, err)
		return
	}

	Success(c, nil)
}

// ListPending 查询待审批的出库申请列表
// GET /outbound-requests/pending
func (h *OutboundRequestHandler) ListPending(c *gin.Context) {
	role := c.Query("role")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	result, err := h.outboundRequestService.ListPending(role, page, pageSize)
	if err != nil {
		handleAppError(c, err)
		return
	}

	Success(c, result)
}

// handleAppError 处理 AppError 类型的错误
func handleAppError(c *gin.Context, err error) {
	if appErr, ok := err.(*service.AppError); ok {
		Error(c, appErr.Code, appErr.Message)
		return
	}
	Error(c, 500, err.Error())
}
