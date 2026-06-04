package service

import (
	"fmt"
	"strings"
	"time"

	"furniture-commission/internal/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ==================== 报表数据结构 ====================

// SalesSummary 销售总览
type SalesSummary struct {
	TotalSales    decimal.Decimal `json:"total_sales"`
	TotalOrders   int64           `json:"total_orders"`
	AvgOrderValue decimal.Decimal `json:"avg_order_value"`
	ProfitRate    decimal.Decimal `json:"profit_rate"`
	TotalProfit   decimal.Decimal `json:"total_profit"`
	TotalCost     decimal.Decimal `json:"total_cost"`
	// 同比（与去年同期对比）
	YoySalesGrowth    *decimal.Decimal `json:"yoy_sales_growth"`
	YoyOrdersGrowth   *decimal.Decimal `json:"yoy_orders_growth"`
	YoyProfitGrowth   *decimal.Decimal `json:"yoy_profit_growth"`
	// 环比（与上期对比）
	MomSalesGrowth    *decimal.Decimal `json:"mom_sales_growth"`
	MomOrdersGrowth   *decimal.Decimal `json:"mom_orders_growth"`
	MomProfitGrowth   *decimal.Decimal `json:"mom_profit_growth"`
}

// TrendData 趋势数据点
type TrendData struct {
	Date        string          `json:"date"`
	SalesAmount decimal.Decimal `json:"sales_amount"`
	OrderCount  int64           `json:"order_count"`
	Profit      decimal.Decimal `json:"profit"`
}

// RankingItem 排行项
type RankingItem struct {
	Name        string          `json:"name"`
	ID          int64           `json:"id"`
	SalesAmount decimal.Decimal `json:"sales_amount"`
	OrderCount  int64           `json:"order_count"`
	Profit      decimal.Decimal `json:"profit"`
	ProfitRate  decimal.Decimal `json:"profit_rate"`
}

// ProfitAnalysis 利润分析
type ProfitAnalysis struct {
	TotalProfit       decimal.Decimal     `json:"total_profit"`
	AvgProfitRate     decimal.Decimal     `json:"avg_profit_rate"`
	ProductCostRatio  decimal.Decimal     `json:"product_cost_ratio"`
	GiftCostRatio     decimal.Decimal     `json:"gift_cost_ratio"`
	ProfitTrend       []TrendData         `json:"profit_trend"`
	CostBreakdown     CostBreakdown       `json:"cost_breakdown"`
}

// CostBreakdown 成本构成
type CostBreakdown struct {
	ProductCost decimal.Decimal `json:"product_cost"`
	GiftCost    decimal.Decimal `json:"gift_cost"`
	OtherCost   decimal.Decimal `json:"other_cost"`
	TotalCost   decimal.Decimal `json:"total_cost"`
}

// PaymentAnalysis 回款分析
type PaymentAnalysis struct {
	CollectionRate      decimal.Decimal       `json:"collection_rate"`
	AvgCollectionDays   float64                `json:"avg_collection_days"`
	TotalPaidAmount     decimal.Decimal        `json:"total_paid_amount"`
	TotalUnpaidAmount   decimal.Decimal        `json:"total_unpaid_amount"`
	PaymentMethodDist   []PaymentMethodDist    `json:"payment_method_dist"`
	CollectionTrend     []TrendData            `json:"collection_trend"`
}

// PaymentMethodDist 回款方式分布
type PaymentMethodDist struct {
	Method string          `json:"method"`
	Count  int64           `json:"count"`
	Amount decimal.Decimal `json:"amount"`
}

// InventoryAnalysis 库存分析
type InventoryAnalysis struct {
	TotalSKUs          int64            `json:"total_skus"`
	TotalStockValue    decimal.Decimal  `json:"total_stock_value"`
	TurnoverRate       decimal.Decimal  `json:"turnover_rate"`
	SlowMovingProducts []SlowMovingItem `json:"slow_moving_products"`
	LowStockStats      LowStockStats    `json:"low_stock_stats"`
}

// SlowMovingItem 滞销商品
type SlowMovingItem struct {
	SKUID        int64           `json:"sku_id"`
	SKUName      string          `json:"sku_name"`
	ProductName  string          `json:"product_name"`
	StockQty     int             `json:"stock_qty"`
	StockValue   decimal.Decimal `json:"stock_value"`
	LastOutDays  int             `json:"last_out_days"`
}

// LowStockStats 库存预警统计
type LowStockStats struct {
	TotalAlerts   int64 `json:"total_alerts"`
	PendingAlerts int64 `json:"pending_alerts"`
	HandledAlerts int64 `json:"handled_alerts"`
}

// CommissionAnalysis 提成分析
type CommissionAnalysis struct {
	TotalCommission     decimal.Decimal          `json:"total_commission"`
	CommissionByType    []CommissionTypeDist     `json:"commission_by_type"`
	IncentiveComparison []IncentiveComparisonItem `json:"incentive_comparison"`
}

// CommissionTypeDist 提成类型分布
type CommissionTypeDist struct {
	TypeName string          `json:"type_name"`
	Type     int8            `json:"type"`
	Count    int64           `json:"count"`
	Amount   decimal.Decimal `json:"amount"`
}

// IncentiveComparisonItem 激励效果对比
type IncentiveComparisonItem struct {
	Period     string          `json:"period"`
	Sales      decimal.Decimal `json:"sales"`
	Profit     decimal.Decimal `json:"profit"`
	Commission decimal.Decimal `json:"commission"`
	Orders     int64           `json:"orders"`
}

// ReportService 报表服务
type ReportService struct {
	db *gorm.DB
}

// NewReportService 创建报表服务实例
func NewReportService(db *gorm.DB) *ReportService {
	return &ReportService{db: db}
}

// ==================== 销售总览 ====================

// GetSalesSummary 获取销售总览
func (s *ReportService) GetSalesSummary(storeID int64, startDate, endDate string) (*SalesSummary, error) {
	summary := &SalesSummary{}

	start, end, err := parseDateRange(startDate, endDate)
	if err != nil {
		return nil, &AppError{Code: 400, Message: "日期格式错误，请使用 YYYY-MM-DD 格式"}
	}

	orderQuery := s.db.Model(&models.Order{}).
		Where("store_id = ? AND order_status = 1 AND order_date BETWEEN ? AND ?", storeID, start, end)

	// 总销售额
	orderQuery.Select("COALESCE(SUM(final_amount), 0)").Scan(&summary.TotalSales)

	// 总订单数
	orderQuery.Session(&gorm.Session{}).Select("COALESCE(COUNT(*), 0)").Scan(&summary.TotalOrders)

	// 总利润
	orderQuery.Session(&gorm.Session{}).Select("COALESCE(SUM(actual_profit), 0)").Scan(&summary.TotalProfit)

	// 总成本
	orderQuery.Session(&gorm.Session{}).Select("COALESCE(SUM(total_cost), 0)").Scan(&summary.TotalCost)

	// 客单价
	if summary.TotalOrders > 0 {
		summary.AvgOrderValue = summary.TotalSales.Div(decimal.NewFromInt(summary.TotalOrders))
	} else {
		summary.AvgOrderValue = decimal.Zero
	}

	// 利润率
	if summary.TotalSales.GreaterThan(decimal.Zero) {
		summary.ProfitRate = summary.TotalProfit.Div(summary.TotalSales).Mul(decimal.NewFromInt(100))
	} else {
		summary.ProfitRate = decimal.Zero
	}

	// 同比计算（与去年同期对比）
	yoyStart := start.AddDate(-1, 0, 0)
	yoyEnd := end.AddDate(-1, 0, 0)
	summary.YoySalesGrowth = calcGrowth(s.db, storeID, "SUM(final_amount)", start, end, yoyStart, yoyEnd)
	summary.YoyOrdersGrowth = calcGrowth(s.db, storeID, "COUNT(*)", start, end, yoyStart, yoyEnd)
	summary.YoyProfitGrowth = calcGrowth(s.db, storeID, "SUM(actual_profit)", start, end, yoyStart, yoyEnd)

	// 环比计算（与上期等长对比）
	duration := end.Sub(start)
	momStart := start.Add(-duration)
	momEnd := start.Add(-time.Second)
	summary.MomSalesGrowth = calcGrowth(s.db, storeID, "SUM(final_amount)", start, end, momStart, momEnd)
	summary.MomOrdersGrowth = calcGrowth(s.db, storeID, "COUNT(*)", start, end, momStart, momEnd)
	summary.MomProfitGrowth = calcGrowth(s.db, storeID, "SUM(actual_profit)", start, end, momStart, momEnd)

	return summary, nil
}

// calcGrowth 计算增长率
func calcGrowth(db *gorm.DB, storeID int64, fieldExpr string, curStart, curEnd, prevStart, prevEnd time.Time) *decimal.Decimal {
	var curVal, prevVal decimal.Decimal

	db.Model(&models.Order{}).
		Where("store_id = ? AND order_status = 1 AND order_date BETWEEN ? AND ?", storeID, curStart, curEnd).
		Select(fmt.Sprintf("COALESCE(%s, 0)", fieldExpr)).Scan(&curVal)

	db.Model(&models.Order{}).
		Where("store_id = ? AND order_status = 1 AND order_date BETWEEN ? AND ?", storeID, prevStart, prevEnd).
		Select(fmt.Sprintf("COALESCE(%s, 0)", fieldExpr)).Scan(&prevVal)

	if prevVal.LessThanOrEqual(decimal.Zero) {
		return nil // 上期无数据，不计算增长率
	}

	growth := curVal.Sub(prevVal).Div(prevVal).Mul(decimal.NewFromInt(100))
	return &growth
}

// ==================== 销售趋势 ====================

// GetSalesTrend 获取销售趋势
func (s *ReportService) GetSalesTrend(storeID int64, startDate, endDate string, dimension string) ([]TrendData, error) {
	start, end, err := parseDateRange(startDate, endDate)
	if err != nil {
		return nil, &AppError{Code: 400, Message: "日期格式错误，请使用 YYYY-MM-DD 格式"}
	}

	// 根据维度确定日期格式和分组
	var dateFormat string
	switch strings.ToLower(dimension) {
	case "week":
		dateFormat = "%Y-%u" // 年-周
	case "month":
		dateFormat = "%Y-%m" // 年-月
	default:
		dateFormat = "%Y-%m-%d" // 日
	}

	type trendRow struct {
		Period      string          `gorm:"column:period"`
		SalesAmount decimal.Decimal `gorm:"column:sales_amount"`
		OrderCount  int64           `gorm:"column:order_count"`
		Profit      decimal.Decimal `gorm:"column:profit"`
	}

	var rows []trendRow
	err = s.db.Model(&models.Order{}).
		Select(
			fmt.Sprintf("DATE_FORMAT(order_date, '%s') as period", dateFormat),
			"COALESCE(SUM(final_amount), 0) as sales_amount",
			"COALESCE(COUNT(*), 0) as order_count",
			"COALESCE(SUM(actual_profit), 0) as profit",
		).
		Where("store_id = ? AND order_status = 1 AND order_date BETWEEN ? AND ?", storeID, start, end).
		Group(fmt.Sprintf("DATE_FORMAT(order_date, '%s')", dateFormat)).
		Order("period ASC").
		Find(&rows).Error

	if err != nil {
		return nil, &AppError{Code: 500, Message: "查询销售趋势失败"}
	}

	result := make([]TrendData, 0, len(rows))
	for _, row := range rows {
		result = append(result, TrendData{
			Date:        row.Period,
			SalesAmount: row.SalesAmount,
			OrderCount:  row.OrderCount,
			Profit:      row.Profit,
		})
	}

	return result, nil
}

// ==================== 业绩排行 ====================

// GetPerformanceRanking 获取业绩排行
func (s *ReportService) GetPerformanceRanking(storeID int64, periodType string, periodValue string, rankBy string) ([]RankingItem, error) {
	start, end, err := parsePeriod(periodType, periodValue)
	if err != nil {
		return nil, &AppError{Code: 400, Message: err.Error()}
	}

	var results []RankingItem

	switch strings.ToLower(rankBy) {
	case "salesman":
		results, err = s.getSalesmanRanking(storeID, start, end)
	case "category":
		results, err = s.getCategoryRanking(storeID, start, end)
	case "product":
		results, err = s.getProductRanking(storeID, start, end)
	default:
		return nil, &AppError{Code: 400, Message: "无效的排行维度，支持: salesman/category/product"}
	}

	if err != nil {
		return nil, err
	}

	return results, nil
}

// getSalesmanRanking 业务员排行
func (s *ReportService) getSalesmanRanking(storeID int64, start, end time.Time) ([]RankingItem, error) {
	type rankRow struct {
		SalesmanID  int64           `gorm:"column:salesman_id"`
		SalesAmount decimal.Decimal `gorm:"column:sales_amount"`
		OrderCount  int64           `gorm:"column:order_count"`
		Profit      decimal.Decimal `gorm:"column:profit"`
	}

	var rows []rankRow
	err := s.db.Model(&models.Order{}).
		Select(
			"salesman_id",
			"COALESCE(SUM(final_amount), 0) as sales_amount",
			"COALESCE(COUNT(*), 0) as order_count",
			"COALESCE(SUM(actual_profit), 0) as profit",
		).
		Where("store_id = ? AND order_status = 1 AND order_date BETWEEN ? AND ?", storeID, start, end).
		Group("salesman_id").
		Order("sales_amount DESC").
		Limit(50).
		Find(&rows).Error
	if err != nil {
		return nil, &AppError{Code: 500, Message: "查询业务员排行失败"}
	}

	// 批量查询业务员姓名
	salesmanIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		salesmanIDs = append(salesmanIDs, row.SalesmanID)
	}

	nameMap := make(map[int64]string)
	if len(salesmanIDs) > 0 {
		var users []models.User
		s.db.Select("id, real_name").Where("id IN ?", salesmanIDs).Find(&users)
		for _, u := range users {
			nameMap[u.ID] = u.RealName
		}
	}

	results := make([]RankingItem, 0, len(rows))
	for _, row := range rows {
		name := nameMap[row.SalesmanID]
		if name == "" {
			name = fmt.Sprintf("ID:%d", row.SalesmanID)
		}
		profitRate := decimal.Zero
		if row.SalesAmount.GreaterThan(decimal.Zero) {
			profitRate = row.Profit.Div(row.SalesAmount).Mul(decimal.NewFromInt(100))
		}
		results = append(results, RankingItem{
			Name:        name,
			ID:          row.SalesmanID,
			SalesAmount: row.SalesAmount,
			OrderCount:  row.OrderCount,
			Profit:      row.Profit,
			ProfitRate:  profitRate,
		})
	}

	return results, nil
}

