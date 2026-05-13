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

// int64Ptr 安全地返回 int64 指针，0 值返回 nil
func int64Ptr(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}

// CostDetail 成本明细
type CostDetail struct {
	BatchID int64 `json:"batch_id" example:1`
	UnitCost decimal.Decimal `json:"unit_cost" example:5000.00`
	Quantity int `json:"quantity" example:10`
	TotalCost decimal.Decimal `json:"total_cost" example:50000.00`
}

// WarehouseStockDetail 仓库库存详情
type WarehouseStockDetail struct {
	models.WarehouseStock
	SKUName string `json:"sku_name,omitempty" example:"真皮沙发-棕色-三座"`
	SKUCode string `json:"sku_code,omitempty" example:"SKU001"`
	Barcode string `json:"barcode,omitempty" example:"6901234567890"`
	GroupName string `json:"group_name,omitempty" example:"沙发"`
}

// InventoryService 库存服务
type InventoryService struct {
	db       *gorm.DB
	invRepo  *repository.InventoryRepository
	skuRepo  *repository.SKURepository
	whRepo   *repository.WarehouseRepository
}

// NewInventoryService 创建库存服务实例
func NewInventoryService(db *gorm.DB, invRepo *repository.InventoryRepository, skuRepo *repository.SKURepository, whRepo *repository.WarehouseRepository) *InventoryService {
	return &InventoryService{
		db:      db,
		invRepo: invRepo,
		skuRepo: skuRepo,
		whRepo:  whRepo,
	}
}

// LockStock 锁定库存（订单创建时调用）
// 使用乐观锁：UPDATE warehouse_stocks SET locked_quantity = locked_quantity + ?, version = version + 1
// WHERE warehouse_id = ? AND sku_id = ? AND version = ? AND available_quantity >= ?
func (s *InventoryService) LockStock(warehouseID, skuID int64, quantity int) error {
	stock, err := s.invRepo.FindStockByWarehouseAndSKU(warehouseID, skuID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrInsufficientStock, Message: "库存记录不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if stock.AvailableQuantity < quantity {
		return &AppError{Code: apperrors.ErrInsufficientStock, Message: "可用库存不足"}
	}

	rowsAffected, err := s.invRepo.UpdateStockWithLock(warehouseID, skuID, stock.Version, map[string]interface{}{
		"locked_quantity":    gorm.Expr("locked_quantity + ?", quantity),
		"available_quantity": gorm.Expr("available_quantity - ?", quantity),
		"version":            gorm.Expr("version + 1"),
	})
	if err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "锁定库存失败"}
	}
	if rowsAffected == 0 {
		return &AppError{Code: apperrors.ErrInsufficientStock, Message: "库存已被修改，请重试"}
	}

	return nil
}

// UnlockStock 释放锁定库存（订单驳回时调用）
func (s *InventoryService) UnlockStock(warehouseID, skuID int64, quantity int) error {
	stock, err := s.invRepo.FindStockByWarehouseAndSKU(warehouseID, skuID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrInsufficientStock, Message: "库存记录不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if stock.LockedQuantity < quantity {
		return &AppError{Code: apperrors.ErrInsufficientStock, Message: "锁定库存不足"}
	}

	rowsAffected, err := s.invRepo.UpdateStockWithLock(warehouseID, skuID, stock.Version, map[string]interface{}{
		"locked_quantity":    gorm.Expr("locked_quantity - ?", quantity),
		"available_quantity": gorm.Expr("available_quantity + ?", quantity),
		"version":            gorm.Expr("version + 1"),
	})
	if err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "释放库存失败"}
	}
	if rowsAffected == 0 {
		return &AppError{Code: apperrors.ErrInsufficientStock, Message: "库存已被修改，请重试"}
	}

	return nil
}

