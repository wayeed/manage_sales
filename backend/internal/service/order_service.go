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

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	StoreID int64 `json:"store_id" binding:"required" example:1`
	SalesmanID int64 `json:"salesman_id" binding:"required" example:1`
	CustomerName string `json:"customer_name" binding:"required" example:"李四"`
	CustomerPhone string `json:"customer_phone" binding:"required" example:"13900139000"`
	CustomerAddress string `json:"customer_address" example:"北京市朝阳区某某路123号"`
	Source int8 `json:"source" example:1`
	IsPeerOrder int8 `json:"is_peer_order" example:0`
	PeerID *int64 `json:"peer_id" example:1`
	IsSpecialApproved int8 `json:"is_special_approved" example:0`
	ApprovalRemark string `json:"approval_remark" example:""`
	Remark string `json:"remark" example:"客户要求尽快配送"`
	OrderDate string `json:"order_date" example:"2024-01-15"`
	Items []CreateOrderItemRequest `json:"items" binding:"required,min=1" example:[]`
	Gifts []CreateOrderGiftRequest `json:"gifts" example:[]`
	IsDraft         int8                     `json:"is_draft" example:"0"` // 0-正式订单, 1-草稿
}

// CreateOrderItemRequest 创建订单商品明细请求
type CreateOrderItemRequest struct {
	SKUID int64 `json:"sku_id" example:1`
	ProductName string `json:"product_name" binding:"required" example:"真皮沙发"`
	SKUName string `json:"sku_name" binding:"required" example:"真皮沙发-棕色-三座"`
	CategoryID *int64 `json:"category_id" example:1`
	Quantity int `json:"quantity" binding:"required,min=1" example:1`
	ListPrice float64 `json:"list_price" binding:"required" example:8999.00`
	SalePrice float64 `json:"sale_price" binding:"required" example:8500.00`
}

// CreateOrderGiftRequest 创建订单赠品请求
type CreateOrderGiftRequest struct {
	GiftID int64 `json:"gift_id" binding:"required" example:1`
	GiftName string `json:"gift_name" binding:"required" example:"抱枕"`
	CostPrice float64 `json:"cost_price" binding:"required" example:50.00`
	Quantity int `json:"quantity" binding:"required,min=1" example:2`
}

// ListOrderRequest 订单列表查询请求
type ListOrderRequest struct {
	StoreID string `form:"store_id" example:"1"`
	SalesmanID string `form:"salesman_id" example:"1"`
	OrderStatus string `form:"order_status" example:"0"`
	PaymentStatus string `form:"payment_status" example:"0"`
	OrderType string `form:"order_type" example:"1"`
	StartDate string `form:"start_date" example:"2024-01-01"`
	EndDate string `form:"end_date" example:"2024-12-31"`
	Keyword string `form:"keyword" example:"李四"`
	Page int `form:"page" example:1`
	PageSize int `form:"page_size" example:10`
}

// OrderDetail 订单详情
type OrderDetail struct {
	Order *models.Order `json:"order" example:{}`
	Items []models.OrderItem `json:"items" example:[]`
	Gifts []models.OrderGift `json:"gifts" example:[]`
	Payments []models.Payment `json:"payments" example:[]`
}

// OrderService 订单核心服务
type OrderService struct {
	db           *gorm.DB
	orderRepo    *repository.OrderRepository
	paymentRepo  *repository.PaymentRepository
	customerRepo *repository.CustomerRepository
	peerRepo     *repository.PeerRepository
	inventorySvc *InventoryService
}

