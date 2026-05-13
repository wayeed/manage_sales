package models

import "time"

// ReferralRelation 老带新引荐关系模型
// status: 1-生效中,0-已终止
type ReferralRelation struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	ReferrerID	int64	`gorm:"column:referrer_id;not null;index" json:"referrer_id" example:"1"`
	ReferredID	int64	`gorm:"column:referred_id;not null;index" json:"referred_id" example:"1"`
	Status	int8	`gorm:"column:status;default:1" json:"status" example:"1"`
	EndedAt	*time.Time	`gorm:"column:ended_at" json:"ended_at" example:"2025-12-31T00:00:00+08:00"`
	EndedReason	string	`gorm:"column:ended_reason;type:varchar(255)" json:"ended_reason" example:"员工离职"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	Referrer *User `gorm:"foreignKey:ReferrerID" json:"referrer,omitempty"`
	Referred *User `gorm:"foreignKey:ReferredID" json:"referred,omitempty"`
}

// TableName 指定表名
func (ReferralRelation) TableName() string { return "referral_relations" }
