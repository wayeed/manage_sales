package integration_test

import (
	"fmt"
	"testing"
	"time"

	"furniture-commission/configs"
	"furniture-commission/internal/models"
	"furniture-commission/internal/repository"
	svc "furniture-commission/internal/service"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// OrderFlowTestSuite 端到端订单流程测试套件
type OrderFlowTestSuite struct {
	suite.Suite
	db              *gorm.DB
	authSvc         *svc.AuthService
	orderSvc        *svc.OrderService
	paymentSvc      *svc.PaymentService
	commissionSvc   *svc.CommissionService
	salarySvc       *svc.SalaryService
	inventorySvc    *svc.InventoryService
	configSvc       *svc.ConfigService
	customerSvc     *svc.CustomerService
	userRepo        *repository.UserRepository
	permRepo        *repository.PermissionRepository
	orderRepo       *repository.OrderRepository
	paymentRepo     *repository.PaymentRepository
	commissionRepo  *repository.CommissionRepository
	salaryRepo      *repository.SalaryRecordRepository
	fundPoolRepo    *repository.FundPoolRepository
	inventoryRepo   *repository.InventoryRepository
	configRepo      *repository.SystemConfigRepository
	customerRepo    *repository.CustomerRepository
}

func (s *OrderFlowTestSuite) SetupSuite() {
	dbName := fmt.Sprintf("file:order_flow_test?mode=memory&cache=shared")
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	s.Require().NoError(err)
	s.db = db

	// 自动迁移所有表
	err = db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.UserRole{},
		&models.RolePermission{},
		&models.Store{},
		&models.Category{},
		&models.Product{},
		&models.ProductSKU{},
		&models.Warehouse{},
		&models.WarehouseStock{},
		&models.InventoryBatch{},
		&models.InventoryTransaction{},
		&models.Order{},
		&models.OrderItem{},
		&models.OrderGift{},
		&models.Payment{},
		&models.Customer{},
		&models.CustomerFollowUp{},
		&models.Peer{},
		&models.PurchaseOrder{},
		&models.PurchaseItem{},
		&models.Supplier{},
		&models.Commission{},
		&models.FundPool{},
		&models.FundPoolShare{},
		&models.ReferralRelation{},
		&models.SalaryRecord{},
		&models.SalaryItem{},
		&models.SystemConfig{},
	)
	s.Require().NoError(err)

	// 初始化仓库
	s.userRepo = repository.NewUserRepository(db)
	s.permRepo = repository.NewPermissionRepository(db)
	s.orderRepo = repository.NewOrderRepository(db)
	s.paymentRepo = repository.NewPaymentRepository(db)
	s.commissionRepo = repository.NewCommissionRepository(db)
	s.salaryRepo = repository.NewSalaryRecordRepository(db)
	s.fundPoolRepo = repository.NewFundPoolRepository(db)
	s.inventoryRepo = repository.NewInventoryRepository(db)
	s.configRepo = repository.NewSystemConfigRepository(db)
	s.customerRepo = repository.NewCustomerRepository(db)
	referralRepo := repository.NewReferralRelationRepository(db)

	// 初始化服务
	s.authSvc = svc.NewAuthService(s.userRepo, s.permRepo)
	s.configSvc = svc.NewConfigService(s.configRepo)
	s.commissionSvc = svc.NewCommissionService(db, s.commissionRepo, s.orderRepo, referralRepo, s.configSvc)
	s.inventorySvc = svc.NewInventoryService(db, s.inventoryRepo, nil, nil)
	s.customerSvc = svc.NewCustomerService(db, s.customerRepo)
	s.orderSvc = svc.NewOrderService(db, s.orderRepo, s.paymentRepo, s.customerRepo, nil, s.inventorySvc)
	s.paymentSvc = svc.NewPaymentService(db, s.paymentRepo, s.orderRepo)
	s.salarySvc = svc.NewSalaryService(db, s.salaryRepo, s.commissionRepo, s.fundPoolRepo)

	// 初始化JWT配置
	configs.GlobalConfig = &configs.Config{
		JWT: configs.JWTConfig{
			Secret:      "test-secret-key-for-integration",
			ExpireHours: 24,
		},
	}
}

