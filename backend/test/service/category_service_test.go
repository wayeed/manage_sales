package service_test

import (
	"testing"

	"furniture-commission/internal/models"
	"furniture-commission/internal/repository"
	svc "furniture-commission/internal/service"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupCategoryTestDB 创建品类测试数据库
func setupCategoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&models.Category{})
	assert.NoError(t, err)
	return db
}

// ========== TestCreateCategory ==========

func TestCreateCategory(t *testing.T) {
	db := setupCategoryTestDB(t)

	categoryRepo := repository.NewCategoryRepository(db)
	categorySvc := svc.NewCategoryService(categoryRepo)

	req := &svc.CreateCategoryRequest{
		StoreID:      1,
		CategoryCode: "SOFA",
		CategoryName: "沙发",
		SortOrder:    1,
	}

	err := categorySvc.Create(req)
	assert.NoError(t, err)

	// 验证品类已创建
	category, err := categoryRepo.FindByCode(1, "SOFA")
	assert.NoError(t, err)
	assert.NotNil(t, category)
	assert.Equal(t, "沙发", category.CategoryName)
	assert.Equal(t, int8(1), category.Status)
}

func TestCreateCategory_DuplicateCode(t *testing.T) {
	db := setupCategoryTestDB(t)

	categoryRepo := repository.NewCategoryRepository(db)
	categorySvc := svc.NewCategoryService(categoryRepo)

	req1 := &svc.CreateCategoryRequest{
		StoreID:      1,
		CategoryCode: "BED",
		CategoryName: "床",
	}
	err := categorySvc.Create(req1)
	assert.NoError(t, err)

	req2 := &svc.CreateCategoryRequest{
		StoreID:      1,
		CategoryCode: "BED",
		CategoryName: "床2",
	}
	err = categorySvc.Create(req2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "品类编码已存在")
}

// ========== TestListCategories ==========

func TestListCategories(t *testing.T) {
	db := setupCategoryTestDB(t)

	categoryRepo := repository.NewCategoryRepository(db)
	categorySvc := svc.NewCategoryService(categoryRepo)

	// 创建多个品类
	categories := []svc.CreateCategoryRequest{
		{StoreID: 1, CategoryCode: "SOFA", CategoryName: "沙发", SortOrder: 1},
		{StoreID: 1, CategoryCode: "BED", CategoryName: "床", SortOrder: 2},
		{StoreID: 1, CategoryCode: "TABLE", CategoryName: "餐桌", SortOrder: 3},
	}
	for _, c := range categories {
		err := categorySvc.Create(&c)
		assert.NoError(t, err)
	}

	// 查询品类列表
	result, err := categorySvc.List(1)
	assert.NoError(t, err)
	assert.Len(t, result, 3)

	// 验证返回数据
	names := make(map[string]bool)
	for _, c := range result {
		names[c.CategoryName] = true
	}
	assert.True(t, names["沙发"])
	assert.True(t, names["床"])
	assert.True(t, names["餐桌"])
}

func TestListCategories_Empty(t *testing.T) {
	db := setupCategoryTestDB(t)

	categoryRepo := repository.NewCategoryRepository(db)
	categorySvc := svc.NewCategoryService(categoryRepo)

	result, err := categorySvc.List(1)
	assert.NoError(t, err)
	assert.Empty(t, result)
}
