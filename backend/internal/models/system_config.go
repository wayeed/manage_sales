package models

import "time"

// SystemConfig 系统配置模型
type SystemConfig struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	ConfigKey	string	`gorm:"column:config_key;type:varchar(128);uniqueIndex;not null" json:"config_key" example:"site_name"`
	ConfigValue	string	`gorm:"column:config_value;type:varchar(512);not null" json:"config_value" example:"ERP管理系统"`
	ConfigType	string	`gorm:"column:config_type;type:varchar(32);default:'string'" json:"config_type" example:"string"`
	Remark	string	`gorm:"column:remark;type:varchar(255)" json:"remark" example:"备注信息"`
	Sort	int	`gorm:"column:sort;default:0" json:"sort" example:"1"` // 排序字段
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`
}

// TableName 指定表名
func (SystemConfig) TableName() string { return "system_configs" }
