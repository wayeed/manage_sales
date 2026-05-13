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

// CreatePurchaseOrderRequest 创建采购订单请求
type CreatePurchaseOrderRequest struct {
	StoreID int64 `json:"store_id" example:1`
	SupplierID int64 `json:"supplier_id" example:1`
	SupplierName string `json:"supplier_name" example:"某某供应商"`
	Remark string `json:"remark" example:"紧急采购"`
	Items []CreatePurchaseItemRequest `json:"items" binding:"required,min=1" example:[]`
}

// CreatePurchaseItemRequest 创建采购明细请求
type CreatePurchaseItemRequest struct {
	SKUID int64 `json:"sku_id" binding:"required" example:1`
	ProductName string `json:"product_name" example:"真皮沙发"`
	SKUName string `json:"sku_name" example:"真皮沙发-棕色-三座"`
	PurchasePrice float64 `json:"purchase_price" binding:"required" example:5000.00`
	Quantity int `json:"quantity" binding:"required,min=1" example:10`
}

// ListPurchaseOrderRequest 采购订单列表查询请求
type ListPurchaseOrderRequest struct {
	StoreID int64 `form:"store_id" example:1`
	SupplierID int64 `form:"supplier_id" example:1`
	Status *int8 `form:"status" example:0`
	Keyword string `form:"keyword" example:"PO"`
	Page int `form:"page" example:1`
	PageSize int `form:"page_size" example:10`
}

// PurchaseService 采购服务
type PurchaseService struct {
	db              *gorm.DB
	purchaseRepo    *repository.BaseRepository[models.PurchaseOrder]
	purchaseItemRepo *repository.BaseRepository[models.PurchaseItem]
	supplierRepo    *repository.BaseRepository[models.Supplier]
	inventoryService *InventoryService
}

// NewPurchaseService 创建采购服务实例
func NewPurchaseService(db *gorm.DB, inventoryService *InventoryService) *PurchaseService {
	return &PurchaseService{
		db:               db,
		purchaseRepo:     repository.NewBaseRepository[models.PurchaseOrder](db),
		purchaseItemRepo: repository.NewBaseRepository[models.PurchaseItem](db),
		supplierRepo:     repository.NewBaseRepository[models.Supplier](db),
		inventoryService: inventoryService,
	}
}

// CreateOrder 创建采购订单
func (s *PurchaseService) CreateOrder(req *CreatePurchaseOrderRequest, warehouseID int64, createdBy int64) error {
	// 如果前端没传 supplier_name，根据 supplier_id 自动查询
	if req.SupplierName == "" && req.SupplierID > 0 {
		var supplier models.Supplier
		if err := s.db.First(&supplier, req.SupplierID).Error; err == nil {
			req.SupplierName = supplier.SupplierName
		}
	}

	// 生成采购单号
	purchaseNo := "PO" + appsnow.GenerateOrderNo()

	// 计算总金额和总数量
	var totalAmount decimal.Decimal
	totalQuantity := 0
	for _, item := range req.Items {
		subtotal := decimal.NewFromFloat(item.PurchasePrice).Mul(decimal.NewFromInt(int64(item.Quantity)))
		totalAmount = totalAmount.Add(subtotal)
		totalQuantity += item.Quantity
	}

	// 创建采购订单
	order := &models.PurchaseOrder{
		StoreID:       req.StoreID,
		PurchaseNo:    purchaseNo,
		SupplierID:    &req.SupplierID,
		SupplierName:  req.SupplierName,
		TotalAmount:   totalAmount,
		TotalQuantity: totalQuantity,
		Status:        0, // 待审核
		Remark:        req.Remark,
		CreatedBy:     &createdBy,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// 创建采购明细
		for _, item := range req.Items {
			subtotal := decimal.NewFromFloat(item.PurchasePrice).Mul(decimal.NewFromInt(int64(item.Quantity)))
			purchaseItem := &models.PurchaseItem{
				PurchaseOrderID: order.ID,
				SKUID:           &item.SKUID,
				ProductName:     item.ProductName,
				SKUName:         item.SKUName,
				PurchasePrice:   decimal.NewFromFloat(item.PurchasePrice),
				Quantity:        item.Quantity,
				Subtotal:        subtotal,
			}
			if err := tx.Create(purchaseItem).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "创建采购订单失败"}
	}
	return nil
}

