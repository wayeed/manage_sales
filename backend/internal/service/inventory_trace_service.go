package service

import (
	"furniture-commission/internal/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ==================== DTO 定义 ====================

// ForwardTraceResult 正向穿透结果
type ForwardTraceResult struct {
	OrderNo         string        `json:"order_no"`
	CustomerName    string        `json:"customer_name"`
	CustomerPhone   string        `json:"customer_phone"`
	CustomerAddress string        `json:"customer_address"`
	SalesmanName    string        `json:"salesman_name"`
	Items           []ForwardItem `json:"items"`
}

// ForwardItem 正向穿透-商品项
type ForwardItem struct {
	SKUID     int64           `json:"sku_id"`
	SKUName   string          `json:"sku_name"`
	SKUCode   string          `json:"sku_code"`
	Quantity  int             `json:"quantity"`
	UnitCost  decimal.Decimal `json:"unit_cost"`
	TotalCost decimal.Decimal `json:"total_cost"`
	Batch     *BatchInfo      `json:"batch"`
	Purchase  *PurchaseInfo   `json:"purchase"`
	Delivery  *DeliveryInfo   `json:"delivery"`
}

// BatchInfo 批次信息
type BatchInfo struct {
	BatchID           int64           `json:"batch_id"`
	BatchNo           string          `json:"batch_no"`
	PurchasePrice     decimal.Decimal `json:"purchase_price"`
	InitialQuantity   int             `json:"initial_quantity"`
	RemainingQuantity int             `json:"remaining_quantity"`
	WarehouseName     string          `json:"warehouse_name"`
	EntryDate         string          `json:"entry_date"`
}

// PurchaseInfo 采购来源信息
type PurchaseInfo struct {
	PurchaseNo      string          `json:"purchase_no"`
	SupplierName    string          `json:"supplier_name"`
	PurchasePrice   decimal.Decimal `json:"purchase_price"`
	PurchaseQuantity int            `json:"purchase_quantity"`
	ReceiptDate     string          `json:"receipt_date"`
}

// DeliveryInfo 出库信息
type DeliveryInfo struct {
	DeliveryNo   string `json:"delivery_no"`
	DeliveryTime string `json:"delivery_time"`
	Quantity     int    `json:"quantity"`
}

// BackwardTraceResult 反向穿透结果
type BackwardTraceResult struct {
	Batch        *BatchDetailInfo   `json:"batch"`
	Transactions []TransactionInfo `json:"transactions"`
	Summary      *BatchSummary      `json:"summary"`
}

// BatchDetailInfo 批次详细信息
type BatchDetailInfo struct {
	ID                int64           `json:"id"`
	BatchNo           string          `json:"batch_no"`
	SKUID             int64           `json:"sku_id"`
	SKUName           string          `json:"sku_name"`
	SKUCode           string          `json:"sku_code"`
	PurchaseNo        string          `json:"purchase_no"`
	SupplierName      string          `json:"supplier_name"`
	PurchasePrice     decimal.Decimal `json:"purchase_price"`
	InitialQuantity   int             `json:"initial_quantity"`
	RemainingQuantity int             `json:"remaining_quantity"`
	WarehouseName     string          `json:"warehouse_name"`
	EntryDate         string          `json:"entry_date"`
}

// TransactionInfo 库存变动记录
type TransactionInfo struct {
	ID                  int64          `json:"id"`
	TransactionType      int8           `json:"transaction_type"`
	TransactionTypeName string         `json:"transaction_type_name"`
	Quantity            int            `json:"quantity"`
	CreatedAt           string         `json:"created_at"`
	Order               *OrderBrief    `json:"order"`
	Delivery            *DeliveryBrief `json:"delivery"`
}

// OrderBrief 订单简要信息
type OrderBrief struct {
	OrderID      int64  `json:"order_id"`
	OrderNo      string `json:"order_no"`
	CustomerName string `json:"customer_name"`
	SalesmanName string `json:"salesman_name"`
}

// DeliveryBrief 出库简要信息
type DeliveryBrief struct {
	DeliveryNo   string `json:"delivery_no"`
	DeliveryTime string `json:"delivery_time"`
	Status       uint8  `json:"status"`
}

// BatchSummary 批次汇总
type BatchSummary struct {
	TotalLocked    int `json:"total_locked"`
	TotalDelivered int `json:"total_delivered"`
	TotalRemaining int `json:"total_remaining"`
}

// SKUBatchTraceResult SKU批次全景结果
type SKUBatchTraceResult struct {
	SKU            *SKUDetail      `json:"sku"`
	TotalStock     int             `json:"total_stock"`
	AvailableStock int             `json:"available_stock"`
	LockedStock    int             `json:"locked_stock"`
	Batches        []BatchUsedInfo `json:"batches"`
}

// SKUDetail SKU详细信息
type SKUDetail struct {
	ID         int64  `json:"id"`
	SKUName    string `json:"sku_name"`
	SKUCode    string `json:"sku_code"`
	ProductName string `json:"product_name"`
}

// BatchUsedInfo 批次使用信息
type BatchUsedInfo struct {
	BatchInfo    *BatchDetailInfo `json:"batch"`
	UsedByOrders []OrderUsedBrief `json:"used_by_orders"`
}

// OrderUsedBrief 订单使用简要信息
type OrderUsedBrief struct {
	OrderNo      string `json:"order_no"`
	CustomerName string `json:"customer_name"`
	Quantity     int    `json:"quantity"`
	Status       string `json:"status"`
}

// ==================== Service ====================

// InventoryTraceService 库存穿透查询服务
type InventoryTraceService struct {
	db *gorm.DB
}

// NewInventoryTraceService 创建库存穿透查询服务
func NewInventoryTraceService(db *gorm.DB) *InventoryTraceService {
	return &InventoryTraceService{db: db}
}

// ForwardTraceByOrderNo 根据订单号正向穿透：订单 → 源头
func (s *InventoryTraceService) ForwardTraceByOrderNo(orderNo string) (*ForwardTraceResult, error) {
	// 1. 查询订单
	var order models.Order
	if err := s.db.Preload("Salesman").Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return nil, err
	}

	return s.forwardTrace(order)
}

