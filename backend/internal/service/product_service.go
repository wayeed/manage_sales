package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	"furniture-commission/internal/pkg/excel"
	"furniture-commission/internal/repository"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SKURequest SKU请求
type SKURequest struct {
	SKUCode   string          `json:"sku_code" example:"SKU001"`
	SKUName   string          `json:"sku_name" example:"真皮沙发-红色"`
	Attributes string         `json:"attributes" example:"{\"颜色\":\"红色\"}"`
	Barcode   string          `json:"barcode" example:"6901234567890"`
	SalePrice float64         `json:"sale_price" example:"8999.00"`
	CostPrice float64         `json:"cost_price" example:"6000.00"`
}

// CreateProductRequest 创建商品请求
type CreateProductRequest struct {
	StoreID int64 `json:"store_id" example:1`
	CategoryID *int64 `json:"category_id" example:1`
	ProductCode string `json:"product_code" example:"P001"`
	ProductName string `json:"product_name" binding:"required" example:"真皮沙发"`
	Brand string `json:"brand" example:"品牌A"`
	Style string `json:"style" example:""`
	Unit string `json:"unit" example:"件"`
	Series string `json:"series" example:"现代系列"`
	SubCategory string `json:"sub_category" example:"A"`
	ProductImage string `json:"product_image" example:"https://example.com/product/sofa.jpg"`
	Description string `json:"description" example:"高档真皮沙发，三座位"`
	ListPrice float64 `json:"list_price" example:8999.00`
	MinPrice float64 `json:"min_price" example:7200.00`
	ReferenceCost float64 `json:"reference_cost" example:5000.00`
	CostPrice float64 `json:"cost_price" example:6000.00`
	TotalCostRate float64 `json:"total_cost_rate" example:1.20`
	WarningStock int `json:"warning_stock" example:10`
	SKUs []SKURequest `json:"skus"`
}