func (s *OrderFlowTestSuite) TearDownSuite() {
	sqlDB, _ := s.db.DB()
	sqlDB.Close()
}

func (s *OrderFlowTestSuite) SetupTest() {
	// 清理所有测试数据
	tables := []string{
		"salary_items", "salary_records", "commissions", "fund_pool_shares", "fund_pools",
		"referral_relations", "inventory_transactions", "inventory_batches", "warehouse_stocks",
		"order_gifts", "order_items", "payments", "orders", "customer_follow_ups", "customers",
		"peers", "purchase_items", "purchase_orders", "suppliers", "product_skus", "products",
		"categories", "warehouses", "user_roles", "role_permissions", "permissions", "roles",
		"users", "stores", "system_configs",
	}
	for _, table := range tables {
		s.db.Exec(fmt.Sprintf("DELETE FROM %s", table))
	}
}

// seedBaseData 准备基础数据
func (s *OrderFlowTestSuite) seedBaseData() {
	// 创建门店
	store := &models.Store{
		ID:        1,
		StoreCode: "STORE-001",
		StoreName: "测试门店",
		Status:    1,
	}
	s.db.Create(store)

	// 创建角色
	roles := []models.Role{
		{ID: 1, RoleCode: "admin", RoleName: "管理员", Status: 1, SortOrder: 1},
		{ID: 2, RoleCode: "salesman", RoleName: "业务员", Status: 1, SortOrder: 2},
	}
	for _, role := range roles {
		s.db.Create(&role)
	}

	// 创建权限
	perms := []models.Permission{
		{ID: 1, PermissionName: "订单管理", PermissionCode: "order:manage", PermissionType: 1, Status: 1},
	}
	for _, perm := range perms {
		s.db.Create(&perm)
	}

	// 创建店长（用于审核）
	storeManagerID := int64(100)
	storeManager := &models.User{
		ID:         storeManagerID,
		StoreID:    intPtr64(1),
		EmployeeNo: "EMP-100",
		Username:   "store_manager",
		Password:   "$2a$10$7xaz/QDlO1axD4kGdf6lre4PQ0hTV1/4lR3mc/vJBfcqZF0DhH2gq", // bcrypt("123456")
		RealName:   "店长",
		Phone:      "13800000100",
		Status:     1,
		BaseSalary: 8000,
	}
	s.db.Create(storeManager)

	// 更新门店店长
	s.db.Model(&models.Store{}).Where("id = ?", 1).Update("manager_id", storeManagerID)

	// 创建业务员
	salesmanID := int64(10)
	salesman := &models.User{
		ID:         salesmanID,
		StoreID:    intPtr64(1),
		EmployeeNo: "EMP-010",
		Username:   "salesman01",
		Password:   "$2a$10$7xaz/QDlO1axD4kGdf6lre4PQ0hTV1/4lR3mc/vJBfcqZF0DhH2gq", // bcrypt("123456")
		RealName:   "业务员A",
		Phone:      "13800000010",
		Status:     1,
		BaseSalary: 3000,
	}
	s.db.Create(salesman)

	// 创建仓库
	warehouse := &models.Warehouse{
		ID:            1,
		StoreID:       1,
		WarehouseCode: "WH-001",
		WarehouseName: "默认仓库",
		WarehouseType: 1,
		Status:        1,
	}
	s.db.Create(warehouse)

	// 创建SKU
	sku := &models.ProductSKU{
		ID:        100,
		ProductID: 1,
		SKUCode:   "SKU-SOFA-001",
		SKUName:   "真皮沙发-标准款",
		Status:    1,
	}
	s.db.Create(sku)

	// 创建库存和批次
	now := time.Now()
	stock := &models.WarehouseStock{
		WarehouseID:       1,
		SKUID:             100,
		StockQuantity:     100,
		AvailableQuantity: 100,
		LockedQuantity:    0,
		WarningStock:      10,
		Version:           0,
	}
	s.db.Create(stock)

	batch := &models.InventoryBatch{
		SKUID:             100,
		BatchNo:           "BATCH-INIT-001",
		PurchasePrice:     decimal.NewFromFloat(100.00),
		TotalCost:         decimal.NewFromFloat(10000.00),
		InitialQuantity:   100,
		RemainingQuantity: 100,
		WarehouseID:       intPtr64(1),
		Status:            1,
		EntryDate:         &now,
	}
	s.db.Create(batch)
}

