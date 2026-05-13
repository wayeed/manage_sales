package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// GiftInventoryBatch 礼品库存批次模型
type GiftInventoryBatch struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	GiftID	int64	`gorm:"column:gift_id;not null" json:"gift_id" example:"1"`
	BatchNo	string	`gorm:"column:batch_no;type:varchar(32);uniqueIndex" json:"batch_no" example:"BATCH20250115001"`
	PurchasePrice	decimal.Decimal	`gorm:"column:purchase_price;type:decimal(10,2);default:0.00" json:"purchase_price" example:"1500.00"`
	InitialQuantity	int	`gorm:"column:initial_quantity;default:0" json:"initial_quantity" example:"100"`
	RemainingQuantity	int	`gorm:"column:remaining_quantity;default:0" json:"remaining_quantity" example:"60"`
	WarehouseID	*int64	`gorm:"column:warehouse_id" json:"warehouse_id" example:"1"`
	Status	int8	`gorm:"column:status;default:1" json:"status" example:"1"`
	EntryDate	*time.Time	`gorm:"column:entry_date" json:"entry_date" example:"2025-01-15T00:00:00+08:00"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`
}

// TableName 指定表名
func (GiftInventoryBatch) TableName() string { return "gift_inventory_batches" }
