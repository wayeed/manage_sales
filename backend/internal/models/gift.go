package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Gift 礼品模型
type Gift struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	StoreID	int64	`gorm:"column:store_id;default:1" json:"store_id" example:"1"`
	GiftCode	string	`gorm:"column:gift_code;type:varchar(32);uniqueIndex" json:"gift_code" example:"GIFT001"`
	GiftName	string	`gorm:"column:gift_name;type:varchar(100)" json:"gift_name" example:"精美礼品盒"`
	GiftImage	string	`gorm:"column:gift_image;type:varchar(255)" json:"gift_image" example:"https://example.com/images/gift001.jpg"`
	CostPrice	decimal.Decimal	`gorm:"column:cost_price;type:decimal(10,2);default:0.00" json:"cost_price" example:"1600.00"`
	StockQuantity	int	`gorm:"column:stock_quantity;default:0" json:"stock_quantity" example:"100"`
	WarningStock	int	`gorm:"column:warning_stock;default:10" json:"warning_stock" example:"10"`
	Status	int8	`gorm:"column:status;default:1" json:"status" example:"1"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`
	DeletedAt     gorm.DeletedAt  `gorm:"index" json:"-"`
	CreatedBy	*int64	`gorm:"column:created_by" json:"created_by" example:"1"`
}

// TableName 指定表名
func (Gift) TableName() string { return "gifts" }
