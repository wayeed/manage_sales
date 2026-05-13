package handler

import (
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// UpdateUserStatusRequest 更新用户状态请求
type UpdateUserStatusRequest struct {
	Status int `json:"status" example:"1"`
}

// AssignRoleRequest 分配角色请求
type AssignRoleRequest struct {
	RoleIDs []int64 `json:"role_ids" binding:"required" example:"1,2,3"`
}

// UserHandler 用户处理器
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户处理器实例
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// List 获取用户列表
// @Summary      获取用户列表
// @Description  分页查询用户列表，支持按关键词、角色、状态筛选
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        page       query  int     false  "页码"          default(1)
// @Param        page_size  query  int     false  "每页数量"       default(10)
// @Param        keyword    query  string  false  "搜索关键词"
// @Param        role       query  int     false  "角色筛选"
// @Param        status     query  int     false  "状态筛选"
// @Param        store_id   query  int     false  "门店ID"
// @Success      200  {object}  handler.Response{data=handler.PageData}  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /users [get]
func (h *UserHandler) List(c *gin.Context) {
	var req service.ListUserRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	// 设置默认分页
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	result, err := h.userService.List(&req)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询用户列表失败")
		return
	}

	PageResponse(c, result.List, result.Total, result.Page, result.PageSize)
}

// Create 创建用户
// @Summary      创建用户
// @Description  创建新用户，需要管理员权限
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        request  body  service.CreateUserRequest  true  "创建用户请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req service.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	createdBy := GetUserID(c)
	if err := h.userService.Create(&req, createdBy); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "创建用户失败")
		return
	}

	Success(c, nil)
}

// Update 更新用户
// @Summary      更新用户
// @Description  根据用户ID更新用户信息
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        id       path  int64                   true  "用户ID"
// @Param        request  body  service.UpdateUserRequest  true  "更新用户请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的用户ID")
		return
	}

	var req service.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if err := h.userService.Update(id, &req); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "更新用户失败")
		return
	}

	Success(c, nil)
}

// GetDetail 获取用户详情
// @Summary      获取用户详情
// @Description  根据用户ID获取用户详细信息
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        id  path  int64  true  "用户ID"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /users/{id} [get]
func (h *UserHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的用户ID")
		return
	}

	detail, err := h.userService.GetDetail(id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "获取用户详情失败")
		return
	}

	Success(c, detail)
}

// Delete 删除用户
// @Summary      删除用户
// @Description  根据用户ID删除用户，需要管理员权限
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        id  path  int64  true  "用户ID"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的用户ID")
		return
	}

	if err := h.userService.Delete(id); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "删除用户失败")
		return
	}

	Success(c, nil)
}

// UpdateStatus 更新用户状态（启用/禁用）
// @Summary      更新用户状态
// @Description  启用或禁用用户账号，需要管理员权限
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        id       path  int64                      true  "用户ID"
// @Param        request  body  UpdateUserStatusRequest     true  "状态更新请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /users/{id}/status [put]
func (h *UserHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的用户ID")
		return
	}

	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if req.Status != 0 && req.Status != 1 {
		Error(c, 400, "状态值无效")
		return
	}

	if err := h.userService.UpdateStatus(id, req.Status); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "更新状态失败")
		return
	}

	Success(c, nil)
}

// ResetPassword 重置密码
// @Summary      重置密码
// @Description  重置指定用户的密码，返回新密码，需要管理员权限
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        id  path  int64  true  "用户ID"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /users/{id}/reset-password [post]
func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的用户ID")
		return
	}

	newPassword, err := h.userService.ResetPassword(id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "重置密码失败")
		return
	}

	Success(c, gin.H{"new_password": newPassword})
}

// AssignRole 分配角色
// @Summary      分配角色
// @Description  为用户分配一个或多个角色，需要管理员权限
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        id       path  int64               true  "用户ID"
// @Param        request  body  AssignRoleRequest   true  "分配角色请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /users/{id}/roles [post]
func (h *UserHandler) AssignRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的用户ID")
		return
	}

	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if err := h.userService.AssignRole(id, req.RoleIDs); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "分配角色失败")
		return
	}

	Success(c, nil)
}
