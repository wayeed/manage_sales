package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Commission 提成记录模型
// commission_type: 1-业务员提成,2-同行分成,3-主管团队分润,4-店长团队分润,5-基金池奖励,6-老带新奖励
// status: 0-待回款,1-可发放,2-已发放
type Commission struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	StoreID	int64	`gorm:"column:store_id;not null;index" json:"store_id" example:"1"`
	OrderID	int64	`gorm:"column:order_id;not null;index" json:"order_id" example:"1"`
	EmployeeID	*int64	`gorm:"column:employee_id;index" json:"employee_id" example:"1"`
	PeerID	*int64	`gorm:"column:peer_id" json:"peer_id" example:"1"`
	CommissionType	int8	`gorm:"column:commission_type;not null" json:"commission_type" example:"1"`
	PeriodValue	string	`gorm:"column:period_value;type:varchar(10);index" json:"period_value" example:"2025-01"`
	BaseAmount	decimal.Decimal	`gorm:"column:base_amount;type:decimal(12,2);default:0.00" json:"base_amount" example:"4998.00"`
	Rate	decimal.Decimal	`gorm:"column:rate;type:decimal(6,4);default:0.00" json:"rate" example:"0.0500"`
	Amount	decimal.Decimal	`gorm:"column:amount;type:decimal(12,2);default:0.00" json:"amount" example:"4998.00"`
	Status	int8	`gorm:"column:status;default:0;index" json:"status" example:"1"`
	SettledAt	*time.Time	`gorm:"column:settled_at" json:"settled_at" example:"2025-02-01T00:00:00+08:00"`
	Remark	string	`gorm:"column:remark;type:varchar(500)" json:"remark" example:"备注信息"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	Order    *Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Employee *User  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Peer     *Peer  `gorm:"foreignKey:PeerID" json:"peer,omitempty"`
}

// TableName 指定表名
func (Commission) TableName() string { return "commissions" }
