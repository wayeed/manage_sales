package service

import (
	"errors"

	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	"furniture-commission/internal/repository"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// CreateProductRequest 创建商品请求
type CreateProductRequest struct {
	StoreID int64 `json:"store_id" example:1`
	CategoryID *int64 `json:"category_id" example:1`
	ProductCode string `json:"product_code" example:"P001"`
	ProductName string `json:"product_name" binding:"required" example:"真皮沙发"`
	Brand string `json:"brand" example:"品牌A"`
	ProductImage string `json:"product_image" example:"https://example.com/product/sofa.jpg"`
	Description string `json:"description" example:"高档真皮沙发，三座位"`
	ListPrice float64 `json:"list_price" example:8999.00`
	MinPrice float64 `json:"min_price" example:7200.00`
	ReferenceCost float64 `json:"reference_cost" example:5000.00`
	CostPrice float64 `json:"cost_price" example:6000.00`
	TotalCostRate float64 `json:"total_cost_rate" example:1.20`
	WarningStock int `json:"warning_stock" example:10`
}

// UpdateProductRequest 更新商品请求
type UpdateProductRequest struct {
	CategoryID *int64 `json:"category_id" example:1`
	ProductName string `json:"product_name" example:"真皮沙发"`
	Brand string `json:"brand" example:"品牌A"`
	ProductImage string `json:"product_image" example:"https://example.com/product/sofa.jpg"`
	Description string `json:"description" example:"高档真皮沙发，三座位"`
	ListPrice *float64 `json:"list_price" example:8999.00`
	MinPrice *float64 `json:"min_price" example:7200.00`
	ReferenceCost *float64 `json:"reference_cost" example:5000.00`
	CostPrice *float64 `json:"cost_price" example:6000.00`
	TotalCostRate *float64 `json:"total_cost_rate" example:1.20`
	WarningStock *int `json:"warning_stock" example:10`
}

// ListProductRequest 商品列表查询请求
type ListProductRequest struct {
	StoreID int64 `form:"store_id" example:1`
	CategoryID *int64 `form:"category_id" example:1`
	Status *int8 `form:"status" example:1`
	Keyword string `form:"keyword" example:"沙发"`
	Page int `form:"page" example:1`
	PageSize int `form:"page_size" example:10`
}

// ProductService 商品服务
type ProductService struct {
	db          *gorm.DB
	productRepo *repository.ProductRepository
	configRepo  *repository.SystemConfigRepository
}

// NewProductService 创建商品服务实例
func NewProductService(db *gorm.DB, productRepo *repository.ProductRepository, configRepo *repository.SystemConfigRepository) *ProductService {
	return &ProductService{
		db:          db,
		productRepo: productRepo,
		configRepo:  configRepo,
	}
}

// getConfigRate 获取系统配置的比率，失败时返回默认值
func (s *ProductService) getConfigRate(key string, defaultVal float64) float64 {
	valStr, err := s.configRepo.Get(key)
	if err != nil || valStr == "" {
		return defaultVal
	}
	val, err := decimal.NewFromString(valStr)
	if err != nil {
		return defaultVal
	}
	f, _ := val.Float64()
	return f
}

// Create 创建商品
func (s *ProductService) Create(req *CreateProductRequest, createdBy int64) error {
	// 检查编码是否重复
	if existing, _ := s.productRepo.FindByCode(req.StoreID, req.ProductCode); existing != nil {
		return &AppError{Code: apperrors.ErrDuplicateKey, Message: "商品编码已存在"}
	}

	// 获取系统配置的系数
	costRate := s.getConfigRate("cost_rate", 1.2)
	minDiscountRate := s.getConfigRate("min_discount_rate", 0.9)

	// 自动计算成本价：如果没传 cost_price，则 = 进货价 × 成本系数
	costPrice := req.CostPrice
	totalCostRate := req.TotalCostRate
	if totalCostRate <= 0 {
		totalCostRate = costRate
	}
	if costPrice <= 0 && req.ReferenceCost > 0 {
		costPrice = req.ReferenceCost * totalCostRate
	}

	// 自动计算最低价：如果没传 min_price，则 = 挂牌价 × 最低折扣系数
	minPrice := req.MinPrice
	if minPrice <= 0 && req.ListPrice > 0 {
		minPrice = req.ListPrice * minDiscountRate
	}

	product := &models.Product{
		StoreID:       req.StoreID,
		CategoryID:    req.CategoryID,
		ProductCode:   req.ProductCode,
		ProductName:   req.ProductName,
		Brand:         req.Brand,
		ProductImage:  req.ProductImage,
		Description:   req.Description,
		ListPrice:     decimal.NewFromFloat(req.ListPrice),
		MinPrice:      decimal.NewFromFloat(minPrice),
		ReferenceCost: decimal.NewFromFloat(req.ReferenceCost),
		CostPrice:     decimal.NewFromFloat(costPrice),
		TotalCostRate: decimal.NewFromFloat(totalCostRate),
		WarningStock:  req.WarningStock,
		Status:        1,
		CreatedBy:     &createdBy,
	}

	if err := s.db.Create(product).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "创建商品失败"}
	}
	return nil
}

