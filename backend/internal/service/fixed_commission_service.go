package service

import (
	"fmt"
	"furniture-commission/internal/models"
	"furniture-commission/internal/repository"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// FixedCommissionService 固定提成服务
type FixedCommissionService struct {
	db             *gorm.DB
	commissionRepo *repository.CommissionRepository
	orderRepo      *repository.OrderRepository
	configService  *ConfigService
}

// NewFixedCommissionService 创建固定提成服务实例
func NewFixedCommissionService(
	db *gorm.DB,
	commissionRepo *repository.CommissionRepository,
	orderRepo *repository.OrderRepository,
	configService *ConfigService,
) *FixedCommissionService {
	return &FixedCommissionService{
		db:             db,
		commissionRepo: commissionRepo,
		orderRepo:      orderRepo,
		configService:  configService,
	}
}

// CalculateMonthlyFixedCommission 计算月度固定提成
// 每月执行，计算指定月份的固定提成
// 固定提成只计算业务员自己订单的当月回款总额（排除同行带单）
func (s *FixedCommissionService) CalculateMonthlyFixedCommission(periodValue string) error {
	// 1. 获取固定提成比例
	rates, err := s.configService.GetCommissionRates()
	if err != nil {
		return fmt.Errorf("获取固定提成比例失败: %w", err)
	}
	fixedRate := rates["fixed_commission_rate"]
	if fixedRate.IsZero() {
		fixedRate = decimal.NewFromFloat(0.05) // 默认5%
	}

	// 2. 查询指定月份有回款的业务员（只统计业务员自己的订单回款，排除同行带单）
	type SalesmanPayment struct {
		SalesmanID   int64
		StoreID      int64
		TotalPayment decimal.Decimal
	}

	var results []SalesmanPayment
	err = s.db.Raw(`
		SELECT o.salesman_id, o.store_id, SUM(p.amount) as total_payment
		FROM payments p
		JOIN orders o ON p.order_id = o.id
		WHERE DATE_FORMAT(p.payment_date, '%Y-%m') = ?
		  AND o.is_peer_order = 0
		  AND o.order_status = 1
		  AND p.status = 1
		GROUP BY o.salesman_id, o.store_id
	`, periodValue).Scan(&results).Error

	if err != nil {
		return fmt.Errorf("查询业务员回款数据失败: %w", err)
	}

	if len(results) == 0 {
		return nil // 没有数据，直接返回
	}

	// 3. 为每个业务员创建固定提成记录
	for _, result := range results {
		// 检查是否已存在该业务员的固定提成记录
		exists, err := s.commissionRepo.ExistsByEmployeeAndPeriod(
			result.SalesmanID,
			7, // commission_type = 7 固定提成
			periodValue,
		)
		if err != nil {
			continue
		}
		if exists {
			// 已存在则跳过（避免重复计算）
			continue
		}

		fixedAmount := result.TotalPayment.Mul(fixedRate)

		commission := &models.Commission{
			StoreID:        result.StoreID,
			OrderID:        0, // 固定提成不关联具体订单
			EmployeeID:     &result.SalesmanID,
			CommissionType: 7, // 7 = 固定提成
			PeriodValue:    periodValue,
			BaseAmount:     result.TotalPayment,
			Rate:           fixedRate,
			Amount:         fixedAmount,
			Status:         1, // 可发放
			Remark:         fmt.Sprintf("%s月度固定提成（自己订单回款）", periodValue),
		}

		if err := s.commissionRepo.Create(commission); err != nil {
			// 记录错误但继续处理其他业务员
			fmt.Printf("创建业务员%d的固定提成记录失败: %v\n", result.SalesmanID, err)
		}
	}

	return nil
}

// CalculateLastMonthFixedCommission 计算上个月的固定提成
// 供定时任务调用
func (s *FixedCommissionService) CalculateLastMonthFixedCommission() error {
	// 获取上个月份
	now := time.Now()
	lastMonth := now.AddDate(0, -1, 0)
	periodValue := lastMonth.Format("2006-01")

	return s.CalculateMonthlyFixedCommission(periodValue)
}
