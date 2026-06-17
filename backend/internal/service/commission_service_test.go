package service

import (
	"fmt"
	"testing"

	"furniture-commission/internal/models"
	"furniture-commission/internal/repository"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupCommissionTestDB 创建提成测试数据库
func setupCommissionTestDB(t *testing.T) *gorm.DB {
	dbName := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(
		&models.Order{},
		&models.OrderItem{},
		&models.OrderGift{},
		&models.Payment{},
		&models.Customer{},
		&models.Peer{},
		&models.WarehouseStock{},
		&models.InventoryBatch{},
		&models.InventoryTransaction{},
		&models.ProductSKU{},
		&models.Warehouse{},
		&models.User{},
		&models.Store{},
		&models.Commission{},
		&models.FundPool{},
		&models.FundPoolShare{},
		&models.ReferralRelation{},
		&models.SalaryRecord{},
		&models.SalaryItem{},
		&models.SystemConfig{},
	)
	assert.NoError(t, err)

	return db
}

// setupCommissionTestService 创建提成测试服务
func setupCommissionTestService(t *testing.T) (*CommissionService, *gorm.DB) {
	db := setupCommissionTestDB(t)
	commissionRepo := repository.NewCommissionRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	referralRepo := repository.NewReferralRelationRepository(db)
	configRepo := repository.NewSystemConfigRepository(db)
	configService := NewConfigService(configRepo)
	commissionSvc := NewCommissionService(db, commissionRepo, orderRepo, referralRepo, configService)
	return commissionSvc, db
}

// createTestStore 创建测试门店
func createTestStore(t *testing.T, db *gorm.DB, id int64, managerID *int64) {
	store := &models.Store{
		ID:        id,
		StoreCode: "STORE-001",
		StoreName: "测试门店",
		ManagerID: managerID,
		Status:    1,
	}
	err := db.Create(store).Error
	assert.NoError(t, err)
}

// createTestUser 创建测试用户
func createTestUser(t *testing.T, db *gorm.DB, id int64, storeID *int64, parentID *int64, baseSalary float64) {
	user := &models.User{
		ID:         id,
		StoreID:    storeID,
		EmployeeNo: fmt.Sprintf("EMP-%d", id),
		Username:   fmt.Sprintf("user%d", id),
		Password:   "hashed_password",
		RealName:   fmt.Sprintf("员工%d", id),
		Phone:      fmt.Sprintf("138%010d", id),
		Status:     1,
		ParentID:   parentID,
		BaseSalary: decimal.NewFromFloat(baseSalary),
	}
	err := db.Create(user).Error
	assert.NoError(t, err)
}

// createApprovedOrder 创建已审核通过的订单
func createApprovedOrder(t *testing.T, db *gorm.DB, orderID int64, storeID, salesmanID int64, orderType int8, actualProfit float64, isPeerOrder int8, peerID *int64) {
	order := &models.Order{
		ID:            orderID,
		StoreID:       storeID,
		OrderNo:       fmt.Sprintf("ORD-%d", orderID),
		SalesmanID:    salesmanID,
		CustomerName:  "测试客户",
		CustomerPhone: "13800000001",
		OrderType:     orderType,
		OrderStatus:   1, // 已生效
		PaymentStatus: 2, // 已回款
		FinalAmount:   decimal.NewFromFloat(actualProfit + 500),
		TotalCost:     decimal.NewFromFloat(500),
		ActualProfit:  decimal.NewFromFloat(actualProfit),
		IsPeerOrder:   isPeerOrder,
		PeerID:        peerID,
	}
	err := db.Create(order).Error
	assert.NoError(t, err)
}

// TestCalculateSingleItemCommission 单品提成20%
func TestCalculateSingleItemCommission(t *testing.T) {
	commissionSvc, db := setupCommissionTestService(t)

	// 准备测试数据
	managerID := int64(100)
	storeManagerID := int64(101)
	createTestStore(t, db, 1, &storeManagerID)
	createTestUser(t, db, 10, intPtr64(1), &managerID, 3000)      // 业务员，主管=100
	createTestUser(t, db, managerID, intPtr64(1), nil, 5000)      // 主管
	createTestUser(t, db, storeManagerID, intPtr64(1), nil, 8000) // 店长

	// 创建单品订单，利润1000元
	createApprovedOrder(t, db, 1, 1, 10, 1, 1000, 0, nil)

	// 计算提成
	err := commissionSvc.CalculateOrderCommission(1)
	assert.NoError(t, err)

	// 验证提成记录
	commissions, err := commissionSvc.GetByOrderID(1)
	assert.NoError(t, err)
	assert.Len(t, commissions, 4) // 业务员提成 + 主管分润 + 店长分润 + 基金池

	// 验证业务员提成 1000 * 0.20 = 200
	var salesCommission *models.Commission
	for i := range commissions {
		if commissions[i].CommissionType == 1 {
			salesCommission = &commissions[i]
			break
		}
	}
	assert.NotNil(t, salesCommission)
	assert.True(t, salesCommission.Amount.Equal(decimal.NewFromFloat(200)))
	assert.Equal(t, int8(1), salesCommission.Status) // 可发放

	// 验证主管团队分润 1000 * 0.03 = 30
	var managerShare *models.Commission
	for i := range commissions {
		if commissions[i].CommissionType == 3 {
			managerShare = &commissions[i]
			break
		}
	}
	assert.NotNil(t, managerShare)
	assert.True(t, managerShare.Amount.Equal(decimal.NewFromFloat(30)))

	// 验证店长团队分润 1000 * 0.02 = 20
	var storeShare *models.Commission
	for i := range commissions {
		if commissions[i].CommissionType == 4 {
			storeShare = &commissions[i]
			break
		}
	}
	assert.NotNil(t, storeShare)
	assert.True(t, storeShare.Amount.Equal(decimal.NewFromFloat(20)))

	// 验证基金池提取 1000 * 0.05 = 50
	var fundPoolCommission *models.Commission
	for i := range commissions {
		if commissions[i].CommissionType == 5 {
			fundPoolCommission = &commissions[i]
			break
		}
	}
	assert.NotNil(t, fundPoolCommission)
	assert.True(t, fundPoolCommission.Amount.Equal(decimal.NewFromFloat(50)))
}

// TestCalculateMultiItemCommission 多品提成22%
func TestCalculateMultiItemCommission(t *testing.T) {
	commissionSvc, db := setupCommissionTestService(t)

	// 准备测试数据
	storeManagerID := int64(101)
	createTestStore(t, db, 1, &storeManagerID)
	createTestUser(t, db, 10, intPtr64(1), nil, 3000)             // 业务员（无主管）
	createTestUser(t, db, storeManagerID, intPtr64(1), nil, 8000) // 店长

	// 创建多品订单，利润2000元
	createApprovedOrder(t, db, 2, 1, 10, 2, 2000, 0, nil)

	// 计算提成
	err := commissionSvc.CalculateOrderCommission(2)
	assert.NoError(t, err)

	// 验证提成记录
	commissions, err := commissionSvc.GetByOrderID(2)
	assert.NoError(t, err)

	// 验证业务员提成 2000 * 0.22 = 440
	var salesCommission *models.Commission
	for i := range commissions {
		if commissions[i].CommissionType == 1 {
			salesCommission = &commissions[i]
			break
		}
	}
	assert.NotNil(t, salesCommission)
	assert.True(t, salesCommission.Amount.Equal(decimal.NewFromFloat(440)))
}

// TestCalculateSpecialCommission 特批提成15%
func TestCalculateSpecialCommission(t *testing.T) {
	commissionSvc, db := setupCommissionTestService(t)

	// 准备测试数据
	storeManagerID := int64(101)
	createTestStore(t, db, 1, &storeManagerID)
	createTestUser(t, db, 10, intPtr64(1), nil, 3000)
	createTestUser(t, db, storeManagerID, intPtr64(1), nil, 8000)

	// 创建特批订单，利润3000元
	createApprovedOrder(t, db, 3, 1, 10, 3, 3000, 0, nil)

	// 计算提成
	err := commissionSvc.CalculateOrderCommission(3)
	assert.NoError(t, err)

	// 验证业务员提成 3000 * 0.15 = 450
	commissions, err := commissionSvc.GetByOrderID(3)
	assert.NoError(t, err)

	var salesCommission *models.Commission
	for i := range commissions {
		if commissions[i].CommissionType == 1 {
			salesCommission = &commissions[i]
			break
		}
	}
	assert.NotNil(t, salesCommission)
	assert.True(t, salesCommission.Amount.Equal(decimal.NewFromFloat(450)))
}

// TestCalculatePeerCommission 同行分成10%/12%/8%
func TestCalculatePeerCommission(t *testing.T) {
	commissionSvc, db := setupCommissionTestService(t)

	// 准备测试数据
	storeManagerID := int64(101)
	createTestStore(t, db, 1, &storeManagerID)
	createTestUser(t, db, 10, intPtr64(1), nil, 3000)
	createTestUser(t, db, storeManagerID, intPtr64(1), nil, 8000)

	// 创建同行
	peer := &models.Peer{
		StoreID:  1,
		PeerName: "测试同行",
		Phone:    "13900000001",
		Status:   1,
	}
	err := db.Create(peer).Error
	assert.NoError(t, err)

	// 测试同行单品 order_type=4, rate=0.10
	createApprovedOrder(t, db, 4, 1, 10, 4, 5000, 1, &peer.ID)

	err = commissionSvc.CalculateOrderCommission(4)
	assert.NoError(t, err)

	commissions, err := commissionSvc.GetByOrderID(4)
	assert.NoError(t, err)

	// 验证业务员提成 5000 * 0.10 = 500
	var salesCommission *models.Commission
	for i := range commissions {
		if commissions[i].CommissionType == 1 {
			salesCommission = &commissions[i]
			break
		}
	}
	assert.NotNil(t, salesCommission)
	assert.True(t, salesCommission.Amount.Equal(decimal.NewFromFloat(500)))

	// 验证同行分成 5000 * 0.10 = 500
	var peerCommission *models.Commission
	for i := range commissions {
		if commissions[i].CommissionType == 2 {
			peerCommission = &commissions[i]
			break
		}
	}
	assert.NotNil(t, peerCommission)
	assert.True(t, peerCommission.Amount.Equal(decimal.NewFromFloat(500)))
	assert.NotNil(t, peerCommission.PeerID)
	assert.Equal(t, peer.ID, *peerCommission.PeerID)
}

// TestCalculatePeerMultiCommission 同行多品分成12%
func TestCalculatePeerMultiCommission(t *testing.T) {
	commissionSvc, db := setupCommissionTestService(t)

	storeManagerID := int64(101)
	createTestStore(t, db, 1, &storeManagerID)
	createTestUser(t, db, 10, intPtr64(1), nil, 3000)
	createTestUser(t, db, storeManagerID, intPtr64(1), nil, 8000)

	peer := &models.Peer{
		StoreID:  1,
		PeerName: "测试同行",
		Phone:    "13900000002",
		Status:   1,
	}
	err := db.Create(peer).Error
	assert.NoError(t, err)

	// 同行多品 order_type=5, rate=0.12
	createApprovedOrder(t, db, 5, 1, 10, 5, 5000, 1, &peer.ID)

	err = commissionSvc.CalculateOrderCommission(5)
	assert.NoError(t, err)

	commissions, err := commissionSvc.GetByOrderID(5)
	assert.NoError(t, err)

	var salesCommission *models.Commission
	for i := range commissions {
		if commissions[i].CommissionType == 1 {
			salesCommission = &commissions[i]
			break
		}
	}
	assert.NotNil(t, salesCommission)
	assert.True(t, salesCommission.Amount.Equal(decimal.NewFromFloat(600))) // 5000 * 0.12
}

// TestCalculateTeamShare 主管/店长团队分润
func TestCalculateTeamShare(t *testing.T) {
	commissionSvc, db := setupCommissionTestService(t)

	// 准备测试数据：业务员 -> 主管 -> (无上级)
	managerID := int64(100)
	storeManagerID := int64(101)
	createTestStore(t, db, 1, &storeManagerID)
	createTestUser(t, db, 10, intPtr64(1), &managerID, 3000)      // 业务员，主管=100
	createTestUser(t, db, managerID, intPtr64(1), nil, 5000)      // 主管
	createTestUser(t, db, storeManagerID, intPtr64(1), nil, 8000) // 店长

	// 创建单品订单，利润10000元
	createApprovedOrder(t, db, 6, 1, 10, 1, 10000, 0, nil)

	err := commissionSvc.CalculateOrderCommission(6)
	assert.NoError(t, err)

	commissions, err := commissionSvc.GetByOrderID(6)
	assert.NoError(t, err)

	// 验证主管团队分润 10000 * 0.03 = 300
	var managerShare *models.Commission
	for i := range commissions {
		if commissions[i].CommissionType == 3 {
			managerShare = &commissions[i]
			break
		}
	}
	assert.NotNil(t, managerShare)
	assert.True(t, managerShare.Amount.Equal(decimal.NewFromFloat(300)))
	assert.Equal(t, managerID, *managerShare.EmployeeID)

	// 验证店长团队分润 10000 * 0.02 = 200
	var storeShare *models.Commission
	for i := range commissions {
		if commissions[i].CommissionType == 4 {
			storeShare = &commissions[i]
			break
		}
	}
	assert.NotNil(t, storeShare)
	assert.True(t, storeShare.Amount.Equal(decimal.NewFromFloat(200)))
	assert.Equal(t, storeManagerID, *storeShare.EmployeeID)
}

// TestCalculateFundPool 基金池提取5%
func TestCalculateFundPool(t *testing.T) {
	commissionSvc, db := setupCommissionTestService(t)

	storeManagerID := int64(101)
	createTestStore(t, db, 1, &storeManagerID)
	createTestUser(t, db, 10, intPtr64(1), nil, 3000)
	createTestUser(t, db, storeManagerID, intPtr64(1), nil, 8000)

	// 创建订单，利润20000元
	createApprovedOrder(t, db, 7, 1, 10, 1, 20000, 0, nil)

	err := commissionSvc.CalculateOrderCommission(7)
	assert.NoError(t, err)

	commissions, err := commissionSvc.GetByOrderID(7)
	assert.NoError(t, err)

	// 验证基金池提取 20000 * 0.05 = 1000
	var fundPoolCommission *models.Commission
	for i := range commissions {
		if commissions[i].CommissionType == 5 {
			fundPoolCommission = &commissions[i]
			break
		}
	}
	assert.NotNil(t, fundPoolCommission)
	assert.True(t, fundPoolCommission.Amount.Equal(decimal.NewFromFloat(1000)))
	assert.Equal(t, int8(1), fundPoolCommission.Status) // 可发放
}

// TestCalculateReferralReward 老带新奖励
func TestCalculateReferralReward(t *testing.T) {
	commissionSvc, db := setupCommissionTestService(t)

	storeManagerID := int64(101)
	referrerID := int64(200) // 引荐人
	createTestStore(t, db, 1, &storeManagerID)
	createTestUser(t, db, referrerID, intPtr64(1), nil, 5000)     // 引荐人
	createTestUser(t, db, 10, intPtr64(1), nil, 3000)             // 业务员
	createTestUser(t, db, storeManagerID, intPtr64(1), nil, 8000) // 店长

	// 创建老带新关系
	referral := &models.ReferralRelation{
		ReferrerID: referrerID,
		ReferredID: 10, // 业务员10被引荐人200引荐
		Status:     1,
	}
	err := db.Create(referral).Error
	assert.NoError(t, err)

	// 创建单品订单，利润1000元
	createApprovedOrder(t, db, 8, 1, 10, 1, 1000, 0, nil)

	err = commissionSvc.CalculateOrderCommission(8)
	assert.NoError(t, err)

	commissions, err := commissionSvc.GetByOrderID(8)
	assert.NoError(t, err)

	// 验证老带新奖励：业务员提成(200) * 0.10 = 20
	var referralReward *models.Commission
	for i := range commissions {
		if commissions[i].CommissionType == 6 {
			referralReward = &commissions[i]
			break
		}
	}
	assert.NotNil(t, referralReward)
	assert.True(t, referralReward.Amount.Equal(decimal.NewFromFloat(20)))
	assert.Equal(t, referrerID, *referralReward.EmployeeID)
	assert.Equal(t, int8(1), referralReward.Status) // 可发放
}

// TestDuplicateCommissionCalculation 重复计算提成应失败
func TestDuplicateCommissionCalculation(t *testing.T) {
	commissionSvc, db := setupCommissionTestService(t)

	storeManagerID := int64(101)
	createTestStore(t, db, 1, &storeManagerID)
	createTestUser(t, db, 10, intPtr64(1), nil, 3000)
	createTestUser(t, db, storeManagerID, intPtr64(1), nil, 8000)

	createApprovedOrder(t, db, 9, 1, 10, 1, 1000, 0, nil)

	// 第一次计算
	err := commissionSvc.CalculateOrderCommission(9)
	assert.NoError(t, err)

	// 第二次计算应失败
	err = commissionSvc.CalculateOrderCommission(9)
	assert.Error(t, err)
	var _ *AppError = err.(*AppError)
}

// TestCalculateCommissionInvalidOrderStatus 无效订单状态不能计算提成
func TestCalculateCommissionInvalidOrderStatus(t *testing.T) {
	commissionSvc, db := setupCommissionTestService(t)

	storeManagerID := int64(101)
	createTestStore(t, db, 1, &storeManagerID)
	createTestUser(t, db, 10, intPtr64(1), nil, 3000)

	// 创建待审批订单（order_status=0）
	order := &models.Order{
		ID:            10,
		StoreID:       1,
		OrderNo:       "ORD-10",
		SalesmanID:    10,
		CustomerName:  "测试客户",
		CustomerPhone: "13800000010",
		OrderType:     1,
		OrderStatus:   0, // 待审批
	}
	err := db.Create(order).Error
	assert.NoError(t, err)

	err = commissionSvc.CalculateOrderCommission(10)
	assert.Error(t, err)
	var _ *AppError = err.(*AppError)
}

// intPtr64 辅助函数：创建int64指针
func intPtr64(v int64) *int64 {
	return &v
}