// NewOrderService 创建订单服务实例
func NewOrderService(db *gorm.DB, orderRepo *repository.OrderRepository, paymentRepo *repository.PaymentRepository, customerRepo *repository.CustomerRepository, peerRepo *repository.PeerRepository, inventorySvc *InventoryService) *OrderService {
	return &OrderService{
		db:           db,
		orderRepo:    orderRepo,
		paymentRepo:  paymentRepo,
		customerRepo: customerRepo,
		peerRepo:     peerRepo,
		inventorySvc: inventorySvc,
	}
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(req *CreateOrderRequest, createdBy int64) (*models.Order, error) {
	// 草稿订单跳过库存验证
	isDraft := req.IsDraft == 1

	// 1. 验证商品SKU存在且上架
	skuSet := make(map[int64]struct{})
	for _, item := range req.Items {
		skuSet[item.SKUID] = struct{}{}
	}
	skuIDs := make([]int64, 0, len(skuSet))
	for id := range skuSet {
		skuIDs = append(skuIDs, id)
	}

	var skus []models.ProductSKU
	if err := s.db.Where("id IN ? AND status = 1", skuIDs).Find(&skus).Error; err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询商品SKU失败"}
	}
	if len(skus) != len(skuIDs) {
		return nil, &AppError{Code: apperrors.BadRequest, Message: "部分商品SKU不存在或已下架"}
	}

	// 2. 查询默认仓库库存（非草稿）
	var defaultWarehouse models.Warehouse
	if !isDraft {
		if err := s.db.Where("store_id = ? AND warehouse_type = 1 AND status = 1", req.StoreID).First(&defaultWarehouse).Error; err != nil {
			return nil, &AppError{Code: apperrors.InternalError, Message: "未找到默认仓库"}
		}

		// 3. 验证库存充足（非草稿）
		for _, item := range req.Items {
			stock, err := s.inventorySvc.GetStock(defaultWarehouse.ID, item.SKUID)
			if err != nil {
				return nil, err
			}
			if stock == nil || stock.AvailableQuantity < item.Quantity {
				return nil, &AppError{Code: apperrors.ErrInsufficientStock, Message: fmt.Sprintf("商品[%s]库存不足", item.SKUName)}
			}
		}
	}

	// 4. 使用事务创建订单
	var order *models.Order
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// a. 锁定库存（非草稿）
		if !isDraft {
			for _, item := range req.Items {
				if err := s.inventorySvc.LockStock(defaultWarehouse.ID, item.SKUID, item.Quantity); err != nil {
					return err
				}
			}
		}

		// b. 计算订单金额
		totalListPrice := decimal.Zero
		totalSalePrice := decimal.Zero
		totalQuantity := 0
		skuCount := len(req.Items)
		categorySet := make(map[int64]bool)

		for _, item := range req.Items {
			lp := decimal.NewFromFloat(item.ListPrice)
			sp := decimal.NewFromFloat(item.SalePrice)
			totalListPrice = totalListPrice.Add(lp.Mul(decimal.NewFromInt(int64(item.Quantity))))
			totalSalePrice = totalSalePrice.Add(sp.Mul(decimal.NewFromInt(int64(item.Quantity))))
			totalQuantity += item.Quantity
			if item.CategoryID != nil {
				categorySet[*item.CategoryID] = true
			}
		}

		categoryCount := len(categorySet)

		// 多品订单折扣（3个及以上品类95折）
		discountAmount := decimal.Zero
		if categoryCount >= 3 {
			discountAmount = totalSalePrice.Mul(decimal.NewFromFloat(0.05))
		}
		finalAmount := totalSalePrice.Sub(discountAmount)

		// c. 创建时total_cost=0, actual_profit=0（审核通过时计算）
		totalCost := decimal.Zero
		actualProfit := decimal.Zero

		// d. 计算gift_cost
		giftCost := decimal.Zero
		for _, gift := range req.Gifts {
			gc := decimal.NewFromFloat(gift.CostPrice).Mul(decimal.NewFromInt(int64(gift.Quantity)))
			giftCost = giftCost.Add(gc)
		}

		// e. 自动判断订单类型
		orderType := int8(1) // 默认单品
		if req.IsPeerOrder == 1 {
			if categoryCount == 1 {
				orderType = 4 // 同行单品
			} else {
				orderType = 5 // 同行多品
			}
			if req.IsSpecialApproved == 1 {
				orderType = 6 // 同行特批
			}
		} else if req.IsSpecialApproved == 1 {
			orderType = 3 // 特殊审批
		} else if categoryCount > 1 {
			orderType = 2 // 多品
		}

		// f. 生成订单号
		orderNo := appsnow.GenerateOrderNo()

		// g. 查找或创建客户
		customer, err := s.customerRepo.FindOrCreate(req.StoreID, req.CustomerName, req.CustomerPhone)
		if err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "查找或创建客户失败"}
		}

		// h. 解析订单日期
		var orderDate *time.Time
		if req.OrderDate != "" {
			od, err := time.Parse("2006-01-02", req.OrderDate)
			if err == nil {
				orderDate = &od
			}
		}
		if orderDate == nil {
			now := time.Now()
			orderDate = &now
		}

		// 创建订单
		order = &models.Order{
			StoreID:          req.StoreID,
			OrderNo:          orderNo,
			SalesmanID:       req.SalesmanID,
			CustomerID:       &customer.ID,
			CustomerName:     req.CustomerName,
			CustomerPhone:    req.CustomerPhone,
			CustomerAddress:  req.CustomerAddress,
			Source:           req.Source,
			OrderType:        orderType,
			OrderStatus:      0, // 待审批
			PaymentStatus:    0, // 未回款
			IsDraft:          req.IsDraft,
			TotalListPrice:   totalListPrice,
			TotalSalePrice:   totalSalePrice,
			DiscountAmount:   discountAmount,
			FinalAmount:      finalAmount,
			TotalCost:        totalCost,
			GiftCost:         giftCost,
			ActualProfit:     actualProfit,
			CategoryCount:    categoryCount,
			SKUCount:         skuCount,
			TotalQuantity:    totalQuantity,
			PaidAmount:       decimal.Zero,
			RemainingAmount:  finalAmount,
			IsPeerOrder:      req.IsPeerOrder,
			PeerID:           req.PeerID,
			IsSpecialApproved: req.IsSpecialApproved,
			ApprovalRemark:   req.ApprovalRemark,
			Remark:           req.Remark,
			OrderDate:        orderDate,
			CreatedBy:        &createdBy,
		}

		if err := tx.Create(order).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "创建订单失败"}
		}

		// 创建订单商品明细
		for _, item := range req.Items {
			listPrice := decimal.NewFromFloat(item.ListPrice)
			salePrice := decimal.NewFromFloat(item.SalePrice)
			// 计算折扣率: 销售价/挂牌价，保留4位小数（0~1之间，如0.85=85折）
			discountRate := decimal.NewFromInt(1)
			if listPrice.GreaterThan(decimal.NewFromInt(0)) {
				discountRate = salePrice.Div(listPrice).Round(4)
			}
			orderItem := &models.OrderItem{
				OrderID:     order.ID,
				SKUID:       item.SKUID,
				ProductName: item.ProductName,
				SKUName:     item.SKUName,
				CategoryID:  item.CategoryID,
				Quantity:    item.Quantity,
				ListPrice:   listPrice,
				SalePrice:   salePrice,
				DiscountRate: discountRate,
			}
			if err := tx.Create(orderItem).Error; err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "创建订单明细失败"}
			}
		}

		// 创建订单赠品
		for _, gift := range req.Gifts {
			gc := decimal.NewFromFloat(gift.CostPrice).Mul(decimal.NewFromInt(int64(gift.Quantity)))
			orderGift := &models.OrderGift{
				OrderID:   order.ID,
				GiftID:    gift.GiftID,
				GiftName:  gift.GiftName,
				CostPrice: decimal.NewFromFloat(gift.CostPrice),
				Quantity:  gift.Quantity,
				TotalCost: gc,
			}
			if err := tx.Create(orderGift).Error; err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "创建订单赠品失败"}
			}
		}

		// i. 记录inventory_transactions（类型=锁定）
		for _, item := range req.Items {
			stock, err := s.inventorySvc.GetStock(defaultWarehouse.ID, item.SKUID)
			if err != nil {
				return err
			}
			if stock == nil {
				continue
			}
			invTx := &models.InventoryTransaction{
				StoreID:         req.StoreID,
				WarehouseID:     &defaultWarehouse.ID,
				TransactionType: 3, // 锁定
				BizType:         1, // 商品
				BizID:           &item.SKUID,
				RelatedOrderID:  &order.ID,
				Quantity:        item.Quantity,
				BeforeStock:     stock.AvailableQuantity + item.Quantity,
				AfterStock:      stock.AvailableQuantity,
				CreatedBy:       &createdBy,
				CreatedAt:       time.Now(),
			}
			if err := tx.Create(invTx).Error; err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "记录库存流水失败"}
			}
		}

		return nil
	})

	if err != nil {
		if appErr, ok := err.(*AppError); ok {
			return nil, appErr
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "创建订单失败"}
	}

	return order, nil
}

