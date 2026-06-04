package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Product 商品模型
type Product struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	StoreID	int64	`gorm:"column:store_id;default:1" json:"store_id" example:"1"`
	CategoryID	*int64	`gorm:"column:category_id" json:"category_id" example:"1"`
	ProductCode	string	`gorm:"column:product_code;type:varchar(32);uniqueIndex" json:"product_code" example:"PRD001"`
	ProductName	string	`gorm:"column:product_name;type:varchar(100)" json:"product_name" example:"智能手机"`
	Brand	string	`gorm:"column:brand;type:varchar(50)" json:"brand" example:"华为"`
	Style         string          `gorm:"column:style;type:varchar(100)" json:"style" example:""`
	Unit          string          `gorm:"column:unit;type:varchar(20)" json:"unit" example:"件"`
	Series        string          `gorm:"column:series;type:varchar(50)" json:"series" example:"现代系列"`
	SubCategory   string          `gorm:"column:sub_category;type:varchar(10)" json:"sub_category" example:"A"`
	ProductImage	string	`gorm:"column:product_image;type:varchar(255)" json:"product_image" example:"https://example.com/images/product001.jpg"`
	Description	string	`gorm:"column:description;type:text" json:"description" example:"高端智能手机，性能卓越"`
	ListPrice	decimal.Decimal	`gorm:"column:list_price;type:decimal(12,2);default:0.00" json:"list_price" example:"2999.00"`
	MinPrice	decimal.Decimal	`gorm:"column:min_price;type:decimal(12,2);default:0.00" json:"min_price" example:"2499.00"`
	ReferenceCost	decimal.Decimal	`gorm:"column:reference_cost;type:decimal(12,2);default:0.00" json:"reference_cost" example:"1800.00"`
	CostPrice	decimal.Decimal	`gorm:"column:cost_price;type:decimal(12,2);default:0.00" json:"cost_price" example:"1600.00"`
	TotalCostRate	decimal.Decimal	`gorm:"column:total_cost_rate;type:decimal(5,4);default:1.2000" json:"total_cost_rate" example:"1.2000"`
	WarningStock	int	`gorm:"column:warning_stock;default:10" json:"warning_stock" example:"10"`
	Status	int8	`gorm:"column:status;default:1" json:"status" example:"1"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`
	CreatedBy	*int64	`gorm:"column:created_by" json:"created_by" example:"1"`

	// 关联
	Category *Category   `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	SKUs     []ProductSKU `gorm:"foreignKey:ProductID" json:"skus,omitempty"`
}

// TableName 指定表名
func (Product) TableName() string { return "products" }
