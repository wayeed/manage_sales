package service_test

import (
	"testing"

	"furniture-commission/internal/models"
	"furniture-commission/internal/repository"
	svc "furniture-commission/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupRoleTestDB 创建角色测试数据库
func setupRoleTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(
		&models.Role{},
		&models.Permission{},
		&models.RolePermission{},
		&models.User{},
		&models.UserRole{},
	)
	assert.NoError(t, err)
	return db
}

// ========== TestCreateRole ==========

func TestCreateRole(t *testing.T) {
	db := setupRoleTestDB(t)

	roleRepo := repository.NewRoleRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	roleSvc := svc.NewRoleService(db, roleRepo, permRepo)

	req := &svc.CreateRoleRequest{
		RoleCode: "manager",
		RoleName: "经理",
		Description: "门店经理角色",
		SortOrder: 1,
	}

	err := roleSvc.Create(req)
	assert.NoError(t, err)

	// 验证角色已创建
	role, err := roleRepo.FindByCode("manager")
	assert.NoError(t, err)
	assert.NotNil(t, role)
	assert.Equal(t, "经理", role.RoleName)
	assert.Equal(t, int8(1), role.Status)
}

func TestCreateRole_DuplicateCode(t *testing.T) {
	db := setupRoleTestDB(t)

	roleRepo := repository.NewRoleRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	roleSvc := svc.NewRoleService(db, roleRepo, permRepo)

	req1 := &svc.CreateRoleRequest{
		RoleCode: "admin",
		RoleName: "管理员",
	}
	err := roleSvc.Create(req1)
	assert.NoError(t, err)

	req2 := &svc.CreateRoleRequest{
		RoleCode: "admin",
		RoleName: "管理员2",
	}
	err = roleSvc.Create(req2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "角色编码已存在")
}

// ========== TestUpdateRole ==========

func TestUpdateRole(t *testing.T) {
	db := setupRoleTestDB(t)

	roleRepo := repository.NewRoleRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	roleSvc := svc.NewRoleService(db, roleRepo, permRepo)

	// 创建角色
	createReq := &svc.CreateRoleRequest{
		RoleCode: "updaterole",
		RoleName: "原名",
	}
	err := roleSvc.Create(createReq)
	assert.NoError(t, err)

	role, _ := roleRepo.FindByCode("updaterole")

	// 更新角色
	status := int8(0)
	updateReq := &svc.UpdateRoleRequest{
		RoleName: "新名称",
		Status:   &status,
		SortOrder: intPtr(5),
	}
	err = roleSvc.Update(role.ID, updateReq)
	assert.NoError(t, err)

	updated, _ := roleRepo.FindByID(role.ID)
	assert.Equal(t, "新名称", updated.RoleName)
	assert.Equal(t, int8(0), updated.Status)
	assert.Equal(t, 5, updated.SortOrder)
}

func TestUpdateRole_NotFound(t *testing.T) {
	db := setupRoleTestDB(t)

	roleRepo := repository.NewRoleRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	roleSvc := svc.NewRoleService(db, roleRepo, permRepo)

	updateReq := &svc.UpdateRoleRequest{
		RoleName: "不存在",
	}
	err := roleSvc.Update(99999, updateReq)
	assert.Error(t, err)
}

// ========== TestAssignPermissions ==========

func TestAssignPermissions(t *testing.T) {
	db := setupRoleTestDB(t)

	roleRepo := repository.NewRoleRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	roleSvc := svc.NewRoleService(db, roleRepo, permRepo)

	// 创建角色
	role := &models.Role{
		RoleCode: "testrole",
		RoleName: "测试角色",
		Status:   1,
	}
	db.Create(role)

	// 创建权限
	perm1 := &models.Permission{
		PermissionName: "用户管理",
		PermissionCode: "user:manage",
		PermissionType: 1,
		Status:         1,
	}
	perm2 := &models.Permission{
		PermissionName: "角色管理",
		PermissionCode: "role:manage",
		PermissionType: 1,
		Status:         1,
	}
	db.Create(perm1)
	db.Create(perm2)

	// 分配权限
	err := roleSvc.AssignPermissions(role.ID, []int64{perm1.ID, perm2.ID})
	assert.NoError(t, err)

	// 验证权限已分配
	roleDetail, err := roleSvc.GetDetail(role.ID)
	assert.NoError(t, err)
	assert.Len(t, roleDetail.Permissions, 2)
}

func TestAssignPermissions_NotFound(t *testing.T) {
	db := setupRoleTestDB(t)

	roleRepo := repository.NewRoleRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	roleSvc := svc.NewRoleService(db, roleRepo, permRepo)

	err := roleSvc.AssignPermissions(99999, []int64{1, 2})
	assert.Error(t, err)
}

// ========== TestListRoles ==========

type RoleListTestSuite struct {
	suite.Suite
	db       *gorm.DB
	roleSvc  *svc.RoleService
	roleRepo *repository.RoleRepository
	permRepo *repository.PermissionRepository
}

func (s *RoleListTestSuite) SetupSuite() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	s.Require().NoError(err)
	s.db = db

	err = db.AutoMigrate(
		&models.Role{},
		&models.Permission{},
		&models.RolePermission{},
		&models.User{},
		&models.UserRole{},
	)
	s.Require().NoError(err)

	s.roleRepo = repository.NewRoleRepository(db)
	s.permRepo = repository.NewPermissionRepository(db)
	s.roleSvc = svc.NewRoleService(db, s.roleRepo, s.permRepo)
}

func (s *RoleListTestSuite) TearDownSuite() {
	sqlDB, _ := s.db.DB()
	sqlDB.Close()
}

func (s *RoleListTestSuite) SetupTest() {
	s.db.Exec("DELETE FROM roles")
}

func (s *RoleListTestSuite) TestListRoles() {
	// 创建多个角色
	roles := []svc.CreateRoleRequest{
		{RoleCode: "admin", RoleName: "管理员"},
		{RoleCode: "staff", RoleName: "普通员工"},
		{RoleCode: "manager", RoleName: "经理"},
	}
	for _, r := range roles {
		err := s.roleSvc.Create(&r)
		s.Require().NoError(err)
	}

	result, err := s.roleSvc.List()
	s.NoError(err)
	s.Len(result, 3)

	// 验证返回的是VO对象
	for _, r := range result {
		s.NotZero(r.ID)
		s.NotEmpty(r.RoleCode)
		s.NotEmpty(r.RoleName)
	}
}

func (s *RoleListTestSuite) TestListRoles_Empty() {
	result, err := s.roleSvc.List()
	s.NoError(err)
	s.Empty(result)
}

func TestRoleListSuite(t *testing.T) {
	suite.Run(t, new(RoleListTestSuite))
}

// intPtr 辅助函数
func intPtr(v int) *int {
	return &v
}
