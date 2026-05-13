package repository

import (
	"fmt"
	"furniture-commission/internal/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// CommissionRepository 提成Repository
type CommissionRepository struct {
	db *gorm.DB
}

// NewCommissionRepository 创建提成Repository实例
func NewCommissionRepository(db *gorm.DB) *CommissionRepository {
	return &CommissionRepository{db: db}
}

// Create 创建提成记录
func (r *CommissionRepository) Create(commission *models.Commission) error {
	return r.db.Create(commission).Error
}

// BatchCreate 批量创建提成记录
func (r *CommissionRepository) BatchCreate(commissions []models.Commission) error {
	return r.db.Create(&commissions).Error
}

// FindByID 根据ID查找提成记录
func (r *CommissionRepository) FindByID(id int64) (*models.Commission, error) {
	var commission models.Commission
	err := r.db.Preload("Order").Preload("Employee").Preload("Peer").First(&commission, id).Error
	if err != nil {
		return nil, err
	}
	return &commission, nil
}

// FindByOrderID 根据订单ID查找提成记录
func (r *CommissionRepository) FindByOrderID(orderID int64) ([]models.Commission, error) {
	var commissions []models.Commission
	err := r.db.Preload("Employee").Preload("Peer").
		Where("order_id = ?", orderID).
		Order("commission_type ASC").
		Find(&commissions).Error
	if err != nil {
		return nil, err
	}
	return commissions, nil
}

// ListWithFilter 带条件分页查询提成列表
func (r *CommissionRepository) ListWithFilter(storeID, employeeID, commissionType, status, periodValue string, page, pageSize int) ([]models.Commission, int64, error) {
	db := r.db.Model(&models.Commission{})

	if storeID != "" {
		db = db.Where("store_id = ?", storeID)
	}
	if employeeID != "" {
		db = db.Where("employee_id = ?", employeeID)
	}
	if commissionType != "" {
		db = db.Where("commission_type = ?", commissionType)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if periodValue != "" {
		db = db.Where("period_value = ?", periodValue)
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

	var commissions []models.Commission
	err := db.Preload("Order").Preload("Employee").Preload("Peer").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").
		Find(&commissions).Error
	if err != nil {
		return nil, 0, err
	}

	return commissions, total, nil
}

// SumByEmployeeAndPeriod 汇总员工某周期的提成
func (r *CommissionRepository) SumByEmployeeAndPeriod(employeeID int64, periodValue string, commissionTypes []int8) (decimal.Decimal, error) {
	var total decimal.Decimal
	db := r.db.Model(&models.Commission{}).
		Where("employee_id = ? AND period_value = ? AND status = 1", employeeID, periodValue)
	if len(commissionTypes) > 0 {
		db = db.Where("commission_type IN ?", commissionTypes)
	}
	err := db.Select("COALESCE(SUM(amount), 0)").Scan(&total).Error
	if err != nil {
		return decimal.Zero, err
	}
	return total, nil
}

// SumByPeriodAndType 汇总某周期某类型的提成总额
func (r *CommissionRepository) SumByPeriodAndType(storeID int64, periodValue string, commissionType int8) (decimal.Decimal, error) {
	var total decimal.Decimal
	err := r.db.Model(&models.Commission{}).
		Where("store_id = ? AND period_value = ? AND commission_type = ?", storeID, periodValue, commissionType).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	if err != nil {
		return decimal.Zero, err
	}
	return total, nil
}

// GetSummary 提成汇总
func (r *CommissionRepository) GetSummary(employeeID int64, startDate, endDate string) (map[string]decimal.Decimal, error) {
	type SummaryResult struct {
		CommissionType int8            `gorm:"column:commission_type"`
		TotalAmount    decimal.Decimal `gorm:"column:total_amount"`
	}
	var results []SummaryResult
	db := r.db.Model(&models.Commission{}).
		Where("employee_id = ?", employeeID)
	if startDate != "" {
		db = db.Where("period_value >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("period_value <= ?", endDate)
	}
	err := db.Select("commission_type, COALESCE(SUM(amount), 0) as total_amount").
		Group("commission_type").
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	summary := make(map[string]decimal.Decimal)
	for _, r := range results {
		summary[fmt.Sprintf("%d", r.CommissionType)] = r.TotalAmount
	}
	return summary, nil
}

// UpdateFields 更新提成指定字段
func (r *CommissionRepository) UpdateFields(id int64, fields map[string]interface{}) error {
	return r.db.Model(&models.Commission{}).Where("id = ?", id).Updates(fields).Error
}

// UpdateStatusByIDs 批量更新提成状态
func (r *CommissionRepository) UpdateStatusByIDs(ids []int64, status int8) error {
	return r.db.Model(&models.Commission{}).Where("id IN ?", ids).Update("status", status).Error
}

// ExistsByOrderID 检查订单是否已有提成记录
func (r *CommissionRepository) ExistsByOrderID(orderID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.Commission{}).Where("order_id = ?", orderID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// FindByPeriod 按周期查询提成记录
func (r *CommissionRepository) FindByPeriod(periodValue string) ([]models.Commission, error) {
	var commissions []models.Commission
	err := r.db.Where("period_value = ?", periodValue).Find(&commissions).Error
	return commissions, err
}
