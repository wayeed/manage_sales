package repository

import (
	"furniture-commission/internal/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// PaymentRepository 回款Repository
type PaymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository 创建回款Repository实例
func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Create 创建回款记录
func (r *PaymentRepository) Create(payment *models.Payment) error {
	return r.db.Create(payment).Error
}

// FindByOrderID 根据订单ID查找回款记录
func (r *PaymentRepository) FindByOrderID(orderID int64) ([]models.Payment, error) {
	var payments []models.Payment
	err := r.db.Where("order_id = ?", orderID).Order("id DESC").Find(&payments).Error
	if err != nil {
		return nil, err
	}
	return payments, nil
}

// FindByID 根据ID查找回款记录
func (r *PaymentRepository) FindByID(id int64) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.First(&payment, id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// SumPaidByOrderID 计算订单已回款总额（仅已审核通过的）
func (r *PaymentRepository) SumPaidByOrderID(orderID int64) (decimal.Decimal, error) {
	var total decimal.Decimal
	err := r.db.Model(&models.Payment{}).
		Where("order_id = ? AND status = 1", orderID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	if err != nil {
		return decimal.Zero, err
	}
	return total, nil
}

// ListWithFilter 带条件分页查询回款列表
func (r *PaymentRepository) ListWithFilter(orderID, status, startDate, endDate string, page, pageSize int) ([]models.Payment, int64, error) {
	db := r.db.Model(&models.Payment{})

	if orderID != "" {
		db = db.Where("order_id = ?", orderID)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if startDate != "" {
		db = db.Where("payment_date >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("payment_date <= ?", endDate)
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

	var payments []models.Payment
	err := db.Preload("Order").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").
		Find(&payments).Error
	if err != nil {
		return nil, 0, err
	}

	return payments, total, nil
}

// UpdateFields 更新回款指定字段
func (r *PaymentRepository) UpdateFields(id int64, fields map[string]interface{}) error {
	return r.db.Model(&models.Payment{}).Where("id = ?", id).Updates(fields).Error
}
