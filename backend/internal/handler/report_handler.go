package handler

import (
	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ReportHandler 报表处理器
type ReportHandler struct {
	reportService *service.ReportService
	db            *gorm.DB
}

// NewReportHandler 创建报表处理器实例
func NewReportHandler(reportService *service.ReportService, db *gorm.DB) *ReportHandler {
	return &ReportHandler{reportService: reportService, db: db}
}

// GetSalesSummary 获取销售总览
// @Summary      获取销售总览
// @Description  获取指定时间范围内的销售汇总数据
// @Tags         数据报表
// @Accept       json
// @Produce      json
// @Param        store_id    query  int64   false  "门店ID（不传则使用当前用户门店）"
// @Param        start_date  query  string  true   "开始日期，格式：2025-01-01"
// @Param        end_date    query  string  true   "结束日期，格式：2025-12-31"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /reports/sales/summary [get]
func (h *ReportHandler) GetSalesSummary(c *gin.Context) {
	storeID, err := getUserStoreID(c, h.db)
	if err != nil || storeID <= 0 {
		Error(c, 400, "无效的门店ID，请先绑定门店")
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate == "" || endDate == "" {
		Error(c, 400, "开始日期和结束日期不能为空")
		return
	}

	summary, err := h.reportService.GetSalesSummary(storeID, startDate, endDate)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "获取销售总览失败")
		return
	}

	Success(c, summary)
}

// GetSalesTrend 获取销售趋势
// @Summary      获取销售趋势
// @Description  获取指定时间范围内的销售趋势数据，支持按日/周/月维度聚合
// @Tags         数据报表
// @Accept       json
// @Produce      json
// @Param        store_id    query  int64   false  "门店ID（不传则使用当前用户门店）"
// @Param        start_date  query  string  true   "开始日期，格式：2025-01-01"
// @Param        end_date    query  string  true   "结束日期，格式：2025-12-31"
// @Param        dimension   query  string  false  "聚合维度：day/week/month"  default(day)
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /reports/sales/trend [get]
func (h *ReportHandler) GetSalesTrend(c *gin.Context) {
	storeID, err := getUserStoreID(c, h.db)
	if err != nil || storeID <= 0 {
		Error(c, 400, "无效的门店ID，请先绑定门店")
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate == "" || endDate == "" {
		Error(c, 400, "开始日期和结束日期不能为空")
		return
	}

	dimension := c.DefaultQuery("dimension", "day")

	trend, err := h.reportService.GetSalesTrend(storeID, startDate, endDate, dimension)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "获取销售趋势失败")
		return
	}

	Success(c, trend)
}

// GetPerformanceRanking 获取业绩排行
// @Summary      获取业绩排行
// @Description  获取指定周期内的业绩排行榜，支持按销售员/设计师维度排行
// @Tags         数据报表
// @Accept       json
// @Produce      json
// @Param        store_id      query  int64   false  "门店ID（不传则使用当前用户门店）"
// @Param        period_type   query  string  false  "周期类型：month/quarter/year"  default(month)
// @Param        period_value  query  string  false  "周期值，如：2025-05"
// @Param        rank_by       query  string  false  "排行维度：salesman/designer"  default(salesman)
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /reports/sales/ranking [get]
func (h *ReportHandler) GetPerformanceRanking(c *gin.Context) {
	storeID, err := getUserStoreID(c, h.db)
	if err != nil || storeID <= 0 {
		Error(c, 400, "无效的门店ID，请先绑定门店")
		return
	}

	periodType := c.DefaultQuery("period_type", "month")
	periodValue := c.DefaultQuery("period_value", "")
	rankBy := c.DefaultQuery("rank_by", "salesman")

	ranking, err := h.reportService.GetPerformanceRanking(storeID, periodType, periodValue, rankBy)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "获取业绩排行失败")
		return
	}

	Success(c, ranking)
}

// GetProfitAnalysis 获取利润分析
// @Summary      获取利润分析
// @Description  获取指定时间范围内的利润分析数据
// @Tags         数据报表
// @Accept       json
// @Produce      json
// @Param        store_id    query  int64   false  "门店ID（不传则使用当前用户门店）"
// @Param        start_date  query  string  true   "开始日期，格式：2025-01-01"
// @Param        end_date    query  string  true   "结束日期，格式：2025-12-31"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /reports/profit/analysis [get]
func (h *ReportHandler) GetProfitAnalysis(c *gin.Context) {
	storeID, err := getUserStoreID(c, h.db)
	if err != nil || storeID <= 0 {
		Error(c, 400, "无效的门店ID，请先绑定门店")
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate == "" || endDate == "" {
		Error(c, 400, "开始日期和结束日期不能为空")
		return
	}

	analysis, err := h.reportService.GetProfitAnalysis(storeID, startDate, endDate)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "获取利润分析失败")
		return
	}

	Success(c, analysis)
}

// GetPaymentAnalysis 获取回款分析
// @Summary      获取回款分析
// @Description  获取指定时间范围内的回款分析数据
// @Tags         数据报表
// @Accept       json
// @Produce      json
// @Param        store_id    query  int64   false  "门店ID（不传则使用当前用户门店）"
// @Param        start_date  query  string  true   "开始日期，格式：2025-01-01"
// @Param        end_date    query  string  true   "结束日期，格式：2025-12-31"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /reports/payment/analysis [get]
func (h *ReportHandler) GetPaymentAnalysis(c *gin.Context) {
	storeID, err := getUserStoreID(c, h.db)
	if err != nil || storeID <= 0 {
		Error(c, 400, "无效的门店ID，请先绑定门店")
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate == "" || endDate == "" {
		Error(c, 400, "开始日期和结束日期不能为空")
		return
	}

	analysis, err := h.reportService.GetPaymentAnalysis(storeID, startDate, endDate)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "获取回款分析失败")
		return
	}

	Success(c, analysis)
}

// GetInventoryAnalysis 获取库存分析
// @Summary      获取库存分析
// @Description  获取当前门店的库存分析数据
// @Tags         数据报表
// @Accept       json
// @Produce      json
// @Param        store_id  query  int64  false  "门店ID（不传则使用当前用户门店）"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /reports/inventory/analysis [get]
func (h *ReportHandler) GetInventoryAnalysis(c *gin.Context) {
	storeID, err := getUserStoreID(c, h.db)
	if err != nil || storeID <= 0 {
		Error(c, 400, "无效的门店ID，请先绑定门店")
		return
	}

	analysis, err := h.reportService.GetInventoryAnalysis(storeID)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "获取库存分析失败")
		return
	}

	Success(c, analysis)
}

// GetCommissionAnalysis 获取提成分析
// @Summary      获取提成分析
// @Description  获取指定时间范围内的提成分析数据
// @Tags         数据报表
// @Accept       json
// @Produce      json
// @Param        store_id    query  int64   false  "门店ID（不传则使用当前用户门店）"
// @Param        start_date  query  string  true   "开始日期，格式：2025-01-01"
// @Param        end_date    query  string  true   "结束日期，格式：2025-12-31"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /reports/commission/analysis [get]
func (h *ReportHandler) GetCommissionAnalysis(c *gin.Context) {
	storeID, err := getUserStoreID(c, h.db)
	if err != nil || storeID <= 0 {
		Error(c, 400, "无效的门店ID，请先绑定门店")
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate == "" || endDate == "" {
		Error(c, 400, "开始日期和结束日期不能为空")
		return
	}

	analysis, err := h.reportService.GetCommissionAnalysis(storeID, startDate, endDate)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "获取提成分析失败")
		return
	}

	Success(c, analysis)
}
