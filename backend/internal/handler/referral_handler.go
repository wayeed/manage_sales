package handler

import (
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// ReferralHandler 引荐关系处理器
type ReferralHandler struct {
	referralService *service.ReferralService
}

// NewReferralHandler 创建引荐关系处理器实例
func NewReferralHandler(referralService *service.ReferralService) *ReferralHandler {
	return &ReferralHandler{
		referralService: referralService,
	}
}

// List 获取引荐关系列表
// GET /api/referrals?page=1&page_size=10
func (h *ReferralHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	relations, total, err := h.referralService.List(page, pageSize)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询引荐关系列表失败")
		return
	}

	PageResponse(c, relations, total, page, pageSize)
}

// Create 创建引荐关系
// POST /api/referrals
func (h *ReferralHandler) Create(c *gin.Context) {
	var req service.CreateReferralRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	if err := h.referralService.Create(&req); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "创建引荐关系失败")
		return
	}

	Success(c, nil)
}

// Terminate 终止引荐关系
// POST /api/referrals/:id/terminate
func (h *ReferralHandler) Terminate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的引荐关系ID")
		return
	}

	if err := h.referralService.Terminate(id); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "终止引荐关系失败")
		return
	}

	Success(c, nil)
}
