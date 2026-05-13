package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Customer 客户模型
type Customer struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	StoreID	int64	`gorm:"column:store_id;default:1" json:"store_id" example:"1"`
	CustomerCode	string	`gorm:"column:customer_code;type:varchar(32);uniqueIndex" json:"customer_code" example:"CUS001"`
	CustomerName	string	`gorm:"column:customer_name;type:varchar(100)" json:"customer_name" example:"王五"`
	Phone	string	`gorm:"column:phone;type:varchar(20)" json:"phone" example:"13800138000"`
	Email	string	`gorm:"column:email;type:varchar(128)" json:"email" example:"admin@example.com"`
	Address	string	`gorm:"column:address;type:varchar(500)" json:"address" example:"北京市朝阳区建国路88号"`
	Gender	int8	`gorm:"column:gender;default:0" json:"gender" example:"1"`
	Birthday	*time.Time	`gorm:"column:birthday" json:"birthday" example:"1990-01-15T00:00:00+08:00"`
	Level	int8	`gorm:"column:level;default:0" json:"level" example:"1"`
	TotalOrders	int	`gorm:"column:total_orders;default:0" json:"total_orders" example:"5"`
	TotalAmount	decimal.Decimal	`gorm:"column:total_amount;type:decimal(12,2);default:0.00" json:"total_amount" example:"24990.00"`
	TotalProfit	decimal.Decimal	`gorm:"column:total_profit;type:decimal(12,2);default:0.00" json:"total_profit" example:"8740.00"`
	LastOrderAt	*time.Time	`gorm:"column:last_order_at" json:"last_order_at" example:"2025-01-15T00:00:00+08:00"`
	Remark	string	`gorm:"column:remark;type:varchar(500)" json:"remark" example:"备注信息"`
	Status	int8	`gorm:"column:status;default:1" json:"status" example:"1"`
	CreatedBy	*int64	`gorm:"column:created_by" json:"created_by" example:"1"`       // 创建者（录入客户的人）
	SalesmanID	*int64	`gorm:"column:salesman_id" json:"salesman_id" example:"1"`     // 负责业务员（跟进客户的人）
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`
	DeletedAt     gorm.DeletedAt  `gorm:"index" json:"-"`

	// 关联
	Salesman      *User           `gorm:"foreignKey:SalesmanID" json:"salesman,omitempty"`
}

// TableName 指定表名
func (Customer) TableName() string { return "customers" }
