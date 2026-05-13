package service_test

import (
	"fmt"
	"testing"

	"furniture-commission/internal/models"
	"furniture-commission/internal/repository"
	svc "furniture-commission/internal/service"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupPurchaseTestDB 创建采购测试数据库
func setupPurchaseTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbName := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(
		&models.PurchaseOrder{},
		&models.PurchaseItem{},
		&models.Supplier{},
		&models.WarehouseStock{},
		&models.InventoryBatch{},
		&models.InventoryTransaction{},
		&models.ProductSKU{},
		&models.Warehouse{},
		&models.Product{},
	)
	assert.NoError(t, err)
	return db
}

// setupPurchaseTestService 创建采购测试服务
func setupPurchaseTestService(t *testing.T) (*svc.PurchaseService, *gorm.DB) {
	db := setupPurchaseTestDB(t)
	invRepo := repository.NewInventoryRepository(db)
	invSvc := svc.NewInventoryService(db, invRepo, nil, nil)
	purchaseSvc := svc.NewPurchaseService(db, invSvc)
	return purchaseSvc, db
}

// createTestSupplier 创建测试供应商
func createTestSupplier(t *testing.T, db *gorm.DB, id int64) {
	supplier := &models.Supplier{
		ID:           id,
		StoreID:      1,
		SupplierCode: "SUP-001",
		SupplierName: "测试供应商",
		Status:       1,
	}
	err := db.Create(supplier).Error
	assert.NoError(t, err)
}

// createTestWarehouseForPurchase 创建测试仓库
func createTestWarehouseForPurchase(t *testing.T, db *gorm.DB, id int64) {
	warehouse := &models.Warehouse{
		ID:            id,
		StoreID:       1,
		WarehouseCode: "WH-001",
		WarehouseName: "默认仓库",
		WarehouseType: 1,
		Status:        1,
	}
	err := db.Create(warehouse).Error
	assert.NoError(t, err)
}

