package handler

import (
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// WarehouseHandler 仓库处理器
type WarehouseHandler struct {
	warehouseRepo *service.WarehouseService
}

// NewWarehouseHandler 创建仓库处理器实例
func NewWarehouseHandler(warehouseService *service.WarehouseService) *WarehouseHandler {
	return &WarehouseHandler{warehouseRepo: warehouseService}
}

// List 获取仓库列表
// @Summary      获取仓库列表
// @Description  获取仓库列表，可按门店ID筛选
// @Tags         仓库管理
// @Accept       json
// @Produce      json
// @Param        store_id  query  int64  false  "门店ID"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /warehouses [get]
func (h *WarehouseHandler) List(c *gin.Context) {
	storeID, _ := strconv.ParseInt(c.Query("store_id"), 10, 64)

	warehouses, err := h.warehouseRepo.List(storeID)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询仓库列表失败")
		return
	}

	Success(c, warehouses)
}

// Create 创建仓库
// @Summary      创建仓库
// @Description  创建新仓库
// @Tags         仓库管理
// @Accept       json
// @Produce      json
// @Param        request  body  service.CreateWarehouseRequest  true  "创建仓库请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /warehouses [post]
func (h *WarehouseHandler) Create(c *gin.Context) {
	var req service.CreateWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	warehouse, err := h.warehouseRepo.Create(&req)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "创建仓库失败")
		return
	}

	Success(c, warehouse)
}

// Update 更新仓库
// @Summary      更新仓库
// @Description  根据仓库ID更新仓库信息
// @Tags         仓库管理
// @Accept       json
// @Produce      json
// @Param        id       path  int64                       true  "仓库ID"
// @Param        request  body  service.UpdateWarehouseRequest  true  "更新仓库请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /warehouses/{id} [put]
func (h *WarehouseHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		Error(c, 400, "无效的仓库ID")
		return
	}

	var req service.UpdateWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	warehouse, err := h.warehouseRepo.Update(id, &req)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "更新仓库失败")
		return
	}

	Success(c, warehouse)
}

// Delete 删除仓库
// @Summary      删除仓库
// @Description  根据仓库ID删除仓库
// @Tags         仓库管理
// @Accept       json
// @Produce      json
// @Param        id  path  int64  true  "仓库ID"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /warehouses/{id} [delete]
func (h *WarehouseHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		Error(c, 400, "无效的仓库ID")
		return
	}

	if err := h.warehouseRepo.Delete(id); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "删除仓库失败")
		return
	}

	Success(c, nil)
}

// GetByID 获取仓库详情
// @Summary      获取仓库详情
// @Description  根据仓库ID获取仓库详细信息
// @Tags         仓库管理
// @Accept       json
// @Produce      json
// @Param        id  path  int64  true  "仓库ID"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /warehouses/{id} [get]
func (h *WarehouseHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		Error(c, 400, "无效的仓库ID")
		return
	}

	warehouse, err := h.warehouseRepo.GetByID(id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询仓库失败")
		return
	}

	Success(c, warehouse)
}
