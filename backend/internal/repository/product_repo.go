package repository

import (
	"furniture-commission/internal/models"

	"gorm.io/gorm"
)

// ProductRepository 商品Repository
type ProductRepository struct {
	*BaseRepository[models.Product]
}

// NewProductRepository 创建商品Repository实例
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{
		BaseRepository: NewBaseRepository[models.Product](db),
	}
}

// ListWithFilter 根据条件分页查询商品列表
func (r *ProductRepository) ListWithFilter(storeID int64, categoryID *int64, status *int8, keyword string, page, pageSize int) ([]models.Product, int64, error) {
	db := r.DB.Model(&models.Product{})

	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	if categoryID != nil && *categoryID > 0 {
		db = db.Where("category_id = ?", *categoryID)
	}
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("product_name LIKE ? OR product_code LIKE ? OR brand LIKE ?", like, like, like)
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

	var products []models.Product
	if err := db.Preload("Category").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").
		Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// FindByCode 根据商品编码查找
func (r *ProductRepository) FindByCode(storeID int64, code string) (*models.Product, error) {
	var product models.Product
	err := r.DB.Where("store_id = ? AND product_code = ?", storeID, code).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// FindByIDWithSKUs 根据ID查找商品（包含SKU）
func (r *ProductRepository) FindByIDWithSKUs(id int64) (*models.Product, error) {
	var product models.Product
	err := r.DB.Preload("Category").Preload("SKUs").First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}
