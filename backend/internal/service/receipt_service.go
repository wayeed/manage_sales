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

// CreateReceiptRequest 创建回货单请求
type CreateReceiptRequest struct {
	ReceiptNo    string                     `json:"receipt_no"`
	StoreID      int64                      `json:"store_id"`
	SupplierID   int64                      `json:"supplier_id"`
	SupplierName string                     `json:"supplier_name"`
	Remark       string                     `json:"remark"`
	Items        []CreateReceiptItemRequest `json:"items"`
}

// CreateReceiptItemRequest 创建回货明细请求
type CreateReceiptItemRequest struct {
	PurchaseItemID *int64  `json:"purchase_item_id"`
	SKUID          int64   `json:"sku_id"`
	ProductName    string  `json:"product_name"`
	SKUName        string  `json:"sku_name"`
	SKUCode        string  `json:"sku_code"`
	BrandStyle     string  `json:"brand_style"`
	ShipQuantity   int     `json:"ship_quantity"`
	CostPrice      float64 `json:"cost_price"`
	Remark         string  `json:"remark"`
}

// ReceiveRequest 入库请求
type ReceiveRequest struct {
	WarehouseID int64                `json:"warehouse_id"`
	Items       []ReceiveItemRequest `json:"items"`
	Remark      string               `json:"remark"`
}

// ReceiveItemRequest 入库明细请求
type ReceiveItemRequest struct {
	ReceiptItemID   int64 `json:"receipt_item_id"`
	ReceiveQuantity int   `json:"receive_quantity"`
}

// ListReceiptRequest 回货单列表查询请求
type ListReceiptRequest struct {
	StoreID    int64  `form:"store_id"`
	SupplierID int64  `form:"supplier_id"`
	Status     *int8  `form:"status"`
	Keyword    string `form:"keyword"`
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
}

// ReceiptService 回货单服务
type ReceiptService struct {
	db               *gorm.DB
	receiptRepo      *repository.BaseRepository[models.ReceiptOrder]
	receiptItemRepo  *repository.BaseRepository[models.ReceiptItem]
	purchaseItemRepo *repository.BaseRepository[models.PurchaseItem]
	inventoryService *InventoryService
}

// NewReceiptService 创建回货单服务实例
func NewReceiptService(db *gorm.DB, inventoryService *InventoryService) *ReceiptService {
	return &ReceiptService{
		db:               db,
		receiptRepo:      repository.NewBaseRepository[models.ReceiptOrder](db),
		receiptItemRepo:  repository.NewBaseRepository[models.ReceiptItem](db),
		purchaseItemRepo: repository.NewBaseRepository[models.PurchaseItem](db),
		inventoryService: inventoryService,
	}
}

// CreateReceiptOrder 创建回货单
func (s *ReceiptService) CreateReceiptOrder(req *CreateReceiptRequest, createdBy int64) error {
	// 验证回货单号
	if req.ReceiptNo == "" {
		return &AppError{Code: apperrors.ErrInvalidParam, Message: "回货单号不能为空"}
	}

	// 如果没传供应商名称，根据ID查询
	if req.SupplierName == "" && req.SupplierID > 0 {
		var supplier models.Supplier
		if err := s.db.First(&supplier, req.SupplierID).Error; err == nil {
			req.SupplierName = supplier.SupplierName
		}
	}

	// 使用前端传递的回货单号
	receiptNo := req.ReceiptNo

	// 计算总金额和总数量
	var totalAmount decimal.Decimal
	totalQuantity := 0
	for _, item := range req.Items {
		subtotal := decimal.NewFromFloat(item.CostPrice).Mul(decimal.NewFromInt(int64(item.ShipQuantity)))
		totalAmount = totalAmount.Add(subtotal)
		totalQuantity += item.ShipQuantity
	}

	// 创建回货单
	order := &models.ReceiptOrder{
		StoreID:       req.StoreID,
		ReceiptNo:     receiptNo,
		SupplierID:    req.SupplierID,
		SupplierName:  req.SupplierName,
		Status:        models.ReceiptStatusPending,
		TotalAmount:   totalAmount,
		TotalQuantity: totalQuantity,
		Remark:        req.Remark,
		CreatedBy:     &createdBy,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// 创建回货明细
		for _, item := range req.Items {
			receiptItem := &models.ReceiptItem{
				ReceiptOrderID:  order.ID,
				PurchaseItemID:  item.PurchaseItemID,
				SKUID:           item.SKUID,
				ProductName:     item.ProductName,
				SKUName:         item.SKUName,
				SKUCode:         item.SKUCode,
				BrandStyle:      item.BrandStyle,
				ShipQuantity:    item.ShipQuantity,
				ReceiveQuantity: 0,
				CostPrice:       decimal.NewFromFloat(item.CostPrice),
				Remark:          item.Remark,
			}
			if err := tx.Create(receiptItem).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "创建回货单失败"}
	}
	return nil
}

