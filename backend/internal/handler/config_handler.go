package handler

import (
	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

// SetConfigRequest 设置系统配置请求
type SetConfigRequest struct {
	Key    string `json:"config_key" binding:"required" example:"commission_rate"`
	Value  string `json:"config_value" binding:"required" example:"0.05"`
	Type   string `json:"config_type" example:"string"`
	Remark string `json:"remark" example:"提成比例"`
}

// UpdateConfigRequest 更新系统配置请求
type UpdateConfigRequest struct {
	Value string `json:"config_value" binding:"required" example:"0.08"`
}

// ConfigHandler 系统配置处理器
type ConfigHandler struct {
	configService *service.ConfigService
}

// NewConfigHandler 创建系统配置处理器实例
func NewConfigHandler(configService *service.ConfigService) *ConfigHandler {
	return &ConfigHandler{configService: configService}
}

// GetAll 获取所有系统配置
// @Summary      获取所有系统配置
// @Description  获取系统中所有配置项列表
// @Tags         系统配置
// @Accept       json
// @Produce      json
// @Success      200  {object}  handler.Response{data=[]interface{}}  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /configs [get]
func (h *ConfigHandler) GetAll(c *gin.Context) {
	configs, err := h.configService.GetAll()
	if err != nil {
		Error(c, 500, "获取系统配置失败")
		return
	}

	Success(c, configs)
}

// Get 获取单个配置
// @Summary      获取单个配置
// @Description  根据配置键获取对应的配置值
// @Tags         系统配置
// @Accept       json
// @Produce      json
// @Param        key   path  string  true  "配置键"
// @Success      200  {object}  handler.Response{data=map[string]string}  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /configs/{key} [get]
func (h *ConfigHandler) Get(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		Error(c, 400, "配置键不能为空")
		return
	}

	value, err := h.configService.Get(key)
	if err != nil {
		Error(c, 500, "获取配置失败")
		return
	}

	Success(c, gin.H{"key": key, "value": value})
}

// Set 设置系统配置
// @Summary      设置系统配置
// @Description  新增系统配置项（管理员）
// @Tags         系统配置
// @Accept       json
// @Produce      json
// @Param        request  body  handler.SetConfigRequest  true  "配置信息"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /configs [post]
func (h *ConfigHandler) Set(c *gin.Context) {
	var req SetConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误: "+err.Error())
		return
	}

	configType := req.Type
	if configType == "" {
		configType = "string"
	}

	if err := h.configService.Set(req.Key, req.Value, configType, req.Remark); err != nil {
		Error(c, 500, "设置配置失败")
		return
	}

	Success(c, nil)
}

// Update 更新系统配置
// @Summary      更新系统配置
// @Description  根据配置键更新配置值（管理员）
// @Tags         系统配置
// @Accept       json
// @Produce      json
// @Param        key       path  string                       true  "配置键"
// @Param        request   body  handler.UpdateConfigRequest   true  "更新配置请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /configs/{key} [put]
func (h *ConfigHandler) Update(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		Error(c, 400, "配置键不能为空")
		return
	}

	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误: "+err.Error())
		return
	}

	// 获取原有配置类型和备注，保持不变
	configType, remark, err := h.configService.GetConfigTypeAndRemark(key)
	if err != nil {
		// 如果获取失败，使用默认值
		configType = "string"
		remark = ""
	}

	if err := h.configService.Set(key, req.Value, configType, remark); err != nil {
		Error(c, 500, "更新配置失败")
		return
	}

	Success(c, nil)
}

// GetCommissionRates 获取所有提成比例配置
// @Summary      获取所有提成比例配置
// @Description  获取系统中所有提成比例相关的配置项
// @Tags         系统配置
// @Accept       json
// @Produce      json
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /configs/commission-rates [get]
func (h *ConfigHandler) GetCommissionRates(c *gin.Context) {
	rates, err := h.configService.GetCommissionRates()
	if err != nil {
		Error(c, 500, "获取提成比例配置失败")
		return
	}

	Success(c, rates)
}
