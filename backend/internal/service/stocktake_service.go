package service

import (
	"errors"
	"fmt"

	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	appsnow "furniture-commission/internal/pkg/snowflake"
	"furniture-commission/internal/repository"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type StocktakeService struct {
	db              *gorm.DB
	stocktakeRepo   *repository.StocktakeRepository
	inventoryRepo   *repository.InventoryRepository
}

func NewStocktakeService(db *gorm.DB, stocktakeRepo *repository.StocktakeRepository, inventoryRepo *repository.InventoryRepository) *StocktakeService {
	return &StocktakeService{db: db, stocktakeRepo: stocktakeRepo, inventoryRepo: inventoryRepo}
}

// --- 请求结构体 ---

type CreateStocktakeRequest struct {
	WarehouseID int64  `json:"warehouse_id" binding:"required"`
	Remark      string `json:"remark"`
}

type UpdateStocktakeRequest struct {
	Items []StocktakeItemRequest `json:"items" binding:"required"`
	Remark string                `json:"remark"`
}

type StocktakeItemRequest struct {
	SKUID       int64  `json:"sku_id" binding:"required"`
	SystemStock int    `json:"system_stock"`
	ActualStock int    `json:"actual_stock" binding:"required"`
	Remark      string `json:"remark"`
}

type ListStocktakeRequest struct {
	WarehouseID int64   `form:"warehouse_id"`
	Status      *int8   `form:"status"`
	Keyword     string  `form:"keyword"`
	Page        int     `form:"page"`
	PageSize    int     `form:"page_size"`
}

// --- 服务方法 ---

// Create 创建盘点单（盘点中状态），自动加载该仓库所有有库存的SKU
func (s *StocktakeService) Create(req *CreateStocktakeRequest, storeID, createdBy int64) (*models.Stocktake, error) {
	stocktakeNo := "ST" + appsnow.GenerateOrderNo()

	stocktake := &models.Stocktake{
		StoreID:     storeID,
		WarehouseID: req.WarehouseID,
		StocktakeNo: stocktakeNo,
		Status:      0, // 盘点中
		Remark:      req.Remark,
		CreatedBy:   &createdBy,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(stocktake).Error; err != nil {
			return err
		}

		// 加载该仓库所有有库存的SKU作为盘点明细
		var stocks []models.WarehouseStock
		if err := tx.Where("warehouse_id = ? AND stock_quantity > 0", req.WarehouseID).
			Preload("SKU").
			Find(&stocks).Error; err != nil {
			return err
		}

		if len(stocks) > 0 {
			items := make([]models.StocktakeItem, 0, len(stocks))
			for _, stock := range stocks {
				items = append(items, models.StocktakeItem{
					StocktakeID: stocktake.ID,
					SKUID:       stock.SKUID,
					SystemStock: stock.StockQuantity,
					ActualStock: stock.StockQuantity, // 默认等于系统库存
					DiffQuantity: 0,
				})
			}
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
			stocktake.TotalItems = len(items)
		}

		return nil
	})

	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "创建盘点单失败"}
	}

	return s.stocktakeRepo.FindByID(stocktake.ID)
}

// GetDetail 获取盘点单详情
func (s *StocktakeService) GetDetail(id int64) (*models.Stocktake, error) {
	stocktake, err := s.stocktakeRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.NotFound, Message: "盘点单不存在"}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}
	return stocktake, nil
}

// List 盘点单列表
func (s *StocktakeService) List(req *ListStocktakeRequest, storeID int64) (*PageResult, error) {
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	list, total, err := s.stocktakeRepo.List(storeID, req.WarehouseID, req.Status, req.Keyword, page, pageSize)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询盘点单列表失败"}
	}

	return &PageResult{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Update 更新盘点单（修改实际盘点数量）
func (s *StocktakeService) Update(id int64, req *UpdateStocktakeRequest, userID int64) (*models.Stocktake, error) {
	stocktake, err := s.stocktakeRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.NotFound, Message: "盘点单不存在"}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if stocktake.Status != 0 {
		return nil, &AppError{Code: apperrors.BadRequest, Message: "只有盘点中的单据才能修改"}
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 删除旧明细
		if err := tx.Where("stocktake_id = ?", id).Delete(&models.StocktakeItem{}).Error; err != nil {
			return err
		}

		// 创建新明细
		profitItems := 0
		lossItems := 0
		items := make([]models.StocktakeItem, 0, len(req.Items))
		for _, item := range req.Items {
			diff := item.ActualStock - item.SystemStock
			if diff > 0 {
				profitItems++
			} else if diff < 0 {
				lossItems++
			}
			items = append(items, models.StocktakeItem{
				StocktakeID:  id,
				SKUID:        item.SKUID,
				SystemStock:  item.SystemStock,
				ActualStock:  item.ActualStock,
				DiffQuantity: diff,
				Remark:       item.Remark,
			})
		}

		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}

		// 更新盘点单
		return tx.Model(stocktake).Updates(map[string]interface{}{
			"total_items":  len(items),
			"profit_items": profitItems,
			"loss_items":   lossItems,
			"remark":       req.Remark,
		}).Error
	})

	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "更新盘点单失败"}
	}

	return s.stocktakeRepo.FindByID(id)
}

