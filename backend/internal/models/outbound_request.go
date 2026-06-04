package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// OutboundRequest 出库申请模型
type OutboundRequest struct {
	ID             int64           `gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	OrderID        int64           `gorm:"column:order_id;not null;uniqueIndex" json:"order_id" example:"1"`
	ApplicantID    int64           `gorm:"column:applicant_id;not null" json:"applicant_id" example:"1"`
	ApplicantName  string          `gorm:"column:applicant_name;not null" json:"applicant_name" example:"张三"`
	Status         int8            `gorm:"column:status;not null;default:0;index" json:"status" example:"1"` // 1=主管待审批,2=财务待审批,3=已拒绝,4=已通过
	RemainingAmount decimal.Decimal `gorm:"column:remaining_amount;type:decimal(12,2);default:0.00" json:"remaining_amount" example:"5000.00"`
	RemainingRate  float64         `gorm:"column:remaining_rate;default:0" json:"remaining_rate" example:"25.5"`
	Remark         string          `gorm:"column:remark;type:varchar(500)" json:"remark" example:"此订单尾款还有5000元"`
	// 主管审批
	SupervisorID    *int64     `gorm:"column:supervisor_id" json:"supervisor_id" example:"2"`
	SupervisorName  *string    `gorm:"column:supervisor_name" json:"supervisor_name" example:"李四"`
	SupervisorAt    *time.Time `gorm:"column:supervisor_at" json:"supervisor_at"`
	SupervisorRemark *string   `gorm:"column:supervisor_remark;type:varchar(500)" json:"supervisor_remark"`
	// 财务审批
	FinanceID       *int64     `gorm:"column:finance_id" json:"finance_id" example:"3"`
	FinanceName     *string    `gorm:"column:finance_name" json:"finance_name" example:"王五"`
	FinanceAt       *time.Time `gorm:"column:finance_at" json:"finance_at"`
	FinanceRemark   *string    `gorm:"column:finance_remark;type:varchar(500)" json:"finance_remark"`
	CreatedAt       time.Time  `json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt       time.Time  `json:"updated_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	Order *Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

// TableName 指定表名
func (OutboundRequest) TableName() string { return "outbound_requests" }
