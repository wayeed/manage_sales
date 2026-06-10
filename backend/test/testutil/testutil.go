package testutil

import (
	"furniture-commission/configs"
	"furniture-commission/internal/models"
	"furniture-commission/internal/pkg/database"

	"gorm.io/gorm"
)

// SetupTestDB 初始化测试数据库连接
func SetupTestDB() *gorm.DB {
	cfg := &configs.DatabaseConfig{
		Host:         "127.0.0.1",
		Port:         3306,
		User:         "root",
		Password:     "root",
		DBName:       "furniture_commission",
		MaxIdleConns: 5,
		MaxOpenConns: 10,
	}
	err := database.InitDB(cfg, "debug")
	if err != nil {
		panic(err)
	}
	return database.GetDB()
}

// AllModels 返回所有需要自动迁移的模型
func AllModels() []interface{} {
	return []interface{}{
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
		&models.SupplierProduct{},
		&models.Commission{},
		&models.FundPool{},
		&models.FundPoolShare{},
		&models.ReferralRelation{},
		&models.SalaryRecord{},
		&models.SalaryItem{},
		&models.SystemConfig{},
		&models.TransferOrder{},
		&models.TransferItem{},
		&models.Gift{},
		&models.GiftInventoryBatch{},
		&models.WarehouseGiftStock{},
		&models.StockAlert{},
	}
}

// AutoMigrateAll 自动迁移所有表
func AutoMigrateAll(db *gorm.DB) error {
	return db.AutoMigrate(AllModels()...)
}