// getCategoryRanking 品类排行
func (s *ReportService) getCategoryRanking(storeID int64, start, end time.Time) ([]RankingItem, error) {
	type rankRow struct {
		CategoryID  *int64          `gorm:"column:category_id"`
		SalesAmount decimal.Decimal `gorm:"column:sales_amount"`
		OrderCount  int64           `gorm:"column:order_count"`
		Profit      decimal.Decimal `gorm:"column:profit"`
	}

	var rows []rankRow
	err := s.db.Model(&models.OrderItem{}).
		Select(
			"oi.category_id",
			"COALESCE(SUM(oi.sale_price * oi.quantity), 0) as sales_amount",
			"COALESCE(COUNT(DISTINCT oi.order_id), 0) as order_count",
			"COALESCE(SUM(oi.sale_price * oi.quantity) - COALESCE(SUM(oi.total_cost), 0), 0) as profit",
		).
		Joins("JOIN orders o ON o.id = oi.order_id").
		Where("o.store_id = ? AND o.order_status = 1 AND o.order_date BETWEEN ? AND ?", storeID, start, end).
		Group("oi.category_id").
		Order("sales_amount DESC").
		Limit(50).
		Find(&rows).Error
	if err != nil {
		return nil, &AppError{Code: 500, Message: "查询品类排行失败"}
	}

	// 批量查询品类名称
	categoryIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row.CategoryID != nil {
			categoryIDs = append(categoryIDs, *row.CategoryID)
		}
	}

	nameMap := make(map[int64]string)
	if len(categoryIDs) > 0 {
		var categories []models.Category
		s.db.Select("id, category_name").Where("id IN ?", categoryIDs).Find(&categories)
		for _, c := range categories {
			nameMap[c.ID] = c.CategoryName
		}
	}

	results := make([]RankingItem, 0, len(rows))
	for _, row := range rows {
		name := "未分类"
		catID := int64(0)
		if row.CategoryID != nil {
			catID = *row.CategoryID
			if n, ok := nameMap[catID]; ok {
				name = n
			}
		}
		profitRate := decimal.Zero
		if row.SalesAmount.GreaterThan(decimal.Zero) {
			profitRate = row.Profit.Div(row.SalesAmount).Mul(decimal.NewFromInt(100))
		}
		results = append(results, RankingItem{
			Name:        name,
			ID:          catID,
			SalesAmount: row.SalesAmount,
			OrderCount:  row.OrderCount,
			Profit:      row.Profit,
			ProfitRate:  profitRate,
		})
	}

	return results, nil
}