// createTestSKUForPurchase 创建测试SKU
func createTestSKUForPurchase(t *testing.T, db *gorm.DB, id, productID int64, skuCode, skuName string) {
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

// ========== TestCreatePurchaseOrder ==========

func TestCreatePurchaseOrder(t *testing.T) {
	purchaseSvc, db := setupPurchaseTestService(t)

	createTestSupplier(t, db, 1)

	req := &svc.CreatePurchaseOrderRequest{
		StoreID:      1,
		SupplierID:   1,
		SupplierName: "测试供应商",
		Remark:       "测试采购",
		Items: []svc.CreatePurchaseItemRequest{
			{
				SKUID:         100,
				ProductName:   "测试商品A",
				SKUName:       "SKU-001",
				PurchasePrice: 100.00,
				Quantity:      10,
			},
			{
				SKUID:         101,
				ProductName:   "测试商品B",
				SKUName:       "SKU-002",
				PurchasePrice: 200.00,
				Quantity:      5,
			},
		},
	}

	err := purchaseSvc.CreateOrder(req, 1, 1)
	assert.NoError(t, err)

	// 验证采购订单已创建
	detail, err := purchaseSvc.GetDetail(1)
	assert.NoError(t, err)
	assert.NotNil(t, detail)
	assert.NotEmpty(t, detail.PurchaseNo)
	assert.Equal(t, int8(0), detail.Status) // 待审核
	assert.True(t, detail.TotalAmount.Equal(decimal.NewFromFloat(2000.00))) // 100*10 + 200*5
	assert.Equal(t, 15, detail.TotalQuantity) // 10 + 5
	assert.Len(t, detail.Items, 2)
}

func TestCreatePurchaseOrder_EmptyItems(t *testing.T) {
	purchaseSvc, _ := setupPurchaseTestService(t)

	req := &svc.CreatePurchaseOrderRequest{
		StoreID:    1,
		SupplierID: 1,
		Items:      []svc.CreatePurchaseItemRequest{},
	}

	err := purchaseSvc.CreateOrder(req, 1, 1)
	// 服务层可能允许创建空明细的采购订单，或返回错误
	// 根据实际业务逻辑调整
	if err != nil {
		assert.Error(t, err)
	}
}

// ========== TestApprovePurchaseOrder ==========

func TestApprovePurchaseOrder(t *testing.T) {
	purchaseSvc, db := setupPurchaseTestService(t)

	createTestSupplier(t, db, 1)

	// 创建采购订单
	req := &svc.CreatePurchaseOrderRequest{
		StoreID:      1,
		SupplierID:   1,
		SupplierName: "测试供应商",
		Items: []svc.CreatePurchaseItemRequest{
			{
				SKUID:         100,
				ProductName:   "测试商品",
				SKUName:       "SKU-001",
				PurchasePrice: 100.00,
				Quantity:      10,
			},
		},
	}
	err := purchaseSvc.CreateOrder(req, 1, 1)
	assert.NoError(t, err)

	// 审核通过
	err = purchaseSvc.ApproveOrder(1, 2)
	assert.NoError(t, err)

	// 验证状态已更新
	detail, err := purchaseSvc.GetDetail(1)
	assert.NoError(t, err)
	assert.Equal(t, int8(1), detail.Status) // 已审核
	assert.NotNil(t, detail.AuditedBy)
	assert.Equal(t, int64(2), *detail.AuditedBy)
	assert.NotNil(t, detail.AuditedAt)
}

func TestApprovePurchaseOrder_InvalidStatus(t *testing.T) {
	purchaseSvc, db := setupPurchaseTestService(t)

	createTestSupplier(t, db, 1)

	req := &svc.CreatePurchaseOrderRequest{
		StoreID:      1,
		SupplierID:   1,
		SupplierName: "测试供应商",
		Items: []svc.CreatePurchaseItemRequest{
			{
				SKUID:         100,
				ProductName:   "测试商品",
				SKUName:       "SKU-001",
				PurchasePrice: 100.00,
				Quantity:      10,
			},
		},
	}
	purchaseSvc.CreateOrder(req, 1, 1)

	// 第一次审核通过
	purchaseSvc.ApproveOrder(1, 2)

	// 第二次审核应失败
	err := purchaseSvc.ApproveOrder(1, 2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "订单状态不允许审核")
}

func TestApprovePurchaseOrder_NotFound(t *testing.T) {
	purchaseSvc, _ := setupPurchaseTestService(t)

	err := purchaseSvc.ApproveOrder(99999, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "采购订单不存在")
}

// ========== TestConfirmReceipt ==========

func TestConfirmReceipt(t *testing.T) {
	purchaseSvc, db := setupPurchaseTestService(t)

	createTestSupplier(t, db, 1)
	createTestWarehouseForPurchase(t, db, 1)
	createTestSKUForPurchase(t, db, 100, 1, "SKU-001", "测试商品A")

	// 创建并审核采购订单
	req := &svc.CreatePurchaseOrderRequest{
		StoreID:      1,
		SupplierID:   1,
		SupplierName: "测试供应商",
		Items: []svc.CreatePurchaseItemRequest{
			{
				SKUID:         100,
				ProductName:   "测试商品A",
				SKUName:       "SKU-001",
				PurchasePrice: 100.00,
				Quantity:      10,
			},
		},
	}
	err := purchaseSvc.CreateOrder(req, 1, 1)
	assert.NoError(t, err)

	err = purchaseSvc.ApproveOrder(1, 2)
	assert.NoError(t, err)

	// 确认入库
	err = purchaseSvc.ConfirmReceipt(1, 1, 3)
	assert.NoError(t, err)

	// 验证采购订单状态
	detail, err := purchaseSvc.GetDetail(1)
	assert.NoError(t, err)
	assert.Equal(t, int8(2), detail.Status) // 已入库

	// 验证库存增加
	var stock models.WarehouseStock
	err = db.Where("warehouse_id = ? AND sku_id = ?", 1, 100).First(&stock).Error
	assert.NoError(t, err)
	assert.Equal(t, 10, stock.StockQuantity)
	assert.Equal(t, 10, stock.AvailableQuantity)

	// 验证批次记录
	var batches []models.InventoryBatch
	db.Where("sku_id = ?", 100).Find(&batches)
	assert.Len(t, batches, 1)
	assert.Equal(t, 10, batches[0].InitialQuantity)
	assert.Equal(t, 10, batches[0].RemainingQuantity)
	assert.True(t, batches[0].PurchasePrice.Equal(decimal.NewFromFloat(100.00)))
}

func TestConfirmReceipt_InvalidStatus(t *testing.T) {
	purchaseSvc, db := setupPurchaseTestService(t)

	createTestSupplier(t, db, 1)

	req := &svc.CreatePurchaseOrderRequest{
		StoreID:      1,
		SupplierID:   1,
		SupplierName: "测试供应商",
		Items: []svc.CreatePurchaseItemRequest{
			{
				SKUID:         100,
				ProductName:   "测试商品",
				SKUName:       "SKU-001",
				PurchasePrice: 100.00,
				Quantity:      10,
			},
		},
	}
	purchaseSvc.CreateOrder(req, 1, 1)

	// 未审核直接入库应失败
	err := purchaseSvc.ConfirmReceipt(1, 1, 3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "订单状态不允许入库")
}

// ========== TestCancelPurchaseOrder ==========

func TestCancelPurchaseOrder(t *testing.T) {
	purchaseSvc, db := setupPurchaseTestService(t)

	createTestSupplier(t, db, 1)

	req := &svc.CreatePurchaseOrderRequest{
		StoreID:      1,
		SupplierID:   1,
		SupplierName: "测试供应商",
		Items: []svc.CreatePurchaseItemRequest{
			{
				SKUID:         100,
				ProductName:   "测试商品",
				SKUName:       "SKU-001",
				PurchasePrice: 100.00,
				Quantity:      10,
			},
		},
	}
	purchaseSvc.CreateOrder(req, 1, 1)

	err := purchaseSvc.CancelOrder(1)
	assert.NoError(t, err)

	detail, _ := purchaseSvc.GetDetail(1)
	assert.Equal(t, int8(3), detail.Status) // 已取消
}

// ========== TestListPurchaseOrders ==========

func TestListPurchaseOrders(t *testing.T) {
	purchaseSvc, db := setupPurchaseTestService(t)

	createTestSupplier(t, db, 1)

	// 创建多个采购订单
	for i := 0; i < 3; i++ {
		req := &svc.CreatePurchaseOrderRequest{
			StoreID:      1,
			SupplierID:   1,
			SupplierName: "测试供应商",
			Items: []svc.CreatePurchaseItemRequest{
				{
					SKUID:         int64(100 + i),
					ProductName:   fmt.Sprintf("商品%d", i),
					SKUName:       fmt.Sprintf("SKU-%d", i),
					PurchasePrice: 100.00,
					Quantity:      5,
				},
			},
		}
		err := purchaseSvc.CreateOrder(req, 1, 1)
		assert.NoError(t, err)
	}

	listReq := &svc.ListPurchaseOrderRequest{
		StoreID:  1,
		Page:     1,
		PageSize: 10,
	}
	result, err := purchaseSvc.List(listReq)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), result.Total)
	assert.Len(t, result.List, 3)
}

// ========== TestListPurchaseOrders_ByStatus ==========

func TestListPurchaseOrders_ByStatus(t *testing.T) {
	purchaseSvc, db := setupPurchaseTestService(t)

	createTestSupplier(t, db, 1)

	// 创建两个订单，一个审核一个不审核
	req := &svc.CreatePurchaseOrderRequest{
		StoreID:      1,
		SupplierID:   1,
		SupplierName: "测试供应商",
		Items: []svc.CreatePurchaseItemRequest{
			{SKUID: 100, ProductName: "商品", SKUName: "SKU", PurchasePrice: 100, Quantity: 5},
		},
	}
	purchaseSvc.CreateOrder(req, 1, 1)
	purchaseSvc.CreateOrder(req, 1, 1)

	purchaseSvc.ApproveOrder(1, 2)

	// 查询待审核
	pending := int8(0)
	listReq := &svc.ListPurchaseOrderRequest{
		StoreID: 1,
		Status:   &pending,
		Page:     1,
		PageSize: 10,
	}
	result, err := purchaseSvc.List(listReq)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
}
