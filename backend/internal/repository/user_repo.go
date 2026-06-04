package repository

import (
	"furniture-commission/internal/models"
	"furniture-commission/internal/pkg/pagination"

	"gorm.io/gorm"
)

// UserRepository 用户Repository
type UserRepository struct {
	*BaseRepository[models.User]
}

// NewUserRepository 创建用户Repository实例
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		BaseRepository: NewBaseRepository[models.User](db),
	}
}

// FindByUsername 根据用户名查找用户
func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByPhone 根据手机号查找用户
func (r *UserRepository) FindByPhone(phone string) (*models.User, error) {
	var user models.User
	err := r.DB.Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmployeeNo 根据工号查找用户
func (r *UserRepository) FindByEmployeeNo(employeeNo string) (*models.User, error) {
	var user models.User
	err := r.DB.Where("employee_no = ?", employeeNo).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindWithRoles 根据ID查找用户（包含角色和门店信息）
func (r *UserRepository) FindWithRoles(id int64) (*models.User, error) {
	var user models.User
	err := r.DB.Preload("Roles").Preload("Store").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ListWithFilter 根据条件分页查询用户列表
func (r *UserRepository) ListWithFilter(storeID, role, status int, keyword string, page, pageSize int) ([]models.User, int64, error) {
	db := r.DB.Model(&models.User{})

	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	if role > 0 {
		db = db.Where("id IN (SELECT user_id FROM user_roles WHERE role_id = ?)", role)
	}
	if status > 0 {
		db = db.Where("status = ?", status)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("username LIKE ? OR real_name LIKE ? OR phone LIKE ? OR employee_no LIKE ?",
			like, like, like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	p := &pagination.Pagination{Page: page, PageSize: pageSize}
	p.SetDefault()

	var users []models.User
	if err := db.Preload("Store").Preload("Roles").Preload("Referrer").
		Offset(p.GetOffset()).Limit(p.GetLimit()).
		Order("id DESC").
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
