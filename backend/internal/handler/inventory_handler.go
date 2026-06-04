package handler

import (
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// InventoryHandler 库存处理器
type InventoryHandler struct {
	inventoryService *service.InventoryService
}

// NewInventoryHandler 创建库存处理器实例
func NewInventoryHandler(inventoryService *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{inventoryService: inventoryService}
}

// GetStock 获取库存
// @Summary      查询库存
// @Description  根据仓库ID和SKU ID查询库存信息
// @Tags         库存管理
// @Accept       json
// @Produce      json
// @Param        warehouse_id  query  int64  true  "仓库ID"
// @Param        sku_id        query  int64  true  "SKU ID"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /inventory/stock [get]
func (h *InventoryHandler) GetStock(c *gin.Context) {
	warehouseID, _ := strconv.ParseInt(c.Query("warehouse_id"), 10, 64)
	skuID, _ := strconv.ParseInt(c.Query("sku_id"), 10, 64)

	if warehouseID <= 0 || skuID <= 0 {
		Error(c, 400, "仓库ID和SKU ID不能为空")
		return
	}

	stock, err := h.inventoryService.GetStock(warehouseID, skuID)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询库存失败")
		return
	}

	Success(c, stock)
}

// GetStockList 获取仓库库存列表
// @Summary      获取库存列表
// @Description  分页查询仓库库存列表，支持按仓库、SKU ID和关键词筛选
// @Tags         库存管理
// @Accept       json
// @Produce      json
// @Param        warehouse_id  query  int64   false  "仓库ID，不传或传0则查询所有仓库"
// @Param        sku_id        query  int64   false  "SKU ID，精确查询指定SKU库存"
// @Param        keyword       query  string  false  "搜索关键词"
// @Param        page          query  int     false  "页码"      default(1)
// @Param        page_size     query  int     false  "每页数量"   default(10)
// @Success      200  {object}  handler.Response{data=handler.PageData}  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /inventory/stocks [get]
func (h *InventoryHandler) GetStockList(c *gin.Context) {
	warehouseID, _ := strconv.ParseInt(c.DefaultQuery("warehouse_id", "0"), 10, 64)
	skuID, _ := strconv.ParseInt(c.Query("sku_id"), 10, 64)
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	stocks, total, err := h.inventoryService.GetStockList(warehouseID, skuID, keyword, page, pageSize)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询库存列表失败")
		return
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	PageResponse(c, stocks, total, page, pageSize)
}

// GetTransactions 获取库存流水列表
// @Summary      获取库存流水
// @Description  分页查询库存流水列表，支持按仓库、类型、日期范围筛选
// @Tags         库存管理
// @Accept       json
// @Produce      json
// @Param        warehouse_id  query  int     false  "仓库ID，不传或传0则查询所有仓库"
// @Param        type          query  int     false  "流水类型"
// @Param        start_date    query  string  false  "开始日期"
// @Param        end_date      query  string  false  "结束日期"
// @Param        page          query  int     false  "页码"      default(1)
// @Param        page_size     query  int     false  "每页数量"   default(10)
// @Success      200  {object}  handler.Response{data=handler.PageData}  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /inventory/transactions [get]
func (h *InventoryHandler) GetTransactions(c *gin.Context) {
	warehouseID, _ := strconv.ParseInt(c.DefaultQuery("warehouse_id", "0"), 10, 64)
	transactionType, _ := strconv.ParseInt(c.DefaultQuery("type", "0"), 10, 8)
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	transactions, total, err := h.inventoryService.GetTransactionList(warehouseID, int8(transactionType), startDate, endDate, page, pageSize)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询库存流水失败")
		return
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	PageResponse(c, transactions, total, page, pageSize)
}

// GetStockDetail 根据ID获取库存详情
// GET /api/inventory/:id
func (h *InventoryHandler) GetStockDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的库存ID")
		return
	}

	stock, err := h.inventoryService.GetStockByID(id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询库存详情失败")
		return
	}

	Success(c, stock)
}
