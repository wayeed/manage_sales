package service

import (
	"errors"
	"fmt"
	"time"

	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	"furniture-commission/internal/repository"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ListSalaryRequest 工资列表查询请求
type ListSalaryRequest struct {
	StoreID string `form:"store_id" example:"1"`
	SalaryMonth string `form:"salary_month" example:"2024-01"`
	Status string `form:"status" example:"0"`
	Page int `form:"page" example:1`
	PageSize int `form:"page_size" example:10`
}

// PaySalaryRequest 发放工资请求
type PaySalaryRequest struct {
	SalaryRecordID int64 `json:"salary_record_id" binding:"required" example:1`
	PayMethod int `json:"pay_method" example:1`
	PayRemark string `json:"pay_remark" example:"银行转账发放"`
}

// SalaryDetailVO 工资详情视图
type SalaryDetailVO struct {
	*models.SalaryRecord
	Items []models.SalaryItem `json:"items" example:[]`
}

// SalaryService 工资服务
type SalaryService struct {
	db             *gorm.DB
	salaryRepo     *repository.SalaryRecordRepository
	commissionRepo *repository.CommissionRepository
	fundPoolRepo   *repository.FundPoolRepository
}

// NewSalaryService 创建工资服务实例
func NewSalaryService(
	db *gorm.DB,
	salaryRepo *repository.SalaryRecordRepository,
	commissionRepo *repository.CommissionRepository,
	fundPoolRepo *repository.FundPoolRepository,
) *SalaryService {
	return &SalaryService{
		db:             db,
		salaryRepo:     salaryRepo,
		commissionRepo: commissionRepo,
		fundPoolRepo:   fundPoolRepo,
	}
}

// GenerateSalary 生成月度工资
func (s *SalaryService) GenerateSalary(storeID int64, salaryMonth string) error {
	// 1. 查找该门店所有在职员工（status=1）
	var employees []models.User
	err := s.db.Where("store_id = ? AND status = 1", storeID).Find(&employees).Error
	if err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "查询员工列表失败"}
	}

	if len(employees) == 0 {
		return &AppError{Code: apperrors.BadRequest, Message: "该门店无在职员工"}
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, employee := range employees {
			// 检查是否已生成工资记录
			var existing models.SalaryRecord
			err := tx.Where("employee_id = ? AND salary_month = ?", employee.ID, salaryMonth).First(&existing).Error
			if err == nil {
				continue // 已存在则跳过
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return &AppError{Code: apperrors.InternalError, Message: "检查工资记录失败"}
			}

			// 2. 计算各项金额
			// a. base_salary
			baseSalary := decimal.NewFromFloat(employee.BaseSalary)

			// b. sales_commission = SUM(commissions WHERE employee_id=? AND commission_type=1 AND period_value=salaryMonth AND status=1)
			salesCommission, _ := s.commissionRepo.SumByEmployeeAndPeriod(employee.ID, salaryMonth, []int8{1})

			// c. team_commission = SUM(commissions WHERE commission_type IN (3,4) AND employee_id=? AND period_value=salaryMonth AND status=1)
			teamCommission, _ := s.commissionRepo.SumByEmployeeAndPeriod(employee.ID, salaryMonth, []int8{3, 4})

			// d. fund_reward = SUM(fund_pool_shares WHERE employee_id=? AND status=1)
			fundShares, err := s.fundPoolRepo.FindPaidSharesByEmployeeAndPeriod(employee.ID, salaryMonth)
			fundReward := decimal.Zero
			if err == nil {
				for _, share := range fundShares {
					fundReward = fundReward.Add(share.RewardAmount)
				}
			}

			// e. referral_reward = SUM(commissions WHERE commission_type=6 AND employee_id=? AND period_value=salaryMonth AND status=1)
			referralReward, _ := s.commissionRepo.SumByEmployeeAndPeriod(employee.ID, salaryMonth, []int8{6})

			// f. deduction = 退货冲减
			var deduction float64
			// 计算该月第一天和下月第一天
			monthStart := salaryMonth + "-01"
			nextMonth := time.Date(
				time.Now().Year(), time.Now().Month()+1, 1, 0, 0, 0, 0, time.Local,
			)
			if t, err := time.Parse("2006-01", salaryMonth); err == nil {
				nextMonth = t.AddDate(0, 1, 0)
			}
			monthEnd := nextMonth.Format("2006-01-02")

			tx.Model(&models.Order{}).
				Where("salesman_id = ? AND is_returned = 1 AND order_date >= ? AND order_date < ?", employee.ID, monthStart, monthEnd).
				Select("COALESCE(SUM(return_profit), 0)").
				Scan(&deduction)

			deductionDecimal := decimal.NewFromFloat(deduction)

			// g. gross_salary
			bonus := decimal.Zero
			grossSalary := baseSalary.Add(salesCommission).Add(teamCommission).Add(fundReward).Add(referralReward).Add(bonus).Sub(deductionDecimal)

			// h. net_salary
			netSalary := grossSalary

			// 3. 创建salary_records记录
			salaryRecord := &models.SalaryRecord{
				StoreID:         storeID,
				EmployeeID:      employee.ID,
				SalaryMonth:     salaryMonth,
				BaseSalary:      baseSalary,
				SalesCommission: salesCommission,
				TeamCommission:  teamCommission,
				FundReward:      fundReward,
				ReferralReward:  referralReward,
				Deduction:       deductionDecimal,
				Bonus:           bonus,
				GrossSalary:     grossSalary,
				NetSalary:       netSalary,
				Status:          0, // 草稿
			}

			if err := tx.Create(salaryRecord).Error; err != nil {
				return &AppError{Code: apperrors.InternalError, Message: fmt.Sprintf("创建员工%s工资记录失败", employee.RealName)}
			}

			// 4. 创建salary_items明细记录
			var items []models.SalaryItem

			if baseSalary.GreaterThan(decimal.Zero) {
				items = append(items, models.SalaryItem{
					SalaryRecordID: salaryRecord.ID,
					ItemType:       1,
					ItemName:       "基本工资",
					Amount:         baseSalary.InexactFloat64(),
				})
			}
			if salesCommission.GreaterThan(decimal.Zero) {
				items = append(items, models.SalaryItem{
					SalaryRecordID: salaryRecord.ID,
					ItemType:       2,
					ItemName:       "销售提成",
					Amount:         salesCommission.InexactFloat64(),
				})
			}
			if teamCommission.GreaterThan(decimal.Zero) {
				items = append(items, models.SalaryItem{
					SalaryRecordID: salaryRecord.ID,
					ItemType:       3,
					ItemName:       "团队分润",
					Amount:         teamCommission.InexactFloat64(),
				})
			}
			if fundReward.GreaterThan(decimal.Zero) {
				items = append(items, models.SalaryItem{
					SalaryRecordID: salaryRecord.ID,
					ItemType:       4,
					ItemName:       "基金池奖励",
					Amount:         fundReward.InexactFloat64(),
				})
			}
			if referralReward.GreaterThan(decimal.Zero) {
				items = append(items, models.SalaryItem{
					SalaryRecordID: salaryRecord.ID,
					ItemType:       5,
					ItemName:       "老带新奖励",
					Amount:         referralReward.InexactFloat64(),
				})
			}
			if deductionDecimal.GreaterThan(decimal.Zero) {
				items = append(items, models.SalaryItem{
					SalaryRecordID: salaryRecord.ID,
					ItemType:       6,
					ItemName:       "退货冲减",
					Amount:         -deductionDecimal.InexactFloat64(),
				})
			}
			if bonus.GreaterThan(decimal.Zero) {
				items = append(items, models.SalaryItem{
					SalaryRecordID: salaryRecord.ID,
					ItemType:       7,
					ItemName:       "奖金",
					Amount:         bonus.InexactFloat64(),
				})
			}

			if len(items) > 0 {
				if err := tx.Create(&items).Error; err != nil {
					return &AppError{Code: apperrors.InternalError, Message: "创建工资明细失败"}
				}
			}
		}

		return nil
	})
}