// ForwardTrace 正向穿透：订单 → 源头
func (s *InventoryTraceService) ForwardTrace(orderID int64) (*ForwardTraceResult, error) {
	// 1. 查询订单
	var order models.Order
	if err := s.db.Preload("Salesman").First(&order, orderID).Error; err != nil {
		return nil, err
	}

	return s.forwardTrace(order)
}

// forwardTrace 内部实现
func (s *InventoryTraceService) forwardTrace(order models.Order) (*ForwardTraceResult, error) {
	result := &ForwardTraceResult{
		OrderNo:         order.OrderNo,
		CustomerName:    order.CustomerName,
		CustomerPhone:   order.CustomerPhone,
		CustomerAddress: order.CustomerAddress,
		SalesmanName:    order.Salesman.RealName,
	}

	// 2. 查询订单商品明细
	var items []models.OrderItem
	if err := s.db.Where("order_id = ? AND item_status != 2", order.ID).Find(&items).Error; err != nil {
		return nil, err
	}

	for _, item := range items {
		fi := ForwardItem{
			SKUID:     item.SKUID,
			SKUName:   item.SKUName,
			Quantity:  item.Quantity,
			UnitCost:  item.UnitCost,
			TotalCost: item.TotalCost,
		}

		// 查询SKU编码
		var sku models.ProductSKU
		if err := s.db.First(&sku, item.SKUID).Error; err == nil {
			fi.SKUCode = sku.SKUCode
		}

		// 3. 查询库存批次
		if item.BatchID != nil && *item.BatchID > 0 {
			var batch models.InventoryBatch
			if err := s.db.First(&batch, *item.BatchID).Error; err == nil {
				bi := &BatchInfo{
					BatchID:           batch.ID,
					BatchNo:           batch.BatchNo,
					PurchasePrice:     batch.PurchasePrice,
					InitialQuantity:   batch.InitialQuantity,
					RemainingQuantity: batch.RemainingQuantity,
				}
				if batch.EntryDate != nil {
					bi.EntryDate = batch.EntryDate.Format("2006-01-02")
				}

				// 查询仓库名称
				if batch.WarehouseID != nil {
					var wh models.Warehouse
					if err := s.db.First(&wh, *batch.WarehouseID).Error; err == nil {
						bi.WarehouseName = wh.WarehouseName
					}
				}

				fi.Batch = bi

				// 4. 查询采购单
				if batch.PurchaseOrderID != nil && *batch.PurchaseOrderID > 0 {
					var purchase models.PurchaseOrder
					if err := s.db.First(&purchase, *batch.PurchaseOrderID).Error; err == nil {
						pi := &PurchaseInfo{
							PurchaseNo:    purchase.PurchaseNo,
							SupplierName:  purchase.SupplierName,
							ReceiptDate:   purchase.AuditedAt.Format("2006-01-02"),
						}

						// 查询该SKU在采购单中的明细
						var purchaseItem models.PurchaseItem
						if err := s.db.Where("purchase_order_id = ? AND sku_id = ?", purchase.ID, item.SKUID).First(&purchaseItem).Error; err == nil {
							pi.PurchasePrice = purchaseItem.PurchasePrice
							pi.PurchaseQuantity = purchaseItem.Quantity
						}

						fi.Purchase = pi
					}
				}
			}
		}

		// 5. 查询出库记录
		var deliveryItems []models.DeliveryItem
		s.db.Where("order_item_id = ? AND batch_id = ?", item.ID, item.BatchID).Find(&deliveryItems)
		if len(deliveryItems) > 0 {
			di := deliveryItems[0]
			var delivery models.DeliveryRecord
			if err := s.db.First(&delivery, di.DeliveryID).Error; err == nil {
				fi.Delivery = &DeliveryInfo{
					DeliveryNo:   delivery.DeliveryNo,
					DeliveryTime: delivery.DeliveryTime.Format("2006-01-02 15:04"),
					Quantity:     di.Quantity,
				}
			}
		}

		result.Items = append(result.Items, fi)
	}

	return result, nil
}

