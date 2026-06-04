package models

import "time"

// WarehouseStock 仓库商品库存模型
type WarehouseStock struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	WarehouseID	int64	`gorm:"column:warehouse_id;not null;uniqueIndex:uk_warehouse_sku" json:"warehouse_id" example:"1"`
	SKUID	int64	`gorm:"column:sku_id;not null;uniqueIndex:uk_warehouse_sku" json:"sku_id" example:"1"`
	StockQuantity	int	`gorm:"column:stock_quantity;default:0" json:"stock_quantity" example:"100"`
	AvailableQuantity	int	`gorm:"column:available_quantity;default:0" json:"available_quantity" example:"80"`
	LockedQuantity	int	`gorm:"column:locked_quantity;default:0" json:"locked_quantity" example:"20"`
	InTransitQuantity	int	`gorm:"column:in_transit_quantity;default:0" json:"in_transit_quantity" example:"0"` // 在途采购库存
	PendingQuantity	int	`gorm:"column:pending_quantity;default:0" json:"pending_quantity" example:"0"` // 待分配库存（入库后尚未匹配缺货订单）
	WarningStock	int	`gorm:"column:warning_stock;default:10" json:"warning_stock" example:"10"`
	Version	int	`gorm:"column:version;default:0" json:"version" example:"1"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	Warehouse *Warehouse  `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	SKU       *ProductSKU `gorm:"foreignKey:SKUID" json:"sku,omitempty"`
}

// TableName 指定表名
func (WarehouseStock) TableName() string { return "warehouse_stocks" }
