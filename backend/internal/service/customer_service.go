package service

import (
	"errors"
	"fmt"
	"time"

	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	"furniture-commission/internal/repository"

	"gorm.io/gorm"
)

// CustomerService 客户服务
type CustomerService struct {
	db       *gorm.DB
	custRepo *repository.CustomerRepository
}

// NewCustomerService 创建客户服务实例
func NewCustomerService(db *gorm.DB, custRepo *repository.CustomerRepository) *CustomerService {
	return &CustomerService{db: db, custRepo: custRepo}
}

// CreateCustomerRequest 创建客户请求
type CreateCustomerRequest struct {
	StoreID int64 `json:"store_id" example:1`
	CustomerName string `json:"customer_name" binding:"required" example:"王五"`
	Phone string `json:"phone" binding:"required" example:"13700137000"`
	Email string `json:"email" example:"wangwu@example.com"`
	Address string `json:"address" example:"上海市浦东新区某某路456号"`
	Gender int8 `json:"gender" example:1`
	Birthday string `json:"birthday" example:"1990-05-15"`
	Level int8 `json:"level" example:1`
	Remark string `json:"remark" example:"VIP客户"`
}

// UpdateCustomerRequest 更新客户请求
type UpdateCustomerRequest struct {
	CustomerName string `json:"customer_name" example:"王五"`
	Phone string `json:"phone" example:"13700137000"`
	Email string `json:"email" example:"wangwu@example.com"`
	Address string `json:"address" example:"上海市浦东新区某某路456号"`
	Gender *int `json:"gender" example:1`
	Birthday string `json:"birthday" example:"1990-05-15"`
	Level *int `json:"level" example:1`
	Remark string `json:"remark" example:"VIP客户"`
	Status *int `json:"status" example:1`
}

// ListCustomerRequest 客户列表查询请求
type ListCustomerRequest struct {
	StoreID string `form:"store_id" example:"1"`
	Keyword string `form:"keyword" example:"王五"`
	Level string `form:"level" example:"1"`
	Page int `form:"page" example:1`
	PageSize int `form:"page_size" example:10`
}

// AddFollowUpRequest 添加跟进记录请求
type AddFollowUpRequest struct {
	CustomerID int64 `json:"customer_id" example:1`
	FollowType int8 `json:"follow_type" example:1`
	Content string `json:"content" binding:"required" example:"电话沟通，客户对产品感兴趣"`
	NextFollowDate string `json:"next_follow_date" example:"2024-02-01"`
	NextFollowContent string `json:"next_follow_content" example:"上门拜访"`
	IsDeal int8 `json:"is_deal" example:0`
}

// Create 创建客户
func (s *CustomerService) Create(req *CreateCustomerRequest, createdBy int64) (*models.Customer, error) {
	// 生成客户编码：C + 年月日 + 6位序号
	customerCode := generateCustomerCode(s.db)

	customer := &models.Customer{
		StoreID:      req.StoreID,
		CustomerCode: customerCode,
		CustomerName: req.CustomerName,
		Phone:        req.Phone,
		Email:        req.Email,
		Address:      req.Address,
		Gender:       req.Gender,
		Level:        req.Level,
		Remark:       req.Remark,
		Status:       1,
		CreatedBy:    &createdBy,
		SalesmanID:   &createdBy,  // 默认负责业务员=创建者
	}

	if req.Birthday != "" {
		birthday, err := time.Parse("2006-01-02", req.Birthday)
		if err == nil {
			customer.Birthday = &birthday
		}
	}

	if err := s.custRepo.Create(customer); err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "创建客户失败"}
	}
	return customer, nil
}

// CustomerListResponse 客户列表响应（含业务员信息）
type CustomerListResponse struct {
	models.Customer
	SalesmanName string `json:"salesman_name"`
}

