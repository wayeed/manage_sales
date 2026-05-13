package repository

import (
	"furniture-commission/internal/models"

	"gorm.io/gorm"
)

// CategoryRepository 品类Repository
type CategoryRepository struct {
	*BaseRepository[models.Category]
}

// NewCategoryRepository 创建品类Repository实例
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{
		BaseRepository: NewBaseRepository[models.Category](db),
	}
}

// List 根据门店ID查询品类列表
func (r *CategoryRepository) List(storeID int64) ([]models.Category, error) {
	var categories []models.Category
	query := r.DB.Order("sort_order ASC, id ASC")
	if storeID > 0 {
		query = query.Where("store_id = ?", storeID)
	}
	err := query.Find(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}

// FindByCode 根据品类编码查找
func (r *CategoryRepository) FindByCode(storeID int64, code string) (*models.Category, error) {
	var category models.Category
	err := r.DB.Where("store_id = ? AND category_code = ?", storeID, code).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}
