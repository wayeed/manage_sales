package handler

import (
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// CategoryHandler 品类处理器
type CategoryHandler struct {
	categoryService *service.CategoryService
}

// NewCategoryHandler 创建品类处理器实例
func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

// List 获取品类列表
// @Summary      获取品类列表
// @Description  获取品类列表，可按门店ID筛选
// @Tags         品类管理
// @Accept       json
// @Produce      json
// @Param        store_id  query  int64  false  "门店ID"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	storeID, _ := strconv.ParseInt(c.Query("store_id"), 10, 64)

	categories, err := h.categoryService.List(storeID)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询品类列表失败")
		return
	}

	Success(c, categories)
}

// Create 创建品类
// @Summary      创建品类
// @Description  创建新品类，需要管理员权限
// @Tags         品类管理
// @Accept       json
// @Produce      json
// @Param        request  body  service.CreateCategoryRequest  true  "创建品类请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var req service.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if err := h.categoryService.Create(&req); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "创建品类失败")
		return
	}

	Success(c, nil)
}

// Update 更新品类
// @Summary      更新品类
// @Description  根据品类ID更新品类信息，需要管理员权限
// @Tags         品类管理
// @Accept       json
// @Produce      json
// @Param        id       path  int64                      true  "品类ID"
// @Param        request  body  service.UpdateCategoryRequest  true  "更新品类请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的品类ID")
		return
	}

	var req service.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if err := h.categoryService.Update(id, &req); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "更新品类失败")
		return
	}

	Success(c, nil)
}

// Delete 删除品类
// @Summary      删除品类
// @Description  根据品类ID删除品类，需要管理员权限
// @Tags         品类管理
// @Accept       json
// @Produce      json
// @Param        id  path  int64  true  "品类ID"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的品类ID")
		return
	}

	if err := h.categoryService.Delete(id); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "删除品类失败")
		return
	}

	Success(c, nil)
}
