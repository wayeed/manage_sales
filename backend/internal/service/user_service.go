package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"furniture-commission/internal/models"
	appauthcache "furniture-commission/internal/pkg/authcache"
	apperrors "furniture-commission/internal/pkg/errors"
	appmd5 "furniture-commission/internal/pkg/md5"
	"furniture-commission/internal/repository"

	"gorm.io/gorm"
)

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	StoreID *int64 `json:"store_id" example:1`
	EmployeeNo string `json:"employee_no" example:"EMP001"`
	Username string `json:"username" binding:"required" example:"zhangsan"`
	Password string `json:"password" binding:"required,min=6" example:"123456"`
	RealName string `json:"real_name" example:"张三"`
	Phone string `json:"phone" example:"13800138000"`
	Email string `json:"email" example:"zhangsan@example.com"`
	DepartmentID *int64 `json:"department_id" example:1`
	Role int `json:"role" example:1`
	EntryDate string `json:"entry_date" example:"2024-01-01"`
	ProbationEndDate string `json:"probation_end_date" example:"2024-04-01"`
	IsFormal int8 `json:"is_formal" example:1`
	ParentID *int64 `json:"parent_id" example:2`
	ReferrerID *int64 `json:"referrer_id" example:3`
	BaseSalary float64 `json:"base_salary" example:5000.00`
	IDCard string `json:"id_card" example:"110101199001011234"`
	BankAccount string `json:"bank_account" example:"6222021234567890123"`
	BankName string `json:"bank_name" example:"中国工商银行"`
	RoleIDs []int64 `json:"role_ids" example:[1, 2]`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	StoreID *int64 `json:"store_id" example:1`
	EmployeeNo string `json:"employee_no" example:"EMP001"`
	RealName string `json:"real_name" example:"张三"`
	Phone string `json:"phone" example:"13800138000"`
	Email string `json:"email" example:"zhangsan@example.com"`
	DepartmentID *int64 `json:"department_id" example:1`
	Role int `json:"role" example:1`
	EntryDate string `json:"entry_date" example:"2024-01-01"`
	ProbationEndDate string `json:"probation_end_date" example:"2024-04-01"`
	IsFormal int8 `json:"is_formal" example:1`
	ParentID *int64 `json:"parent_id" example:2`
	ReferrerID *int64 `json:"referrer_id" example:3`
	BaseSalary float64 `json:"base_salary" example:5000.00`
	IDCard string `json:"id_card" example:"110101199001011234"`
	BankAccount string `json:"bank_account" example:"6222021234567890123"`
	BankName string `json:"bank_name" example:"中国工商银行"`
	Status int8 `json:"status" example:1`
}

// ListUserRequest 用户列表查询请求
type ListUserRequest struct {
	StoreID int `form:"store_id" example:1`
	Role int `form:"role" example:1`
	Status int `form:"status" example:1`
	Keyword string `form:"keyword" example:"张三"`
	Page int `form:"page" example:1`
	PageSize int `form:"page_size" example:10`
}

// PageResult 分页结果
type PageResult struct {
	List interface{} `json:"list" example:[]`
	Total int64 `json:"total" example:100`
	Page int `json:"page" example:1`
	PageSize int `json:"page_size" example:10`
}

// UserService 用户服务
type UserService struct {
	db       *gorm.DB
	userRepo *repository.UserRepository
	roleRepo *repository.RoleRepository
}

// NewUserService 创建用户服务实例
func NewUserService(db *gorm.DB, userRepo *repository.UserRepository, roleRepo *repository.RoleRepository) *UserService {
	return &UserService{
		db:       db,
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// Create 创建用户
func (s *UserService) Create(req *CreateUserRequest, createdBy int64) error {
	// 检查用户名是否已存在
	if existing, _ := s.userRepo.FindByUsername(req.Username); existing != nil {
		return &AppError{Code: apperrors.ErrDuplicateKey, Message: "用户名已存在"}
	}

	// 检查手机号是否已存在
	if req.Phone != "" {
		if existing, _ := s.userRepo.FindByPhone(req.Phone); existing != nil {
			return &AppError{Code: apperrors.ErrDuplicateKey, Message: "手机号已存在"}
		}
	}

	// 检查工号是否已存在
	if req.EmployeeNo != "" {
		if existing, _ := s.userRepo.FindByEmployeeNo(req.EmployeeNo); existing != nil {
			return &AppError{Code: apperrors.ErrDuplicateKey, Message: "工号已存在"}
		}
	}

	user := &models.User{
		StoreID:     req.StoreID,
		EmployeeNo:  req.EmployeeNo,
		Username:    req.Username,
		Password:    appmd5.MD5Encode(req.Password),
		RealName:    req.RealName,
		Phone:       req.Phone,
		Email:       req.Email,
		DepartmentID: req.DepartmentID,
		Role:        req.Role,
		Status:      1,
		IsFormal:    req.IsFormal,
		ParentID:    req.ParentID,
		ReferrerID:  req.ReferrerID,
		BaseSalary:  req.BaseSalary,
		IDCard:      req.IDCard,
		BankAccount: req.BankAccount,
		BankName:    req.BankName,
		CreatedBy:   &createdBy,
	}

	// 解析日期
	if req.EntryDate != "" {
		if t, err := time.Parse("2006-01-02", req.EntryDate); err == nil {
			user.EntryDate = &t
		}
	}
	if req.ProbationEndDate != "" {
		if t, err := time.Parse("2006-01-02", req.ProbationEndDate); err == nil {
			user.ProbationEndDate = &t
		}
	}

	// 创建用户
	if err := s.db.Create(user).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "创建用户失败"}
	}

	// 分配角色
	if len(req.RoleIDs) > 0 {
		if err := s.assignRoles(user.ID, req.RoleIDs); err != nil {
			return err
		}
	}

	return nil
}

