package repository

import (
	"furniture-commission/internal/models"

	"gorm.io/gorm"
)

// WarehouseRepository 仓库Repository
type WarehouseRepository struct {
	*BaseRepository[models.Warehouse]
}

// NewWarehouseRepository 创建仓库Repository实例
func NewWarehouseRepository(db *gorm.DB) *WarehouseRepository {
	return &WarehouseRepository{
		BaseRepository: NewBaseRepository[models.Warehouse](db),
	}
}

// List 根据门店ID查询仓库列表
func (r *WarehouseRepository) List(storeID int64) ([]models.Warehouse, error) {
	var warehouses []models.Warehouse
	db := r.DB
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	err := db.Order("id ASC").Find(&warehouses).Error
	if err != nil {
		return nil, err
	}
	return warehouses, nil
}

// FindByCode 根据仓库编码查找
func (r *WarehouseRepository) FindByCode(storeID int64, code string) (*models.Warehouse, error) {
	var warehouse models.Warehouse
	err := r.DB.Where("store_id = ? AND warehouse_code = ?", storeID, code).First(&warehouse).Error
	if err != nil {
		return nil, err
	}
	return &warehouse, nil
}
