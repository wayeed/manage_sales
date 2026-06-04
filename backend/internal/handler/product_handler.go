package handler

import (
	"io"
	"net/http"
	"strconv"

	"furniture-commission/internal/pkg/excel"
	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// UpdateProductStatusRequest 更新商品状态请求
type UpdateProductStatusRequest struct {
	Status int8 `json:"status" binding:"required" example:"1"`
}

// ProductHandler 商品处理器
type ProductHandler struct {
	productService *service.ProductService
	skuService     *service.SKUService
}

// NewProductHandler 创建商品处理器实例
func NewProductHandler(productService *service.ProductService, skuService *service.SKUService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		skuService:     skuService,
	}
}

// List 获取商品列表
// @Summary      获取商品列表
// @Description  分页查询商品列表，支持按门店、品类、状态、关键词筛选
// @Tags         商品管理
// @Accept       json
// @Produce      json
// @Param        page         query  int     false  "页码"      default(1)
// @Param        page_size    query  int     false  "每页数量"   default(10)
// @Param        store_id     query  int64   false  "门店ID"
// @Param        category_id  query  int64   false  "品类ID"
// @Param        status       query  int8    false  "状态"
// @Param        keyword      query  string  false  "搜索关键词"
// @Success      200  {object}  handler.Response{data=handler.PageData}  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /products [get]
func (h *ProductHandler) List(c *gin.Context) {
	var req service.ListProductRequest
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
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	result, err := h.productService.List(&req)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询商品列表失败")
		return
	}

	PageResponse(c, result.List, result.Total, result.Page, result.PageSize)
}

// Create 创建商品
// @Summary      创建商品
// @Description  创建新商品，需要管理员权限
// @Tags         商品管理
// @Accept       json
// @Produce      json
// @Param        request  body  service.CreateProductRequest  true  "创建商品请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /products [post]
func (h *ProductHandler) Create(c *gin.Context) {
	var req service.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	createdBy := GetUserID(c)
	if err := h.productService.Create(&req, createdBy); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "创建商品失败")
		return
	}

	Success(c, nil)
}

// Update 更新商品
// @Summary      更新商品
// @Description  根据商品ID更新商品信息，需要管理员权限
// @Tags         商品管理
// @Accept       json
// @Produce      json
// @Param        id       path  int64                      true  "商品ID"
// @Param        request  body  service.UpdateProductRequest  true  "更新商品请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /products/{id} [put]
func (h *ProductHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的商品ID")
		return
	}

	var req service.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if err := h.productService.Update(id, &req); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "更新商品失败")
		return
	}

	Success(c, nil)
}

// GetDetail 获取商品详情
// @Summary      获取商品详情
// @Description  根据商品ID获取商品详细信息
// @Tags         商品管理
// @Accept       json
// @Produce      json
// @Param        id  path  int64  true  "商品ID"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /products/{id} [get]
func (h *ProductHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的商品ID")
		return
	}

	product, err := h.productService.GetDetail(id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "获取商品详情失败")
		return
	}

	Success(c, product)
}

// Delete 删除商品
// @Summary      删除商品
// @Description  根据商品ID删除商品，需要管理员权限
// @Tags         商品管理
// @Accept       json
// @Produce      json
// @Param        id  path  int64  true  "商品ID"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /products/{id} [delete]
func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的商品ID")
		return
	}

	if err := h.productService.Delete(id); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "删除商品失败")
		return
	}

	Success(c, nil)
}

// UpdateStatus 更新商品状态（上下架）
// @Summary      更新商品状态
// @Description  更新商品的上下架状态，需要管理员权限
// @Tags         商品管理
// @Accept       json
// @Produce      json
// @Param        id       path  int64                       true  "商品ID"
// @Param        request  body  UpdateProductStatusRequest   true  "状态更新请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /products/{id}/status [put]
func (h *ProductHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的商品ID")
		return
	}

	var req UpdateProductStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if err := h.productService.UpdateStatus(id, req.Status); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "更新商品状态失败")
		return
	}

	Success(c, nil)
}

// ListSKU 获取商品SKU列表
// GET /api/products/:id/skus
func (h *ProductHandler) ListSKU(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的商品ID")
		return
	}

	skus, err := h.skuService.ListByProduct(productID)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询SKU列表失败")
		return
	}

	Success(c, skus)
}

