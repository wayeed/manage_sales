package handler

import (
	"furniture-commission/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AppVersionHandler APP版本管理处理器
type AppVersionHandler struct {
	appVersionService *service.AppVersionService
}

// NewAppVersionHandler 创建APP版本管理处理器实例
func NewAppVersionHandler(appVersionService *service.AppVersionService) *AppVersionHandler {
	return &AppVersionHandler{
		appVersionService: appVersionService,
	}
}

// Create 创建APP版本
// POST /api/app-versions
func (h *AppVersionHandler) Create(c *gin.Context) {
	var req service.CreateAppVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	userID := c.GetInt64("user_id")
	version, err := h.appVersionService.Create(&req, userID)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, http.StatusInternalServerError, "创建版本失败")
		return
	}

	Success(c, version)
}

// Update 更新APP版本
// PUT /api/app-versions/:id
func (h *AppVersionHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的版本ID")
		return
	}

	var req service.UpdateAppVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	version, err := h.appVersionService.Update(id, &req)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, http.StatusInternalServerError, "更新版本失败")
		return
	}

	Success(c, version)
}

// Delete 删除APP版本
// DELETE /api/app-versions/:id
func (h *AppVersionHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的版本ID")
		return
	}

	if err := h.appVersionService.Delete(id); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, http.StatusInternalServerError, "删除版本失败")
		return
	}

	Success(c, nil)
}

// GetByID 根据ID获取版本
// GET /api/app-versions/:id
func (h *AppVersionHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的版本ID")
		return
	}

	version, err := h.appVersionService.GetByID(id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, http.StatusInternalServerError, "查询版本失败")
		return
	}

	Success(c, version)
}

// List 获取版本列表
// GET /api/app-versions
func (h *AppVersionHandler) List(c *gin.Context) {
	var req service.ListAppVersionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.appVersionService.List(&req)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, http.StatusInternalServerError, "查询版本列表失败")
		return
	}

	Success(c, result)
}

// GetLatest 获取最新版本（APP端调用）
// GET /api/app-versions/latest
func (h *AppVersionHandler) GetLatest(c *gin.Context) {
	platform := c.Query("platform")
	if platform == "" {
		platform = "android"
	}

	version, err := h.appVersionService.GetLatest(platform)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, http.StatusInternalServerError, "查询版本失败")
		return
	}

	Success(c, version)
}
