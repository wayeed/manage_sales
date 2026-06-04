package service

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
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
	StoreID int64 `json:"store_id" example:1`
	SalesmanID int64 `json:"salesman_id" example:1`
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
	ItemID int64 `json:"item_id" example:"0"` // 已有商品ID(修改订单时用于标识已有商品, 0或不传表示新增)
	SKUID int64 `json:"sku_id" example:1`
	ProductName string `json:"product_name" binding:"required" example:"真皮沙发"`
	SKUName string `json:"sku_name" binding:"required" example:"真皮沙发-棕色-三座"`
	SKUCode string `json:"sku_code" example:"SKU001"`
	CategoryID *int64 `json:"category_id" example:1`
	Quantity int `json:"quantity" binding:"required,min=1" example:1`
	ListPrice float64 `json:"list_price" binding:"required" example:8999.00`
	SalePrice float64 `json:"sale_price" binding:"required" example:8500.00`
	CostPrice float64 `json:"cost_price" example:"2000.00"`
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
	Order           *models.Order          `json:"order" example:{}`
	Items           []models.OrderItem     `json:"items" example:[]`
	Gifts           []models.OrderGift     `json:"gifts" example:[]`
	Payments        []models.Payment       `json:"payments" example:[]`
	OutboundRequest *models.OutboundRequest `json:"outbound_request,omitempty"`
	DeliveryRecords []models.DeliveryRecord `json:"delivery_records" example:[]"`
}

// OrderService 订单核心服务
type OrderService struct {
	db             *gorm.DB
	orderRepo      *repository.OrderRepository
	paymentRepo    *repository.PaymentRepository
	customerRepo   *repository.CustomerRepository
	peerRepo       *repository.PeerRepository
	inventorySvc    *InventoryService
	orderReturnRepo *repository.OrderReturnRepository
	commissionSvc   *CommissionService
	userSvc        *UserService
}

// NewOrderService 创建订单服务实例
func NewOrderService(
	db *gorm.DB,
	orderRepo *repository.OrderRepository,
	paymentRepo *repository.PaymentRepository,
	customerRepo *repository.CustomerRepository,
	peerRepo *repository.PeerRepository,
	inventorySvc *InventoryService,
	orderReturnRepo *repository.OrderReturnRepository,
	commissionSvc *CommissionService,
	userSvc *UserService,
) *OrderService {
	return &OrderService{
		db:             db,
		orderRepo:      orderRepo,
		paymentRepo:    paymentRepo,
		customerRepo:   customerRepo,
		peerRepo:       peerRepo,
		inventorySvc:   inventorySvc,
		orderReturnRepo: orderReturnRepo,
		commissionSvc:  commissionSvc,
		userSvc:        userSvc,
	}
}

// GetUserByID 根据ID获取用户信息
func (s *OrderService) GetUserByID(userID int64) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
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

	// 2. 查询默认仓库（非草稿）
	var defaultWarehouse models.Warehouse
	if !isDraft {
		if err := s.db.Where("store_id = ? AND warehouse_type = 1 AND status = 1", req.StoreID).First(&defaultWarehouse).Error; err != nil {
			return nil, &AppError{Code: apperrors.InternalError, Message: "未找到默认仓库"}
		}
		// 注：新流程不检查库存，0库存也可下单
	}

	// 3. 使用事务创建订单
	var order *models.Order
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 注：新流程不锁定库存，库存检查移到送货出库时

		// b. 计算订单金额
		totalListPrice := decimal.Zero
		totalSalePrice := decimal.Zero
		totalCost := decimal.Zero
		totalQuantity := 0
		skuCount := len(req.Items)
		categorySet := make(map[int64]bool)

		for _, item := range req.Items {
			lp := decimal.NewFromFloat(item.ListPrice)
			sp := decimal.NewFromFloat(item.SalePrice)
			cp := decimal.NewFromFloat(item.CostPrice)
			totalListPrice = totalListPrice.Add(lp.Mul(decimal.NewFromInt(int64(item.Quantity))))
			totalSalePrice = totalSalePrice.Add(sp.Mul(decimal.NewFromInt(int64(item.Quantity))))
			totalCost = totalCost.Add(cp.Mul(decimal.NewFromInt(int64(item.Quantity)))) // 累加商品成本
			totalQuantity += item.Quantity

			// 获取商品分类ID
			categoryID := item.CategoryID
			if categoryID == nil && item.SKUID > 0 {
				// 如果前端没传分类ID，根据SKUID查询商品分类
				var sku models.ProductSKU
				if err := tx.Preload("Product").First(&sku, item.SKUID).Error; err == nil {
					if sku.Product != nil && sku.Product.CategoryID != nil {
						categoryID = sku.Product.CategoryID
					}
				}
			}
			if categoryID != nil {
				categorySet[*categoryID] = true
			}
		}

		categoryCount := len(categorySet)

		// 多品订单折扣（3个及以上品类95折）
		discountAmount := decimal.Zero
		if categoryCount >= 3 {
			discountAmount = totalSalePrice.Mul(decimal.NewFromFloat(0.05))
		}
		finalAmount := totalSalePrice.Sub(discountAmount)

		// d. 计算gift_cost并累加到total_cost
		giftCost := decimal.Zero
		for _, gift := range req.Gifts {
			gc := decimal.NewFromFloat(gift.CostPrice).Mul(decimal.NewFromInt(int64(gift.Quantity)))
			giftCost = giftCost.Add(gc)
			totalCost = totalCost.Add(gc) // 累加礼品成本到总成本
		}

		// e. 计算actual_profit = final_amount - total_cost
		actualProfit := finalAmount.Sub(totalCost)
		if actualProfit.LessThan(decimal.Zero) {
			actualProfit = decimal.Zero // 利润不能为负
		}

		// f. 自动判断订单类型
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
		} else if categoryCount >= 3 {
			orderType = 2 // 多品（3种及以上品类）
		}

		// g. 生成订单号
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
				SKUCode:     item.SKUCode,
				CategoryID:  item.CategoryID,
				Quantity:    item.Quantity,
				ListPrice:   listPrice,
				SalePrice:   salePrice,
				DiscountRate: discountRate,
				UnitCost:  decimal.NewFromFloat(item.CostPrice),
				TotalCost: decimal.NewFromFloat(item.CostPrice).Mul(decimal.NewFromInt(int64(item.Quantity))),
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

		// i. 记录inventory_transactions（类型=库存锁定）
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
				TransactionType: models.TransactionTypeLock, // 9: 库存锁定
				BizType:         1,                          // 商品
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
func (s *OrderService) ApproveOrder(orderID int64, approvedBy int64, approved bool, remark string, depositAmount string, roleCodes []string) error {
	// 检查审核权限（主管、店长、财务可审核）
	if !hasRole(roleCodes, "SUPERVISOR") && !hasRole(roleCodes, "STORE_MANAGER") && !hasRole(roleCodes, "FINANCE") && !hasRole(roleCodes, "BOSS") {
		return &AppError{Code: apperrors.Forbidden, Message: "无权审核订单"}
	}

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
	return s.approveOrder(order, approvedBy, remark, depositAmount)
}

