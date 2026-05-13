package repository

import (
	"furniture-commission/internal/models"

	"gorm.io/gorm"
)

// OperationLogRepository 操作日志Repository
type OperationLogRepository struct {
	db *gorm.DB
}

// NewOperationLogRepository 创建操作日志Repository实例
func NewOperationLogRepository(db *gorm.DB) *OperationLogRepository {
	return &OperationLogRepository{db: db}
}

// List 分页查询操作日志
func (r *OperationLogRepository) List(keyword string, page, pageSize int) ([]models.OperationLog, int64, error) {
	db := r.db.Model(&models.OperationLog{})

	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("username LIKE ? OR action LIKE ? OR detail LIKE ?", like, like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var logs []models.OperationLog
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").
		Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// Create 创建操作日志
func (r *OperationLogRepository) Create(log *models.OperationLog) error {
	return r.db.Create(log).Error
}
