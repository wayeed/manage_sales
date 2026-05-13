package repository

import (
	"furniture-commission/internal/models"

	"gorm.io/gorm"
)

// FundPoolRepository 基金池Repository
type FundPoolRepository struct {
	db *gorm.DB
}

// NewFundPoolRepository 创建基金池Repository实例
func NewFundPoolRepository(db *gorm.DB) *FundPoolRepository {
	return &FundPoolRepository{db: db}
}

// Create 创建基金池
func (r *FundPoolRepository) Create(fundPool *models.FundPool) error {
	return r.db.Create(fundPool).Error
}

// FindByID 根据ID查找基金池
func (r *FundPoolRepository) FindByID(id int64) (*models.FundPool, error) {
	var fundPool models.FundPool
	err := r.db.Preload("Shares.Employee").First(&fundPool, id).Error
	if err != nil {
		return nil, err
	}
	return &fundPool, nil
}

// ListWithFilter 带条件分页查询基金池列表
func (r *FundPoolRepository) ListWithFilter(storeID string, periodType int, page, pageSize int) ([]models.FundPool, int64, error) {
	db := r.db.Model(&models.FundPool{})

	if storeID != "" {
		db = db.Where("store_id = ?", storeID)
	}
	if periodType > 0 {
		db = db.Where("period_type = ?", periodType)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var fundPools []models.FundPool
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").
		Find(&fundPools).Error
	if err != nil {
		return nil, 0, err
	}

	return fundPools, total, nil
}

// FindByPeriod 查找某周期的基金池
func (r *FundPoolRepository) FindByPeriod(storeID int64, periodType int, periodValue string) (*models.FundPool, error) {
	var fundPool models.FundPool
	err := r.db.Where("store_id = ? AND period_type = ? AND period_value = ?", storeID, periodType, periodValue).First(&fundPool).Error
	if err != nil {
		return nil, err
	}
	return &fundPool, nil
}

// UpdateFields 更新基金池指定字段
func (r *FundPoolRepository) UpdateFields(id int64, fields map[string]interface{}) error {
	return r.db.Model(&models.FundPool{}).Where("id = ?", id).Updates(fields).Error
}

// CreateShare 创建基金池份额
func (r *FundPoolRepository) CreateShare(share *models.FundPoolShare) error {
	return r.db.Create(share).Error
}

// BatchCreateShares 批量创建基金池份额
func (r *FundPoolRepository) BatchCreateShares(shares []models.FundPoolShare) error {
	return r.db.Create(&shares).Error
}

// FindSharesByFundPoolID 查找基金池份额
func (r *FundPoolRepository) FindSharesByFundPoolID(fundPoolID int64) ([]models.FundPoolShare, error) {
	var shares []models.FundPoolShare
	err := r.db.Preload("Employee").
		Where("fund_pool_id = ?", fundPoolID).
		Order("reward_amount DESC").
		Find(&shares).Error
	if err != nil {
		return nil, err
	}
	return shares, nil
}

// UpdateShareStatus 更新份额状态
func (r *FundPoolRepository) UpdateShareStatus(fundPoolID int64, status int8) error {
	return r.db.Model(&models.FundPoolShare{}).Where("fund_pool_id = ?", fundPoolID).Update("status", status).Error
}

// FindPaidSharesByEmployeeAndPeriod 查找员工某周期已发放的基金池奖励
func (r *FundPoolRepository) FindPaidSharesByEmployeeAndPeriod(employeeID int64, periodValue string) ([]models.FundPoolShare, error) {
	var shares []models.FundPoolShare
	err := r.db.Joins("JOIN fund_pools ON fund_pools.id = fund_pool_shares.fund_pool_id").
		Where("fund_pool_shares.employee_id = ? AND fund_pools.period_value = ? AND fund_pool_shares.status = 1", employeeID, periodValue).
		Find(&shares).Error
	if err != nil {
		return nil, err
	}
	return shares, nil
}
