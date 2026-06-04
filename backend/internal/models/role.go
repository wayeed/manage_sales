package models

import (
	"time"

	"gorm.io/gorm"
)

// Role 角色模型
type Role struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	RoleCode	string	`gorm:"column:role_code;type:varchar(50);uniqueIndex;not null" json:"role_code" example:"admin"`
	RoleName	string	`gorm:"column:role_name;type:varchar(50);not null" json:"role_name" example:"管理员"`
	RoleType	int8	`gorm:"column:role_type;default:1" json:"role_type" example:"1"`
	SortOrder	int	`gorm:"column:sort_order;default:0" json:"sort_order" example:"1"`
	Description	string	`gorm:"column:description;type:varchar(255)" json:"description" example:"高端智能手机，性能卓越"`
	Status	int8	`gorm:"column:status;default:1;not null" json:"status" example:"1"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"-"`

	// 关联
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	Users       []User       `gorm:"many2many:user_roles;" json:"users,omitempty"`
}

// TableName 指定表名
func (Role) TableName() string { return "roles" }