// DeductStock 实际扣减库存（订单审核通过时调用）
// 核心FIFO逻辑
func (s *InventoryService) DeductStock(warehouseID, skuID int64, quantity int, storeID int64, orderID int64, createdBy int64, transactionType int8) ([]CostDetail, error) {
	// 1. 查找可用批次，按entry_date ASC（FIFO）
	batches, err := s.invRepo.FindBatchesBySKU(skuID, warehouseID)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询批次失败"}
	}

	// 检查总可用库存
	totalRemaining := 0
	for _, batch := range batches {
		totalRemaining += batch.RemainingQuantity
	}
	if totalRemaining < quantity {
		return nil, &AppError{Code: apperrors.ErrInsufficientStock, Message: fmt.Sprintf("库存不足，可用库存: %d，需要: %d", totalRemaining, quantity)}
	}

	// 2. 获取当前库存
	stock, err := s.invRepo.FindStockByWarehouseAndSKU(warehouseID, skuID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.ErrInsufficientStock, Message: "库存记录不存在"}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if stock.LockedQuantity < quantity {
		return nil, &AppError{Code: apperrors.ErrInsufficientStock, Message: "锁定库存不足"}
	}

	// 3. FIFO扣减
	var costDetails []CostDetail
	remaining := quantity
	now := time.Now()

	for i := range batches {
		if remaining <= 0 {
			break
		}

		batch := batches[i]
		deductQty := batch.RemainingQuantity
		if deductQty > remaining {
			deductQty = remaining
		}

		beforeStock := batch.RemainingQuantity
		afterStock := beforeStock - deductQty

		// 更新批次
		batch.RemainingQuantity = afterStock
		if afterStock == 0 {
			batch.Status = 0 // 已耗尽
		}
		if err := s.invRepo.UpdateBatch(&batch); err != nil {
			return nil, &AppError{Code: apperrors.InternalError, Message: "更新批次失败"}
		}

		// 记录库存流水
		tx := &models.InventoryTransaction{
			StoreID:         storeID,
			WarehouseID:     &warehouseID,
			TransactionType: transactionType, // 动态传入
			BizType:         1, // 商品
			BizID:           &skuID,
			BatchID:         &batch.ID,
			RelatedOrderID:  &orderID,
			Quantity:        deductQty,
			BeforeStock:     beforeStock,
			AfterStock:      afterStock,
			UnitCost:        batch.PurchasePrice,
			TotalCost:       batch.PurchasePrice.Mul(decimal.NewFromInt(int64(deductQty))),
			CreatedBy:       int64Ptr(createdBy),
			CreatedAt:       now,
		}
		if err := s.invRepo.CreateTransaction(tx); err != nil {
			return nil, &AppError{Code: apperrors.InternalError, Message: "记录库存流水失败"}
		}

		costDetails = append(costDetails, CostDetail{
			BatchID:   batch.ID,
			UnitCost:  batch.PurchasePrice,
			Quantity:  deductQty,
			TotalCost: batch.PurchasePrice.Mul(decimal.NewFromInt(int64(deductQty))),
		})

		remaining -= deductQty
	}

	// 4. 使用乐观锁更新warehouse_stocks
	// 注意：available_quantity在LockStock时已减少，这里只需减少locked_quantity和stock_quantity
	rowsAffected, err := s.invRepo.UpdateStockWithLock(warehouseID, skuID, stock.Version, map[string]interface{}{
		"locked_quantity":    gorm.Expr("locked_quantity - ?", quantity),
		"stock_quantity":     gorm.Expr("stock_quantity - ?", quantity),
		"version":            gorm.Expr("version + 1"),
	})
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "更新库存失败"}
	}
	if rowsAffected == 0 {
		return nil, &AppError{Code: apperrors.ErrInsufficientStock, Message: "库存已被修改，请重试"}
	}

	return costDetails, nil
}

// AddStock 增加库存（采购入库时调用）
func (s *InventoryService) AddStock(warehouseID, skuID int64, quantity int, purchasePrice, totalCost decimal.Decimal, batchNo string, purchaseOrderID int64, storeID int64, createdBy int64, transactionType int8) error {
	now := time.Now()

	// 1. 创建 inventory_batch 记录
	batch := &models.InventoryBatch{
		SKUID:            skuID,
		BatchNo:          batchNo,
		PurchaseOrderID:  &purchaseOrderID,
		PurchasePrice:    purchasePrice,
		TotalCost:        totalCost,
		InitialQuantity:  quantity,
		RemainingQuantity: quantity,
		WarehouseID:      &warehouseID,
		Status:           1,
		EntryDate:        &now,
	}
	if err := s.invRepo.CreateBatch(batch); err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "创建库存批次失败"}
	}

	// 2. 更新或创建 warehouse_stocks
	stock, err := s.invRepo.FindStockByWarehouseAndSKU(warehouseID, skuID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新库存记录
			newStock := &models.WarehouseStock{
				WarehouseID:       warehouseID,
				SKUID:             skuID,
				StockQuantity:     quantity,
				AvailableQuantity: quantity,
				LockedQuantity:    0,
				WarningStock:      10,
				Version:           0,
			}
			if err := s.invRepo.CreateStock(newStock); err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "创建库存记录失败"}
			}
		} else {
			return &AppError{Code: apperrors.InternalError, Message: "查询库存失败"}
		}
	} else {
		// 更新已有库存记录
		_, err := s.invRepo.UpdateStockWithLock(warehouseID, skuID, stock.Version, map[string]interface{}{
			"stock_quantity":     gorm.Expr("stock_quantity + ?", quantity),
			"available_quantity": gorm.Expr("available_quantity + ?", quantity),
			"version":            gorm.Expr("version + 1"),
		})
		if err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "更新库存失败"}
		}
	}

	// 3. 记录 inventory_transaction
	beforeStock := 0
	if stock != nil {
		beforeStock = stock.StockQuantity
	}
	afterStock := beforeStock + quantity

	tx := &models.InventoryTransaction{
		StoreID:           storeID,
		WarehouseID:       &warehouseID,
		TransactionType:   transactionType, // 动态传入
		BizType:           1, // 商品
		BizID:             &skuID,
		BatchID:           &batch.ID,
		RelatedPurchaseID: &purchaseOrderID,
		Quantity:          quantity,
		BeforeStock:       beforeStock,
		AfterStock:        afterStock,
		UnitCost:          purchasePrice,
		TotalCost:         totalCost,
		CreatedBy:         int64Ptr(createdBy),
		CreatedAt:         now,
	}
	if err := s.invRepo.CreateTransaction(tx); err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "记录库存流水失败"}
	}

	return nil
}

