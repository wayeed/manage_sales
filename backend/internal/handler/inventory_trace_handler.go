package handler

import (
	"net/http"
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// InventoryTraceHandler 库存穿透查询处理器
type InventoryTraceHandler struct {
	traceService *service.InventoryTraceService
}

// NewInventoryTraceHandler 创建库存穿透查询处理器
func NewInventoryTraceHandler(traceService *service.InventoryTraceService) *InventoryTraceHandler {
	return &InventoryTraceHandler{traceService: traceService}
}

// ForwardTrace 正向穿透：订单 → 源头
// GET /api/inventory/trace/forward?order_no=xxx
func (h *InventoryTraceHandler) ForwardTrace(c *gin.Context) {
	orderNo := c.Query("order_no")
	if orderNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请输入订单号"})
		return
	}

	result, err := h.traceService.ForwardTraceByOrderNo(orderNo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result})
}

// BackwardTrace 反向穿透：批次 → 去向
// GET /api/inventory/trace/backward?batch_no=xxx
func (h *InventoryTraceHandler) BackwardTrace(c *gin.Context) {
	batchNo := c.Query("batch_no")
	if batchNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请输入批次号"})
		return
	}

	result, err := h.traceService.BackwardTrace(batchNo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "未找到该批次: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result})
}

// SKUBatchTrace SKU库存全景
// GET /api/inventory/trace/sku?sku_id=xxx
func (h *InventoryTraceHandler) SKUBatchTrace(c *gin.Context) {
	skuIDStr := c.Query("sku_id")
	if skuIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请输入SKU ID"})
		return
	}
	skuID, err := strconv.ParseInt(skuIDStr, 10, 64)
	if err != nil || skuID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "SKU ID格式错误"})
		return
	}

	result, err := h.traceService.SKUBatchTrace(skuID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result})
}