// getProductRanking 商品排行
func (s *ReportService) getProductRanking(storeID int64, start, end time.Time) ([]RankingItem, error) {
	type rankRow struct {
		SKUID       int64           `gorm:"column:sku_id"`
		ProductName string          `gorm:"column:product_name"`
		SKUName     string          `gorm:"column:sku_name"`
		SalesAmount decimal.Decimal `gorm:"column:sales_amount"`
		OrderCount  int64           `gorm:"column:order_count"`
		Profit      decimal.Decimal `gorm:"column:profit"`
	}

	var rows []rankRow
	err := s.db.Model(&models.OrderItem{}).
		Select(
			"oi.sku_id",
			"oi.product_name",
			"oi.sku_name",
			"COALESCE(SUM(oi.sale_price * oi.quantity), 0) as sales_amount",
			"COALESCE(COUNT(DISTINCT oi.order_id), 0) as order_count",
			"COALESCE(SUM(oi.sale_price * oi.quantity) - COALESCE(SUM(oi.total_cost), 0), 0) as profit",
		).
		Joins("JOIN orders o ON o.id = oi.order_id").
		Where("o.store_id = ? AND o.order_status = 1 AND o.order_date BETWEEN ? AND ?", storeID, start, end).
		Group("oi.sku_id").
		Order("sales_amount DESC").
		Limit(50).
		Find(&rows).Error
	if err != nil {
		return nil, &AppError{Code: 500, Message: "查询商品排行失败"}
	}

	results := make([]RankingItem, 0, len(rows))
	for _, row := range rows {
		name := row.ProductName
		if row.SKUName != "" {
			name = fmt.Sprintf("%s - %s", row.ProductName, row.SKUName)
		}
		profitRate := decimal.Zero
		if row.SalesAmount.GreaterThan(decimal.Zero) {
			profitRate = row.Profit.Div(row.SalesAmount).Mul(decimal.NewFromInt(100))
		}
		results = append(results, RankingItem{
			Name:        name,
			ID:          row.SKUID,
			SalesAmount: row.SalesAmount,
			OrderCount:  row.OrderCount,
			Profit:      row.Profit,
			ProfitRate:  profitRate,
		})
	}

	return results, nil
}