// CreateSKU 创建SKU
// @Summary      创建SKU
// @Description  为指定商品创建SKU，需要管理员权限
// @Tags         商品管理
// @Accept       json
// @Produce      json
// @Param        id       path  int64                   true  "商品ID"
// @Param        request  body  service.CreateSKURequest  true  "创建SKU请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /products/{id}/skus [post]
func (h *ProductHandler) CreateSKU(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的商品ID")
		return
	}

	var req service.CreateSKURequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}
	req.ProductID = productID

	if err := h.skuService.Create(&req); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "创建SKU失败")
		return
	}

	Success(c, nil)
}

// UpdateSKU 更新SKU
// PUT /api/skus/:id
func (h *ProductHandler) UpdateSKU(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的SKU ID")
		return
	}

	var req service.UpdateSKURequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if err := h.skuService.Update(id, &req); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "更新SKU失败")
		return
	}

	Success(c, nil)
}

// DeleteSKU 删除SKU
// DELETE /api/skus/:id
func (h *ProductHandler) DeleteSKU(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的SKU ID")
		return
	}

	if err := h.skuService.Delete(id); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "删除SKU失败")
		return
	}

	Success(c, nil)
}

// ListAllSKU 获取所有SKU列表（支持搜索）
// @Summary      获取所有SKU列表
// @Description  分页查询所有SKU列表，支持关键词搜索
// @Tags         商品管理
// @Accept       json
// @Produce      json
// @Param        keyword    query  string  false  "搜索关键词"
// @Param        page       query  int     false  "页码"      default(1)
// @Param        page_size  query  int     false  "每页数量"   default(10)
// @Success      200  {object}  handler.Response{data=handler.PageData}  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /skus [get]
func (h *ProductHandler) ListAllSKU(c *gin.Context) {
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	
	skus, total, err := h.skuService.ListAll(keyword, page, pageSize)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询SKU列表失败")
		return
	}
	
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	
	PageResponse(c, skus, total, page, pageSize)
}

// Import 批量导入商品
// @Summary      批量导入商品
// @Description  上传xlsx文件批量导入商品及SKU
// @Tags         商品管理
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "xlsx文件"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /products/import [post]
func (h *ProductHandler) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		Error(c, 400, "请上传文件")
		return
	}

	// 检查文件后缀
	if file.Filename == "" || len(file.Filename) < 5 || file.Filename[len(file.Filename)-5:] != ".xlsx" {
		Error(c, 400, "仅支持xlsx格式文件")
		return
	}

	// 读取文件内容
	src, err := file.Open()
	if err != nil {
		Error(c, 400, "读取文件失败")
		return
	}
	defer src.Close()

	fileData, err := io.ReadAll(src)
	if err != nil {
		Error(c, 400, "读取文件内容失败")
		return
	}

	// 获取用户门店ID
	storeID := GetStoreID(c)
	createdBy := GetUserID(c)

	// 执行导入
	result, err := h.productService.BatchImport(storeID, createdBy, fileData)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "导入失败")
		return
	}

	Success(c, result)
}

// DownloadTemplate 下载导入模板
// @Summary      下载商品导入模板
// @Description  下载商品批量导入的xlsx模板文件
// @Tags         商品管理
// @Produce      octet-stream
// @Success      200  {file}  binary  "xlsx文件"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /products/import-template [get]
func (h *ProductHandler) DownloadTemplate(c *gin.Context) {
	data, err := excel.GenerateTemplate()
	if err != nil {
		Error(c, 500, "生成模板失败")
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=product_import_template.xlsx")
	c.Header("Content-Length", strconv.Itoa(len(data)))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// ListSKUWithStock 获取带库存的SKU列表（用于订单选商品）
// @Summary      获取带库存的SKU列表
// @Description  分页查询带库存信息的SKU列表，用于订单选商品
// @Tags         商品管理
// @Accept       json
// @Produce      json
// @Param        store_id   query  int64   true   "门店ID"
// @Param        keyword    query  string  false  "搜索关键词"
// @Param        page       query  int     false  "页码"      default(1)
// @Param        page_size  query  int     false  "每页数量"   default(10)
// @Success      200  {object}  handler.Response{data=handler.PageData}  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /skus/with-stock [get]
func (h *ProductHandler) ListSKUWithStock(c *gin.Context) {
	keyword := c.Query("keyword")
	storeID, _ := strconv.ParseInt(c.Query("store_id"), 10, 64)
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	if storeID <= 0 {
		Error(c, 400, "缺少门店ID")
		return
	}

	skus, total, err := h.skuService.ListWithStock(storeID, keyword, page, pageSize)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询SKU列表失败")
		return
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	PageResponse(c, skus, total, page, pageSize)
}
