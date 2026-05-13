package models

import "time"

// WarehouseGiftStock 仓库礼品库存模型
type WarehouseGiftStock struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	WarehouseID	int64	`gorm:"column:warehouse_id;not null;uniqueIndex:uk_warehouse_gift" json:"warehouse_id" example:"1"`
	GiftID	int64	`gorm:"column:gift_id;not null;uniqueIndex:uk_warehouse_gift" json:"gift_id" example:"1"`
	StockQuantity	int	`gorm:"column:stock_quantity;default:0" json:"stock_quantity" example:"100"`
	AvailableQuantity	int	`gorm:"column:available_quantity;default:0" json:"available_quantity" example:"80"`
	LockedQuantity	int	`gorm:"column:locked_quantity;default:0" json:"locked_quantity" example:"20"`
	WarningStock	int	`gorm:"column:warning_stock;default:10" json:"warning_stock" example:"10"`
	Version	int	`gorm:"column:version;default:0" json:"version" example:"1"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	Warehouse *Warehouse `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	Gift      *Gift      `gorm:"foreignKey:GiftID" json:"gift,omitempty"`
}

// TableName 指定表名
func (WarehouseGiftStock) TableName() string { return "warehouse_gift_stocks" }