// Submit 提交盘点单（状态从盘点中 -> 已提交）
func (s *StocktakeService) Submit(id int64, userID int64) error {
	stocktake, err := s.stocktakeRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "盘点单不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if stocktake.Status != 0 {
		return &AppError{Code: apperrors.BadRequest, Message: "只有盘点中的单据才能提交"}
	}

	return s.db.Model(stocktake).Update("status", 1).Error
}

// Approve 审核盘点单，生成盘盈/盘亏流水并调整库存
func (s *StocktakeService) Approve(id int64, approvedBy int64, approved bool) error {
	stocktake, err := s.stocktakeRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "盘点单不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if stocktake.Status != 1 {
		return &AppError{Code: apperrors.BadRequest, Message: "只有已提交的单据才能审核"}
	}

	if !approved {
		// 驳回：退回到盘点中
		return s.db.Model(stocktake).Update("status", 0).Error
	}

	// 审核通过：生成盘盈/盘亏流水并调整库存
	items, err := s.stocktakeRepo.GetItemsByStocktakeID(id)
	if err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "获取盘点明细失败"}
	}

	now := gorm.Expr("NOW()")

	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if item.DiffQuantity == 0 {
				continue
			}

			// 查询当前库存
			stock, err := s.inventoryRepo.FindStockByWarehouseAndSKU(stocktake.WarehouseID, item.SKUID)
			if err != nil {
				// 如果库存记录不存在，创建一条
				if errors.Is(err, gorm.ErrRecordNotFound) {
					stock = &models.WarehouseStock{
						WarehouseID: stocktake.WarehouseID,
						SKUID:       item.SKUID,
					}
					if err := tx.Create(stock).Error; err != nil {
						return fmt.Errorf("创建库存记录失败: %w", err)
					}
				} else {
					return fmt.Errorf("查询库存失败: %w", err)
				}
			}

			beforeStock := stock.StockQuantity
			afterStock := beforeStock + item.DiffQuantity

			var transactionType int8
			var remark string
			if item.DiffQuantity > 0 {
				transactionType = models.TransactionTypeProfit
				remark = fmt.Sprintf("盘点盘盈[%s]", stocktake.StocktakeNo)
			} else {
				transactionType = models.TransactionTypeLoss
				remark = fmt.Sprintf("盘点盘亏[%s]", stocktake.StocktakeNo)
			}

			// 更新库存
			if err := tx.Model(stock).Updates(map[string]interface{}{
				"stock_quantity":     afterStock,
				"available_quantity": afterStock,
				"version":            gorm.Expr("version + 1"),
			}).Error; err != nil {
				return fmt.Errorf("更新库存失败: %w", err)
			}

			// 记录流水
			txItem := models.InventoryTransaction{
				StoreID:         stocktake.StoreID,
				WarehouseID:     &stocktake.WarehouseID,
				TransactionType: transactionType,
				BizType:         1, // 商品
				BizID:           &item.SKUID,
				Quantity:        item.DiffQuantity,
				BeforeStock:     beforeStock,
				AfterStock:      afterStock,
				UnitCost:        decimal.Zero,
				TotalCost:       decimal.Zero,
				Remark:          remark,
				CreatedBy:       &approvedBy,
			}
			if err := tx.Create(&txItem).Error; err != nil {
				return fmt.Errorf("创建库存流水失败: %w", err)
			}
		}

		// 更新盘点单状态
		return tx.Model(stocktake).Updates(map[string]interface{}{
			"status":      2,
			"approved_by": approvedBy,
			"approved_at": now,
		}).Error
	})

	if err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "审核盘点单失败: " + err.Error()}
	}

	return nil
}

// Delete 删除盘点单（仅盘点中状态可删除）
func (s *StocktakeService) Delete(id int64) error {
	stocktake, err := s.stocktakeRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "盘点单不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if stocktake.Status != 0 {
		return &AppError{Code: apperrors.BadRequest, Message: "只有盘点中的单据才能删除"}
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("stocktake_id = ?", id).Delete(&models.StocktakeItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(stocktake).Error
	})
}
