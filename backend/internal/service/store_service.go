package service

import (
	"errors"

	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	"furniture-commission/internal/repository"

	"gorm.io/gorm"
)

// StoreService 门店服务
type StoreService struct {
	storeRepo *repository.StoreRepository
}

// NewStoreService 创建门店服务实例
func NewStoreService(storeRepo *repository.StoreRepository) *StoreService {
	return &StoreService{storeRepo: storeRepo}
}

// CreateStoreRequest 创建门店请求
type CreateStoreRequest struct {
	StoreCode    string `json:"store_code" binding:"required"`
	StoreName    string `json:"store_name" binding:"required"`
	Address      string `json:"address" binding:"required"`
	ContactPhone string `json:"contact_phone"`
	Status       int8   `json:"status"`
}

// UpdateStoreRequest 更新门店请求
type UpdateStoreRequest struct {
	StoreName    string `json:"store_name"`
	Address      string `json:"address"`
	ContactPhone string `json:"contact_phone"`
	Status       *int8  `json:"status"`
}

// List 获取所有门店
func (s *StoreService) List() ([]models.Store, error) {
	stores, err := s.storeRepo.List()
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询门店列表失败"}
	}
	return stores, nil
}

// Create 创建门店
func (s *StoreService) Create(req *CreateStoreRequest) (*models.Store, error) {
	// 检查编码是否已存在
	if existing, _ := s.storeRepo.FindByCode(req.StoreCode); existing != nil {
		return nil, &AppError{Code: apperrors.ErrDuplicateKey, Message: "门店编码已存在"}
	}

	store := &models.Store{
		StoreCode:    req.StoreCode,
		StoreName:    req.StoreName,
		Address:      req.Address,
		ContactPhone: req.ContactPhone,
		Status:       req.Status,
	}
	if store.Status == 0 {
		store.Status = 1
	}

	if err := s.storeRepo.Create(store); err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "创建门店失败"}
	}
	return store, nil
}

// Update 更新门店
func (s *StoreService) Update(id int64, req *UpdateStoreRequest) error {
	store, err := s.storeRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "门店不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if req.StoreName != "" {
		store.StoreName = req.StoreName
	}
	if req.Address != "" {
		store.Address = req.Address
	}
	if req.ContactPhone != "" {
		store.ContactPhone = req.ContactPhone
	}
	if req.Status != nil {
		store.Status = *req.Status
	}

	if err := s.storeRepo.Update(store); err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "更新门店失败"}
	}
	return nil
}

// Delete 删除门店
func (s *StoreService) Delete(id int64) error {
	if err := s.storeRepo.Delete(id); err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "删除门店失败"}
	}
	return nil
}
