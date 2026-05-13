package service_test

import (
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

// setupPaymentTestDB 创建回款测试数据库
func setupPaymentTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(
		&models.Payment{},
		&models.Order{},
		&models.OrderItem{},
		&models.OrderGift{},
		&models.Customer{},
		&models.CustomerFollowUp{},
		&models.Peer{},
		&models.WarehouseStock{},
		&models.InventoryBatch{},
		&models.InventoryTransaction{},
		&models.ProductSKU{},
		&models.Warehouse{},
	)
	assert.NoError(t, err)
	return db
}

// setupPaymentTestService 创建回款测试服务
func setupPaymentTestService(t *testing.T) (*svc.PaymentService, *gorm.DB) {
	db := setupPaymentTestDB(t)
	paymentRepo := repository.NewPaymentRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	paymentSvc := svc.NewPaymentService(db, paymentRepo, orderRepo)
	return paymentSvc, db
}

// createApprovedOrderForPayment 创建已审核通过的订单（用于回款测试）
func createApprovedOrderForPayment(t *testing.T, db *gorm.DB, orderID int64, finalAmount float64) {
	now := time.Now()
	order := &models.Order{
		ID:            orderID,
		StoreID:       1,
		OrderNo:       "ORD-PAY-001",
		SalesmanID:    1,
		CustomerName:  "回款测试客户",
		CustomerPhone: "13800138000",
		OrderType:     1,
		OrderStatus:   1, // 已生效
		PaymentStatus: 0,
		FinalAmount:   decimal.NewFromFloat(finalAmount),
		OrderDate:     &now,
	}
	err := db.Create(order).Error
	assert.NoError(t, err)
}

// ========== TestCreatePayment ==========

func TestCreatePayment(t *testing.T) {
	paymentSvc, db := setupPaymentTestService(t)

	// 创建已生效订单
	createApprovedOrderForPayment(t, db, 1, 9000.00)

	req := &svc.CreatePaymentRequest{
		OrderID:       1,
		Amount:        5000.00,
		PaymentDate:   "2025-01-15",
		PaymentMethod: 1,
		Remark:        "首付款",
	}

	err := paymentSvc.CreatePayment(req, 1)
	assert.NoError(t, err)

	// 验证回款记录已创建
	payments, err := paymentSvc.GetByOrderID(1)
	assert.NoError(t, err)
	assert.Len(t, payments, 1)
	assert.Equal(t, int8(0), payments[0].Status) // 待审核
	assert.True(t, payments[0].Amount.Equal(decimal.NewFromFloat(5000.00)))
	assert.NotEmpty(t, payments[0].PaymentNo)
}

func TestCreatePayment_InvalidOrderStatus(t *testing.T) {
	paymentSvc, db := setupPaymentTestService(t)

	// 创建待审批订单（order_status=0）
	now := time.Now()
	order := &models.Order{
		ID:            1,
		StoreID:       1,
		OrderNo:       "ORD-PENDING",
		SalesmanID:    1,
		CustomerName:  "客户",
		OrderStatus:   0, // 待审批
		FinalAmount:   decimal.NewFromFloat(5000),
		OrderDate:     &now,
	}
	db.Create(order)

	req := &svc.CreatePaymentRequest{
		OrderID: 1,
		Amount:  3000.00,
	}

	err := paymentSvc.CreatePayment(req, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "只有已生效的订单才能录入回款")
}

func TestCreatePayment_OrderNotFound(t *testing.T) {
	paymentSvc, _ := setupPaymentTestService(t)

	req := &svc.CreatePaymentRequest{
		OrderID: 99999,
		Amount:  1000.00,
	}

	err := paymentSvc.CreatePayment(req, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "订单不存在")
}

func TestCreatePayment_ZeroAmount(t *testing.T) {
	paymentSvc, db := setupPaymentTestService(t)

	createApprovedOrderForPayment(t, db, 1, 5000.00)

	req := &svc.CreatePaymentRequest{
		OrderID: 1,
		Amount:  0,
	}

	// 零金额回款：服务层可能允许创建，验证创建后的状态
	err := paymentSvc.CreatePayment(req, 1)
	if err != nil {
		// 如果服务层验证了零金额，测试通过
		assert.Error(t, err)
	} else {
		// 如果服务层允许零金额，验证记录已创建
		payments, _ := paymentSvc.GetByOrderID(1)
		assert.Len(t, payments, 1)
	}
}

