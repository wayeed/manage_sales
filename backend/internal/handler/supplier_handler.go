package handler

import (
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// SupplierHandler 供应商处理器
type SupplierHandler struct {
	supplierService *service.SupplierService
}

// NewSupplierHandler 创建供应商处理器实例
func NewSupplierHandler(supplierService *service.SupplierService) *SupplierHandler {
	return &SupplierHandler{supplierService: supplierService}
}

// List 获取供应商列表
// GET /api/suppliers?page=1&page_size=10&store_id=1&status=1&keyword=
func (h *SupplierHandler) List(c *gin.Context) {
	var req service.ListSupplierRequest
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

	result, err := h.supplierService.List(&req)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询供应商列表失败")
		return
	}

	PageResponse(c, result.List, result.Total, result.Page, result.PageSize)
}

// Create 创建供应商
// POST /api/suppliers
func (h *SupplierHandler) Create(c *gin.Context) {
	var req service.CreateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if err := h.supplierService.Create(&req); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "创建供应商失败")
		return
	}

	Success(c, nil)
}

// Update 更新供应商
// PUT /api/suppliers/:id
func (h *SupplierHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的供应商ID")
		return
	}

	var req service.UpdateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if err := h.supplierService.Update(id, &req); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "更新供应商失败")
		return
	}

	Success(c, nil)
}

// GetDetail 获取供应商详情
// GET /api/suppliers/:id
func (h *SupplierHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的供应商ID")
		return
	}

	supplier, err := h.supplierService.GetDetail(id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "获取供应商详情失败")
		return
	}

	Success(c, supplier)
}

// Delete 删除供应商
// DELETE /api/suppliers/:id
func (h *SupplierHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的供应商ID")
		return
	}

	if err := h.supplierService.Delete(id); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "删除供应商失败")
		return
	}

	Success(c, nil)
}

// AddProduct 添加供应商商品
// POST /api/suppliers/:id/products
func (h *SupplierHandler) AddProduct(c *gin.Context) {
	supplierID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的供应商ID")
		return
	}

	var req service.AddSupplierProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if err := h.supplierService.AddProduct(supplierID, &req); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "添加供应商商品失败")
		return
	}

	Success(c, nil)
}

// RemoveProduct 移除供应商商品
// DELETE /api/suppliers/:id/products/:sku_id
func (h *SupplierHandler) RemoveProduct(c *gin.Context) {
	supplierID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的供应商ID")
		return
	}

	skuID, err := strconv.ParseInt(c.Param("sku_id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的SKU ID")
		return
	}

	if err := h.supplierService.RemoveProduct(supplierID, skuID); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "移除供应商商品失败")
		return
	}

	Success(c, nil)
}

// GetProducts 获取供应商商品列表
// GET /api/suppliers/:id/products
func (h *SupplierHandler) GetProducts(c *gin.Context) {
	supplierID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的供应商ID")
		return
	}

	products, err := h.supplierService.GetProducts(supplierID)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询供应商商品列表失败")
		return
	}

	Success(c, products)
}
