package service

import (
	"errors"
	"fmt"

	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	"furniture-commission/internal/repository"

	"gorm.io/gorm"
)

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	RoleCode string `json:"role_code" binding:"required" example:"admin"`
	RoleName string `json:"role_name" binding:"required" example:"管理员"`
	Description string `json:"description" example:"系统管理员角色"`
	Status int8 `json:"status" example:1`
	PermissionIDs []int64 `json:"permission_ids" example:[1, 2, 3]`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	RoleCode string `json:"role_code" example:"admin"`
	RoleName string `json:"role_name" example:"管理员"`
	Description string `json:"description" example:"系统管理员角色"`
	Status *int8 `json:"status" example:1`
	PermissionIDs []int64 `json:"permission_ids" example:[1, 2, 3]`
}

// RoleVO 角色视图对象
type RoleVO struct {
	ID int64 `json:"id" example:1`
	RoleCode string `json:"role_code" example:"admin"`
	RoleName string `json:"role_name" example:"管理员"`
	Description string `json:"description" example:"系统管理员角色"`
	Status int8 `json:"status" example:1`
}

// RoleDetail 角色详情
type RoleDetail struct {
	*RoleVO
	Permissions []models.Permission `json:"permissions" example:[]`
}

// RoleService 角色服务
type RoleService struct {
	db       *gorm.DB
	roleRepo *repository.RoleRepository
	permRepo *repository.PermissionRepository
}

// NewRoleService 创建角色服务实例
func NewRoleService(db *gorm.DB, roleRepo *repository.RoleRepository, permRepo *repository.PermissionRepository) *RoleService {
	return &RoleService{
		db:       db,
		roleRepo: roleRepo,
		permRepo: permRepo,
	}
}

// Create 创建角色
func (s *RoleService) Create(req *CreateRoleRequest) error {
	// 检查角色编码是否已存在
	if existing, _ := s.roleRepo.FindByCode(req.RoleCode); existing != nil {
		return &AppError{Code: apperrors.ErrDuplicateKey, Message: "角色编码已存在"}
	}

	role := &models.Role{
		RoleCode:    req.RoleCode,
		RoleName:    req.RoleName,
		Description: req.Description,
		Status:      req.Status,
	}

	if req.Status == 0 {
		role.Status = 1
	}

	// 创建角色
	if err := s.db.Create(role).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "创建角色失败"}
	}

	// 分配权限
	if len(req.PermissionIDs) > 0 {
		if err := s.assignPermissions(role.ID, req.PermissionIDs); err != nil {
			return err
		}
	}

	return nil
}

// Update 更新角色
func (s *RoleService) Update(id int64, req *UpdateRoleRequest) error {
	role, err := s.roleRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: apperrors.GetMessage(apperrors.NotFound)}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	// 检查角色编码是否被其他人使用
	if req.RoleCode != "" && req.RoleCode != role.RoleCode {
		if existing, _ := s.roleRepo.FindByCode(req.RoleCode); existing != nil && existing.ID != id {
			return &AppError{Code: apperrors.ErrDuplicateKey, Message: "角色编码已存在"}
		}
		role.RoleCode = req.RoleCode
	}

	if req.RoleName != "" {
		role.RoleName = req.RoleName
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	if req.Status != nil {
		role.Status = *req.Status
	}

	if err := s.db.Save(role).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "更新角色失败"}
	}

	// 如果有权限ID，则更新权限
	if req.PermissionIDs != nil {
		if err := s.assignPermissions(role.ID, req.PermissionIDs); err != nil {
			return err
		}
	}

	return nil
}

// Delete 删除角色
func (s *RoleService) Delete(id int64) error {
	// 检查角色是否存在
	if _, err := s.roleRepo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: apperrors.GetMessage(apperrors.NotFound)}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	// 检查是否有用户关联此角色
	count, err := s.roleRepo.CountUsersByRoleID(id)
	if err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}
	if count > 0 {
		return &AppError{Code: apperrors.Forbidden, Message: fmt.Sprintf("该角色下有 %d 个用户，无法删除", count)}
	}

	// 删除角色权限关联
	if err := s.db.Where("role_id = ?", id).Delete(&models.RolePermission{}).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "删除角色权限关联失败"}
	}

	// 删除角色
	if err := s.db.Delete(&models.Role{}, id).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "删除角色失败"}
	}

	return nil
}

// List 获取角色列表
func (s *RoleService) List() ([]RoleVO, error) {
	roles, err := s.roleRepo.List()
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询角色列表失败"}
	}

	result := make([]RoleVO, len(roles))
	for i, role := range roles {
		result[i] = RoleVO{
			ID:          role.ID,
			RoleCode:    role.RoleCode,
			RoleName:    role.RoleName,
			Description: role.Description,
			Status:      role.Status,
		}
	}

	return result, nil
}

// GetDetail 获取角色详情
func (s *RoleService) GetDetail(id int64) (*RoleDetail, error) {
	role, err := s.roleRepo.FindWithPermissions(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.NotFound, Message: apperrors.GetMessage(apperrors.NotFound)}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	return &RoleDetail{
		RoleVO: &RoleVO{
			ID:          role.ID,
			RoleCode:    role.RoleCode,
			RoleName:    role.RoleName,
			Description: role.Description,
			Status:      role.Status,
		},
		Permissions: role.Permissions,
	}, nil
}

// AssignPermissions 分配权限
func (s *RoleService) AssignPermissions(roleID int64, permIDs []int64) error {
	// 检查角色是否存在
	if _, err := s.roleRepo.FindByID(roleID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: apperrors.GetMessage(apperrors.NotFound)}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	return s.assignPermissions(roleID, permIDs)
}

// assignPermissions 内部方法：分配权限
func (s *RoleService) assignPermissions(roleID int64, permIDs []int64) error {
	// 先清除已有权限
	if err := s.db.Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "清除权限失败"}
	}

	// 批量添加新权限
	if len(permIDs) > 0 {
		rolePerms := make([]models.RolePermission, len(permIDs))
		for i, permID := range permIDs {
			rolePerms[i] = models.RolePermission{
				RoleID:       roleID,
				PermissionID: permID,
			}
		}
		if err := s.db.Create(&rolePerms).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "分配权限失败"}
		}
	}

	return nil
}