// GetStock 获取库存
func (s *InventoryService) GetStock(warehouseID, skuID int64) (*models.WarehouseStock, error) {
	stock, err := s.invRepo.FindStockByWarehouseAndSKU(warehouseID, skuID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询库存失败"}
	}
	return stock, nil
}

// GetStockList 获取仓库库存列表
func (s *InventoryService) GetStockList(warehouseID int64, keyword string, page, pageSize int) ([]models.WarehouseStock, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	stocks, total, err := s.invRepo.GetStockList(warehouseID, keyword, page, pageSize)
	if err != nil {
		return nil, 0, &AppError{Code: apperrors.InternalError, Message: "查询库存列表失败"}
	}
	return stocks, total, nil
}

// TransferStock 调拨库存
func (s *InventoryService) TransferStock(fromWarehouseID, toWarehouseID, skuID int64, quantity int, storeID int64, createdBy int64) error {
	// 1. 扣减调出仓库库存（使用FIFO）
	costDetails, err := s.DeductStock(fromWarehouseID, skuID, quantity, storeID, 0, createdBy, 3) // 3=调拨出库
	if err != nil {
		return err
	}

	// 计算总成本
	totalCost := decimal.Zero
	for _, detail := range costDetails {
		totalCost = totalCost.Add(detail.TotalCost)
	}

	// 平均单价
	avgPrice := decimal.Zero
	if quantity > 0 {
		avgPrice = totalCost.Div(decimal.NewFromInt(int64(quantity)))
	}

	// 2. 增加调入仓库库存
	batchNo := fmt.Sprintf("TR%s%d", time.Now().Format("20060102150405"), skuID)
	if err := s.AddStock(toWarehouseID, skuID, quantity, avgPrice, totalCost, batchNo, 0, storeID, createdBy, 4); err != nil { // 4=调拨入库
		return err
	}

	return nil
}

// AddGiftStock 增加礼品库存
func (s *InventoryService) AddGiftStock(warehouseID, giftID int64, quantity int, purchasePrice decimal.Decimal, batchNo string, storeID int64, createdBy int64) error {
	now := time.Now()

	// 创建礼品批次
	batch := &models.GiftInventoryBatch{
		GiftID:            giftID,
		BatchNo:          batchNo,
		PurchasePrice:    purchasePrice,
		InitialQuantity:  quantity,
		RemainingQuantity: quantity,
		WarehouseID:      &warehouseID,
		Status:           1,
		EntryDate:        &now,
	}
	if err := s.invRepo.CreateGiftBatch(batch); err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "创建礼品批次失败"}
	}

	// 更新或创建仓库礼品库存
	stock, err := s.invRepo.FindGiftStockByWarehouseAndGift(warehouseID, giftID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newStock := &models.WarehouseGiftStock{
				WarehouseID:       warehouseID,
				GiftID:            giftID,
				StockQuantity:     quantity,
				AvailableQuantity: quantity,
				LockedQuantity:    0,
				WarningStock:      10,
				Version:           0,
			}
			if err := s.invRepo.CreateGiftStock(newStock); err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "创建礼品库存记录失败"}
			}
		} else {
			return &AppError{Code: apperrors.InternalError, Message: "查询礼品库存失败"}
		}
	} else {
		_, err := s.invRepo.UpdateGiftStockWithLock(warehouseID, giftID, stock.Version, map[string]interface{}{
			"stock_quantity":     gorm.Expr("stock_quantity + ?", quantity),
			"available_quantity": gorm.Expr("available_quantity + ?", quantity),
			"version":            gorm.Expr("version + 1"),
		})
		if err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "更新礼品库存失败"}
		}
	}

	// 记录库存流水
	beforeStock := 0
	if stock != nil {
		beforeStock = stock.StockQuantity
	}
	afterStock := beforeStock + quantity

	tx := &models.InventoryTransaction{
		StoreID:         storeID,
		WarehouseID:     &warehouseID,
		TransactionType: 8, // 礼品入库
		BizType:         2, // 礼品
		BizID:           &giftID,
		Quantity:        quantity,
		BeforeStock:     beforeStock,
		AfterStock:      afterStock,
		UnitCost:        purchasePrice,
		TotalCost:       purchasePrice.Mul(decimal.NewFromInt(int64(quantity))),
		CreatedBy:       int64Ptr(createdBy),
		CreatedAt:       now,
	}
	if err := s.invRepo.CreateTransaction(tx); err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "记录库存流水失败"}
	}

	return nil
}

