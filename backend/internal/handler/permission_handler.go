package handler

import (
	"strconv"

	"furniture-commission/internal/repository"
	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// PermissionHandler 权限处理器
type PermissionHandler struct {
	permRepo    *repository.PermissionRepository
	roleService *service.RoleService
}

// NewPermissionHandler 创建权限处理器实例
func NewPermissionHandler(permRepo *repository.PermissionRepository, roleService *service.RoleService) *PermissionHandler {
	return &PermissionHandler{
		permRepo:    permRepo,
		roleService: roleService,
	}
}

// List 获取所有权限列表
// @Summary      获取权限列表
// @Description  获取所有权限列表
// @Tags         权限管理
// @Accept       json
// @Produce      json
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /permissions [get]
func (h *PermissionHandler) List(c *gin.Context) {
	permissions, err := h.permRepo.FindAll()
	if err != nil {
		Error(c, 500, "查询权限列表失败")
		return
	}

	Success(c, permissions)
}

// GetTree 获取权限树
// @Summary      获取权限树
// @Description  获取权限树形结构数据
// @Tags         权限管理
// @Accept       json
// @Produce      json
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /permissions/tree [get]
func (h *PermissionHandler) GetTree(c *gin.Context) {
	tree, err := h.permRepo.FindTree()
	if err != nil {
		Error(c, 500, "查询权限树失败")
		return
	}

	Success(c, tree)
}

// GetByRole 根据角色获取权限列表
// @Summary      获取角色权限
// @Description  根据角色ID获取该角色拥有的权限列表
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        id  path  int64  true  "角色ID"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /roles/{id}/permissions [get]
func (h *PermissionHandler) GetByRole(c *gin.Context) {
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
		Error(c, 500, "获取角色权限失败")
		return
	}

	Success(c, detail.Permissions)
}