// ConfirmSalary 审核确认工资（status 0->1）
func (s *SalaryService) ConfirmSalary(salaryRecordID int64, confirmedBy int64) error {
	record, err := s.salaryRepo.FindByID(salaryRecordID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "工资记录不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "查询工资记录失败"}
	}

	if record.Status != 0 {
		return &AppError{Code: apperrors.BadRequest, Message: "只有草稿状态的工资才能确认"}
	}

	now := time.Now()
	return s.salaryRepo.UpdateFields(salaryRecordID, map[string]interface{}{
		"status":       1,
		"confirmed_by": confirmedBy,
		"confirmed_at": now,
	})
}

// PaySalary 发放工资（status 1->2）
// 同时更新相关commission记录状态为2(已发放)
func (s *SalaryService) PaySalary(salaryRecordID int64, paidBy int64, payMethod int, payRemark string) error {
	record, err := s.salaryRepo.FindByID(salaryRecordID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "工资记录不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "查询工资记录失败"}
	}

	if record.Status != 1 {
		return &AppError{Code: apperrors.BadRequest, Message: "只有已确认的工资才能发放"}
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		// 更新工资记录状态
		if err := tx.Model(&models.SalaryRecord{}).Where("id = ?", salaryRecordID).Updates(map[string]interface{}{
			"status":     2,
			"paid_amount": record.NetSalary,
			"paid_at":    now,
			"paid_by":    paidBy,
			"pay_method": payMethod,
			"pay_remark": payRemark,
		}).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "更新工资状态失败"}
		}

		// 更新相关commission记录状态为2(已发放)
		// 查找该员工该月的所有可发放提成
		var commissionIDs []int64
		tx.Model(&models.Commission{}).
			Where("employee_id = ? AND period_value = ? AND status = 1", record.EmployeeID, record.SalaryMonth).
			Pluck("id", &commissionIDs)

		if len(commissionIDs) > 0 {
			if err := tx.Model(&models.Commission{}).
				Where("id IN ?", commissionIDs).
				Updates(map[string]interface{}{
					"status":     2,
					"settled_at": now,
				}).Error; err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "更新提成状态失败"}
			}
		}

		return nil
	})
}

// List 工资列表
func (s *SalaryService) List(req *ListSalaryRequest) (*PageResult, error) {
	records, total, err := s.salaryRepo.ListWithFilter(req.StoreID, req.SalaryMonth, req.Status, req.Page, req.PageSize)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询工资列表失败"}
	}

	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	return &PageResult{
		List:     records,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetDetail 工资详情（含明细）
func (s *SalaryService) GetDetail(salaryRecordID int64) (*SalaryDetailVO, error) {
	record, err := s.salaryRepo.FindByID(salaryRecordID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.NotFound, Message: "工资记录不存在"}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询工资记录失败"}
	}

	return &SalaryDetailVO{
		SalaryRecord: record,
		Items:        record.Items,
	}, nil
}

// GetEmployeeSalary 员工月度工资
func (s *SalaryService) GetEmployeeSalary(employeeID int64, salaryMonth string) (*models.SalaryRecord, error) {
	record, err := s.salaryRepo.FindByEmployeeAndMonth(employeeID, salaryMonth)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.NotFound, Message: "工资记录不存在"}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询工资记录失败"}
	}
	return record, nil
}
