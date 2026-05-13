package repository

import (
	"furniture-commission/internal/models"

	"gorm.io/gorm"
)

// RoleRepository 角色Repository
type RoleRepository struct {
	*BaseRepository[models.Role]
}

// NewRoleRepository 创建角色Repository实例
func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{
		BaseRepository: NewBaseRepository[models.Role](db),
	}
}

// FindByCode 根据角色编码查找角色
func (r *RoleRepository) FindByCode(code string) (*models.Role, error) {
	var role models.Role
	err := r.DB.Where("role_code = ?", code).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// List 获取所有角色列表
func (r *RoleRepository) List() ([]models.Role, error) {
	var roles []models.Role
	err := r.DB.Where("status = 1").Order("id ASC").Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// FindWithPermissions 根据ID查找角色（包含权限信息）
func (r *RoleRepository) FindWithPermissions(id int64) (*models.Role, error) {
	var role models.Role
	err := r.DB.Preload("Permissions").First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// CountUsersByRoleID 统计角色下的用户数量
func (r *RoleRepository) CountUsersByRoleID(roleID int64) (int64, error) {
	var count int64
	err := r.DB.Table("user_roles").Where("role_id = ?", roleID).Count(&count).Error
	return count, err
}
