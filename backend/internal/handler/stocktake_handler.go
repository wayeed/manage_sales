package handler

import (
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

type StocktakeHandler struct {
	stocktakeService *service.StocktakeService
}

func NewStocktakeHandler(stocktakeService *service.StocktakeService) *StocktakeHandler {
	return &StocktakeHandler{stocktakeService: stocktakeService}
}

// Create 创建盘点单
// POST /api/stocktakes
func (h *StocktakeHandler) Create(c *gin.Context) {
	var req service.CreateStocktakeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误: "+err.Error())
		return
	}

	userID := GetUserID(c)
	storeID := GetStoreID(c)

	stocktake, err := h.stocktakeService.Create(&req, storeID, userID)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "创建盘点单失败")
		return
	}

	Success(c, stocktake)
}

// List 盘点单列表
// GET /api/stocktakes?warehouse_id=1&status=0&page=1&page_size=10
func (h *StocktakeHandler) List(c *gin.Context) {
	var req service.ListStocktakeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	storeID := GetStoreID(c)

	result, err := h.stocktakeService.List(&req, storeID)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询盘点单列表失败")
		return
	}

	PageResponse(c, result.List, result.Total, result.Page, result.PageSize)
}

// GetDetail 盘点单详情
// GET /api/stocktakes/:id
func (h *StocktakeHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的盘点单ID")
		return
	}

	stocktake, err := h.stocktakeService.GetDetail(id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询盘点单详情失败")
		return
	}

	Success(c, stocktake)
}

// Update 更新盘点单
// PUT /api/stocktakes/:id
func (h *StocktakeHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的盘点单ID")
		return
	}

	var req service.UpdateStocktakeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误: "+err.Error())
		return
	}

	userID := GetUserID(c)

	stocktake, err := h.stocktakeService.Update(id, &req, userID)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "更新盘点单失败")
		return
	}

	Success(c, stocktake)
}

// Submit 提交盘点单
// POST /api/stocktakes/:id/submit
func (h *StocktakeHandler) Submit(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的盘点单ID")
		return
	}

	userID := GetUserID(c)

	if err := h.stocktakeService.Submit(id, userID); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "提交盘点单失败")
		return
	}

	Success(c, nil)
}

// Approve 审核盘点单
// POST /api/stocktakes/:id/approve
func (h *StocktakeHandler) Approve(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的盘点单ID")
		return
	}

	var req struct {
		Approved bool `json:"approved"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	userID := GetUserID(c)

	if err := h.stocktakeService.Approve(id, userID, req.Approved); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "审核盘点单失败")
		return
	}

	Success(c, nil)
}

// Delete 删除盘点单
// DELETE /api/stocktakes/:id
func (h *StocktakeHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的盘点单ID")
		return
	}

	if err := h.stocktakeService.Delete(id); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "删除盘点单失败")
		return
	}

	Success(c, nil)
}