// ApproveOrder 审核订单
func (s *OrderService) ApproveOrder(orderID int64, approvedBy int64, approved bool, remark string) error {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrOrderNotFound, Message: "订单不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	// 验证订单状态=0(待审批)
	if order.OrderStatus != 0 {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "订单状态不允许审核"}
	}

	if !approved {
		// 审核驳回
		return s.rejectOrder(order, approvedBy, remark)
	}

	// 审核通过
	return s.approveOrder(order, approvedBy, remark)
}

// approveOrder 审核通过
func (s *OrderService) approveOrder(order *models.Order, approvedBy int64, remark string) error {
	// 查询默认仓库
	var defaultWarehouse models.Warehouse
	if err := s.db.Where("store_id = ? AND warehouse_type = 1 AND status = 1", order.StoreID).First(&defaultWarehouse).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "未找到默认仓库"}
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		// a. 调用inventorySvc.DeductStock实际扣减库存（FIFO）
		totalCost := decimal.Zero
		for i := range order.Items {
			item := &order.Items[i]
			costDetails, err := s.inventorySvc.DeductStock(
				defaultWarehouse.ID, item.SKUID, item.Quantity,
				order.StoreID, order.ID, approvedBy, 2, // 2=销售出库
			)
			if err != nil {
				return err
			}

			// b. 用返回的成本明细更新order_items的unit_cost和total_cost
			itemCost := decimal.Zero
			for _, detail := range costDetails {
				itemCost = itemCost.Add(detail.TotalCost)
			}
			unitCost := decimal.Zero
			if item.Quantity > 0 {
				unitCost = itemCost.Div(decimal.NewFromInt(int64(item.Quantity)))
			}

			if err := tx.Model(&models.OrderItem{}).Where("id = ?", item.ID).Updates(map[string]interface{}{
				"unit_cost":  unitCost,
				"total_cost": itemCost,
			}).Error; err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "更新订单明细成本失败"}
			}

			totalCost = totalCost.Add(itemCost)
		}

		// c. 计算订单总成本
		// d. 计算实际利润 actual_profit = final_amount - total_cost - gift_cost
		actualProfit := order.FinalAmount.Sub(totalCost).Sub(order.GiftCost)

		// e. 更新订单状态
		if err := tx.Model(order).Updates(map[string]interface{}{
			"order_status":  1, // 已生效
			"total_cost":    totalCost,
			"actual_profit": actualProfit,
			"approved_by":   approvedBy,
			"approved_at":   now,
		}).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "更新订单状态失败"}
		}

		// g. 更新客户累计统计
		if order.CustomerID != nil {
			if err := tx.Model(&models.Customer{}).Where("id = ?", *order.CustomerID).Updates(map[string]interface{}{
				"total_orders":  gorm.Expr("total_orders + 1"),
				"total_amount":  gorm.Expr("total_amount + ?", order.FinalAmount),
				"total_profit":  gorm.Expr("total_profit + ?", actualProfit),
				"last_order_at": now,
			}).Error; err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "更新客户统计失败"}
			}
		}

		// h. 如果是同行订单，更新同行累计统计
		if order.IsPeerOrder == 1 && order.PeerID != nil {
			if err := tx.Model(&models.Peer{}).Where("id = ?", *order.PeerID).Updates(map[string]interface{}{
				"total_orders":  gorm.Expr("total_orders + 1"),
				"total_amount":  gorm.Expr("total_amount + ?", order.FinalAmount),
				"total_profit":  gorm.Expr("total_profit + ?", actualProfit),
				"last_order_at": now,
			}).Error; err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "更新同行统计失败"}
			}
		}

		return nil
	})

	if err != nil {
		if appErr, ok := err.(*AppError); ok {
			return appErr
		}
		return &AppError{Code: apperrors.InternalError, Message: "审核订单失败"}
	}

	return nil
}