// ==================== 利润分析 ====================

// GetProfitAnalysis 获取利润分析
func (s *ReportService) GetProfitAnalysis(storeID int64, startDate, endDate string) (*ProfitAnalysis, error) {
	start, end, err := parseDateRange(startDate, endDate)
	if err != nil {
		return nil, &AppError{Code: 400, Message: "日期格式错误，请使用 YYYY-MM-DD 格式"}
	}

	analysis := &ProfitAnalysis{}

	orderQuery := s.db.Model(&models.Order{}).
		Where("store_id = ? AND order_status = 1 AND order_date BETWEEN ? AND ?", storeID, start, end)

	// 总利润
	orderQuery.Select("COALESCE(SUM(actual_profit), 0)").Scan(&analysis.TotalProfit)

	// 平均利润率
	var totalSales decimal.Decimal
	orderQuery.Session(&gorm.Session{}).Select("COALESCE(SUM(final_amount), 0)").Scan(&totalSales)
	if totalSales.GreaterThan(decimal.Zero) {
		analysis.AvgProfitRate = analysis.TotalProfit.Div(totalSales).Mul(decimal.NewFromInt(100))
	} else {
		analysis.AvgProfitRate = decimal.Zero
	}

	// 成本构成
	var totalCost, totalGiftCost decimal.Decimal
	orderQuery.Session(&gorm.Session{}).Select("COALESCE(SUM(total_cost), 0)").Scan(&totalCost)
	orderQuery.Session(&gorm.Session{}).Select("COALESCE(SUM(gift_cost), 0)").Scan(&totalGiftCost)

	analysis.CostBreakdown = CostBreakdown{
		ProductCost: totalCost,
		GiftCost:    totalGiftCost,
		TotalCost:   totalCost.Add(totalGiftCost),
	}

	// 成本占比
	if analysis.CostBreakdown.TotalCost.GreaterThan(decimal.Zero) {
		analysis.ProductCostRatio = totalCost.Div(analysis.CostBreakdown.TotalCost).Mul(decimal.NewFromInt(100))
		analysis.GiftCostRatio = totalGiftCost.Div(analysis.CostBreakdown.TotalCost).Mul(decimal.NewFromInt(100))
	} else {
		analysis.ProductCostRatio = decimal.Zero
		analysis.GiftCostRatio = decimal.Zero
	}

	// 利润趋势（按月）
	type trendRow struct {
		Period string          `gorm:"column:period"`
		Profit decimal.Decimal `gorm:"column:profit"`
	}
	var trendRows []trendRow
	s.db.Model(&models.Order{}).
		Select(
			"DATE_FORMAT(order_date, '%Y-%m') as period",
			"COALESCE(SUM(actual_profit), 0) as profit",
		).
		Where("store_id = ? AND order_status = 1 AND order_date BETWEEN ? AND ?", storeID, start, end).
		Group("DATE_FORMAT(order_date, '%Y-%m')").
		Order("period ASC").
		Find(&trendRows)

	analysis.ProfitTrend = make([]TrendData, 0, len(trendRows))
	for _, row := range trendRows {
		analysis.ProfitTrend = append(analysis.ProfitTrend, TrendData{
			Date:   row.Period,
			Profit: row.Profit,
		})
	}

	return analysis, nil
}

