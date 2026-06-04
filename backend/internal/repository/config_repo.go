package repository

import (
	"furniture-commission/internal/models"

	"gorm.io/gorm"
)

// SystemConfigRepository 系统配置Repository
type SystemConfigRepository struct {
	db *gorm.DB
}

// NewSystemConfigRepository 创建系统配置Repository实例
func NewSystemConfigRepository(db *gorm.DB) *SystemConfigRepository {
	return &SystemConfigRepository{db: db}
}

// Get 获取配置值
func (r *SystemConfigRepository) Get(key string) (string, error) {
	var config models.SystemConfig
	err := r.db.Where("config_key = ?", key).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return config.ConfigValue, nil
}

// GetConfigType 获取配置类型
func (r *SystemConfigRepository) GetConfigType(key string) (string, error) {
	var config models.SystemConfig
	err := r.db.Where("config_key = ?", key).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return config.ConfigType, nil
}

// GetConfigTypeAndRemark 获取配置类型和备注
func (r *SystemConfigRepository) GetConfigTypeAndRemark(key string) (string, string, error) {
	var config models.SystemConfig
	err := r.db.Where("config_key = ?", key).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", "", nil
		}
		return "", "", err
	}
	return config.ConfigType, config.Remark, nil
}

// Set 设置配置值（存在则更新，不存在则创建）
func (r *SystemConfigRepository) Set(key, value, configType, remark string) error {
	var config models.SystemConfig
	err := r.db.Where("config_key = ?", key).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			config = models.SystemConfig{
				ConfigKey:   key,
				ConfigValue: value,
				ConfigType:  configType,
				Remark:      remark,
			}
			return r.db.Create(&config).Error
		}
		return err
	}
	return r.db.Model(&config).Updates(map[string]interface{}{
		"config_value": value,
		"config_type":  configType,
		"remark":       remark,
	}).Error
}

// GetAll 获取所有配置
func (r *SystemConfigRepository) GetAll() ([]models.SystemConfig, error) {
	var configs []models.SystemConfig
	err := r.db.Order("sort ASC, id ASC").Find(&configs).Error
	if err != nil {
		return nil, err
	}
	return configs, nil
}

// BatchSet 批量设置配置
func (r *SystemConfigRepository) BatchSet(configs []models.SystemConfig) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, c := range configs {
			var existing models.SystemConfig
			err := tx.Where("config_key = ?", c.ConfigKey).First(&existing).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					if err := tx.Create(&c).Error; err != nil {
						return err
					}
					continue
				}
				return err
			}
			if err := tx.Model(&existing).Updates(map[string]interface{}{
				"config_value": c.ConfigValue,
				"config_type":  c.ConfigType,
				"remark":       c.Remark,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
