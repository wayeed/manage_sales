package models

import "time"

// StockQueue 缺货排队模型
// 当订单审核通过但库存不足时，订单商品进入排队队列
// 采购入库后按订单先后顺序自动分配库存
type StockQueue struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	OrderID      int64     `gorm:"column:order_id;not null;index:idx_sku_created" json:"order_id" example:"1"`
	OrderItemID  int64     `gorm:"column:order_item_id;not null" json:"order_item_id" example:"1"`
	SKUID        int64     `gorm:"column:sku_id;not null;index:idx_sku_created" json:"sku_id" example:"1"`
	Quantity     int       `gorm:"column:quantity;not null" json:"quantity" example:"5"`
	AllocatedQty int       `gorm:"column:allocated_qty;default:0" json:"allocated_qty" example:"0"` // 已分配数量
	Status       int8      `gorm:"column:status;default:0" json:"status" example:"0"` // 0=排队中 1=部分分配 2=全部分配
	CreatedAt    time.Time `json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt    time.Time `json:"updated_at" example:"2025-01-15T00:00:00+08:00"`
}

// TableName 指定表名
func (StockQueue) TableName() string { return "stock_queues" }
