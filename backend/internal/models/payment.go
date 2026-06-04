package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Payment 回款记录模型
type Payment struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	OrderID	int64	`gorm:"column:order_id;not null" json:"order_id" example:"1"`
	PaymentNo	string	`gorm:"column:payment_no;type:varchar(32);uniqueIndex" json:"payment_no" example:"PAY20250115001"`
	Amount	decimal.Decimal	`gorm:"column:amount;type:decimal(12,2);default:0.00" json:"amount" example:"4998.00"`
	PaymentDate	*time.Time	`gorm:"column:payment_date" json:"payment_date" example:"2025-01-15T00:00:00+08:00"`
	PaymentMethod	int8	`gorm:"column:payment_method;default:0" json:"payment_method" example:"1"`
	Status	int8	`gorm:"column:status;default:0" json:"status" example:"1"`
	PaymentType	int8	`gorm:"column:payment_type;default:0" json:"payment_type" example:"0"` // 0=普通回款, 1=订金, 2=尾款
	Remark	string	`gorm:"column:remark;type:varchar(500)" json:"remark" example:"备注信息"`
	VoucherURL	string	`gorm:"column:voucher_url;type:varchar(500)" json:"voucher_url" example:"/uploads/images/2025/01/15/xxx.jpg"`
	CreatedBy	*int64	`gorm:"column:created_by" json:"created_by" example:"1"`
	AuditedBy	*int64	`gorm:"column:audited_by" json:"audited_by" example:"1"`
	AuditedAt	*time.Time	`gorm:"column:audited_at" json:"audited_at" example:"2025-01-15T00:00:00+08:00"`
	RejectReason	*string	`gorm:"column:reject_reason;type:varchar(500)" json:"reject_reason" example:"驳回原因"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	Order *Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

// TableName 指定表名
func (Payment) TableName() string { return "payments" }
