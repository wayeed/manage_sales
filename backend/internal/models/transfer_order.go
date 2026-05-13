package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// TransferOrder 调拨单模型
type TransferOrder struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	StoreID	int64	`gorm:"column:store_id;default:1" json:"store_id" example:"1"`
	TransferNo	string	`gorm:"column:transfer_no;type:varchar(32);uniqueIndex" json:"transfer_no" example:"TO20250115001"`
	FromWarehouseID	*int64	`gorm:"column:from_warehouse_id" json:"from_warehouse_id" example:"1"`
	ToWarehouseID	*int64	`gorm:"column:to_warehouse_id" json:"to_warehouse_id" example:"1"`
	TotalQuantity	int	`gorm:"column:total_quantity;default:0" json:"total_quantity" example:"5"`
	TotalAmount	decimal.Decimal	`gorm:"column:total_amount;type:decimal(12,2);default:0.00" json:"total_amount" example:"49980.00"`
	Status	int8	`gorm:"column:status;default:0" json:"status" example:"1"`
	Remark	string	`gorm:"column:remark;type:varchar(500)" json:"remark" example:"备注信息"`
	CreatedBy	*int64	`gorm:"column:created_by" json:"created_by" example:"1"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`
	AuditedBy	*int64	`gorm:"column:audited_by" json:"audited_by" example:"1"`
	AuditedAt	*time.Time	`gorm:"column:audited_at" json:"audited_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	FromWarehouse *Warehouse     `gorm:"foreignKey:FromWarehouseID" json:"from_warehouse,omitempty"`
	ToWarehouse   *Warehouse     `gorm:"foreignKey:ToWarehouseID" json:"to_warehouse,omitempty"`
	Items         []TransferItem `gorm:"foreignKey:TransferOrderID" json:"items,omitempty"`
}

// TableName 指定表名
func (TransferOrder) TableName() string { return "transfer_orders" }
