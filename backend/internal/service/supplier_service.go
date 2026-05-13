package service

import (
	"errors"

	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	"furniture-commission/internal/repository"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// CreateSupplierRequest 创建供应商请求
type CreateSupplierRequest struct {
	StoreID int64 `json:"store_id" example:1`
	SupplierCode string `json:"supplier_code" binding:"required" example:"SUP001"`
	SupplierName string `json:"supplier_name" binding:"required" example:"某某家具制造厂"`
	ContactPerson string `json:"contact_person" example:"孙七"`
	ContactPhone string `json:"contact_phone" example:"13500135000"`
	Address string `json:"address" example:"广东省佛山市某某工业区"`
	BusinessScope string `json:"business_scope" example:"沙发、床、餐桌"`
	BankName string `json:"bank_name" example:"中国农业银行"`
	BankAccount string `json:"bank_account" example:"6228481234567890789"`
	TaxNo string `json:"tax_no" example:"91440600MA5XXXXX"`
	Remark string `json:"remark" example:"主要沙发供应商"`
}

// UpdateSupplierRequest 更新供应商请求
type UpdateSupplierRequest struct {
	SupplierName string `json:"supplier_name" example:"某某家具制造厂"`
	ContactPerson string `json:"contact_person" example:"孙七"`
	ContactPhone string `json:"contact_phone" example:"13500135000"`
	Address string `json:"address" example:"广东省佛山市某某工业区"`
	BusinessScope string `json:"business_scope" example:"沙发、床、餐桌"`
	BankName string `json:"bank_name" example:"中国农业银行"`
	BankAccount string `json:"bank_account" example:"6228481234567890789"`
	TaxNo string `json:"tax_no" example:"91440600MA5XXXXX"`
	Remark string `json:"remark" example:"主要沙发供应商"`
	Status *int8 `json:"status" example:1`
}

// AddSupplierProductRequest 添加供应商商品请求
type AddSupplierProductRequest struct {
	SKUID int64 `json:"sku_id" binding:"required" example:1`
	SupplyPrice float64 `json:"supply_price" example:4800.00`
	MinOrderQuantity int `json:"min_order_quantity" example:5`
	LeadTime *int `json:"lead_time" example:7`
	IsDefault int8 `json:"is_default" example:1`
}

// ListSupplierRequest 供应商列表查询请求
type ListSupplierRequest struct {
	StoreID int64 `form:"store_id" example:1`
	Status *int8 `form:"status" example:1`
	Keyword string `form:"keyword" example:"某某"`
	Page int `form:"page" example:1`
	PageSize int `form:"page_size" example:10`
}

// SupplierService 供应商服务
type SupplierService struct {
	db       *gorm.DB
	supplierRepo *repository.BaseRepository[models.Supplier]
	spRepo      *repository.BaseRepository[models.SupplierProduct]
}

// NewSupplierService 创建供应商服务实例
func NewSupplierService(db *gorm.DB) *SupplierService {
	return &SupplierService{
		db:           db,
		supplierRepo: repository.NewBaseRepository[models.Supplier](db),
		spRepo:       repository.NewBaseRepository[models.SupplierProduct](db),
	}
}

// Create 创建供应商
func (s *SupplierService) Create(req *CreateSupplierRequest) error {
	supplier := &models.Supplier{
		StoreID:       req.StoreID,
		SupplierCode:  req.SupplierCode,
		SupplierName:  req.SupplierName,
		ContactPerson: req.ContactPerson,
		ContactPhone:  req.ContactPhone,
		Address:       req.Address,
		BusinessScope: req.BusinessScope,
		BankName:      req.BankName,
		BankAccount:   req.BankAccount,
		TaxNo:         req.TaxNo,
		Remark:        req.Remark,
		Status:        1,
	}

	if err := s.db.Create(supplier).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "创建供应商失败"}
	}
	return nil
}

// Update 更新供应商
func (s *SupplierService) Update(id int64, req *UpdateSupplierRequest) error {
	supplier, err := s.supplierRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "供应商不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if req.SupplierName != "" {
		supplier.SupplierName = req.SupplierName
	}
	if req.ContactPerson != "" {
		supplier.ContactPerson = req.ContactPerson
	}
	if req.ContactPhone != "" {
		supplier.ContactPhone = req.ContactPhone
	}
	supplier.Address = req.Address
	supplier.BusinessScope = req.BusinessScope
	supplier.BankName = req.BankName
	supplier.BankAccount = req.BankAccount
	supplier.TaxNo = req.TaxNo
	supplier.Remark = req.Remark
	if req.Status != nil {
		supplier.Status = *req.Status
	}

	if err := s.db.Save(supplier).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "更新供应商失败"}
	}
	return nil
}

// Delete 删除供应商
func (s *SupplierService) Delete(id int64) error {
	_, err := s.supplierRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "供应商不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if err := s.db.Delete(&models.Supplier{}, id).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "删除供应商失败"}
	}
	return nil
}

// List 获取供应商列表
func (s *SupplierService) List(req *ListSupplierRequest) (*PageResult, error) {
	db := s.db.Model(&models.Supplier{})

	if req.StoreID > 0 {
		db = db.Where("store_id = ?", req.StoreID)
	}
	if req.Status != nil {
		db = db.Where("status = ?", *req.Status)
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		db = db.Where("supplier_name LIKE ? OR supplier_code LIKE ? OR contact_person LIKE ?", like, like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询供应商列表失败"}
	}

	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var suppliers []models.Supplier
	if err := db.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").
		Find(&suppliers).Error; err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询供应商列表失败"}
	}

	return &PageResult{
		List:     suppliers,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetDetail 获取供应商详情
func (s *SupplierService) GetDetail(id int64) (*models.Supplier, error) {
	supplier, err := s.supplierRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.NotFound, Message: "供应商不存在"}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}
	return supplier, nil
}

// AddProduct 添加供应商商品关联
func (s *SupplierService) AddProduct(supplierID int64, req *AddSupplierProductRequest) error {
	_, err := s.supplierRepo.FindByID(supplierID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "供应商不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	sp := &models.SupplierProduct{
		SupplierID:       supplierID,
		SKUID:            req.SKUID,
		SupplyPrice:      decimal.NewFromFloat(req.SupplyPrice),
		MinOrderQuantity: req.MinOrderQuantity,
		LeadTime:         req.LeadTime,
		IsDefault:        req.IsDefault,
		Status:           1,
	}

	if err := s.db.Create(sp).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "添加供应商商品失败"}
	}
	return nil
}

// RemoveProduct 移除供应商商品关联
func (s *SupplierService) RemoveProduct(supplierID, skuID int64) error {
	if err := s.db.Where("supplier_id = ? AND sku_id = ?", supplierID, skuID).Delete(&models.SupplierProduct{}).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "移除供应商商品失败"}
	}
	return nil
}

// GetProducts 获取供应商商品列表
func (s *SupplierService) GetProducts(supplierID int64) ([]models.SupplierProduct, error) {
	var products []models.SupplierProduct
	if err := s.db.Preload("SKU").Where("supplier_id = ?", supplierID).Find(&products).Error; err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询供应商商品列表失败"}
	}
	return products, nil
}
