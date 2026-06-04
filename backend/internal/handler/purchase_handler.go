package handler

import (
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// ConfirmReceiptRequest 确认入库请求
type ConfirmReceiptRequest struct {
	WarehouseID int64  `json:"warehouse_id" example:"1"`
	Remark      string `json:"remark" example:"入库备注"`
}

// PurchaseHandler 采购处理器
type PurchaseHandler struct {
	purchaseService *service.PurchaseService
}

// NewPurchaseHandler 创建采购处理器实例
func NewPurchaseHandler(purchaseService *service.PurchaseService) *PurchaseHandler {
	return &PurchaseHandler{purchaseService: purchaseService}
}

// CreateOrder 创建采购订单
// @Summary      创建采购订单
// @Description  创建新的采购订单，需要管理员权限
// @Tags         采购管理
// @Accept       json
// @Produce      json
// @Param        warehouse_id  query  int64                              false  "仓库ID"
// @Param        request      body  service.CreatePurchaseOrderRequest    true   "创建采购订单请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /purchases [post]
func (h *PurchaseHandler) CreateOrder(c *gin.Context) {
	var req service.CreatePurchaseOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	warehouseID, _ := strconv.ParseInt(c.Query("warehouse_id"), 10, 64)
	createdBy := GetUserID(c)

	if err := h.purchaseService.CreateOrder(&req, warehouseID, createdBy); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "创建采购订单失败")
		return
	}

	Success(c, nil)
}

// List 获取采购订单列表
// @Summary      获取采购订单列表
// @Description  分页查询采购订单列表，支持按门店、状态、关键词筛选
// @Tags         采购管理
// @Accept       json
// @Produce      json
// @Param        page         query  int     false  "页码"      default(1)
// @Param        page_size    query  int     false  "每页数量"   default(10)
// @Param        store_id     query  int64   false  "门店ID"
// @Param        status       query  int8    false  "订单状态"
// @Param        keyword      query  string  false  "搜索关键词"
// @Success      200  {object}  handler.Response{data=handler.PageData}  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /purchases [get]
func (h *PurchaseHandler) List(c *gin.Context) {
	var req service.ListPurchaseOrderRequest
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

	result, err := h.purchaseService.List(&req)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询采购订单列表失败")
		return
	}

	PageResponse(c, result.List, result.Total, result.Page, result.PageSize)
}

// GetDetail 获取采购订单详情
// @Summary      获取采购订单详情
// @Description  根据采购订单ID获取订单详细信息
// @Tags         采购管理
// @Accept       json
// @Produce      json
// @Param        id  path  int64  true  "采购订单ID"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /purchases/{id} [get]
func (h *PurchaseHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的采购订单ID")
		return
	}

	order, err := h.purchaseService.GetDetail(id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "获取采购订单详情失败")
		return
	}

	Success(c, order)
}

// ApproveOrder 审核采购订单
// @Summary      审核采购订单
// @Description  审核通过采购订单，需要管理员权限
// @Tags         采购管理
// @Accept       json
// @Produce      json
// @Param        id  path  int64  true  "采购订单ID"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /purchases/{id}/approve [put]
func (h *PurchaseHandler) ApproveOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的采购订单ID")
		return
	}

	auditedBy := GetUserID(c)
	if err := h.purchaseService.ApproveOrder(id, auditedBy); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "审核采购订单失败")
		return
	}

	Success(c, nil)
}

// ConfirmReceipt 确认入库
// @Summary      确认入库
// @Description  确认采购订单入库，需要管理员权限
// @Tags         采购管理
// @Accept       json
// @Produce      json
// @Param        id       path  int64                   true  "采购订单ID"
// @Param        request  body  ConfirmReceiptRequest    true  "确认入库请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /purchases/{id}/receipt [put]
func (h *PurchaseHandler) ConfirmReceipt(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的采购订单ID")
		return
	}

	var req ConfirmReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 兼容 query 参数方式
		req.WarehouseID, _ = strconv.ParseInt(c.Query("warehouse_id"), 10, 64)
	}

	createdBy := GetUserID(c)

	if err := h.purchaseService.ConfirmReceipt(id, req.WarehouseID, req.Remark, createdBy); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "确认入库失败")
		return
	}

	Success(c, nil)
}

// UpdateOrder 更新采购订单
// @Summary      更新采购订单
// @Description  更新采购订单信息及明细，仅待审核和已审核状态允许修改
// @Tags         采购管理
// @Accept       json
// @Produce      json
// @Param        id       path  int64                            true  "采购订单ID"
// @Param        request  body  service.UpdatePurchaseOrderRequest  true  "更新采购订单请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /purchases/{id} [put]
func (h *PurchaseHandler) UpdateOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的采购订单ID")
		return
	}

	var req service.UpdatePurchaseOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if err := h.purchaseService.UpdateOrder(id, &req); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "更新采购订单失败")
		return
	}

	Success(c, nil)
}

// ListMergeableOrders 查询可合并的采购单
// @Summary      查询可合并采购单
// @Description  查询状态为待审核或已审核的采购单，用于一键生成采购单时选择加入已有采购单
// @Tags         采购管理
// @Produce      json
// @Param        store_id  query  int64  false  "门店ID"
// @Success      200  {object}  handler.Response  "成功"
// @Security     BearerAuth
// @Router       /purchases/mergeable [get]
func (h *PurchaseHandler) ListMergeableOrders(c *gin.Context) {
	storeID, _ := strconv.ParseInt(c.Query("store_id"), 10, 64)

	orders, err := h.purchaseService.ListMergeableOrders(storeID)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询可合并采购单失败")
		return
	}

	Success(c, orders)
}

// AppendItems 向已有采购单追加商品
// @Summary      追加采购商品
// @Description  向已有的未入库采购单追加商品明细
// @Tags         采购管理
// @Accept       json
// @Produce      json
// @Param        id       path  int64                    true  "采购订单ID"
// @Param        request  body  service.AppendItemsRequest  true  "追加商品请求"
// @Success      200  {object}  handler.Response  "成功"
// @Security     BearerAuth
// @Router       /purchases/{id}/items [post]
func (h *PurchaseHandler) AppendItems(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的采购订单ID")
		return
	}

	var req service.AppendItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if err := h.purchaseService.AppendItems(id, &req); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "追加采购商品失败")
		return
	}

	Success(c, nil)
}

// CancelOrder 取消采购订单
// @Summary      取消采购订单
// @Description  取消采购订单，需要管理员权限
// @Tags         采购管理
// @Accept       json
// @Produce      json
// @Param        id  path  int64  true  "采购订单ID"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /purchases/{id}/cancel [put]
func (h *PurchaseHandler) CancelOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的采购订单ID")
		return
	}

	if err := h.purchaseService.CancelOrder(id); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "取消采购订单失败")
		return
	}

	Success(c, nil)
}
