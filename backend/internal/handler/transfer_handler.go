package handler

import (
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// TransferHandler 调拨处理器
type TransferHandler struct {
	transferService *service.TransferService
}

// NewTransferHandler 创建调拨处理器实例
func NewTransferHandler(transferService *service.TransferService) *TransferHandler {
	return &TransferHandler{transferService: transferService}
}

// CreateOrder 创建调拨单
// POST /api/transfers
func (h *TransferHandler) CreateOrder(c *gin.Context) {
	var req service.CreateTransferOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	createdBy := GetUserID(c)
	if err := h.transferService.CreateOrder(&req, createdBy); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "创建调拨单失败")
		return
	}

	Success(c, nil)
}

// List 获取调拨单列表
// GET /api/transfers?page=1&page_size=10&store_id=1&status=0
func (h *TransferHandler) List(c *gin.Context) {
	var req service.ListTransferOrderRequest
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

	result, err := h.transferService.List(&req)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询调拨单列表失败")
		return
	}

	PageResponse(c, result.List, result.Total, result.Page, result.PageSize)
}

// GetDetail 获取调拨单详情
// GET /api/transfers/:id
func (h *TransferHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的调拨单ID")
		return
	}

	order, err := h.transferService.GetDetail(id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "获取调拨单详情失败")
		return
	}

	Success(c, order)
}

// ApproveOrder 审核调拨单
// PUT /api/transfers/:id/approve
func (h *TransferHandler) ApproveOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的调拨单ID")
		return
	}

	auditedBy := GetUserID(c)
	if err := h.transferService.ApproveOrder(id, auditedBy); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "审核调拨单失败")
		return
	}

	Success(c, nil)
}

// ConfirmOut 确认出库
// PUT /api/transfers/:id/out
func (h *TransferHandler) ConfirmOut(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的调拨单ID")
		return
	}

	createdBy := GetUserID(c)
	if err := h.transferService.ConfirmOut(id, createdBy); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "确认出库失败")
		return
	}

	Success(c, nil)
}

// ConfirmIn 确认入库
// PUT /api/transfers/:id/in
func (h *TransferHandler) ConfirmIn(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的调拨单ID")
		return
	}

	createdBy := GetUserID(c)
	if err := h.transferService.ConfirmIn(id, createdBy); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "确认入库失败")
		return
	}

	Success(c, nil)
}

// CancelOrder 取消调拨单
// PUT /api/transfers/:id/cancel
func (h *TransferHandler) CancelOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的调拨单ID")
		return
	}

	if err := h.transferService.CancelOrder(id); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "取消调拨单失败")
		return
	}

	Success(c, nil)
}
