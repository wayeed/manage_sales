package models

import (
	"time"
)

// FollowUpApproval 申请跟进审批模型
type FollowUpApproval struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	StoreID	int64	`gorm:"column:store_id;not null;index" json:"store_id" example:"1"`
	CustomerID	int64	`gorm:"column:customer_id;not null;index" json:"customer_id" example:"1"`
	OrderID	*int64	`gorm:"column:order_id;index" json:"order_id" example:"1"` // 关联订单ID（送货单打印审批时使用）
	ApplicantID	int64	`gorm:"column:applicant_id;not null" json:"applicant_id" example:"1"`
	ApproverID	*int64	`gorm:"column:approver_id" json:"approver_id" example:"1"`
	ApprovalType	int8	`gorm:"column:approval_type;default:1" json:"approval_type" example:"1"` // 1-跟进转交审批, 2-送货单打印审批
	Status	int8	`gorm:"column:status;default:0;index" json:"status" example:"1"` // 0-待审批,1-已通过,2-已拒绝
	Remark	string	`gorm:"column:remark;type:varchar(255)" json:"remark" example:"备注信息"`
	RejectReason	string	`gorm:"column:reject_reason;type:varchar(255)" json:"reject_reason" example:""`
	ApprovedAt	*time.Time	`gorm:"column:approved_at" json:"approved_at" example:"2025-01-15T14:00:00+08:00"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	Customer   *Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Order      *Order    `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Applicant  *User     `gorm:"foreignKey:ApplicantID" json:"applicant,omitempty"`
	Approver   *User     `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`
}

// TableName 指定表名
func (FollowUpApproval) TableName() string { return "follow_up_approvals" }
