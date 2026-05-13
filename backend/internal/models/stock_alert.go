package models

import "time"

// StockAlert 库存预警模型
type StockAlert struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	StoreID	int64	`gorm:"column:store_id;default:1" json:"store_id" example:"1"`
	WarehouseID	*int64	`gorm:"column:warehouse_id" json:"warehouse_id" example:"1"`
	AlertType	int8	`gorm:"column:alert_type" json:"alert_type" example:"1"`
	SKUID	*int64	`gorm:"column:sku_id" json:"sku_id" example:"1"`
	GiftID	*int64	`gorm:"column:gift_id" json:"gift_id" example:"1"`
	CurrentStock	int	`gorm:"column:current_stock;default:0" json:"current_stock" example:"5"`
	WarningStock	int	`gorm:"column:warning_stock;default:0" json:"warning_stock" example:"10"`
	AlertStatus	int8	`gorm:"column:alert_status;default:0" json:"alert_status" example:"0"`
	HandledBy	*int64	`gorm:"column:handled_by" json:"handled_by" example:"1"`
	HandledAt	*time.Time	`gorm:"column:handled_at" json:"handled_at" example:"2025-01-15T00:00:00+08:00"`
	Remark	string	`gorm:"column:remark;type:varchar(500)" json:"remark" example:"备注信息"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	Warehouse *Warehouse  `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	SKU       *ProductSKU `gorm:"foreignKey:SKUID" json:"sku,omitempty"`
	Gift      *Gift       `gorm:"foreignKey:GiftID" json:"gift,omitempty"`
}

// TableName 指定表名
func (StockAlert) TableName() string { return "stock_alerts" }