// rejectOrder 审核驳回
func (s *OrderService) rejectOrder(order *models.Order, approvedBy int64, remark string) error {
	// 查询默认仓库
	var defaultWarehouse models.Warehouse
	if err := s.db.Where("store_id = ? AND warehouse_type = 1 AND status = 1", order.StoreID).First(&defaultWarehouse).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "未找到默认仓库"}
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		// 更新订单状态为已驳回
		if err := tx.Model(order).Updates(map[string]interface{}{
			"order_status":  2, // 已驳回
			"approved_by":   approvedBy,
			"approved_at":   now,
			"approval_remark": remark,
		}).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "更新订单状态失败"}
		}

		// 释放锁定库存（在事务内直接操作）
		for _, item := range order.Items {
			if err := s.unlockStockInTx(tx, defaultWarehouse.ID, item.SKUID, item.Quantity); err != nil {
				return err
			}
		}

		// 记录inventory_transactions（类型=释放锁定）
		for _, item := range order.Items {
			var stock models.WarehouseStock
			if err := tx.Where("warehouse_id = ? AND sku_id = ?", defaultWarehouse.ID, item.SKUID).First(&stock).Error; err != nil {
				continue
			}
			invTx := &models.InventoryTransaction{
				StoreID:         order.StoreID,
				WarehouseID:     &defaultWarehouse.ID,
				TransactionType: 4, // 释放锁定
				BizType:         1, // 商品
				BizID:           &item.SKUID,
				RelatedOrderID:  &order.ID,
				Quantity:        item.Quantity,
				BeforeStock:     stock.AvailableQuantity - item.Quantity,
				AfterStock:      stock.AvailableQuantity,
				CreatedBy:       &approvedBy,
				CreatedAt:       now,
			}
			if err := tx.Create(invTx).Error; err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "记录库存流水失败"}
			}
		}

		return nil
	})

	if err != nil {
		if appErr, ok := err.(*AppError); ok {
			return appErr
		}
		return &AppError{Code: apperrors.InternalError, Message: "驳回订单失败"}
	}

	return nil
}

