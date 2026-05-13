package handler

import (
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// AssignPermissionsRequest 分配权限请求
type AssignPermissionsRequest struct {
	PermissionIDs []int64 `json:"permission_ids" binding:"required" example:"1,2,3"`
}

// RoleHandler 角色处理器
type RoleHandler struct {
	roleService *service.RoleService
}

// NewRoleHandler 创建角色处理器实例
func NewRoleHandler(roleService *service.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

// List 获取角色列表
// @Summary      获取角色列表
// @Description  获取所有角色列表
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /roles [get]
func (h *RoleHandler) List(c *gin.Context) {
	roles, err := h.roleService.List()
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询角色列表失败")
		return
	}

	Success(c, roles)
}

// Create 创建角色
// @Summary      创建角色
// @Description  创建新角色，需要管理员权限
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        request  body  service.CreateRoleRequest  true  "创建角色请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /roles [post]
func (h *RoleHandler) Create(c *gin.Context) {
	var req service.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if err := h.roleService.Create(&req); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "创建角色失败")
		return
	}

	Success(c, nil)
}

// Update 更新角色
// @Summary      更新角色
// @Description  根据角色ID更新角色信息，需要管理员权限
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        id       path  int64                    true  "角色ID"
// @Param        request  body  service.UpdateRoleRequest  true  "更新角色请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /roles/{id} [put]
func (h *RoleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的角色ID")
		return
	}

	var req service.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if err := h.roleService.Update(id, &req); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "更新角色失败")
		return
	}

	Success(c, nil)
}

// Delete 删除角色
// @Summary      删除角色
// @Description  根据角色ID删除角色，需要管理员权限
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        id  path  int64  true  "角色ID"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /roles/{id} [delete]
func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的角色ID")
		return
	}

	if err := h.roleService.Delete(id); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "删除角色失败")
		return
	}

	Success(c, nil)
}

// GetDetail 获取角色详情
// @Summary      获取角色详情
// @Description  根据角色ID获取角色详细信息
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        id  path  int64  true  "角色ID"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /roles/{id} [get]
func (h *RoleHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的角色ID")
		return
	}

	detail, err := h.roleService.GetDetail(id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "获取角色详情失败")
		return
	}

	Success(c, detail)
}

// AssignPermissions 分配权限
// @Summary      分配权限
// @Description  为角色分配权限，需要管理员权限
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        id       path  int64                      true  "角色ID"
// @Param        request  body  AssignPermissionsRequest    true  "分配权限请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /roles/{id}/permissions [post]
func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的角色ID")
		return
	}

	var req AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if err := h.roleService.AssignPermissions(id, req.PermissionIDs); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "分配权限失败")
		return
	}

	Success(c, nil)
}
