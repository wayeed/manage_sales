package handler

import (
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// CommissionHandler 提成管理处理器
type CommissionHandler struct {
	commissionService *service.CommissionService
}

// NewCommissionHandler 创建提成管理处理器实例
func NewCommissionHandler(commissionService *service.CommissionService) *CommissionHandler {
	return &CommissionHandler{commissionService: commissionService}
}

// List 获取提成列表
// GET /api/commissions?page=1&page_size=10&store_id=1&employee_id=1&commission_type=1&status=1&period_value=2024-01
func (h *CommissionHandler) List(c *gin.Context) {
	var req service.ListCommissionRequest
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

	result, err := h.commissionService.List(&req)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询提成列表失败")
		return
	}

	PageResponse(c, result.List, result.Total, result.Page, result.PageSize)
}

// GetByOrderID 获取订单提成明细
// GET /api/commissions/order/:order_id
func (h *CommissionHandler) GetByOrderID(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("order_id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的订单ID")
		return
	}

	commissions, err := h.commissionService.GetByOrderID(orderID)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询订单提成明细失败")
		return
	}

	Success(c, commissions)
}

// GetSummary 获取提成汇总
// GET /api/commissions/summary?employee_id=1&start_date=2024-01&end_date=2024-12
// 或 GET /api/commissions/summary?month=2024-01（查询所有员工）
func (h *CommissionHandler) GetSummary(c *gin.Context) {
	// 支持按月份查询所有员工
	month := c.Query("month")
	if month != "" {
		list, err := h.commissionService.GetMonthlySummary(month)
		if err != nil {
			if appErr, ok := err.(*service.AppError); ok {
				Error(c, appErr.Code, appErr.Message)
				return
			}
			Error(c, 500, "查询提成汇总失败")
			return
		}
		Success(c, list)
		return
	}

	// 按员工ID查询
	employeeID, err := strconv.ParseInt(c.Query("employee_id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的员工ID")
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	summary, err := h.commissionService.GetSummary(employeeID, startDate, endDate)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询提成汇总失败")
		return
	}

	Success(c, summary)
}

// ManualAdjust 手工调整提成
// POST /api/commissions/adjust
func (h *CommissionHandler) ManualAdjust(c *gin.Context) {
	var req service.ManualAdjustRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误: "+err.Error())
		return
	}

	amount := decimal.NewFromFloat(req.Amount)
	if err := h.commissionService.ManualAdjust(req.CommissionID, amount, req.Remark); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "调整提成失败")
		return
	}

	Success(c, nil)
}

// EstimateCommission 预估提成
// POST /api/commissions/estimate
func (h *CommissionHandler) EstimateCommission(c *gin.Context) {
	var req service.EstimateCommissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.commissionService.EstimateCommission(&req)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "预估提成失败")
		return
	}

	Success(c, result)
}

// CalculateOrderCommission 计算订单提成
// POST /api/commissions/calculate/:order_id
func (h *CommissionHandler) CalculateOrderCommission(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("order_id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的订单ID")
		return
	}

	if err := h.commissionService.CalculateOrderCommission(orderID); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "计算订单提成失败")
		return
	}

	Success(c, nil)
}
