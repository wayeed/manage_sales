package handler

import (
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// StoreHandler 门店处理器
type StoreHandler struct {
	storeService *service.StoreService
}

// NewStoreHandler 创建门店处理器实例
func NewStoreHandler(storeService *service.StoreService) *StoreHandler {
	return &StoreHandler{storeService: storeService}
}

// List 获取门店列表
// @Summary      获取门店列表
// @Description  获取所有门店信息列表
// @Tags         门店管理
// @Accept       json
// @Produce      json
// @Success      200  {object}  handler.Response{data=[]interface{}}  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /stores [get]
func (h *StoreHandler) List(c *gin.Context) {
	stores, err := h.storeService.List()
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询门店列表失败")
		return
	}

	Success(c, stores)
}

// Create 创建门店
// @Summary      创建门店
// @Description  创建新门店（管理员）
// @Tags         门店管理
// @Accept       json
// @Produce      json
// @Param        request  body  service.CreateStoreRequest  true  "门店信息"
// @Success      200  {object}  handler.Response{data=interface{}}  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /stores [post]
func (h *StoreHandler) Create(c *gin.Context) {
	var req service.CreateStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误: "+err.Error())
		return
	}

	store, err := h.storeService.Create(&req)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "创建门店失败")
		return
	}

	Success(c, store)
}

// Update 更新门店
// @Summary      更新门店
// @Description  根据门店ID更新门店信息（管理员）
// @Tags         门店管理
// @Accept       json
// @Produce      json
// @Param        id       path  int                        true  "门店ID"
// @Param        request  body  service.UpdateStoreRequest  true  "更新门店信息"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /stores/{id} [put]
func (h *StoreHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的门店ID")
		return
	}

	var req service.UpdateStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误: "+err.Error())
		return
	}

	if err := h.storeService.Update(id, &req); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "更新门店失败")
		return
	}

	Success(c, nil)
}

// Delete 删除门店
// @Summary      删除门店
// @Description  根据门店ID删除门店（管理员）
// @Tags         门店管理
// @Accept       json
// @Produce      json
// @Param        id   path  int64  true  "门店ID"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /stores/{id} [delete]
func (h *StoreHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的门店ID")
		return
	}

	if err := h.storeService.Delete(id); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "删除门店失败")
		return
	}

	Success(c, nil)
}
