package handler

import (
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// StockAlertHandler 库存预警处理器
type StockAlertHandler struct {
	alertService *service.StockAlertService
}

// NewStockAlertHandler 创建库存预警处理器实例
func NewStockAlertHandler(alertService *service.StockAlertService) *StockAlertHandler {
	return &StockAlertHandler{alertService: alertService}
}

// CheckAlerts 检查库存预警
// POST /api/stock-alerts/check?store_id=1
func (h *StockAlertHandler) CheckAlerts(c *gin.Context) {
	storeID, _ := strconv.ParseInt(c.Query("store_id"), 10, 64)

	count, err := h.alertService.CheckAlerts(storeID)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "检查库存预警失败")
		return
	}

	Success(c, gin.H{"alert_count": count})
}

// List 获取预警列表
// GET /api/stock-alerts?page=1&page_size=10&store_id=1&alert_type=1&alert_status=0
func (h *StockAlertHandler) List(c *gin.Context) {
	var req service.ListAlertRequest
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

	result, err := h.alertService.List(&req)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询预警列表失败")
		return
	}

	PageResponse(c, result.List, result.Total, result.Page, result.PageSize)
}

// Handle 处理预警
// PUT /api/stock-alerts/:id
func (h *StockAlertHandler) Handle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的预警ID")
		return
	}

	var req service.HandleAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	handledBy := GetUserID(c)
	if err := h.alertService.Handle(id, &req, handledBy); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "处理预警失败")
		return
	}

	Success(c, nil)
}
