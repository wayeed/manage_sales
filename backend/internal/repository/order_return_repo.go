package repository

import (
	"furniture-commission/internal/models"
	"gorm.io/gorm"
)

type OrderReturnRepository struct {
	db *gorm.DB
}

func NewOrderReturnRepository(db *gorm.DB) *OrderReturnRepository {
	return &OrderReturnRepository{db: db}
}

// Create 创建退货记录
func (r *OrderReturnRepository) Create(record *models.OrderReturn) error {
	return r.db.Create(record).Error
}

// FindByOrderID 根据订单ID查询退货记录
func (r *OrderReturnRepository) FindByOrderID(orderID int64) (*models.OrderReturn, error) {
	var record models.OrderReturn
	err := r.db.Where("order_id = ?", orderID).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// FindByReturnNo 根据退货单号查询
func (r *OrderReturnRepository) FindByReturnNo(returnNo string) (*models.OrderReturn, error) {
	var record models.OrderReturn
	err := r.db.Where("return_no = ?", returnNo).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// ListByStoreID 查询门店的所有退货记录
func (r *OrderReturnRepository) ListByStoreID(storeID int64, page, pageSize int) ([]models.OrderReturn, int64, error) {
	var records []models.OrderReturn
	var total int64

	query := r.db.Model(&models.OrderReturn{}).Where("store_id = ?", storeID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}
