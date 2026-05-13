package service

import (
	"errors"
	"fmt"
	"time"

	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	appsnow "furniture-commission/internal/pkg/snowflake"
	"furniture-commission/internal/repository"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// CreateTransferOrderRequest 创建调拨单请求
type CreateTransferOrderRequest struct {
	StoreID int64 `json:"store_id" example:1`
	FromWarehouseID int64 `json:"from_warehouse_id" binding:"required" example:1`
	ToWarehouseID int64 `json:"to_warehouse_id" binding:"required" example:2`
	Remark string `json:"remark" example:"门店间调拨"`
	Items []CreateTransferItemRequest `json:"items" binding:"required,min=1" example:[]`
}

// CreateTransferItemRequest 创建调拨明细请求
type CreateTransferItemRequest struct {
	SKUID int64 `json:"sku_id" binding:"required" example:1`
	Quantity int `json:"quantity" binding:"required,min=1" example:5`
}

// ListTransferOrderRequest 调拨单列表查询请求
type ListTransferOrderRequest struct {
	StoreID int64 `form:"store_id" example:1`
	Status *int8 `form:"status" example:0`
	Page int `form:"page" example:1`
	PageSize int `form:"page_size" example:10`
}

// TransferService 调拨服务
type TransferService struct {
	db               *gorm.DB
	transferRepo     *repository.BaseRepository[models.TransferOrder]
	transferItemRepo *repository.BaseRepository[models.TransferItem]
	inventoryService *InventoryService
}

// NewTransferService 创建调拨服务实例
func NewTransferService(db *gorm.DB, inventoryService *InventoryService) *TransferService {
	return &TransferService{
		db:               db,
		transferRepo:     repository.NewBaseRepository[models.TransferOrder](db),
		transferItemRepo: repository.NewBaseRepository[models.TransferItem](db),
		inventoryService: inventoryService,
	}
}

// CreateOrder 创建调拨单
func (s *TransferService) CreateOrder(req *CreateTransferOrderRequest, createdBy int64) error {
	if req.FromWarehouseID == req.ToWarehouseID {
		return &AppError{Code: apperrors.BadRequest, Message: "调出和调入仓库不能相同"}
	}

	transferNo := "TR" + appsnow.GenerateOrderNo()

	totalQuantity := 0
	for _, item := range req.Items {
		totalQuantity += item.Quantity
	}

	order := &models.TransferOrder{
		StoreID:         req.StoreID,
		TransferNo:      transferNo,
		FromWarehouseID: &req.FromWarehouseID,
		ToWarehouseID:   &req.ToWarehouseID,
		TotalQuantity:   totalQuantity,
		Status:          0, // 待审核
		Remark:          req.Remark,
		CreatedBy:       &createdBy,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		for _, item := range req.Items {
			ti := &models.TransferItem{
				TransferOrderID: order.ID,
				SKUID:           &item.SKUID,
				Quantity:        item.Quantity,
			}
			if err := tx.Create(ti).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "创建调拨单失败"}
	}
	return nil
}

// ApproveOrder 审核调拨单（status 0->1）
func (s *TransferService) ApproveOrder(id int64, auditedBy int64) error {
	order, err := s.transferRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrOrderNotFound, Message: "调拨单不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if order.Status != 0 {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "调拨单状态不允许审核"}
	}

	now := time.Now()
	if err := s.db.Model(order).Updates(map[string]interface{}{
		"status":     1,
		"audited_by": auditedBy,
		"audited_at": now,
	}).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "审核调拨单失败"}
	}

	return nil
}

// ConfirmOut 确认出库（status 1->2）
func (s *TransferService) ConfirmOut(id int64, createdBy int64) error {
	order, err := s.transferRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrOrderNotFound, Message: "调拨单不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if order.Status != 1 {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "调拨单状态不允许出库"}
	}

	// 查询调拨明细
	var items []models.TransferItem
	if err := s.db.Where("transfer_order_id = ?", order.ID).Find(&items).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "查询调拨明细失败"}
	}

	// 锁定调出仓库库存
	for _, item := range items {
		if item.SKUID == nil {
			continue
		}
		if err := s.inventoryService.LockStock(*order.FromWarehouseID, *item.SKUID, item.Quantity); err != nil {
			return err
		}
	}

	if err := s.db.Model(order).Update("status", 2).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "确认出库失败"}
	}

	return nil
}

// ConfirmIn 确认入库（status 2->3）
func (s *TransferService) ConfirmIn(id int64, createdBy int64) error {
	order, err := s.transferRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrOrderNotFound, Message: "调拨单不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if order.Status != 2 {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "调拨单状态不允许入库"}
	}

	// 查询调拨明细
	var items []models.TransferItem
	if err := s.db.Where("transfer_order_id = ?", order.ID).Find(&items).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "查询调拨明细失败"}
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if item.SKUID == nil {
				continue
			}

			// 调用TransferStock完成调拨（扣减调出+增加调入）
			if err := s.inventoryService.TransferStock(
				*order.FromWarehouseID, *order.ToWarehouseID,
				*item.SKUID, item.Quantity,
				order.StoreID, createdBy,
			); err != nil {
				return err
			}
		}

		// 更新调拨单状态
		if err := tx.Model(order).Update("status", 3).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if appErr, ok := err.(*AppError); ok {
			return appErr
		}
		return &AppError{Code: apperrors.InternalError, Message: "确认入库失败"}
	}

	return nil
}

// CancelOrder 取消调拨单
func (s *TransferService) CancelOrder(id int64) error {
	order, err := s.transferRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrOrderNotFound, Message: "调拨单不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if order.Status == 3 {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "已完成调拨不能取消"}
	}
	if order.Status == 4 {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "调拨单已取消"}
	}

	// 如果已出库但未入库，释放锁定的库存
	if order.Status == 2 {
		var items []models.TransferItem
		if err := s.db.Where("transfer_order_id = ?", order.ID).Find(&items).Error; err == nil {
			for _, item := range items {
				if item.SKUID != nil && order.FromWarehouseID != nil {
					s.inventoryService.UnlockStock(*order.FromWarehouseID, *item.SKUID, item.Quantity)
				}
			}
		}
	}

	if err := s.db.Model(order).Update("status", 4).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "取消调拨单失败"}
	}

	return nil
}

// List 调拨单列表
func (s *TransferService) List(req *ListTransferOrderRequest) (*PageResult, error) {
	db := s.db.Model(&models.TransferOrder{})

	if req.StoreID > 0 {
		db = db.Where("store_id = ?", req.StoreID)
	}
	if req.Status != nil {
		db = db.Where("status = ?", *req.Status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询调拨单列表失败"}
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

	var orders []models.TransferOrder
	if err := db.Preload("FromWarehouse").Preload("ToWarehouse").Preload("Items.SKU.Product").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").
		Find(&orders).Error; err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询调拨单列表失败"}
	}

	return &PageResult{
		List:     orders,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetDetail 调拨单详情
func (s *TransferService) GetDetail(id int64) (*models.TransferOrder, error) {
	var order models.TransferOrder
	if err := s.db.Preload("FromWarehouse").Preload("ToWarehouse").Preload("Items.SKU").
		First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.ErrOrderNotFound, Message: "调拨单不存在"}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}
	return &order, nil
}

// _ = fmt.Sprintf used for batch number generation
var _ = fmt.Sprintf
var _ = decimal.Zero
