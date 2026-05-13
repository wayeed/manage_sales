package handler

import (
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// SettleFundPoolRequest 基金池结算请求
type SettleFundPoolRequest struct {
	StoreID     int64  `json:"store_id" binding:"required" example:"1"`
	PeriodType  int    `json:"period_type" binding:"required" example:"1"`
	PeriodValue string `json:"period_value" binding:"required" example:"2024-01"`
}

// FundPoolHandler 基金池处理器
type FundPoolHandler struct {
	fundPoolService *service.FundPoolService
}

// NewFundPoolHandler 创建基金池处理器实例
func NewFundPoolHandler(fundPoolService *service.FundPoolService) *FundPoolHandler {
	return &FundPoolHandler{fundPoolService: fundPoolService}
}

// List 获取基金池列表
// @Summary      获取基金池列表
// @Description  分页查询基金池列表，支持按门店和周期类型筛选
// @Tags         基金池
// @Accept       json
// @Produce      json
// @Param        page         query  int   false  "页码"      default(1)
// @Param        page_size    query  int   false  "每页数量"  default(10)
// @Param        store_id     query  string  false  "门店ID"
// @Param        period_type  query  int   false  "周期类型"
// @Success      200  {object}  handler.Response{data=handler.PageData{list=[]interface{}}}  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /fund-pools [get]
func (h *FundPoolHandler) List(c *gin.Context) {
	var req service.ListFundPoolRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	result, err := h.fundPoolService.List(&req)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询基金池列表失败")
		return
	}

	PageResponse(c, result.List, result.Total, result.Page, result.PageSize)
}

// GetShares 获取基金池份额详情
// @Summary      获取基金池份额详情
// @Description  根据基金池ID获取份额分配详情
// @Tags         基金池
// @Accept       json
// @Produce      json
// @Param        id   path  int64  true  "基金池ID"
// @Success      200  {object}  handler.Response{data=interface{}}  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /fund-pools/{id}/shares [get]
func (h *FundPoolHandler) GetShares(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的基金池ID")
		return
	}

	fundPool, err := h.fundPoolService.GetShares(id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "获取基金池份额详情失败")
		return
	}

	Success(c, fundPool)
}

// SettleFundPool 基金池结算
// @Summary      基金池结算
// @Description  对指定门店和周期进行基金池结算（管理员）
// @Tags         基金池
// @Accept       json
// @Produce      json
// @Param        request  body  handler.SettleFundPoolRequest  true  "结算请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /fund-pools/settle [post]
func (h *FundPoolHandler) SettleFundPool(c *gin.Context) {
	var req SettleFundPoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误: "+err.Error())
		return
	}

	if err := h.fundPoolService.SettleFundPool(req.StoreID, req.PeriodType, req.PeriodValue); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "基金池结算失败")
		return
	}

	Success(c, nil)
}
