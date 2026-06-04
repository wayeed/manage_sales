package service

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"furniture-commission/internal/dto"
	"furniture-commission/internal/models"
	"furniture-commission/internal/repository"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// DeliveryService 送货出库服务接口
type DeliveryService interface {
        // CreateDelivery 创建送货出库记录
        CreateDelivery(req *dto.CreateDeliveryRequest, operatorID uint64) (*models.DeliveryRecord, error)
        // GetDeliveryList 获取送货出库列表
        GetDeliveryList(req *dto.DeliveryListRequest) (*PageResult, error)
        // GetDeliveryDetail 获取送货出库详情
        GetDeliveryDetail(deliveryID uint64) (*dto.DeliveryDTO, error)
        // CancelDelivery 作废送货记录
        CancelDelivery(deliveryID uint64, operatorID uint64, remark string) error
        // GetPendingDeliveryOrders 获取待送货订单列表
        GetPendingDeliveryOrders(req *dto.PendingDeliveryOrderQuery) (*PageResult, error)
        // ConfirmDelivery 確认送达
        ConfirmDelivery(orderID int64) error
        // GetOrderStockStatus 获取订单库存状态
        GetOrderStockStatus(orderID int64, warehouseID int64) ([]map[string]interface{}, error)
        // PrintDelivery 打印送货单时更新配送状态
        PrintDelivery(orderID int64) error
}

// deliveryService 送货出库服务实现
type deliveryService struct {
	deliveryRepo     repository.DeliveryRepository
	orderRepo        *repository.OrderRepository
	userRepo         *repository.UserRepository
	inventoryService *InventoryService
	db               *gorm.DB
}

// NewDeliveryService 创建送货出库服务实例
func NewDeliveryService(
	deliveryRepo repository.DeliveryRepository,
	orderRepo *repository.OrderRepository,
	userRepo *repository.UserRepository,
	inventoryService *InventoryService,
	db *gorm.DB,
) DeliveryService {
	return &deliveryService{
		deliveryRepo:     deliveryRepo,
		orderRepo:        orderRepo,
		userRepo:         userRepo,
		inventoryService: inventoryService,
		db:               db,
	}
}

