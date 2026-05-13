package repository

import (
	"furniture-commission/internal/models"

	"gorm.io/gorm"
)

// SalaryRecordRepository 工资记录Repository
type SalaryRecordRepository struct {
	db *gorm.DB
}

// NewSalaryRecordRepository 创建工资记录Repository实例
func NewSalaryRecordRepository(db *gorm.DB) *SalaryRecordRepository {
	return &SalaryRecordRepository{db: db}
}

// Create 创建工资记录
func (r *SalaryRecordRepository) Create(record *models.SalaryRecord) error {
	return r.db.Create(record).Error
}

// FindByID 根据ID查找工资记录
func (r *SalaryRecordRepository) FindByID(id int64) (*models.SalaryRecord, error) {
	var record models.SalaryRecord
	err := r.db.Preload("Employee").Preload("Items").First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// FindByEmployeeAndMonth 查找员工某月工资记录
func (r *SalaryRecordRepository) FindByEmployeeAndMonth(employeeID int64, salaryMonth string) (*models.SalaryRecord, error) {
	var record models.SalaryRecord
	err := r.db.Where("employee_id = ? AND salary_month = ?", employeeID, salaryMonth).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// ListWithFilter 带条件分页查询工资列表
func (r *SalaryRecordRepository) ListWithFilter(storeID, salaryMonth, status string, page, pageSize int) ([]models.SalaryRecord, int64, error) {
	db := r.db.Model(&models.SalaryRecord{})

	if storeID != "" {
		db = db.Where("store_id = ?", storeID)
	}
	if salaryMonth != "" {
		db = db.Where("salary_month = ?", salaryMonth)
	}
	if status != "" {
		db = db.Where("status = ?", status)
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

	var records []models.SalaryRecord
	err := db.Preload("Employee").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").
		Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// UpdateFields 更新工资记录指定字段
func (r *SalaryRecordRepository) UpdateFields(id int64, fields map[string]interface{}) error {
	return r.db.Model(&models.SalaryRecord{}).Where("id = ?", id).Updates(fields).Error
}

// CreateItem 创建工资明细
func (r *SalaryRecordRepository) CreateItem(item *models.SalaryItem) error {
	return r.db.Create(item).Error
}

// BatchCreateItems 批量创建工资明细
func (r *SalaryRecordRepository) BatchCreateItems(items []models.SalaryItem) error {
	return r.db.Create(&items).Error
}

// DeleteItemsBySalaryRecordID 删除工资记录的所有明细
func (r *SalaryRecordRepository) DeleteItemsBySalaryRecordID(salaryRecordID int64) error {
	return r.db.Where("salary_record_id = ?", salaryRecordID).Delete(&models.SalaryItem{}).Error
}

// SumDeductionByEmployeeAndMonth 汇总员工某月退货冲减
func (r *SalaryRecordRepository) SumDeductionByEmployeeAndMonth(db *gorm.DB, employeeID int64, salaryMonth string) (float64, error) {
	var total float64
	err := db.Model(&models.Order{}).
		Where("salesman_id = ? AND is_returned = 1 AND order_date >= ? AND order_date < ?", employeeID, salaryMonth+"-01", salaryMonth+"-32").
		Select("COALESCE(SUM(return_profit), 0)").
		Scan(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}
