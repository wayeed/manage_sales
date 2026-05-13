package models

import (
	"time"

	"gorm.io/gorm"
)

// Warehouse 仓库模型
type Warehouse struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	StoreID	int64	`gorm:"column:store_id;default:1" json:"store_id" example:"1"`
	WarehouseCode	string	`gorm:"column:warehouse_code;type:varchar(32);uniqueIndex" json:"warehouse_code" example:"WH001"`
	WarehouseName	string	`gorm:"column:warehouse_name;type:varchar(100)" json:"warehouse_name" example:"北京主仓库"`
	WarehouseType	int8	`gorm:"column:warehouse_type;default:1" json:"warehouse_type" example:"1"`
	Address	string	`gorm:"column:address;type:varchar(255)" json:"address" example:"北京市朝阳区建国路88号"`
	ContactPerson	string	`gorm:"column:contact_person;type:varchar(50)" json:"contact_person" example:"李四"`
	ContactPhone	string	`gorm:"column:contact_phone;type:varchar(20)" json:"contact_phone" example:"13900139000"`
	ManagerID	*int64	`gorm:"column:manager_id" json:"manager_id" example:"1"`
	Status	int8	`gorm:"column:status;default:1" json:"status" example:"1"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`
	DeletedAt     gorm.DeletedAt  `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Warehouse) TableName() string { return "warehouses" }