// BackwardTrace 反向穿透：批次 → 去向
func (s *InventoryTraceService) BackwardTrace(batchNo string) (*BackwardTraceResult, error) {
	// 1. 查询批次
	var batch models.InventoryBatch
	if err := s.db.Where("batch_no = ?", batchNo).First(&batch).Error; err != nil {
		return nil, err
	}

	// 查询SKU信息
	var sku models.ProductSKU
	skuName, skuCode := "", ""
	if err := s.db.Preload("Product").First(&sku, batch.SKUID).Error; err == nil {
		skuName = sku.SKUName
		skuCode = sku.SKUCode
	}

	// 查询仓库名称
	warehouseName := ""
	if batch.WarehouseID != nil {
		var wh models.Warehouse
		if err := s.db.First(&wh, *batch.WarehouseID).Error; err == nil {
			warehouseName = wh.WarehouseName
		}
	}

	entryDate := ""
	if batch.EntryDate != nil {
		entryDate = batch.EntryDate.Format("2006-01-02")
	}

	// 查询采购单信息
	purchaseNo, supplierName := "", ""
	if batch.PurchaseOrderID != nil && *batch.PurchaseOrderID > 0 {
		var purchase models.PurchaseOrder
		if err := s.db.First(&purchase, *batch.PurchaseOrderID).Error; err == nil {
			purchaseNo = purchase.PurchaseNo
			supplierName = purchase.SupplierName
		}
	}

	batchDetail := &BatchDetailInfo{
		ID:                batch.ID,
		BatchNo:           batch.BatchNo,
		SKUID:             batch.SKUID,
		SKUName:           skuName,
		SKUCode:           skuCode,
		PurchaseNo:        purchaseNo,
		SupplierName:      supplierName,
		PurchasePrice:     batch.PurchasePrice,
		InitialQuantity:   batch.InitialQuantity,
		RemainingQuantity: batch.RemainingQuantity,
		WarehouseName:     warehouseName,
		EntryDate:         entryDate,
	}

	result := &BackwardTraceResult{
		Batch: batchDetail,
	}

	// 2. 查询库存流水（锁定、锁定转出库、退货入库）
	var transactions []models.InventoryTransaction
	s.db.Where("batch_id = ? AND transaction_type IN (9, 10, 11, 12)", batch.ID).
		Order("created_at ASC").Find(&transactions)

	totalLocked := 0
	totalDelivered := 0

	for _, tx := range transactions {
		ti := TransactionInfo{
			ID:                  tx.ID,
			TransactionType:      tx.TransactionType,
			TransactionTypeName: models.GetTransactionTypeName(tx.TransactionType),
			Quantity:            tx.Quantity,
			CreatedAt:           tx.CreatedAt.Format("2006-01-02 15:04"),
		}

		switch tx.TransactionType {
		case models.TransactionTypeLock:
			totalLocked += tx.Quantity
		case models.TransactionTypeLockToOut:
			totalDelivered += tx.Quantity
		}

		// 查询关联订单
		if tx.RelatedOrderID != nil && *tx.RelatedOrderID > 0 {
			var order models.Order
			if err := s.db.Preload("Salesman").First(&order, *tx.RelatedOrderID).Error; err == nil {
				ti.Order = &OrderBrief{
					OrderID:      order.ID,
					OrderNo:      order.OrderNo,
					CustomerName: order.CustomerName,
					SalesmanName: order.Salesman.RealName,
				}
			}
		}

		// 查询出库记录
		if tx.TransactionType == models.TransactionTypeLockToOut {
			var deliveryItems []models.DeliveryItem
			s.db.Where("batch_id = ?", batch.ID).Find(&deliveryItems)
			for _, di := range deliveryItems {
				var delivery models.DeliveryRecord
				if err := s.db.First(&delivery, di.DeliveryID).Error; err == nil {
					ti.Delivery = &DeliveryBrief{
						DeliveryNo:   delivery.DeliveryNo,
						DeliveryTime: delivery.DeliveryTime.Format("2006-01-02 15:04"),
						Status:       delivery.Status,
					}
					break
				}
			}
		}

		result.Transactions = append(result.Transactions, ti)
	}

	result.Summary = &BatchSummary{
		TotalLocked:    totalLocked,
		TotalDelivered: totalDelivered,
		TotalRemaining: batch.RemainingQuantity,
	}

	return result, nil
}

