package handler

import (
	"net/http"
	"strconv"

	"furniture-commission/internal/dto"
	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// DeliveryHandler 送货出库处理器
type DeliveryHandler struct {
	deliveryService service.DeliveryService
}

// NewDeliveryHandler 创建送货出库处理器
func NewDeliveryHandler(deliveryService service.DeliveryService) *DeliveryHandler {
	return &DeliveryHandler{
		deliveryService: deliveryService,
	}
}

// CreateDelivery 创建送货出库
func (h *DeliveryHandler) CreateDelivery(c *gin.Context) {
	var req dto.CreateDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 获取当前用户ID
	operatorID, exists := c.Get("user_id")
	if !exists {
		Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	record, err := h.deliveryService.CreateDelivery(&req, uint64(operatorID.(int64)))
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessWithMessage(c, "送货出库成功", record)
}

// GetDeliveryList 获取送货出库列表
func (h *DeliveryHandler) GetDeliveryList(c *gin.Context) {
	var req dto.DeliveryListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	result, err := h.deliveryService.GetDeliveryList(&req)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, result)
}

// GetDeliveryDetail 获取送货出库详情
func (h *DeliveryHandler) GetDeliveryDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	detail, err := h.deliveryService.GetDeliveryDetail(id)
	if err != nil {
		Error(c, http.StatusNotFound, err.Error())
		return
	}

	Success(c, detail)
}

// CancelDelivery 作废送货记录
func (h *DeliveryHandler) CancelDelivery(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var req dto.CancelDeliveryRequest
	c.ShouldBindJSON(&req) // 作废原因可选

	// 获取当前用户ID
	operatorID, exists := c.Get("user_id")
	if !exists {
		Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	if err := h.deliveryService.CancelDelivery(id, uint64(operatorID.(int64)), req.Remark); err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessWithMessage(c, "作废成功", nil)
}

// ConfirmDelivery 确认送达
func (h *DeliveryHandler) ConfirmDelivery(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}

	if err := h.deliveryService.ConfirmDelivery(int64(orderID)); err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessWithMessage(c, "已确认送达", nil)
}

// GetPendingDeliveryOrders 获取待送货订单列表
func (h *DeliveryHandler) GetPendingDeliveryOrders(c *gin.Context) {
	var req dto.PendingDeliveryOrderQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	result, err := h.deliveryService.GetPendingDeliveryOrders(&req)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, result)
}

// RegisterDeliveryRoutes 注册送货管理路由
func RegisterDeliveryRoutes(router *gin.RouterGroup, handler *DeliveryHandler, permissionMiddleware func(...string) gin.HandlerFunc) {
	deliveryGroup := router.Group("/deliveries")
	{
		// 创建送货出库 - 需要送货出库权限
		deliveryGroup.POST("", handler.CreateDelivery)

		// 获取送货列表 - 需要查看送货权限
		deliveryGroup.GET("", permissionMiddleware("delivery:view"), handler.GetDeliveryList)

		// 获取待送货订单列表 - 需要查看送货权限
		deliveryGroup.GET("/orders/pending", permissionMiddleware("delivery:view"), handler.GetPendingDeliveryOrders)

		// 获取送货详情 - 需要查看送货权限
		deliveryGroup.GET("/:id", permissionMiddleware("delivery:view"), handler.GetDeliveryDetail)

		// 作废送货记录 - 需要作废送货权限
		deliveryGroup.PUT("/:id/cancel", permissionMiddleware("delivery:cancel"), handler.CancelDelivery)

		// 确认送达 - 需要送货出库权限
		deliveryGroup.PUT("/:id/confirm", permissionMiddleware("delivery:create"), handler.ConfirmDelivery)

                // 获取订单库存状态 - 需要查看送货权限
                deliveryGroup.GET("/stock-status", handler.GetOrderStockStatus)

                // 打印送货单时更新配送状态
                deliveryGroup.POST("/:order_id/print", handler.PrintDelivery)
        }
}

// GetOrderStockStatus 获取订单库存状态
func (h *DeliveryHandler) GetOrderStockStatus(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Query("order_id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的订单ID")
		return
	}
	warehouseID, err := strconv.ParseInt(c.Query("warehouse_id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的仓库ID")
		return
	}

	stockStatus, err := h.deliveryService.GetOrderStockStatus(orderID, warehouseID)
	if err != nil {
		Error(c, 500, err.Error())
		return
	}

	Success(c, stockStatus)
}

// PrintDelivery 打印送货单时更新配送状态
func (h *DeliveryHandler) PrintDelivery(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		Error(c, 400, "无效的订单ID")
		return
	}

	if err := h.deliveryService.PrintDelivery(orderID); err != nil {
		Error(c, 500, err.Error())
		return
	}

	Success(c, nil)
}
