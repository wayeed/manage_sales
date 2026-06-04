package repository

import (
	"furniture-commission/internal/models"

	"gorm.io/gorm"
)

type StocktakeRepository struct {
	db *gorm.DB
}

func NewStocktakeRepository(db *gorm.DB) *StocktakeRepository {
	return &StocktakeRepository{db: db}
}

// Create 创建盘点单
func (r *StocktakeRepository) Create(stocktake *models.Stocktake) error {
	return r.db.Create(stocktake).Error
}

// FindByID 根据ID查询盘点单
func (r *StocktakeRepository) FindByID(id int64) (*models.Stocktake, error) {
	var stocktake models.Stocktake
	err := r.db.Preload("Warehouse").Preload("Creator").Preload("Items.SKU.Product").First(&stocktake, id).Error
	return &stocktake, err
}

// List 分页查询盘点单列表
func (r *StocktakeRepository) List(storeID int64, warehouseID int64, status *int8, keyword string, page, pageSize int) ([]models.Stocktake, int64, error) {
	db := r.db.Model(&models.Stocktake{})

	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	if warehouseID > 0 {
		db = db.Where("warehouse_id = ?", warehouseID)
	}
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("stocktake_no LIKE ? OR remark LIKE ?", like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	var list []models.Stocktake
	err := db.Preload("Warehouse").Preload("Creator").
		Order("id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error

	return list, total, err
}

// Update 更新盘点单
func (r *StocktakeRepository) Update(stocktake *models.Stocktake) error {
	return r.db.Save(stocktake).Error
}

// CreateItem 创建盘点明细
func (r *StocktakeRepository) CreateItem(item *models.StocktakeItem) error {
	return r.db.Create(item).Error
}

// BatchCreateItems 批量创建盘点明细
func (r *StocktakeRepository) BatchCreateItems(items []models.StocktakeItem) error {
	return r.db.Create(&items).Error
}

// GetItemsByStocktakeID 获取盘点单的所有明细
func (r *StocktakeRepository) GetItemsByStocktakeID(stocktakeID int64) ([]models.StocktakeItem, error) {
	var items []models.StocktakeItem
	err := r.db.Preload("SKU.Product").Where("stocktake_id = ?", stocktakeID).Find(&items).Error
	return items, err
}

// DeleteItemsByStocktakeID 删除盘点单的所有明细
func (r *StocktakeRepository) DeleteItemsByStocktakeID(stocktakeID int64) error {
	return r.db.Where("stocktake_id = ?", stocktakeID).Delete(&models.StocktakeItem{}).Error
}
