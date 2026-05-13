package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// OrderItem 订单商品明细模型
type OrderItem struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	OrderID	int64	`gorm:"column:order_id;not null" json:"order_id" example:"1"`
	SKUID	int64	`gorm:"column:sku_id;default:0" json:"sku_id" example:"1"`
	ProductName	string	`gorm:"column:product_name;type:varchar(100)" json:"product_name" example:"智能手机"`
	SKUName	string	`gorm:"column:sku_name;type:varchar(100)" json:"sku_name" example:"智能手机-黑色-128G"`
	CategoryID	*int64	`gorm:"column:category_id" json:"category_id" example:"1"`
	Quantity	int	`gorm:"column:quantity;default:0" json:"quantity" example:"10"`
	ListPrice	decimal.Decimal	`gorm:"column:list_price;type:decimal(12,2);default:0.00" json:"list_price" example:"2999.00"`
	SalePrice	decimal.Decimal	`gorm:"column:sale_price;type:decimal(12,2);default:0.00" json:"sale_price" example:"2499.00"`
	DiscountRate	decimal.Decimal	`gorm:"column:discount_rate;type:decimal(5,4);default:1.0000" json:"discount_rate" example:"0.8333"`
	BatchID	*int64	`gorm:"column:batch_id" json:"batch_id" example:"1"`
	UnitCost	decimal.Decimal	`gorm:"column:unit_cost;type:decimal(12,2);default:0.00" json:"unit_cost" example:"150.00"`
	TotalCost	decimal.Decimal	`gorm:"column:total_cost;type:decimal(12,2);default:0.00" json:"total_cost" example:"3200.00"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	Order *Order      `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	SKU   *ProductSKU `gorm:"foreignKey:SKUID" json:"sku,omitempty"`
}

// TableName 指定表名
func (OrderItem) TableName() string { return "order_items" }
