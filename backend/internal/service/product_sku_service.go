package service

import (
	"errors"

	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	"furniture-commission/internal/repository"

	"gorm.io/gorm"
)

// CreateSKURequest 创建SKU请求
type CreateSKURequest struct {
	ProductID int64 `json:"product_id" example:1`
	SKUCode string `json:"sku_code" binding:"required" example:"SKU001"`
	SKUName string `json:"sku_name" binding:"required" example:"真皮沙发-棕色-三座"`
	Attributes string `json:"attributes" example:"颜色:棕色;规格:三座"`
	Barcode string `json:"barcode" example:"6901234567890"`
}

// UpdateSKURequest 更新SKU请求
type UpdateSKURequest struct {
	SKUName string `json:"sku_name" example:"真皮沙发-棕色-三座"`
	Attributes string `json:"attributes" example:"颜色:棕色;规格:三座"`
	Barcode string `json:"barcode" example:"6901234567890"`
	Status *int8 `json:"status" example:1`
}

// SKUService SKU服务
type SKUService struct {
	db      *gorm.DB
	skuRepo *repository.SKURepository
}

// NewSKUService 创建SKU服务实例
func NewSKUService(db *gorm.DB, skuRepo *repository.SKURepository) *SKUService {
	return &SKUService{
		db:      db,
		skuRepo: skuRepo,
	}
}

// Create 创建SKU
func (s *SKUService) Create(req *CreateSKURequest) error {
	// 检查编码是否重复
	if existing, _ := s.skuRepo.FindByCode(req.SKUCode); existing != nil {
		return &AppError{Code: apperrors.ErrDuplicateKey, Message: "SKU编码已存在"}
	}

	sku := &models.ProductSKU{
		ProductID:  req.ProductID,
		SKUCode:    req.SKUCode,
		SKUName:    req.SKUName,
		Attributes: []byte(req.Attributes),
		Barcode:    req.Barcode,
		Status:     1,
	}

	if err := s.db.Create(sku).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "创建SKU失败"}
	}
	return nil
}

// Update 更新SKU
func (s *SKUService) Update(id int64, req *UpdateSKURequest) error {
	sku, err := s.skuRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "SKU不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if req.SKUName != "" {
		sku.SKUName = req.SKUName
	}
	if req.Attributes != "" {
		sku.Attributes = []byte(req.Attributes)
	}
	if req.Barcode != "" {
		sku.Barcode = req.Barcode
	}
	if req.Status != nil {
		sku.Status = *req.Status
	}

	if err := s.db.Save(sku).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "更新SKU失败"}
	}
	return nil
}

// Delete 删除SKU
func (s *SKUService) Delete(id int64) error {
	_, err := s.skuRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "SKU不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if err := s.db.Delete(&models.ProductSKU{}, id).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "删除SKU失败"}
	}
	return nil
}

// ListByProduct 获取商品下的SKU列表
func (s *SKUService) ListByProduct(productID int64) ([]models.ProductSKU, error) {
	skus, err := s.skuRepo.FindByProductID(productID)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询SKU列表失败"}
	}
	return skus, nil
}

// ListAll 获取所有SKU列表（支持搜索）
func (s *SKUService) ListAll(keyword string, page, pageSize int) ([]models.ProductSKU, int64, error) {
	db := s.db.Model(&models.ProductSKU{}).Preload("Product")

	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("sku_code LIKE ? OR sku_name LIKE ? OR barcode LIKE ?", like, like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, &AppError{Code: apperrors.InternalError, Message: "查询SKU列表失败"}
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

	var skus []models.ProductSKU
	if err := db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&skus).Error; err != nil {
		return nil, 0, &AppError{Code: apperrors.InternalError, Message: "查询SKU列表失败"}
	}

	return skus, total, nil
}

// SKUWithStock 带库存的SKU
type SKUWithStock struct {
	models.ProductSKU
	AvailableStock int `json:"available_stock" example:50`
}

// ListWithStock 获取带库存信息的SKU列表（用于订单选商品）
func (s *SKUService) ListWithStock(storeID int64, keyword string, page, pageSize int) ([]SKUWithStock, int64, error) {
	// 子查询：汇总门店所有仓库的可用库存
	stockSubQuery := s.db.Model(&models.WarehouseStock{}).
		Select("sku_id, SUM(available_quantity) as total_available").
		Joins("JOIN warehouses ON warehouses.id = warehouse_stocks.warehouse_id").
		Where("warehouses.store_id = ? AND warehouses.status = 1", storeID).
		Group("sku_id")

	db := s.db.Model(&models.ProductSKU{}).
		Select("product_skus.*, COALESCE(ws.total_available, 0) as available_stock").
		Joins("LEFT JOIN (?) ws ON ws.sku_id = product_skus.id", stockSubQuery).
		Preload("Product").
		Where("product_skus.status = 1")

	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("product_skus.sku_code LIKE ? OR product_skus.sku_name LIKE ? OR product_skus.barcode LIKE ?", like, like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, &AppError{Code: apperrors.InternalError, Message: "查询SKU列表失败"}
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

	var skus []SKUWithStock
	if err := db.Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&skus).Error; err != nil {
		return nil, 0, &AppError{Code: apperrors.InternalError, Message: "查询SKU列表失败"}
	}

	return skus, total, nil
}
