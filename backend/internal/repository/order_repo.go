package repository

import (
	"furniture-commission/internal/models"

	"gorm.io/gorm"
)

// OrderRepository 订单Repository
type OrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository 创建订单Repository实例
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// ListWithFilter 带条件分页查询订单列表
func (r *OrderRepository) ListWithFilter(storeID, salesmanID, orderStatus, paymentStatus, orderType, startDate, endDate, keyword string, page, pageSize int) ([]models.Order, int64, error) {
	db := r.db.Model(&models.Order{})

	if storeID != "" {
		db = db.Where("store_id = ?", storeID)
	}
	if salesmanID != "" {
		db = db.Where("salesman_id = ?", salesmanID)
	}
	if orderStatus != "" {
		db = db.Where("order_status = ?", orderStatus)
	}
	if paymentStatus != "" {
		db = db.Where("payment_status = ?", paymentStatus)
	}
	if orderType != "" {
		db = db.Where("order_type = ?", orderType)
	}
	if startDate != "" {
		db = db.Where("order_date >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("order_date <= ?", endDate)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("order_no LIKE ? OR customer_name LIKE ? OR customer_phone LIKE ?", like, like, like)
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

	var orders []models.Order
	err := db.Preload("Salesman").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").
		Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// FindByID 根据ID查找订单（包含Items和Gifts）
func (r *OrderRepository) FindByID(id int64) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("Items").Preload("Gifts").Preload("Payments").Preload("Salesman").Preload("Customer").Preload("Peer").
		First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// FindByOrderNo 根据订单号查找订单
func (r *OrderRepository) FindByOrderNo(orderNo string) (*models.Order, error) {
	var order models.Order
	err := r.db.Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// Create 创建订单
func (r *OrderRepository) Create(order *models.Order) error {
	return r.db.Create(order).Error
}

// Update 更新订单
func (r *OrderRepository) Update(order *models.Order) error {
	return r.db.Save(order).Error
}

// UpdateFields 更新订单指定字段
func (r *OrderRepository) UpdateFields(id int64, fields map[string]interface{}) error {
	return r.db.Model(&models.Order{}).Where("id = ?", id).Updates(fields).Error
}

// GetOrderFeed 获取最近订单动态
func (r *OrderRepository) GetOrderFeed(userID int64, limit int) ([]models.Order, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	var orders []models.Order
	err := r.db.Where("salesman_id = ?", userID).
		Order("id DESC").
		Limit(limit).
		Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}