// unlockStockInTx 在事务内释放锁定库存
func (s *OrderService) unlockStockInTx(tx *gorm.DB, warehouseID, skuID int64, quantity int) error {
	var stock models.WarehouseStock
	if err := tx.Where("warehouse_id = ? AND sku_id = ?", warehouseID, skuID).First(&stock).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrInsufficientStock, Message: "库存记录不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if stock.LockedQuantity < quantity {
		return &AppError{Code: apperrors.ErrInsufficientStock, Message: "锁定库存不足"}
	}

	result := tx.Model(&models.WarehouseStock{}).
		Where("warehouse_id = ? AND sku_id = ? AND version = ?", warehouseID, skuID, stock.Version).
		Updates(map[string]interface{}{
			"locked_quantity":    gorm.Expr("locked_quantity - ?", quantity),
			"available_quantity": gorm.Expr("available_quantity + ?", quantity),
			"version":            gorm.Expr("version + 1"),
		})
	if result.Error != nil {
		return &AppError{Code: apperrors.InternalError, Message: "释放库存失败"}
	}
	if result.RowsAffected == 0 {
		return &AppError{Code: apperrors.ErrInsufficientStock, Message: "库存已被修改，请重试"}
	}

	return nil
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(orderID int64) error {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrOrderNotFound, Message: "订单不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	// 验证订单状态=0(待审批)
	if order.OrderStatus != 0 {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "只有待审批的订单才能取消"}
	}

	// 查询默认仓库
	var defaultWarehouse models.Warehouse
	if err := s.db.Where("store_id = ? AND warehouse_type = 1 AND status = 1", order.StoreID).First(&defaultWarehouse).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "未找到默认仓库"}
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 释放锁定库存（在事务内直接操作）
		for _, item := range order.Items {
			if err := s.unlockStockInTx(tx, defaultWarehouse.ID, item.SKUID, item.Quantity); err != nil {
				return err
			}
		}

		// 更新订单状态为已取消
		if err := tx.Model(order).Update("order_status", 3).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "更新订单状态失败"}
		}

		return nil
	})

	if err != nil {
		if appErr, ok := err.(*AppError); ok {
			return appErr
		}
		return &AppError{Code: apperrors.InternalError, Message: "取消订单失败"}
	}

	return nil
}