// TestOrderFullFlow 完整订单流程测试
func (s *OrderFlowTestSuite) TestOrderFullFlow() {
	s.seedBaseData()

	// ========== 步骤1: 登录获取Token ==========
	loginResp, err := s.authSvc.Login("salesman01", "123456", "127.0.0.1")
	s.NoError(err)
	s.NotNil(loginResp)
	s.NotEmpty(loginResp.Token)
	s.Equal(int64(10), loginResp.User.ID)
	s.Equal("业务员A", loginResp.User.RealName)

	// ========== 步骤2: 创建订单 ==========
	catID := int64(1)
	createOrderReq := &svc.CreateOrderRequest{
		StoreID:        1,
		SalesmanID:     10,
		CustomerName:   "流程测试客户",
		CustomerPhone:  "13800139000",
		CustomerAddress: "测试地址",
		Items: []svc.CreateOrderItemRequest{
			{
				SKUID:       100,
				ProductName: "真皮沙发-标准款",
				SKUName:     "SKU-SOFA-001",
				CategoryID:  &catID,
				Quantity:    5,
				ListPrice:   200.00,
				SalePrice:   180.00,
			},
		},
	}

	order, err := s.orderSvc.CreateOrder(createOrderReq, 10)
	s.NoError(err)
	s.NotNil(order)
	s.NotZero(order.ID)
	s.Equal(int8(0), order.OrderStatus) // 待审批
	s.True(s.orderSvc != nil) // 确保服务可用

	// 验证库存已锁定
	var stock models.WarehouseStock
	s.db.Where("warehouse_id = ? AND sku_id = ?", 1, 100).First(&stock)
	s.Equal(95, stock.AvailableQuantity)  // 100 - 5 = 95
	s.Equal(5, stock.LockedQuantity)     // 0 + 5 = 5

	// ========== 步骤3: 审核通过 ==========
	err = s.orderSvc.ApproveOrder(order.ID, 100, true, "审核通过")
	s.NoError(err)

	// 验证订单状态
	orderDetail, err := s.orderSvc.GetDetail(order.ID)
	s.NoError(err)
	s.Equal(int8(1), orderDetail.Order.OrderStatus) // 已生效

	// 验证成本计算：5个@100元 = 500
	s.True(orderDetail.Order.TotalCost.Equal(decimal.NewFromFloat(500.00)))
	// 验证利润：final_amount(900) - total_cost(500) = 400
	s.True(orderDetail.Order.ActualProfit.Equal(decimal.NewFromFloat(400.00)))

	// 验证库存扣减
	s.db.Where("warehouse_id = ? AND sku_id = ?", 1, 100).First(&stock)
	s.Equal(95, stock.StockQuantity)      // 100 - 5 = 95
	s.Equal(95, stock.AvailableQuantity)  // 95 (锁定已转为扣减)
	s.Equal(0, stock.LockedQuantity)     // 5 - 5 = 0

	// ========== 步骤4: 录入回款 ==========
	paymentReq := &svc.CreatePaymentRequest{
		OrderID:       order.ID,
		Amount:        900.00, // 全额回款
		PaymentDate:   time.Now().Format("2006-01-02"),
		PaymentMethod: 1,
		Remark:        "全额回款",
	}

	err = s.paymentSvc.CreatePayment(paymentReq, 10)
	s.NoError(err)

	// 获取回款记录
	payments, err := s.paymentSvc.GetByOrderID(order.ID)
	s.NoError(err)
	s.Len(payments, 1)

	// 审核回款
	err = s.paymentSvc.ApprovePayment(payments[0].ID, 100, true)
	s.NoError(err)

	// 验证订单回款状态
	updatedOrder, _ := s.orderRepo.FindByID(order.ID)
	s.Equal(int8(2), updatedOrder.PaymentStatus) // 已回款

	// ========== 步骤5: 验证提成生成 ==========
	err = s.commissionSvc.CalculateOrderCommission(order.ID)
	s.NoError(err)

	// 验证提成记录
	commissions, err := s.commissionSvc.GetByOrderID(order.ID)
	s.NoError(err)
	s.NotEmpty(commissions)

	// 验证业务员提成：400 * 0.20 = 80
	var salesCommission *models.Commission
	for i := range commissions {
		if commissions[i].CommissionType == 1 {
			salesCommission = &commissions[i]
			break
		}
	}
	s.NotNil(salesCommission)
	s.True(salesCommission.Amount.Equal(decimal.NewFromFloat(80.00)))
	s.Equal(int8(1), salesCommission.Status) // 可发放

	// ========== 步骤6: 生成工资 ==========
	err = s.salarySvc.GenerateSalary(1, time.Now().Format("2006-01"))
	s.NoError(err)

	// 验证工资记录已生成
	salaryRecord, err := s.salarySvc.GetEmployeeSalary(10, time.Now().Format("2006-01"))
	s.NoError(err)
	s.NotNil(salaryRecord)
	s.Equal(int64(10), salaryRecord.EmployeeID)
	s.Equal(int8(0), salaryRecord.Status) // 草稿

	// 验证基本工资
	s.True(salaryRecord.BaseSalary.Equal(decimal.NewFromFloat(3000.00)))

	// 验证销售提成
	s.True(salaryRecord.SalesCommission.GreaterThan(decimal.Zero))

	// 验证净工资
	s.True(salaryRecord.NetSalary.GreaterThan(decimal.NewFromFloat(3000.00)))

	// 确认工资
	err = s.salarySvc.ConfirmSalary(salaryRecord.ID, 100)
	s.NoError(err)

	confirmedRecord, _ := s.salarySvc.GetDetail(salaryRecord.ID)
	s.Equal(int8(1), confirmedRecord.Status) // 已确认
}

