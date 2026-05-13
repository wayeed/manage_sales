package repository

import (
	"furniture-commission/internal/models"

	"gorm.io/gorm"
)

// FollowUpApprovalRepository 申请跟进审批Repository
type FollowUpApprovalRepository struct {
	db *gorm.DB
}

// NewFollowUpApprovalRepository 创建实例
func NewFollowUpApprovalRepository(db *gorm.DB) *FollowUpApprovalRepository {
	return &FollowUpApprovalRepository{db: db}
}

// Create 创建审批记录
func (r *FollowUpApprovalRepository) Create(approval *models.FollowUpApproval) error {
	return r.db.Create(approval).Error
}

// FindByID 根据ID查找
func (r *FollowUpApprovalRepository) FindByID(id int64) (*models.FollowUpApproval, error) {
	var approval models.FollowUpApproval
	err := r.db.Preload("Customer").Preload("Applicant").Preload("Approver").
		First(&approval, id).Error
	if err != nil {
		return nil, err
	}
	return &approval, nil
}

// FindPendingByCustomer 查找客户待审批记录
func (r *FollowUpApprovalRepository) FindPendingByCustomer(customerID int64) (*models.FollowUpApproval, error) {
	var approval models.FollowUpApproval
	err := r.db.Where("customer_id = ? AND status = 0", customerID).First(&approval).Error
	if err != nil {
		return nil, err
	}
	return &approval, nil
}

// Update 更新
func (r *FollowUpApprovalRepository) Update(approval *models.FollowUpApproval) error {
	return r.db.Save(approval).Error
}

// UpdateStatus 更新状态
func (r *FollowUpApprovalRepository) UpdateStatus(id int64, status int8, approverID int64, rejectReason string) error {
	updates := map[string]interface{}{
		"status":       status,
		"approver_id":  approverID,
		"reject_reason": rejectReason,
	}
	if status == 1 {
		updates["approved_at"] = gorm.Expr("NOW()")
	}
	return r.db.Model(&models.FollowUpApproval{}).Where("id = ?", id).Updates(updates).Error
}

// ListByApplicant 查询申请人的申请列表
func (r *FollowUpApprovalRepository) ListByApplicant(applicantID int64, page, pageSize int) ([]models.FollowUpApproval, int64, error) {
	db := r.db.Model(&models.FollowUpApproval{}).Where("applicant_id = ?", applicantID)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var approvals []models.FollowUpApproval
	err := db.Preload("Customer").Preload("Approver").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").
		Find(&approvals).Error
	if err != nil {
		return nil, 0, err
	}

	return approvals, total, nil
}
