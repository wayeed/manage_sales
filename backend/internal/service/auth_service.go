package service

import (
	"errors"
	"time"

	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	appjwt "furniture-commission/internal/pkg/jwt"
	appmd5 "furniture-commission/internal/pkg/md5"
	"furniture-commission/internal/repository"

	"gorm.io/gorm"
)

// LoginResponse 登录响应
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User *UserVO `json:"user" example:{}`
}

// UserVO 用户视图对象
type UserVO struct {
	ID int64 `json:"id" example:1`
	Username string `json:"username" example:"zhangsan"`
	RealName string `json:"real_name" example:"张三"`
	Phone string `json:"phone" example:"13800138000"`
	Avatar string `json:"avatar" example:"https://example.com/avatar/default.png"`
	Role int `json:"role" example:1`
	Status int8 `json:"status" example:1`
	StoreID *int64 `json:"store_id" example:1`
	StoreName string `json:"store_name,omitempty" example:"总店"`
}

// UserDetail 用户详情（包含角色和权限）
type UserDetail struct {
	*UserVO
	Roles []models.Role `json:"roles" example:[]`
	Permissions []models.Permission `json:"permissions" example:[]`
}

// AuthService 认证服务
type AuthService struct {
	userRepo *repository.UserRepository
	permRepo *repository.PermissionRepository
}

// NewAuthService 创建认证服务实例
func NewAuthService(userRepo *repository.UserRepository, permRepo *repository.PermissionRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		permRepo: permRepo,
	}
}

// Login 用户登录
func (s *AuthService) Login(username, password, clientIP string) (*LoginResponse, error) {
	// 查找用户（支持用户名或手机号登录）
	var user *models.User
	var err error

	user, err = s.userRepo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user, err = s.userRepo.FindByPhone(username)
			if err != nil {
				return nil, &AppError{Code: apperrors.ErrUserNotFound, Message: apperrors.GetMessage(apperrors.ErrUserNotFound)}
			}
		} else {
			return nil, &AppError{Code: apperrors.InternalError, Message: "系统错误"}
		}
	}

	// 验证密码（MD5）
	if !appmd5.MD5Verify(password, user.Password) {
		return nil, &AppError{Code: apperrors.ErrPasswordWrong, Message: apperrors.GetMessage(apperrors.ErrPasswordWrong)}
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, &AppError{Code: apperrors.ErrUserDisabled, Message: apperrors.GetMessage(apperrors.ErrUserDisabled)}
	}

	// 生成JWT Token（不再携带 role，由中间件从 DB 查询）
	token, err := appjwt.GenerateToken(user.ID)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "令牌生成失败"}
	}

	// 更新最后登录时间和IP
	now := time.Now()
	s.userRepo.DB.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
		"last_login_at": now,
		"last_login_ip": clientIP,
	})

	// 构建返回数据
	storeName := ""
	if user.Store != nil {
		storeName = user.Store.StoreName
	}

	return &LoginResponse{
		Token: token,
		User: &UserVO{
			ID:        user.ID,
			Username:  user.Username,
			RealName:  user.RealName,
			Phone:     user.Phone,
			Avatar:    user.Avatar,
			Role:      user.Role,
			Status:    user.Status,
			StoreID:   user.StoreID,
			StoreName: storeName,
		},
	}, nil
}

// Logout 用户登出
func (s *AuthService) Logout(userID int64) error {
	// JWT是无状态的，登出主要在前端清除token
	// 这里可以记录登出日志等操作
	return nil
}

// GetCurrentUser 获取当前用户详情
func (s *AuthService) GetCurrentUser(userID int64) (*UserDetail, error) {
	user, err := s.userRepo.FindWithRoles(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.ErrUserNotFound, Message: apperrors.GetMessage(apperrors.ErrUserNotFound)}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	// 获取用户权限
	permissions, err := s.permRepo.FindByUserID(userID)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "获取权限失败"}
	}

	storeName := ""
	if user.Store != nil {
		storeName = user.Store.StoreName
	}

	return &UserDetail{
		UserVO: &UserVO{
			ID:        user.ID,
			Username:  user.Username,
			RealName:  user.RealName,
			Phone:     user.Phone,
			Avatar:    user.Avatar,
			Role:      user.Role,
			Status:    user.Status,
			StoreID:   user.StoreID,
			StoreName: storeName,
		},
		Roles:       user.Roles,
		Permissions: permissions,
	}, nil
}

// AppError 应用错误
type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}