// ApproveReceiptOrder 审核回货单
func (s *ReceiptService) ApproveReceiptOrder(id int64, auditedBy int64) error {
	order, err := s.receiptRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrOrderNotFound, Message: "回货单不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if order.Status != models.ReceiptStatusPending {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "当前状态不允许审核"}
	}

	now := time.Now()
	if err := s.db.Model(order).Updates(map[string]interface{}{
		"status":     models.ReceiptStatusApproved,
		"audited_by": auditedBy,
		"audited_at": now,
	}).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "审核回货单失败"}
	}

	return nil
}

// ReceiveReceiptOrder 入库操作
func (s *ReceiptService) ReceiveReceiptOrder(id int64, req *ReceiveRequest, createdBy int64) error {
	order, err := s.receiptRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrOrderNotFound, Message: "回货单不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if order.Status != models.ReceiptStatusApproved {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "当前状态不允许入库"}
	}

	// 查询回货明细
	var items []models.ReceiptItem
	if err := s.db.Where("receipt_order_id = ?", order.ID).Find(&items).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "查询回货明细失败"}
	}

	// 构建ID到item的映射
	itemMap := make(map[int64]*models.ReceiptItem)
	for i := range items {
		itemMap[items[i].ID] = &items[i]
	}

	// 生成唯一的批次号后缀（支持分批入库）
	batchSuffix := time.Now().Format("150405") + fmt.Sprintf("%d", time.Now().UnixNano()%1000000)

	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, receiveItem := range req.Items {
			item, ok := itemMap[receiveItem.ReceiptItemID]
			if !ok {
				return &AppError{Code: apperrors.ErrInvalidParam, Message: fmt.Sprintf("回货明细ID %d 不存在", receiveItem.ReceiptItemID)}
			}

			// 检查是否超过发货数量
			if item.ReceiveQuantity+receiveItem.ReceiveQuantity > item.ShipQuantity {
				return &AppError{Code: apperrors.ErrInvalidParam, Message: "实际收货数量不能超过发货数量"}
			}

			// 更新回货明细实际收货数量
			newReceiveQty := item.ReceiveQuantity + receiveItem.ReceiveQuantity
			if err := tx.Model(item).Update("receive_quantity", newReceiveQty).Error; err != nil {
				return err
			}
			// 同步更新 Go 变量，确保后续判断使用正确的值
			item.ReceiveQuantity = newReceiveQty

			// 更新采购明细已入库数量（如果关联了采购明细）
			if item.PurchaseItemID != nil {
				var purchaseItem models.PurchaseItem
				if err := tx.First(&purchaseItem, *item.PurchaseItemID).Error; err == nil {
					newReceivedQty := purchaseItem.ReceivedQuantity + receiveItem.ReceiveQuantity
					if err := tx.Model(&purchaseItem).Update("received_quantity", newReceivedQty).Error; err != nil {
						return err
					}
				}
			}

			// 调用库存服务增加库存（传入事务连接）
			// 使用时间戳后缀确保批次号唯一，支持分批入库
			batchNo := fmt.Sprintf("RC%dSKU%d-%s", order.ID, item.SKUID, batchSuffix)
			totalCost := item.CostPrice.Mul(decimal.NewFromInt(int64(receiveItem.ReceiveQuantity)))

			if err := s.inventoryService.AddStockWithTx(
				tx, req.WarehouseID, item.SKUID, receiveItem.ReceiveQuantity,
				item.CostPrice, totalCost, batchNo,
				order.ID, order.StoreID, createdBy, 1, // 1=采购入库
			); err != nil {
				return err
			}

			// 查询刚创建的批次
			var batch models.InventoryBatch
			if err := tx.Where("batch_no = ?", batchNo).First(&batch).Error; err != nil {
				continue
			}

			// 自动分配库存到缺货排队订单
			if err := s.inventoryService.AllocateStockToQueue(tx, req.WarehouseID, item.SKUID, batch.ID, batch.PurchasePrice, batch.RemainingQuantity, order.StoreID); err != nil {
				// 分配失败不影响入库，仅记录日志
				fmt.Printf("[WARN] 回货单自动分配缺货订单失败 receipt_order_id=%d, sku_id=%d: %v\n", order.ID, item.SKUID, err)
			}
		}

		// 检查是否所有商品都已入库
		allReceived := true
		for _, item := range items {
			if item.ReceiveQuantity < item.ShipQuantity {
				allReceived = false
				break
			}
		}

		// 如果全部入库，更新回货单状态
		if allReceived {
			if err := tx.Model(order).Update("status", models.ReceiptStatusReceived).Error; err != nil {
				return err
			}

			// 更新关联的采购单状态（如果所有商品都已入库）
			s.updatePurchaseOrderStatus(tx, order.ID)
		}

		return nil
	})

	if err != nil {
		if appErr, ok := err.(*AppError); ok {
			return appErr
		}
		return &AppError{Code: apperrors.InternalError, Message: "入库失败"}
	}

	return nil
}

