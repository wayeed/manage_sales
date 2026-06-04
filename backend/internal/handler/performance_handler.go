package handler

import (
	"strconv"
	"time"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PerformanceHandler 业绩查询 Handler
type PerformanceHandler struct {
	db            *gorm.DB
	commissionSvc *service.CommissionService
}

// NewPerformanceHandler 创建业绩查询 Handler
func NewPerformanceHandler(db *gorm.DB, commissionSvc *service.CommissionService) *PerformanceHandler {
	return &PerformanceHandler{
		db:            db,
		commissionSvc: commissionSvc,
	}
}

// PerformanceOverview 业绩概览响应
type PerformanceOverview struct {
	TotalSales      string `json:"totalSales" example:"50000.00"`
	TotalProfit     string `json:"totalProfit" example:"8000.00"`
	TotalCommission string `json:"totalCommission" example:"3500.00"`
	CommissionRank  int    `json:"commissionRank" example:"3"`
	SalesCommission string `json:"salesCommission" example:"2500.00"`
	TeamShare       string `json:"teamShare" example:"500.00"`
	FundReward      string `json:"fundReward" example:"300.00"`
	MentorReward    string `json:"mentorReward" example:"200.00"`
}

// GetOverview 获取业绩概览
// @Summary      获取业绩概览
// @Description  获取当前用户指定月份的业绩概览数据
// @Tags         业绩查询
// @Accept       json
// @Produce      json
// @Param        year   query  int  true  "年份"
// @Param        month  query  int  true  "月份"
// @Success      200  {object}  handler.Response{data=PerformanceOverview}  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /performance/overview [get]
func (h *PerformanceHandler) GetOverview(c *gin.Context) {
	// 从上下文获取用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		Error(c, 401, "未登录")
		return
	}
	userID, ok := userIDVal.(int64)
	if !ok {
		Error(c, 401, "用户ID格式错误")
		return
	}

	yearStr := c.Query("year")
	monthStr := c.Query("month")

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		year = time.Now().Year()
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		month = int(time.Now().Month())
	}

	// 构建日期范围
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")
	periodValue := startDate.Format("2006-01")

	// 获取提成汇总（使用 period_value 格式 YYYY-MM）
	summary, err := h.commissionSvc.GetSummary(userID, periodValue, periodValue)
	if err != nil {
		summary = &service.CommissionSummary{}
	}

	// 获取员工月度汇总排名
	monthlySummaries, err := h.commissionSvc.GetMonthlySummary(periodValue)
	if err != nil {
		monthlySummaries = []service.EmployeeCommissionSummary{}
	}

	// 计算排名
	rank := 0
	for i, s := range monthlySummaries {
		if s.EmployeeID == userID {
			rank = i + 1
			break
		}
	}

	// 获取销售额和利润
	type salesStats struct {
		TotalSales  float64
		TotalProfit float64
	}
	var stats salesStats
	h.db.Raw(`
		SELECT COALESCE(SUM(final_amount), 0) as total_sales, COALESCE(SUM(actual_profit), 0) as total_profit
		FROM orders
		WHERE salesman_id = ? AND order_status = 1
		AND order_date >= ? AND order_date <= ?
	`, userID, startDateStr, endDateStr).Scan(&stats)

	overview := PerformanceOverview{
		TotalSales:      strconv.FormatFloat(stats.TotalSales, 'f', 2, 64),
		TotalProfit:     strconv.FormatFloat(stats.TotalProfit, 'f', 2, 64),
		TotalCommission: summary.Total.String(),
		CommissionRank:  rank,
		SalesCommission: summary.SalesCommission.String(),
		TeamShare:       summary.TeamShare.String(),
		FundReward:      summary.FundPool.String(),
		MentorReward:    summary.ReferralReward.String(),
	}

	Success(c, overview)
}
