package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// FundPoolShare 基金池份额模型
// status: 0-待发放,1-已发放
type FundPoolShare struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	FundPoolID	int64	`gorm:"column:fund_pool_id;not null;index" json:"fund_pool_id" example:"1"`
	EmployeeID	int64	`gorm:"column:employee_id;not null;index" json:"employee_id" example:"1"`
	PersonalProfit	decimal.Decimal	`gorm:"column:personal_profit;type:decimal(14,2);default:0.00" json:"personal_profit" example:"50000.00"`
	Shares	decimal.Decimal	`gorm:"column:shares;type:decimal(14,4);default:0.00" json:"shares" example:"50.0000"`
	RewardAmount	decimal.Decimal	`gorm:"column:reward_amount;type:decimal(12,2);default:0.00" json:"reward_amount" example:"5000.00"`
	Status	int8	`gorm:"column:status;default:0" json:"status" example:"1"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	FundPool *FundPool `gorm:"foreignKey:FundPoolID" json:"fund_pool,omitempty"`
	Employee *User     `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

// TableName 指定表名
func (FundPoolShare) TableName() string { return "fund_pool_shares" }
