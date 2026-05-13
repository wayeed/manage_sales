package database

import (
	"fmt"
	"time"

	"furniture-commission/configs"
	"furniture-commission/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(cfg *configs.DatabaseConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Info),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// 自动迁移新增的表（不会修改已有表结构）
	if err := DB.AutoMigrate(
		&models.CustomerFollowUp{},
		&models.FollowUpApproval{},
	); err != nil {
		fmt.Printf("[WARN] AutoMigrate failed: %v\n", err)
	}

	// 为已存在的表添加新列
	migrateColumns()

	return nil
}

// migrateColumns 为已存在的表添加新列或修改列类型
func migrateColumns() {
	// 添加 orders.is_draft 列
	if err := DB.Migrator().AddColumn(&models.Order{}, "is_draft"); err != nil {
		fmt.Printf("[INFO] AddColumn is_draft: %v\n", err)
	}
	// 添加 customers.salesman_id 列
	if err := DB.Migrator().AddColumn(&models.Customer{}, "salesman_id"); err != nil {
		fmt.Printf("[INFO] AddColumn salesman_id: %v\n", err)
	}
	// 修改 order_items.discount_rate 列类型为 DECIMAL(5,4)，并将旧数据从百分比转为小数
	migrateDiscountRate()

	// 初始化权限数据（INSERT IGNORE 幂等）
	seedPermissions()
}

// seedPermissions 初始化权限数据到 permissions 表
func seedPermissions() {
	// permission_type: 1=菜单 2=按钮 3=接口
	// 使用 INSERT IGNORE 避免重复插入
	perms := []struct {
		Code    string
		Name    string
		Type    int8
		Parent  string
		Sort    int
	}{
		// ===== 用户管理 =====
		{"user:create", "创建用户", 3, "", 101},
		{"user:delete", "删除用户", 3, "", 102},
		// ===== 角色管理 =====
		{"role:manage", "角色管理", 3, "", 201},
		// ===== 品类管理 =====
		{"category:manage", "品类管理", 3, "", 301},
		// ===== 商品管理 =====
		{"product:manage", "商品管理", 3, "", 401},
		// ===== 采购管理 =====
		{"purchase:manage", "采购管理", 3, "", 501},
		// ===== 礼品管理 =====
		{"gift:manage", "礼品管理", 3, "", 601},
		// ===== 供应商管理 =====
		{"supplier:manage", "供应商管理", 3, "", 701},
		// ===== 调拨管理 =====
		{"transfer:manage", "调拨管理", 3, "", 801},
		// ===== 库存预警 =====
		{"stock_alert:manage", "库存预警管理", 3, "", 901},
		// ===== 订单管理 =====
		{"order:approve", "订单审核", 3, "", 1001},
		// ===== 回款管理 =====
		{"payment:manage", "回款管理", 3, "", 1101},
		// ===== 同行管理 =====
		{"peer:manage", "同行管理", 3, "", 1201},
		// ===== 提成管理 =====
		{"commission:manage", "提成管理", 3, "", 1301},
		// ===== 工资管理 =====
		{"salary:manage", "工资管理", 3, "", 1401},
		// ===== 基金池 =====
		{"fund_pool:manage", "基金池管理", 3, "", 1501},
		// ===== 系统配置 =====
		{"config:manage", "系统配置管理", 3, "", 1601},
		// ===== 门店管理 =====
		{"store:manage", "门店管理", 3, "", 1701},
	}

	for _, p := range perms {
		DB.Exec(
			`INSERT IGNORE INTO permissions (permission_code, permission_name, permission_type, sort_order, status, created_at)
			 VALUES (?, ?, ?, ?, 1, NOW())`,
			p.Code, p.Name, p.Type, p.Sort,
		)
	}

	fmt.Printf("[INFO] 权限数据初始化完成，共 %d 项\n", len(perms))

	// 为 ADMIN 角色分配所有权限
	assignAllPermissionsToAdminRole()

	// 添加 payments 表缺少的 updated_at 列
	migratePaymentsTable()
}

// migratePaymentsTable 为 payments 表添加缺少的列
func migratePaymentsTable() {
	// 添加 updated_at 列
	if err := DB.Exec("ALTER TABLE payments ADD COLUMN IF NOT EXISTS updated_at datetime(3) DEFAULT NULL").Error; err != nil {
		fmt.Printf("[INFO] AddColumn payments.updated_at: %v\n", err)
	}
}

// assignAllPermissionsToAdminRole 为 admin 角色分配所有权限
func assignAllPermissionsToAdminRole() {
	// 获取 admin 角色 ID
	var adminRoleID int64
	if err := DB.Table("roles").Select("id").Where("role_code = ?", "admin").Scan(&adminRoleID).Error; err != nil {
		fmt.Printf("[INFO] 获取 admin 角色失败: %v\n", err)
		return
	}
	if adminRoleID == 0 {
		fmt.Printf("[INFO] admin 角色不存在，跳过权限分配\n")
		return
	}

	// 获取所有权限 ID
	var permIDs []int64
	if err := DB.Table("permissions").Select("id").Pluck("id", &permIDs).Error; err != nil {
		fmt.Printf("[INFO] 获取权限列表失败: %v\n", err)
		return
	}

	// 为 admin 角色分配所有权限（INSERT IGNORE 避免重复）
	assignedCount := 0
	for _, permID := range permIDs {
		result := DB.Exec(
			`INSERT IGNORE INTO role_permissions (role_id, permission_id) VALUES (?, ?)`,
			adminRoleID, permID,
		)
		if result.Error == nil && result.RowsAffected > 0 {
			assignedCount++
		}
	}

	fmt.Printf("[INFO] 已为 admin 角色分配 %d 项权限\n", assignedCount)
}

// migrateDiscountRate 修改折扣率列类型并转换数据
func migrateDiscountRate() {
	// 检查是否需要迁移：如果列已经是 decimal(5,4) 则跳过
	var columnType string
	DB.Raw("SELECT DATA_TYPE FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'order_items' AND COLUMN_NAME = 'discount_rate'").Scan(&columnType)
	if columnType == "decimal" {
		// 检查精度是否已经是 (5,4)
		// 直接执行 ALTER TABLE，如果已经是目标类型会报 warning 但不会报错
	}
	// 修改列类型
	if err := DB.Exec("ALTER TABLE order_items MODIFY COLUMN discount_rate DECIMAL(5,4) NOT NULL DEFAULT 1.0000").Error; err != nil {
		fmt.Printf("[INFO] Migrate discount_rate type: %v\n", err)
	}
	// 将旧的百分比值（>1）转换为小数（除以100）
	DB.Exec("UPDATE order_items SET discount_rate = ROUND(discount_rate / 100, 4) WHERE discount_rate > 1")
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
