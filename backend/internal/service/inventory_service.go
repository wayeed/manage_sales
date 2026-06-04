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

// DeductLockedStock 将锁定库存转为实际出库（送货出库时调用）
// 与DeductStock的区别：available_quantity已经在LockStock时减少，这里只需减少locked_quantity和stock_quantity
func (s *InventoryService) DeductLockedStock(warehouseID, skuID int64, quantity int, storeID int64, orderID int64, createdBy int64) ([]CostDetail, error) {
	// 新流程：直接使用审核时已锁定的批次
	// 查询订单商品中已绑定的批次
	var orderItems []models.OrderItem
	if err := s.db.Where("order_id = ? AND sku_id = ? AND batch_id IS NOT NULL AND batch_id > 0", orderID, skuID).Find(&orderItems).Error; err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询订单商品失败"}
	}

	// 获取当前库存
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

	var costDetails []CostDetail
	now := time.Now()

	// 使用已绑定的批次扣减
	for _, item := range orderItems {
		if item.BatchID == nil || *item.BatchID == 0 {
			continue
		}

		var batch models.InventoryBatch
		if err := s.db.First(&batch, *item.BatchID).Error; err != nil {
			continue
		}

		deductQty := item.Quantity

		// 扣减批次剩余数量
		batch.RemainingQuantity -= deductQty
		if batch.RemainingQuantity <= 0 {
			batch.RemainingQuantity = 0
			batch.Status = 0
		}
		if err := s.invRepo.UpdateBatch(&batch); err != nil {
			return nil, &AppError{Code: apperrors.InternalError, Message: "更新批次失败"}
		}

		// 记录库存流水
		tx := &models.InventoryTransaction{
			StoreID:         storeID,
			WarehouseID:     &warehouseID,
			TransactionType: models.TransactionTypeLockToOut, // 11: 销售锁定转出库
			BizType:         1,
			BizID:           &skuID,
			BatchID:         item.BatchID,
			RelatedOrderID:  &orderID,
			Quantity:        deductQty,
			BeforeStock:     batch.RemainingQuantity + deductQty,
			AfterStock:      batch.RemainingQuantity,
			UnitCost:        batch.PurchasePrice,
			TotalCost:       batch.PurchasePrice.Mul(decimal.NewFromInt(int64(deductQty))),
			CreatedBy:       int64Ptr(createdBy),
			CreatedAt:       now,
		}
		if err := s.invRepo.CreateTransaction(tx); err != nil {
			return nil, &AppError{Code: apperrors.InternalError, Message: "记录库存流水失败"}
		}

		costDetails = append(costDetails, CostDetail{
			BatchID:   *item.BatchID,
			UnitCost:  batch.PurchasePrice,
			Quantity:  deductQty,
			TotalCost: batch.PurchasePrice.Mul(decimal.NewFromInt(int64(deductQty))),
		})
	}

	// 更新warehouse_stocks：locked -= qty, stock -= qty
	rowsAffected, err := s.invRepo.UpdateStockWithLock(warehouseID, skuID, stock.Version, map[string]interface{}{
		"locked_quantity": gorm.Expr("locked_quantity - ?", quantity),
		"stock_quantity":  gorm.Expr("stock_quantity - ?", quantity),
		"version":         gorm.Expr("version + 1"),
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

// LockStockByBatch 按批次FIFO锁定库存（审核通过时调用）
// 优先从指定仓库锁定，不足时跨仓库锁定
// 返回锁定的批次明细列表
func (s *InventoryService) LockStockByBatch(warehouseID, skuID int64, quantity int, orderID int64, storeID int64) ([]CostDetail, error) {
	// 1. 查询指定仓库的可用库存总量
	stock, err := s.invRepo.FindStockByWarehouseAndSKU(warehouseID, skuID)
	if err == nil && stock.AvailableQuantity >= quantity {
		// 指定仓库库存充足，只在指定仓库锁定
		return s.lockStockInWarehouse(warehouseID, skuID, quantity, stock.Version)
	}

	// 2. 指定仓库不足或无记录，跨仓库查询所有可用库存
	crossWarehouseBatches, err := s.invRepo.FindBatchesBySKU(skuID, 0)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询批次失败"}
	}

	// 计算跨仓库可用总量
	var totalAvailable int
	for _, b := range crossWarehouseBatches {
		if b.RemainingQuantity > 0 && b.Status == 1 {
			totalAvailable += b.RemainingQuantity
		}
	}

	if totalAvailable < quantity {
		return nil, nil // 可用库存不足，返回空（表示缺货，进入排队）
	}

	// 3. 逐批次锁定（跨仓库FIFO）
	var costDetails []CostDetail
	remaining := quantity

	// 先锁定指定仓库的库存
	if stock != nil && stock.AvailableQuantity > 0 {
		whBatches, _ := s.invRepo.FindBatchesBySKU(skuID, warehouseID)
		for i := range whBatches {
			if remaining <= 0 {
				break
			}
			batch := &whBatches[i]
			if batch.RemainingQuantity <= 0 || batch.Status == 0 {
				continue
			}

			lockQty := batch.RemainingQuantity
			if lockQty > remaining {
				lockQty = remaining
			}

			batch.RemainingQuantity -= lockQty
			if batch.RemainingQuantity == 0 {
				batch.Status = 0
			}
			if err := s.invRepo.UpdateBatch(batch); err != nil {
				return nil, &AppError{Code: apperrors.InternalError, Message: "更新批次失败"}
			}

			unitCost := batch.PurchasePrice
			totalCost := unitCost.Mul(decimal.NewFromInt(int64(lockQty)))
			costDetails = append(costDetails, CostDetail{
				BatchID:   batch.ID,
				UnitCost:  unitCost,
				Quantity:  lockQty,
				TotalCost: totalCost,
			})
			remaining -= lockQty
		}

		// 更新指定仓库库存
		lockedInWH := quantity - remaining
		if lockedInWH > 0 {
			s.invRepo.UpdateStockWithLock(warehouseID, skuID, stock.Version, map[string]interface{}{
				"locked_quantity":    gorm.Expr("locked_quantity + ?", lockedInWH),
				"available_quantity": gorm.Expr("available_quantity - ?", lockedInWH),
				"version":            gorm.Expr("version + 1"),
			})
		}
	}

	// 再锁定其他仓库的库存
	for i := range crossWarehouseBatches {
		if remaining <= 0 {
			break
		}
		batch := &crossWarehouseBatches[i]
		if batch.RemainingQuantity <= 0 || batch.Status == 0 {
			continue
		}
		// 跳过已在指定仓库锁定的批次
		if batch.WarehouseID != nil && *batch.WarehouseID == warehouseID {
			continue
		}

		lockQty := batch.RemainingQuantity
		if lockQty > remaining {
			lockQty = remaining
		}

		batch.RemainingQuantity -= lockQty
		if batch.RemainingQuantity == 0 {
			batch.Status = 0
		}
		if err := s.invRepo.UpdateBatch(batch); err != nil {
			return nil, &AppError{Code: apperrors.InternalError, Message: "更新批次失败"}
		}

		unitCost := batch.PurchasePrice
		totalCost := unitCost.Mul(decimal.NewFromInt(int64(lockQty)))
		costDetails = append(costDetails, CostDetail{
			BatchID:   batch.ID,
			UnitCost:  unitCost,
			Quantity:  lockQty,
			TotalCost: totalCost,
		})
		remaining -= lockQty

		// 更新对应仓库的库存
		if batch.WarehouseID != nil {
			var whStock models.WarehouseStock
			if _, err := s.invRepo.FindStockByWarehouseAndSKU(*batch.WarehouseID, skuID); err == nil {
				// 重新查询获取最新版本
				s.db.Where("warehouse_id = ? AND sku_id = ?", *batch.WarehouseID, skuID).First(&whStock)
				s.invRepo.UpdateStockWithLock(*batch.WarehouseID, skuID, whStock.Version, map[string]interface{}{
					"locked_quantity":    gorm.Expr("locked_quantity + ?", lockQty),
					"available_quantity": gorm.Expr("available_quantity - ?", lockQty),
					"version":            gorm.Expr("version + 1"),
				})
			}
		}
	}

	if remaining > 0 {
		return nil, &AppError{Code: apperrors.InternalError, Message: "库存锁定异常：批次库存不足"}
	}

	return costDetails, nil
}

// lockStockInWarehouse 在指定仓库内按批次FIFO锁定库存
func (s *InventoryService) lockStockInWarehouse(warehouseID, skuID int64, quantity int, version int) ([]CostDetail, error) {
	batches, err := s.invRepo.FindBatchesBySKU(skuID, warehouseID)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询批次失败"}
	}

	var costDetails []CostDetail
	remaining := quantity

	for i := range batches {
		if remaining <= 0 {
			break
		}
		batch := &batches[i]
		if batch.RemainingQuantity <= 0 || batch.Status == 0 {
			continue
		}

		lockQty := batch.RemainingQuantity
		if lockQty > remaining {
			lockQty = remaining
		}

		batch.RemainingQuantity -= lockQty
		if batch.RemainingQuantity == 0 {
			batch.Status = 0
		}
		if err := s.invRepo.UpdateBatch(batch); err != nil {
			return nil, &AppError{Code: apperrors.InternalError, Message: "更新批次失败"}
		}

		unitCost := batch.PurchasePrice
		totalCost := unitCost.Mul(decimal.NewFromInt(int64(lockQty)))
		costDetails = append(costDetails, CostDetail{
			BatchID:   batch.ID,
			UnitCost:  unitCost,
			Quantity:  lockQty,
			TotalCost: totalCost,
		})
		remaining -= lockQty
	}

	if remaining > 0 {
		return nil, &AppError{Code: apperrors.InternalError, Message: "库存锁定异常：批次库存不足"}
	}

	// 更新仓库库存
	rowsAffected, err := s.invRepo.UpdateStockWithLock(warehouseID, skuID, version, map[string]interface{}{
		"locked_quantity":    gorm.Expr("locked_quantity + ?", quantity),
		"available_quantity": gorm.Expr("available_quantity - ?", quantity),
		"version":            gorm.Expr("version + 1"),
	})
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "锁定库存失败"}
	}
	if rowsAffected == 0 {
		return nil, &AppError{Code: apperrors.ErrInsufficientStock, Message: "库存已被修改，请重试"}
	}

	return costDetails, nil
}

