package repository

import (
	"furniture-commission/internal/models"

	"gorm.io/gorm"
)

// CustomerRepository 客户Repository
type CustomerRepository struct {
	db *gorm.DB
}

// NewCustomerRepository 创建客户Repository实例
func NewCustomerRepository(db *gorm.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

// ListWithFilter 带条件分页查询客户列表
// salesmanID: 业务员ID，无关键字时用于过滤自己的客户；有关键字时精确匹配客户
func (r *CustomerRepository) ListWithFilter(storeID, keyword, level string, page, pageSize int, salesmanID int64) ([]models.Customer, int64, error) {
	db := r.db.Model(&models.Customer{})

	if storeID != "" {
		db = db.Where("store_id = ?", storeID)
	}

	if keyword != "" {
		// 有关键字时，精确匹配（完全相等）
		db = db.Where("customer_name = ? OR phone = ? OR customer_code = ?", keyword, keyword, keyword)
	} else if salesmanID > 0 {
		// 无关键字时，只显示该业务员的客户（created_by 或 salesman_id）
		db = db.Where("created_by = ? OR salesman_id = ?", salesmanID, salesmanID)
	}

	if level != "" {
		db = db.Where("level = ?", level)
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

	var customers []models.Customer
	err := db.Preload("Salesman").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").
		Find(&customers).Error
	if err != nil {
		return nil, 0, err
	}

	return customers, total, nil
}

// FindByID 根据ID查找客户
func (r *CustomerRepository) FindByID(id int64) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.Preload("Salesman").First(&customer, id).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// FindByPhone 根据手机号查找客户
func (r *CustomerRepository) FindByPhone(storeID int64, phone string) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.Where("store_id = ? AND phone = ?", storeID, phone).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// FindOrCreate 查找或创建客户
func (r *CustomerRepository) FindOrCreate(storeID int64, customerName, phone string) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.Where("store_id = ? AND phone = ?", storeID, phone).First(&customer).Error
	if err == nil {
		return &customer, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 创建新客户
	customer = models.Customer{
		StoreID:      storeID,
		CustomerName: customerName,
		Phone:        phone,
		Status:       1,
	}
	if err := r.db.Create(&customer).Error; err != nil {
		return nil, err
	}
	return &customer, nil
}

// Create 创建客户
func (r *CustomerRepository) Create(customer *models.Customer) error {
	return r.db.Create(customer).Error
}

// Update 更新客户
func (r *CustomerRepository) Update(customer *models.Customer) error {
	return r.db.Save(customer).Error
}

// Delete 删除客户
func (r *CustomerRepository) Delete(id int64) error {
	return r.db.Delete(&models.Customer{}, id).Error
}

// UpdateStats 更新客户累计统计
func (r *CustomerRepository) UpdateStats(customerID int64, totalOrders int, totalAmount, totalProfit float64) error {
	return r.db.Model(&models.Customer{}).Where("id = ?", customerID).Updates(map[string]interface{}{
		"total_orders":  gorm.Expr("total_orders + ?", totalOrders),
		"total_amount":  gorm.Expr("total_amount + ?", totalAmount),
		"total_profit":  gorm.Expr("total_profit + ?", totalProfit),
	}).Error
}