// List 客户列表
func (s *CustomerService) List(req *ListCustomerRequest) (*PageResult, error) {
	customers, total, err := s.custRepo.ListWithFilter(req.StoreID, req.Keyword, req.Level, req.Page, req.PageSize)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询客户列表失败"}
	}

	// 转换响应，添加业务员姓名
	list := make([]CustomerListResponse, len(customers))
	for i, c := range customers {
		list[i] = CustomerListResponse{
			Customer: c,
		}
		if c.Salesman != nil {
			list[i].SalesmanName = c.Salesman.RealName
			if list[i].SalesmanName == "" {
				list[i].SalesmanName = c.Salesman.Username
			}
		}
	}

	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	return &PageResult{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Update 更新客户
func (s *CustomerService) Update(id int64, req *UpdateCustomerRequest) error {
	customer, err := s.custRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrOrderNotFound, Message: "客户不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if req.CustomerName != "" {
		customer.CustomerName = req.CustomerName
	}
	if req.Phone != "" {
		customer.Phone = req.Phone
	}
	if req.Email != "" {
		customer.Email = req.Email
	}
	if req.Address != "" {
		customer.Address = req.Address
	}
	if req.Gender != nil {
		customer.Gender = int8(*req.Gender)
	}
	if req.Birthday != "" {
		birthday, err := time.Parse("2006-01-02", req.Birthday)
		if err == nil {
			customer.Birthday = &birthday
		}
	}
	if req.Level != nil {
		customer.Level = int8(*req.Level)
	}
	if req.Remark != "" {
		customer.Remark = req.Remark
	}
	if req.Status != nil {
		customer.Status = int8(*req.Status)
	}

	if err := s.custRepo.Update(customer); err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "更新客户失败"}
	}
	return nil
}

// Delete 删除客户
func (s *CustomerService) Delete(id int64) error {
	if err := s.custRepo.Delete(id); err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "删除客户失败"}
	}
	return nil
}

// GetDetail 获取客户详情
func (s *CustomerService) GetDetail(id int64) (*models.Customer, error) {
	customer, err := s.custRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.ErrOrderNotFound, Message: "客户不存在"}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}
	return customer, nil
}

// AddFollowUp 添加跟进记录
func (s *CustomerService) AddFollowUp(req *AddFollowUpRequest, followerID int64) error {
	followUp := &models.CustomerFollowUp{
		CustomerID: req.CustomerID,
		FollowerID: followerID,
		FollowType: req.FollowType,
		Content:    req.Content,
		IsDeal:     req.IsDeal,
	}

	if req.NextFollowDate != "" {
		nextDate, err := time.Parse("2006-01-02", req.NextFollowDate)
		if err == nil {
			followUp.NextFollowDate = &nextDate
		}
	}
	if req.NextFollowContent != "" {
		followUp.NextFollowContent = req.NextFollowContent
	}

	if err := s.db.Create(followUp).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "添加跟进记录失败"}
	}
	return nil
}

// GetFollowUps 获取跟进记录
func (s *CustomerService) GetFollowUps(customerID int64) ([]models.CustomerFollowUp, error) {
	var followUps []models.CustomerFollowUp
	err := s.db.Preload("Follower").
		Where("customer_id = ?", customerID).
		Order("id DESC").
		Find(&followUps).Error
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询跟进记录失败"}
	}
	return followUps, nil
}

// UpdateOrderStats 更新客户累计统计
func (s *CustomerService) UpdateOrderStats(customerID int64, totalOrders int, totalAmount, totalProfit float64) error {
	if err := s.custRepo.UpdateStats(customerID, totalOrders, totalAmount, totalProfit); err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "更新客户统计失败"}
	}
	return nil
}

// generateCustomerCode 生成客户编码：C + 年月日 + 6位序号
func generateCustomerCode(db *gorm.DB) string {
	date := time.Now().Format("20060102")
	prefix := "C" + date

	var count int64
	db.Model(&models.Customer{}).Where("customer_code LIKE ?", prefix+"%").Count(&count)
	seq := count + 1
	return fmt.Sprintf("%s%06d", prefix, seq)
}