// ==================== 回款分析 ====================

// GetPaymentAnalysis 获取回款分析
func (s *ReportService) GetPaymentAnalysis(storeID int64, startDate, endDate string) (*PaymentAnalysis, error) {
	start, end, err := parseDateRange(startDate, endDate)
	if err != nil {
		return nil, &AppError{Code: 400, Message: "日期格式错误，请使用 YYYY-MM-DD 格式"}
	}

	analysis := &PaymentAnalysis{}

	// 查询时间范围内的已生效订单
	var totalFinalAmount, totalPaidAmount decimal.Decimal
	s.db.Model(&models.Order{}).
		Where("store_id = ? AND order_status = 1 AND order_date BETWEEN ? AND ?", storeID, start, end).
		Select("COALESCE(SUM(final_amount), 0)").Scan(&totalFinalAmount)

	s.db.Model(&models.Order{}).
		Where("store_id = ? AND order_status = 1 AND order_date BETWEEN ? AND ?", storeID, start, end).
		Select("COALESCE(SUM(paid_amount), 0)").Scan(&totalPaidAmount)

	analysis.TotalPaidAmount = totalPaidAmount
	analysis.TotalUnpaidAmount = totalFinalAmount.Sub(totalPaidAmount)

	// 回款率
	if totalFinalAmount.GreaterThan(decimal.Zero) {
		analysis.CollectionRate = totalPaidAmount.Div(totalFinalAmount).Mul(decimal.NewFromInt(100))
	} else {
		analysis.CollectionRate = decimal.Zero
	}

	// 平均回款周期（天）：从订单日期到回款日期的平均天数
	type collectionDays struct {
		AvgDays *float64 `gorm:"column:avg_days"`
	}
	var cd collectionDays
	s.db.Model(&models.Payment{}).
		Joins("JOIN orders ON orders.id = payments.order_id").
		Where("orders.store_id = ? AND payments.status = 1 AND payments.payment_date IS NOT NULL AND orders.order_date IS NOT NULL", storeID).
		Select("AVG(DATEDIFF(payments.payment_date, orders.order_date)) as avg_days").
		Scan(&cd)
	if cd.AvgDays != nil {
		analysis.AvgCollectionDays = *cd.AvgDays
	}

	// 回款方式分布
	paymentMethodNames := map[int8]string{
		0: "未指定",
		1: "银行转账",
		2: "微信",
		3: "支付宝",
		4: "现金",
		5: "刷卡",
	}

	type methodRow struct {
		Method int8            `gorm:"column:payment_method"`
		Count  int64           `gorm:"column:cnt"`
		Amount decimal.Decimal `gorm:"column:total_amount"`
	}
	var methodRows []methodRow
	s.db.Model(&models.Payment{}).
		Joins("JOIN orders ON orders.id = payments.order_id").
		Where("orders.store_id = ? AND payments.status = 1", storeID).
		Select("payments.payment_method, COUNT(*) as cnt, COALESCE(SUM(payments.amount), 0) as total_amount").
		Group("payments.payment_method").
		Find(&methodRows)

	analysis.PaymentMethodDist = make([]PaymentMethodDist, 0, len(methodRows))
	for _, row := range methodRows {
		name := paymentMethodNames[row.Method]
		if name == "" {
			name = fmt.Sprintf("其他(%d)", row.Method)
		}
		analysis.PaymentMethodDist = append(analysis.PaymentMethodDist, PaymentMethodDist{
			Method: name,
			Count:  row.Count,
			Amount: row.Amount,
		})
	}

	// 回款趋势（按月）
	type trendRow struct {
		Period string          `gorm:"column:period"`
		Amount decimal.Decimal `gorm:"column:amount"`
	}
	var trendRows []trendRow
	s.db.Model(&models.Payment{}).
		Joins("JOIN orders ON orders.id = payments.order_id").
		Where("orders.store_id = ? AND payments.status = 1 AND payments.payment_date BETWEEN ? AND ?", storeID, start, end).
		Select("DATE_FORMAT(payments.payment_date, '%Y-%m') as period, COALESCE(SUM(payments.amount), 0) as amount").
		Group("DATE_FORMAT(payments.payment_date, '%Y-%m')").
		Order("period ASC").
		Find(&trendRows)

	analysis.CollectionTrend = make([]TrendData, 0, len(trendRows))
	for _, row := range trendRows {
		analysis.CollectionTrend = append(analysis.CollectionTrend, TrendData{
			Date:        row.Period,
			SalesAmount: row.Amount,
		})
	}

	return analysis, nil
}

