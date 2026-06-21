package integration_test

import (
	"fmt"
	"testing"
	"time"

	"furniture-commission/configs"
	"furniture-commission/internal/models"
	"furniture-commission/internal/repository"
	svc "furniture-commission/internal/service"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ReceiptFlowTestSuite 回货单流程测试套件
type ReceiptFlowTestSuite struct {
	suite.Suite
	db               *gorm.DB
	receiptSvc       *svc.ReceiptService
	inventorySvc     *svc.InventoryService
	inventoryRepo    *repository.InventoryRepository
	purchaseItemRepo *repository.BaseRepository[models.PurchaseItem]
}

func (s *ReceiptFlowTestSuite) SetupSuite() {
	dbName := fmt.Sprintf("file:/tmp/receipt_flow_%d.db", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	s.Require().NoError(err)
	s.db = db

	err = db.AutoMigrate(
		&models.Store{},
		&models.Warehouse{},
		&models.WarehouseStock{},
		&models.InventoryBatch{},
		&models.InventoryTransaction{},
		&models.Product{},
		&models.ProductSKU{},
		&models.Supplier{},
		&models.PurchaseOrder{},
		&models.PurchaseItem{},
		&models.ReceiptOrder{},
		&models.ReceiptItem{},
	)
	s.Require().NoError(err)

	s.inventoryRepo = repository.NewInventoryRepository(db)
	s.purchaseItemRepo = repository.NewBaseRepository[models.PurchaseItem](db)

	s.inventorySvc = svc.NewInventoryService(db, s.inventoryRepo, nil, nil)
	s.receiptSvc = svc.NewReceiptService(db, s.inventorySvc)

	configs.GlobalConfig = &configs.Config{
		JWT: configs.JWTConfig{
			Secret:      "test-secret-key",
			ExpireHours: 24,
		},
	}
}

func (s *ReceiptFlowTestSuite) TearDownSuite() {
	sqlDB, _ := s.db.DB()
	sqlDB.Close()
}

func (s *ReceiptFlowTestSuite) SetupTest() {
	tables := []string{
		"receipt_items", "receipt_orders", "purchase_items", "purchase_orders",
		"inventory_transactions", "inventory_batches", "warehouse_stocks",
		"product_skus", "products", "suppliers", "warehouses", "stores",
	}
	for _, table := range tables {
		s.db.Exec(fmt.Sprintf("DELETE FROM %s", table))
	}
}

func (s *ReceiptFlowTestSuite) seedBaseData() {
	store := &models.Store{
		ID:        1,
		StoreCode: "STORE-001",
		StoreName: "测试门店",
		Status:    1,
	}
	s.db.Create(store)

	warehouse := &models.Warehouse{
		ID:            1,
		StoreID:       1,
		WarehouseCode: "WH-001",
		WarehouseName: "默认仓库",
		WarehouseType: 1,
		Status:        1,
	}
	s.db.Create(warehouse)

	supplier := &models.Supplier{
		ID:           1,
		SupplierCode: "SUP-001",
		SupplierName: "测试供应商",
		Status:       1,
	}
	s.db.Create(supplier)

	sku := &models.ProductSKU{
		ID:        100,
		ProductID: 1,
		SKUCode:   "SKU-SOFA-001",
		SKUName:   "真皮沙发-标准款",
		Status:    1,
	}
	s.db.Create(sku)

	stock := &models.WarehouseStock{
		WarehouseID:       1,
		SKUID:             100,
		StockQuantity:     0,
		AvailableQuantity: 0,
		LockedQuantity:    0,
		WarningStock:      10,
		Version:           0,
	}
	s.db.Create(stock)

	supplierID := int64(1)
	purchaseOrder := &models.PurchaseOrder{
		ID:           1,
		PurchaseNo:   "PO-20240101-001",
		SupplierID:   &supplierID,
		SupplierName: "测试供应商",
		StoreID:      1,
		Status:       1,
		TotalAmount:  decimal.NewFromFloat(5000.00),
	}
	s.db.Create(purchaseOrder)

	skuID := int64(100)
	purchaseItem := &models.PurchaseItem{
		ID:              1,
		PurchaseOrderID: 1,
		SKUID:           &skuID,
		ProductName:     "真皮沙发",
		SKUName:         "真皮沙发-标准款",
		PurchasePrice:   decimal.NewFromFloat(1000.00),
		Quantity:        10,
		ReceivedQuantity: 0,
		Subtotal:        decimal.NewFromFloat(10000.00),
	}
	s.db.Create(purchaseItem)
}

func (s *ReceiptFlowTestSuite) TestCreateReceiptOrder() {
	s.seedBaseData()

	req := &svc.CreateReceiptRequest{
		StoreID:      1,
		SupplierID:   1,
		SupplierName: "测试供应商",
		Remark:       "测试回货单",
		Items: []svc.CreateReceiptItemRequest{
			{
				PurchaseItemID: intPtr64(1),
				SKUID:          100,
				ProductName:    "真皮沙发",
				SKUCode:        "SKU-SOFA-001",
				SpecInfo:       "标准款-黑色",
				ShipQuantity:   10,
				CostPrice:      1000.00,
			},
		},
	}

	err := s.receiptSvc.CreateReceiptOrder(req, 1)
	s.NoError(err)

	var orders []models.ReceiptOrder
	s.db.Find(&orders)
	s.Len(orders, 1)
	s.Equal(models.ReceiptStatusPending, orders[0].Status)
	s.Equal(10, orders[0].TotalQuantity)
	s.True(orders[0].TotalAmount.Equal(decimal.NewFromFloat(10000.00)))

	var items []models.ReceiptItem
	s.db.Where("receipt_order_id = ?", orders[0].ID).Find(&items)
	s.Len(items, 1)
	s.Equal(10, items[0].ShipQuantity)
	s.Equal(0, items[0].ReceiveQuantity)
}

func (s *ReceiptFlowTestSuite) TestApproveReceiptOrder() {
	s.seedBaseData()

	req := &svc.CreateReceiptRequest{
		StoreID:      1,
		SupplierID:   1,
		SupplierName: "测试供应商",
		Items: []svc.CreateReceiptItemRequest{
			{
				SKUID:        100,
				ProductName:  "真皮沙发",
				SKUCode:      "SKU-SOFA-001",
				SpecInfo:     "标准款",
				ShipQuantity: 5,
				CostPrice:    1000.00,
			},
		},
	}

	err := s.receiptSvc.CreateReceiptOrder(req, 1)
	s.NoError(err)

	var order models.ReceiptOrder
	s.db.First(&order)

	err = s.receiptSvc.ApproveReceiptOrder(order.ID, 1)
	s.NoError(err)

	s.db.First(&order, order.ID)
	s.Equal(models.ReceiptStatusApproved, order.Status)
	s.NotNil(order.AuditedAt)
	s.Equal(int64(1), *order.AuditedBy)
}

func (s *ReceiptFlowTestSuite) TestReceiveReceiptOrder_Partial() {
	s.seedBaseData()

	req := &svc.CreateReceiptRequest{
		StoreID:      1,
		SupplierID:   1,
		SupplierName: "测试供应商",
		Items: []svc.CreateReceiptItemRequest{
			{
				PurchaseItemID: intPtr64(1),
				SKUID:          100,
				ProductName:    "真皮沙发",
				SKUCode:        "SKU-SOFA-001",
				SpecInfo:       "标准款-黑色",
				ShipQuantity:   10,
				CostPrice:      1000.00,
			},
		},
	}

	err := s.receiptSvc.CreateReceiptOrder(req, 1)
	s.NoError(err)

	var order models.ReceiptOrder
	s.db.First(&order)

	var receiptItems []models.ReceiptItem
	s.db.Where("receipt_order_id = ?", order.ID).Find(&receiptItems)
	s.Len(receiptItems, 1)
	receiptItemID := receiptItems[0].ID

	err = s.receiptSvc.ApproveReceiptOrder(order.ID, 1)
	s.NoError(err)

	receiveReq := &svc.ReceiveRequest{
		WarehouseID: 1,
		Items: []svc.ReceiveItemRequest{
			{
				ReceiptItemID:   receiptItemID,
				ReceiveQuantity: 3,
			},
		},
	}

	err = s.receiptSvc.ReceiveReceiptOrder(order.ID, receiveReq, 1)
	s.NoError(err)

	var receiptItem models.ReceiptItem
	s.db.First(&receiptItem, receiptItemID)
	s.Equal(3, receiptItem.ReceiveQuantity)

	var purchaseItem models.PurchaseItem
	s.db.First(&purchaseItem, 1)
	s.Equal(3, purchaseItem.ReceivedQuantity)

	var stock models.WarehouseStock
	s.db.Where("warehouse_id = ? AND sku_id = ?", 1, 100).First(&stock)
	s.Equal(3, stock.StockQuantity)
	s.Equal(3, stock.AvailableQuantity)

	s.db.First(&order, order.ID)
	s.Equal(models.ReceiptStatusApproved, order.Status)
}

func (s *ReceiptFlowTestSuite) TestReceiveReceiptOrder_Full() {
	s.seedBaseData()

	req := &svc.CreateReceiptRequest{
		StoreID:      1,
		SupplierID:   1,
		SupplierName: "测试供应商",
		Items: []svc.CreateReceiptItemRequest{
			{
				PurchaseItemID: intPtr64(1),
				SKUID:          100,
				ProductName:    "真皮沙发",
				SKUCode:        "SKU-SOFA-001",
				SpecInfo:       "标准款-黑色",
				ShipQuantity:   10,
				CostPrice:      1000.00,
			},
		},
	}

	err := s.receiptSvc.CreateReceiptOrder(req, 1)
	s.NoError(err)

	var order models.ReceiptOrder
	s.db.First(&order)

	var receiptItems []models.ReceiptItem
	s.db.Where("receipt_order_id = ?", order.ID).Find(&receiptItems)
	s.Len(receiptItems, 1)
	receiptItemID := receiptItems[0].ID

	err = s.receiptSvc.ApproveReceiptOrder(order.ID, 1)
	s.NoError(err)

	receiveReq := &svc.ReceiveRequest{
		WarehouseID: 1,
		Items: []svc.ReceiveItemRequest{
			{
				ReceiptItemID:   receiptItemID,
				ReceiveQuantity: 10,
			},
		},
	}

	err = s.receiptSvc.ReceiveReceiptOrder(order.ID, receiveReq, 1)
	s.NoError(err)

	var receiptItem models.ReceiptItem
	s.db.First(&receiptItem, receiptItemID)
	s.Equal(10, receiptItem.ReceiveQuantity)

	var purchaseItem models.PurchaseItem
	s.db.First(&purchaseItem, 1)
	s.Equal(10, purchaseItem.ReceivedQuantity)

	var stock models.WarehouseStock
	s.db.Where("warehouse_id = ? AND sku_id = ?", 1, 100).First(&stock)
	s.Equal(10, stock.StockQuantity)
	s.Equal(10, stock.AvailableQuantity)

	s.db.First(&order, order.ID)
	s.Equal(models.ReceiptStatusReceived, order.Status)

	var purchaseOrder models.PurchaseOrder
	s.db.First(&purchaseOrder, 1)
	s.Equal(int8(2), purchaseOrder.Status)
}

func (s *ReceiptFlowTestSuite) TestReceiveReceiptOrder_MultiplePartial() {
	s.seedBaseData()

	req := &svc.CreateReceiptRequest{
		StoreID:      1,
		SupplierID:   1,
		SupplierName: "测试供应商",
		Items: []svc.CreateReceiptItemRequest{
			{
				PurchaseItemID: intPtr64(1),
				SKUID:          100,
				ProductName:    "真皮沙发",
				SKUCode:        "SKU-SOFA-001",
				SpecInfo:       "标准款-黑色",
				ShipQuantity:   10,
				CostPrice:      1000.00,
			},
		},
	}

	err := s.receiptSvc.CreateReceiptOrder(req, 1)
	s.NoError(err)

	var order models.ReceiptOrder
	s.db.First(&order)

	var receiptItems []models.ReceiptItem
	s.db.Where("receipt_order_id = ?", order.ID).Find(&receiptItems)
	s.Len(receiptItems, 1)
	receiptItemID := receiptItems[0].ID

	err = s.receiptSvc.ApproveReceiptOrder(order.ID, 1)
	s.NoError(err)

	receiveReq1 := &svc.ReceiveRequest{
		WarehouseID: 1,
		Items: []svc.ReceiveItemRequest{
			{
				ReceiptItemID:   receiptItemID,
				ReceiveQuantity: 4,
			},
		},
	}
	err = s.receiptSvc.ReceiveReceiptOrder(order.ID, receiveReq1, 1)
	s.NoError(err)

	var receiptItem models.ReceiptItem
	s.db.First(&receiptItem, receiptItemID)
	s.Equal(4, receiptItem.ReceiveQuantity)

	receiveReq2 := &svc.ReceiveRequest{
		WarehouseID: 1,
		Items: []svc.ReceiveItemRequest{
			{
				ReceiptItemID:   receiptItemID,
				ReceiveQuantity: 3,
			},
		},
	}
	err = s.receiptSvc.ReceiveReceiptOrder(order.ID, receiveReq2, 1)
	s.NoError(err)

	s.db.First(&receiptItem, receiptItemID)
	s.Equal(7, receiptItem.ReceiveQuantity)

	receiveReq3 := &svc.ReceiveRequest{
		WarehouseID: 1,
		Items: []svc.ReceiveItemRequest{
			{
				ReceiptItemID:   receiptItemID,
				ReceiveQuantity: 3,
			},
		},
	}
	err = s.receiptSvc.ReceiveReceiptOrder(order.ID, receiveReq3, 1)
	s.NoError(err)

	s.db.First(&receiptItem, receiptItemID)
	s.Equal(10, receiptItem.ReceiveQuantity)

	s.db.First(&order, order.ID)
	s.Equal(models.ReceiptStatusReceived, order.Status)
}

func (s *ReceiptFlowTestSuite) TestReceiveReceiptOrder_ExceedsShipQuantity() {
	s.seedBaseData()

	req := &svc.CreateReceiptRequest{
		StoreID:      1,
		SupplierID:   1,
		SupplierName: "测试供应商",
		Items: []svc.CreateReceiptItemRequest{
			{
				SKUID:        100,
				ProductName:  "真皮沙发",
				SKUCode:      "SKU-SOFA-001",
				SpecInfo:     "标准款",
				ShipQuantity: 5,
				CostPrice:    1000.00,
			},
		},
	}

	err := s.receiptSvc.CreateReceiptOrder(req, 1)
	s.NoError(err)

	var order models.ReceiptOrder
	s.db.First(&order)

	var receiptItems []models.ReceiptItem
	s.db.Where("receipt_order_id = ?", order.ID).Find(&receiptItems)
	s.Len(receiptItems, 1)
	receiptItemID := receiptItems[0].ID

	err = s.receiptSvc.ApproveReceiptOrder(order.ID, 1)
	s.NoError(err)

	receiveReq := &svc.ReceiveRequest{
		WarehouseID: 1,
		Items: []svc.ReceiveItemRequest{
			{
				ReceiptItemID:   receiptItemID,
				ReceiveQuantity: 6,
			},
		},
	}

	err = s.receiptSvc.ReceiveReceiptOrder(order.ID, receiveReq, 1)
	s.Error(err)
	s.Contains(err.Error(), "实际收货数量不能超过发货数量")
}

func (s *ReceiptFlowTestSuite) TestCancelReceiptOrder() {
	s.seedBaseData()

	req := &svc.CreateReceiptRequest{
		StoreID:      1,
		SupplierID:   1,
		SupplierName: "测试供应商",
		Items: []svc.CreateReceiptItemRequest{
			{
				SKUID:        100,
				ProductName:  "真皮沙发",
				SKUCode:      "SKU-SOFA-001",
				SpecInfo:     "标准款",
				ShipQuantity: 5,
				CostPrice:    1000.00,
			},
		},
	}

	err := s.receiptSvc.CreateReceiptOrder(req, 1)
	s.NoError(err)

	var order models.ReceiptOrder
	s.db.First(&order)

	err = s.receiptSvc.CancelReceiptOrder(order.ID)
	s.NoError(err)

	s.db.First(&order, order.ID)
	s.Equal(models.ReceiptStatusCancelled, order.Status)
}

func (s *ReceiptFlowTestSuite) TestCancelReceiptOrder_AfterReceived() {
	s.seedBaseData()

	req := &svc.CreateReceiptRequest{
		StoreID:      1,
		SupplierID:   1,
		SupplierName: "测试供应商",
		Items: []svc.CreateReceiptItemRequest{
			{
				SKUID:        100,
				ProductName:  "真皮沙发",
				SKUCode:      "SKU-SOFA-001",
				SpecInfo:     "标准款",
				ShipQuantity: 5,
				CostPrice:    1000.00,
			},
		},
	}

	err := s.receiptSvc.CreateReceiptOrder(req, 1)
	s.NoError(err)

	var order models.ReceiptOrder
	s.db.First(&order)

	var receiptItems []models.ReceiptItem
	s.db.Where("receipt_order_id = ?", order.ID).Find(&receiptItems)
	s.Len(receiptItems, 1)
	receiptItemID := receiptItems[0].ID

	err = s.receiptSvc.ApproveReceiptOrder(order.ID, 1)
	s.NoError(err)

	receiveReq := &svc.ReceiveRequest{
		WarehouseID: 1,
		Items: []svc.ReceiveItemRequest{
			{
				ReceiptItemID:   receiptItemID,
				ReceiveQuantity: 5,
			},
		},
	}
	err = s.receiptSvc.ReceiveReceiptOrder(order.ID, receiveReq, 1)
	s.NoError(err)

	err = s.receiptSvc.CancelReceiptOrder(order.ID)
	s.Error(err)
	s.Contains(err.Error(), "已入库回货单不能取消")
}

func TestReceiptFlowSuite(t *testing.T) {
	suite.Run(t, new(ReceiptFlowTestSuite))
}