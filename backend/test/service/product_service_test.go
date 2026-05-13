package service_test

import (
	"fmt"
	"testing"

	"furniture-commission/internal/models"
	"furniture-commission/internal/repository"
	svc "furniture-commission/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupProductTestDB 创建商品测试数据库
func setupProductTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(
		&models.Product{},
		&models.Category{},
		&models.ProductSKU{},
	)
	assert.NoError(t, err)
	return db
}

// ========== TestCreateProduct ==========

func TestCreateProduct(t *testing.T) {
	db := setupProductTestDB(t)

	productRepo := repository.NewProductRepository(db)
	productSvc := svc.NewProductService(db, productRepo)

	// 先创建品类
	cat := &models.Category{
		StoreID:      1,
		CategoryCode: "SOFA",
		CategoryName: "沙发",
		Status:       1,
	}
	db.Create(cat)
	catID := cat.ID

	req := &svc.CreateProductRequest{
		StoreID:       1,
		CategoryID:    &catID,
		ProductCode:   "SF-001",
		ProductName:   "真皮沙发",
		Brand:         "品牌A",
		ListPrice:     10000,
		MinPrice:      8000,
		ReferenceCost: 5000,
		WarningStock:  5,
	}

	err := productSvc.Create(req, 1)
	assert.NoError(t, err)

	// 验证商品已创建
	product, err := productRepo.FindByCode(1, "SF-001")
	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, "真皮沙发", product.ProductName)
	assert.Equal(t, int8(1), product.Status)
}

func TestCreateProduct_DuplicateCode(t *testing.T) {
	db := setupProductTestDB(t)

	productRepo := repository.NewProductRepository(db)
	productSvc := svc.NewProductService(db, productRepo)

	req1 := &svc.CreateProductRequest{
		StoreID:     1,
		ProductCode: "PD-001",
		ProductName: "商品1",
		ListPrice:   1000,
		MinPrice:    800,
	}
	err := productSvc.Create(req1, 1)
	assert.NoError(t, err)

	req2 := &svc.CreateProductRequest{
		StoreID:     1,
		ProductCode: "PD-001",
		ProductName: "商品2",
		ListPrice:   2000,
		MinPrice:    1500,
	}
	err = productSvc.Create(req2, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "商品编码已存在")
}

// ========== TestUpdateProductStatus ==========

func TestUpdateProductStatus(t *testing.T) {
	db := setupProductTestDB(t)

	productRepo := repository.NewProductRepository(db)
	productSvc := svc.NewProductService(db, productRepo)

	// 创建商品
	req := &svc.CreateProductRequest{
		StoreID:     1,
		ProductCode: "SF-002",
		ProductName: "布艺沙发",
		ListPrice:   5000,
		MinPrice:    3000,
	}
	err := productSvc.Create(req, 1)
	assert.NoError(t, err)

	product, _ := productRepo.FindByCode(1, "SF-002")

	// 下架
	err = productSvc.UpdateStatus(product.ID, 0)
	assert.NoError(t, err)

	updated, _ := productRepo.FindByID(product.ID)
	assert.Equal(t, int8(0), updated.Status)

	// 重新上架
	err = productSvc.UpdateStatus(product.ID, 1)
	assert.NoError(t, err)

	updated, _ = productRepo.FindByID(product.ID)
	assert.Equal(t, int8(1), updated.Status)
}

func TestUpdateProductStatus_NotFound(t *testing.T) {
	db := setupProductTestDB(t)

	productRepo := repository.NewProductRepository(db)
	productSvc := svc.NewProductService(db, productRepo)

	err := productSvc.UpdateStatus(99999, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "商品不存在")
}

// ========== TestListProducts ==========

type ProductListTestSuite struct {
	suite.Suite
	db          *gorm.DB
	productSvc  *svc.ProductService
	productRepo *repository.ProductRepository
}

func (s *ProductListTestSuite) SetupSuite() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	s.Require().NoError(err)
	s.db = db

	err = db.AutoMigrate(
		&models.Product{},
		&models.Category{},
		&models.ProductSKU{},
	)
	s.Require().NoError(err)

	s.productRepo = repository.NewProductRepository(db)
	s.productSvc = svc.NewProductService(db, s.productRepo)
}

func (s *ProductListTestSuite) TearDownSuite() {
	sqlDB, _ := s.db.DB()
	sqlDB.Close()
}

func (s *ProductListTestSuite) SetupTest() {
	s.db.Exec("DELETE FROM products")
}

func (s *ProductListTestSuite) TestListProducts() {
	// 创建多个商品
	for i := 1; i <= 3; i++ {
		req := &svc.CreateProductRequest{
			StoreID:     1,
			ProductCode: fmt.Sprintf("PD-%03d", i),
			ProductName: fmt.Sprintf("商品%d", i),
			ListPrice:   float64(1000 * i),
			MinPrice:    float64(800 * i),
		}
		err := s.productSvc.Create(req, 1)
		s.Require().NoError(err)
	}

	listReq := &svc.ListProductRequest{
		StoreID:  1,
		Page:     1,
		PageSize: 10,
	}
	result, err := s.productSvc.List(listReq)
	s.NoError(err)
	s.NotNil(result)
	s.Equal(int64(3), result.Total)
	s.Len(result.List, 3)
}

func (s *ProductListTestSuite) TestListProducts_WithStatus() {
	// 创建上架和下架商品
	req1 := &svc.CreateProductRequest{
		StoreID:     1,
		ProductCode: "PD-ON",
		ProductName: "上架商品",
		ListPrice:   1000,
		MinPrice:    800,
	}
	err := s.productSvc.Create(req1, 1)
	s.Require().NoError(err)

	req2 := &svc.CreateProductRequest{
		StoreID:     1,
		ProductCode: "PD-OFF",
		ProductName: "下架商品",
		ListPrice:   2000,
		MinPrice:    1500,
	}
	err = s.productSvc.Create(req2, 1)
	s.Require().NoError(err)

	// 下架第二个商品
	products, _ := s.productRepo.FindByCode(1, "PD-OFF")
	s.productSvc.UpdateStatus(products.ID, 0)

	// 只查询上架商品
	onStatus := int8(1)
	listReq := &svc.ListProductRequest{
		StoreID:  1,
		Status:   &onStatus,
		Page:     1,
		PageSize: 10,
	}
	result, err := s.productSvc.List(listReq)
	s.NoError(err)
	s.Equal(int64(1), result.Total)
}

func TestProductListSuite(t *testing.T) {
	suite.Run(t, new(ProductListTestSuite))
}
