package models

import "time"

// OperationLog 操作日志模型
type OperationLog struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	UserID	int64	`gorm:"column:user_id" json:"user_id" example:"1"`
	Username	string	`gorm:"column:username;type:varchar(50)" json:"username" example:"admin"`
	Action	string	`gorm:"column:action;type:varchar(100)" json:"action" example:"创建订单"`
	BizType	string	`gorm:"column:biz_type;type:varchar(50)" json:"biz_type" example:"order"`
	BizID	int64	`gorm:"column:biz_id" json:"biz_id" example:"1"`
	Detail	string	`gorm:"column:detail;type:text" json:"detail" example:"创建订单 ORD20250115001"`
	IPAddress	string	`gorm:"column:ip_address;type:varchar(50)" json:"ip_address" example:"192.168.1.100"`
	UserAgent	string	`gorm:"column:user_agent;type:varchar(255)" json:"user_agent" example:"Mozilla/5.0 (Windows NT 10.0; Win64; x64)"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
}

// TableName 指定表名
func (OperationLog) TableName() string { return "operation_logs" }
