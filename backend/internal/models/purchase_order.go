package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// PurchaseOrder 采购订单模型
type PurchaseOrder struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	StoreID	int64	`gorm:"column:store_id;default:1" json:"store_id" example:"1"`
	PurchaseNo	string	`gorm:"column:purchase_no;type:varchar(32);uniqueIndex" json:"purchase_no" example:"PO20250115001"`
	SupplierID	*int64	`gorm:"column:supplier_id" json:"supplier_id" example:"1"`
	SupplierName	string	`gorm:"column:supplier_name;type:varchar(100)" json:"supplier_name" example:"深圳某某科技有限公司"`
	TotalAmount	decimal.Decimal	`gorm:"column:total_amount;type:decimal(12,2);default:0.00" json:"total_amount" example:"49980.00"`
	TotalQuantity	int	`gorm:"column:total_quantity;default:0" json:"total_quantity" example:"5"`
	Status	int8	`gorm:"column:status;default:0" json:"status" example:"1"`
	Remark	string	`gorm:"column:remark;type:varchar(500)" json:"remark" example:"备注信息"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`
	CreatedBy	*int64	`gorm:"column:created_by" json:"created_by" example:"1"`
	AuditedBy	*int64	`gorm:"column:audited_by" json:"audited_by" example:"1"`
	AuditedAt	*time.Time	`gorm:"column:audited_at" json:"audited_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	Items []PurchaseItem `gorm:"foreignKey:PurchaseOrderID" json:"items,omitempty"`
}

// TableName 指定表名
func (PurchaseOrder) TableName() string { return "purchase_orders" }
