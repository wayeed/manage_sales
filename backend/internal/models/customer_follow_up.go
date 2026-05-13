package models

import "time"

// CustomerFollowUp 客户跟进记录模型
type CustomerFollowUp struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	CustomerID	int64	`gorm:"column:customer_id;not null" json:"customer_id" example:"1"`
	FollowerID	int64	`gorm:"column:follower_id;not null" json:"follower_id" example:"1"`
	FollowType	int8	`gorm:"column:follow_type;default:0" json:"follow_type" example:"1"`
	Content	string	`gorm:"column:content;type:text" json:"content" example:"客户对产品很感兴趣，需要后续跟进"`
	NextFollowDate	*time.Time	`gorm:"column:next_follow_date" json:"next_follow_date" example:"2025-01-22T00:00:00+08:00"`
	NextFollowContent	string	`gorm:"column:next_follow_content;type:varchar(500)" json:"next_follow_content" example:"发送产品详细资料"`
	IsDeal	int8	`gorm:"column:is_deal;default:0" json:"is_deal" example:"0"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	Customer *Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Follower *User     `gorm:"foreignKey:FollowerID" json:"follower,omitempty"`
}

// TableName 指定表名
func (CustomerFollowUp) TableName() string { return "customer_follow_ups" }
