package handler

import (
	"furniture-commission/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// MaintenanceHandler 平台维护处理器
type MaintenanceHandler struct {
	maintenanceService *service.MaintenanceService
}

// NewMaintenanceHandler 创建平台维护处理器实例
func NewMaintenanceHandler(maintenanceService *service.MaintenanceService) *MaintenanceHandler {
	return &MaintenanceHandler{
		maintenanceService: maintenanceService,
	}
}

// GetDataTables 获取可清除的数据表列表
// GET /api/maintenance/data-tables
func (h *MaintenanceHandler) GetDataTables(c *gin.Context) {
	tables := h.maintenanceService.GetDataTables()
	Success(c, tables)
}

// CheckRecentBackup 检查是否有10分钟内备份
// POST /api/maintenance/check-recent-backup
func (h *MaintenanceHandler) CheckRecentBackup(c *gin.Context) {
	hasBackup, backup, err := h.maintenanceService.CheckRecentBackup()
	if err != nil {
		Error(c, http.StatusInternalServerError, "检查备份状态失败")
		return
	}

	Success(c, gin.H{
		"has_backup": hasBackup,
		"backup":     backup,
	})
}

// CreateBackup 创建数据库备份
// POST /api/backups
func (h *MaintenanceHandler) CreateBackup(c *gin.Context) {
	var req service.CreateBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	userID := c.GetInt64("user_id")
	backup, err := h.maintenanceService.CreateBackup(&req, userID)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, http.StatusInternalServerError, "创建备份失败")
		return
	}

	Success(c, backup)
}

// ListBackups 获取备份列表
// GET /api/backups
func (h *MaintenanceHandler) ListBackups(c *gin.Context) {
	var req service.ListBackupRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.maintenanceService.ListBackups(&req)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, http.StatusInternalServerError, "查询备份列表失败")
		return
	}

	Success(c, result)
}

// DeleteBackup 删除备份
// DELETE /api/backups/:id
func (h *MaintenanceHandler) DeleteBackup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的备份ID")
		return
	}

	if err := h.maintenanceService.DeleteBackup(id); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, http.StatusInternalServerError, "删除备份失败")
		return
	}

	Success(c, nil)
}

// RestoreBackup 还原备份
// POST /api/backups/:id/restore
func (h *MaintenanceHandler) RestoreBackup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的备份ID")
		return
	}

	if err := h.maintenanceService.RestoreBackup(id); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, http.StatusInternalServerError, "还原备份失败")
		return
	}

	Success(c, nil)
}

// ClearData 清除业务数据
// POST /api/maintenance/clear-data
func (h *MaintenanceHandler) ClearData(c *gin.Context) {
	var req service.ClearDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	if err := h.maintenanceService.ClearData(&req); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, http.StatusInternalServerError, "清除数据失败")
		return
	}

	Success(c, nil)
}
