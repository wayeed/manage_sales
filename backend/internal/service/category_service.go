package service

import (
	"errors"

	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	"furniture-commission/internal/repository"

	"gorm.io/gorm"
)

// CreateCategoryRequest 创建品类请求
type CreateCategoryRequest struct {
	StoreID int64 `json:"store_id" example:1`
	CategoryCode string `json:"category_code" binding:"required" example:"CAT001"`
	CategoryName string `json:"category_name" binding:"required" example:"沙发"`
	SortOrder int `json:"sort_order" example:1`
}

// UpdateCategoryRequest 更新品类请求
type UpdateCategoryRequest struct {
	CategoryName string `json:"category_name" example:"沙发"`
	SortOrder int `json:"sort_order" example:1`
	Status *int8 `json:"status" example:1`
}

// CategoryService 品类服务
type CategoryService struct {
	categoryRepo *repository.CategoryRepository
}

// NewCategoryService 创建品类服务实例
func NewCategoryService(categoryRepo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

// Create 创建品类
func (s *CategoryService) Create(req *CreateCategoryRequest) error {
	// 检查编码是否重复
	if existing, _ := s.categoryRepo.FindByCode(req.StoreID, req.CategoryCode); existing != nil {
		return &AppError{Code: apperrors.ErrDuplicateKey, Message: "品类编码已存在"}
	}

	category := &models.Category{
		StoreID:      req.StoreID,
		CategoryCode: req.CategoryCode,
		CategoryName: req.CategoryName,
		SortOrder:    req.SortOrder,
		Status:       1,
	}

	if err := s.categoryRepo.Create(category); err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "创建品类失败"}
	}
	return nil
}

// Update 更新品类
func (s *CategoryService) Update(id int64, req *UpdateCategoryRequest) error {
	category, err := s.categoryRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "品类不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if req.CategoryName != "" {
		category.CategoryName = req.CategoryName
	}
	if req.SortOrder > 0 {
		category.SortOrder = req.SortOrder
	}
	if req.Status != nil {
		category.Status = *req.Status
	}

	if err := s.categoryRepo.Update(category); err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "更新品类失败"}
	}
	return nil
}

// Delete 删除品类
func (s *CategoryService) Delete(id int64) error {
	_, err := s.categoryRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "品类不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if err := s.categoryRepo.Delete(id); err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "删除品类失败"}
	}
	return nil
}

// List 获取品类列表
func (s *CategoryService) List(storeID int64) ([]models.Category, error) {
	categories, err := s.categoryRepo.List(storeID)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询品类列表失败"}
	}
	return categories, nil
}
