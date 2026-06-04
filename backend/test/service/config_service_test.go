package service_test

import (
	"testing"

	"furniture-commission/internal/models"
	"furniture-commission/internal/repository"
	svc "furniture-commission/internal/service"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupConfigTestDB 创建配置测试数据库
func setupConfigTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&models.SystemConfig{})
	assert.NoError(t, err)
	return db
}

// ========== TestGetConfig ==========

func TestGetConfig(t *testing.T) {
	db := setupConfigTestDB(t)

	configRepo := repository.NewSystemConfigRepository(db)
	configSvc := svc.NewConfigService(configRepo)

	// 先设置配置
	err := configSvc.Set("test_key", "test_value", "string", "测试配置")
	assert.NoError(t, err)

	// 获取配置
	val, err := configSvc.Get("test_key")
	assert.NoError(t, err)
	assert.Equal(t, "test_value", val)
}

func TestGetConfig_NotFound(t *testing.T) {
	db := setupConfigTestDB(t)

	configRepo := repository.NewSystemConfigRepository(db)
	configSvc := svc.NewConfigService(configRepo)

	val, err := configSvc.Get("nonexistent_key")
	assert.NoError(t, err)
	assert.Equal(t, "", val)
}

// ========== TestSetConfig ==========

func TestSetConfig(t *testing.T) {
	db := setupConfigTestDB(t)

	configRepo := repository.NewSystemConfigRepository(db)
	configSvc := svc.NewConfigService(configRepo)

	// 设置新配置
	err := configSvc.Set("site_name", "家具提成系统", "string", "站点名称")
	assert.NoError(t, err)

	// 验证
	val, _ := configSvc.Get("site_name")
	assert.Equal(t, "家具提成系统", val)
}

func TestSetConfig_UpdateExisting(t *testing.T) {
	db := setupConfigTestDB(t)

	configRepo := repository.NewSystemConfigRepository(db)
	configSvc := svc.NewConfigService(configRepo)

	// 设置配置
	configSvc.Set("theme", "dark", "string", "主题")

	// 更新配置
	err := configSvc.Set("theme", "light", "string", "主题")
	assert.NoError(t, err)

	// 验证已更新
	val, _ := configSvc.Get("theme")
	assert.Equal(t, "light", val)
}

func TestSetConfig_DecimalValue(t *testing.T) {
	db := setupConfigTestDB(t)

	configRepo := repository.NewSystemConfigRepository(db)
	configSvc := svc.NewConfigService(configRepo)

	// 设置小数配置
	err := configSvc.Set("commission_rate_level1_single", "0.08", "decimal", "初级单品提成比例")
	assert.NoError(t, err)

	val, _ := configSvc.Get("commission_rate_level1_single")
	assert.Equal(t, "0.08", val)
}

// ========== TestGetCommissionRates ==========

func TestGetCommissionRates_Defaults(t *testing.T) {
	db := setupConfigTestDB(t)

	configRepo := repository.NewSystemConfigRepository(db)
	configSvc := svc.NewConfigService(configRepo)

	// 未设置任何配置时，应返回默认值
	rates, err := configSvc.GetCommissionRates()
	assert.NoError(t, err)
	assert.NotNil(t, rates)

	// 验证默认值
	assert.True(t, rates["commission_rate_level1_single"].Equal(decimal.NewFromFloat(0.08)))
	assert.True(t, rates["commission_rate_level1_multi"].Equal(decimal.NewFromFloat(0.10)))
	assert.True(t, rates["commission_rate_level2_single"].Equal(decimal.NewFromFloat(0.18)))
	assert.True(t, rates["commission_rate_level2_multi"].Equal(decimal.NewFromFloat(0.22)))
	assert.True(t, rates["commission_rate_level3_single"].Equal(decimal.NewFromFloat(0.35)))
	assert.True(t, rates["commission_rate_level3_multi"].Equal(decimal.NewFromFloat(0.38)))
	assert.True(t, rates["commission_rate_peer_single"].Equal(decimal.NewFromFloat(0.10)))
	assert.True(t, rates["commission_rate_peer_multi"].Equal(decimal.NewFromFloat(0.12)))
	assert.True(t, rates["commission_rate_peer_special"].Equal(decimal.NewFromFloat(0.08)))
	assert.True(t, rates["fund_pool_extract_rate"].Equal(decimal.NewFromFloat(0.05)))
	assert.True(t, rates["team_share_rate_manager"].Equal(decimal.NewFromFloat(0.03)))
	assert.True(t, rates["team_share_rate_store"].Equal(decimal.NewFromFloat(0.02)))
	assert.True(t, rates["referral_reward_rate"].Equal(decimal.NewFromFloat(0.10)))
}

func TestGetCommissionRates_CustomValues(t *testing.T) {
	db := setupConfigTestDB(t)

	configRepo := repository.NewSystemConfigRepository(db)
	configSvc := svc.NewConfigService(configRepo)

	// 自定义提成比例
	configSvc.Set("commission_rate_level1_single", "0.10", "decimal", "初级单品提成比例")
	configSvc.Set("commission_rate_level1_multi", "0.12", "decimal", "初级多品提成比例")

	rates, err := configSvc.GetCommissionRates()
	assert.NoError(t, err)

	assert.True(t, rates["commission_rate_level1_single"].Equal(decimal.NewFromFloat(0.10)))
	assert.True(t, rates["commission_rate_level1_multi"].Equal(decimal.NewFromFloat(0.12)))
	// 未设置的仍使用默认值
	assert.True(t, rates["commission_rate_level2_single"].Equal(decimal.NewFromFloat(0.18)))
}

// ========== TestGetAllConfigs ==========

func TestGetAllConfigs(t *testing.T) {
	db := setupConfigTestDB(t)

	configRepo := repository.NewSystemConfigRepository(db)
	configSvc := svc.NewConfigService(configRepo)

	// 设置多个配置
	configSvc.Set("key1", "value1", "string", "配置1")
	configSvc.Set("key2", "value2", "string", "配置2")
	configSvc.Set("key3", "0.15", "decimal", "配置3")

	configs, err := configSvc.GetAll()
	assert.NoError(t, err)
	assert.Len(t, configs, 3)
}

func TestGetAllConfigs_Empty(t *testing.T) {
	db := setupConfigTestDB(t)

	configRepo := repository.NewSystemConfigRepository(db)
	configSvc := svc.NewConfigService(configRepo)

	configs, err := configSvc.GetAll()
	assert.NoError(t, err)
	assert.Empty(t, configs)
}

// ========== TestGetRate ==========

func TestGetRate(t *testing.T) {
	db := setupConfigTestDB(t)

	configRepo := repository.NewSystemConfigRepository(db)
	configSvc := svc.NewConfigService(configRepo)

	// 使用默认值
	rate := configSvc.GetRate("commission_rate_level1_single")
	assert.True(t, rate.Equal(decimal.NewFromFloat(0.08)))

	// 不存在的key返回0
	rate = configSvc.GetRate("nonexistent_rate")
	assert.True(t, rate.Equal(decimal.Zero))
}
