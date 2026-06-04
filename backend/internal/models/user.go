package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// User 用户模型
type User struct {
	ID               int64          `gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	StoreID          *int64         `gorm:"column:store_id" json:"store_id" example:"1"`
	EmployeeNo       string         `gorm:"column:employee_no;type:varchar(32);uniqueIndex" json:"employee_no" example:"EMP001"`
	Username         string         `gorm:"column:username;type:varchar(64);uniqueIndex;not null" json:"username" example:"admin"`
	Password         string         `gorm:"column:password;type:varchar(128);not null" json:"-"`
	RealName         string         `gorm:"column:real_name;type:varchar(64)" json:"real_name" example:"张三"`
	Phone            string         `gorm:"column:phone;type:varchar(20);uniqueIndex" json:"phone" example:"13800138000"`
	Email            string         `gorm:"column:email;type:varchar(128)" json:"email" example:"admin@example.com"`
	DepartmentID     *int64         `gorm:"column:department_id" json:"department_id" example:"1"`
	Role             int8           `gorm:"column:role;default:0;not null" json:"role" example:"1"`
	Status           int8           `gorm:"column:status;default:1;not null" json:"status" example:"1"`
	EntryDate        *time.Time     `gorm:"column:entry_date" json:"entry_date" example:"2025-01-15T00:00:00+08:00"`
	ProbationEndDate *time.Time     `gorm:"column:probation_end_date" json:"probation_end_date" example:"2025-04-15T00:00:00+08:00"`
	IsFormal         int8           `gorm:"column:is_formal;default:0" json:"is_formal" example:"1"`
	Level            int8           `gorm:"column:level;default:1" json:"level" example:"1"` // 1-初级, 2-中级, 3-高级
	ParentID         *int64         `gorm:"column:parent_id" json:"parent_id" example:"1"`
	ReferrerID       *int64         `gorm:"column:referrer_id" json:"referrer_id" example:"1"`
	BaseSalary       decimal.Decimal `gorm:"column:base_salary;type:decimal(10,2);default:0" json:"base_salary" example:"5000.00"`
	IDCard           string         `gorm:"column:id_card;type:varchar(18)" json:"id_card" example:"110101199001011234"`
	BankAccount      string         `gorm:"column:bank_account;type:varchar(32)" json:"bank_account" example:"6222021234567890123"`
	BankName         string         `gorm:"column:bank_name;type:varchar(64)" json:"bank_name" example:"中国工商银行"`
	Avatar           string         `gorm:"column:avatar;type:varchar(255)" json:"avatar" example:"https://example.com/avatar/default.png"`
	LastLoginAt      *time.Time     `gorm:"column:last_login_at" json:"last_login_at" example:"2025-01-15T10:30:00+08:00"`
	LastLoginIP      string         `gorm:"column:last_login_ip;type:varchar(45)" json:"last_login_ip" example:"192.168.1.100"`
	CreatedBy        *int64         `gorm:"column:created_by" json:"created_by" example:"1"`
	CreatedAt        time.Time      `json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt        time.Time      `json:"updated_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	Roles    []Role `gorm:"many2many:user_roles;" json:"roles,omitempty"`
	Store    *Store `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	Parent   *User  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Referrer *User  `gorm:"foreignKey:ReferrerID" json:"referrer,omitempty"`
}

// TableName 指定表名
func (User) TableName() string { return "users" }