// UpdateProductRequest 更新商品请求
type UpdateProductRequest struct {
	CategoryID *int64 `json:"category_id" example:1`
	ProductName string `json:"product_name" example:"真皮沙发"`
	Brand string `json:"brand" example:"品牌A"`
	Style *string `json:"style" example:""`
	Unit *string `json:"unit" example:"件"`
	Series *string `json:"series" example:"现代系列"`
	SubCategory *string `json:"sub_category" example:"A"`
	ProductImage string `json:"product_image" example:"https://example.com/product/sofa.jpg"`
	Description string `json:"description" example:"高档真皮沙发，三座位"`
	ListPrice *float64 `json:"list_price" example:8999.00`
	MinPrice *float64 `json:"min_price" example:7200.00`
	ReferenceCost *float64 `json:"reference_cost" example:5000.00`
	CostPrice *float64 `json:"cost_price" example:6000.00`
	TotalCostRate *float64 `json:"total_cost_rate" example:1.20`
	WarningStock *int `json:"warning_stock" example:10`
	SKUs []SKURequest `json:"skus"`
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
		Style:         req.Style,
		Unit:          req.Unit,
		Series:        req.Series,
		SubCategory:   req.SubCategory,
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

	// 单位默认值
	if product.Unit == "" {
		product.Unit = "件"
	}

	// 使用事务创建商品和SKU
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 创建商品
		if err := tx.Create(product).Error; err != nil {
			return err
		}

		// 创建SKU
			if len(req.SKUs) > 0 {
				for _, skuReq := range req.SKUs {
					sku := &models.ProductSKU{
						ProductID:  product.ID,
						SKUCode:    skuReq.SKUCode,
						SKUName:    skuReq.SKUName,
						Attributes: datatypes.JSON(skuReq.Attributes),
						Barcode:    skuReq.Barcode,
						Status:     1,
					}
					if err := tx.Create(sku).Error; err != nil {
						return err
					}
				}
			}

		return nil
	})

	if err != nil {
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
	if req.Style != nil {
		product.Style = *req.Style
	}
	if req.Unit != nil {
		product.Unit = *req.Unit
	}
	if req.Series != nil {
		product.Series = *req.Series
	}
	if req.SubCategory != nil {
		product.SubCategory = *req.SubCategory
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

	// 使用事务更新商品和SKU
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 更新商品
		if err := tx.Save(product).Error; err != nil {
			return fmt.Errorf("保存商品失败: %w", err)
		}

		// 更新SKU（删除旧SKU，创建新SKU）
		if len(req.SKUs) > 0 {
			// 先查询该商品的所有旧SKU编码
			var oldSkus []models.ProductSKU
			if err := tx.Where("product_id = ?", product.ID).Find(&oldSkus).Error; err != nil {
				return fmt.Errorf("查询旧SKU失败: %w", err)
			}

			// 收集需要保留的SKU编码（新提交的SKU编码）
			newSkuCodes := make(map[string]bool)
			for _, skuReq := range req.SKUs {
				if skuReq.SKUCode != "" {
					newSkuCodes[skuReq.SKUCode] = true
				}
			}

			// 删除不再需要的旧SKU（编码不在新列表中的）
			for _, oldSku := range oldSkus {
				if !newSkuCodes[oldSku.SKUCode] {
					if err := tx.Unscoped().Delete(&oldSku).Error; err != nil {
						return fmt.Errorf("删除旧SKU[%s]失败: %w", oldSku.SKUCode, err)
					}
				}
			}

			// 更新或创建SKU
			for i, skuReq := range req.SKUs {
				// 检查SKU编码是否为空
				if skuReq.SKUCode == "" {
					return fmt.Errorf("第%d个SKU编码不能为空", i+1)
				}

				// 查找是否已存在该SKU编码
				var existingSku models.ProductSKU
				err := tx.Where("sku_code = ?", skuReq.SKUCode).First(&existingSku).Error
				if err == nil {
					// 已存在，更新
					existingSku.ProductID = product.ID
					existingSku.SKUName = skuReq.SKUName
					existingSku.Attributes = datatypes.JSON(skuReq.Attributes)
					existingSku.Barcode = skuReq.Barcode
					existingSku.Status = 1
					if err := tx.Save(&existingSku).Error; err != nil {
						return fmt.Errorf("更新SKU[%s]失败: %w", skuReq.SKUCode, err)
					}
				} else if errors.Is(err, gorm.ErrRecordNotFound) {
					// 不存在，创建新SKU
					sku := &models.ProductSKU{
						ProductID:  product.ID,
						SKUCode:    skuReq.SKUCode,
						SKUName:    skuReq.SKUName,
						Attributes: datatypes.JSON(skuReq.Attributes),
						Barcode:    skuReq.Barcode,
						Status:     1,
					}
					if err := tx.Create(sku).Error; err != nil {
						return fmt.Errorf("创建SKU[%s]失败: %w", skuReq.SKUCode, err)
					}
				} else {
					return fmt.Errorf("查询SKU[%s]失败: %w", skuReq.SKUCode, err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return &AppError{Code: apperrors.InternalError, Message: fmt.Sprintf("更新商品失败: %v", err)}
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

// ImportResult 导入结果
type ImportResult struct {
	TotalCount   int           `json:"total_count"`
	SuccessCount int           `json:"success_count"`
	FailCount    int           `json:"fail_count"`
	Errors       []ImportError `json:"errors"`
}

// ImportError 导入错误详情
type ImportError struct {
	Row     int    `json:"row"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// BatchImport 批量导入商品（单Sheet：每行=一个商品，规格颜色自动生成SKU组合）
func (s *ProductService) BatchImport(storeID int64, createdBy int64, fileData []byte) (*ImportResult, error) {
	// 1. 解析Excel文件
	rows, err := excel.ParseImportFile(fileData)
	if err != nil {
		return nil, &AppError{Code: 400, Message: err.Error()}
	}

	if len(rows) == 0 {
		return nil, &AppError{Code: 400, Message: "Excel中没有数据"}
	}

	result := &ImportResult{
		TotalCount: len(rows),
	}

	// 2. 预处理：自动编码
	maxProductSeq, _ := s.getMaxProductCodeSeq(storeID)

	// 记录本次导入已使用的编码（避免同一文件内重复）
	usedProductCodes := make(map[string]bool)
	usedSKUCodes := make(map[string]bool)

	// 3. 逐行创建商品（一行一个商品，规格颜色生成SKU组合）
	for _, row := range rows {
		if row.ProductName == "" {
			result.Errors = append(result.Errors, ImportError{
				Row: row.Row, Code: row.ProductCode, Message: "商品名称不能为空",
			})
			result.FailCount++
			continue
		}

		// 商品编码处理
		code := row.ProductCode
		if code == "" {
			// 自动生成编码，确保唯一
			for {
				maxProductSeq++
				code = fmt.Sprintf("ZD%d", maxProductSeq)
				// 检查数据库中是否已存在
				existing, _ := s.productRepo.FindByCode(storeID, code)
				if existing == nil && !usedProductCodes[code] {
					break
				}
			}
		} else {
			// 手动填写编码，检查唯一性
			if usedProductCodes[code] {
				result.Errors = append(result.Errors, ImportError{
					Row: row.Row, Code: code, Message: "商品编码在本文件中重复",
				})
				result.FailCount++
				continue
			}
			existing, _ := s.productRepo.FindByCode(storeID, code)
			if existing != nil {
				result.Errors = append(result.Errors, ImportError{
					Row: row.Row, Code: code, Message: "商品编码已存在",
				})
				result.FailCount++
				continue
			}
		}
		usedProductCodes[code] = true

		// 查找品类ID（按品类名称查找）
		var categoryID *int64
		if row.CategoryName != "" {
			var category models.Category
			if err := s.db.Where("category_name = ?", row.CategoryName).First(&category).Error; err != nil {
				result.Errors = append(result.Errors, ImportError{
					Row: row.Row, Code: code,
					Message: fmt.Sprintf("品类名称[%s]不存在", row.CategoryName),
				})
				result.FailCount++
				continue
			}
			categoryID = &category.ID
		}

		// 解析规格和颜色，生成SKU组合
		attrCombos := excel.ParseSpecColor(row.Spec, row.Color)

		// 生成SKU列表
		var skus []SKURequest
		for i, attrs := range attrCombos {
			// SKU编码格式：商品编码-01、-02...
			skuSeq := i + 1
			skuCode := fmt.Sprintf("%s-%02d", code, skuSeq)

			// 检查SKU编码是否已存在（数据库或本次文件内）
			var existingSku models.ProductSKU
			err := s.db.Where("sku_code = ?", skuCode).First(&existingSku).Error
			if err == nil || usedSKUCodes[skuCode] {
				// 如果冲突，使用递增后缀
				for j := 1; j <= 1000; j++ {
					skuCode = fmt.Sprintf("%s-%s", code, fmt.Sprintf("%02d%02d", skuSeq, j))
					var exist models.ProductSKU
					if err := s.db.Where("sku_code = ?", skuCode).First(&exist).Error; err != nil {
						if !usedSKUCodes[skuCode] {
							break
						}
					}
				}
			}
			usedSKUCodes[skuCode] = true

			skuName := excel.SKUNameFromAttrs(row.ProductName, attrs)

			// 条码：第一个SKU使用输入的条码，其他留空
			barcode := ""
			if i == 0 {
				barcode = row.Barcode
			}

			skus = append(skus, SKURequest{
				SKUCode:    skuCode,
				SKUName:    skuName,
				Attributes: excel.AttributesToJSON(attrs),
				Barcode:    barcode,
			})
		}

		req := &CreateProductRequest{
			StoreID:       storeID,
			CategoryID:    categoryID,
			ProductCode:   code,
			ProductName:   row.ProductName,
			Brand:         row.Brand,
			Style:         row.Style,
			Unit:          row.Unit,
			Series:        row.Series,
			SubCategory:   row.SubCategory,
			ListPrice:     row.ListPrice,
			ReferenceCost: row.ReferenceCost,
			TotalCostRate: row.TotalCostRate,
			WarningStock:  row.WarningStock,
			SKUs:          skus,
		}

		if err := s.Create(req, createdBy); err != nil {
			result.Errors = append(result.Errors, ImportError{
				Row: row.Row, Code: code, Message: err.Error(),
			})
			result.FailCount++
			continue
		}

		result.SuccessCount++
	}

	return result, nil
}

// getMaxProductCodeSeq 获取当前最大商品编码序号
func (s *ProductService) getMaxProductCodeSeq(storeID int64) (int, error) {
	var maxCode string
	err := s.db.Model(&models.Product{}).
		Where("store_id = ? AND product_code LIKE 'ZD%'", storeID).
		Select("MAX(product_code)").
		Scan(&maxCode).Error
	if err != nil {
		return 100000, nil // 出错时从100000开始
	}

	if maxCode == "" {
		return 100000, nil
	}

	// 提取数字部分
	maxCode = strings.TrimPrefix(maxCode, "ZD")
	seq, err := strconv.Atoi(maxCode)
	if err != nil {
		return 100000, nil
	}
	return seq, nil
}

// getMaxSKUCodeSeq 获取当前最大SKU编码序号
func (s *ProductService) getMaxSKUCodeSeq() (int, error) {
	var maxCode string
	err := s.db.Model(&models.ProductSKU{}).
		Where("sku_code LIKE 'SKU%'").
		Select("MAX(sku_code)").
		Scan(&maxCode).Error
	if err != nil {
		return 100000, nil
	}

	if maxCode == "" {
		return 100000, nil
	}

	maxCode = strings.TrimPrefix(maxCode, "SKU")
	seq, err := strconv.Atoi(maxCode)
	if err != nil {
		return 100000, nil
	}
	return seq, nil
}
