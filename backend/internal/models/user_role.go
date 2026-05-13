package models

// UserRole 用户角色关联模型
type UserRole struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	UserID	int64	`gorm:"column:user_id;uniqueIndex:idx_user_role;not null" json:"user_id" example:"1"`
	RoleID	int64	`gorm:"column:role_id;uniqueIndex:idx_user_role;not null" json:"role_id" example:"1"`
}

// TableName 指定表名
func (UserRole) TableName() string { return "user_roles" }