// approveOrder 审核通过
// 优化后：审核通过时不扣减库存，只更新订单状态。库存将在送货出库时扣减。
func (s *OrderService) approveOrder(order *models.Order, approvedBy int64, remark string, depositAmount string) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		// 更新订单状态为已生效
		// 注意：total_cost和actual_profit将在送货出库时计算
		if err := tx.Model(order).Updates(map[string]interface{}{
			"order_status":  1, // 已生效
			"approved_by":   approvedBy,
			"approved_at":   now,
			"approval_remark": remark,
		}).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "更新订单状态失败"}
		}

		// 查询默认仓库
		var defaultWarehouse models.Warehouse
		if err := tx.Where("store_id = ? AND warehouse_type = 1 AND status = 1", order.StoreID).First(&defaultWarehouse).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "未找到默认仓库"}
		}

		// ===== 新流程：审核通过时按批次FIFO锁定库存 =====
		// 查询订单商品（排除已标记移除的）
		var orderItems []models.OrderItem
		if err := tx.Where("order_id = ? AND item_status != 2", order.ID).Find(&orderItems).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "查询订单商品失败"}
		}

		// 统计库存状态
		var hasStockCount, noStockCount int

		for _, item := range orderItems {
			// 尝试按批次FIFO锁定库存
			costDetails, err := s.inventorySvc.LockStockByBatch(defaultWarehouse.ID, item.SKUID, item.Quantity, order.ID, order.StoreID)
			if err != nil {
				return err
			}

			if len(costDetails) > 0 {
				// 有库存：将第一个批次的ID写入order_items（简化为一单一批次绑定）
				// 如果跨多个批次，记录第一个批次ID，成本取加权平均
				var totalCost decimal.Decimal
				var totalQty int
				for _, cd := range costDetails {
					totalCost = totalCost.Add(cd.TotalCost)
					totalQty += cd.Quantity
				}
				avgUnitCost := decimal.Zero
				if totalQty > 0 {
					avgUnitCost = totalCost.Div(decimal.NewFromInt(int64(totalQty)))
				}

				if err := tx.Model(&models.OrderItem{}).Where("id = ?", item.ID).Updates(map[string]interface{}{
					"batch_id":    costDetails[0].BatchID,
					"unit_cost":   avgUnitCost,
					"total_cost":  totalCost,
				}).Error; err != nil {
					return &AppError{Code: apperrors.InternalError, Message: "更新订单商品批次失败"}
				}
				hasStockCount++
			} else {
				// 无库存：创建缺货排队记录
				queue := &models.StockQueue{
					OrderID:     order.ID,
					OrderItemID: item.ID,
					SKUID:       item.SKUID,
					Quantity:    item.Quantity,
					Status:      0, // 排队中
					CreatedAt:   now,
					UpdatedAt:   now,
				}
				if err := tx.Create(queue).Error; err != nil {
					return &AppError{Code: apperrors.InternalError, Message: "创建缺货排队记录失败"}
				}
				noStockCount++
			}
		}

		// 计算订单库存状态：0-全部有库存, 1-部分缺货, 2-全部缺货
		var stockStatus int8 = 0
		if noStockCount > 0 {
			if hasStockCount == 0 {
				stockStatus = 2 // 全部缺货
			} else {
				stockStatus = 1 // 部分缺货
			}
		}
		if err := tx.Model(order).Update("stock_status", stockStatus).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "更新订单库存状态失败"}
		}
		// 新流程结束

		// 处理修改订单的商品变更
		if order.EditCount > 0 {
			// 释放移除商品的库存（item_status=2）
			var removedItems []models.OrderItem
			tx.Where("order_id = ? AND item_status = 2", order.ID).Find(&removedItems)
			for _, item := range removedItems {
				if err := s.unlockStockInTx(tx, defaultWarehouse.ID, item.SKUID, item.Quantity); err != nil {
					return err
				}
				// 记录库存流水
				var stock models.WarehouseStock
				if err := tx.Where("warehouse_id = ? AND sku_id = ?", defaultWarehouse.ID, item.SKUID).First(&stock).Error; err == nil {
					invTx := &models.InventoryTransaction{
						StoreID:         order.StoreID,
						WarehouseID:     &defaultWarehouse.ID,
						TransactionType: models.TransactionTypeUnlock,
						BizType:         1,
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
			}

			// 释放保留商品中数量减少部分的库存
			// 需要对比当前数量和原始锁定数量，但由于数量已在修改时更新，
			// 我们需要通过其他方式获取差额。这里通过对比当前 item 的锁定库存
			// 和实际数量来处理。实际上在修改订单时数量已更新，减少的部分
			// 库存仍然处于锁定状态，需要在审核通过后释放。
			// 由于我们无法直接知道原始数量，这里采用另一种方式：
			// 查询 item_status=0 的商品，检查是否有数量减少的情况
			// 但由于数量已经在 UpdateOrder 中被更新了，我们需要在 UpdateOrder 中
			// 记录数量差额。为了简化，我们在此处不处理数量减少的库存释放，
			// 因为减少的数量对应的库存仍然处于锁定状态，会在送货出库时自然处理。
			// （如果需要精确处理，可以在 UpdateOrder 时记录一个数量变更日志）

			// 物理删除移除的商品
			if err := tx.Where("order_id = ? AND item_status = 2", order.ID).Delete(&models.OrderItem{}).Error; err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "删除移除商品失败"}
			}

			// 将新增商品状态改为正常（item_status=1 -> 0）
			if err := tx.Model(&models.OrderItem{}).
				Where("order_id = ? AND item_status = 1", order.ID).
				Update("item_status", 0).Error; err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "更新新增商品状态失败"}
			}
		}

		// 更新客户累计统计（此时不计算利润，等送货后再更新）
		if order.CustomerID != nil {
			if err := tx.Model(&models.Customer{}).Where("id = ?", *order.CustomerID).Updates(map[string]interface{}{
				"total_orders":  gorm.Expr("total_orders + 1"),
				"total_amount":  gorm.Expr("total_amount + ?", order.FinalAmount),
				"last_order_at": now,
			}).Error; err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "更新客户统计失败"}
			}
		}

		// 如果是同行订单，更新同行累计统计
		if order.IsPeerOrder == 1 && order.PeerID != nil {
			if err := tx.Model(&models.Peer{}).Where("id = ?", *order.PeerID).Updates(map[string]interface{}{
				"total_orders":  gorm.Expr("total_orders + 1"),
				"total_amount":  gorm.Expr("total_amount + ?", order.FinalAmount),
				"last_order_at": now,
			}).Error; err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "更新同行统计失败"}
			}
		}

		// 如果传入了订金金额，创建订金回款记录
		if depositAmount != "" {
			depAmount, err := decimal.NewFromString(depositAmount)
			if err != nil {
				return &AppError{Code: apperrors.BadRequest, Message: "订金金额格式错误"}
			}
			if depAmount.GreaterThan(decimal.Zero) {
				// 校验订金不超过订单成交价
				if depAmount.GreaterThan(order.FinalAmount) {
					return &AppError{Code: apperrors.BadRequest, Message: "订金金额不能超过订单成交价"}
				}

				paymentNo := "PAY" + appsnow.GenerateOrderNo()
				payment := &models.Payment{
					OrderID:       order.ID,
					PaymentNo:     paymentNo,
					Amount:        depAmount,
					PaymentDate:   &now,
					PaymentMethod: 0,
					Status:        1, // 已审核
					PaymentType:   1, // 订金
					Remark:        "审批时录入订金",
					CreatedBy:     &approvedBy,
					AuditedBy:     &approvedBy,
					AuditedAt:     &now,
				}
				if err := tx.Create(payment).Error; err != nil {
					return &AppError{Code: apperrors.InternalError, Message: "创建订金回款记录失败"}
				}

				// 更新订单 paid_amount
				newPaidAmount := order.PaidAmount.Add(depAmount)
				remainingAmount := order.FinalAmount.Sub(newPaidAmount)
				if remainingAmount.IsNegative() {
					remainingAmount = decimal.Zero
				}

				paymentStatus := int8(0)
				if newPaidAmount.GreaterThanOrEqual(order.FinalAmount) {
					paymentStatus = 2 // 已回款
				} else if newPaidAmount.GreaterThan(decimal.Zero) {
					paymentStatus = 1 // 部分回款
				}

				if err := tx.Model(order).Updates(map[string]interface{}{
					"paid_amount":      newPaidAmount,
					"remaining_amount": remainingAmount,
					"payment_status":   paymentStatus,
				}).Error; err != nil {
					return &AppError{Code: apperrors.InternalError, Message: "更新订单回款信息失败"}
				}
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
// 新流程：驳回时无库存变动（因为审核通过时才锁定库存，驳回时还未锁定）
func (s *OrderService) rejectOrder(order *models.Order, approvedBy int64, remark string) error {
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

		// 新流程：驳回时不需要释放库存（审核通过时才锁定，驳回时还未锁定）
		// 如果是已生效订单被修改后重新审批驳回（status=1 -> 修改 -> status=0 -> 驳回），
		// 此时也不需要释放，因为修改时状态重置为0，锁定已在修改流程中处理

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

// lockStockInTx 在事务内锁定库存
func (s *OrderService) lockStockInTx(tx *gorm.DB, warehouseID, skuID int64, quantity int) error {
	var stock models.WarehouseStock
	if err := tx.Where("warehouse_id = ? AND sku_id = ?", warehouseID, skuID).First(&stock).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrInsufficientStock, Message: "库存记录不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if stock.AvailableQuantity < quantity {
		return &AppError{Code: apperrors.ErrInsufficientStock, Message: "可用库存不足"}
	}

	result := tx.Model(&models.WarehouseStock{}).
		Where("warehouse_id = ? AND sku_id = ? AND version = ?", warehouseID, skuID, stock.Version).
		Updates(map[string]interface{}{
			"locked_quantity":    gorm.Expr("locked_quantity + ?", quantity),
			"available_quantity": gorm.Expr("available_quantity - ?", quantity),
			"version":            gorm.Expr("version + 1"),
		})
	if result.Error != nil {
		return &AppError{Code: apperrors.InternalError, Message: "锁定库存失败"}
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
func (s *OrderService) ReturnOrder(orderID int64, returnAmount float64, returnProfit float64, warehouseID int64, remark string, operatorID int64, operatorName string) error {
	// 加载订单详情（包括商品明细）
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

	// 验证订单有商品明细
	if len(order.Items) == 0 {
		return &AppError{Code: apperrors.InternalError, Message: "订单没有商品明细"}
	}

	ra := decimal.NewFromFloat(returnAmount)
	rp := decimal.NewFromFloat(returnProfit)

	// 生成退货单号（格式：RET + 年月日时分秒 + 6位随机数，确保不超过32位）
	returnNo := fmt.Sprintf("RET%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建退货记录
		returnRecord := &models.OrderReturn{
			StoreID:      order.StoreID,
			OrderID:      orderID,
			ReturnNo:     returnNo,
			ReturnAmount: ra,
			ReturnProfit: rp,
			Reason:       remark,
			OperatorID:   operatorID,
			OperatorName: operatorName,
			ReturnTime:   time.Now(),
		}
		if err := tx.Create(returnRecord).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "创建退货记录失败"}
		}

			// 2. 库存处理
		// 根据是否出库决定如何处理库存
		if order.DeliveryStatus == 0 {
			// 未出库：释放锁定库存（将 locked_quantity 转回 available_quantity）
			for _, item := range order.Items {
				if item.BatchID != nil && *item.BatchID > 0 {
					// 获取批次信息以确定仓库
					var batch models.InventoryBatch
					if err := tx.First(&batch, *item.BatchID).Error; err == nil && batch.WarehouseID != nil {
						if err := s.inventorySvc.UnlockStock(*batch.WarehouseID, int64(item.SKUID), item.Quantity); err != nil {
							return &AppError{Code: apperrors.InternalError, Message: fmt.Sprintf("释放锁定库存失败: %v", err)}
						}
					}
				}
			}
		} else {
			// 已出库：增加库存（退货入库）
			for _, item := range order.Items {
				if err := s.inventorySvc.IncreaseStock(
					int64(item.SKUID),
					item.Quantity,
					order.StoreID,
					warehouseID,
					orderID,
					operatorID,
					"退货入库",
				); err != nil {
					return &AppError{Code: apperrors.InternalError, Message: fmt.Sprintf("库存回滚失败: %v", err)}
				}
			}
		}

		// 3. 提成回滚
		if s.commissionSvc != nil {
			if err := s.commissionSvc.ReverseCommission(orderID, tx); err != nil {
				return &AppError{Code: apperrors.InternalError, Message: fmt.Sprintf("提成回滚失败: %v", err)}
			}
		}

		// 4. 更新订单状态
		if err := tx.Model(order).Updates(map[string]interface{}{
			"is_returned":    1,
			"return_amount":  ra,
			"return_profit":  rp,
			"order_status":   4, // 已退货
		}).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "更新退货信息失败"}
		}

		// 5. 更新客户累计统计（冲减）
		if order.CustomerID != nil {
			if err := tx.Model(&models.Customer{}).Where("id = ?", *order.CustomerID).Updates(map[string]interface{}{
				"total_orders":  gorm.Expr("total_orders - 1"),
				"total_amount": gorm.Expr("total_amount - ?", ra),
				"total_profit": gorm.Expr("total_profit - ?", rp),
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

// ListWithPermission 带权限控制的订单列表查询
// roleCodes: 用户角色编码列表
func (s *OrderService) ListWithPermission(req *ListOrderRequest, userID int64, roleCodes []string) (*PageResult, error) {
	// 获取用户信息
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "获取用户信息失败"}
	}
	if user == nil {
		return nil, &AppError{Code: apperrors.ErrUserNotFound, Message: "用户不存在"}
	}

	var salesmanIDs []int64
	storeID := req.StoreID

	// 根据角色确定查询范围
	// 管理员、老板、店长、财务可以查看门店所有订单
	if hasRole(roleCodes, "ADMIN") || hasRole(roleCodes, "BOSS") || hasRole(roleCodes, "STORE_MANAGER") || hasRole(roleCodes, "FINANCE") {
		// 管理员、老板、店长、财务：查看门店所有订单，不限制业务员
		if user.StoreID != nil {
			storeID = strconv.FormatInt(*user.StoreID, 10)
		}
		salesmanIDs = nil // 不限制业务员
	} else if hasRole(roleCodes, "SUPERVISOR") {
		// 主管：查看自己及直属下级的订单
		subordinateIDs, err := s.userSvc.GetDirectSubordinateIDs(userID)
		if err != nil {
			subordinateIDs = []int64{}
		}
		salesmanIDs = append(subordinateIDs, userID)
	} else {
		// 普通业务员：只能查看自己的订单
		salesmanIDs = []int64{userID}
	}

	// 调试日志
	fmt.Printf("[DEBUG] ListWithPermission: userID=%d, roleCodes=%v, storeID=%s, salesmanIDs=%v\n",
		userID, roleCodes, storeID, salesmanIDs)
	fmt.Printf("[DEBUG] ListWithPermission: req.StoreID=%s, user.StoreID=%v\n",
		req.StoreID, user.StoreID)

	// 执行查询
	orders, total, err := s.orderRepo.ListWithFilterAndSalesmanIDs(
		storeID, salesmanIDs, "", req.OrderStatus, req.PaymentStatus,
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

// hasRole 检查用户是否拥有指定角色
func hasRole(roleCodes []string, role string) bool {
	for _, code := range roleCodes {
		if code == role {
			return true
		}
	}
	return false
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

	// 加载业务员名称
	if order.SalesmanID > 0 {
		var salesman models.User
		if err := s.db.Select("real_name").First(&salesman, order.SalesmanID).Error; err == nil {
			order.SalesmanName = salesman.RealName
		}
	}

	// 查询送货记录
	var deliveryRecords []models.DeliveryRecord
	s.db.Where("order_id = ?", orderID).Order("id DESC").Find(&deliveryRecords)

	return &OrderDetail{
		Order:           order,
		Items:           order.Items,
		Gifts:           order.Gifts,
		Payments:        order.Payments,
		OutboundRequest: order.OutboundRequest,
		DeliveryRecords: deliveryRecords,
	}, nil
}

// GetCustomerDraft 获取客户最新草稿订单
func (s *OrderService) GetCustomerDraft(customerID int64) (*OrderDetail, error) {
	var order models.Order
	err := s.db.Where("customer_id = ? AND is_draft = ?", customerID, 1).
		Order("id DESC").
		First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询草稿订单失败"}
	}

	// 加载订单详情
	return s.GetDetail(order.ID)
}

// DeleteOrder 删除订单
func (s *OrderService) DeleteOrder(orderID int64) error {
	var order models.Order
	if err := s.db.First(&order, orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: 404, Message: "订单不存在"}
		}
		return &AppError{Code: 500, Message: "查询订单失败"}
	}

	// 删除订单明细
	s.db.Where("order_id = ?", orderID).Delete(&models.OrderItem{})
	// 删除赠品明细
	s.db.Where("order_id = ?", orderID).Delete(&models.OrderGift{})
	// 删除订单
	if err := s.db.Delete(&order).Error; err != nil {
		return &AppError{Code: 500, Message: "删除订单失败"}
	}

	return nil
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

// OrderCommissionDetail 订单利润提成详情
type OrderCommissionDetail struct {
	OrderID          int64                `json:"order_id"`
	OrderNo          string               `json:"order_no"`
	OrderType        int8                 `json:"order_type"`
	OrderTypeName    string               `json:"order_type_name"`
	FinalAmount      decimal.Decimal      `json:"final_amount"`
	TotalCost        decimal.Decimal      `json:"total_cost"`
	GiftCost         decimal.Decimal      `json:"gift_cost"`
	ActualProfit     decimal.Decimal      `json:"actual_profit"`
	CommissionRate   decimal.Decimal      `json:"commission_rate"`
	CommissionAmount decimal.Decimal      `json:"commission_amount"`
	Items            []OrderItemCommission `json:"items"`
}

// OrderItemCommission 订单商品提成明细
type OrderItemCommission struct {
	ProductName string          `json:"product_name"`
	SKUName     string          `json:"sku_name"`
	Quantity    int             `json:"quantity"`
	SalePrice   decimal.Decimal `json:"sale_price"`
	UnitCost    decimal.Decimal `json:"unit_cost"`
	TotalCost   decimal.Decimal `json:"total_cost"`
}

// GetOrderCommissionDetail 获取订单利润提成详情
func (s *OrderService) GetOrderCommissionDetail(orderID int64) (*OrderCommissionDetail, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.ErrOrderNotFound, Message: "订单不存在"}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询订单失败"}
	}

	// 获取订单明细（包含成本信息，排除已标记移除的商品）
	var items []models.OrderItem
	if err := s.db.Where("order_id = ? AND item_status != 2", orderID).Find(&items).Error; err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询订单明细失败"}
	}

	// 组装商品明细
	itemCommissions := make([]OrderItemCommission, 0, len(items))
	for _, item := range items {
		itemCommissions = append(itemCommissions, OrderItemCommission{
			ProductName: item.ProductName,
			SKUName:     item.SKUName,
			Quantity:    item.Quantity,
			SalePrice:   item.SalePrice,
			UnitCost:    item.UnitCost,
			TotalCost:   item.TotalCost,
		})
	}

	// 获取业务员等级
	var salesmanLevel int8 = 1 // 默认初级
	var salesman models.User
	if err := s.db.First(&salesman, order.SalesmanID).Error; err == nil {
		if salesman.Level > 0 {
			salesmanLevel = salesman.Level
		}
	}

	// 获取提成比例（根据订单类型和业务员等级）
	rate := s.getCommissionRateByOrderTypeAndLevel(order.OrderType, salesmanLevel)

	// 计算预估提成
	commissionAmount := decimal.Zero
	if order.ActualProfit.GreaterThan(decimal.Zero) {
		commissionAmount = order.ActualProfit.Mul(rate)
	}

	// 订单类型名称
	orderTypeName := s.getOrderTypeName(order.OrderType)

	return &OrderCommissionDetail{
		OrderID:          order.ID,
		OrderNo:          order.OrderNo,
		OrderType:        order.OrderType,
		OrderTypeName:    orderTypeName,
		FinalAmount:      order.FinalAmount,
		TotalCost:        order.TotalCost,
		GiftCost:         order.GiftCost,
		ActualProfit:     order.ActualProfit,
		CommissionRate:   rate,
		CommissionAmount: commissionAmount,
		Items:            itemCommissions,
	}, nil
}

