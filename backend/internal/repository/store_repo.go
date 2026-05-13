package repository

import (
	"furniture-commission/internal/models"

	"gorm.io/gorm"
)

// StoreRepository 门店Repository
type StoreRepository struct {
	db *gorm.DB
}

// NewStoreRepository 创建门店Repository实例
func NewStoreRepository(db *gorm.DB) *StoreRepository {
	return &StoreRepository{db: db}
}

// List 获取所有门店
func (r *StoreRepository) List() ([]models.Store, error) {
	var stores []models.Store
	err := r.db.Order("id ASC").Find(&stores).Error
	if err != nil {
		return nil, err
	}
	return stores, nil
}

// FindByID 根据ID查找门店
func (r *StoreRepository) FindByID(id int64) (*models.Store, error) {
	var store models.Store
	err := r.db.First(&store, id).Error
	if err != nil {
		return nil, err
	}
	return &store, nil
}

// FindByCode 根据编码查找门店
func (r *StoreRepository) FindByCode(code string) (*models.Store, error) {
	var store models.Store
	err := r.db.Where("store_code = ?", code).First(&store).Error
	if err != nil {
		return nil, err
	}
	return &store, nil
}

// Create 创建门店
func (r *StoreRepository) Create(store *models.Store) error {
	return r.db.Create(store).Error
}

// Update 更新门店
func (r *StoreRepository) Update(store *models.Store) error {
	return r.db.Save(store).Error
}

// Delete 删除门店
func (r *StoreRepository) Delete(id int64) error {
	return r.db.Delete(&models.Store{}, id).Error
}
