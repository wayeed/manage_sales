package service_test

import (
	"fmt"
	"testing"

	"furniture-commission/internal/models"
	appmd5 "furniture-commission/internal/pkg/md5"
	"furniture-commission/internal/repository"
	svc "furniture-commission/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupUserTestDB 创建用户测试数据库
func setupUserTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.UserRole{},
		&models.RolePermission{},
		&models.Store{},
	)
	assert.NoError(t, err)
	return db
}

// ========== TestCreateUser ==========

func TestCreateUser(t *testing.T) {
	db := setupUserTestDB(t)

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userSvc := svc.NewUserService(db, userRepo, roleRepo)

	req := &svc.CreateUserRequest{
		Username:    "testuser",
		Password:    "123456",
		RealName:    "测试用户",
		Phone:       "13800138000",
		Role:        1,
		BaseSalary:  5000,
	}

	err := userSvc.Create(req, 1)
	assert.NoError(t, err)

	// 验证用户已创建
	user, err := userRepo.FindByUsername("testuser")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "测试用户", user.RealName)
	assert.Equal(t, "13800138000", user.Phone)
	assert.Equal(t, int8(1), user.Status)
	assert.Equal(t, 5000.0, user.BaseSalary)

	// 验证密码已加密
	assert.True(t, appmd5.CheckPassword("123456", user.Password))
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	db := setupUserTestDB(t)

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userSvc := svc.NewUserService(db, userRepo, roleRepo)

	req1 := &svc.CreateUserRequest{
		Username: "dupuser",
		Password: "123456",
		RealName: "用户1",
	}
	err := userSvc.Create(req1, 1)
	assert.NoError(t, err)

	req2 := &svc.CreateUserRequest{
		Username: "dupuser",
		Password: "654321",
		RealName: "用户2",
	}
	err = userSvc.Create(req2, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "用户名已存在")
}

func TestCreateUser_DuplicatePhone(t *testing.T) {
	db := setupUserTestDB(t)

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userSvc := svc.NewUserService(db, userRepo, roleRepo)

	req1 := &svc.CreateUserRequest{
		Username: "user1",
		Password: "123456",
		Phone:    "13800138000",
	}
	err := userSvc.Create(req1, 1)
	assert.NoError(t, err)

	req2 := &svc.CreateUserRequest{
		Username: "user2",
		Password: "123456",
		Phone:    "13800138000",
	}
	err = userSvc.Create(req2, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "手机号已存在")
}

// ========== TestUpdateUser ==========

func TestUpdateUser(t *testing.T) {
	db := setupUserTestDB(t)

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userSvc := svc.NewUserService(db, userRepo, roleRepo)

	createReq := &svc.CreateUserRequest{
		Username:    "updateuser",
		Password:    "123456",
		RealName:    "原名",
		Phone:       "13800138000",
		BaseSalary:  3000,
	}
	err := userSvc.Create(createReq, 1)
	assert.NoError(t, err)

	user, _ := userRepo.FindByUsername("updateuser")

	updateReq := &svc.UpdateUserRequest{
		RealName:    "新名",
		Phone:       "13900139000",
		BaseSalary:  6000,
		Status:      1,
	}
	err = userSvc.Update(user.ID, updateReq)
	assert.NoError(t, err)

	updated, _ := userRepo.FindByID(user.ID)
	assert.Equal(t, "新名", updated.RealName)
	assert.Equal(t, "13900139000", updated.Phone)
	assert.Equal(t, 6000.0, updated.BaseSalary)
}

func TestUpdateUser_NotFound(t *testing.T) {
	db := setupUserTestDB(t)

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userSvc := svc.NewUserService(db, userRepo, roleRepo)

	updateReq := &svc.UpdateUserRequest{
		RealName: "不存在",
	}
	err := userSvc.Update(99999, updateReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "用户不存在")
}

// ========== TestDeleteUser ==========

func TestDeleteUser(t *testing.T) {
	db := setupUserTestDB(t)

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userSvc := svc.NewUserService(db, userRepo, roleRepo)

	createReq := &svc.CreateUserRequest{
		Username: "deleteuser",
		Password: "123456",
		RealName: "待删除",
	}
	err := userSvc.Create(createReq, 1)
	assert.NoError(t, err)

	user, _ := userRepo.FindByUsername("deleteuser")

	err = userSvc.Delete(user.ID)
	assert.NoError(t, err)

	deleted, _ := userRepo.FindByID(user.ID)
	assert.Equal(t, int8(2), deleted.Status)
}

func TestDeleteUser_NotFound(t *testing.T) {
	db := setupUserTestDB(t)

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userSvc := svc.NewUserService(db, userRepo, roleRepo)

	err := userSvc.Delete(99999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "用户不存在")
}

// ========== TestResetPassword ==========