// ReturnOrder 退货处理
func (s *OrderService) ReturnOrder(orderID int64, returnAmount float64, returnProfit float64, remark string) error {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrOrderNotFound, Message: "订单不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	// 验证订单状态=1(已生效)
	if order.OrderStatus != 1 {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "只有已生效的订单才能退货"}
	}

	ra := decimal.NewFromFloat(returnAmount)
	rp := decimal.NewFromFloat(returnProfit)

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 更新退货信息
		if err := tx.Model(order).Updates(map[string]interface{}{
			"is_returned":    1,
			"return_amount":  ra,
			"return_profit":  rp,
			"order_status":   4, // 已退货
		}).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "更新退货信息失败"}
		}

		// 更新客户累计统计（冲减）
		if order.CustomerID != nil {
			if err := tx.Model(&models.Customer{}).Where("id = ?", *order.CustomerID).Updates(map[string]interface{}{
				"total_orders":  gorm.Expr("total_orders - 1"),
				"total_amount":  gorm.Expr("total_amount - ?", ra),
				"total_profit":  gorm.Expr("total_profit - ?", rp),
			}).Error; err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "更新客户统计失败"}
			}
		}

		return nil
	})

	if err != nil {
		if appErr, ok := err.(*AppError); ok {
			return appErr
		}
		return &AppError{Code: apperrors.InternalError, Message: "退货处理失败"}
	}

	return nil
}

// List 订单列表
func (s *OrderService) List(req *ListOrderRequest) (*PageResult, error) {
	orders, total, err := s.orderRepo.ListWithFilter(
		req.StoreID, req.SalesmanID, req.OrderStatus, req.PaymentStatus,
		req.OrderType, req.StartDate, req.EndDate, req.Keyword,
		req.Page, req.PageSize,
	)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询订单列表失败"}
	}

	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	return &PageResult{
		List:     orders,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetDetail 订单详情
func (s *OrderService) GetDetail(orderID int64) (*OrderDetail, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.ErrOrderNotFound, Message: "订单不存在"}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	return &OrderDetail{
		Order:    order,
		Items:    order.Items,
		Gifts:    order.Gifts,
		Payments: order.Payments,
	}, nil
}

// GetOrderFeed 获取最近订单动态
func (s *OrderService) GetOrderFeed(userID int64, limit int) ([]models.Order, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	orders, err := s.orderRepo.GetOrderFeed(userID, limit)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询订单动态失败"}
	}
	return orders, nil
}
