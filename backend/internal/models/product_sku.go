package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ProductSKU 商品SKU模型
type ProductSKU struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	ProductID	int64	`gorm:"column:product_id;not null" json:"product_id" example:"1"`
	SKUCode	string	`gorm:"column:sku_code;type:varchar(50);uniqueIndex" json:"sku_code" example:"SKU001"`
	SKUName	string	`gorm:"column:sku_name;type:varchar(100)" json:"sku_name" example:"智能手机-黑色-128G"`
	Attributes	datatypes.JSON	`gorm:"column:attributes;type:json" json:"attributes" example:"{"color":"黑色","storage":"128G"}"`
	Barcode	string	`gorm:"column:barcode;type:varchar(50)" json:"barcode" example:"6901234567890"`
	Status	int8	`gorm:"column:status;default:1" json:"status" example:"1"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// TableName 指定表名
func (ProductSKU) TableName() string { return "product_skus" }
