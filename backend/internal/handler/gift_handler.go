package handler

import (
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// GiftHandler 礼品处理器
type GiftHandler struct {
	giftService *service.GiftService
}

// NewGiftHandler 创建礼品处理器实例
func NewGiftHandler(giftService *service.GiftService) *GiftHandler {
	return &GiftHandler{giftService: giftService}
}

// List 获取礼品列表
// @Summary      获取礼品列表
// @Description  分页查询礼品列表，支持按门店、状态、关键词筛选
// @Tags         礼品管理
// @Accept       json
// @Produce      json
// @Param        page       query  int     false  "页码"      default(1)
// @Param        page_size  query  int     false  "每页数量"   default(10)
// @Param        store_id   query  int64   false  "门店ID"
// @Param        status     query  int8    false  "状态"
// @Param        keyword    query  string  false  "搜索关键词"
// @Success      200  {object}  handler.Response{data=handler.PageData}  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /gifts [get]
func (h *GiftHandler) List(c *gin.Context) {
	var req service.ListGiftRequest
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

	result, err := h.giftService.List(&req)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询礼品列表失败")
		return
	}

	PageResponse(c, result.List, result.Total, result.Page, result.PageSize)
}

// Create 创建礼品
// @Summary      创建礼品
// @Description  创建新礼品，需要管理员权限
// @Tags         礼品管理
// @Accept       json
// @Produce      json
// @Param        request  body  service.CreateGiftRequest  true  "创建礼品请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /gifts [post]
func (h *GiftHandler) Create(c *gin.Context) {
	var req service.CreateGiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	createdBy := GetUserID(c)
	if err := h.giftService.Create(&req, createdBy); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "创建礼品失败")
		return
	}

	Success(c, nil)
}

// Update 更新礼品
// @Summary      更新礼品
// @Description  根据礼品ID更新礼品信息，需要管理员权限
// @Tags         礼品管理
// @Accept       json
// @Produce      json
// @Param        id       path  int64                    true  "礼品ID"
// @Param        request  body  service.UpdateGiftRequest  true  "更新礼品请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /gifts/{id} [put]
func (h *GiftHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的礼品ID")
		return
	}

	var req service.UpdateGiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if err := h.giftService.Update(id, &req); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "更新礼品失败")
		return
	}

	Success(c, nil)
}

// GetDetail 获取礼品详情
// @Summary      获取礼品详情
// @Description  根据礼品ID获取礼品详细信息
// @Tags         礼品管理
// @Accept       json
// @Produce      json
// @Param        id  path  int64  true  "礼品ID"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /gifts/{id} [get]
func (h *GiftHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的礼品ID")
		return
	}

	gift, err := h.giftService.GetDetail(id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "获取礼品详情失败")
		return
	}

	Success(c, gift)
}

// Delete 删除礼品
// @Summary      删除礼品
// @Description  根据礼品ID删除礼品，需要管理员权限
// @Tags         礼品管理
// @Accept       json
// @Produce      json
// @Param        id  path  int64  true  "礼品ID"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /gifts/{id} [delete]
func (h *GiftHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的礼品ID")
		return
	}

	if err := h.giftService.Delete(id); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "删除礼品失败")
		return
	}

	Success(c, nil)
}

// AddStock 增加礼品库存
// @Summary      增加礼品库存
// @Description  为指定礼品增加库存，需要管理员权限
// @Tags         礼品管理
// @Accept       json
// @Produce      json
// @Param        id         path  int64                        true  "礼品ID"
// @Param        store_id   query  int64                       false  "门店ID"
// @Param        request    body  service.AddGiftStockRequest   true  "增加库存请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /gifts/{id}/stock [post]
func (h *GiftHandler) AddStock(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的礼品ID")
		return
	}

	var req service.AddGiftStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	createdBy := GetUserID(c)
	storeID, _ := strconv.ParseInt(c.Query("store_id"), 10, 64)

	if err := h.giftService.AddStock(id, &req, storeID, createdBy); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "增加礼品库存失败")
		return
	}

	Success(c, nil)
}
