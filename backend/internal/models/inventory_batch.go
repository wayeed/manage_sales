package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// InventoryBatch 库存批次模型
type InventoryBatch struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	SKUID	int64	`gorm:"column:sku_id;not null" json:"sku_id" example:"1"`
	BatchNo	string	`gorm:"column:batch_no;type:varchar(32);uniqueIndex" json:"batch_no" example:"BATCH20250115001"`
	PurchaseOrderID	*int64	`gorm:"column:purchase_order_id" json:"purchase_order_id" example:"1"`
	PurchasePrice	decimal.Decimal	`gorm:"column:purchase_price;type:decimal(12,2);default:0.00" json:"purchase_price" example:"1500.00"`
	TotalCost	decimal.Decimal	`gorm:"column:total_cost;type:decimal(12,2);default:0.00" json:"total_cost" example:"3200.00"`
	InitialQuantity	int	`gorm:"column:initial_quantity;default:0" json:"initial_quantity" example:"100"`
	RemainingQuantity	int	`gorm:"column:remaining_quantity;default:0" json:"remaining_quantity" example:"60"`
	WarehouseID	*int64	`gorm:"column:warehouse_id" json:"warehouse_id" example:"1"`
	Status	int8	`gorm:"column:status;default:1" json:"status" example:"1"`
	EntryDate	*time.Time	`gorm:"column:entry_date" json:"entry_date" example:"2025-01-15T00:00:00+08:00"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`
}

// TableName 指定表名
func (InventoryBatch) TableName() string { return "inventory_batches" }