// ==================== 库存分析 ====================

// GetInventoryAnalysis 获取库存分析
func (s *ReportService) GetInventoryAnalysis(storeID int64) (*InventoryAnalysis, error) {
	analysis := &InventoryAnalysis{}

	// 总SKU数
	s.db.Model(&models.WarehouseStock{}).
		Joins("JOIN warehouses w ON w.id = warehouse_stocks.warehouse_id").
		Where("w.store_id = ? AND warehouse_stocks.stock_quantity > 0", storeID).
		Select("COALESCE(COUNT(DISTINCT warehouse_stocks.sku_id), 0)").Scan(&analysis.TotalSKUs)

	// 库存总值（库存数量 * 参考成本价）
	var stockValue decimal.Decimal
	s.db.Model(&models.WarehouseStock{}).
		Joins("JOIN warehouses w ON w.id = warehouse_stocks.warehouse_id").
		Joins("JOIN product_skus ps ON ps.id = warehouse_stocks.sku_id").
		Joins("JOIN products p ON p.id = ps.product_id").
		Where("w.store_id = ? AND warehouse_stocks.stock_quantity > 0", storeID).
		Select("COALESCE(SUM(warehouse_stocks.stock_quantity * p.reference_cost), 0)").Scan(&stockValue)
	analysis.TotalStockValue = stockValue

	// 库存周转率 = 期间销售成本 / 平均库存
	// 使用近90天数据
	now := time.Now()
	ninetyDaysAgo := now.AddDate(0, 0, -90)
	var salesCost decimal.Decimal
	s.db.Model(&models.Order{}).
		Where("store_id = ? AND order_status = 1 AND order_date BETWEEN ? AND ?", storeID, ninetyDaysAgo, now).
		Select("COALESCE(SUM(total_cost), 0)").Scan(&salesCost)
	if stockValue.GreaterThan(decimal.Zero) {
		analysis.TurnoverRate = salesCost.Div(stockValue).Mul(decimal.NewFromInt(100))
	} else {
		analysis.TurnoverRate = decimal.Zero
	}

	// 滞销商品（30天无出库记录）
	thirtyDaysAgo := now.AddDate(0, 0, -30)
	type slowItem struct {
		SKUID       int64           `gorm:"column:sku_id"`
		SKUName     string          `gorm:"column:sku_name"`
		ProductName string          `gorm:"column:product_name"`
		StockQty    int             `gorm:"column:stock_qty"`
		StockValue  decimal.Decimal `gorm:"column:stock_value"`
		LastOutDate *time.Time      `gorm:"column:last_out_date"`
	}
	var slowItems []slowItem
	s.db.Raw(`
		SELECT
			ws.sku_id,
			COALESCE(ps.sku_name, '') as sku_name,
			COALESCE(p.product_name, '') as product_name,
			ws.stock_quantity as stock_qty,
			ws.stock_quantity * COALESCE(p.reference_cost, 0) as stock_value,
			(
				SELECT MAX(it.created_at)
				FROM inventory_transactions it
				WHERE it.biz_id = ws.sku_id AND it.transaction_type = 2 AND it.biz_type = 1
			) as last_out_date
		FROM warehouse_stocks ws
		JOIN warehouses w ON w.id = ws.warehouse_id
		LEFT JOIN product_skus ps ON ps.id = ws.sku_id
		LEFT JOIN products p ON p.id = ps.product_id
		WHERE w.store_id = ? AND ws.stock_quantity > 0
		HAVING last_out_date IS NULL OR last_out_date < ?
		ORDER BY stock_value DESC
		LIMIT 50
	`, storeID, thirtyDaysAgo).Scan(&slowItems)

	analysis.SlowMovingProducts = make([]SlowMovingItem, 0, len(slowItems))
	for _, item := range slowItems {
		lastOutDays := 999
		if item.LastOutDate != nil {
			lastOutDays = int(now.Sub(*item.LastOutDate).Hours() / 24)
		}
		analysis.SlowMovingProducts = append(analysis.SlowMovingProducts, SlowMovingItem{
			SKUID:       item.SKUID,
			SKUName:     item.SKUName,
			ProductName: item.ProductName,
			StockQty:    item.StockQty,
			StockValue:  item.StockValue,
			LastOutDays: lastOutDays,
		})
	}

	// 库存预警统计
	s.db.Model(&models.StockAlert{}).
		Where("store_id = ?", storeID).
		Select("COALESCE(COUNT(*), 0)").Scan(&analysis.LowStockStats.TotalAlerts)

	s.db.Model(&models.StockAlert{}).
		Where("store_id = ? AND alert_status = 0", storeID).
		Select("COALESCE(COUNT(*), 0)").Scan(&analysis.LowStockStats.PendingAlerts)

	s.db.Model(&models.StockAlert{}).
		Where("store_id = ? AND alert_status = 1", storeID).
		Select("COALESCE(COUNT(*), 0)").Scan(&analysis.LowStockStats.HandledAlerts)

	return analysis, nil
}

