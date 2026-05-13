package models

import (
	"time"

	"gorm.io/gorm"
)

// Store 门店模型
type Store struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	StoreCode	string	`gorm:"column:store_code;type:varchar(32);uniqueIndex;not null" json:"store_code" example:"STORE001"`
	StoreName	string	`gorm:"column:store_name;type:varchar(100);not null" json:"store_name" example:"北京旗舰店"`
	Address	string	`gorm:"column:address;type:varchar(255)" json:"address" example:"北京市朝阳区建国路88号"`
	ContactPhone	string	`gorm:"column:contact_phone;type:varchar(20)" json:"contact_phone" example:"13900139000"`
	ManagerID	*int64	`gorm:"column:manager_id" json:"manager_id" example:"1"`
	Status	int8	`gorm:"column:status;default:1" json:"status" example:"1"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`
	DeletedAt    gorm.DeletedAt  `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Store) TableName() string { return "stores" }