// CreateDelivery 创建送货出库记录
func (s *deliveryService) CreateDelivery(req *dto.CreateDeliveryRequest, operatorID uint64) (*models.DeliveryRecord, error) {
	// 1. 获取订单信息
	order, err := s.orderRepo.FindByID(int64(req.OrderID))
	if err != nil {
		return nil, errors.New("订单不存在")
	}

	// 2. 验证订单状态
	if order.OrderStatus != 1 { // 已生效
		return nil, errors.New("订单未审核通过，无法送货")
	}
	if order.DeliveryStatus != 0 { // 未配送
		return nil, errors.New("订单已配送或配送中")
	}

	// 检查出库申请状态（如果有出库申请记录）
	var outboundReq models.OutboundRequest
	outboundRemark := ""
	if err := s.db.Where("order_id = ?", order.ID).First(&outboundReq).Error; err == nil {
		if outboundReq.Status != 4 {
			return nil, errors.New("出库申请尚未审批通过，无法出库")
		}
		// 获取出库申请的备注，用于写入送货记录
		outboundRemark = outboundReq.Remark
	}

	// 3. 获取操作人信息
	operator, err := s.userRepo.FindByID(int64(operatorID))
	if err != nil {
		return nil, errors.New("操作人不存在")
	}

	// 4. 验证订单明细 - 从订单的Items关联获取
	if len(order.Items) == 0 {
		return nil, errors.New("订单没有商品明细")
	}

	// 5. 非打印模式时检查库存充足性
	if !req.PrintMode && req.WarehouseID > 0 {
		insufficientItems, err := s.inventoryService.CheckStockForOrder(int64(req.WarehouseID), order.Items)
		if err != nil {
			return nil, fmt.Errorf("库存检查失败: %v", err)
		}
		if len(insufficientItems) > 0 {
			// 构建库存不足的商品信息
			var msg strings.Builder
			msg.WriteString("库存不足，无法出库:\n")
			for _, item := range insufficientItems {
				msg.WriteString(fmt.Sprintf("- %s: 需要%d，可用%d\n", item.SKUName, item.RequiredQty, item.AvailableQty))
			}
			return nil, errors.New(msg.String())
		}
	}

	// 构建订单明细映射
	orderItemMap := make(map[uint64]*models.OrderItem)
	for i := range order.Items {
		orderItemMap[uint64(order.Items[i].ID)] = &order.Items[i]
	}

	// 验证请求的商品明细
	for _, item := range req.Items {
		orderItem, exists := orderItemMap[item.OrderItemID]
		if !exists {
			return nil, fmt.Errorf("订单明细ID %d 不存在", item.OrderItemID)
		}
		if item.Quantity > orderItem.Quantity {
			return nil, fmt.Errorf("商品 %s 的送货数量不能超过订单数量", orderItem.ProductName)
		}
	}

	// 5. 生成送货单号
	deliveryNo := s.generateDeliveryNo(order.StoreID)

	// 6. 开启事务
	var deliveryRecord *models.DeliveryRecord
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 6.1 创建送货记录
		deliveryTime := req.DeliveryTime
		if deliveryTime.IsZero() {
			deliveryTime = time.Now()
		}
		// 合并备注：出库申请备注 + 用户输入备注
		combinedRemark := outboundRemark
		if req.Remark != "" {
			if combinedRemark != "" {
				combinedRemark = combinedRemark + "；" + req.Remark
			} else {
				combinedRemark = req.Remark
			}
		}

		deliveryRecord = &models.DeliveryRecord{
			StoreID:         uint64(order.StoreID),
			OrderID:         uint64(order.ID),
			OrderNo:         order.OrderNo,
			DeliveryNo:      deliveryNo,
			WarehouseID:     req.WarehouseID,
			OperatorID:      operatorID,
			OperatorName:    operator.RealName,
			DeliveryTime:    deliveryTime,
			DeliveryType:    req.DeliveryType,
			LogisticsNo:     req.LogisticsNo,
			ReceiverName:    req.ReceiverName,
			ReceiverPhone:   req.ReceiverPhone,
			ReceiverAddress: req.ReceiverAddress,
			Remark:          combinedRemark,
			Status:          models.DeliveryStatusNormal,
		}

		if err := tx.Create(deliveryRecord).Error; err != nil {
			return err
		}

		// 6.2 处理商品明细
		var totalQuantity int
		var totalAmount decimal.Decimal
		deliveryItems := make([]models.DeliveryItem, 0)

		// 打印模式：从订单获取所有商品
		// 普通模式：使用请求中的商品明细
		var itemList []struct {
			OrderItemID uint64
			Quantity    int
			OrderItem   *models.OrderItem
		}

		if req.PrintMode {
			// 打印模式：使用订单的所有商品
			for i := range order.Items {
				itemList = append(itemList, struct {
					OrderItemID uint64
					Quantity    int
					OrderItem   *models.OrderItem
				}{
					OrderItemID: uint64(order.Items[i].ID),
					Quantity:    order.Items[i].Quantity,
					OrderItem:   &order.Items[i],
				})
			}
		} else {
			// 普通模式：使用请求中的商品明细
			for _, itemReq := range req.Items {
				orderItem, exists := orderItemMap[itemReq.OrderItemID]
				if !exists {
					return fmt.Errorf("订单明细ID %d 不存在", itemReq.OrderItemID)
				}
				itemList = append(itemList, struct {
					OrderItemID uint64
					Quantity    int
					OrderItem   *models.OrderItem
				}{
					OrderItemID: itemReq.OrderItemID,
					Quantity:    itemReq.Quantity,
					OrderItem:   orderItem,
				})
			}
		}

		for _, item := range itemList {
			orderItem := item.OrderItem

			// 只有非打印模式才执行库存扣减
			if !req.PrintMode && req.WarehouseID > 0 {
				// 执行FIFO库存扣减
				costDetails, err := s.inventoryService.DeductLockedStock(
					int64(req.WarehouseID),
					orderItem.SKUID,
					item.Quantity,
					order.StoreID,
					int64(order.ID),
					int64(operatorID),
				)
				if err != nil {
					return fmt.Errorf("库存扣减失败: %v", err)
				}

				// 计算成本
				var totalCost decimal.Decimal
				for _, detail := range costDetails {
					totalCost = totalCost.Add(detail.TotalCost)
				}

				var unitCost decimal.Decimal
				if item.Quantity > 0 {
					unitCost = totalCost.Div(decimal.NewFromInt(int64(item.Quantity)))
				}

				var deliveryItem models.DeliveryItem
				deliveryItem.DeliveryID = deliveryRecord.ID
				deliveryItem.OrderItemID = item.OrderItemID
				deliveryItem.SkuID = uint64(orderItem.SKUID)
				deliveryItem.ProductName = orderItem.ProductName
				deliveryItem.SkuName = orderItem.SKUName
				deliveryItem.SkuCode = orderItem.SKUCode
				deliveryItem.Quantity = item.Quantity
				deliveryItem.UnitCost = unitCost
				deliveryItem.TotalCost = totalCost

				// 设置批次ID
				if len(costDetails) > 0 && costDetails[0].BatchID > 0 {
					batchID := uint64(costDetails[0].BatchID)
					deliveryItem.BatchID = &batchID
				}

				deliveryItems = append(deliveryItems, deliveryItem)

				// 更新订单明细的成本信息
				if err := tx.Model(&models.OrderItem{}).
					Where("id = ?", item.OrderItemID).
					Updates(map[string]interface{}{
						"unit_cost":  unitCost,
						"total_cost": totalCost,
					}).Error; err != nil {
					return fmt.Errorf("更新订单明细成本失败: %v", err)
				}
			} else {
				// 打印模式或无仓库ID：使用订单中的成本
				unitCost := orderItem.UnitCost
				totalCost := orderItem.TotalCost
				if totalCost.IsZero() && unitCost.GreaterThan(decimal.Zero) && item.Quantity > 0 {
					totalCost = unitCost.Mul(decimal.NewFromInt(int64(item.Quantity)))
				}

				deliveryItem := models.DeliveryItem{
					DeliveryID:  deliveryRecord.ID,
					OrderItemID: item.OrderItemID,
					SkuID:       uint64(orderItem.SKUID),
					ProductName: orderItem.ProductName,
					SkuName:     orderItem.SKUName,
					SkuCode:     orderItem.SKUCode,
					Quantity:    item.Quantity,
					UnitCost:    unitCost,
					TotalCost:   totalCost,
				}
				deliveryItems = append(deliveryItems, deliveryItem)
			}

			itemAmount := orderItem.SalePrice.Mul(decimal.NewFromInt(int64(item.Quantity)))
			totalQuantity += item.Quantity
			totalAmount = totalAmount.Add(itemAmount)
		}

		// 批量创建送货明细
		if len(deliveryItems) > 0 {
			if err := tx.Create(&deliveryItems).Error; err != nil {
				return err
			}
		}

		// 更新送货记录统计
		deliveryRecord.TotalQuantity = totalQuantity
		deliveryRecord.TotalAmount = totalAmount
		if err := tx.Save(deliveryRecord).Error; err != nil {
			return err
		}

		// 6.3 出库确认时不更新 delivery_status，打印送货单时才更新为配送中
		// 标记订单已出库确认
		if err := tx.Model(&models.Order{}).
			Where("id = ?", order.ID).
			Update("outbound_confirmed", true).Error; err != nil {
			return err
		}

		// 6.4 更新订单成本信息（仅非打印模式才重新计算）
		if !req.PrintMode && len(deliveryItems) > 0 {
			var orderTotalCost decimal.Decimal
			for _, item := range deliveryItems {
				orderTotalCost = orderTotalCost.Add(item.TotalCost)
			}
			actualProfit := order.FinalAmount.Sub(orderTotalCost).Sub(order.GiftCost)

			if err := tx.Model(&models.Order{}).
				Where("id = ?", order.ID).
				Updates(map[string]interface{}{
					"total_cost":    orderTotalCost,
					"actual_profit": actualProfit,
				}).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return deliveryRecord, nil
}

// ConfirmDelivery 确认送达
func (s *deliveryService) ConfirmDelivery(orderID int64) error {
	// 1. 获取订单信息
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return errors.New("订单不存在")
	}

	// 2. 验证配送状态（必须是配送中才能确认送达）
	if order.DeliveryStatus != 1 {
		return errors.New("订单未处于配送中状态")
	}

	// 3. 更新订单配送状态为已配送
	if err := s.db.Model(&models.Order{}).
		Where("id = ?", orderID).
		Update("delivery_status", 2).Error; err != nil { // 2-已配送
		return errors.New("更新配送状态失败")
	}

	return nil
}

