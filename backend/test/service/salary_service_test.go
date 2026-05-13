package service_test

import (
	"fmt"
	"testing"
	"time"

	"furniture-commission/internal/models"
	"furniture-commission/internal/repository"
	svc "furniture-commission/internal/service"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupSalaryTestDB 创建工资测试数据库
func setupSalaryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(
		&models.User{},
		&models.Store{},
		&models.SalaryRecord{},
		&models.SalaryItem{},
		&models.Commission{},
		&models.Order{},
		&models.FundPool{},
		&models.FundPoolShare{},
	)
	assert.NoError(t, err)
	return db
}

// setupSalaryTestService 创建工资测试服务
func setupSalaryTestService(t *testing.T) (*svc.SalaryService, *gorm.DB) {
	db := setupSalaryTestDB(t)
	salaryRepo := repository.NewSalaryRecordRepository(db)
	commissionRepo := repository.NewCommissionRepository(db)
	fundPoolRepo := repository.NewFundPoolRepository(db)
	salarySvc := svc.NewSalaryService(db, salaryRepo, commissionRepo, fundPoolRepo)
	return salarySvc, db
}

// createEmployeeForSalary 创建测试员工
func createEmployeeForSalary(t *testing.T, db *gorm.DB, id int64, storeID int64, baseSalary float64) {
	user := &models.User{
		ID:          id,
		StoreID:     &storeID,
		EmployeeNo:  fmt.Sprintf("EMP-%d", id),
		Username:    fmt.Sprintf("salaryuser%d", id),
		Password:    "hashed",
		RealName:    fmt.Sprintf("员工%d", id),
		Phone:       fmt.Sprintf("138%08d", id),
		Status:      1,
		BaseSalary:  baseSalary,
	}
	err := db.Create(user).Error
	assert.NoError(t, err)
}

// ========== TestGenerateSalary ==========

func TestGenerateSalary(t *testing.T) {
	salarySvc, db := setupSalaryTestService(t)

	// 创建门店
	store := &models.Store{
		ID:        1,
		StoreCode: "STORE-001",
		StoreName: "测试门店",
		Status:    1,
	}
	db.Create(store)

	// 创建员工
	createEmployeeForSalary(t, db, 10, 1, 5000)
	createEmployeeForSalary(t, db, 11, 1, 6000)

	// 生成工资
	err := salarySvc.GenerateSalary(1, "2025-01")
	assert.NoError(t, err)

	// 验证工资记录已生成
	listReq := &svc.ListSalaryRequest{
		StoreID:     "1",
		SalaryMonth: "2025-01",
		Page:        1,
		PageSize:    10,
	}
	result, err := salarySvc.List(listReq)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)

	// 验证工资明细
	records, ok := result.List.([]models.SalaryRecord)
	assert.True(t, ok)
	for _, sr := range records {
		assert.Equal(t, int8(0), sr.Status) // 草稿
		assert.True(t, sr.BaseSalary.GreaterThan(decimal.Zero))

		// 验证明细项
		detail, err := salarySvc.GetDetail(sr.ID)
		assert.NoError(t, err)
		assert.Len(t, detail.Items, 1) // 至少有基本工资
		assert.Equal(t, "基本工资", detail.Items[0].ItemName)
	}
}

func TestGenerateSalary_NoEmployees(t *testing.T) {
	salarySvc, db := setupSalaryTestService(t)

	// 创建门店但无员工
	store := &models.Store{
		ID:        1,
		StoreCode: "STORE-001",
		StoreName: "空门店",
		Status:    1,
	}
	db.Create(store)

	err := salarySvc.GenerateSalary(1, "2025-01")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "该门店无在职员工")
}

func TestGenerateSalary_Idempotent(t *testing.T) {
	salarySvc, db := setupSalaryTestService(t)

	store := &models.Store{
		ID:        1,
		StoreCode: "STORE-001",
		StoreName: "测试门店",
		Status:    1,
	}
	db.Create(store)

	createEmployeeForSalary(t, db, 10, 1, 5000)

	// 第一次生成
	err := salarySvc.GenerateSalary(1, "2025-01")
	assert.NoError(t, err)

	// 第二次生成（应跳过已存在的记录）
	err = salarySvc.GenerateSalary(1, "2025-01")
	assert.NoError(t, err)

	// 验证只有一条记录
	listReq := &svc.ListSalaryRequest{
		StoreID:     "1",
		SalaryMonth: "2025-01",
		Page:        1,
		PageSize:    10,
	}
	result, _ := salarySvc.List(listReq)
	assert.Equal(t, int64(1), result.Total)
}

// ========== TestConfirmSalary ==========

func TestConfirmSalary(t *testing.T) {
	salarySvc, db := setupSalaryTestService(t)

	store := &models.Store{
		ID:        1,
		StoreCode: "STORE-001",
		StoreName: "测试门店",
		Status:    1,
	}
	db.Create(store)

	createEmployeeForSalary(t, db, 10, 1, 5000)

	// 生成工资
	err := salarySvc.GenerateSalary(1, "2025-01")
	assert.NoError(t, err)

	// 获取工资记录ID
	listReq := &svc.ListSalaryRequest{
		StoreID:     "1",
		SalaryMonth: "2025-01",
		Page:        1,
		PageSize:    10,
	}
	result, _ := salarySvc.List(listReq)
	records := result.List.([]models.SalaryRecord)
	sr := records[0]

	// 确认工资
	err = salarySvc.ConfirmSalary(sr.ID, 2)
	assert.NoError(t, err)

	// 验证状态已更新
	detail, err := salarySvc.GetDetail(sr.ID)
	assert.NoError(t, err)
	assert.Equal(t, int8(1), detail.Status) // 已确认
	assert.NotNil(t, detail.ConfirmedBy)
	assert.Equal(t, int64(2), *detail.ConfirmedBy)
	assert.NotNil(t, detail.ConfirmedAt)
}

