package repository

import (
	"furniture-commission/internal/models"

	"gorm.io/gorm"
)

// SKURepository SKU Repository
type SKURepository struct {
	*BaseRepository[models.ProductSKU]
}

// NewSKURepository 创建SKU Repository实例
func NewSKURepository(db *gorm.DB) *SKURepository {
	return &SKURepository{
		BaseRepository: NewBaseRepository[models.ProductSKU](db),
	}
}

// FindByProductID 根据商品ID查找SKU列表
func (r *SKURepository) FindByProductID(productID int64) ([]models.ProductSKU, error) {
	var skus []models.ProductSKU
	err := r.DB.Where("product_id = ?", productID).Order("id ASC").Find(&skus).Error
	if err != nil {
		return nil, err
	}
	return skus, nil
}

// FindByCode 根据SKU编码查找
func (r *SKURepository) FindByCode(code string) (*models.ProductSKU, error) {
	var sku models.ProductSKU
	err := r.DB.Where("sku_code = ?", code).First(&sku).Error
	if err != nil {
		return nil, err
	}
	return &sku, nil
}