// GetDeliveryList 获取送货出库列表
func (s *deliveryService) GetDeliveryList(req *dto.DeliveryListRequest) (*PageResult, error) {
	query := &dto.DeliveryListQuery{
		StoreID:    req.StoreID,
		OrderID:    req.OrderID,
		OrderNo:    req.OrderNo,
		OperatorID: req.OperatorID,
		Status:     req.Status,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}

	if !req.StartTime.IsZero() {
		query.StartTime = &req.StartTime
	}
	if !req.EndTime.IsZero() {
		query.EndTime = &req.EndTime
	}

	records, total, err := s.deliveryRepo.GetList(query)
	if err != nil {
		return nil, err
	}

	// 转换为DTO
	dtos := make([]dto.DeliveryDTO, 0, len(records))
	for _, record := range records {
		dtos = append(dtos, s.convertToDTO(&record))
	}

	return &PageResult{
		List:  dtos,
		Total: total,
		Page:  req.Page,
		PageSize:  req.PageSize,
	}, nil
}

// GetDeliveryDetail 获取送货出库详情
func (s *deliveryService) GetDeliveryDetail(deliveryID uint64) (*dto.DeliveryDTO, error) {
	record, err := s.deliveryRepo.GetByID(deliveryID)
	if err != nil {
		return nil, errors.New("送货记录不存在")
	}

	dto := s.convertToDTO(record)
	return &dto, nil
}

