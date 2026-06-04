package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// OrderReturn 订单退货记录
type OrderReturn struct {
	ID           int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	StoreID      int64           `gorm:"column:store_id;not null" json:"store_id"`
	OrderID      int64           `gorm:"column:order_id;not null;index" json:"order_id"`
	ReturnNo     string          `gorm:"column:return_no;type:varchar(32);uniqueIndex" json:"return_no"`
	ReturnAmount decimal.Decimal `gorm:"column:return_amount;type:decimal(12,2)" json:"return_amount"`
	ReturnProfit decimal.Decimal `gorm:"column:return_profit;type:decimal(12,2)" json:"return_profit"`
	Reason       string          `gorm:"column:reason;type:varchar(500)" json:"reason"`
	OperatorID   int64           `gorm:"column:operator_id" json:"operator_id"`
	OperatorName string          `gorm:"column:operator_name;type:varchar(50)" json:"operator_name"`
	ReturnTime   time.Time       `gorm:"column:return_time" json:"return_time"`
	CreatedAt    time.Time       `json:"created_at"`
}

// TableName 指定表名
func (OrderReturn) TableName() string {
	return "order_returns"
}
