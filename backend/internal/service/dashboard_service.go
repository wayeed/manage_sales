package service

import (
	"time"

	"furniture-commission/internal/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// DashboardOverview 仪表盘概览数据
type DashboardOverview struct {
	TotalSales     decimal.Decimal `json:"total_sales"`
	TotalOrders    int64           `json:"total_orders"`
	TotalProfit    decimal.Decimal `json:"total_profit"`
	PendingOrders  int64           `json:"pending_orders"`
	PendingPayments int64          `json:"pending_payments"`
	UnpaidAmount   decimal.Decimal `json:"unpaid_amount"`
	LowStockCount  int64           `json:"low_stock_count"`
	TodayOrders    int64           `json:"today_orders"`
}

// DashboardService 仪表盘服务
type DashboardService struct {
	db *gorm.DB
}

// NewDashboardService 创建仪表盘服务实例
func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{db: db}
}

// GetOverview 获取仪表盘概览数据
func (s *DashboardService) GetOverview(storeID int64) (*DashboardOverview, error) {
	overview := &DashboardOverview{}

	now := time.Now()
	// 本月起止
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)
	// 今日起止
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24 * time.Hour).Add(-time.Second)

	// 构建基础查询条件
	orderQuery := s.db.Model(&models.Order{}).Where("store_id = ?", storeID)

	// 本月销售总额（已生效订单）
	var totalSales decimal.Decimal
	orderQuery.Where("order_status = 1 AND order_date BETWEEN ? AND ?", monthStart, monthEnd).
		Select("COALESCE(SUM(final_amount), 0)").Scan(&totalSales)
	overview.TotalSales = totalSales

	// 本月订单数（已生效）
	orderQuery.Session(&gorm.Session{}).Where("order_status = 1 AND order_date BETWEEN ? AND ?", monthStart, monthEnd).
		Select("COALESCE(COUNT(*), 0)").Scan(&overview.TotalOrders)

	// 本月利润总额
	var totalProfit decimal.Decimal
	orderQuery.Session(&gorm.Session{}).Where("order_status = 1 AND order_date BETWEEN ? AND ?", monthStart, monthEnd).
		Select("COALESCE(SUM(actual_profit), 0)").Scan(&totalProfit)
	overview.TotalProfit = totalProfit

	// 待审批订单数
	orderQuery.Session(&gorm.Session{}).Where("order_status = 0").
		Select("COALESCE(COUNT(*), 0)").Scan(&overview.PendingOrders)

	// 待审核回款数
	s.db.Model(&models.Payment{}).
		Joins("JOIN orders ON orders.id = payments.order_id").
		Where("orders.store_id = ? AND payments.status = 0", storeID).
		Select("COALESCE(COUNT(*), 0)").Scan(&overview.PendingPayments)

	// 未回款总额（已生效订单的 remaining_amount 之和）
	var unpaidAmount decimal.Decimal
	s.db.Model(&models.Order{}).
		Where("store_id = ? AND order_status = 1 AND payment_status != 2 AND remaining_amount > 0", storeID).
		Select("COALESCE(SUM(remaining_amount), 0)").Scan(&unpaidAmount)
	overview.UnpaidAmount = unpaidAmount

	// 库存预警数量（未处理的预警）
	s.db.Model(&models.StockAlert{}).
		Where("store_id = ? AND alert_status = 0", storeID).
		Select("COALESCE(COUNT(*), 0)").Scan(&overview.LowStockCount)

	// 今日新增订单数
	orderQuery.Session(&gorm.Session{}).Where("created_at BETWEEN ? AND ?", todayStart, todayEnd).
		Select("COALESCE(COUNT(*), 0)").Scan(&overview.TodayOrders)

	return overview, nil
}
