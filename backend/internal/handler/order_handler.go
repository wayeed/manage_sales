package handler

import (
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// ApproveOrderRequest 审核订单请求
type ApproveOrderRequest struct {
	Approved bool   `json:"approved" example:"true"`
	Remark   string `json:"remark" example:"审核通过"`
}

// ReturnOrderRequest 退货处理请求
type ReturnOrderRequest struct {
	ReturnAmount float64 `json:"return_amount" example:"100.00"`
	ReturnProfit float64 `json:"return_profit" example:"20.00"`
	Remark       string  `json:"remark" example:"客户退货"`
}

// OrderHandler 订单处理器
type OrderHandler struct {
	orderService *service.OrderService
}

// NewOrderHandler 创建订单处理器实例
func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// CreateOrder 创建订单
// POST /api/orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req service.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误: "+err.Error())
		return
	}

	createdBy := GetUserID(c)

	order, err := h.orderService.CreateOrder(&req, createdBy)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "创建订单失败")
		return
	}

	Success(c, order)
}

// List 获取订单列表
// GET /api/orders?page=1&page_size=10&store_id=1&order_status=0&keyword=
func (h *OrderHandler) List(c *gin.Context) {
	var req service.ListOrderRequest
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

	result, err := h.orderService.List(&req)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询订单列表失败")
		return
	}

	PageResponse(c, result.List, result.Total, result.Page, result.PageSize)
}

// GetDetail 获取订单详情
// GET /api/orders/:id
func (h *OrderHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的订单ID")
		return
	}

	detail, err := h.orderService.GetDetail(id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "获取订单详情失败")
		return
	}

	Success(c, detail)
}

// ApproveOrder 审核订单
// POST /api/orders/:id/approve
func (h *OrderHandler) ApproveOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的订单ID")
		return
	}

	var req ApproveOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	approvedBy := GetUserID(c)

	if err := h.orderService.ApproveOrder(id, approvedBy, req.Approved, req.Remark); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "审核订单失败")
		return
	}

	Success(c, nil)
}

// CancelOrder 取消订单
// POST /api/orders/:id/cancel
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的订单ID")
		return
	}

	if err := h.orderService.CancelOrder(id); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "取消订单失败")
		return
	}

	Success(c, nil)
}

// ReturnOrder 退货处理
// POST /api/orders/:id/return
func (h *OrderHandler) ReturnOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的订单ID")
		return
	}

	var req ReturnOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if err := h.orderService.ReturnOrder(id, req.ReturnAmount, req.ReturnProfit, req.Remark); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "退货处理失败")
		return
	}

	Success(c, nil)
}

// GetOrderFeed 获取订单动态
// GET /api/orders/feed
func (h *OrderHandler) GetOrderFeed(c *gin.Context) {
	userID := GetUserID(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	orders, err := h.orderService.GetOrderFeed(userID, limit)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询订单动态失败")
		return
	}

	Success(c, orders)
}