// getCommissionRateByOrderTypeAndLevel 根据订单类型和业务员等级获取提成比例
// 用于订单详情页显示预估提成
func (s *OrderService) getCommissionRateByOrderTypeAndLevel(orderType int8, salesmanLevel int8) decimal.Decimal {
	// 同行订单：业务员提成 = (级别比例 - 同行比例)
	peerRates := map[int8]string{
		4: "commission_rate_peer_single",
		5: "commission_rate_peer_multi",
		6: "commission_rate_peer_special",
	}
	if key, ok := peerRates[orderType]; ok {
		// 获取业务员级别对应的比例
		salesmanRate := s.getSalesmanRateByLevelForOrder(orderType, salesmanLevel)
		// 获取同行比例
		peerRate := s.getRateFromDB(key)
		if peerRate.LessThanOrEqual(decimal.Zero) {
			peerRate = decimal.NewFromFloat(0.10) // 默认10%
		}
		// 实际提成比例 = 级别比例 - 同行比例
		actualRate := salesmanRate.Sub(peerRate)
		if actualRate.LessThan(decimal.Zero) {
			return decimal.Zero
		}
		return actualRate
	}

	// 非同行订单，根据业务员等级确定比例
	return s.getSalesmanRateByLevelForOrder(orderType, salesmanLevel)
}

