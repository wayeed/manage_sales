package service

import (
	"errors"
	"fmt"
	"time"

	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	"furniture-commission/internal/repository"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// CreateGiftRequest 创建礼品请求
type CreateGiftRequest struct {
	StoreID int64 `json:"store_id" example:1`
	GiftCode string `json:"gift_code" binding:"required" example:"GIFT001"`
	GiftName string `json:"gift_name" binding:"required" example:"精美抱枕"`
	GiftImage string `json:"gift_image" example:"https://example.com/gift/pillow.jpg"`
	CostPrice float64 `json:"cost_price" example:50.00`
	WarningStock int `json:"warning_stock" example:20`
}

// UpdateGiftRequest 更新礼品请求
type UpdateGiftRequest struct {
	GiftName string `json:"gift_name" example:"精美抱枕"`
	GiftImage string `json:"gift_image" example:"https://example.com/gift/pillow.jpg"`
	CostPrice float64 `json:"cost_price" example:50.00`
	WarningStock int `json:"warning_stock" example:20`
	Status *int8 `json:"status" example:1`
}

// AddGiftStockRequest 增加礼品库存请求
type AddGiftStockRequest struct {
	WarehouseID int64 `json:"warehouse_id" binding:"required" example:1`
	Quantity int `json:"quantity" binding:"required,min=1" example:100`
	PurchasePrice float64 `json:"purchase_price" example:45.00`
}

// ListGiftRequest 礼品列表查询请求
type ListGiftRequest struct {
	StoreID int64 `form:"store_id" example:1`
	Status *int8 `form:"status" example:1`
	Keyword string `form:"keyword" example:"抱枕"`
	Page int `form:"page" example:1`
	PageSize int `form:"page_size" example:10`
}

// GiftService 礼品服务
type GiftService struct {
	db               *gorm.DB
	giftRepo         *repository.BaseRepository[models.Gift]
	inventoryService *InventoryService
}

// NewGiftService 创建礼品服务实例
func NewGiftService(db *gorm.DB, inventoryService *InventoryService) *GiftService {
	return &GiftService{
		db:               db,
		giftRepo:         repository.NewBaseRepository[models.Gift](db),
		inventoryService: inventoryService,
	}
}

// Create 创建礼品
func (s *GiftService) Create(req *CreateGiftRequest, createdBy int64) error {
	gift := &models.Gift{
		StoreID:      req.StoreID,
		GiftCode:     req.GiftCode,
		GiftName:     req.GiftName,
		GiftImage:    req.GiftImage,
		CostPrice:    decimal.NewFromFloat(req.CostPrice),
		WarningStock: req.WarningStock,
		Status:       1,
		CreatedBy:    &createdBy,
	}

	if err := s.db.Create(gift).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "创建礼品失败"}
	}
	return nil
}

// Update 更新礼品
func (s *GiftService) Update(id int64, req *UpdateGiftRequest) error {
	gift, err := s.giftRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "礼品不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if req.GiftName != "" {
		gift.GiftName = req.GiftName
	}
	gift.GiftImage = req.GiftImage
	if req.CostPrice > 0 {
		gift.CostPrice = decimal.NewFromFloat(req.CostPrice)
	}
	if req.WarningStock > 0 {
		gift.WarningStock = req.WarningStock
	}
	if req.Status != nil {
		gift.Status = *req.Status
	}

	if err := s.db.Save(gift).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "更新礼品失败"}
	}
	return nil
}

// Delete 删除礼品
func (s *GiftService) Delete(id int64) error {
	_, err := s.giftRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "礼品不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if err := s.db.Delete(&models.Gift{}, id).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "删除礼品失败"}
	}
	return nil
}

// List 获取礼品列表
func (s *GiftService) List(req *ListGiftRequest) (*PageResult, error) {
	db := s.db.Model(&models.Gift{})

	if req.StoreID > 0 {
		db = db.Where("store_id = ?", req.StoreID)
	}
	if req.Status != nil {
		db = db.Where("status = ?", *req.Status)
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		db = db.Where("gift_name LIKE ? OR gift_code LIKE ?", like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询礼品列表失败"}
	}

	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var gifts []models.Gift
	if err := db.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").
		Find(&gifts).Error; err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询礼品列表失败"}
	}

	return &PageResult{
		List:     gifts,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetDetail 获取礼品详情
func (s *GiftService) GetDetail(id int64) (*models.Gift, error) {
	gift, err := s.giftRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.NotFound, Message: "礼品不存在"}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}
	return gift, nil
}

// AddStock 增加礼品库存
func (s *GiftService) AddStock(id int64, req *AddGiftStockRequest, storeID int64, createdBy int64) error {
	gift, err := s.giftRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "礼品不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	batchNo := fmt.Sprintf("GIFT%s%d", time.Now().Format("20060102150405"), gift.ID)
	purchasePrice := decimal.NewFromFloat(req.PurchasePrice)
	if req.PurchasePrice <= 0 {
		purchasePrice = gift.CostPrice
	}

	if err := s.inventoryService.AddGiftStock(
		req.WarehouseID, gift.ID, req.Quantity,
		purchasePrice, batchNo, storeID, createdBy,
	); err != nil {
		return err
	}

	// 更新礼品总库存
	s.db.Model(gift).UpdateColumn("stock_quantity", gorm.Expr("stock_quantity + ?", req.Quantity))

	return nil
}
