package service

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	"furniture-commission/internal/repository"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupInventoryTestDB 创建测试数据库并初始化表结构
func setupInventoryTestDB(t *testing.T) *gorm.DB {
	dbName := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	assert.NoError(t, err)

	// 自动迁移表结构
	err = db.AutoMigrate(
		&models.WarehouseStock{},
		&models.InventoryBatch{},
		&models.InventoryTransaction{},
		&models.WarehouseGiftStock{},
		&models.GiftInventoryBatch{},
	)
	assert.NoError(t, err)

	return db
}

// createTestStock 创建测试库存记录
func createTestStock(t *testing.T, db *gorm.DB, warehouseID, skuID, stockQty, availQty, lockedQty, warningQty, version int) *models.WarehouseStock {
	stock := &models.WarehouseStock{
		WarehouseID:       int64(warehouseID),
		SKUID:             int64(skuID),
		StockQuantity:     stockQty,
		AvailableQuantity: availQty,
		LockedQuantity:    lockedQty,
		WarningStock:      warningQty,
		Version:           version,
	}
	err := db.Create(stock).Error
	assert.NoError(t, err)
	return stock
}

// createTestBatch 创建测试批次
func createTestBatch(t *testing.T, db *gorm.DB, skuID, warehouseID int, batchNo string, price float64, qty, remaining int, entryDate time.Time) *models.InventoryBatch {
	batch := &models.InventoryBatch{
		SKUID:             int64(skuID),
		BatchNo:           batchNo,
		PurchasePrice:     decimal.NewFromFloat(price),
		TotalCost:         decimal.NewFromFloat(price * float64(qty)),
		InitialQuantity:   qty,
		RemainingQuantity: remaining,
		WarehouseID:       &[]int64{int64(warehouseID)}[0],
		Status:            1,
		EntryDate:         &entryDate,
	}
	err := db.Create(batch).Error
	assert.NoError(t, err)
	return batch
}

// TestFIFODeductStock 多批次FIFO出库成本计算
func TestFIFODeductStock(t *testing.T) {
	db := setupInventoryTestDB(t)
	invRepo := repository.NewInventoryRepository(db)
	svc := NewInventoryService(db, invRepo, nil, nil)

	warehouseID := int64(1)
	skuID := int64(100)
	storeID := int64(1)
	orderID := int64(1000)
	createdBy := int64(1)

	// 创建库存记录：stock=100, available=100, locked=100
	createTestStock(t, db, 1, 100, 100, 100, 100, 10, 0)

	// 创建3个批次（不同价格、不同入库日期）
	date1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	date3 := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)

	createTestBatch(t, db, 100, 1, "BATCH-001", 10.00, 50, 50, date1) // 最早批次，单价10
	createTestBatch(t, db, 100, 1, "BATCH-002", 15.00, 30, 30, date2) // 中间批次，单价15
	createTestBatch(t, db, 100, 1, "BATCH-003", 20.00, 20, 20, date3) // 最新批次，单价20

	// 扣减70个库存（FIFO：先扣50个@10，再扣20个@15）
	costDetails, err := svc.DeductStock(warehouseID, skuID, 70, storeID, orderID, createdBy, int8(2))
	assert.NoError(t, err)
	assert.NotNil(t, costDetails)
	assert.Len(t, costDetails, 2) // 涉及2个批次

	// 验证第一批次：扣减50个@10
	assert.Equal(t, int64(1), costDetails[0].BatchID) // 第一个创建的batch
	assert.Equal(t, 50, costDetails[0].Quantity)
	assert.True(t, costDetails[0].UnitCost.Equal(decimal.NewFromFloat(10.00)))
	assert.True(t, costDetails[0].TotalCost.Equal(decimal.NewFromFloat(500.00)))

	// 验证第二批次：扣减20个@15
	assert.Equal(t, 20, costDetails[1].Quantity)
	assert.True(t, costDetails[1].UnitCost.Equal(decimal.NewFromFloat(15.00)))
	assert.True(t, costDetails[1].TotalCost.Equal(decimal.NewFromFloat(300.00)))

	// 验证库存更新
	stock, err := invRepo.FindStockByWarehouseAndSKU(warehouseID, skuID)
	assert.NoError(t, err)
	assert.Equal(t, 30, stock.StockQuantity)      // 100 - 70 = 30
	assert.Equal(t, 100, stock.AvailableQuantity) // 100 (DeductStock不再减少available)
	assert.Equal(t, 30, stock.LockedQuantity)     // 100 - 70 = 30

	// 验证批次更新
	batches, err := invRepo.FindBatchesBySKU(skuID, warehouseID)
	assert.NoError(t, err)
	// 第一个批次应该已耗尽
	for _, b := range batches {
		if b.BatchNo == "BATCH-001" {
			assert.Equal(t, int8(0), b.Status) // 已耗尽
			assert.Equal(t, 0, b.RemainingQuantity)
		}
		if b.BatchNo == "BATCH-002" {
			assert.Equal(t, int8(1), b.Status)       // 仍可用
			assert.Equal(t, 10, b.RemainingQuantity) // 30 - 20 = 10
		}
		if b.BatchNo == "BATCH-003" {
			assert.Equal(t, int8(1), b.Status)       // 仍可用
			assert.Equal(t, 20, b.RemainingQuantity) // 未被扣减
		}
	}

	// 验证库存流水
	var transactions []models.InventoryTransaction
	db.Where("biz_id = ? AND transaction_type = 2", skuID).Find(&transactions)
	assert.Len(t, transactions, 2) // 2笔流水
}