// ApproveOrder 审核采购订单（status 0->1）
func (s *PurchaseService) ApproveOrder(id int64, auditedBy int64) error {
	order, err := s.purchaseRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrOrderNotFound, Message: "采购订单不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if order.Status != 0 {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "订单状态不允许审核"}
	}

	now := time.Now()
	if err := s.db.Model(order).Updates(map[string]interface{}{
		"status":     1,
		"audited_by": auditedBy,
		"audited_at": now,
	}).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "审核采购订单失败"}
	}

	return nil
}

// ConfirmReceipt 确认入库（status 1->2）
func (s *PurchaseService) ConfirmReceipt(id int64, warehouseID int64, createdBy int64) error {
	order, err := s.purchaseRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrOrderNotFound, Message: "采购订单不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if order.Status != 1 {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "订单状态不允许入库"}
	}

	// 查询采购明细
	var items []models.PurchaseItem
	if err := s.db.Where("purchase_order_id = ?", order.ID).Find(&items).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "查询采购明细失败"}
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 遍历采购明细，入库
		for _, item := range items {
			if item.SKUID == nil {
				continue
			}

			// 生成批次号
			batchNo := fmt.Sprintf("PO%dSKU%d", order.ID, *item.SKUID)

			// 调用inventory_service.AddStock增加库存
			purchasePrice := item.PurchasePrice
			totalCost := item.PurchasePrice.Mul(decimal.NewFromInt(int64(item.Quantity)))

			if err := s.inventoryService.AddStock(
				warehouseID, *item.SKUID, item.Quantity,
				purchasePrice, totalCost, batchNo,
				order.ID, order.StoreID, createdBy, 1, // 1=采购入库
			); err != nil {
				return err
			}
		}

		// 更新采购订单状态
		if err := tx.Model(order).Update("status", 2).Error; err != nil {
			return err
		}

		// 更新供应商统计
		if order.SupplierID != nil {
			s.db.Model(&models.Supplier{}).Where("id = ?", *order.SupplierID).Updates(map[string]interface{}{
				"total_purchase_amount":  gorm.Expr("total_purchase_amount + ?", order.TotalAmount),
				"total_purchase_orders":  gorm.Expr("total_purchase_orders + 1"),
				"last_purchase_at":       time.Now(),
			})
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

// CancelOrder 取消采购订单
func (s *PurchaseService) CancelOrder(id int64) error {
	order, err := s.purchaseRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrOrderNotFound, Message: "采购订单不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if order.Status == 2 {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "已入库订单不能取消"}
	}
	if order.Status == 3 {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "订单已取消"}
	}

	if err := s.db.Model(order).Update("status", 3).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "取消采购订单失败"}
	}

	return nil
}

// List 采购订单列表
func (s *PurchaseService) List(req *ListPurchaseOrderRequest) (*PageResult, error) {
	db := s.db.Model(&models.PurchaseOrder{})

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
		db = db.Where("purchase_no LIKE ?", like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询采购订单列表失败"}
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

	var orders []models.PurchaseOrder
	if err := db.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").
		Find(&orders).Error; err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询采购订单列表失败"}
	}

	return &PageResult{
		List:     orders,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetDetail 采购订单详情
func (s *PurchaseService) GetDetail(id int64) (*models.PurchaseOrder, error) {
	var order models.PurchaseOrder
	if err := s.db.Preload("Items.SKU.Product").First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.ErrOrderNotFound, Message: "采购订单不存在"}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}
	return &order, nil
}
