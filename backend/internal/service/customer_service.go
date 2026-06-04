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
	SourceType int8 `json:"source_type" example:"0"` // 0=自然进店 1=主动邀约 2=同行带单
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
	SourceType int8 `json:"source_type" example:"0"` // 0=自然进店 1=主动邀约 2=同行带单
	Remark string `json:"remark" example:"VIP客户"`
	Status *int `json:"status" example:1`
}

// ListCustomerRequest 客户列表查询请求
type ListCustomerRequest struct {
	StoreID    string `form:"store_id" example:"1"`
	Keyword    string `form:"keyword" example:"王五"`
	Level      string `form:"level" example:"1"`
	Page       int    `form:"page" example:1`
	PageSize   int    `form:"page_size" example:10`
	SalesmanID int64  `form:"salesman_id" example:"1"` // 业务员ID，用于过滤自己的客户
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
		SourceType:   req.SourceType,
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
	customers, total, err := s.custRepo.ListWithFilter(req.StoreID, req.Keyword, req.Level, req.Page, req.PageSize, req.SalesmanID)
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
	// 手机号变更时记录原手机号
	if req.Phone != "" && req.Phone != customer.Phone {
		if customer.Phone != "" {
			customer.OriginalPhone = customer.Phone
		}
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
	// 更新客户来源
	customer.SourceType = req.SourceType
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

// CustomerWithDraft 客户（含草稿状态）
type CustomerWithDraft struct {
	models.Customer
	SalesmanName string `json:"salesman_name"`
	HasDraft     bool   `json:"has_draft"`
	DraftID      int64  `json:"draft_id"`
	DraftItems   int    `json:"draft_items"`
}

// GetCustomersWithDraftStatus 获取业务员负责的客户列表（含草稿状态）
func (s *CustomerService) GetCustomersWithDraftStatus(salesmanID int64, storeID int64, keyword string, page, pageSize int) (*PageResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	// 查询业务员负责的客户
	query := s.db.Model(&models.Customer{}).Where("salesman_id = ? AND status = 1", salesmanID)
	if storeID > 0 {
		query = query.Where("store_id = ?", storeID)
	}
	if keyword != "" {
		query = query.Where("customer_name LIKE ? OR phone LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	var customers []models.Customer
	query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Preload("Salesman").Find(&customers)

	// 批量查询草稿状态
	customerIDs := make([]int64, len(customers))
	for i, c := range customers {
		customerIDs[i] = c.ID
	}

	// 查询这些客户的草稿订单（每个客户最新的草稿）
	type DraftInfo struct {
		CustomerID int64 `json:"customer_id"`
		OrderID    int64 `json:"order_id"`
		ItemCount  int   `json:"item_count"`
	}
	var drafts []DraftInfo
	if len(customerIDs) > 0 {
		// 先查询每个客户最新的草稿订单ID
		var draftOrders []models.Order
		s.db.Where("customer_id IN ? AND is_draft = 1", customerIDs).
			Order("id DESC").
			Find(&draftOrders)

		// 去重，只保留每个客户最新的草稿
		seen := make(map[int64]bool)
		for _, order := range draftOrders {
			cid := int64(0)
			if order.CustomerID != nil {
				cid = *order.CustomerID
			}
			if !seen[cid] {
				seen[cid] = true
				// 查询该订单的商品数量
				var itemCount int64
				s.db.Model(&models.OrderItem{}).Where("order_id = ?", order.ID).Count(&itemCount)
				drafts = append(drafts, DraftInfo{
					CustomerID: cid,
					OrderID:    order.ID,
					ItemCount:  int(itemCount),
				})
			}
		}
	}

	// 构建草稿映射
	draftMap := make(map[int64]DraftInfo)
	for _, d := range drafts {
		draftMap[d.CustomerID] = d
	}

	// 组装返回
	list := make([]CustomerWithDraft, len(customers))
	for i, c := range customers {
		list[i] = CustomerWithDraft{
			Customer: c,
		}
		if c.Salesman != nil {
			list[i].SalesmanName = c.Salesman.RealName
			if list[i].SalesmanName == "" {
				list[i].SalesmanName = c.Salesman.Username
			}
		}
		if draft, ok := draftMap[c.ID]; ok {
			list[i].HasDraft = true
			list[i].DraftID = draft.OrderID
			list[i].DraftItems = draft.ItemCount
		}
	}

	return &PageResult{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
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
