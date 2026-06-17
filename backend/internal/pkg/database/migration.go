package database

import (
	"fmt"
	"furniture-commission/internal/models"

	"gorm.io/gorm"
)

// SafeMigration 提供安全的数据库迁移方法
type SafeMigration struct {
	db *gorm.DB
}

// NewSafeMigration 创建安全迁移实例
func NewSafeMigration(db *gorm.DB) *SafeMigration {
	return &SafeMigration{db: db}
}

// AddColumnIfNotExists 安全地添加列（使用GORM API，自动处理转义）
func (m *SafeMigration) AddColumnIfNotExists(model interface{}, columnName string) error {
	if m.db.Migrator().HasColumn(model, columnName) {
		return nil
	}
	return m.db.Migrator().AddColumn(model, columnName)
}

// HasColumn 安全地检查列是否存在
func (m *SafeMigration) HasColumn(model interface{}, columnName string) bool {
	return m.db.Migrator().HasColumn(model, columnName)
}

// HasTable 安全地检查表是否存在
func (m *SafeMigration) HasTable(model interface{}) bool {
	return m.db.Migrator().HasTable(model)
}

// CreateTable 安全地创建表
func (m *SafeMigration) CreateTable(model interface{}) error {
	return m.db.AutoMigrate(model)
}

// MigratePaymentsTable 安全地为payments表添加缺少的列
func (m *SafeMigration) MigratePaymentsTable() error {
	if m.HasColumn(&models.Payment{}, "updated_at") {
		return nil
	}
	return m.AddColumnIfNotExists(&models.Payment{}, "updated_at")
}

// ExecuteRawSQL 已废弃 - 禁止使用原始SQL
func (m *SafeMigration) ExecuteRawSQL(sql string) error {
	return fmt.Errorf("ExecuteRawSQL is deprecated, use GORM Migrator API instead")
}

// isValidMD5 检查字符串是否为有效的MD5哈希
func isValidMD5(s string) bool {
	if len(s) != 32 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