// SKUBatchTrace SKU批次全景
func (s *InventoryTraceService) SKUBatchTrace(skuID int64) (*SKUBatchTraceResult, error) {
	// 1. 查询SKU信息
	var sku models.ProductSKU
	if err := s.db.Preload("Product").First(&sku, skuID).Error; err != nil {
		return nil, err
	}

	productName := ""
	if sku.Product != nil {
		productName = sku.Product.ProductName
	}

	result := &SKUBatchTraceResult{
		SKU: &SKUDetail{
			ID:          sku.ID,
			SKUName:     sku.SKUName,
			SKUCode:     sku.SKUCode,
			ProductName: productName,
		},
	}

	// 2. 查询SKU总库存
	var totalStock, availableStock, lockedStock int
	s.db.Model(&models.WarehouseStock{}).
		Select("COALESCE(SUM(stock_quantity), 0)").
		Where("sku_id = ?", skuID).Scan(&totalStock)
	s.db.Model(&models.WarehouseStock{}).
		Select("COALESCE(SUM(available_quantity), 0)").
		Where("sku_id = ?", skuID).Scan(&availableStock)
	s.db.Model(&models.WarehouseStock{}).
		Select("COALESCE(SUM(locked_quantity), 0)").
		Where("sku_id = ?", skuID).Scan(&lockedStock)

	result.TotalStock = totalStock
	result.AvailableStock = availableStock
	result.LockedStock = lockedStock

	// 3. 查询所有批次
	var batches []models.InventoryBatch
	s.db.Where("sku_id = ? AND status = 1", skuID).
		Order("entry_date ASC").Find(&batches)

	for _, batch := range batches {
		bi := &BatchDetailInfo{
			ID:                batch.ID,
			BatchNo:           batch.BatchNo,
			SKUID:             batch.SKUID,
			SKUName:           sku.SKUName,
			SKUCode:           sku.SKUCode,
			PurchasePrice:     batch.PurchasePrice,
			InitialQuantity:   batch.InitialQuantity,
			RemainingQuantity: batch.RemainingQuantity,
		}

		if batch.EntryDate != nil {
			bi.EntryDate = batch.EntryDate.Format("2006-01-02")
		}

		// 仓库名称
		if batch.WarehouseID != nil {
			var wh models.Warehouse
			if err := s.db.First(&wh, *batch.WarehouseID).Error; err == nil {
				bi.WarehouseName = wh.WarehouseName
			}
		}

		// 采购单信息
		if batch.PurchaseOrderID != nil && *batch.PurchaseOrderID > 0 {
			var purchase models.PurchaseOrder
			if err := s.db.First(&purchase, *batch.PurchaseOrderID).Error; err == nil {
				bi.PurchaseNo = purchase.PurchaseNo
				bi.SupplierName = purchase.SupplierName
			}
		}

		bui := BatchUsedInfo{
			BatchInfo: bi,
		}

		// 4. 查询使用该批次的订单
		var orderItems []models.OrderItem
		s.db.Where("batch_id = ? AND item_status != 2", batch.ID).Find(&orderItems)
		for _, oi := range orderItems {
			var order models.Order
			if err := s.db.First(&order, oi.OrderID).Error; err == nil {
				status := "已锁定"
				// 检查是否已出库
				var deliveryItems []models.DeliveryItem
				s.db.Where("order_item_id = ? AND batch_id = ?", oi.ID, batch.ID).Find(&deliveryItems)
				if len(deliveryItems) > 0 {
					status = "已出库"
				}

				bui.UsedByOrders = append(bui.UsedByOrders, OrderUsedBrief{
					OrderNo:      order.OrderNo,
					CustomerName: order.CustomerName,
					Quantity:     oi.Quantity,
					Status:       status,
				})
			}
		}

		result.Batches = append(result.Batches, bui)
	}

	return result, nil
}