// Update 更新用户
func (s *UserService) Update(id int64, req *UpdateUserRequest) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrUserNotFound, Message: apperrors.GetMessage(apperrors.ErrUserNotFound)}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	// 检查手机号是否被其他人使用
	if req.Phone != "" && req.Phone != user.Phone {
		if existing, _ := s.userRepo.FindByPhone(req.Phone); existing != nil && existing.ID != id {
			return &AppError{Code: apperrors.ErrDuplicateKey, Message: "手机号已存在"}
		}
		user.Phone = req.Phone
	}

	// 检查工号是否被其他人使用
	if req.EmployeeNo != "" && req.EmployeeNo != user.EmployeeNo {
		if existing, _ := s.userRepo.FindByEmployeeNo(req.EmployeeNo); existing != nil && existing.ID != id {
			return &AppError{Code: apperrors.ErrDuplicateKey, Message: "工号已存在"}
		}
		user.EmployeeNo = req.EmployeeNo
	}

	// 更新字段
	user.StoreID = req.StoreID
	user.RealName = req.RealName
	user.Email = req.Email
	user.DepartmentID = req.DepartmentID
	user.Role = req.Role
	user.IsFormal = req.IsFormal
	user.ParentID = req.ParentID
	user.ReferrerID = req.ReferrerID
	user.BaseSalary = req.BaseSalary
	user.IDCard = req.IDCard
	user.BankAccount = req.BankAccount
	user.BankName = req.BankName
	user.Status = req.Status

	// 解析日期
	if req.EntryDate != "" {
		if t, err := time.Parse("2006-01-02", req.EntryDate); err == nil {
			user.EntryDate = &t
		}
	}
	if req.ProbationEndDate != "" {
		if t, err := time.Parse("2006-01-02", req.ProbationEndDate); err == nil {
			user.ProbationEndDate = &t
		}
	}

	if err := s.db.Save(user).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "更新用户失败"}
	}

	return nil
}

// Delete 删除用户（软删除，设置status=2）
func (s *UserService) Delete(id int64) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrUserNotFound, Message: apperrors.GetMessage(apperrors.ErrUserNotFound)}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if err := s.db.Delete(user).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "删除用户失败"}
	}

	return nil
}

// UpdateStatus 更新用户状态（启用/禁用）
func (s *UserService) UpdateStatus(id int64, status int) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrUserNotFound, Message: apperrors.GetMessage(apperrors.ErrUserNotFound)}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if err := s.db.Model(user).Update("status", status).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "更新状态失败"}
	}

	return nil
}

// ResetPassword 重置密码，返回新密码
func (s *UserService) ResetPassword(id int64) (string, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", &AppError{Code: apperrors.ErrUserNotFound, Message: apperrors.GetMessage(apperrors.ErrUserNotFound)}
		}
		return "", &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	// 生成随机密码
	newPassword := generateRandomPassword(8)
	hashedPassword := appmd5.MD5Encode(newPassword)

	if err := s.db.Model(user).Update("password", hashedPassword).Error; err != nil {
		return "", &AppError{Code: apperrors.InternalError, Message: "重置密码失败"}
	}

	return newPassword, nil
}

// List 获取用户列表
func (s *UserService) List(req *ListUserRequest) (*PageResult, error) {
	users, total, err := s.userRepo.ListWithFilter(req.StoreID, req.Role, req.Status, req.Keyword, req.Page, req.PageSize)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询用户列表失败"}
	}

	return &PageResult{
		List:     users,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetDetail 获取用户详情
func (s *UserService) GetDetail(id int64) (*UserDetail, error) {
	user, err := s.userRepo.FindWithRoles(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.ErrUserNotFound, Message: apperrors.GetMessage(apperrors.ErrUserNotFound)}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "系统错误"}
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
		Roles: user.Roles,
	}, nil
}

// AssignRole 分配角色
func (s *UserService) AssignRole(userID int64, roleIDs []int64) error {
	// 检查用户是否存在
	if _, err := s.userRepo.FindByID(userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrUserNotFound, Message: apperrors.GetMessage(apperrors.ErrUserNotFound)}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if err := s.assignRoles(userID, roleIDs); err != nil {
		return err
	}

	// 清除角色缓存，使权限立即生效
	appauthcache.Invalidate(userID)

	return nil
}

// assignRoles 内部方法：分配角色
func (s *UserService) assignRoles(userID int64, roleIDs []int64) error {
	// 先清除已有角色
	if err := s.db.Where("user_id = ?", userID).Delete(&models.UserRole{}).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "清除角色失败"}
	}

	// 批量添加新角色
	if len(roleIDs) > 0 {
		userRoles := make([]models.UserRole, len(roleIDs))
		for i, roleID := range roleIDs {
			userRoles[i] = models.UserRole{
				UserID: userID,
				RoleID: roleID,
			}
		}
		if err := s.db.Create(&userRoles).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "分配角色失败"}
		}
	}

	return nil
}

// generateRandomPassword 生成随机密码
func generateRandomPassword(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		// fallback
		return fmt.Sprintf("Reset%d", time.Now().Unix())
	}
	return hex.EncodeToString(b)[:length]
}
