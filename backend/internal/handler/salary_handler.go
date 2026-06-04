package handler

import (
	"fmt"
	"strconv"

	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// GenerateSalaryRequest 生成月度工资请求
type GenerateSalaryRequest struct {
	StoreID     int64  `json:"store_id" binding:"required" example:"1"`
	SalaryMonth string `json:"salary_month" binding:"required" example:"2024-01"`
}

// SalaryHandler 工资管理处理器
type SalaryHandler struct {
	salaryService *service.SalaryService
}

// NewSalaryHandler 创建工资管理处理器实例
func NewSalaryHandler(salaryService *service.SalaryService) *SalaryHandler {
	return &SalaryHandler{salaryService: salaryService}
}

// List 获取工资列表
// @Summary      获取工资列表
// @Description  分页查询工资记录列表，支持按门店、月份、状态筛选
// @Tags         工资管理
// @Accept       json
// @Produce      json
// @Param        page          query  int     false  "页码"          default(1)
// @Param        page_size     query  int     false  "每页数量"      default(10)
// @Param        store_id      query  string  false  "门店ID"
// @Param        salary_month  query  string  false  "工资月份"      default(2024-01)
// @Param        status        query  string  false  "状态"
// @Success      200  {object}  handler.Response{data=handler.PageData{list=[]interface{}}}  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /salaries [get]
func (h *SalaryHandler) List(c *gin.Context) {
	var req service.ListSalaryRequest
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

	result, err := h.salaryService.List(&req)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "查询工资列表失败")
		return
	}

	PageResponse(c, result.List, result.Total, result.Page, result.PageSize)
}

// GetDetail 获取工资详情
// @Summary      获取工资详情
// @Description  根据工资记录ID获取工资详细信息
// @Tags         工资管理
// @Accept       json
// @Produce      json
// @Param        id   path  int64  true  "工资记录ID"
// @Success      200  {object}  handler.Response{data=interface{}}  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /salaries/{id} [get]
func (h *SalaryHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的工资记录ID")
		return
	}

	detail, err := h.salaryService.GetDetail(id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "获取工资详情失败")
		return
	}

	Success(c, detail)
}

// GetEmployeeSalary 获取员工月度工资
// @Summary      获取员工月度工资
// @Description  根据员工ID和工资月份获取该员工的月度工资记录
// @Tags         工资管理
// @Accept       json
// @Produce      json
// @Param        employee_id   path  int64  true  "员工ID"
// @Param        salary_month  path  string  true  "工资月份，格式：2024-01"
// @Success      200  {object}  handler.Response{data=interface{}}  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /salaries/employee/{employee_id}/{salary_month} [get]
func (h *SalaryHandler) GetEmployeeSalary(c *gin.Context) {
	employeeID, err := strconv.ParseInt(c.Param("employee_id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的员工ID")
		return
	}

	salaryMonth := c.Param("salary_month")
	if salaryMonth == "" {
		Error(c, 400, "工资月份不能为空")
		return
	}

	record, err := h.salaryService.GetEmployeeSalary(employeeID, salaryMonth)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "获取员工月度工资失败")
		return
	}

	Success(c, record)
}

// GenerateSalary 生成月度工资
// @Summary      生成月度工资
// @Description  根据门店和月份生成月度工资记录（管理员）
// @Tags         工资管理
// @Accept       json
// @Produce      json
// @Param        request  body  handler.GenerateSalaryRequest  true  "生成工资请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /salaries/generate [post]
func (h *SalaryHandler) GenerateSalary(c *gin.Context) {
	var req GenerateSalaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误: "+err.Error())
		return
	}

	if err := h.salaryService.GenerateSalary(req.StoreID, req.SalaryMonth); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "生成月度工资失败")
		return
	}

	Success(c, nil)
}

// ExportSalarySlip 导出工资条
// @Summary      导出工资条
// @Description  导出指定工资记录的工资条文件
// @Tags         工资管理
// @Produce      application/octet-stream
// @Param        id   path  int64  true  "工资记录ID"
// @Success      200  "工资条文件"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /salaries/{id}/export [get]
func (h *SalaryHandler) ExportSalarySlip(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的工资记录ID")
		return
	}

	data, filename, err := h.salaryService.ExportSalarySlip(id)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "导出工资条失败")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", filename))
	c.Data(200, "application/octet-stream", data)
}

// ConfirmSalary 审核确认工资
// @Summary      审核确认工资
// @Description  审核确认指定工资记录（管理员）
// @Tags         工资管理
// @Accept       json
// @Produce      json
// @Param        id   path  int64  true  "工资记录ID"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /salaries/{id}/confirm [post]
func (h *SalaryHandler) ConfirmSalary(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的工资记录ID")
		return
	}

	confirmedBy := GetUserID(c)

	if err := h.salaryService.ConfirmSalary(id, confirmedBy); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "审核确认工资失败")
		return
	}

	Success(c, nil)
}

// PaySalary 发放工资
// @Summary      发放工资
// @Description  发放指定工资记录的工资（管理员）
// @Tags         工资管理
// @Accept       json
// @Produce      json
// @Param        id   path  int64                        true  "工资记录ID"
// @Param        request  body  map[string]interface{}  false  "发放信息"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /salaries/{id}/pay [post]
func (h *SalaryHandler) PaySalary(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 400, "无效的工资记录ID")
		return
	}

	var req struct {
		PayMethod int    `json:"pay_method"`
		PayRemark string `json:"pay_remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	paidBy := GetUserID(c)

	if err := h.salaryService.PaySalary(id, paidBy, req.PayMethod, req.PayRemark); err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "发放工资失败")
		return
	}

	Success(c, nil)
}
