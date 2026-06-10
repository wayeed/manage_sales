package service

import (
	"testing"

	"furniture-commission/configs"
	"furniture-commission/internal/models"
	"furniture-commission/internal/pkg/database"
	"furniture-commission/internal/repository"
	svc "furniture-commission/internal/service"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 初始化测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// 初始化配置（JWT生成需要）
	configs.GlobalConfig = &configs.Config{
		JWT: configs.JWTConfig{
			Secret:      "test-secret-key",
			ExpireHours: 24,
		},
	}

	// 使用SQLite内存数据库进行测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 自动建表
	err = db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.UserRole{},
		&models.RolePermission{},
		&models.Store{},
	)
	assert.NoError(t, err)

	// 替换全局DB
	database.DB = db

	return db
}

// seedTestData 填充测试数据
func seedTestData(t *testing.T, db *gorm.DB) {
	t.Helper()

	// 创建测试角色
	roles := []models.Role{
		{ID: 1, RoleCode: "admin", RoleName: "管理员", Status: 1, SortOrder: 1},
		{ID: 2, RoleCode: "staff", RoleName: "普通员工", Status: 1, SortOrder: 2},
	}
	for _, role := range roles {
		db.Create(&role)
	}

	// 创建测试权限
	perms := []models.Permission{
		{ID: 1, ParentID: nil, PermissionName: "用户管理", PermissionCode: "user:manage", PermissionType: 1, SortOrder: 1, Status: 1},
		{ID: 2, ParentID: nil, PermissionName: "角色管理", PermissionCode: "role:manage", PermissionType: 1, SortOrder: 2, Status: 1},
	}
	for _, perm := range perms {
		db.Create(&perm)
	}

	// 创建测试用户（密码使用 bcrypt 哈希）
	bcryptPassword := "$2a$10$7xaz/QDlO1axD4kGdf6lre4PQ0hTV1/4lR3mc/vJBfcqZF0DhH2gq" // bcrypt("123456")
	users := []models.User{
		{
			ID:       1,
			Username: "admin",
			Password: bcryptPassword,
			RealName: "管理员",
			Phone:    "13800138000",
			Role:     5,
			Status:   1,
		},
		{
			ID:       2,
			Username: "disabled_user",
			Password: bcryptPassword,
			RealName: "已禁用用户",
			Phone:    "13800138001",
			Role:     1,
			Status:   0, // 已禁用
		},
		{
			ID:       3,
			Username: "normal_user",
			Password: bcryptPassword,
			RealName: "普通用户",
			Phone:    "13800138002",
			Role:     1,
			Status:   1,
		},
	}
	for _, user := range users {
		db.Create(&user)
	}

	// 分配角色给admin用户
	db.Create(&models.UserRole{UserID: 1, RoleID: 1})

	// 分配权限给admin角色
	db.Create(&models.RolePermission{RoleID: 1, PermissionID: 1})
	db.Create(&models.RolePermission{RoleID: 1, PermissionID: 2})
}

// TestLogin_Success 测试登录成功
func TestLogin_Success(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)

	userRepo := repository.NewUserRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	authService := svc.NewAuthService(userRepo, permRepo)

	// 测试使用用户名登录
	resp, err := authService.Login("admin", "123456", "127.0.0.1")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, int64(1), resp.User.ID)
	assert.Equal(t, "admin", resp.User.Username)
	assert.Equal(t, "管理员", resp.User.RealName)

	// 测试使用手机号登录
	resp2, err := authService.Login("13800138000", "123456", "127.0.0.1")
	assert.NoError(t, err)
	assert.NotNil(t, resp2)
	assert.Equal(t, int64(1), resp2.User.ID)
}

// TestLogin_WrongPassword 测试密码错误
func TestLogin_WrongPassword(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)

	userRepo := repository.NewUserRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	authService := svc.NewAuthService(userRepo, permRepo)

	resp, err := authService.Login("admin", "wrong_password", "127.0.0.1")
	assert.Error(t, err)
	assert.Nil(t, resp)

	assert.Contains(t, err.Error(), "密码错误")
}

// TestLogin_UserNotFound 测试用户不存在
func TestLogin_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)

	userRepo := repository.NewUserRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	authService := svc.NewAuthService(userRepo, permRepo)

	resp, err := authService.Login("nonexistent", "123456", "127.0.0.1")
	assert.Error(t, err)
	assert.Nil(t, resp)

	assert.Contains(t, err.Error(), "用户不存在")
}

// TestLogin_UserDisabled 测试用户禁用
func TestLogin_UserDisabled(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)

	userRepo := repository.NewUserRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	authService := svc.NewAuthService(userRepo, permRepo)

	resp, err := authService.Login("disabled_user", "123456", "127.0.0.1")
	assert.Error(t, err)
	assert.Nil(t, resp)

	// 禁用用户登录时，服务可能返回"用户不存在"或"用户已被禁用"
	// 取决于登录查询是否包含status过滤
	assert.Contains(t, err.Error(), "不存在")
}

// TestGetCurrentUser 测试获取当前用户信息
func TestGetCurrentUser(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)

	userRepo := repository.NewUserRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	authService := svc.NewAuthService(userRepo, permRepo)

	detail, err := authService.GetCurrentUser(1)
	assert.NoError(t, err)
	assert.NotNil(t, detail)
	assert.Equal(t, int64(1), detail.ID)
	assert.Equal(t, "admin", detail.Username)
	assert.Len(t, detail.Roles, 1)
	assert.Equal(t, "admin", detail.Roles[0].RoleCode)
	assert.Len(t, detail.Permissions, 2)
}

// TestGetCurrentUser_NotFound 测试获取不存在的用户
func TestGetCurrentUser_NotFound(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)

	userRepo := repository.NewUserRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	authService := svc.NewAuthService(userRepo, permRepo)

	detail, err := authService.GetCurrentUser(999)
	assert.Error(t, err)
	assert.Nil(t, detail)

	assert.Contains(t, err.Error(), "用户不存在")
}

// TestLogout 测试登出
func TestLogout(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)

	userRepo := repository.NewUserRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	authService := svc.NewAuthService(userRepo, permRepo)

	err := authService.Logout(1)
	assert.NoError(t, err)
}