// GetAvailableStock 获取SKU可用库存
func (s *InventoryService) GetAvailableStock(warehouseID, skuID int64) (int, error) {
	stock, err := s.invRepo.FindStockByWarehouseAndSKU(warehouseID, skuID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return stock.AvailableQuantity, nil
}

// AllocateStockToQueue 采购入库后自动分配库存到缺货排队订单
// 按订单先后顺序（FIFO）依次分配
func (s *InventoryService) AllocateStockToQueue(db *gorm.DB, warehouseID, skuID int64, batchID int64, batchPurchasePrice decimal.Decimal, availableQty int, storeID int64) error {
	if availableQty <= 0 {
		return nil
	}

	// 查询该SKU的排队记录（按创建时间升序）
	var queues []models.StockQueue
	if err := db.Where("sku_id = ? AND status IN (0, 1) AND quantity > allocated_qty", skuID).
		Order("created_at ASC").
		Find(&queues).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "查询排队记录失败"}
	}

	remaining := availableQty
	for i := range queues {
		if remaining <= 0 {
			break
		}
		q := &queues[i]
		needQty := q.Quantity - q.AllocatedQty // 还需要分配的数量

		allocQty := needQty
		if allocQty > remaining {
			allocQty = remaining
		}

		// 更新排队记录
		q.AllocatedQty += allocQty
		if q.AllocatedQty >= q.Quantity {
			q.Status = 2 // 全部分配
		} else {
			q.Status = 1 // 部分分配
		}
		if err := db.Save(q).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "更新排队记录失败"}
		}

		// 更新order_items：写入batch_id和成本
		unitCost := batchPurchasePrice
		totalCost := unitCost.Mul(decimal.NewFromInt(int64(allocQty)))
		if err := db.Model(&models.OrderItem{}).Where("id = ?", q.OrderItemID).Updates(map[string]interface{}{
			"batch_id":   batchID,
			"unit_cost":  unitCost,
			"total_cost": totalCost,
		}).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "更新订单商品批次失败"}
		}

		// 更新仓库库存：available -= qty, locked += qty
		var stock models.WarehouseStock
		if err := db.Where("warehouse_id = ? AND sku_id = ?", warehouseID, skuID).First(&stock).Error; err == nil {
			db.Model(&stock).Updates(map[string]interface{}{
				"locked_quantity":    gorm.Expr("locked_quantity + ?", allocQty),
				"available_quantity": gorm.Expr("available_quantity - ?", allocQty),
				"version":            gorm.Expr("version + 1"),
			})
		}

		// 记录库存流水
		invTx := &models.InventoryTransaction{
			StoreID:         storeID,
			WarehouseID:     &warehouseID,
			TransactionType: 9, // 库存锁定
			BizType:         1,
			BizID:           &skuID,
			RelatedOrderID:  &q.OrderID,
			Quantity:        allocQty,
			Remark:          "采购入库自动分配缺货订单",
			CreatedAt:       time.Now(),
		}
		db.Create(invTx)

		remaining -= allocQty

		// 检查该订单是否所有商品都已分配库存，更新订单stock_status
		if q.AllocatedQty >= q.Quantity {
			// 查询该订单还有多少个未分配的排队记录
			var pendingCount int64
			db.Model(&models.StockQueue{}).Where("order_id = ? AND status IN (0, 1)", q.OrderID).Count(&pendingCount)
			if pendingCount == 0 {
				// 全部缺货商品都已分配，更新订单为全部有库存
				db.Model(&models.Order{}).Where("id = ?", q.OrderID).Update("stock_status", 0)
			} else {
				// 还有部分缺货，更新为部分缺货状态
				db.Model(&models.Order{}).Where("id = ?", q.OrderID).Update("stock_status", 1)
			}
		}
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

