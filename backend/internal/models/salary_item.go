package models

import "time"

// SalaryItem 工资明细模型
// item_type: 1-基本工资,2-销售提成,3-团队分润,4-基金池奖励,5-老带新奖励,6-扣减,7-奖金
type SalaryItem struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	SalaryRecordID	int64	`gorm:"column:salary_record_id;not null;index" json:"salary_record_id" example:"1"`
	ItemType	int8	`gorm:"column:item_type;not null" json:"item_type" example:"1"`
	ItemName	string	`gorm:"column:item_name;type:varchar(64);not null" json:"item_name" example:"基本工资"`
	Amount	float64	`gorm:"column:amount;type:decimal(12,2);default:0.00" json:"amount" example:"4998.00"`
	Remark	string	`gorm:"column:remark;type:varchar(255)" json:"remark" example:"备注信息"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	SalaryRecord *SalaryRecord `gorm:"foreignKey:SalaryRecordID" json:"salary_record,omitempty"`
}

// TableName 指定表名
func (SalaryItem) TableName() string { return "salary_items" }
