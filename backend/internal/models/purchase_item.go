package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// PurchaseItem 采购商品明细模型
type PurchaseItem struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	PurchaseOrderID	int64	`gorm:"column:purchase_order_id;not null" json:"purchase_order_id" example:"1"`
	SKUID	*int64	`gorm:"column:sku_id" json:"sku_id" example:"1"`
	ProductName	string	`gorm:"column:product_name;type:varchar(100)" json:"product_name" example:"智能手机"`
	SKUName	string	`gorm:"column:sku_name;type:varchar(100)" json:"sku_name" example:"智能手机-黑色-128G"`
	SKUCode	string	`gorm:"column:sku_code;type:varchar(100)" json:"sku_code" example:"SKU001"`
	BrandStyle	string	`gorm:"column:brand_style;type:varchar(200)" json:"brand_style" example:"品牌-款式"`
	PurchasePrice	decimal.Decimal	`gorm:"column:purchase_price;type:decimal(12,2);default:0.00" json:"purchase_price" example:"1500.00"`
	Quantity	int	`gorm:"column:quantity;default:0" json:"quantity" example:"10"`
	ReceivedQuantity int `gorm:"column:received_quantity;default:0" json:"received_quantity" example:"0"`
	Subtotal	decimal.Decimal	`gorm:"column:subtotal;type:decimal(12,2);default:0.00" json:"subtotal" example:"15000.00"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	PurchaseOrder *PurchaseOrder `gorm:"foreignKey:PurchaseOrderID" json:"purchase_order,omitempty"`
	SKU           *ProductSKU    `gorm:"foreignKey:SKUID" json:"sku,omitempty"`
}

// TableName 指定表名
func (PurchaseItem) TableName() string { return "purchase_items" }
