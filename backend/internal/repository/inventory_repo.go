package repository

import (
	"furniture-commission/internal/models"

	"gorm.io/gorm"
)

// InventoryRepository 库存Repository
type InventoryRepository struct {
	db *gorm.DB
}

// NewInventoryRepository 创建库存Repository实例
func NewInventoryRepository(db *gorm.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

// FindStockByWarehouseAndSKU 查找仓库SKU库存
func (r *InventoryRepository) FindStockByWarehouseAndSKU(warehouseID, skuID int64) (*models.WarehouseStock, error) {
	var stock models.WarehouseStock
	err := r.db.Where("warehouse_id = ? AND sku_id = ?", warehouseID, skuID).First(&stock).Error
	if err != nil {
		return nil, err
	}
	return &stock, nil
}

// FindBatchesBySKU 查找SKU的可用批次（按entry_date ASC）
func (r *InventoryRepository) FindBatchesBySKU(skuID int64, warehouseID int64) ([]models.InventoryBatch, error) {
	var batches []models.InventoryBatch
	query := r.db.Where("sku_id = ? AND status = 1", skuID)
	if warehouseID > 0 {
		query = query.Where("warehouse_id = ?", warehouseID)
	}
	err := query.Order("entry_date ASC, id ASC").Find(&batches).Error
	if err != nil {
		return nil, err
	}
	return batches, nil
}

// FindGiftStockByWarehouseAndGift 查找仓库礼品库存
func (r *InventoryRepository) FindGiftStockByWarehouseAndGift(warehouseID, giftID int64) (*models.WarehouseGiftStock, error) {
	var stock models.WarehouseGiftStock
	err := r.db.Where("warehouse_id = ? AND gift_id = ?", warehouseID, giftID).First(&stock).Error
	if err != nil {
		return nil, err
	}
	return &stock, nil
}

// FindGiftBatchesByGift 查找礼品的可用批次
func (r *InventoryRepository) FindGiftBatchesByGift(giftID int64, warehouseID int64) ([]models.GiftInventoryBatch, error) {
	var batches []models.GiftInventoryBatch
	query := r.db.Where("gift_id = ? AND status = 1", giftID)
	if warehouseID > 0 {
		query = query.Where("warehouse_id = ?", warehouseID)
	}
	err := query.Order("entry_date ASC, id ASC").Find(&batches).Error
	if err != nil {
		return nil, err
	}
	return batches, nil
}

// CreateStock 创建库存记录
func (r *InventoryRepository) CreateStock(stock *models.WarehouseStock) error {
	return r.db.Create(stock).Error
}

// CreateGiftStock 创建礼品库存记录
func (r *InventoryRepository) CreateGiftStock(stock *models.WarehouseGiftStock) error {
	return r.db.Create(stock).Error
}

// UpdateStockWithLock 乐观锁更新库存
func (r *InventoryRepository) UpdateStockWithLock(warehouseID, skuID int64, version int, updates map[string]interface{}) (int64, error) {
	result := r.db.Model(&models.WarehouseStock{}).
		Where("warehouse_id = ? AND sku_id = ? AND version = ?", warehouseID, skuID, version).
		Updates(updates)
	return result.RowsAffected, result.Error
}

// UpdateGiftStockWithLock 乐观锁更新礼品库存
func (r *InventoryRepository) UpdateGiftStockWithLock(warehouseID, giftID int64, version int, updates map[string]interface{}) (int64, error) {
	result := r.db.Model(&models.WarehouseGiftStock{}).
		Where("warehouse_id = ? AND gift_id = ? AND version = ?", warehouseID, giftID, version).
		Updates(updates)
	return result.RowsAffected, result.Error
}

// CreateTransaction 创建库存流水
func (r *InventoryRepository) CreateTransaction(tx *models.InventoryTransaction) error {
	return r.db.Create(tx).Error
}

// CreateBatch 创建库存批次
func (r *InventoryRepository) CreateBatch(batch *models.InventoryBatch) error {
	return r.db.Create(batch).Error
}

// CreateGiftBatch 创建礼品库存批次
func (r *InventoryRepository) CreateGiftBatch(batch *models.GiftInventoryBatch) error {
	return r.db.Create(batch).Error
}

// UpdateBatch 更新库存批次
func (r *InventoryRepository) UpdateBatch(batch *models.InventoryBatch) error {
	return r.db.Save(batch).Error
}

// UpdateGiftBatch 更新礼品库存批次
func (r *InventoryRepository) UpdateGiftBatch(batch *models.GiftInventoryBatch) error {
	return r.db.Save(batch).Error
}

// GetStockList 获取仓库库存列表（带分页和关键字搜索）
func (r *InventoryRepository) GetStockList(warehouseID int64, skuID int64, keyword string, page, pageSize int) ([]models.WarehouseStock, int64, error) {
	db := r.db.Model(&models.WarehouseStock{})
	if warehouseID > 0 {
		db = db.Where("warehouse_id = ?", warehouseID)
	}

	if skuID > 0 {
		db = db.Where("sku_id = ?", skuID)
	}

	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Joins("LEFT JOIN product_skus ON product_skus.id = warehouse_stocks.sku_id").
			Where("product_skus.sku_name LIKE ? OR product_skus.sku_code LIKE ? OR product_skus.barcode LIKE ?", like, like, like)
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
	if pageSize > 100 {
		pageSize = 100
	}

	var stocks []models.WarehouseStock
	err := db.Preload("SKU.Product").
		Preload("Warehouse").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").
		Find(&stocks).Error
	if err != nil {
		return nil, 0, err
	}

	return stocks, total, nil
}

// GetTransactionList 获取库存流水列表（带分页）
func (r *InventoryRepository) GetTransactionList(warehouseID int64, transactionType int8, startDate, endDate string, page, pageSize int) ([]models.InventoryTransaction, int64, error) {
	db := r.db.Model(&models.InventoryTransaction{})
	if warehouseID > 0 {
		db = db.Where("warehouse_id = ?", warehouseID)
	}
	if transactionType > 0 {
		db = db.Where("transaction_type = ?", transactionType)
	}
	if startDate != "" {
		db = db.Where("created_at >= ?", startDate+" 00:00:00")
	}
	if endDate != "" {
		db = db.Where("created_at <= ?", endDate+" 23:59:59")
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
	if pageSize > 100 {
		pageSize = 100
	}

	var transactions []models.InventoryTransaction
	err := db.Preload("SKU.Product").Preload("Warehouse").Preload("Creator").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").
		Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}