// GetStockByID 根据ID查询库存详情
func (s *InventoryService) GetStockByID(id int64) (*models.WarehouseStock, error) {
	var stock models.WarehouseStock
	if err := s.db.Preload("Warehouse").Preload("SKU.Product").First(&stock, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.NotFound, Message: "库存记录不存在"}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询库存失败"}
	}
	return &stock, nil
}

// GetStockList 获取仓库库存列表
func (s *InventoryService) GetStockList(warehouseID int64, skuID int64, keyword string, page, pageSize int) ([]models.WarehouseStock, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	stocks, total, err := s.invRepo.GetStockList(warehouseID, skuID, keyword, page, pageSize)
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

// CheckStockForOrder 检查订单商品库存是否充足
// 返回库存不足的商品列表，如果全部充足则返回空列表
func (s *InventoryService) CheckStockForOrder(warehouseID int64, items []models.OrderItem) ([]StockCheckResult, error) {
	results := make([]StockCheckResult, 0)

	for _, item := range items {
		stock, err := s.GetStock(warehouseID, item.SKUID)
		if err != nil {
			return nil, err
		}

                availableQty := 0
                lockedQty := 0
                if stock != nil {
                        availableQty = stock.AvailableQuantity
                        lockedQty = stock.LockedQuantity
                }

                // 出库时：可用库存 + 锁定库存（审核时已锁定，锁定库存也是为该订单准备的）
                totalAvailable := availableQty + lockedQty
                if totalAvailable < item.Quantity {
                        results = append(results, StockCheckResult{
                                SKUID:           item.SKUID,
                                ProductName:     item.ProductName,
                                SKUName:         item.SKUName,
                                RequiredQty:     item.Quantity,
                                AvailableQty:    totalAvailable,
                                IsSufficient:    false,
                        })

                }
        }

	return results, nil
}

// StockCheckResult 库存检查结果
type StockCheckResult struct {
	SKUID        int64  `json:"sku_id"`
	ProductName  string `json:"product_name"`
	SKUName      string `json:"sku_name"`
	RequiredQty  int    `json:"required_qty"`     // 订单需求数量
	AvailableQty int    `json:"available_qty"`    // 可用库存数量
	IsSufficient bool   `json:"is_sufficient"`    // 是否充足
}

// IncreaseStock 增加库存（退货入库）
func (s *InventoryService) IncreaseStock(skuID int64, quantity int, storeID int64, warehouseID int64, orderID int64, operatorID int64, remark string) error {
	// 使用传入的仓库ID

	// 尝试查找现有库存记录
	var stock models.WarehouseStock
	err := s.db.Where("sku_id = ? AND warehouse_id = ?", skuID, warehouseID).First(&stock).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return &AppError{Code: apperrors.InternalError, Message: "查询库存失败"}
	}

	beforeQty := 0
	afterQty := 0

	// 更新库存数量
	if err == gorm.ErrRecordNotFound {
		// 创建新库存记录
		stock = models.WarehouseStock{
			WarehouseID:       warehouseID,
			SKUID:             skuID,
			StockQuantity:     quantity,
			AvailableQuantity: quantity,
			LockedQuantity:    0,
			WarningStock:      10,
			Version:           0,
		}
		if err := s.db.Create(&stock).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "创建库存记录失败"}
		}
		beforeQty = 0
		afterQty = quantity
	} else {
		// 更新已有库存记录
		beforeQty = stock.StockQuantity
		afterQty = beforeQty + quantity

		rowsAffected, err := s.invRepo.UpdateStockWithLock(warehouseID, skuID, stock.Version, map[string]interface{}{
			"stock_quantity":     gorm.Expr("stock_quantity + ?", quantity),
			"available_quantity": gorm.Expr("available_quantity + ?", quantity),
			"version":            gorm.Expr("version + 1"),
		})
		if err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "更新库存失败"}
		}
		if rowsAffected == 0 {
			return &AppError{Code: apperrors.ErrInsufficientStock, Message: "库存已被修改，请重试"}
		}
	}

	// 创建库存流水记录（使用退货入库类型）
	tx := &models.InventoryTransaction{
		StoreID:         storeID,
		WarehouseID:     &warehouseID,
		TransactionType: 12, // 退货入库
		BizType:         1, // 商品
		BizID:           &skuID,
		RelatedOrderID:   &orderID,
		Quantity:        quantity,
		BeforeStock:      beforeQty,
		AfterStock:       afterQty,
		Remark:          remark,
		CreatedBy:        int64Ptr(operatorID),
		CreatedAt:        time.Now(),
	}
	if err := s.invRepo.CreateTransaction(tx); err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "记录库存流水失败"}
	}

	return nil
}