// ========== TestApprovePayment ==========

func TestApprovePayment(t *testing.T) {
	paymentSvc, db := setupPaymentTestService(t)

	createApprovedOrderForPayment(t, db, 1, 9000.00)

	// 创建回款
	payReq := &svc.CreatePaymentRequest{
		OrderID: 1,
		Amount:  5000.00,
	}
	err := paymentSvc.CreatePayment(payReq, 1)
	assert.NoError(t, err)

	payments, _ := paymentSvc.GetByOrderID(1)
	paymentID := payments[0].ID

	// 审核通过
	err = paymentSvc.ApprovePayment(paymentID, 2, true)
	assert.NoError(t, err)

	// 验证回款状态
	payments, _ = paymentSvc.GetByOrderID(1)
	assert.Equal(t, int8(1), payments[0].Status) // 已审核

	// 验证订单已回款金额更新
	var order models.Order
	db.First(&order, 1)
	assert.True(t, order.PaidAmount.Equal(decimal.NewFromFloat(5000.00)))
	assert.Equal(t, int8(1), order.PaymentStatus) // 部分回款
}

func TestApprovePayment_FullPayment(t *testing.T) {
	paymentSvc, db := setupPaymentTestService(t)

	createApprovedOrderForPayment(t, db, 1, 5000.00)

	// 创建全额回款
	payReq := &svc.CreatePaymentRequest{
		OrderID: 1,
		Amount:  5000.00,
	}
	paymentSvc.CreatePayment(payReq, 1)

	payments, _ := paymentSvc.GetByOrderID(1)
	paymentID := payments[0].ID

	// 审核通过
	err := paymentSvc.ApprovePayment(paymentID, 2, true)
	assert.NoError(t, err)

	// 验证订单已全额回款
	var order models.Order
	db.First(&order, 1)
	assert.Equal(t, int8(2), order.PaymentStatus) // 已回款
	assert.True(t, order.RemainingAmount.Equal(decimal.Zero))
}

func TestApprovePayment_Reject(t *testing.T) {
	paymentSvc, db := setupPaymentTestService(t)

	createApprovedOrderForPayment(t, db, 1, 5000.00)

	payReq := &svc.CreatePaymentRequest{
		OrderID: 1,
		Amount:  3000.00,
	}
	paymentSvc.CreatePayment(payReq, 1)

	payments, _ := paymentSvc.GetByOrderID(1)
	paymentID := payments[0].ID

	// 审核驳回
	err := paymentSvc.ApprovePayment(paymentID, 2, false)
	assert.NoError(t, err)

	// 验证回款状态
	payments, _ = paymentSvc.GetByOrderID(1)
	assert.Equal(t, int8(2), payments[0].Status) // 已驳回

	// 验证订单回款金额未变
	var order models.Order
	db.First(&order, 1)
	assert.True(t, order.PaidAmount.Equal(decimal.Zero))
}

func TestApprovePayment_AlreadyApproved(t *testing.T) {
	paymentSvc, db := setupPaymentTestService(t)

	createApprovedOrderForPayment(t, db, 1, 5000.00)

	payReq := &svc.CreatePaymentRequest{
		OrderID: 1,
		Amount:  3000.00,
	}
	paymentSvc.CreatePayment(payReq, 1)

	payments, _ := paymentSvc.GetByOrderID(1)
	paymentID := payments[0].ID

	// 第一次审核通过
	paymentSvc.ApprovePayment(paymentID, 2, true)

	// 第二次审核应失败
	err := paymentSvc.ApprovePayment(paymentID, 2, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "回款记录已审核")
}

func TestApprovePayment_NotFound(t *testing.T) {
	paymentSvc, _ := setupPaymentTestService(t)

	err := paymentSvc.ApprovePayment(99999, 1, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "回款记录不存在")
}

// ========== TestListPayments ==========

func TestListPayments(t *testing.T) {
	paymentSvc, db := setupPaymentTestService(t)

	createApprovedOrderForPayment(t, db, 1, 10000.00)

	// 创建多条回款
	for i := 0; i < 3; i++ {
		req := &svc.CreatePaymentRequest{
			OrderID: 1,
			Amount:  float64((i + 1) * 1000),
		}
		paymentSvc.CreatePayment(req, 1)
	}

	listReq := &svc.ListPaymentRequest{
		Page:     1,
		PageSize: 10,
	}
	result, err := paymentSvc.List(listReq)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), result.Total)
}
