package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// FundPool 基金池模型
// period_type: 1-月度,2-季度,3-年度
// status: 0-待结算,1-已结算
type FundPool struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	StoreID	int64	`gorm:"column:store_id;not null;index" json:"store_id" example:"1"`
	PeriodType	int8	`gorm:"column:period_type;not null" json:"period_type" example:"1"`
	PeriodValue	string	`gorm:"column:period_value;type:varchar(10);not null" json:"period_value" example:"2025-01"`
	TotalProfit	decimal.Decimal	`gorm:"column:total_profit;type:decimal(14,2);default:0.00" json:"total_profit" example:"100000.00"`
	ExtractRate	decimal.Decimal	`gorm:"column:extract_rate;type:decimal(6,4);default:0.00" json:"extract_rate" example:"0.1000"`
	PoolAmount	decimal.Decimal	`gorm:"column:pool_amount;type:decimal(14,2);default:0.00" json:"pool_amount" example:"10000.00"`
	TotalShares	decimal.Decimal	`gorm:"column:total_shares;type:decimal(14,4);default:0.00" json:"total_shares" example:"100.0000"`
	PerShareAmount	decimal.Decimal	`gorm:"column:per_share_amount;type:decimal(12,2);default:0.00" json:"per_share_amount" example:"100.00"`
	Status	int8	`gorm:"column:status;default:0" json:"status" example:"1"`
	SettledAt	*time.Time	`gorm:"column:settled_at" json:"settled_at" example:"2025-02-01T00:00:00+08:00"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	Shares []FundPoolShare `gorm:"foreignKey:FundPoolID" json:"shares,omitempty"`
}

// TableName 指定表名
func (FundPool) TableName() string { return "fund_pools" }
