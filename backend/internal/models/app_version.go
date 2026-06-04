package models

import (
	"time"

	"gorm.io/gorm"
)

// AppVersion APP版本模型
type AppVersion struct {
	ID             int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Platform       string         `gorm:"column:platform;type:varchar(20);not null" json:"platform"`                    // 平台: ios, android
	VersionCode    string         `gorm:"column:version_code;type:varchar(32);not null" json:"version_code"`            // 版本号(如: 1.0.0)
	VersionName    string         `gorm:"column:version_name;type:varchar(64);not null" json:"version_name"`            // 版本名称
	DownloadURL    string         `gorm:"column:download_url;type:varchar(500);not null" json:"download_url"`           // 安装包下载地址
	FileSize       int64          `gorm:"column:file_size;default:0" json:"file_size"`                                  // 文件大小(字节)
	UpdateType     int8           `gorm:"column:update_type;default:0" json:"update_type"`                              // 更新类型: 0-整包更新, 1-热更新
	IsForceUpdate  int8           `gorm:"column:is_force_update;default:0" json:"is_force_update"`                      // 是否强制更新: 0-否, 1-是
	MinVersion     string         `gorm:"column:min_version;type:varchar(32);default:''" json:"min_version"`            // 最低支持版本
	UpdateContent  string         `gorm:"column:update_content;type:text" json:"update_content"`                        // 更新内容
	Status         int8           `gorm:"column:status;default:1" json:"status"`                                        // 状态: 0-禁用, 1-启用
	PublishedAt    *time.Time     `gorm:"column:published_at" json:"published_at"`                                      // 发布时间
	CreatedBy      *int64         `gorm:"column:created_by" json:"created_by"`                                          // 创建人ID
	CreatedAt      time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`                           // 创建时间
	UpdatedAt      time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`                           // 更新时间
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`                                    // 删除时间
}

// TableName 表名
func (AppVersion) TableName() string {
	return "api_app_versions"
}
