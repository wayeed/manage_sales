package models

import (
	"time"

	"gorm.io/gorm"
)

// Category 品类模型
type Category struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	StoreID	int64	`gorm:"column:store_id;default:1" json:"store_id" example:"1"`
	CategoryCode	string	`gorm:"column:category_code;type:varchar(32);uniqueIndex" json:"category_code" example:"CAT001"`
	CategoryName	string	`gorm:"column:category_name;type:varchar(50)" json:"category_name" example:"电子产品"`
	SortOrder	int	`gorm:"column:sort_order;default:0" json:"sort_order" example:"1"`
	Status	int8	`gorm:"column:status;default:1" json:"status" example:"1"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`
	DeletedAt    gorm.DeletedAt  `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Category) TableName() string { return "categories" }
