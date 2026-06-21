package handler

import (
	"furniture-commission/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ReceiptHandler 回货单控制器
type ReceiptHandler struct {
	receiptService *service.ReceiptService
}

// NewReceiptHandler 创建回货单控制器实例
func NewReceiptHandler(receiptService *service.ReceiptService) *ReceiptHandler {
	return &ReceiptHandler{
		receiptService: receiptService,
	}
}

// CreateReceipt 创建回货单
func (h *ReceiptHandler) CreateReceipt(c *gin.Context) {
	var req service.CreateReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 获取当前用户ID
	userID := GetUserID(c)
	if userID <= 0 {
		Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	// 获取当前用户所属门店ID
	storeID := GetStoreID(c)

	// 如果请求中没有指定门店ID，使用用户所属门店
	if req.StoreID <= 0 {
		req.StoreID = storeID
	}

	err := h.receiptService.CreateReceiptOrder(&req, userID)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "创建回货单失败")
		return
	}

	Success(c, nil)
}

// ApproveReceipt 审核回货单
func (h *ReceiptHandler) ApproveReceipt(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的回货单ID")
		return
	}

	userID := GetUserID(c)
	if userID <= 0 {
		Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	err = h.receiptService.ApproveReceiptOrder(id, userID)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "审核回货单失败")
		return
	}

	Success(c, nil)
}

// ReceiveReceipt 入库操作
func (h *ReceiptHandler) ReceiveReceipt(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的回货单ID")
		return
	}

	var req service.ReceiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	userID := GetUserID(c)
	if userID <= 0 {
		Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	err = h.receiptService.ReceiveReceiptOrder(id, &req, userID)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "入库失败")
		return
	}

	Success(c, nil)
}

// GetReceiptDetail 获取回货单详情
func (h *ReceiptHandler) GetReceiptDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的回货单ID")
		return
	}

	detail, err := h.receiptService.GetReceiptOrderDetail(id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "获取回货单详情失败")
		return
	}

	Success(c, detail)
}

// ListReceipts 查询回货单列表
func (h *ReceiptHandler) ListReceipts(c *gin.Context) {
	var req service.ListReceiptRequest

	storeID, _ := strconv.ParseInt(c.Query("store_id"), 10, 64)
	supplierID, _ := strconv.ParseInt(c.Query("supplier_id"), 10, 64)
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	req.StoreID = storeID
	req.SupplierID = supplierID
	req.Page = page
	req.PageSize = pageSize
	req.Keyword = c.Query("keyword")

	if statusStr := c.Query("status"); statusStr != "" {
		status, err := strconv.ParseInt(statusStr, 10, 8)
		if err == nil {
			status8 := int8(status)
			req.Status = &status8
		}
	}

	result, err := h.receiptService.ListReceiptOrders(&req)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询回货单列表失败")
		return
	}

	Success(c, result)
}

// CancelReceipt 取消回货单
func (h *ReceiptHandler) CancelReceipt(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的回货单ID")
		return
	}

	err = h.receiptService.CancelReceiptOrder(id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "取消回货单失败")
		return
	}

	Success(c, nil)
}
