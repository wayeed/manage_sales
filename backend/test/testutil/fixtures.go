package testutil

import (
	"fmt"
	"time"

	"furniture-commission/internal/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// FixtureManager 测试数据管理器
type FixtureManager struct {
	db *gorm.DB
}

// NewFixtureManager 创建测试数据管理器
func NewFixtureManager(db *gorm.DB) *FixtureManager {
	return &FixtureManager{db: db}
}

// Cleanup 清理所有测试数据（按外键依赖顺序）
func (f *FixtureManager) Cleanup() {
	tables := []string{
		"salary_items",
		"salary_records",
		"commissions",
		"fund_pool_shares",
		"fund_pools",
		"referral_relations",
		"inventory_transactions",
		"inventory_batches",
		"warehouse_stocks",
		"warehouse_gift_stocks",
		"gift_inventory_batches",
		"gifts",
		"transfer_items",
		"transfer_orders",
		"stock_alerts",
		"order_gifts",
		"order_items",
		"payments",
		"orders",
		"customer_follow_ups",
		"customers",
		"peers",
		"purchase_items",
		"purchase_orders",
		"supplier_products",
		"suppliers",
		"product_skus",
		"products",
		"categories",
		"warehouses",
		"user_roles",
		"role_permissions",
		"permissions",
		"roles",
		"users",
		"stores",
		"system_configs",
	}
	for _, table := range tables {
		f.db.Exec(fmt.Sprintf("DELETE FROM %s", table))
	}
}

// CreateStore 创建测试门店
func (f *FixtureManager) CreateStore(id int64, managerID *int64) *models.Store {
	store := &models.Store{
		ID:        id,
		StoreCode: fmt.Sprintf("STORE-%d", id),
		StoreName: "测试门店",
		ManagerID: managerID,
		Status:    1,
	}
	f.db.Create(store)
	return store
}

// CreateUser 创建测试用户
func (f *FixtureManager) CreateUser(id int64, storeID *int64, username, password, realName, phone string, status int8, role int8, baseSalary decimal.Decimal, parentID *int64) *models.User {
	user := &models.User{
		ID:          id,
		StoreID:     storeID,
		EmployeeNo:  fmt.Sprintf("EMP-%d", id),
		Username:    username,
		Password:    password,
		RealName:    realName,
		Phone:       phone,
		Status:      status,
		Role:        role,
		BaseSalary:  baseSalary,
		ParentID:    parentID,
	}
	f.db.Create(user)
	return user
}

// CreateRole 创建测试角色
func (f *FixtureManager) CreateRole(id int64, roleCode, roleName string, status int8, sortOrder int) *models.Role {
	role := &models.Role{
		ID:        id,
		RoleCode:  roleCode,
		RoleName:  roleName,
		Status:    status,
	}
	f.db.Create(role)
	return role
}

// CreatePermission 创建测试权限
func (f *FixtureManager) CreatePermission(id int64, parentID *int64, permissionName, permissionCode string, permissionType int8, sortOrder int, status int8) *models.Permission {
	perm := &models.Permission{
		ID:             id,
		ParentID:       parentID,
		PermissionName: permissionName,
		PermissionCode: permissionCode,
		PermissionType: permissionType,
		SortOrder:      sortOrder,
		Status:         status,
	}
	f.db.Create(perm)
	return perm
}

// CreateCategory 创建测试品类
func (f *FixtureManager) CreateCategory(id int64, storeID int64, code, name string, sortOrder int) *models.Category {
	cat := &models.Category{
		ID:           id,
		StoreID:      storeID,
		CategoryCode: code,
		CategoryName: name,
		SortOrder:    sortOrder,
		Status:       1,
	}
	f.db.Create(cat)
	return cat
}

// CreateProduct 创建测试商品
func (f *FixtureManager) CreateProduct(id int64, storeID int64, categoryID *int64, code, name string, listPrice, minPrice, refCost float64) *models.Product {
	product := &models.Product{
		ID:            id,
		StoreID:       storeID,
		CategoryID:    categoryID,
		ProductCode:   code,
		ProductName:   name,
		ListPrice:     decimal.NewFromFloat(listPrice),
		MinPrice:      decimal.NewFromFloat(minPrice),
		ReferenceCost: decimal.NewFromFloat(refCost),
		TotalCostRate: decimal.NewFromFloat(1.2),
		WarningStock:  10,
		Status:        1,
	}
	f.db.Create(product)
	return product
}

// CreateProductSKU 创建测试SKU
func (f *FixtureManager) CreateProductSKU(id, productID int64, skuCode, skuName string) *models.ProductSKU {
	sku := &models.ProductSKU{
		ID:        id,
		ProductID: productID,
		SKUCode:   skuCode,
		SKUName:   skuName,
		Status:    1,
	}
	f.db.Create(sku)
	return sku
}

// CreateWarehouse 创建测试仓库
func (f *FixtureManager) CreateWarehouse(id, storeID int64, code, name string) *models.Warehouse {
	warehouse := &models.Warehouse{
		ID:            id,
		StoreID:       storeID,
		WarehouseCode: code,
		WarehouseName: name,
		WarehouseType: 1,
		Status:        1,
	}
	f.db.Create(warehouse)
	return warehouse
}

// CreateWarehouseStock 创建测试库存
func (f *FixtureManager) CreateWarehouseStock(warehouseID, skuID int64, stockQty, availQty, lockedQty, warningQty int) *models.WarehouseStock {
	stock := &models.WarehouseStock{
		WarehouseID:       warehouseID,
		SKUID:             skuID,
		StockQuantity:     stockQty,
		AvailableQuantity: availQty,
		LockedQuantity:    lockedQty,
		WarningStock:      warningQty,
		Version:           0,
	}
	f.db.Create(stock)
	return stock
}

// CreateInventoryBatch 创建测试批次
func (f *FixtureManager) CreateInventoryBatch(skuID, warehouseID int64, batchNo string, price float64, qty, remaining int) *models.InventoryBatch {
	now := time.Now()
	batch := &models.InventoryBatch{
		SKUID:             skuID,
		BatchNo:           batchNo,
		PurchasePrice:     decimal.NewFromFloat(price),
		TotalCost:         decimal.NewFromFloat(price * float64(qty)),
		InitialQuantity:   qty,
		RemainingQuantity: remaining,
		WarehouseID:       &warehouseID,
		Status:            1,
		EntryDate:         &now,
	}
	f.db.Create(batch)
	return batch
}

// CreateSupplier 创建测试供应商
func (f *FixtureManager) CreateSupplier(id, storeID int64, code, name string) *models.Supplier {
	supplier := &models.Supplier{
		ID:           id,
		StoreID:      storeID,
		SupplierCode: code,
		SupplierName: name,
		Status:       1,
	}
	f.db.Create(supplier)
	return supplier
}

// CreateCustomer 创建测试客户
func (f *FixtureManager) CreateCustomer(id, storeID int64, name, phone string) *models.Customer {
	customer := &models.Customer{
		ID:           id,
		StoreID:      storeID,
		CustomerCode: fmt.Sprintf("CUST-%d", id),
		CustomerName: name,
		Phone:        phone,
		Status:       1,
	}
	f.db.Create(customer)
	return customer
}

// CreateApprovedOrder 创建已审核通过的订单
func (f *FixtureManager) CreateApprovedOrder(id, storeID, salesmanID int64, orderType int8, finalAmount, totalCost, actualProfit float64) *models.Order {
	now := time.Now()
	order := &models.Order{
		ID:            id,
		StoreID:       storeID,
		OrderNo:       fmt.Sprintf("ORD-%d", id),
		SalesmanID:    salesmanID,
		CustomerName:  "测试客户",
		CustomerPhone: "13800000001",
		OrderType:     orderType,
		OrderStatus:   1,
		PaymentStatus: 0,
		FinalAmount:   decimal.NewFromFloat(finalAmount),
		TotalCost:     decimal.NewFromFloat(totalCost),
		ActualProfit:  decimal.NewFromFloat(actualProfit),
		OrderDate:     &now,
	}
	f.db.Create(order)
	return order
}

// CreateConfig 创建测试配置
func (f *FixtureManager) CreateConfig(key, value, configType, remark string) *models.SystemConfig {
	config := &models.SystemConfig{
		ConfigKey:   key,
		ConfigValue: value,
		ConfigType:  configType,
		Remark:      remark,
	}
	f.db.Create(config)
	return config
}

// Int64Ptr int64指针辅助函数
func Int64Ptr(v int64) *int64 {
	return &v
}

// Int8Ptr int8指针辅助函数
func Int8Ptr(v int8) *int8 {
	return &v
}
