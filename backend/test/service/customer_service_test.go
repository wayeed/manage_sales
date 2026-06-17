package service_test

import (
	"fmt"
	"testing"

	"furniture-commission/internal/models"
	"furniture-commission/internal/repository"
	svc "furniture-commission/internal/service"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupCustomerTestDB 创建客户测试数据库
func setupCustomerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(
		&models.Customer{},
		&models.CustomerFollowUp{},
		&models.User{},
	)
	assert.NoError(t, err)
	return db
}

// ========== TestCreateCustomer ==========

func TestCreateCustomer(t *testing.T) {
	db := setupCustomerTestDB(t)

	custRepo := repository.NewCustomerRepository(db)
	customerSvc := svc.NewCustomerService(db, custRepo)

	req := &svc.CreateCustomerRequest{
		StoreID:      1,
		CustomerName: "张三",
		Phone:        "13800138000",
		Email:        "zhangsan@example.com",
		Address:      "北京市朝阳区",
		Gender:       1,
		Level:        2,
		Remark:       "VIP客户",
	}

	customer, err := customerSvc.Create(req, 1)
	assert.NoError(t, err)
	assert.NotNil(t, customer)
	assert.NotZero(t, customer.ID)
	assert.Equal(t, "张三", customer.CustomerName)
	assert.Equal(t, "13800138000", customer.Phone)
	assert.Equal(t, "zhangsan@example.com", customer.Email)
	assert.Equal(t, int8(1), customer.Status)
	assert.Equal(t, int8(2), customer.Level)

	// 验证可以从数据库查询到
	found, err := custRepo.FindByID(customer.ID)
	assert.NoError(t, err)
	assert.Equal(t, "张三", found.CustomerName)
}

func TestCreateCustomer_WithBirthday(t *testing.T) {
	db := setupCustomerTestDB(t)

	custRepo := repository.NewCustomerRepository(db)
	customerSvc := svc.NewCustomerService(db, custRepo)

	req := &svc.CreateCustomerRequest{
		StoreID:      1,
		CustomerName: "李四",
		Phone:        "13800138001",
		Birthday:     "1990-05-15",
	}

	customer, err := customerSvc.Create(req, 1)
	assert.NoError(t, err)
	assert.NotNil(t, customer.Birthday)
}

func TestCreateCustomer_InvalidBirthday(t *testing.T) {
	db := setupCustomerTestDB(t)

	custRepo := repository.NewCustomerRepository(db)
	customerSvc := svc.NewCustomerService(db, custRepo)

	req := &svc.CreateCustomerRequest{
		StoreID:      1,
		CustomerName: "王五",
		Phone:        "13800138002",
		Birthday:     "invalid-date",
	}

	customer, err := customerSvc.Create(req, 1)
	assert.NoError(t, err)
	assert.Nil(t, customer.Birthday) // 无效日期不设置
}

// ========== TestAddFollowUp ==========

func TestAddFollowUp(t *testing.T) {
	db := setupCustomerTestDB(t)

	custRepo := repository.NewCustomerRepository(db)
	customerSvc := svc.NewCustomerService(db, custRepo)

	// 先创建客户
	customerReq := &svc.CreateCustomerRequest{
		StoreID:      1,
		CustomerName: "跟进客户",
		Phone:        "13800138010",
	}
	customer, err := customerSvc.Create(customerReq, 1)
	assert.NoError(t, err)

	// 添加跟进记录
	followUpReq := &svc.AddFollowUpRequest{
		CustomerID:        customer.ID,
		FollowType:        1,
		Content:           "电话沟通，客户对产品感兴趣",
		NextFollowDate:    "2025-02-01",
		NextFollowContent: "预约到店体验",
		IsDeal:            0,
	}

	err = customerSvc.AddFollowUp(followUpReq, 1)
	assert.NoError(t, err)

	// 验证跟进记录已创建
	followUps, err := customerSvc.GetFollowUps(customer.ID)
	assert.NoError(t, err)
	assert.Len(t, followUps, 1)
	assert.Equal(t, "电话沟通，客户对产品感兴趣", followUps[0].Content)
	assert.Equal(t, int64(1), followUps[0].FollowerID)
	assert.Equal(t, int8(1), followUps[0].FollowType)
}

