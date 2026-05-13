package service

import (
	"errors"
	"time"

	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	"furniture-commission/internal/repository"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ListFundPoolRequest 基金池列表查询请求
type ListFundPoolRequest struct {
	StoreID string `form:"store_id" example:"1"`
	PeriodType int `form:"period_type" example:1`
	Page int `form:"page" example:1`
	PageSize int `form:"page_size" example:10`
}

// FundPoolService 基金池服务
type FundPoolService struct {
	db             *gorm.DB
	fundPoolRepo   *repository.FundPoolRepository
	commissionRepo *repository.CommissionRepository
	configService  *ConfigService
}

// NewFundPoolService 创建基金池服务实例
func NewFundPoolService(
	db *gorm.DB,
	fundPoolRepo *repository.FundPoolRepository,
	commissionRepo *repository.CommissionRepository,
	configService *ConfigService,
) *FundPoolService {
	return &FundPoolService{
		db:             db,
		fundPoolRepo:   fundPoolRepo,
		commissionRepo: commissionRepo,
		configService:  configService,
	}
}

// SettleFundPool 月度/季度/年度基金池结算
func (s *FundPoolService) SettleFundPool(storeID int64, periodType int, periodValue string) error {
	// 检查是否已结算
	existing, err := s.fundPoolRepo.FindByPeriod(storeID, periodType, periodValue)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return &AppError{Code: apperrors.InternalError, Message: "查询基金池失败"}
	}
	if existing != nil && existing.Status == 1 {
		return &AppError{Code: apperrors.BadRequest, Message: "该周期基金池已结算"}
	}

	// 获取提成比例配置
	rates, err := s.configService.GetCommissionRates()
	if err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "获取提成比例配置失败"}
	}
	extractRate := rates["fund_pool_extract_rate"]

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 汇总该周期内所有commission_type=5(基金池)的提成记录
		var fundCommissions []models.Commission
		err := tx.Where("store_id = ? AND commission_type = 5 AND period_value = ? AND status = 1", storeID, periodValue).
			Find(&fundCommissions).Error
		if err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "查询基金池提成记录失败"}
		}

		if len(fundCommissions) == 0 {
			return &AppError{Code: apperrors.BadRequest, Message: "该周期无基金池提成记录"}
		}

		// 2. 计算pool_amount
		poolAmount := decimal.Zero
		for _, c := range fundCommissions {
			poolAmount = poolAmount.Add(c.Amount)
		}

		// 3. 汇总该周期内所有员工的个人利润
		// 通过commission_type=1(业务员提成)的base_amount汇总
		type EmployeeProfit struct {
			EmployeeID     int64
			PersonalProfit decimal.Decimal
		}
		var employeeProfits []EmployeeProfit
		err = tx.Model(&models.Commission{}).
			Select("employee_id, COALESCE(SUM(base_amount), 0) as personal_profit").
			Where("store_id = ? AND commission_type = 1 AND period_value = ? AND status = 1", storeID, periodValue).
			Group("employee_id").
			Scan(&employeeProfits).Error
		if err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "汇总员工利润失败"}
		}

		// 4. 计算总利润
		totalProfit := decimal.Zero
		for _, ep := range employeeProfits {
			totalProfit = totalProfit.Add(ep.PersonalProfit)
		}

		// 5. 计算每个员工的份额
		var shares []models.FundPoolShare
		totalShares := decimal.Zero
		perShareAmount := decimal.Zero

		if totalProfit.GreaterThan(decimal.Zero) {
			for _, ep := range employeeProfits {
				shareRatio := ep.PersonalProfit.Div(totalProfit)
				shares = append(shares, models.FundPoolShare{
					FundPoolID:     0, // 后续更新
					EmployeeID:     ep.EmployeeID,
					PersonalProfit: ep.PersonalProfit,
					Shares:         shareRatio,
					RewardAmount:   poolAmount.Mul(shareRatio),
					Status:         0, // 待发放
				})
				totalShares = totalShares.Add(shareRatio)
			}

			if totalShares.GreaterThan(decimal.Zero) {
				perShareAmount = poolAmount.Div(totalShares)
			}
		}

		// 6. 创建或更新fund_pools记录
		now := time.Now()
		fundPool := &models.FundPool{
			StoreID:        storeID,
			PeriodType:     int8(periodType),
			PeriodValue:    periodValue,
			TotalProfit:    totalProfit,
			ExtractRate:    extractRate,
			PoolAmount:     poolAmount,
			TotalShares:    totalShares,
			PerShareAmount: perShareAmount,
			Status:         1, // 已结算
			SettledAt:      &now,
		}

		if existing != nil {
			fundPool.ID = existing.ID
			if err := tx.Model(fundPool).Where("id = ?", existing.ID).Updates(map[string]interface{}{
				"total_profit":    totalProfit,
				"extract_rate":    extractRate,
				"pool_amount":     poolAmount,
				"total_shares":    totalShares,
				"per_share_amount": perShareAmount,
				"status":          1,
				"settled_at":      now,
			}).Error; err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "更新基金池失败"}
			}
		} else {
			if err := tx.Create(fundPool).Error; err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "创建基金池失败"}
			}
		}

		// 7. 创建fund_pool_shares记录
		if len(shares) > 0 {
			// 删除旧的份额记录
			if existing != nil {
				tx.Where("fund_pool_id = ?", fundPool.ID).Delete(&models.FundPoolShare{})
			}
			for i := range shares {
				shares[i].FundPoolID = fundPool.ID
			}
			if err := tx.Create(&shares).Error; err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "创建基金池份额失败"}
			}
		}

		// 8. 更新commission记录状态为已结算（status=1保持不变，基金池提成仍为可发放状态）
		// 基金池提成记录保持status=1，后续由工资发放时统一更新为status=2

		return nil
	})
}

// List 基金池列表
func (s *FundPoolService) List(req *ListFundPoolRequest) (*PageResult, error) {
	fundPools, total, err := s.fundPoolRepo.ListWithFilter(req.StoreID, req.PeriodType, req.Page, req.PageSize)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询基金池列表失败"}
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
		List:     fundPools,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetShares 基金池份额详情
func (s *FundPoolService) GetShares(fundPoolID int64) (*models.FundPool, error) {
	fundPool, err := s.fundPoolRepo.FindByID(fundPoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.NotFound, Message: "基金池不存在"}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询基金池失败"}
	}
	return fundPool, nil
}