// TestOrderFullFlow_WithPaymentValidation 带回款验证的完整流程
func (s *OrderFlowTestSuite) TestOrderFullFlow_WithPaymentValidation() {
	s.seedBaseData()

	// 创建订单
	catID := int64(1)
	req := &svc.CreateOrderRequest{
		StoreID:       1,
		SalesmanID:    10,
		CustomerName:  "回款验证客户",
		CustomerPhone: "13800139001",
		Items: []svc.CreateOrderItemRequest{
			{
				SKUID:       100,
				ProductName: "真皮沙发-标准款",
				SKUName:     "SKU-SOFA-001",
				CategoryID:  &catID,
				Quantity:    2,
				ListPrice:   200.00,
				SalePrice:   180.00,
			},
		},
	}

	order, err := s.orderSvc.CreateOrder(req, 10)
	s.NoError(err)

	// 审核通过
	err = s.orderSvc.ApproveOrder(order.ID, 100, true, "审核通过")
	s.NoError(err)

	// 部分回款
	payReq1 := &svc.CreatePaymentRequest{
		OrderID: order.ID,
		Amount:  200.00,
	}
	s.paymentSvc.CreatePayment(payReq1, 10)

	payments, _ := s.paymentSvc.GetByOrderID(order.ID)
	s.paymentSvc.ApprovePayment(payments[0].ID, 100, true)

	// 验证部分回款状态
	updatedOrder, _ := s.orderRepo.FindByID(order.ID)
	s.Equal(int8(1), updatedOrder.PaymentStatus) // 部分回款
	s.True(updatedOrder.PaidAmount.Equal(decimal.NewFromFloat(200.00)))

	// 剩余回款 - 一次性全额回款
	payReq2 := &svc.CreatePaymentRequest{
		OrderID: order.ID,
		Amount:  160.00, // 360 - 200 = 160
	}
	err = s.paymentSvc.CreatePayment(payReq2, 10)
	s.NoError(err) // 确保第二次回款也成功

	payments, _ = s.paymentSvc.GetByOrderID(order.ID)
	s.Require().GreaterOrEqual(len(payments), 2, "应该有2笔回款记录")
	// 审核第二笔回款（找到未审核的那条）
	for _, p := range payments {
		if p.Status == 0 {
			err = s.paymentSvc.ApprovePayment(p.ID, 100, true)
			s.NoError(err)
		}
	}

	// 验证全额回款状态
	updatedOrder, _ = s.orderRepo.FindByID(order.ID)
	// 调试：打印实际值
	s.T().Logf("DEBUG: PaidAmount=%s, FinalAmount=%s, PaymentStatus=%d",
		updatedOrder.PaidAmount.String(), updatedOrder.FinalAmount.String(), updatedOrder.PaymentStatus)
	s.Equal(int8(2), updatedOrder.PaymentStatus) // 已回款
	s.True(updatedOrder.PaidAmount.Equal(decimal.NewFromFloat(360.00)))
}