func TestAddFollowUp_MultipleRecords(t *testing.T) {
	db := setupCustomerTestDB(t)

	custRepo := repository.NewCustomerRepository(db)
	customerSvc := svc.NewCustomerService(db, custRepo)

	// 创建客户
	customerReq := &svc.CreateCustomerRequest{
		StoreID:      1,
		CustomerName: "多次跟进客户",
		Phone:        "13800138011",
	}
	customer, err := customerSvc.Create(customerReq, 1)
	assert.NoError(t, err)

	// 添加多条跟进记录
	for i := 0; i < 3; i++ {
		req := &svc.AddFollowUpRequest{
			CustomerID: customer.ID,
			FollowType: int8(i + 1),
			Content:    "第%d次跟进",
		}
		err = customerSvc.AddFollowUp(req, 1)
		assert.NoError(t, err)
	}

	followUps, err := customerSvc.GetFollowUps(customer.ID)
	assert.NoError(t, err)
	assert.Len(t, followUps, 3)
}

func TestAddFollowUp_Deal(t *testing.T) {
	db := setupCustomerTestDB(t)

	custRepo := repository.NewCustomerRepository(db)
	customerSvc := svc.NewCustomerService(db, custRepo)

	customerReq := &svc.CreateCustomerRequest{
		StoreID:      1,
		CustomerName: "成交客户",
		Phone:        "13800138012",
	}
	customer, err := customerSvc.Create(customerReq, 1)
	assert.NoError(t, err)

	req := &svc.AddFollowUpRequest{
		CustomerID: customer.ID,
		FollowType: 3,
		Content:    "客户已下单购买",
		IsDeal:     1,
	}

	err = customerSvc.AddFollowUp(req, 1)
	assert.NoError(t, err)

	followUps, _ := customerSvc.GetFollowUps(customer.ID)
	assert.Len(t, followUps, 1)
	assert.Equal(t, int8(1), followUps[0].IsDeal)
}

// ========== TestListCustomers ==========

func TestListCustomers(t *testing.T) {
	db := setupCustomerTestDB(t)

	custRepo := repository.NewCustomerRepository(db)
	customerSvc := svc.NewCustomerService(db, custRepo)

	// 创建多个客户（直接使用DB创建以设置唯一的CustomerCode）
	for i := 0; i < 5; i++ {
		customer := &models.Customer{
			StoreID:      1,
			CustomerCode: fmt.Sprintf("CUST-%03d", i),
			CustomerName: fmt.Sprintf("列表客户%d", i),
			Phone:        fmt.Sprintf("1390013802%d", i),
			Status:       1,
		}
		err := db.Create(customer).Error
		assert.NoError(t, err)
	}

	listReq := &svc.ListCustomerRequest{
		Page:     1,
		PageSize: 10,
	}
	result, err := customerSvc.List(listReq)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), result.Total)
}

// ========== TestUpdateCustomer ==========

func TestUpdateCustomer(t *testing.T) {
	db := setupCustomerTestDB(t)

	custRepo := repository.NewCustomerRepository(db)
	customerSvc := svc.NewCustomerService(db, custRepo)

	createReq := &svc.CreateCustomerRequest{
		StoreID:      1,
		CustomerName: "原名",
		Phone:        "13800138030",
	}
	customer, err := customerSvc.Create(createReq, 1)
	assert.NoError(t, err)

	level := int8(3)
	updateReq := &svc.UpdateCustomerRequest{
		CustomerName: "新名",
		Level:        &level,
		Remark:       "已更新",
	}

	err = customerSvc.Update(customer.ID, updateReq)
	assert.NoError(t, err)

	updated, _ := custRepo.FindByID(customer.ID)
	assert.Equal(t, "新名", updated.CustomerName)
	assert.Equal(t, int8(3), updated.Level)
	assert.Equal(t, "已更新", updated.Remark)
}

// ========== TestGetCustomerDetail ==========

func TestGetCustomerDetail(t *testing.T) {
	db := setupCustomerTestDB(t)

	custRepo := repository.NewCustomerRepository(db)
	customerSvc := svc.NewCustomerService(db, custRepo)

	createReq := &svc.CreateCustomerRequest{
		StoreID:      1,
		CustomerName: "详情客户",
		Phone:        "13800138040",
		Address:      "上海市浦东新区",
	}
	customer, err := customerSvc.Create(createReq, 1)
	assert.NoError(t, err)

	detail, err := customerSvc.GetDetail(customer.ID)
	assert.NoError(t, err)
	assert.Equal(t, "详情客户", detail.CustomerName)
	assert.Equal(t, "上海市浦东新区", detail.Address)
}

func TestGetCustomerDetail_NotFound(t *testing.T) {
	db := setupCustomerTestDB(t)

	custRepo := repository.NewCustomerRepository(db)
	customerSvc := svc.NewCustomerService(db, custRepo)

	_, err := customerSvc.GetDetail(99999)
	assert.Error(t, err)
}