// getSalesmanRateByLevelForOrder 根据订单类型和业务员等级获取业务员比例
func (s *OrderService) getSalesmanRateByLevelForOrder(orderType int8, salesmanLevel int8) decimal.Decimal {
	if salesmanLevel <= 0 {
		salesmanLevel = 1
	}

	// 判断单品/多品
	isMulti := orderType == 2 // 多品订单
	// 同行多品也是多品
	if orderType == 5 {
		isMulti = true
	}

	key := fmt.Sprintf("commission_rate_level%d_", salesmanLevel)
	if isMulti {
		key += "multi"
	} else {
		key += "single"
	}

	rate := s.getRateFromDB(key)
	if rate.GreaterThan(decimal.Zero) {
		return rate
	}

	// 降级使用初级等级比例
	if isMulti {
		return decimal.NewFromFloat(0.10)
	}
	return decimal.NewFromFloat(0.08)
}

// getRateFromDB 从 system_configs 表获取配置比例
func (s *OrderService) getRateFromDB(configKey string) decimal.Decimal {
	var config models.SystemConfig
	if err := s.db.Where("config_key = ?", configKey).First(&config).Error; err != nil {
		return decimal.Zero
	}
	rate, err := decimal.NewFromString(config.ConfigValue)
	if err != nil {
		return decimal.Zero
	}
	return rate
}