// CancelDelivery 作废送货记录
func (s *deliveryService) CancelDelivery(deliveryID uint64, operatorID uint64, remark string) error {
	record, err := s.deliveryRepo.GetByID(deliveryID)
	if err != nil {
		return errors.New("送货记录不存在")
	}

	if record.Status == models.DeliveryStatusCancelled {
		return errors.New("送货记录已作废")
	}

	// TODO: 作废送货记录时是否需要恢复库存？
	// 根据业务需求，作废后可能需要创建退货入库单来恢复库存

	return s.deliveryRepo.Cancel(deliveryID)
}

// GetPendingDeliveryOrders 获取待送货订单列表
func (s *deliveryService) GetPendingDeliveryOrders(req *dto.PendingDeliveryOrderQuery) (*PageResult, error) {
	// 查询已生效(order_status=1)且未配送(delivery_status=0)且非草稿(is_draft=0)的订单
	query := s.db.Model(&models.Order{}).
		Where("order_status = 1 AND delivery_status = 0 AND (is_draft = 0 OR is_draft IS NULL)")

	if req.StoreID > 0 {
		query = query.Where("store_id = ?", req.StoreID)
	}
	if req.OrderNo != "" {
		query = query.Where("order_no LIKE ?", "%"+req.OrderNo+"%")
	}
	if req.CustomerName != "" {
		query = query.Where("customer_name LIKE ?", "%"+req.CustomerName+"%")
	}
	if req.SalesmanID > 0 {
		query = query.Where("salesman_id = ?", req.SalesmanID)
	}

	var total int64
	query.Count(&total)

	var orders []models.Order
	query.Order("id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Where("item_status = 0")
		}).
		Find(&orders)

	// 组装 DTO
	list := make([]dto.PendingDeliveryOrderDTO, 0, len(orders))
	for _, order := range orders {
		item := dto.PendingDeliveryOrderDTO{
			OrderID:         uint64(order.ID),
			OrderNo:         order.OrderNo,
			CustomerName:    order.CustomerName,
			CustomerPhone:   order.CustomerPhone,
			CustomerAddress: order.CustomerAddress,
			SalesmanID:      uint64(order.SalesmanID),
			TotalAmount:     order.FinalAmount,
			TotalQuantity:   order.TotalQuantity,
			OrderTime:       order.CreatedAt,
		}

		// 查询业务员姓名
		if order.SalesmanID > 0 {
			var user models.User
			if s.db.Select("real_name, username").First(&user, order.SalesmanID).Error == nil {
				item.SalesmanName = user.RealName
				if item.SalesmanName == "" {
					item.SalesmanName = user.Username
				}
			}
		}

		// 组装商品明细
		if len(order.Items) > 0 {
			items := make([]dto.PendingDeliveryOrderItemDTO, 0, len(order.Items))
			for _, oi := range order.Items {
				items = append(items, dto.PendingDeliveryOrderItemDTO{
					OrderItemID: uint64(oi.ID),
					SkuID:       uint64(oi.SKUID),
					ProductName: oi.ProductName,
					SkuName:     oi.SKUName,
					SkuCode:     oi.SKUCode,
					Quantity:    oi.Quantity,
					SalePrice:   oi.SalePrice,
				})
			}
			item.Items = items
		}

		list = append(list, item)
	}

	return &PageResult{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// generateDeliveryNo 生成送货单号
func (s *deliveryService) generateDeliveryNo(storeID int64) string {
	// 格式: DL + 年月日 + 门店ID(3位) + 随机数(4位)
	now := time.Now()
	randomNum := rand.Intn(9000) + 1000
	return fmt.Sprintf("DL%s%03d%d", now.Format("20060102"), storeID%1000, randomNum)
}

// convertToDTO 转换为DTO
func (s *deliveryService) convertToDTO(record *models.DeliveryRecord) dto.DeliveryDTO {
	result := dto.DeliveryDTO{
		ID:               record.ID,
		StoreID:          record.StoreID,
		OrderID:          record.OrderID,
		OrderNo:          record.OrderNo,
		DeliveryNo:       record.DeliveryNo,
		WarehouseID:      record.WarehouseID,
		OperatorID:       record.OperatorID,
		OperatorName:     record.OperatorName,
		DeliveryTime:     record.DeliveryTime,
		DeliveryType:     record.DeliveryType,
		DeliveryTypeName: models.GetDeliveryTypeName(record.DeliveryType),
		LogisticsNo:      record.LogisticsNo,
		ReceiverName:     record.ReceiverName,
		ReceiverPhone:    record.ReceiverPhone,
		ReceiverAddress:  record.ReceiverAddress,
		Remark:           record.Remark,
		TotalQuantity:    record.TotalQuantity,
		TotalAmount:      record.TotalAmount,
		Status:           record.Status,
		StatusName:       models.GetDeliveryStatusName(record.Status),
		CreatedAt:        record.CreatedAt,
	}

	// 设置仓库名称
	if record.Warehouse.ID > 0 {
		result.WarehouseName = record.Warehouse.WarehouseName
	}

	// 转换明细
	if len(record.Items) > 0 {
		items := make([]dto.DeliveryItemDTO, 0, len(record.Items))
		for _, item := range record.Items {
			items = append(items, dto.DeliveryItemDTO{
				ID:          item.ID,
				DeliveryID:  item.DeliveryID,
				OrderItemID: item.OrderItemID,
				SkuID:       item.SkuID,
				ProductName: item.ProductName,
				SkuName:     item.SkuName,
				Quantity:    item.Quantity,
				BatchID:     item.BatchID,
				UnitCost:    item.UnitCost,
				TotalCost:   item.TotalCost,
			})
		}
		result.Items = items
	}

	return result
}

// GetOrderStockStatus 获取订单库存状态（含锁定批次信息）
func (s *deliveryService) GetOrderStockStatus(orderID int64, warehouseID int64) ([]map[string]interface{}, error) {
	// 1. 获取订单信息
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, errors.New("订单不存在")
	}

	// 2. 获取库存
	stockMap := make(map[int64]*models.WarehouseStock)
	var stocks []models.WarehouseStock
	s.db.Where("warehouse_id = ?", warehouseID).Find(&stocks)
	for i := range stocks {
		stockMap[stocks[i].SKUID] = &stocks[i]
	}

	// 3. 获取批次绑定信息
	var orderItems []models.OrderItem
	s.db.Where("order_id = ? AND batch_id IS NOT NULL AND batch_id > 0", orderID).Find(&orderItems)
	batchMap := make(map[int64][]map[string]interface{})
	for _, item := range orderItems {
		if item.BatchID != nil && *item.BatchID > 0 {
			var batch models.InventoryBatch
			if err := s.db.First(&batch, *item.BatchID).Error; err == nil {
				batchInfo := map[string]interface{}{
					"batch_id":   batch.ID,
					"batch_no":   batch.BatchNo,
					"quantity":   item.Quantity,
				}
				batchMap[item.SKUID] = append(batchMap[item.SKUID], batchInfo)
			}
		}
	}

	// 4. 构建返回数据
	var result []map[string]interface{}
	for _, item := range order.Items {
		stock := stockMap[item.SKUID]
		availableQty := 0
		lockedQty := 0
		if stock != nil {
			availableQty = stock.AvailableQuantity
			lockedQty = stock.LockedQuantity
		}

		result = append(result, map[string]interface{}{
			"sku_id":        item.SKUID,
			"product_name":  item.ProductName,
			"sku_name":      item.SKUName,
			"quantity":      item.Quantity,
			"available_qty": availableQty,
			"locked_qty":    lockedQty,
			"batches":       batchMap[item.SKUID],
		})
	}

	return result, nil
}

// PrintDelivery 打印送货单时更新配送状态为配送中
func (s *deliveryService) PrintDelivery(orderID int64) error {
	// 1. 查询订单
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("订单不存在")
		}
		return errors.New("查询订单失败")
	}

	// 2. 校验订单状态
	if order.OrderStatus != 1 {
		return errors.New("订单未审核通过")
	}

	// 3. 更新配送状态为配送中（delivery_status = 1）
	if err := s.db.Model(&models.Order{}).
		Where("id = ?", orderID).
		Update("delivery_status", 1).Error; err != nil {
		return errors.New("更新配送状态失败")
	}

	return nil
}
