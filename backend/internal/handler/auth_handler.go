package handler

import (
	"furniture-commission/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	contextUserID  = "user_id"
	contextRole    = "role"
	contextStoreID = "storeID"
)

// GetUserID 从上下文中获取用户ID
func GetUserID(c *gin.Context) int64 {
	if userID, exists := c.Get(contextUserID); exists {
		if id, ok := userID.(int64); ok {
			return id
		}
	}
	return 0
}

// GetStoreID 从上下文中获取门店ID
func GetStoreID(c *gin.Context) int64 {
	if storeID, exists := c.Get(contextStoreID); exists {
		if id, ok := storeID.(int64); ok {
			return id
		}
	}
	return 0
}

// GetRole 从上下文中获取用户角色
func GetRole(c *gin.Context) int {
	if role, exists := c.Get(contextRole); exists {
		if r, ok := role.(int); ok {
			return r
		}
	}
	return 0
}

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// loginRequest 登录请求
type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 用户登录
// @Summary      用户登录
// @Description  通过用户名和密码进行登录，返回JWT令牌
// @Tags         认证管理
// @Accept       json
// @Produce      json
// @Param        request  body  loginRequest  true  "登录请求"
// @Success      200  {object}  handler.Response  "成功"
// @Failure      200  {object}  handler.Response  "业务错误"
// @Router       /login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, "请求参数错误")
		return
	}

	resp, err := h.authService.Login(req.Username, req.Password, c.ClientIP())
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "登录失败")
		return
	}

	Success(c, resp)
}

// Logout 用户登出
// @Summary      用户登出
// @Description  退出登录，使当前JWT令牌失效
// @Tags         认证管理
// @Accept       json
// @Produce      json
// @Success      200  {object}  handler.Response  "成功"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID := GetUserID(c)
	if err := h.authService.Logout(userID); err != nil {
		Error(c, 500, "登出失败")
		return
	}
	Success(c, nil)
}

// GetCurrentUser 获取当前用户信息
// @Summary      获取当前用户信息
// @Description  根据JWT令牌获取当前登录用户的详细信息
// @Tags         认证管理
// @Accept       json
// @Produce      json
// @Success      200  {object}  handler.Response  "成功"
// @Failure      401  {object}  handler.Response  "未认证"
// @Security     BearerAuth
// @Router       /users/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID := GetUserID(c)
	detail, err := h.authService.GetCurrentUser(userID)
	if err != nil {
		if appErr, ok := err.(*service.AppError); ok {
			Error(c, appErr.Code, appErr.Message)
			return
		}
		Error(c, 500, "获取用户信息失败")
		return
	}

	Success(c, detail)
}
