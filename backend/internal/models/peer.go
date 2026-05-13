package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Peer 同行模型
type Peer struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	StoreID	int64	`gorm:"column:store_id;default:1" json:"store_id" example:"1"`
	PeerName	string	`gorm:"column:peer_name;type:varchar(50)" json:"peer_name" example:"赵六"`
	Phone	string	`gorm:"column:phone;type:varchar(20)" json:"phone" example:"13800138000"`
	IDCard	string	`gorm:"column:id_card;type:varchar(18)" json:"id_card" example:"110101199001011234"`
	Company	string	`gorm:"column:company;type:varchar(100)" json:"company" example:"某某科技有限公司"`
	BankAccount	string	`gorm:"column:bank_account;type:varchar(50)" json:"bank_account" example:"6222021234567890123"`
	BankName	string	`gorm:"column:bank_name;type:varchar(100)" json:"bank_name" example:"中国工商银行"`
	TotalOrders	int	`gorm:"column:total_orders;default:0" json:"total_orders" example:"10"`
	TotalAmount	decimal.Decimal	`gorm:"column:total_amount;type:decimal(14,2);default:0.00" json:"total_amount" example:"49980.00"`
	TotalProfit	decimal.Decimal	`gorm:"column:total_profit;type:decimal(14,2);default:0.00" json:"total_profit" example:"17480.00"`
	TotalCommission	decimal.Decimal	`gorm:"column:total_commission;type:decimal(14,2);default:0.00" json:"total_commission" example:"1748.00"`
	PaidCommission	decimal.Decimal	`gorm:"column:paid_commission;type:decimal(14,2);default:0.00" json:"paid_commission" example:"1000.00"`
	UnpaidCommission	decimal.Decimal	`gorm:"column:unpaid_commission;type:decimal(14,2);default:0.00" json:"unpaid_commission" example:"748.00"`
	LastOrderAt	*time.Time	`gorm:"column:last_order_at" json:"last_order_at" example:"2025-01-15T00:00:00+08:00"`
	CommissionRate	*decimal.Decimal	`gorm:"column:commission_rate;type:decimal(5,4)" json:"commission_rate" example:"0.1000"`
	Remark	string	`gorm:"column:remark;type:varchar(255)" json:"remark" example:"备注信息"`
	Status	int8	`gorm:"column:status;default:1" json:"status" example:"1"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`
	DeletedAt        gorm.DeletedAt  `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Peer) TableName() string { return "peers" }
