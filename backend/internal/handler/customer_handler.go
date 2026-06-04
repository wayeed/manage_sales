package handler

import (
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// CustomerHandler 客户处理器
type CustomerHandler struct {
	customerService *service.CustomerService
}

// NewCustomerHandler 创建客户处理器实例
func NewCustomerHandler(customerService *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{customerService: customerService}
}

// Create 创建客户
// POST /api/customers
func (h *CustomerHandler) Create(c *gin.Context) {
	var req service.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误: "+err.Error())
		return
	}

	userID := GetUserID(c)
	storeID, _ := c.Get("storeID")

	// 如果前端没传 store_id，使用当前用户的门店
	if req.StoreID == 0 {
		if sid, ok := storeID.(int64); ok && sid > 0 {
			req.StoreID = sid
		}
	}

	customer, err := h.customerService.Create(&req, userID)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "创建客户失败")
		return
	}

	Success(c, customer)
}

// List 获取客户列表
// GET /api/customers?store_id=1&keyword=&level=&page=1&page_size=10
func (h *CustomerHandler) List(c *gin.Context) {
	var req service.ListCustomerRequest
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

	result, err := h.customerService.List(&req)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询客户列表失败: "+err.Error())
		return
	}

	PageResponse(c, result.List, result.Total, result.Page, result.PageSize)
}

// GetDetail 获取客户详情
// GET /api/customers/:id
func (h *CustomerHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的客户ID")
		return
	}

	customer, err := h.customerService.GetDetail(id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "获取客户详情失败")
		return
	}

	Success(c, customer)
}

// Update 更新客户
// PUT /api/customers/:id
func (h *CustomerHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的客户ID")
		return
	}

	var req service.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if err := h.customerService.Update(id, &req); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "更新客户失败")
		return
	}

	Success(c, nil)
}

// Delete 删除客户
// DELETE /api/customers/:id
func (h *CustomerHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的客户ID")
		return
	}

	if err := h.customerService.Delete(id); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "删除客户失败")
		return
	}

	Success(c, nil)
}

// AddFollowUp 添加跟进记录
// POST /api/customers/:id/follow-ups
func (h *CustomerHandler) AddFollowUp(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的客户ID")
		return
	}

	var req service.AddFollowUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}
	req.CustomerID = id

	followerID := GetUserID(c)

	if err := h.customerService.AddFollowUp(&req, followerID); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "添加跟进记录失败: "+err.Error())
		return
	}

	Success(c, nil)
}

// GetFollowUps 获取跟进记录
// GET /api/customers/:id/follow-ups
func (h *CustomerHandler) GetFollowUps(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的客户ID")
		return
	}

	followUps, err := h.customerService.GetFollowUps(id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询跟进记录失败")
		return
	}

	Success(c, followUps)
}

// GetCustomersWithDraftStatus 获取客户列表（含草稿状态）
// GET /api/customers/with-draft-status?keyword=&page=1&page_size=20
func (h *CustomerHandler) GetCustomersWithDraftStatus(c *gin.Context) {
	userID := GetUserID(c)
	storeID, _ := c.Get("storeID")
	var sid int64
	if s, ok := storeID.(int64); ok {
		sid = s
	}

	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.customerService.GetCustomersWithDraftStatus(userID, sid, keyword, page, pageSize)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询客户列表失败")
		return
	}

	PageResponse(c, result.List, result.Total, result.Page, result.PageSize)
}
