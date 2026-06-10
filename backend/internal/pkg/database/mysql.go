package database

import (
	"fmt"
	"time"

	"furniture-commission/configs"
	"furniture-commission/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
// serverMode: debug/release/test，控制 SQL 日志输出级别
func InitDB(cfg *configs.DatabaseConfig, serverMode string) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	// 根据运行模式设置 GORM SQL 日志级别
	// debug: 打印所有 SQL（Info）
	// release: 仅打印慢查询和错误（Warn）
	// test: 静默（Silent）
	gormLogLevel := logger.Warn
	switch serverMode {
	case "debug":
		gormLogLevel = logger.Info
	case "release":
		gormLogLevel = logger.Warn
	case "test":
		gormLogLevel = logger.Silent
	}

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:                                   logger.Default.LogMode(gormLogLevel),
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

	// 自动迁移新增的表（只创建不存在的表，不修改已有表结构）
	// 注意：GORM AutoMigrate 会清除 MySQL 的 COMMENT，所以只用于创建新表
	autoMigrateNewTablesOnly()

	// 为已存在的表添加新列
	migrateColumns()

	return nil
}

// migrateColumns 为已存在的表添加新列或修改列类型
func migrateColumns() {
	// 添加 orders.is_draft 列（如果不存在）
	if !DB.Migrator().HasColumn(&models.Order{}, "is_draft") {
		if err := DB.Migrator().AddColumn(&models.Order{}, "is_draft"); err != nil {
			fmt.Printf("[ERROR] AddColumn is_draft: %v\n", err)
		} else {
			fmt.Println("[INFO] AddColumn is_draft: success")
		}
	}
	// 添加 customers.salesman_id 列（如果不存在）
	if !DB.Migrator().HasColumn(&models.Customer{}, "salesman_id") {
		if err := DB.Migrator().AddColumn(&models.Customer{}, "salesman_id"); err != nil {
			fmt.Printf("[ERROR] AddColumn salesman_id: %v\n", err)
		} else {
			fmt.Println("[INFO] AddColumn salesman_id: success")
		}
	}
	// 添加 customers.original_phone 列（如果不存在）
	if !DB.Migrator().HasColumn(&models.Customer{}, "original_phone") {
		if err := DB.Migrator().AddColumn(&models.Customer{}, "original_phone"); err != nil {
			fmt.Printf("[ERROR] AddColumn original_phone: %v\n", err)
		} else {
			fmt.Println("[INFO] AddColumn original_phone: success")
		}
	}
	// 添加 warehouse_stocks.in_transit_quantity 列（如果不存在）
	if !DB.Migrator().HasColumn(&models.WarehouseStock{}, "in_transit_quantity") {
		if err := DB.Migrator().AddColumn(&models.WarehouseStock{}, "in_transit_quantity"); err != nil {
			fmt.Printf("[ERROR] AddColumn in_transit_quantity: %v\n", err)
		} else {
			fmt.Println("[INFO] AddColumn in_transit_quantity: success")
		}
	}
	// 添加 warehouse_stocks.pending_quantity 列（如果不存在）
	if !DB.Migrator().HasColumn(&models.WarehouseStock{}, "pending_quantity") {
		if err := DB.Migrator().AddColumn(&models.WarehouseStock{}, "pending_quantity"); err != nil {
			fmt.Printf("[ERROR] AddColumn pending_quantity: %v\n", err)
		} else {
			fmt.Println("[INFO] AddColumn pending_quantity: success")
		}
	}
	// 添加 orders.stock_status 列（如果不存在）
	if !DB.Migrator().HasColumn(&models.Order{}, "stock_status") {
		if err := DB.Migrator().AddColumn(&models.Order{}, "stock_status"); err != nil {
			fmt.Printf("[ERROR] AddColumn stock_status: %v\n", err)
		} else {
			fmt.Println("[INFO] AddColumn stock_status: success")
		}
	}
	// 添加 purchase_orders.receipt_remark 列（如果不存在）
	if !DB.Migrator().HasColumn(&models.PurchaseOrder{}, "receipt_remark") {
		if err := DB.Migrator().AddColumn(&models.PurchaseOrder{}, "receipt_remark"); err != nil {
			fmt.Printf("[ERROR] AddColumn receipt_remark: %v\n", err)
		} else {
			fmt.Println("[INFO] AddColumn receipt_remark: success")
		}
	}
	// 修改 order_items.discount_rate 列类型为 DECIMAL(5,4)，并将旧数据从百分比转为小数
	migrateDiscountRate()

	// 添加 users.level 列（如果不存在）
	if !DB.Migrator().HasColumn(&models.User{}, "level") {
		if err := DB.Migrator().AddColumn(&models.User{}, "level"); err != nil {
			fmt.Printf("[ERROR] AddColumn users.level: %v\n", err)
		} else {
			fmt.Println("[INFO] AddColumn users.level: success")
		}
	}

	// 添加 system_configs.sort 列（如果不存在）
	if !DB.Migrator().HasColumn(&models.SystemConfig{}, "sort") {
		if err := DB.Migrator().AddColumn(&models.SystemConfig{}, "sort"); err != nil {
			fmt.Printf("[ERROR] AddColumn system_configs.sort: %v\n", err)
		} else {
			fmt.Println("[INFO] AddColumn system_configs.sort: success")
		}
	}

	// 添加 follow_up_approvals.approval_type 和 order_id 列（如果不存在）
	if !DB.Migrator().HasColumn(&models.FollowUpApproval{}, "approval_type") {
		if err := DB.Migrator().AddColumn(&models.FollowUpApproval{}, "approval_type"); err != nil {
			fmt.Printf("[ERROR] AddColumn follow_up_approvals.approval_type: %v\n", err)
		} else {
			fmt.Println("[INFO] AddColumn follow_up_approvals.approval_type: success")
		}
	}
	if !DB.Migrator().HasColumn(&models.FollowUpApproval{}, "order_id") {
		if err := DB.Migrator().AddColumn(&models.FollowUpApproval{}, "order_id"); err != nil {
			fmt.Printf("[ERROR] AddColumn follow_up_approvals.order_id: %v\n", err)
		} else {
			fmt.Println("[INFO] AddColumn follow_up_approvals.order_id: success")
		}
	}

	// 添加 roles.sort_order 列（如果不存在）
	if !DB.Migrator().HasColumn(&models.Role{}, "sort_order") {
		if err := DB.Migrator().AddColumn(&models.Role{}, "sort_order"); err != nil {
			fmt.Printf("[ERROR] AddColumn roles.sort_order: %v\n", err)
		} else {
			fmt.Println("[INFO] AddColumn roles.sort_order: success")
		}
	}

	// 添加 app_versions.update_type 列（如果不存在）
	if !DB.Migrator().HasColumn(&models.AppVersion{}, "update_type") {
		if err := DB.Migrator().AddColumn(&models.AppVersion{}, "update_type"); err != nil {
			fmt.Printf("[ERROR] AddColumn app_versions.update_type: %v\n", err)
		} else {
			fmt.Println("[INFO] AddColumn app_versions.update_type: success")
		}
	}

	// 添加 products.series 列（如果不存在）
	if !DB.Migrator().HasColumn(&models.Product{}, "series") {
		if err := DB.Migrator().AddColumn(&models.Product{}, "series"); err != nil {
			fmt.Printf("[ERROR] AddColumn products.series: %v\n", err)
		} else {
			fmt.Println("[INFO] AddColumn products.series: success")
		}
	}

	// 添加 products.sub_category 列（如果不存在）
	if !DB.Migrator().HasColumn(&models.Product{}, "sub_category") {
		if err := DB.Migrator().AddColumn(&models.Product{}, "sub_category"); err != nil {
			fmt.Printf("[ERROR] AddColumn products.sub_category: %v\n", err)
		} else {
			fmt.Println("[INFO] AddColumn products.sub_category: success")
		}
	}

	// 迁移旧 MD5 密码到 bcrypt（仅处理未升级的记录）
	migratePasswordsToBcrypt()

	// 初始化角色数据（INSERT IGNORE 幂等）
	seedRoles()

	// 初始化权限数据（INSERT IGNORE 幂等）
	seedPermissions()

	// 初始化提成比例配置数据（INSERT IGNORE 幂等）
	seedCommissionConfigs()
}

