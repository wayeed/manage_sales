package repository

import (
	"furniture-commission/internal/models"
	"furniture-commission/internal/pkg/pagination"

	"gorm.io/gorm"
)

// BaseRepository 通用Repository封装
type BaseRepository[T any] struct {
	DB *gorm.DB
}

// Create 创建记录
func (r *BaseRepository[T]) Create(entity *T) error {
	return r.DB.Create(entity).Error
}

// Update 更新记录
func (r *BaseRepository[T]) Update(entity *T) error {
	return r.DB.Save(entity).Error
}

// Delete 根据ID删除记录
func (r *BaseRepository[T]) Delete(id int64) error {
	entity := new(T)
	return r.DB.Delete(entity, id).Error
}

// FindByID 根据ID查找记录
func (r *BaseRepository[T]) FindByID(id int64) (*T, error) {
	var entity T
	err := r.DB.First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// FindPage 分页查询
func (r *BaseRepository[T]) FindPage(db *gorm.DB, pageReq *pagination.Pagination, list *[]T) (int64, error) {
	if db == nil {
		db = r.DB
	}

	var total int64
	if err := db.Model(new(T)).Count(&total).Error; err != nil {
		return 0, err
	}

	offset := pageReq.GetOffset()
	limit := pageReq.GetLimit()

	if err := db.Offset(offset).Limit(limit).Find(list).Error; err != nil {
		return 0, err
	}

	return total, nil
}

// NewBaseRepository 创建通用Repository实例
func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{DB: db}
}

// ModelName 获取模型表名
func ModelName[T any]() string {
	var entity T
	if m, ok := any(&entity).(interface{ TableName() string }); ok {
		return m.TableName()
	}
	return ""
}

// Ensure models are used (避免未使用导入警告)
var _ models.Store
var _ models.User
var _ models.Role
var _ models.Permission
var _ models.UserRole
var _ models.RolePermission
