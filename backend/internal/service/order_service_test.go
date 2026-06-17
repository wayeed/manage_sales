package service

import (
	"fmt"
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

// setupOrderTestDB 创建订单测试数据库并初始化表结构
func setupOrderTestDB(t *testing.T) *gorm.DB {
	// 每个测试使用独立的数据库文件避免SQLite锁冲突
	dbName := fmt.Sprintf("file:%s_%d?mode=memory&cache=shared", t.Name(), time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(
		&models.Order{},
		&models.OrderItem{},
		&models.OrderGift{},
		&models.Payment{},
		&models.Customer{},
		&models.CustomerFollowUp{},
		&models.Peer{},
		&models.WarehouseStock{},
		&models.InventoryBatch{},
		&models.InventoryTransaction{},
		&models.ProductSKU{},
		&models.Warehouse{},
	)
	assert.NoError(t, err)

	return db
}

// setupOrderTestService 创建订单测试服务
func setupOrderTestService(t *testing.T) (*OrderService, *gorm.DB) {
	db := setupOrderTestDB(t)
	orderRepo := repository.NewOrderRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	peerRepo := repository.NewPeerRepository(db)
	inventoryRepo := repository.NewInventoryRepository(db)
	inventorySvc := NewInventoryService(db, inventoryRepo, nil, nil)
	orderSvc := NewOrderService(db, orderRepo, paymentRepo, customerRepo, peerRepo, inventorySvc, nil, nil, nil)
	return orderSvc, db
}

// createTestWarehouse 创建测试仓库
func createTestWarehouse(t *testing.T, db *gorm.DB, id int64) {
	warehouse := &models.Warehouse{
		ID:            id,
		StoreID:       1,
		WarehouseCode: "WH-DEFAULT",
		WarehouseName: "默认仓库",
		WarehouseType: 1,
		Status:        1,
	}
	err := db.Create(warehouse).Error
	assert.NoError(t, err)
}

// createTestSKU 创建测试SKU
func createTestSKU(t *testing.T, db *gorm.DB, id, productID int64, skuCode, skuName string) {
	sku := &models.ProductSKU{
		ID:        id,
		ProductID: productID,
		SKUCode:   skuCode,
		SKUName:   skuName,
		Status:    1,
	}
	err := db.Create(sku).Error
	assert.NoError(t, err)
}

// createTestStockWithBatch 创建测试库存和批次
func createTestStockWithBatch(t *testing.T, db *gorm.DB, warehouseID, skuID int64, stockQty, batchQty int, purchasePrice float64) {
	// 创建库存
	stock := &models.WarehouseStock{
		WarehouseID:       warehouseID,
		SKUID:             skuID,
		StockQuantity:     stockQty,
		AvailableQuantity: stockQty,
		LockedQuantity:    0,
		WarningStock:      10,
		Version:           0,
	}
	err := db.Create(stock).Error
	assert.NoError(t, err)

	// 创建批次
	now := time.Now()
	batch := &models.InventoryBatch{
		SKUID:             skuID,
		BatchNo:           fmt.Sprintf("BATCH-SKU%d", skuID),
		PurchasePrice:     decimal.NewFromFloat(purchasePrice),
		TotalCost:         decimal.NewFromFloat(purchasePrice * float64(batchQty)),
		InitialQuantity:   batchQty,
		RemainingQuantity: batchQty,
		WarehouseID:       &warehouseID,
		Status:            1,
		EntryDate:         &now,
	}
	err = db.Create(batch).Error
	assert.NoError(t, err)
}

// TestCreateSingleItemOrder 创建单品订单
func TestCreateSingleItemOrder(t *testing.T) {
	orderSvc, db := setupOrderTestService(t)

	// 准备测试数据
	createTestWarehouse(t, db, 1)
	createTestSKU(t, db, 100, 1, "SKU-001", "测试商品A")
	createTestStockWithBatch(t, db, 1, 100, 50, 50, 100.00)

	// 创建单品订单
	req := &CreateOrderRequest{
		StoreID:         1,
		SalesmanID:      1,
		CustomerName:    "测试客户",
		CustomerPhone:   "13800138000",
		CustomerAddress: "测试地址",
		Items: []CreateOrderItemRequest{
			{
				SKUID:       100,
				ProductName: "测试商品A",
				SKUName:     "SKU-001",
				CategoryID:  intPtr(1),
				Quantity:    5,
				ListPrice:   200.00,
				SalePrice:   180.00,
			},
		},
	}

	order, err := orderSvc.CreateOrder(req, 1)
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.NotZero(t, order.ID)
	assert.Equal(t, int8(0), order.OrderStatus)                               // 待审批
	assert.Equal(t, int8(1), order.OrderType)                                 // 单品
	assert.True(t, order.TotalListPrice.Equal(decimal.NewFromFloat(1000.00))) // 200*5
	assert.True(t, order.TotalSalePrice.Equal(decimal.NewFromFloat(900.00)))  // 180*5
	assert.True(t, order.DiscountAmount.Equal(decimal.Zero))                  // 单品无折扣
	assert.True(t, order.FinalAmount.Equal(decimal.NewFromFloat(900.00)))
	assert.True(t, order.TotalCost.Equal(decimal.Zero))    // 创建时成本为0
	assert.True(t, order.ActualProfit.Equal(decimal.Zero)) // 创建时利润为0
	assert.Equal(t, 1, order.CategoryCount)
	assert.Equal(t, 1, order.SKUCount)
	assert.Equal(t, 5, order.TotalQuantity)

	// 验证库存锁定
	stock, _ := db.Model(&models.WarehouseStock{}).Where("warehouse_id = ? AND sku_id = ?", 1, 100).Rows()
	assert.True(t, stock.Next())
}

// TestCreateMultiItemOrder 创建多品订单
func TestCreateMultiItemOrder(t *testing.T) {
	orderSvc, db := setupOrderTestService(t)

	// 准备测试数据
	createTestWarehouse(t, db, 1)
	createTestSKU(t, db, 100, 1, "SKU-001", "测试商品A")
	createTestSKU(t, db, 101, 2, "SKU-002", "测试商品B")
	createTestStockWithBatch(t, db, 1, 100, 50, 50, 100.00)
	createTestStockWithBatch(t, db, 1, 101, 50, 50, 80.00)

	// 创建多品订单（不同品类）
	catID1 := int64(1)
	catID2 := int64(2)
	req := &CreateOrderRequest{
		StoreID:       1,
		SalesmanID:    1,
		CustomerName:  "测试客户",
		CustomerPhone: "13800138001",
		Items: []CreateOrderItemRequest{
			{
				SKUID:       100,
				ProductName: "测试商品A",
				SKUName:     "SKU-001",
				CategoryID:  &catID1,
				Quantity:    3,
				ListPrice:   200.00,
				SalePrice:   180.00,
			},
			{
				SKUID:       101,
				ProductName: "测试商品B",
				SKUName:     "SKU-002",
				CategoryID:  &catID2,
				Quantity:    2,
				ListPrice:   150.00,
				SalePrice:   130.00,
			},
		},
	}

	order, err := orderSvc.CreateOrder(req, 1)
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, int8(2), order.OrderType) // 多品
	assert.Equal(t, 2, order.CategoryCount)
	assert.Equal(t, 2, order.SKUCount)
	assert.Equal(t, 5, order.TotalQuantity)

	// total_list_price = 200*3 + 150*2 = 900
	assert.True(t, order.TotalListPrice.Equal(decimal.NewFromFloat(900.00)))
	// total_sale_price = 180*3 + 130*2 = 800
	assert.True(t, order.TotalSalePrice.Equal(decimal.NewFromFloat(800.00)))
	// 多品折扣 = 800 * 0.05 = 40
	assert.True(t, order.DiscountAmount.Equal(decimal.NewFromFloat(40.00)))
	// final_amount = 800 - 40 = 760
	assert.True(t, order.FinalAmount.Equal(decimal.NewFromFloat(760.00)))
}

// TestApproveOrder 审核通过订单
func TestApproveOrder(t *testing.T) {
	orderSvc, db := setupOrderTestService(t)

	// 准备测试数据
	createTestWarehouse(t, db, 1)
	createTestSKU(t, db, 100, 1, "SKU-001", "测试商品A")
	createTestStockWithBatch(t, db, 1, 100, 50, 50, 100.00)

	// 创建订单
	catID := int64(1)
	req := &CreateOrderRequest{
		StoreID:       1,
		SalesmanID:    1,
		CustomerName:  "测试客户",
		CustomerPhone: "13800138002",
		Items: []CreateOrderItemRequest{
			{
				SKUID:       100,
				ProductName: "测试商品A",
				SKUName:     "SKU-001",
				CategoryID:  &catID,
				Quantity:    5,
				ListPrice:   200.00,
				SalePrice:   180.00,
			},
		},
	}

	order, err := orderSvc.CreateOrder(req, 1)
	assert.NoError(t, err)
	assert.NotNil(t, order)

	// 审核通过
	err = orderSvc.ApproveOrder(order.ID, 2, true, "审核通过", "", nil)
	assert.NoError(t, err)

	// 验证订单状态
	updatedOrder, err := orderSvc.GetDetail(order.ID)
	assert.NoError(t, err)
	assert.Equal(t, int8(1), updatedOrder.Order.OrderStatus) // 已生效
	assert.NotNil(t, updatedOrder.Order.ApprovedBy)
	assert.Equal(t, int64(2), *updatedOrder.Order.ApprovedBy)

	// 验证成本计算：5个@100元 = 500
	assert.True(t, updatedOrder.Order.TotalCost.Equal(decimal.NewFromFloat(500.00)))
	// 验证利润计算：final_amount(900) - total_cost(500) - gift_cost(0) = 400
	assert.True(t, updatedOrder.Order.ActualProfit.Equal(decimal.NewFromFloat(400.00)))

	// 验证库存扣减
	var stock models.WarehouseStock
	db.Where("warehouse_id = ? AND sku_id = ?", 1, 100).First(&stock)
	assert.Equal(t, 45, stock.StockQuantity)     // 50 - 5 = 45
	assert.Equal(t, 45, stock.AvailableQuantity) // 50 - 5(lock) = 45, deduct不再减available
	assert.Equal(t, 0, stock.LockedQuantity)     // 5 - 5 = 0

	// 验证客户统计更新
	var customer models.Customer
	db.First(&customer, updatedOrder.Order.CustomerID)
	assert.Equal(t, 1, customer.TotalOrders)
	assert.True(t, customer.TotalAmount.Equal(decimal.NewFromFloat(900.00)))
	assert.True(t, customer.TotalProfit.Equal(decimal.NewFromFloat(400.00)))
}

// TestRejectOrder 审核驳回订单
func TestRejectOrder(t *testing.T) {
	orderSvc, db := setupOrderTestService(t)

	// 准备测试数据
	createTestWarehouse(t, db, 1)
	createTestSKU(t, db, 100, 1, "SKU-001", "测试商品A")
	createTestStockWithBatch(t, db, 1, 100, 50, 50, 100.00)

	// 创建订单
	catID := int64(1)
	req := &CreateOrderRequest{
		StoreID:       1,
		SalesmanID:    1,
		CustomerName:  "测试客户",
		CustomerPhone: "13800138003",
		Items: []CreateOrderItemRequest{
			{
				SKUID:       100,
				ProductName: "测试商品A",
				SKUName:     "SKU-001",
				CategoryID:  &catID,
				Quantity:    5,
				ListPrice:   200.00,
				SalePrice:   180.00,
			},
		},
	}

	order, err := orderSvc.CreateOrder(req, 1)
	assert.NoError(t, err)
	assert.NotNil(t, order)

	// 验证创建后库存已锁定
	var stockBefore models.WarehouseStock
	db.Where("warehouse_id = ? AND sku_id = ?", 1, 100).First(&stockBefore)
	assert.Equal(t, 45, stockBefore.AvailableQuantity) // 50 - 5 = 45
	assert.Equal(t, 5, stockBefore.LockedQuantity)     // 0 + 5 = 5

	// 审核驳回
	err = orderSvc.ApproveOrder(order.ID, 2, false, "价格不合理", "", nil)
	assert.NoError(t, err)

	// 验证订单状态
	updatedOrder, err := orderSvc.GetDetail(order.ID)
	assert.NoError(t, err)
	assert.Equal(t, int8(2), updatedOrder.Order.OrderStatus) // 已驳回

	// 验证库存释放
	var stockAfter models.WarehouseStock
	db.Where("warehouse_id = ? AND sku_id = ?", 1, 100).First(&stockAfter)
	assert.Equal(t, 50, stockAfter.AvailableQuantity) // 45 + 5 = 50
	assert.Equal(t, 0, stockAfter.LockedQuantity)     // 5 - 5 = 0
}

// TestCancelOrder 取消订单
func TestCancelOrder(t *testing.T) {
	orderSvc, db := setupOrderTestService(t)

	// 准备测试数据
	createTestWarehouse(t, db, 1)
	createTestSKU(t, db, 100, 1, "SKU-001", "测试商品A")
	createTestStockWithBatch(t, db, 1, 100, 50, 50, 100.00)

	// 创建订单
	catID := int64(1)
	req := &CreateOrderRequest{
		StoreID:       1,
		SalesmanID:    1,
		CustomerName:  "测试客户",
		CustomerPhone: "13800138004",
		Items: []CreateOrderItemRequest{
			{
				SKUID:       100,
				ProductName: "测试商品A",
				SKUName:     "SKU-001",
				CategoryID:  &catID,
				Quantity:    3,
				ListPrice:   200.00,
				SalePrice:   180.00,
			},
		},
	}

	order, err := orderSvc.CreateOrder(req, 1)
	assert.NoError(t, err)

	// 取消订单
	err = orderSvc.CancelOrder(order.ID)
	assert.NoError(t, err)

	// 验证订单状态
	updatedOrder, err := orderSvc.GetDetail(order.ID)
	assert.NoError(t, err)
	assert.Equal(t, int8(3), updatedOrder.Order.OrderStatus) // 已取消

	// 验证库存释放
	var stock models.WarehouseStock
	db.Where("warehouse_id = ? AND sku_id = ?", 1, 100).First(&stock)
	assert.Equal(t, 50, stock.AvailableQuantity) // 恢复
	assert.Equal(t, 0, stock.LockedQuantity)
}

// TestReturnOrder 退货处理
func TestReturnOrder(t *testing.T) {
	orderSvc, db := setupOrderTestService(t)

	// 准备测试数据
	createTestWarehouse(t, db, 1)
	createTestSKU(t, db, 100, 1, "SKU-001", "测试商品A")
	createTestStockWithBatch(t, db, 1, 100, 50, 50, 100.00)

	// 创建并审核通过订单
	catID := int64(1)
	req := &CreateOrderRequest{
		StoreID:       1,
		SalesmanID:    1,
		CustomerName:  "测试客户",
		CustomerPhone: "13800138005",
		Items: []CreateOrderItemRequest{
			{
				SKUID:       100,
				ProductName: "测试商品A",
				SKUName:     "SKU-001",
				CategoryID:  &catID,
				Quantity:    5,
				ListPrice:   200.00,
				SalePrice:   180.00,
			},
		},
	}

	order, err := orderSvc.CreateOrder(req, 1)
	assert.NoError(t, err)

	err = orderSvc.ApproveOrder(order.ID, 2, true, "审核通过", "", nil)
	assert.NoError(t, err)

	// 退货处理
	err = orderSvc.ReturnOrder(order.ID, 300.00, 100.00, 1, "客户退货", 1, "测试用户")
	assert.NoError(t, err)

	// 验证订单状态
	updatedOrder, err := orderSvc.GetDetail(order.ID)
	assert.NoError(t, err)
	assert.Equal(t, int8(4), updatedOrder.Order.OrderStatus) // 已退货
	assert.Equal(t, int8(1), updatedOrder.Order.IsReturned)
	assert.True(t, updatedOrder.Order.ReturnAmount.Equal(decimal.NewFromFloat(300.00)))
	assert.True(t, updatedOrder.Order.ReturnProfit.Equal(decimal.NewFromFloat(100.00)))
}

// TestCreateOrderWithGift 创建带赠品的订单
func TestCreateOrderWithGift(t *testing.T) {
	orderSvc, db := setupOrderTestService(t)

	// 准备测试数据
	createTestWarehouse(t, db, 1)
	createTestSKU(t, db, 100, 1, "SKU-001", "测试商品A")
	createTestStockWithBatch(t, db, 1, 100, 50, 50, 100.00)

	catID := int64(1)
	req := &CreateOrderRequest{
		StoreID:       1,
		SalesmanID:    1,
		CustomerName:  "测试客户",
		CustomerPhone: "13800138006",
		Items: []CreateOrderItemRequest{
			{
				SKUID:       100,
				ProductName: "测试商品A",
				SKUName:     "SKU-001",
				CategoryID:  &catID,
				Quantity:    2,
				ListPrice:   200.00,
				SalePrice:   180.00,
			},
		},
		Gifts: []CreateOrderGiftRequest{
			{
				GiftID:    1,
				GiftName:  "赠品A",
				CostPrice: 50.00,
				Quantity:  1,
			},
			{
				GiftID:    2,
				GiftName:  "赠品B",
				CostPrice: 30.00,
				Quantity:  2,
			},
		},
	}

	order, err := orderSvc.CreateOrder(req, 1)
	assert.NoError(t, err)
	assert.NotNil(t, order)

	// 验证赠品成本：50*1 + 30*2 = 110
	assert.True(t, order.GiftCost.Equal(decimal.NewFromFloat(110.00)))
}

// TestApproveInvalidStatus 审核无效状态的订单
func TestApproveInvalidStatus(t *testing.T) {
	orderSvc, db := setupOrderTestService(t)

	// 准备测试数据
	createTestWarehouse(t, db, 1)
	createTestSKU(t, db, 100, 1, "SKU-001", "测试商品A")
	createTestStockWithBatch(t, db, 1, 100, 50, 50, 100.00)

	catID := int64(1)
	req := &CreateOrderRequest{
		StoreID:       1,
		SalesmanID:    1,
		CustomerName:  "测试客户",
		CustomerPhone: "13800138007",
		Items: []CreateOrderItemRequest{
			{
				SKUID:       100,
				ProductName: "测试商品A",
				SKUName:     "SKU-001",
				CategoryID:  &catID,
				Quantity:    1,
				ListPrice:   200.00,
				SalePrice:   180.00,
			},
		},
	}

	order, err := orderSvc.CreateOrder(req, 1)
	assert.NoError(t, err)

	// 审核通过
	err = orderSvc.ApproveOrder(order.ID, 2, true, "审核通过", "", nil)
	assert.NoError(t, err)

	// 再次审核（应失败，因为状态已不是待审批）
	err = orderSvc.ApproveOrder(order.ID, 2, true, "再次审核", "", nil)
	assert.Error(t, err)
	appErr, ok := err.(*AppError)
	assert.True(t, ok)
	assert.Equal(t, apperrors.ErrInvalidOrderStatus, appErr.Code)
}

// TestCreateOrderInsufficientStock 库存不足
func TestCreateOrderInsufficientStock(t *testing.T) {
	orderSvc, db := setupOrderTestService(t)

	// 准备测试数据
	createTestWarehouse(t, db, 1)
	createTestSKU(t, db, 100, 1, "SKU-001", "测试商品A")
	createTestStockWithBatch(t, db, 1, 100, 3, 3, 100.00) // 只有3个库存

	catID := int64(1)
	req := &CreateOrderRequest{
		StoreID:       1,
		SalesmanID:    1,
		CustomerName:  "测试客户",
		CustomerPhone: "13800138008",
		Items: []CreateOrderItemRequest{
			{
				SKUID:       100,
				ProductName: "测试商品A",
				SKUName:     "SKU-001",
				CategoryID:  &catID,
				Quantity:    5, // 需要5个，但只有3个
				ListPrice:   200.00,
				SalePrice:   180.00,
			},
		},
	}

	_, err := orderSvc.CreateOrder(req, 1)
	assert.Error(t, err)
	appErr, ok := err.(*AppError)
	assert.True(t, ok)
	assert.Equal(t, apperrors.ErrInsufficientStock, appErr.Code)
}

// TestPeerOrder 同行订单
func TestPeerOrder(t *testing.T) {
	orderSvc, db := setupOrderTestService(t)

	// 准备测试数据
	createTestWarehouse(t, db, 1)
	createTestSKU(t, db, 100, 1, "SKU-001", "测试商品A")
	createTestStockWithBatch(t, db, 1, 100, 50, 50, 100.00)

	// 创建同行
	peer := &models.Peer{
		StoreID:  1,
		PeerName: "测试同行",
		Phone:    "13900139000",
		Status:   1,
	}
	err := db.Create(peer).Error
	assert.NoError(t, err)

	catID := int64(1)
	req := &CreateOrderRequest{
		StoreID:       1,
		SalesmanID:    1,
		CustomerName:  "同行客户",
		CustomerPhone: "13800138009",
		IsPeerOrder:   1,
		PeerID:        &peer.ID,
		Items: []CreateOrderItemRequest{
			{
				SKUID:       100,
				ProductName: "测试商品A",
				SKUName:     "SKU-001",
				CategoryID:  &catID,
				Quantity:    2,
				ListPrice:   200.00,
				SalePrice:   180.00,
			},
		},
	}

	order, err := orderSvc.CreateOrder(req, 1)
	assert.NoError(t, err)
	assert.Equal(t, int8(4), order.OrderType) // 同行单品
	assert.Equal(t, int8(1), order.IsPeerOrder)
}

// TestSpecialApprovedOrder 特批订单
func TestSpecialApprovedOrder(t *testing.T) {
	orderSvc, db := setupOrderTestService(t)

	// 准备测试数据
	createTestWarehouse(t, db, 1)
	createTestSKU(t, db, 100, 1, "SKU-001", "测试商品A")
	createTestStockWithBatch(t, db, 1, 100, 50, 50, 100.00)

	catID := int64(1)
	req := &CreateOrderRequest{
		StoreID:           1,
		SalesmanID:        1,
		CustomerName:      "特批客户",
		CustomerPhone:     "13800138010",
		IsSpecialApproved: 1,
		ApprovalRemark:    "特批折扣",
		Items: []CreateOrderItemRequest{
			{
				SKUID:       100,
				ProductName: "测试商品A",
				SKUName:     "SKU-001",
				CategoryID:  &catID,
				Quantity:    2,
				ListPrice:   200.00,
				SalePrice:   180.00,
			},
		},
	}

	order, err := orderSvc.CreateOrder(req, 1)
	assert.NoError(t, err)
	assert.Equal(t, int8(3), order.OrderType) // 特殊审批
}

// intPtr 辅助函数：创建int64指针
func intPtr(v int64) *int64 {
	return &v
}
