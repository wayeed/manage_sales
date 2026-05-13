package models

// RolePermission 角色权限关联模型
type RolePermission struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	RoleID	int64	`gorm:"column:role_id;uniqueIndex:idx_role_perm;not null" json:"role_id" example:"1"`
	PermissionID	int64	`gorm:"column:permission_id;uniqueIndex:idx_role_perm;not null" json:"permission_id" example:"1"`
}

// TableName 指定表名
func (RolePermission) TableName() string { return "role_permissions" }
