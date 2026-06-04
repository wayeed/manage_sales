package models

import (
	"time"

	"gorm.io/gorm"
)

// SystemBackup 系统备份记录模型
type SystemBackup struct {
	ID           int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	BackupType   int8           `gorm:"column:backup_type;default:0" json:"backup_type"`                 // 备份类型: 0-手动, 1-自动
	FileName     string         `gorm:"column:file_name;type:varchar(255);not null" json:"file_name"`     // 备份文件名
	FilePath     string         `gorm:"column:file_path;type:varchar(500);not null" json:"file_path"`     // 备份文件路径
	FileSize     int64          `gorm:"column:file_size;default:0" json:"file_size"`                      // 文件大小(字节)
	Status       int8           `gorm:"column:status;default:0" json:"status"`                            // 状态: 0-进行中, 1-成功, 2-失败
	ErrorMessage string         `gorm:"column:error_message;type:text" json:"error_message"`              // 错误信息
	CreatedBy    *int64         `gorm:"column:created_by" json:"created_by"`                              // 创建人ID
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`               // 创建时间
	FinishedAt   *time.Time     `gorm:"column:finished_at" json:"finished_at"`                            // 完成时间
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`                        // 删除时间
}

// TableName 表名
func (SystemBackup) TableName() string {
	return "system_backups"
}
