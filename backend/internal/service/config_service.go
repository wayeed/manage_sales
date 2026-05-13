package service

import (
	"furniture-commission/internal/models"
	"furniture-commission/internal/repository"

	"github.com/shopspring/decimal"
)

// ConfigService 系统配置服务
type ConfigService struct {
	configRepo *repository.SystemConfigRepository
}

// NewConfigService 创建系统配置服务实例
func NewConfigService(configRepo *repository.SystemConfigRepository) *ConfigService {
	return &ConfigService{configRepo: configRepo}
}

// Get 获取配置值
func (s *ConfigService) Get(key string) (string, error) {
	return s.configRepo.Get(key)
}

// GetConfigType 获取配置类型
func (s *ConfigService) GetConfigType(key string) (string, error) {
	return s.configRepo.GetConfigType(key)
}

// GetConfigTypeAndRemark 获取配置类型和备注
func (s *ConfigService) GetConfigTypeAndRemark(key string) (string, string, error) {
	return s.configRepo.GetConfigTypeAndRemark(key)
}

// Set 设置配置值
func (s *ConfigService) Set(key, value, configType, remark string) error {
	return s.configRepo.Set(key, value, configType, remark)
}

// GetAll 获取所有配置
func (s *ConfigService) GetAll() ([]models.SystemConfig, error) {
	return s.configRepo.GetAll()
}

// GetCommissionRates 获取所有提成比例配置
func (s *ConfigService) GetCommissionRates() (map[string]decimal.Decimal, error) {
	keys := []string{
		"commission_rate_single",
		"commission_rate_multi",
		"commission_rate_special",
		"commission_rate_peer_single",
		"commission_rate_peer_multi",
		"commission_rate_peer_special",
		"fund_pool_extract_rate",
		"team_share_rate_manager",
		"team_share_rate_store",
		"referral_reward_rate",
	}

	// 默认值
	defaults := map[string]decimal.Decimal{
		"commission_rate_single":     decimal.NewFromFloat(0.20),
		"commission_rate_multi":      decimal.NewFromFloat(0.22),
		"commission_rate_special":    decimal.NewFromFloat(0.15),
		"commission_rate_peer_single": decimal.NewFromFloat(0.10),
		"commission_rate_peer_multi":  decimal.NewFromFloat(0.12),
		"commission_rate_peer_special": decimal.NewFromFloat(0.08),
		"fund_pool_extract_rate":      decimal.NewFromFloat(0.05),
		"team_share_rate_manager":     decimal.NewFromFloat(0.03),
		"team_share_rate_store":       decimal.NewFromFloat(0.02),
		"referral_reward_rate":        decimal.NewFromFloat(0.10),
	}

	rates := make(map[string]decimal.Decimal)
	for _, key := range keys {
		val, err := s.configRepo.Get(key)
		if err != nil {
			return nil, err
		}
		if val != "" {
			d, err := decimal.NewFromString(val)
			if err != nil {
				rates[key] = defaults[key]
			} else {
				rates[key] = d
			}
		} else {
			rates[key] = defaults[key]
		}
	}

	return rates, nil
}

// GetRate 获取单个提成比例配置
func (s *ConfigService) GetRate(key string) decimal.Decimal {
	rates, err := s.GetCommissionRates()
	if err != nil {
		return decimal.Zero
	}
	if rate, ok := rates[key]; ok {
		return rate
	}
	return decimal.Zero
}
