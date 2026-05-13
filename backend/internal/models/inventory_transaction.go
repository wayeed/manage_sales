package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// InventoryTransaction 库存流水模型
type InventoryTransaction struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	StoreID	int64	`gorm:"column:store_id;default:1" json:"store_id" example:"1"`
	WarehouseID	*int64	`gorm:"column:warehouse_id" json:"warehouse_id" example:"1"`
	TransactionType	int8	`gorm:"column:transaction_type" json:"transaction_type" example:"1"`
	BizType	int8	`gorm:"column:biz_type" json:"biz_type" example:"order"`
	BizID	*int64	`gorm:"column:biz_id" json:"biz_id" example:"1"`
	BatchID	*int64	`gorm:"column:batch_id" json:"batch_id" example:"1"`
	RelatedOrderID	*int64	`gorm:"column:related_order_id" json:"related_order_id" example:"1"`
	RelatedPurchaseID	*int64	`gorm:"column:related_purchase_id" json:"related_purchase_id" example:"1"`
	Quantity	int	`gorm:"column:quantity" json:"quantity" example:"10"`
	BeforeStock	int	`gorm:"column:before_stock" json:"before_stock" example:"100"`
	AfterStock	int	`gorm:"column:after_stock" json:"after_stock" example:"90"`
	UnitCost	decimal.Decimal	`gorm:"column:unit_cost;type:decimal(12,2);default:0.00" json:"unit_cost" example:"150.00"`
	TotalCost	decimal.Decimal	`gorm:"column:total_cost;type:decimal(12,2);default:0.00" json:"total_cost" example:"3200.00"`
	Remark	string	`gorm:"column:remark;type:varchar(500)" json:"remark" example:"备注信息"`
	CreatedBy	*int64	`gorm:"column:created_by" json:"created_by" example:"1"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	SKU       *ProductSKU `gorm:"foreignKey:BizID;references:ID" json:"sku,omitempty"`
	Warehouse *Warehouse  `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	Creator   *User       `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// TableName 指定表名
func (InventoryTransaction) TableName() string { return "inventory_transactions" }
