package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// OrderGift 订单赠品模型
type OrderGift struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	OrderID	int64	`gorm:"column:order_id;not null" json:"order_id" example:"1"`
	GiftID	int64	`gorm:"column:gift_id;not null" json:"gift_id" example:"1"`
	GiftName	string	`gorm:"column:gift_name;type:varchar(100)" json:"gift_name" example:"精美礼品盒"`
	CostPrice	decimal.Decimal	`gorm:"column:cost_price;type:decimal(10,2);default:0.00" json:"cost_price" example:"1600.00"`
	Quantity	int	`gorm:"column:quantity;default:0" json:"quantity" example:"10"`
	TotalCost	decimal.Decimal	`gorm:"column:total_cost;type:decimal(12,2);default:0.00" json:"total_cost" example:"3200.00"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	Order *Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Gift  *Gift  `gorm:"foreignKey:GiftID" json:"gift,omitempty"`
}

// TableName 指定表名
func (OrderGift) TableName() string { return "order_gifts" }
