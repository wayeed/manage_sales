package repository

import (
	"furniture-commission/internal/models"

	"gorm.io/gorm"
)

// PermissionRepository 权限Repository
type PermissionRepository struct {
	*BaseRepository[models.Permission]
}

// NewPermissionRepository 创建权限Repository实例
func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{
		BaseRepository: NewBaseRepository[models.Permission](db),
	}
}

// FindByRoleID 根据角色ID查找权限列表
func (r *PermissionRepository) FindByRoleID(roleID int64) ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.DB.
		Joins("INNER JOIN role_permissions rp ON rp.permission_id = permissions.id").
		Where("rp.role_id = ? AND permissions.status = 1", roleID).
		Order("permissions.sort_order ASC, permissions.id ASC").
		Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// FindMenuByRoleID 根据角色ID查找菜单权限（permission_type=1）
func (r *PermissionRepository) FindMenuByRoleID(roleID int64) ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.DB.
		Joins("INNER JOIN role_permissions rp ON rp.permission_id = permissions.id").
		Where("rp.role_id = ? AND permissions.status = 1 AND permissions.perm_type = 1", roleID).
		Order("permissions.sort_order ASC, permissions.id ASC").
		Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// FindByUserID 根据用户ID查找所有权限（通过角色关联）
func (r *PermissionRepository) FindByUserID(userID int64) ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.DB.
		Joins("INNER JOIN role_permissions rp ON rp.permission_id = permissions.id").
		Joins("INNER JOIN user_roles ur ON ur.role_id = rp.role_id").
		Where("ur.user_id = ? AND permissions.status = 1", userID).
		Group("permissions.id").
		Order("permissions.sort_order ASC, permissions.id ASC").
		Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// FindAll 获取所有权限列表
func (r *PermissionRepository) FindAll() ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.DB.Where("status = 1").Order("sort_order ASC, id ASC").Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// FindTree 获取权限树（只获取顶级节点，子节点通过Children关联加载）
func (r *PermissionRepository) FindTree() ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.DB.Where("status = 1 AND parent_id = 0").
		Preload("Children", func(db *gorm.DB) *gorm.DB {
			return db.Where("status = 1").Order("sort_order ASC, id ASC")
		}).
		Order("sort_order ASC, id ASC").
		Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}
