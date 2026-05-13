package service

import (
	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	"furniture-commission/internal/repository"

	"gorm.io/gorm"
)

// HandleAlertRequest 处理预警请求
type HandleAlertRequest struct {
	AlertStatus int8 `json:"alert_status" binding:"required" example:1`
	Remark string `json:"remark" example:"已补货"`
}

// ListAlertRequest 预警列表查询请求
type ListAlertRequest struct {
	StoreID int64 `form:"store_id" example:1`
	WarehouseID *int64 `form:"warehouse_id" example:1`
	AlertType *int8 `form:"alert_type" example:1`
	AlertStatus *int8 `form:"alert_status" example:0`
	Page int `form:"page" example:1`
	PageSize int `form:"page_size" example:10`
}

// StockAlertService 库存预警服务
type StockAlertService struct {
	db       *gorm.DB
	alertRepo *repository.BaseRepository[models.StockAlert]
	invRepo  *repository.InventoryRepository
}

// NewStockAlertService 创建库存预警服务实例
func NewStockAlertService(db *gorm.DB, invRepo *repository.InventoryRepository) *StockAlertService {
	return &StockAlertService{
		db:        db,
		alertRepo: repository.NewBaseRepository[models.StockAlert](db),
		invRepo:   invRepo,
	}
}

// CheckAlerts 检查库存预警（对比warehouse_stocks.stock_quantity和warning_stock）
func (s *StockAlertService) CheckAlerts(storeID int64) (int, error) {
	// 查询所有仓库
	var warehouses []models.Warehouse
	if err := s.db.Where("store_id = ? AND status = 1", storeID).Find(&warehouses).Error; err != nil {
		return 0, &AppError{Code: apperrors.InternalError, Message: "查询仓库列表失败"}
	}

	alertCount := 0

	for _, wh := range warehouses {
		// 检查商品库存预警
		var stocks []models.WarehouseStock
		if err := s.db.Where("warehouse_id = ? AND stock_quantity <= warning_stock", wh.ID).Find(&stocks).Error; err != nil {
			continue
		}

		for _, stock := range stocks {
			// 检查是否已有未处理的预警
			var existingAlert int64
			s.db.Model(&models.StockAlert{}).
				Where("warehouse_id = ? AND sku_id = ? AND alert_type = 1 AND alert_status = 0", wh.ID, stock.SKUID).
				Count(&existingAlert)

			if existingAlert > 0 {
				continue
			}

			alert := &models.StockAlert{
				StoreID:      storeID,
				WarehouseID:  &wh.ID,
				AlertType:    1, // 商品库存不足
				SKUID:        &stock.SKUID,
				CurrentStock: stock.StockQuantity,
				WarningStock: stock.WarningStock,
				AlertStatus:  0, // 未处理
			}
			if err := s.db.Create(alert).Error; err == nil {
				alertCount++
			}
		}

		// 检查礼品库存预警
		var giftStocks []models.WarehouseGiftStock
		if err := s.db.Where("warehouse_id = ? AND stock_quantity <= warning_stock", wh.ID).Find(&giftStocks).Error; err != nil {
			continue
		}

		for _, gs := range giftStocks {
			var existingAlert int64
			s.db.Model(&models.StockAlert{}).
				Where("warehouse_id = ? AND gift_id = ? AND alert_type = 2 AND alert_status = 0", wh.ID, gs.GiftID).
				Count(&existingAlert)

			if existingAlert > 0 {
				continue
			}

			alert := &models.StockAlert{
				StoreID:      storeID,
				WarehouseID:  &wh.ID,
				AlertType:    2, // 礼品库存不足
				GiftID:       &gs.GiftID,
				CurrentStock: gs.StockQuantity,
				WarningStock: gs.WarningStock,
				AlertStatus:  0,
			}
			if err := s.db.Create(alert).Error; err == nil {
				alertCount++
			}
		}
	}

	return alertCount, nil
}

// List 获取预警列表
func (s *StockAlertService) List(req *ListAlertRequest) (*PageResult, error) {
	db := s.db.Model(&models.StockAlert{})

	if req.StoreID > 0 {
		db = db.Where("store_id = ?", req.StoreID)
	}
	if req.WarehouseID != nil {
		db = db.Where("warehouse_id = ?", *req.WarehouseID)
	}
	if req.AlertType != nil {
		db = db.Where("alert_type = ?", *req.AlertType)
	}
	if req.AlertStatus != nil {
		db = db.Where("alert_status = ?", *req.AlertStatus)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询预警列表失败"}
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

	var alerts []models.StockAlert
	if err := db.Preload("Warehouse").Preload("SKU").Preload("Gift").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").
		Find(&alerts).Error; err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询预警列表失败"}
	}

	return &PageResult{
		List:     alerts,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Handle 处理预警
func (s *StockAlertService) Handle(id int64, req *HandleAlertRequest, handledBy int64) error {
	alert, err := s.alertRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &AppError{Code: apperrors.NotFound, Message: "预警记录不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if alert.AlertStatus != 0 {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "预警已处理"}
	}

	now := gorm.Expr("NOW()")
	if err := s.db.Model(alert).Updates(map[string]interface{}{
		"alert_status": req.AlertStatus,
		"handled_by":   handledBy,
		"handled_at":   now,
		"remark":       req.Remark,
	}).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "处理预警失败"}
	}

	return nil
}