func TestConfirmSalary_NotFound(t *testing.T) {
	salarySvc, _ := setupSalaryTestService(t)

	err := salarySvc.ConfirmSalary(99999, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "工资记录不存在")
}

func TestConfirmSalary_InvalidStatus(t *testing.T) {
	salarySvc, db := setupSalaryTestService(t)

	store := &models.Store{
		ID:        1,
		StoreCode: "STORE-001",
		StoreName: "测试门店",
		Status:    1,
	}
	db.Create(store)

	createEmployeeForSalary(t, db, 10, 1, 5000)
	salarySvc.GenerateSalary(1, "2025-01")

	listReq := &svc.ListSalaryRequest{
		StoreID:     "1",
		SalaryMonth: "2025-01",
		Page:        1,
		PageSize:    10,
	}
	result, _ := salarySvc.List(listReq)
	records := result.List.([]models.SalaryRecord)
	sr := records[0]

	// 确认工资
	salarySvc.ConfirmSalary(sr.ID, 2)

	// 再次确认应失败
	err := salarySvc.ConfirmSalary(sr.ID, 2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "只有草稿状态的工资才能确认")
}

// ========== TestPaySalary ==========

func TestPaySalary(t *testing.T) {
	salarySvc, db := setupSalaryTestService(t)

	store := &models.Store{
		ID:        1,
		StoreCode: "STORE-001",
		StoreName: "测试门店",
		Status:    1,
	}
	db.Create(store)

	createEmployeeForSalary(t, db, 10, 1, 5000)
	salarySvc.GenerateSalary(1, "2025-01")

	listReq := &svc.ListSalaryRequest{
		StoreID:     "1",
		SalaryMonth: "2025-01",
		Page:        1,
		PageSize:    10,
	}
	result, _ := salarySvc.List(listReq)
	records := result.List.([]models.SalaryRecord)
	sr := records[0]

	// 确认
	salarySvc.ConfirmSalary(sr.ID, 2)

	// 发放
	err := salarySvc.PaySalary(sr.ID, 3, 1, "银行转账")
	assert.NoError(t, err)

	// 验证状态
	detail, err := salarySvc.GetDetail(sr.ID)
	assert.NoError(t, err)
	assert.Equal(t, int8(2), detail.Status) // 已发放
	assert.NotNil(t, detail.PaidAt)
	assert.Equal(t, int64(3), *detail.PaidBy)
	assert.Equal(t, int8(1), detail.PayMethod)
}

func TestPaySalary_InvalidStatus(t *testing.T) {
	salarySvc, db := setupSalaryTestService(t)

	store := &models.Store{
		ID:        1,
		StoreCode: "STORE-001",
		StoreName: "测试门店",
		Status:    1,
	}
	db.Create(store)

	createEmployeeForSalary(t, db, 10, 1, 5000)
	salarySvc.GenerateSalary(1, "2025-01")

	listReq := &svc.ListSalaryRequest{
		StoreID:     "1",
		SalaryMonth: "2025-01",
		Page:        1,
		PageSize:    10,
	}
	result, _ := salarySvc.List(listReq)
	records := result.List.([]models.SalaryRecord)
	sr := records[0]

	// 草稿状态直接发放应失败
	err := salarySvc.PaySalary(sr.ID, 3, 1, "银行转账")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "只有已确认的工资才能发放")
}

// ========== TestGetEmployeeSalary ==========

func TestGetEmployeeSalary(t *testing.T) {
	salarySvc, db := setupSalaryTestService(t)

	store := &models.Store{
		ID:        1,
		StoreCode: "STORE-001",
		StoreName: "测试门店",
		Status:    1,
	}
	db.Create(store)

	createEmployeeForSalary(t, db, 10, 1, 5000)
	salarySvc.GenerateSalary(1, "2025-01")

	record, err := salarySvc.GetEmployeeSalary(10, "2025-01")
	assert.NoError(t, err)
	assert.NotNil(t, record)
	assert.Equal(t, int64(10), record.EmployeeID)
	assert.Equal(t, "2025-01", record.SalaryMonth)
	assert.True(t, record.BaseSalary.Equal(decimal.NewFromFloat(5000)))
}

func TestGetEmployeeSalary_NotFound(t *testing.T) {
	salarySvc, _ := setupSalaryTestService(t)

	_, err := salarySvc.GetEmployeeSalary(99999, "2025-01")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "工资记录不存在")
}

// ========== TestSalaryList ==========

func TestSalaryList(t *testing.T) {
	salarySvc, db := setupSalaryTestService(t)

	store := &models.Store{
		ID:        1,
		StoreCode: "STORE-001",
		StoreName: "测试门店",
		Status:    1,
	}
	db.Create(store)

	createEmployeeForSalary(t, db, 10, 1, 5000)
	createEmployeeForSalary(t, db, 11, 1, 6000)

	salarySvc.GenerateSalary(1, "2025-01")
	salarySvc.GenerateSalary(1, "2025-02")

	listReq := &svc.ListSalaryRequest{
		StoreID:  "1",
		Page:     1,
		PageSize: 10,
	}
	result, err := salarySvc.List(listReq)
	assert.NoError(t, err)
	assert.Equal(t, int64(4), result.Total) // 2个月 * 2个员工
}

// 辅助：确保 time 被使用
var _ = time.Now
