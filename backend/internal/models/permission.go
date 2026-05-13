package models

import "time"

// Permission 权限模型
type Permission struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	PermissionCode	string	`gorm:"column:permission_code;type:varchar(50);uniqueIndex" json:"permission_code" example:"system:user:list"`
	PermissionName	string	`gorm:"column:permission_name;type:varchar(50)" json:"permission_name" example:"用户列表"`
	PermissionType	int8	`gorm:"column:permission_type;default:1" json:"permission_type" example:"1"` // 1=菜单 2=按钮 3=接口
	ParentID	*int64	`gorm:"column:parent_id" json:"parent_id" example:"1"`
	Path	string	`gorm:"column:path;type:varchar(100)" json:"path" example:"/system/user"`
	Icon	string	`gorm:"column:icon;type:varchar(50)" json:"icon" example:"user"`
	SortOrder	int	`gorm:"column:sort_order;default:0" json:"sort_order" example:"1"`
	Status	int8	`gorm:"column:status;default:1" json:"status" example:"1"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	Children []Permission `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

// TableName 指定表名
func (Permission) TableName() string { return "permissions" }