// updatePurchaseOrderStatus 更新采购单状态
func (s *ReceiptService) updatePurchaseOrderStatus(tx *gorm.DB, receiptOrderID int64) {
	// 查询回货单关联的采购明细ID
	var receiptItems []models.ReceiptItem
	if err := tx.Where("receipt_order_id = ? AND purchase_item_id IS NOT NULL", receiptOrderID).Find(&receiptItems).Error; err != nil {
		return
	}

	// 获取所有关联的采购订单ID
	purchaseOrderIDs := make(map[int64]bool)
	for _, item := range receiptItems {
		if item.PurchaseItemID != nil {
			var purchaseItem models.PurchaseItem
			if err := tx.First(&purchaseItem, *item.PurchaseItemID).Error; err == nil && purchaseItem.PurchaseOrderID > 0 {
				purchaseOrderIDs[purchaseItem.PurchaseOrderID] = true
			}
		}
	}

	// 检查每个采购单是否所有商品都已入库
	for poID := range purchaseOrderIDs {
		type qtyResult struct {
			TotalQty    int
			ReceivedQty int
		}
		var result qtyResult
		if err := tx.Model(&models.PurchaseItem{}).Where("purchase_order_id = ?", poID).
			Select("SUM(quantity) as total_qty, SUM(received_quantity) as received_qty").
			Scan(&result).Error; err != nil {
			continue
		}

		if result.TotalQty > 0 && result.TotalQty == result.ReceivedQty {
			// 所有商品都已入库，更新采购单状态
			tx.Model(&models.PurchaseOrder{}).Where("id = ?", poID).Update("status", 2) // 2=已入库
		}
	}
}

// GetReceiptOrderDetail 获取回货单详情
func (s *ReceiptService) GetReceiptOrderDetail(id int64) (*models.ReceiptOrder, error) {
	var order models.ReceiptOrder
	if err := s.db.Preload("Items").First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.ErrOrderNotFound, Message: "回货单不存在"}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}
	return &order, nil
}

// ListReceiptOrders 查询回货单列表
func (s *ReceiptService) ListReceiptOrders(req *ListReceiptRequest) (*PageResult, error) {
	db := s.db.Model(&models.ReceiptOrder{})

	if req.StoreID > 0 {
		db = db.Where("store_id = ?", req.StoreID)
	}
	if req.SupplierID > 0 {
		db = db.Where("supplier_id = ?", req.SupplierID)
	}
	if req.Status != nil {
		db = db.Where("status = ?", *req.Status)
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		db = db.Where("receipt_no LIKE ? OR supplier_name LIKE ?", like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询回货单列表失败"}
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

	var orders []models.ReceiptOrder
	if err := db.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").
		Find(&orders).Error; err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询回货单列表失败"}
	}

	return &PageResult{
		List:     orders,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// CancelReceiptOrder 取消回货单
func (s *ReceiptService) CancelReceiptOrder(id int64) error {
	order, err := s.receiptRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrOrderNotFound, Message: "回货单不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if order.Status == models.ReceiptStatusReceived {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "已入库回货单不能取消"}
	}

	if err := s.db.Model(order).Update("status", models.ReceiptStatusCancelled).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "取消回货单失败"}
	}

	return nil
}
