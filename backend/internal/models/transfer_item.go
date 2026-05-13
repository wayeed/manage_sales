package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// TransferItem 调拨明细模型
type TransferItem struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	TransferOrderID	int64	`gorm:"column:transfer_order_id;not null" json:"transfer_order_id" example:"1"`
	SKUID	*int64	`gorm:"column:sku_id" json:"sku_id" example:"1"`
	Quantity	int	`gorm:"column:quantity;default:0" json:"quantity" example:"10"`
	UnitCost	decimal.Decimal	`gorm:"column:unit_cost;type:decimal(12,2);default:0.00" json:"unit_cost" example:"150.00"`
	Subtotal	decimal.Decimal	`gorm:"column:subtotal;type:decimal(12,2);default:0.00" json:"subtotal" example:"15000.00"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	TransferOrder *TransferOrder `gorm:"foreignKey:TransferOrderID" json:"transfer_order,omitempty"`
	SKU           *ProductSKU    `gorm:"foreignKey:SKUID" json:"sku,omitempty"`
}

// TableName 指定表名
func (TransferItem) TableName() string { return "transfer_items" }