// ==================== 提成分析 ====================

// GetCommissionAnalysis 获取提成分析
func (s *ReportService) GetCommissionAnalysis(storeID int64, startDate, endDate string) (*CommissionAnalysis, error) {
	start, end, err := parseDateRange(startDate, endDate)
	if err != nil {
		return nil, &AppError{Code: 400, Message: "日期格式错误，请使用 YYYY-MM-DD 格式"}
	}

	analysis := &CommissionAnalysis{}

	// 提成总额
	s.db.Model(&models.Commission{}).
		Where("store_id = ? AND created_at BETWEEN ? AND ?", storeID, start, end).
		Select("COALESCE(SUM(amount), 0)").Scan(&analysis.TotalCommission)

	// 提成类型分布
	commissionTypeNames := map[int8]string{
		1: "业务员提成",
		2: "同行分成",
		3: "主管团队分润",
		4: "店长团队分润",
		5: "基金池奖励",
		6: "老带新奖励",
	}

	type typeRow struct {
		CommissionType int8            `gorm:"column:commission_type"`
		Count          int64           `gorm:"column:cnt"`
		Amount         decimal.Decimal `gorm:"column:total_amount"`
	}
	var typeRows []typeRow
	s.db.Model(&models.Commission{}).
		Where("store_id = ? AND created_at BETWEEN ? AND ?", storeID, start, end).
		Select("commission_type, COUNT(*) as cnt, COALESCE(SUM(amount), 0) as total_amount").
		Group("commission_type").
		Find(&typeRows)

	analysis.CommissionByType = make([]CommissionTypeDist, 0, len(typeRows))
	for _, row := range typeRows {
		name := commissionTypeNames[row.CommissionType]
		if name == "" {
			name = fmt.Sprintf("其他(%d)", row.CommissionType)
		}
		analysis.CommissionByType = append(analysis.CommissionByType, CommissionTypeDist{
			TypeName: name,
			Type:     row.CommissionType,
			Count:    row.Count,
			Amount:   row.Amount,
		})
	}

	// 激励效果对比（按月：销售额、利润、提成、订单数）
	type incentiveRow struct {
		Period     string          `gorm:"column:period"`
		Sales      decimal.Decimal `gorm:"column:sales"`
		Profit     decimal.Decimal `gorm:"column:profit"`
		Commission decimal.Decimal `gorm:"column:commission"`
		Orders     int64           `gorm:"column:orders"`
	}
	var incentiveRows []incentiveRow
	s.db.Raw(`
		SELECT
			DATE_FORMAT(o.order_date, '%Y-%m') as period,
			COALESCE(SUM(o.final_amount), 0) as sales,
			COALESCE(SUM(o.actual_profit), 0) as profit,
			COALESCE(SUM(c.amount), 0) as commission,
			COUNT(DISTINCT o.id) as orders
		FROM orders o
		LEFT JOIN commissions c ON c.order_id = o.id
		WHERE o.store_id = ? AND o.order_status = 1 AND o.order_date BETWEEN ? AND ?
		GROUP BY DATE_FORMAT(o.order_date, '%Y-%m')
		ORDER BY period ASC
	`, storeID, start, end).Scan(&incentiveRows)

	analysis.IncentiveComparison = make([]IncentiveComparisonItem, 0, len(incentiveRows))
	for _, row := range incentiveRows {
		analysis.IncentiveComparison = append(analysis.IncentiveComparison, IncentiveComparisonItem{
			Period:     row.Period,
			Sales:      row.Sales,
			Profit:     row.Profit,
			Commission: row.Commission,
			Orders:     row.Orders,
		})
	}

	return analysis, nil
}