// getOrderTypeName 获取订单类型名称
func (s *OrderService) getOrderTypeName(orderType int8) string {
	names := map[int8]string{
		1: "单品",
		2: "多品",
		3: "特批",
		4: "同行单品",
		5: "同行多品",
		6: "同行特批",
	}
	if name, ok := names[orderType]; ok {
		return name
	}
	return "未知"
}

// GeneratePurchaseFromOrder 从订单生成采购单
// 根据订单商品明细自动生成采购单，用于"先下单后采购"的业务场景
func (s *OrderService) GeneratePurchaseFromOrder(orderID int64, supplierID int64, supplierName string, warehouseID int64, createdBy int64) (*models.PurchaseOrder, error) {
	// 1. 查询订单详情
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.ErrOrderNotFound, Message: "订单不存在"}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询订单失败"}
	}

	// 2. 验证订单状态（已审核通过才能生成采购单）
	if order.OrderStatus != 1 {
		return nil, &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "只有已审核通过的订单才能生成采购单"}
	}

	// 3. 验证订单有商品明细
	if len(order.Items) == 0 {
		return nil, &AppError{Code: apperrors.BadRequest, Message: "订单没有商品明细"}
	}

	// 4. 构建采购明细（仅包含库存不足的商品）
	purchaseItems := make([]models.PurchaseItem, 0)
	var totalAmount decimal.Decimal
	totalQuantity := 0
	skippedCount := 0

	for _, item := range order.Items {
		// 查询该SKU在目标仓库的可用库存
		availableQty, _ := s.inventorySvc.GetAvailableStock(warehouseID, item.SKUID)
		if availableQty >= item.Quantity {
			// 库存充足，跳过
			skippedCount++
			continue
		}

		// 库存不足，计算需要采购的数量
		purchaseQty := item.Quantity - availableQty
		if purchaseQty <= 0 {
			skippedCount++
			continue
		}

		// 使用订单的成本价作为采购价
		purchasePrice := item.UnitCost
		if purchasePrice.IsZero() {
			purchasePrice = item.SalePrice.Mul(decimal.NewFromFloat(0.7))
		}

		subtotal := purchasePrice.Mul(decimal.NewFromInt(int64(purchaseQty)))
		totalAmount = totalAmount.Add(subtotal)
		totalQuantity += purchaseQty

		purchaseItems = append(purchaseItems, models.PurchaseItem{
			SKUID:         &item.SKUID,
			ProductName:   item.ProductName,
			SKUName:       item.SKUName,
			PurchasePrice: purchasePrice,
			Quantity:      purchaseQty,
			Subtotal:      subtotal,
		})
	}

	// 如果所有商品都有库存，无需采购
	if len(purchaseItems) == 0 {
		return nil, &AppError{Code: apperrors.BadRequest, Message: "所有商品库存充足，无需生成采购单"}
	}

	// 5. 生成采购单号
	purchaseNo := "PO" + appsnow.GenerateOrderNo()

	// 6. 创建采购订单
	purchaseOrder := &models.PurchaseOrder{
		StoreID:       order.StoreID,
		PurchaseNo:    purchaseNo,
		SupplierID:    &supplierID,
		SupplierName:  supplierName,
		TotalAmount:   totalAmount,
		TotalQuantity: totalQuantity,
		Status:        0, // 待审核
		Remark:        fmt.Sprintf("由订单[%s]自动生成(共%d件商品,跳过%d件有库存)", order.OrderNo, len(order.Items), skippedCount),
		CreatedBy:     &createdBy,
	}

	// 7. 事务保存
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 保存采购订单
		if err := tx.Create(purchaseOrder).Error; err != nil {
			return err
		}

		// 保存采购明细
		for i := range purchaseItems {
			purchaseItems[i].PurchaseOrderID = purchaseOrder.ID
			if err := tx.Create(&purchaseItems[i]).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "创建采购订单失败"}
	}

	return purchaseOrder, nil
}