// autoMigrateNewTablesOnly 只创建不存在的表，不修改已有表
// GORM 的 AutoMigrate 会清除 MySQL 的 COMMENT，所以必须谨慎使用
func autoMigrateNewTablesOnly() {
	// 定义需要自动迁移的模型列表
	modelsToMigrate := []interface{}{
		&models.CustomerFollowUp{},
		&models.FollowUpApproval{},
		&models.AppVersion{},
		&models.StockQueue{},
		&models.SystemBackup{},
		&models.Role{},
		&models.Stocktake{},
		&models.StocktakeItem{},
		&models.OutboundRequest{},
	}

	for _, model := range modelsToMigrate {
		// 检查表是否已存在
		if !DB.Migrator().HasTable(model) {
			// 表不存在，创建表
			if err := DB.AutoMigrate(model); err != nil {
				fmt.Printf("[WARN] AutoMigrate failed for %T: %v\n", model, err)
			} else {
				fmt.Printf("[INFO] Created table for %T\n", model)
			}
		} else {
			// 表已存在，跳过（避免清除 COMMENT）
			fmt.Printf("[INFO] Table for %T already exists, skipping AutoMigrate\n", model)
		}
	}
}

// seedRoles 初始化角色数据到 roles 表
func seedRoles() {
	roles := []struct {
		Code string
		Name string
		Sort int
	}{
		{"BOSS", "老板", 1},
		{"ADMIN", "管理员", 2},
		{"STORE_MANAGER", "店长", 3},
		{"FINANCE", "财务", 4},
		{"SUPERVISOR", "主管", 5},
		{"SALESMAN", "业务员", 6},
		{"WAREHOUSE", "仓管", 7},
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	for _, r := range roles {
		DB.Exec(`INSERT IGNORE INTO roles (role_code, role_name, sort_order, status, created_at, updated_at) VALUES (?, ?, ?, 1, ?, ?)`,
			r.Code, r.Name, r.Sort, now, now)
	}
	fmt.Println("[INFO] 角色初始化完成")
}

// seedPermissions 初始化权限数据到 permissions 表
func seedPermissions() {
	// permission_type: 1=菜单 2=按钮 3=接口
	// 使用 INSERT IGNORE 避免重复插入
	perms := []struct {
		Code   string
		Name   string
		Type   int8
		Parent string
		Sort   int
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
		// ===== 送货管理 =====
		{"delivery:view", "查看送货", 3, "", 1251},
		{"delivery:create", "创建送货", 3, "", 1252},
		{"delivery:cancel", "作废送货", 3, "", 1253},
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
		// ===== APP版本管理 =====
		{"app_version:manage", "APP版本管理", 3, "", 1801},
		// ===== 平台维护 =====
		{"maintenance:manage", "平台维护", 3, "", 1901},
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
	// 检查 updated_at 列是否存在
	var count int64
	DB.Raw(`
		SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'payments' 
		AND COLUMN_NAME = 'updated_at'
	`).Scan(&count)

	// 如果不存在则添加
	if count == 0 {
		if err := DB.Exec("ALTER TABLE payments ADD COLUMN updated_at datetime(3) DEFAULT NULL").Error; err != nil {
			fmt.Printf("[ERROR] AddColumn payments.updated_at: %v\n", err)
		} else {
			fmt.Println("[INFO] AddColumn payments.updated_at: success")
		}
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

// seedCommissionConfigs 初始化提成比例配置数据
func seedCommissionConfigs() {
	configs := []struct {
		Key   string
		Value string
		Type  string
		Desc  string
		Sort  int
	}{
		// 等级提成比例 - 初级 (sort: 10-12)
		{"commission_rate_level1_single", "0.08", "commission", "初级业务员-单品提成比例", 10},
		{"commission_rate_level1_multi", "0.10", "commission", "初级业务员-多品提成比例", 11},
		{"commission_rate_level1_remark", "建议底薪3000-4000", "commission", "初级业务员-备注", 12},
		// 等级提成比例 - 中级 (sort: 20-22)
		{"commission_rate_level2_single", "0.18", "commission", "中级业务员-单品提成比例", 20},
		{"commission_rate_level2_multi", "0.22", "commission", "中级业务员-多品提成比例", 21},
		{"commission_rate_level2_remark", "建议底薪1500-2500", "commission", "中级业务员-备注", 22},
		// 等级提成比例 - 高级 (sort: 30-32)
		{"commission_rate_level3_single", "0.35", "commission", "高级业务员-单品提成比例", 30},
		{"commission_rate_level3_multi", "0.38", "commission", "高级业务员-多品提成比例", 31},
		{"commission_rate_level3_remark", "建议底薪0", "commission", "高级业务员-备注", 32},
		// 同行提成比例 (sort: 40-42)
		{"commission_rate_peer_single", "0.10", "commission", "同行单品提成比例", 40},
		{"commission_rate_peer_multi", "0.12", "commission", "同行多品提成比例", 41},
		{"commission_rate_peer_special", "0.08", "commission", "同行特批提成比例", 42},
		// 团队分润 (sort: 50-53)
		{"fund_pool_extract_rate", "0.05", "commission", "基金池提取比例", 50},
		{"team_share_rate_manager", "0.03", "commission", "主管团队分润比例", 51},
		{"team_share_rate_store", "0.02", "commission", "店长团队分润比例", 52},
		{"referral_reward_rate", "0.10", "commission", "老带新奖励比例", 53},
		// 固定提成比例 (sort: 60)
		{"fixed_commission_rate", "0.05", "commission", "固定提成比例（月度回款提成）", 60},
	}

	inserted := 0
	for _, c := range configs {
		result := DB.Exec(
			`INSERT IGNORE INTO system_configs (config_key, config_value, config_type, remark, sort, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, NOW(), NOW())`,
			c.Key, c.Value, c.Type, c.Desc, c.Sort,
		)
		if result.Error == nil && result.RowsAffected > 0 {
			inserted++
		}
	}

	fmt.Printf("[INFO] 提成配置数据初始化完成，新增 %d 项\n", inserted)
}

// migratePasswordsToBcrypt 将数据库中旧的 MD5 密码升级为 bcrypt 哈希
// MD5 哈希特征是 32 位十六进制字符串，bcrypt 以 "$2a$" 开头
func migratePasswordsToBcrypt() {
	var users []models.User
	// 查找所有密码不是 bcrypt 格式的用户（bcrypt 哈希始终以 "$2a$" 开头）
	if err := DB.Where("password NOT LIKE ?", "$2a$%").Find(&users).Error; err != nil {
		fmt.Printf("[ERROR] 查询待迁移密码用户失败: %v\n", err)
		return
	}

	migrated := 0
	skipped := 0
	for _, user := range users {
		// 跳过空密码
		if user.Password == "" {
			skipped++
			continue
		}

		// 对明文密码进行 bcrypt 哈希
		// 注意：如果数据库中存储的是明文密码（非哈希），这会将明文视为密码进行哈希
		// 如果存储的是 MD5 哈希（32位hex），系统将无法自动迁移（因为不知道原始密码）
		// 对于种子数据中的 MD5 哈希密码，需要重新生成
		hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Printf("[ERROR] 用户 %d 密码迁移失败: %v\n", user.ID, err)
			continue
		}

		if err := DB.Model(&models.User{}).Where("id = ?", user.ID).Update("password", string(hash)).Error; err != nil {
			fmt.Printf("[ERROR] 用户 %d 密码迁移更新失败: %v\n", user.ID, err)
			continue
		}
		migrated++
	}

	if migrated > 0 || skipped > 0 {
		fmt.Printf("[INFO] 密码迁移完成：已升级 %d 条，跳过 %d 条（空密码）\n", migrated, skipped)
	}
}
