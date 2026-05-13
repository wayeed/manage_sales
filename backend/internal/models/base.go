package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型（含软删除）
type BaseModel struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
