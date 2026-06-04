package service

import (
	"time"

	"furniture-commission/internal/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// DashboardOverview 仪表盘概览数据
type DashboardOverview struct {
	TotalSales      decimal.Decimal `json:"totalSales"`
	TotalOrders     int64          `json:"totalOrders"`
	TotalProfit     decimal.Decimal `json:"totalProfit"`
	PendingApproval int64          `json:"pendingApproval"`
	PendingPayments int64          `json:"pendingPayments"`
	UnpaidAmount    decimal.Decimal `json:"unpaidAmount"`
	LowStockCount   int64          `json:"lowStockCount"`
	TodayOrders     int64          `json:"todayOrders"`
}

// DashboardService 仪表盘服务
type DashboardService struct {
	db      *gorm.DB
	userSvc *UserService
}

// NewDashboardService 创建仪表盘服务实例
func NewDashboardService(db *gorm.DB, userSvc *UserService) *DashboardService {
	return &DashboardService{db: db, userSvc: userSvc}
}

// GetOverview 获取仪表盘概览数据
func (s *DashboardService) GetOverview(storeID int64, userID int64, roleCodes []string) (*DashboardOverview, error) {
	overview := &DashboardOverview{}

	now := time.Now()
	// 本月起止
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)
	// 今日起止
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24 * time.Hour).Add(-time.Second)

	// 根据角色确定查询范围
	var salesmanIDs []int64
	var orderQuery *gorm.DB

	if hasRole(roleCodes, "STORE_MANAGER") || hasRole(roleCodes, "FINANCE") || hasRole(roleCodes, "BOSS") {
		// 老板、店长、财务：查看门店所有数据
		orderQuery = s.db.Model(&models.Order{}).Where("store_id = ?", storeID)
	} else if hasRole(roleCodes, "SUPERVISOR") {
		// 主管：查看自己及直属下级的数据
		subordinateIDs, err := s.userSvc.GetDirectSubordinateIDs(userID)
		if err != nil {
			subordinateIDs = []int64{}
		}
		salesmanIDs = append(subordinateIDs, userID)
		orderQuery = s.db.Model(&models.Order{}).Where("store_id = ? AND salesman_id IN ?", storeID, salesmanIDs)
	} else {
		// 业务员：只查看自己的数据
		orderQuery = s.db.Model(&models.Order{}).Where("store_id = ? AND salesman_id = ?", storeID, userID)
	}

	// 本月销售总额（已生效订单）
	var totalSales decimal.Decimal
	orderQuery.Session(&gorm.Session{}).Where("order_status = 1 AND order_date BETWEEN ? AND ?", monthStart, monthEnd).
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

	// 待审批订单数（根据角色范围，排除草稿）
	if hasRole(roleCodes, "STORE_MANAGER") || hasRole(roleCodes, "FINANCE") || hasRole(roleCodes, "BOSS") {
		// 店长/财务/老板：查看门店所有待审批
		s.db.Model(&models.Order{}).
			Where("store_id = ? AND order_status = 0 AND is_draft = 0", storeID).
			Select("COALESCE(COUNT(*), 0)").Scan(&overview.PendingApproval)
	} else if hasRole(roleCodes, "SUPERVISOR") {
		// 主管：查看自己及下级的待审批
		s.db.Model(&models.Order{}).
			Where("store_id = ? AND order_status = 0 AND is_draft = 0 AND salesman_id IN ?", storeID, salesmanIDs).
			Select("COALESCE(COUNT(*), 0)").Scan(&overview.PendingApproval)
	} else {
		// 业务员：不显示待审批（自己不能审批自己的订单）
		overview.PendingApproval = 0
	}

	// 待审核回款数（根据角色范围）
	if hasRole(roleCodes, "STORE_MANAGER") || hasRole(roleCodes, "FINANCE") || hasRole(roleCodes, "BOSS") {
		s.db.Model(&models.Payment{}).
			Joins("JOIN orders ON orders.id = payments.order_id").
			Where("orders.store_id = ? AND payments.status = 0", storeID).
			Select("COALESCE(COUNT(*), 0)").Scan(&overview.PendingPayments)
	} else if hasRole(roleCodes, "SUPERVISOR") {
		s.db.Model(&models.Payment{}).
			Joins("JOIN orders ON orders.id = payments.order_id").
			Where("orders.store_id = ? AND payments.status = 0 AND orders.salesman_id IN ?", storeID, salesmanIDs).
			Select("COALESCE(COUNT(*), 0)").Scan(&overview.PendingPayments)
	} else {
		overview.PendingPayments = 0
	}

	// 未回款总额（根据角色范围）
	if hasRole(roleCodes, "STORE_MANAGER") || hasRole(roleCodes, "FINANCE") || hasRole(roleCodes, "BOSS") {
		var unpaidAmount decimal.Decimal
		s.db.Model(&models.Order{}).
			Where("store_id = ? AND order_status = 1 AND payment_status != 2 AND remaining_amount > 0", storeID).
			Select("COALESCE(SUM(remaining_amount), 0)").Scan(&unpaidAmount)
		overview.UnpaidAmount = unpaidAmount
	} else if hasRole(roleCodes, "SUPERVISOR") {
		var unpaidAmount decimal.Decimal
		s.db.Model(&models.Order{}).
			Where("store_id = ? AND order_status = 1 AND payment_status != 2 AND remaining_amount > 0 AND salesman_id IN ?", storeID, salesmanIDs).
			Select("COALESCE(SUM(remaining_amount), 0)").Scan(&unpaidAmount)
		overview.UnpaidAmount = unpaidAmount
	} else {
		var unpaidAmount decimal.Decimal
		s.db.Model(&models.Order{}).
			Where("store_id = ? AND order_status = 1 AND payment_status != 2 AND remaining_amount > 0 AND salesman_id = ?", storeID, userID).
			Select("COALESCE(SUM(remaining_amount), 0)").Scan(&unpaidAmount)
		overview.UnpaidAmount = unpaidAmount
	}

	// 库存预警数量（仅店长/财务/老板可见）
	if hasRole(roleCodes, "STORE_MANAGER") || hasRole(roleCodes, "FINANCE") || hasRole(roleCodes, "BOSS") {
		s.db.Model(&models.StockAlert{}).
			Where("store_id = ? AND alert_status = 0", storeID).
			Select("COALESCE(COUNT(*), 0)").Scan(&overview.LowStockCount)
	} else {
		overview.LowStockCount = 0
	}

	// 今日新增订单数（根据角色范围）
	orderQuery.Session(&gorm.Session{}).Where("created_at BETWEEN ? AND ?", todayStart, todayEnd).
		Select("COALESCE(COUNT(*), 0)").Scan(&overview.TodayOrders)

	return overview, nil
}