// UpdateOrder 修改订单
func (s *OrderService) UpdateOrder(req *CreateOrderRequest, orderID int64, userID int64) error {
	// 1. 查询订单
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrOrderNotFound, Message: "订单不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	// 2. 校验订单状态: 允许草稿(0)、已生效未送货(1)或已驳回(2)的订单修改
	if order.OrderStatus != 0 && order.OrderStatus != 1 && order.OrderStatus != 2 {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "当前订单状态不允许修改"}
	}

	// 3. 安全校验: 只能修改自己的订单
	if order.SalesmanID != userID {
		return &AppError{Code: apperrors.Forbidden, Message: "只能修改自己的订单"}
	}

	// 4. 校验: 未计算提成
	var commissionCount int64
	s.db.Model(&models.Commission{}).Where("order_id = ?", orderID).Count(&commissionCount)
	if commissionCount > 0 {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "订单已计算提成，不允许修改"}
	}

	// 5. 查询已回款金额（排除已驳回的回款）
	var totalPaid decimal.Decimal
	s.db.Model(&models.Payment{}).
		Where("order_id = ? AND status != 2", orderID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalPaid)

	// 6. 验证商品SKU存在且上架
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
		return &AppError{Code: apperrors.InternalError, Message: "查询商品SKU失败"}
	}
	if len(skus) != len(skuIDs) {
		return &AppError{Code: apperrors.BadRequest, Message: "部分商品SKU不存在或已下架"}
	}

	// 7. 查询默认仓库
	var defaultWarehouse models.Warehouse
	if err := s.db.Where("store_id = ? AND warehouse_type = 1 AND status = 1", order.StoreID).First(&defaultWarehouse).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "未找到默认仓库"}
	}

	// 8. 构建旧商品映射 (itemID -> OrderItem) 和旧SKU映射 (skuID -> OrderItem)
	// 只取 item_status != 2 的旧商品（已标记移除的不再参与对比）
	oldItemByID := make(map[int64]models.OrderItem)
	oldItemBySKU := make(map[int64]models.OrderItem)
	for _, item := range order.Items {
		if item.ItemStatus != 2 {
			oldItemByID[item.ID] = item
			oldItemBySKU[item.SKUID] = item
		}
	}

	// 构建新商品映射 (skuID -> CreateOrderItemRequest)
	newItemBySKU := make(map[int64]CreateOrderItemRequest)
	for _, item := range req.Items {
		newItemBySKU[item.SKUID] = item
	}

	// 9. 验证新增商品的库存充足
	for _, item := range req.Items {
		if _, exists := oldItemBySKU[item.SKUID]; !exists {
			// 新增商品，需要验证库存
			stock, err := s.inventorySvc.GetStock(defaultWarehouse.ID, item.SKUID)
			if err != nil {
				return err
			}
			if stock == nil || stock.AvailableQuantity < item.Quantity {
				return &AppError{Code: apperrors.ErrInsufficientStock, Message: fmt.Sprintf("商品[%s]库存不足", item.SKUName)}
			}
		}
	}

	// 对于保留商品中数量增加的，验证差额库存
	for _, newItem := range req.Items {
		if oldItem, exists := oldItemBySKU[newItem.SKUID]; exists {
			if newItem.Quantity > oldItem.Quantity {
				diff := newItem.Quantity - oldItem.Quantity
				stock, err := s.inventorySvc.GetStock(defaultWarehouse.ID, newItem.SKUID)
				if err != nil {
					return err
				}
				if stock == nil || stock.AvailableQuantity < diff {
					return &AppError{Code: apperrors.ErrInsufficientStock, Message: fmt.Sprintf("商品[%s]库存不足", newItem.SKUName)}
				}
			}
		}
	}

	originalStatus := order.OrderStatus

	// 10. 使用事务处理
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// a. 对比旧商品和新商品，进行增量处理

		// 遍历旧商品
		for _, oldItem := range oldItemByID {
			if newItem, exists := newItemBySKU[oldItem.SKUID]; exists {
				// 保留：更新价格/数量等字段，item_status=0
				listPrice := decimal.NewFromFloat(newItem.ListPrice)
				salePrice := decimal.NewFromFloat(newItem.SalePrice)
				discountRate := decimal.NewFromInt(1)
				if listPrice.GreaterThan(decimal.NewFromInt(0)) {
					discountRate = salePrice.Div(listPrice).Round(4)
				}
				unitCost := decimal.NewFromFloat(newItem.CostPrice)
				totalCost := unitCost.Mul(decimal.NewFromInt(int64(newItem.Quantity)))

				if err := tx.Model(&models.OrderItem{}).Where("id = ?", oldItem.ID).Updates(map[string]interface{}{
					"product_name":  newItem.ProductName,
					"sku_name":      newItem.SKUName,
					"category_id":   newItem.CategoryID,
					"quantity":      newItem.Quantity,
					"list_price":    listPrice,
					"sale_price":    salePrice,
					"discount_rate": discountRate,
					"unit_cost":     unitCost,
					"total_cost":    totalCost,
					"item_status":   0, // 正常
				}).Error; err != nil {
					return &AppError{Code: apperrors.InternalError, Message: "更新订单明细失败"}
				}

				// 如果数量增加，锁定差额库存
				if newItem.Quantity > oldItem.Quantity {
					diff := newItem.Quantity - oldItem.Quantity
					if err := s.lockStockInTx(tx, defaultWarehouse.ID, oldItem.SKUID, diff); err != nil {
						return err
					}
				}
				// 如果数量减少，不释放库存（审核通过后处理）
			} else {
				// 不在新列表中 → 标记移除（item_status=2），不释放库存
				if err := tx.Model(&models.OrderItem{}).Where("id = ?", oldItem.ID).Update("item_status", 2).Error; err != nil {
					return &AppError{Code: apperrors.InternalError, Message: "标记移除订单明细失败"}
				}
			}
		}

		// 遍历新商品
		for _, newItem := range req.Items {
			if _, exists := oldItemBySKU[newItem.SKUID]; !exists {
				// 不在旧列表中 → 新增（item_status=1），插入记录并锁定库存
				listPrice := decimal.NewFromFloat(newItem.ListPrice)
				salePrice := decimal.NewFromFloat(newItem.SalePrice)
				discountRate := decimal.NewFromInt(1)
				if listPrice.GreaterThan(decimal.NewFromInt(0)) {
					discountRate = salePrice.Div(listPrice).Round(4)
				}
				unitCost := decimal.NewFromFloat(newItem.CostPrice)
				totalCost := unitCost.Mul(decimal.NewFromInt(int64(newItem.Quantity)))

				orderItem := &models.OrderItem{
					OrderID:      orderID,
					SKUID:        newItem.SKUID,
					ProductName:  newItem.ProductName,
					SKUName:      newItem.SKUName,
					SKUCode:      newItem.SKUCode,
					CategoryID:   newItem.CategoryID,
					Quantity:     newItem.Quantity,
					ItemStatus:   1, // 新增
					ListPrice:    listPrice,
					SalePrice:    salePrice,
					DiscountRate: discountRate,
					UnitCost:     unitCost,
					TotalCost:    totalCost,
				}
				if err := tx.Create(orderItem).Error; err != nil {
					return &AppError{Code: apperrors.InternalError, Message: "创建新增订单明细失败"}
				}

				// 锁定库存
				if err := s.lockStockInTx(tx, defaultWarehouse.ID, newItem.SKUID, newItem.Quantity); err != nil {
					return err
				}
			}
		}

		// b. 删除旧 order_gifts（赠品仍使用全量替换逻辑）
		if err := tx.Where("order_id = ?", orderID).Delete(&models.OrderGift{}).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "删除旧订单赠品失败"}
		}

		// c. 计算订单金额（只统计 item_status != 2 的商品，即保留+新增）
		totalListPrice := decimal.Zero
		totalSalePrice := decimal.Zero
		totalCost := decimal.Zero
		totalQuantity := 0
		skuCount := 0
		categorySet := make(map[int64]bool)

		// 收集所有有效商品（保留的 + 新增的）
		for _, newItem := range req.Items {
			lp := decimal.NewFromFloat(newItem.ListPrice)
			sp := decimal.NewFromFloat(newItem.SalePrice)
			cp := decimal.NewFromFloat(newItem.CostPrice)
			totalListPrice = totalListPrice.Add(lp.Mul(decimal.NewFromInt(int64(newItem.Quantity))))
			totalSalePrice = totalSalePrice.Add(sp.Mul(decimal.NewFromInt(int64(newItem.Quantity))))
			totalCost = totalCost.Add(cp.Mul(decimal.NewFromInt(int64(newItem.Quantity))))
			totalQuantity += newItem.Quantity
			skuCount++

			categoryID := newItem.CategoryID
			if categoryID == nil && newItem.SKUID > 0 {
				var sku models.ProductSKU
				if err := tx.Preload("Product").First(&sku, newItem.SKUID).Error; err == nil {
					if sku.Product != nil && sku.Product.CategoryID != nil {
						categoryID = sku.Product.CategoryID
					}
				}
			}
			if categoryID != nil {
				categorySet[*categoryID] = true
			}
		}

		categoryCount := len(categorySet)

		// 多品订单折扣（3个及以上品类95折）
		discountAmount := decimal.Zero
		if categoryCount >= 3 {
			discountAmount = totalSalePrice.Mul(decimal.NewFromFloat(0.05))
		}
		finalAmount := totalSalePrice.Sub(discountAmount)

		// 计算 gift_cost
		giftCost := decimal.Zero
		for _, gift := range req.Gifts {
			gc := decimal.NewFromFloat(gift.CostPrice).Mul(decimal.NewFromInt(int64(gift.Quantity)))
			giftCost = giftCost.Add(gc)
			totalCost = totalCost.Add(gc)
		}

		// 计算 actual_profit
		actualProfit := finalAmount.Sub(totalCost)
		if actualProfit.LessThan(decimal.Zero) {
			actualProfit = decimal.Zero
		}

		// 校验: 新订单金额必须大于已回款金额
		if finalAmount.LessThan(totalPaid) {
			return &AppError{Code: apperrors.BadRequest, Message: fmt.Sprintf("修改后金额不能小于已回款金额(%s)", totalPaid.String())}
		}

		// 自动判断订单类型
		orderType := int8(1)
		if req.IsPeerOrder == 1 {
			if categoryCount == 1 {
				orderType = 4
			} else {
				orderType = 5
			}
			if req.IsSpecialApproved == 1 {
				orderType = 6
			}
		} else if req.IsSpecialApproved == 1 {
			orderType = 3
		} else if categoryCount >= 3 {
			orderType = 2
		}

		// d. 插入新 order_gifts
		for _, gift := range req.Gifts {
			gc := decimal.NewFromFloat(gift.CostPrice).Mul(decimal.NewFromInt(int64(gift.Quantity)))
			orderGift := &models.OrderGift{
				OrderID:   orderID,
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

		// e. 更新订单主表
		updates := map[string]interface{}{
			"customer_name":      req.CustomerName,
			"customer_phone":     req.CustomerPhone,
			"customer_address":   req.CustomerAddress,
			"source":             req.Source,
			"is_peer_order":      req.IsPeerOrder,
			"peer_id":            req.PeerID,
			"is_special_approved": req.IsSpecialApproved,
			"remark":             req.Remark,
			"total_list_price":   totalListPrice,
			"total_sale_price":   totalSalePrice,
			"discount_amount":    discountAmount,
			"final_amount":       finalAmount,
			"total_cost":         totalCost,
			"gift_cost":          giftCost,
			"actual_profit":      actualProfit,
			"category_count":     categoryCount,
			"sku_count":          skuCount,
			"total_quantity":     totalQuantity,
			"order_type":         orderType,
			"order_status":       0, // 重置为待审批
			"approved_by":        nil,
			"approved_at":        nil,
			"approval_remark":    "",
			"edit_count":         gorm.Expr("edit_count + 1"),
		}

		// 保留已回款金额，重新计算剩余金额和回款状态
		remainingAmount := finalAmount.Sub(totalPaid)
		if remainingAmount.IsNegative() {
			remainingAmount = decimal.Zero
		}

		paymentStatus := int8(0)
		if totalPaid.GreaterThanOrEqual(finalAmount) {
			paymentStatus = 2 // 已回款
		} else if totalPaid.GreaterThan(decimal.Zero) {
			paymentStatus = 1 // 部分回款
		}

		updates["paid_amount"] = totalPaid
		updates["payment_status"] = paymentStatus
		updates["remaining_amount"] = remainingAmount

		// 解析订单日期
		var orderDate *time.Time
		if req.OrderDate != "" {
			od, err := time.Parse("2006-01-02", req.OrderDate)
			if err == nil {
				orderDate = &od
			}
		}
		if orderDate != nil {
			updates["order_date"] = orderDate
		}

		if err := tx.Model(order).Updates(updates).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "更新订单失败"}
		}

		// f. 原状态=1时回退客户/同行累计统计
		if originalStatus == 1 {
			if order.CustomerID != nil {
				if err := tx.Model(&models.Customer{}).Where("id = ?", *order.CustomerID).Updates(map[string]interface{}{
					"total_orders": gorm.Expr("GREATEST(total_orders - 1, 0)"),
					"total_amount": gorm.Expr("GREATEST(total_amount - ?, 0)", order.FinalAmount),
				}).Error; err != nil {
					return &AppError{Code: apperrors.InternalError, Message: "回退客户统计失败"}
				}
			}

			if order.IsPeerOrder == 1 && order.PeerID != nil {
				if err := tx.Model(&models.Peer{}).Where("id = ?", *order.PeerID).Updates(map[string]interface{}{
					"total_orders": gorm.Expr("GREATEST(total_orders - 1, 0)"),
					"total_amount": gorm.Expr("GREATEST(total_amount - ?, 0)", order.FinalAmount),
				}).Error; err != nil {
					return &AppError{Code: apperrors.InternalError, Message: "回退同行统计失败"}
				}
			}
		}

		// g. 记录库存流水（新增商品锁定 + 保留商品数量增加的差额锁定）
		for _, newItem := range req.Items {
			if oldItem, exists := oldItemBySKU[newItem.SKUID]; exists {
				// 保留商品，只记录数量增加的差额
				if newItem.Quantity > oldItem.Quantity {
					diff := newItem.Quantity - oldItem.Quantity
					stock, err := s.inventorySvc.GetStock(defaultWarehouse.ID, newItem.SKUID)
					if err != nil {
						return err
					}
					if stock == nil {
						continue
					}
					invTx := &models.InventoryTransaction{
						StoreID:         order.StoreID,
						WarehouseID:     &defaultWarehouse.ID,
						TransactionType: models.TransactionTypeLock,
						BizType:         1,
						BizID:           &newItem.SKUID,
						RelatedOrderID:  &orderID,
						Quantity:        diff,
						BeforeStock:     stock.AvailableQuantity + diff,
						AfterStock:      stock.AvailableQuantity,
						CreatedBy:       &userID,
						CreatedAt:       time.Now(),
					}
					if err := tx.Create(invTx).Error; err != nil {
						return &AppError{Code: apperrors.InternalError, Message: "记录库存流水失败"}
					}
				}
			} else {
				// 新增商品，记录全部数量锁定
				stock, err := s.inventorySvc.GetStock(defaultWarehouse.ID, newItem.SKUID)
				if err != nil {
					return err
				}
				if stock == nil {
					continue
				}
				invTx := &models.InventoryTransaction{
					StoreID:         order.StoreID,
					WarehouseID:     &defaultWarehouse.ID,
					TransactionType: models.TransactionTypeLock,
					BizType:         1,
					BizID:           &newItem.SKUID,
					RelatedOrderID:  &orderID,
					Quantity:        newItem.Quantity,
					BeforeStock:     stock.AvailableQuantity + newItem.Quantity,
					AfterStock:      stock.AvailableQuantity,
					CreatedBy:       &userID,
					CreatedAt:       time.Now(),
				}
				if err := tx.Create(invTx).Error; err != nil {
					return &AppError{Code: apperrors.InternalError, Message: "记录库存流水失败"}
				}
			}
		}

		return nil
	})

	if err != nil {
		if appErr, ok := err.(*AppError); ok {
			return appErr
		}
		return &AppError{Code: apperrors.InternalError, Message: "修改订单失败"}
	}

	return nil
}