// TestOrderFullFlow_PartialPaymentAndCommission 部分回款后生成提成
func (s *OrderFlowTestSuite) TestOrderFullFlow_PartialPaymentAndCommission() {
	s.seedBaseData()

	// 创建大额订单
	catID := int64(1)
	req := &svc.CreateOrderRequest{
		StoreID:       1,
		SalesmanID:    10,
		CustomerName:  "大额订单客户",
		CustomerPhone: "13800139002",
		Items: []svc.CreateOrderItemRequest{
			{
				SKUID:       100,
				ProductName: "真皮沙发-标准款",
				SKUName:     "SKU-SOFA-001",
				CategoryID:  &catID,
				Quantity:    10,
				ListPrice:   200.00,
				SalePrice:   180.00,
			},
		},
	}

	order, err := s.orderSvc.CreateOrder(req, 10)
	s.NoError(err)

	// 审核通过
	err = s.orderSvc.ApproveOrder(order.ID, 100, true, "审核通过")
	s.NoError(err)

	// 验证利润：final_amount(1800) - total_cost(1000) = 800
	orderDetail, _ := s.orderSvc.GetDetail(order.ID)
	s.True(orderDetail.Order.ActualProfit.Equal(decimal.NewFromFloat(800.00)))

	// 部分回款
	payReq := &svc.CreatePaymentRequest{
		OrderID: order.ID,
		Amount:  1000.00,
	}
	s.paymentSvc.CreatePayment(payReq, 10)

	payments, _ := s.paymentSvc.GetByOrderID(order.ID)
	s.paymentSvc.ApprovePayment(payments[0].ID, 100, true)

	// 生成提成
	err = s.commissionSvc.CalculateOrderCommission(order.ID)
	s.NoError(err)

	commissions, _ := s.commissionSvc.GetByOrderID(order.ID)
	s.NotEmpty(commissions)

	// 验证业务员提成：800 * 0.20 = 160
	var salesCommission *models.Commission
	for i := range commissions {
		if commissions[i].CommissionType == 1 {
			salesCommission = &commissions[i]
			break
		}
	}
	s.NotNil(salesCommission)
	s.True(salesCommission.Amount.Equal(decimal.NewFromFloat(160.00)))
}

// 辅助函数
func intPtr64(v int64) *int64 {
	return &v
}

func TestOrderFlowSuite(t *testing.T) {
	suite.Run(t, new(OrderFlowTestSuite))
}
