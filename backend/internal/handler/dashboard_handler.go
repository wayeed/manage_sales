package handler

import (
	"strconv"

	"furniture-commission/internal/models"
	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DashboardHandler 仪表盘处理器
type DashboardHandler struct {
	dashboardService *service.DashboardService
	db               *gorm.DB
}

// NewDashboardHandler 创建仪表盘处理器实例
func NewDashboardHandler(dashboardService *service.DashboardService, db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService, db: db}
}

// getUserStoreID 获取用户门店ID：优先从查询参数获取，其次从当前登录用户获取
func getUserStoreID(c *gin.Context, db *gorm.DB) (int64, error) {
	// 优先从查询参数获取（管理员可跨门店查看）
	if storeIDStr := c.Query("store_id"); storeIDStr != "" {
		storeID, err := strconv.ParseInt(storeIDStr, 10, 64)
		if err == nil && storeID > 0 {
			return storeID, nil
		}
	}

	// 从当前登录用户获取门店ID
	userID := GetUserID(c)
	if userID > 0 {
		var user models.User
		if err := db.Select("store_id").First(&user, userID).Error; err == nil && user.StoreID != nil && *user.StoreID > 0 {
			return *user.StoreID, nil
		}
	}

	return 0, nil
}

// GetOverview 获取仪表盘概览
// @Summary      获取仪表盘概览
// @Description  获取当前门店的仪表盘概览数据，包括销售、订单、客户等核心指标
// @Tags         仪表盘
// @Accept       json
// @Produce      json
// @Param        store_id  query  int64  false  "门店ID（不传则使用当前用户门店）"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /dashboard/overview [get]
func (h *DashboardHandler) GetOverview(c *gin.Context) {
	storeID, err := getUserStoreID(c, h.db)
	if err != nil || storeID <= 0 {
		Error(c, 400, "无效的门店ID，请先绑定门店")
		return
	}

	// 获取当前用户ID和角色
	userID := GetUserID(c)
	roleCodes := GetRoleCodes(c)

	overview, err := h.dashboardService.GetOverview(storeID, userID, roleCodes)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "获取仪表盘概览失败")
		return
	}

	Success(c, overview)
}