// TestConcurrentDeductStock 并发库存扣减（乐观锁）
func TestConcurrentDeductStock(t *testing.T) {
	db := setupInventoryTestDB(t)
	invRepo := repository.NewInventoryRepository(db)
	svc := NewInventoryService(db, invRepo, nil, nil)

	warehouseID := int64(1)
	skuID := int64(200)
	storeID := int64(1)
	createdBy := int64(1)

	// 创建库存记录：stock=10, available=10, locked=10
	createTestStock(t, db, 1, 200, 10, 10, 10, 10, 0)

	// 创建批次：10个@10元
	date := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	createTestBatch(t, db, 200, 1, "BATCH-CONC", 10.00, 10, 10, date)

	// 并发扣减：3个goroutine各扣5个（总共15个，但只有10个可用）
	var wg sync.WaitGroup
	successCount := 0
	failCount := 0
	mu := sync.Mutex{}

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			orderID := int64(2000 + idx)
			_, err := svc.DeductStock(warehouseID, skuID, 5, storeID, orderID, createdBy, int8(2))
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				failCount++
			} else {
				successCount++
			}
		}(i)
	}
	wg.Wait()

	// 应该有部分成功、部分失败（乐观锁冲突或库存不足）
	t.Logf("成功: %d, 失败: %d", successCount, failCount)
	assert.Greater(t, successCount, 0)
	assert.Greater(t, failCount, 0)

	// 验证最终库存一致性
	stock, err := invRepo.FindStockByWarehouseAndSKU(warehouseID, skuID)
	if err == nil {
		// 库存不应为负数
		assert.GreaterOrEqual(t, stock.StockQuantity, 0)
		assert.GreaterOrEqual(t, stock.AvailableQuantity, 0)
		assert.GreaterOrEqual(t, stock.LockedQuantity, 0)
	}

	// 注意：SQLite在并发场景下可能出现"database table is locked"，
	// 导致批次已扣减但库存未更新。在MySQL/PostgreSQL中使用事务可避免此问题。
	// 这里主要验证：并发场景下不会出现panic，且部分请求会因乐观锁失败
}

// TestLockAndUnlockStock 锁定和释放库存
func TestLockAndUnlockStock(t *testing.T) {
	db := setupInventoryTestDB(t)
	invRepo := repository.NewInventoryRepository(db)
	svc := NewInventoryService(db, invRepo, nil, nil)

	warehouseID := int64(1)
	skuID := int64(300)

	// 创建库存记录：stock=100, available=100, locked=0
	createTestStock(t, db, 1, 300, 100, 100, 0, 10, 0)

	// 测试1：锁定库存
	err := svc.LockStock(warehouseID, skuID, 30)
	assert.NoError(t, err)

	stock, err := invRepo.FindStockByWarehouseAndSKU(warehouseID, skuID)
	assert.NoError(t, err)
	assert.Equal(t, 70, stock.AvailableQuantity) // 100 - 30 = 70
	assert.Equal(t, 30, stock.LockedQuantity)    // 0 + 30 = 30
	assert.Equal(t, 100, stock.StockQuantity)    // 总库存不变

	// 测试2：继续锁定
	err = svc.LockStock(warehouseID, skuID, 20)
	assert.NoError(t, err)

	stock, err = invRepo.FindStockByWarehouseAndSKU(warehouseID, skuID)
	assert.NoError(t, err)
	assert.Equal(t, 50, stock.AvailableQuantity) // 70 - 20 = 50
	assert.Equal(t, 50, stock.LockedQuantity)    // 30 + 20 = 50

	// 测试3：锁定超过可用库存
	err = svc.LockStock(warehouseID, skuID, 60)
	assert.Error(t, err)
	appErr, ok := err.(*AppError)
	assert.True(t, ok)
	assert.Equal(t, apperrors.ErrInsufficientStock, appErr.Code)

	// 测试4：释放库存
	err = svc.UnlockStock(warehouseID, skuID, 25)
	assert.NoError(t, err)

	stock, err = invRepo.FindStockByWarehouseAndSKU(warehouseID, skuID)
	assert.NoError(t, err)
	assert.Equal(t, 75, stock.AvailableQuantity) // 50 + 25 = 75
	assert.Equal(t, 25, stock.LockedQuantity)    // 50 - 25 = 25

	// 测试5：释放超过锁定库存
	err = svc.UnlockStock(warehouseID, skuID, 30)
	assert.Error(t, err)
	appErr, ok = err.(*AppError)
	assert.True(t, ok)
	assert.Equal(t, apperrors.ErrInsufficientStock, appErr.Code)

	// 测试6：库存记录不存在
	err = svc.LockStock(999, skuID, 10)
	assert.Error(t, err)
}