func TestResetPassword(t *testing.T) {
	db := setupUserTestDB(t)

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userSvc := svc.NewUserService(db, userRepo, roleRepo)

	createReq := &svc.CreateUserRequest{
		Username: "resetuser",
		Password: "oldpassword",
		RealName: "重置密码",
	}
	err := userSvc.Create(createReq, 1)
	assert.NoError(t, err)

	user, _ := userRepo.FindByUsername("resetuser")

	newPassword, err := userSvc.ResetPassword(user.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, newPassword)
	assert.Len(t, newPassword, 8)

	updated, _ := userRepo.FindByID(user.ID)
	assert.True(t, appmd5.CheckPassword(newPassword, updated.Password))
	assert.False(t, appmd5.CheckPassword("oldpassword", updated.Password))
}

func TestResetPassword_NotFound(t *testing.T) {
	db := setupUserTestDB(t)

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userSvc := svc.NewUserService(db, userRepo, roleRepo)

	_, err := userSvc.ResetPassword(99999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "用户不存在")
}

// ========== TestListUsers ==========

type UserListTestSuite struct {
	suite.Suite
	db       *gorm.DB
	userSvc  *svc.UserService
	userRepo *repository.UserRepository
	roleRepo *repository.RoleRepository
}

func (s *UserListTestSuite) SetupSuite() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	s.Require().NoError(err)
	s.db = db

	err = db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.UserRole{},
		&models.RolePermission{},
		&models.Store{},
	)
	s.Require().NoError(err)

	s.userRepo = repository.NewUserRepository(db)
	s.roleRepo = repository.NewRoleRepository(db)
	s.userSvc = svc.NewUserService(db, s.userRepo, s.roleRepo)
}

func (s *UserListTestSuite) TearDownSuite() {
	sqlDB, _ := s.db.DB()
	sqlDB.Close()
}

func (s *UserListTestSuite) SetupTest() {
	s.db.Exec("DELETE FROM users")
}

func (s *UserListTestSuite) TestListUsers() {
	for i := 1; i <= 5; i++ {
		req := &svc.CreateUserRequest{
			Username:    fmt.Sprintf("listuser%d", i),
			Password:    "123456",
			RealName:    fmt.Sprintf("列表用户%d", i),
			Phone:       fmt.Sprintf("138%08d", i),
			Role:        1,
			BaseSalary:  3000,
			EmployeeNo:  fmt.Sprintf("EMP-LIST-%d", i),
		}
		err := s.userSvc.Create(req, 1)
		s.Require().NoError(err)
	}

	listReq := &svc.ListUserRequest{
		Page:     1,
		PageSize: 10,
	}
	result, err := s.userSvc.List(listReq)
	s.NoError(err)
	s.NotNil(result)
	s.Equal(int64(5), result.Total)
	s.Len(result.List, 5)
}

func (s *UserListTestSuite) TestListUsers_WithKeyword() {
	req := &svc.CreateUserRequest{
		Username:   "searchuser",
		Password:   "123456",
		RealName:   "搜索目标用户",
		Phone:      "13800138000",
		EmployeeNo: "EMP-SEARCH",
	}
	err := s.userSvc.Create(req, 1)
	s.Require().NoError(err)

	req2 := &svc.CreateUserRequest{
		Username:   "otheruser",
		Password:   "123456",
		RealName:   "其他用户",
		Phone:      "13800138001",
		EmployeeNo: "EMP-OTHER",
	}
	err = s.userSvc.Create(req2, 1)
	s.Require().NoError(err)

	listReq := &svc.ListUserRequest{
		Keyword:  "搜索",
		Page:     1,
		PageSize: 10,
	}
	result, err := s.userSvc.List(listReq)
	s.NoError(err)
	s.Equal(int64(1), result.Total)
}

func (s *UserListTestSuite) TestListUsers_WithStatus() {
	req1 := &svc.CreateUserRequest{
		Username:   "activeuser",
		Password:   "123456",
		RealName:   "启用用户",
		Phone:      "13800138001",
		EmployeeNo: "EMP-ACT",
	}
	err := s.userSvc.Create(req1, 1)
	s.Require().NoError(err)

	req2 := &svc.CreateUserRequest{
		Username:   "inactiveuser",
		Password:   "123456",
		RealName:   "禁用用户",
		Phone:      "13800138002",
		EmployeeNo: "EMP-INACT",
	}
	err = s.userSvc.Create(req2, 1)
	s.Require().NoError(err)

	// 手动禁用第二个用户
	inactiveUser, _ := s.userRepo.FindByUsername("inactiveuser")
	s.Require().NotNil(inactiveUser)
	s.db.Model(&models.User{}).Where("id = ?", inactiveUser.ID).Update("status", 0)

	listReq := &svc.ListUserRequest{
		Status:   1,
		Page:     1,
		PageSize: 10,
	}
	result, err := s.userSvc.List(listReq)
	s.NoError(err)
	s.Equal(int64(1), result.Total)
}

func TestUserListSuite(t *testing.T) {
	suite.Run(t, new(UserListTestSuite))
}