// ==================== 辅助函数 ====================

// parseDateRange 解析日期范围
func parseDateRange(startDate, endDate string) (time.Time, time.Time, error) {
	if startDate == "" || endDate == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("开始日期和结束日期不能为空")
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("开始日期格式错误")
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("结束日期格式错误")
	}

	// 设置为当天的起止时间
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, end.Location())

	return start, end, nil
}

// parsePeriod 解析周期参数
func parsePeriod(periodType, periodValue string) (time.Time, time.Time, error) {
	now := time.Now()
	loc := now.Location()

	switch strings.ToLower(periodType) {
	case "year":
		year, err := parseYear(periodValue, now)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		start := time.Date(year, 1, 1, 0, 0, 0, 0, loc)
		end := time.Date(year, 12, 31, 23, 59, 59, 0, loc)
		return start, end, nil

	case "month":
		year, month, err := parseYearMonth(periodValue, now)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		start := time.Date(year, month, 1, 0, 0, 0, 0, loc)
		end := start.AddDate(0, 1, 0).Add(-time.Second)
		return start, end, nil

	case "quarter":
		year, quarter, err := parseYearQuarter(periodValue, now)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		startMonth := (quarter-1)*3 + 1
		start := time.Date(year, time.Month(startMonth), 1, 0, 0, 0, 0, loc)
		end := start.AddDate(0, 3, 0).Add(-time.Second)
		return start, end, nil

	default:
		// 默认当月
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		end := start.AddDate(0, 1, 0).Add(-time.Second)
		return start, end, nil
	}
}

func parseYear(value string, now time.Time) (int, error) {
	if value == "" {
		return now.Year(), nil
	}
	year := 0
	for _, ch := range value {
		if ch >= '0' && ch <= '9' {
			year = year*10 + int(ch-'0')
		}
	}
	if year < 2000 || year > 2100 {
		return now.Year(), nil
	}
	return year, nil
}

func parseYearMonth(value string, now time.Time) (int, time.Month, error) {
	if value == "" {
		return now.Year(), now.Month(), nil
	}
	parts := strings.Split(value, "-")
	if len(parts) == 2 {
		year := 0
		for _, ch := range parts[0] {
			if ch >= '0' && ch <= '9' {
				year = year*10 + int(ch-'0')
			}
		}
		month := 0
		for _, ch := range parts[1] {
			if ch >= '0' && ch <= '9' {
				month = month*10 + int(ch-'0')
			}
		}
		if year >= 2000 && year <= 2100 && month >= 1 && month <= 12 {
			return year, time.Month(month), nil
		}
	}
	return now.Year(), now.Month(), nil
}

func parseYearQuarter(value string, now time.Time) (int, int, error) {
	if value == "" {
		quarter := (int(now.Month()) - 1) / 3
		if quarter < 1 {
			quarter = 1
		}
		return now.Year(), quarter, nil
	}
	parts := strings.Split(value, "-")
	if len(parts) == 2 {
		year := 0
		for _, ch := range parts[0] {
			if ch >= '0' && ch <= '9' {
				year = year*10 + int(ch-'0')
			}
		}
		quarter := 0
		for _, ch := range parts[1] {
			if ch >= '0' && ch <= '9' {
				quarter = quarter*10 + int(ch-'0')
			}
		}
		if year >= 2000 && year <= 2100 && quarter >= 1 && quarter <= 4 {
			return year, quarter, nil
		}
	}
	quarter := (int(now.Month()) - 1) / 3
	if quarter < 1 {
		quarter = 1
	}
	return now.Year(), quarter, nil
}