// DeductGiftStock 扣减礼品库存
func (s *InventoryService) DeductGiftStock(warehouseID, giftID int64, quantity int, storeID int64, orderID int64, createdBy int64) error {
	// 获取库存
	stock, err := s.invRepo.FindGiftStockByWarehouseAndGift(warehouseID, giftID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrInsufficientStock, Message: "礼品库存记录不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if stock.AvailableQuantity < quantity {
		return &AppError{Code: apperrors.ErrInsufficientStock, Message: "礼品可用库存不足"}
	}

	// 获取可用批次
	batches, err := s.invRepo.FindGiftBatchesByGift(giftID, warehouseID)
	if err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "查询礼品批次失败"}
	}

	remaining := quantity
	now := time.Now()
	var totalCost decimal.Decimal

	for i := range batches {
		if remaining <= 0 {
			break
		}

		batch := batches[i]
		deductQty := batch.RemainingQuantity
		if deductQty > remaining {
			deductQty = remaining
		}

		beforeStock := batch.RemainingQuantity
		afterStock := beforeStock - deductQty

		batch.RemainingQuantity = afterStock
		if afterStock == 0 {
			batch.Status = 0
		}
		if err := s.invRepo.UpdateGiftBatch(&batch); err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "更新礼品批次失败"}
		}

		cost := batch.PurchasePrice.Mul(decimal.NewFromInt(int64(deductQty)))
		totalCost = totalCost.Add(cost)

		remaining -= deductQty
	}

	// 更新仓库礼品库存
	rowsAffected, err := s.invRepo.UpdateGiftStockWithLock(warehouseID, giftID, stock.Version, map[string]interface{}{
		"available_quantity": gorm.Expr("available_quantity - ?", quantity),
		"stock_quantity":     gorm.Expr("stock_quantity - ?", quantity),
		"version":            gorm.Expr("version + 1"),
	})
	if err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "更新礼品库存失败"}
	}
	if rowsAffected == 0 {
		return &AppError{Code: apperrors.ErrInsufficientStock, Message: "礼品库存已被修改，请重试"}
	}

	// 记录流水
	tx := &models.InventoryTransaction{
		StoreID:         storeID,
		WarehouseID:     &warehouseID,
		TransactionType: 7, // 礼品出库
		BizType:         2, // 礼品
		BizID:           &giftID,
		RelatedOrderID:  &orderID,
		Quantity:        quantity,
		BeforeStock:     stock.AvailableQuantity,
		AfterStock:      stock.AvailableQuantity - quantity,
		TotalCost:       totalCost,
		CreatedBy:       int64Ptr(createdBy),
		CreatedAt:       now,
	}
	if err := s.invRepo.CreateTransaction(tx); err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "记录库存流水失败"}
	}

	return nil
}

// GetTransactionList 获取库存流水列表
func (s *InventoryService) GetTransactionList(warehouseID int64, transactionType int8, startDate, endDate string, page, pageSize int) ([]models.InventoryTransaction, int64, error) {
	transactions, total, err := s.invRepo.GetTransactionList(warehouseID, transactionType, startDate, endDate, page, pageSize)
	if err != nil {
		return nil, 0, &AppError{Code: apperrors.InternalError, Message: "查询库存流水失败"}
	}
	return transactions, total, nil
}