// Update 更新商品
func (s *ProductService) Update(id int64, req *UpdateProductRequest) error {
	product, err := s.productRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "商品不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if req.CategoryID != nil {
		product.CategoryID = req.CategoryID
	}
	if req.ProductName != "" {
		product.ProductName = req.ProductName
	}
	if req.Brand != "" {
		product.Brand = req.Brand
	}
	product.ProductImage = req.ProductImage
	product.Description = req.Description
	if req.ListPrice != nil {
		product.ListPrice = decimal.NewFromFloat(*req.ListPrice)
	}
	if req.MinPrice != nil {
		product.MinPrice = decimal.NewFromFloat(*req.MinPrice)
	}
	if req.ReferenceCost != nil {
		product.ReferenceCost = decimal.NewFromFloat(*req.ReferenceCost)
	}
	if req.CostPrice != nil {
		product.CostPrice = decimal.NewFromFloat(*req.CostPrice)
	}
	if req.TotalCostRate != nil {
		product.TotalCostRate = decimal.NewFromFloat(*req.TotalCostRate)
	}
	if req.WarningStock != nil {
		product.WarningStock = *req.WarningStock
	}

	if err := s.db.Save(product).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "更新商品失败"}
	}
	return nil
}

// Delete 删除商品
func (s *ProductService) Delete(id int64) error {
	_, err := s.productRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "商品不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if err := s.db.Delete(&models.Product{}, id).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "删除商品失败"}
	}
	return nil
}

// UpdateStatus 更新商品状态（上下架）
func (s *ProductService) UpdateStatus(id int64, status int8) error {
	product, err := s.productRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "商品不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if err := s.db.Model(product).Update("status", status).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "更新商品状态失败"}
	}
	return nil
}

// List 获取商品列表
func (s *ProductService) List(req *ListProductRequest) (*PageResult, error) {
	products, total, err := s.productRepo.ListWithFilter(req.StoreID, req.CategoryID, req.Status, req.Keyword, req.Page, req.PageSize)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询商品列表失败"}
	}

	// 自动计算成本价 = 进货价 × 成本系数
	for i := range products {
		if products[i].TotalCostRate.GreaterThan(decimal.Zero) {
			products[i].CostPrice = products[i].ReferenceCost.Mul(products[i].TotalCostRate)
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
		List:     products,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetDetail 获取商品详情
func (s *ProductService) GetDetail(id int64) (*models.Product, error) {
	product, err := s.productRepo.FindByIDWithSKUs(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.NotFound, Message: "商品不存在"}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}
	// 自动计算成本价 = 进货价 × 成本系数
	if product.TotalCostRate.GreaterThan(decimal.Zero) {
		product.CostPrice = product.ReferenceCost.Mul(product.TotalCostRate)
	}
	return product, nil
}

// BatchImport 批量导入商品（预留）
func (s *ProductService) BatchImport(storeID int64, createdBy int64, items []CreateProductRequest) (int, error) {
	successCount := 0
	for _, item := range items {
		item.StoreID = storeID
		if err := s.Create(&item, createdBy); err != nil {
			continue
		}
		successCount++
	}
	return successCount, nil
}