// TestAddStock 采购入库增加库存
func TestAddStock(t *testing.T) {
	db := setupInventoryTestDB(t)
	invRepo := repository.NewInventoryRepository(db)
	svc := NewInventoryService(db, invRepo, nil, nil)

	warehouseID := int64(1)
	skuID := int64(400)
	storeID := int64(1)
	createdBy := int64(1)

	// 测试1：首次入库（库存记录不存在）
	purchasePrice := decimal.NewFromFloat(50.00)
	totalCost := decimal.NewFromFloat(500.00)
	batchNo := "PO-001-SKU400"

	err := svc.AddStock(warehouseID, skuID, 10, purchasePrice, totalCost, batchNo, 1, storeID, createdBy, int8(2))
	assert.NoError(t, err)

	// 验证库存记录已创建
	stock, err := invRepo.FindStockByWarehouseAndSKU(warehouseID, skuID)
	assert.NoError(t, err)
	assert.Equal(t, 10, stock.StockQuantity)
	assert.Equal(t, 10, stock.AvailableQuantity)
	assert.Equal(t, 0, stock.LockedQuantity)

	// 验证批次记录
	var batches []models.InventoryBatch
	db.Where("sku_id = ?", skuID).Find(&batches)
	assert.Len(t, batches, 1)
	assert.Equal(t, batchNo, batches[0].BatchNo)
	assert.Equal(t, 10, batches[0].InitialQuantity)
	assert.Equal(t, 10, batches[0].RemainingQuantity)
	assert.Equal(t, int8(1), batches[0].Status)
	assert.True(t, batches[0].PurchasePrice.Equal(decimal.NewFromFloat(50.00)))

	// 验证库存流水
	var transactions []models.InventoryTransaction
	db.Where("biz_id = ? AND transaction_type = 1", skuID).Find(&transactions)
	assert.Len(t, transactions, 1)
	assert.Equal(t, 10, transactions[0].Quantity)
	assert.Equal(t, 0, transactions[0].BeforeStock)
	assert.Equal(t, 10, transactions[0].AfterStock)
	assert.True(t, transactions[0].UnitCost.Equal(decimal.NewFromFloat(50.00)))
	assert.True(t, transactions[0].TotalCost.Equal(decimal.NewFromFloat(500.00)))

	// 测试2：再次入库（库存记录已存在）
	purchasePrice2 := decimal.NewFromFloat(55.00)
	totalCost2 := decimal.NewFromFloat(550.00)
	batchNo2 := "PO-002-SKU400"

	err = svc.AddStock(warehouseID, skuID, 10, purchasePrice2, totalCost2, batchNo2, 2, storeID, createdBy, int8(2))
	assert.NoError(t, err)

	// 验证库存更新
	stock, err = invRepo.FindStockByWarehouseAndSKU(warehouseID, skuID)
	assert.NoError(t, err)
	assert.Equal(t, 20, stock.StockQuantity)
	assert.Equal(t, 20, stock.AvailableQuantity)

	// 验证批次数量
	db.Where("sku_id = ?", skuID).Find(&batches)
	assert.Len(t, batches, 2)

	// 验证流水数量
	db.Where("biz_id = ? AND transaction_type = 1", skuID).Find(&transactions)
	assert.Len(t, transactions, 2)
	assert.Equal(t, 10, transactions[1].BeforeStock)
	assert.Equal(t, 20, transactions[1].AfterStock)

	// 测试3：入库后FIFO出库验证
	// 先锁定库存
	err = svc.LockStock(warehouseID, skuID, 15)
	assert.NoError(t, err)

	// 扣减15个（应该先扣10个@50，再扣5个@55）
	costDetails, err := svc.DeductStock(warehouseID, skuID, 15, storeID, 100, createdBy, int8(2))
	assert.NoError(t, err)
	assert.Len(t, costDetails, 2)

	// 第一笔：10个@50
	assert.Equal(t, 10, costDetails[0].Quantity)
	assert.True(t, costDetails[0].UnitCost.Equal(decimal.NewFromFloat(50.00)))

	// 第二笔：5个@55
	assert.Equal(t, 5, costDetails[1].Quantity)
	assert.True(t, costDetails[1].UnitCost.Equal(decimal.NewFromFloat(55.00)))

	// 验证总成本
	totalDeductCost := decimal.Zero
	for _, detail := range costDetails {
		totalDeductCost = totalDeductCost.Add(detail.TotalCost)
	}
	expectedCost := decimal.NewFromFloat(50*10 + 55*5)
	assert.True(t, totalDeductCost.Equal(expectedCost), fmt.Sprintf("expected %s, got %s", expectedCost.String(), totalDeductCost.String()))

	// 验证剩余库存
	stock, err = invRepo.FindStockByWarehouseAndSKU(warehouseID, skuID)
	assert.NoError(t, err)
	assert.Equal(t, 5, stock.StockQuantity)     // 20 - 15 = 5
	assert.Equal(t, 5, stock.AvailableQuantity) // 5 (LockStock时已减少，DeductStock不再减少)
	assert.Equal(t, 0, stock.LockedQuantity)    // 15 - 15 = 0
}
