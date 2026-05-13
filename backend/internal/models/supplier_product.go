package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// SupplierProduct 供应商商品关联模型
type SupplierProduct struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	SupplierID	int64	`gorm:"column:supplier_id;not null;uniqueIndex:uk_supplier_sku" json:"supplier_id" example:"1"`
	SKUID	int64	`gorm:"column:sku_id;not null;uniqueIndex:uk_supplier_sku" json:"sku_id" example:"1"`
	SupplyPrice	decimal.Decimal	`gorm:"column:supply_price;type:decimal(12,2);default:0.00" json:"supply_price" example:"1500.00"`
	MinOrderQuantity	int	`gorm:"column:min_order_quantity;default:1" json:"min_order_quantity" example:"10"`
	LeadTime	*int	`gorm:"column:lead_time" json:"lead_time" example:"7"`
	IsDefault	int8	`gorm:"column:is_default;default:0" json:"is_default" example:"1"`
	Status	int8	`gorm:"column:status;default:1" json:"status" example:"1"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	Supplier *Supplier   `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	SKU      *ProductSKU `gorm:"foreignKey:SKUID" json:"sku,omitempty"`
}

// TableName 指定表名
func (SupplierProduct) TableName() string { return "supplier_products" }
