package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// SalaryRecord 工资记录模型
// status: 0-草稿,1-已确认,2-已发放
type SalaryRecord struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	StoreID	int64	`gorm:"column:store_id;not null;index" json:"store_id" example:"1"`
	EmployeeID	int64	`gorm:"column:employee_id;not null;index" json:"employee_id" example:"1"`
	SalaryMonth	string	`gorm:"column:salary_month;type:varchar(10);not null;index" json:"salary_month" example:"2025-01"`
	BaseSalary	decimal.Decimal	`gorm:"column:base_salary;type:decimal(12,2);default:0.00" json:"base_salary" example:"5000.00"`
	SalesCommission	decimal.Decimal	`gorm:"column:sales_commission;type:decimal(12,2);default:0.00" json:"sales_commission" example:"3000.00"`
	TeamCommission	decimal.Decimal	`gorm:"column:team_commission;type:decimal(12,2);default:0.00" json:"team_commission" example:"500.00"`
	FundReward	decimal.Decimal	`gorm:"column:fund_reward;type:decimal(12,2);default:0.00" json:"fund_reward" example:"200.00"`
	ReferralReward	decimal.Decimal	`gorm:"column:referral_reward;type:decimal(12,2);default:0.00" json:"referral_reward" example:"300.00"`
	Deduction	decimal.Decimal	`gorm:"column:deduction;type:decimal(12,2);default:0.00" json:"deduction" example:"0.00"`
	Bonus	decimal.Decimal	`gorm:"column:bonus;type:decimal(12,2);default:0.00" json:"bonus" example:"1000.00"`
	GrossSalary	decimal.Decimal	`gorm:"column:gross_salary;type:decimal(12,2);default:0.00" json:"gross_salary" example:"10000.00"`
	NetSalary	decimal.Decimal	`gorm:"column:net_salary;type:decimal(12,2);default:0.00" json:"net_salary" example:"10000.00"`
	Status	int8	`gorm:"column:status;default:0" json:"status" example:"1"`
	PaidAmount	decimal.Decimal	`gorm:"column:paid_amount;type:decimal(12,2);default:0.00" json:"paid_amount" example:"10000.00"`
	PaidAt	*time.Time	`gorm:"column:paid_at" json:"paid_at" example:"2025-02-01T00:00:00+08:00"`
	PaidBy	*int64	`gorm:"column:paid_by" json:"paid_by" example:"1"`
	PayMethod	int8	`gorm:"column:pay_method;default:0" json:"pay_method" example:"1"`
	PayRemark	string	`gorm:"column:pay_remark;type:varchar(500)" json:"pay_remark" example:"银行转账"`
	Remark	string	`gorm:"column:remark;type:varchar(500)" json:"remark" example:"备注信息"`
	ConfirmedBy	*int64	`gorm:"column:confirmed_by" json:"confirmed_by" example:"1"`
	ConfirmedAt	*time.Time	`gorm:"column:confirmed_at" json:"confirmed_at" example:"2025-01-31T00:00:00+08:00"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	Employee *User         `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Items    []SalaryItem  `gorm:"foreignKey:SalaryRecordID" json:"items,omitempty"`
}

// TableName 指定表名
func (SalaryRecord) TableName() string { return "salary_records" }
