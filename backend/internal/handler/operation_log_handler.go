package handler

import (
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// OperationLogHandler 操作日志处理器
type OperationLogHandler struct {
	logService *service.OperationLogService
}

// NewOperationLogHandler 创建操作日志处理器实例
func NewOperationLogHandler(logService *service.OperationLogService) *OperationLogHandler {
	return &OperationLogHandler{logService: logService}
}

// List 获取操作日志列表
// @Summary      获取操作日志列表
// @Description  分页查询操作日志列表，支持关键字搜索
// @Tags         操作日志
// @Accept       json
// @Produce      json
// @Param        keyword    query  string  false  "搜索关键字"
// @Param        page       query  int     false  "页码"      default(1)
// @Param        pageSize   query  int     false  "每页数量"  default(20)
// @Success      200  {object}  handler.Response{data=handler.PageData{list=[]interface{}}}  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /operation-logs [get]
func (h *OperationLogHandler) List(c *gin.Context) {
	var req service.ListOperationLogRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	// 兼容前端传的 pageSize 参数
	if req.PageSize == 0 {
		if ps, err := strconv.Atoi(c.DefaultQuery("pageSize", "20")); err == nil {
			req.PageSize = ps
		}
	}

	result, err := h.logService.List(&req)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询操作日志失败")
		return
	}

	Success(c, result)
}
