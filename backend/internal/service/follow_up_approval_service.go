package service

import (
	"errors"

	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	"furniture-commission/internal/repository"

	"gorm.io/gorm"
)

// CreateFollowUpApprovalRequest 申请跟进请求
type CreateFollowUpApprovalRequest struct {
	CustomerID int64 `json:"customer_id" binding:"required" example:1`
	Remark string `json:"remark" example:"客户长时间未跟进，申请重新分配"`
}

// FollowUpApprovalService 申请跟进审批服务
type FollowUpApprovalService struct {
	db           *gorm.DB
	approvalRepo *repository.FollowUpApprovalRepository
	customerRepo *repository.CustomerRepository
	userRepo     *repository.UserRepository
	storeRepo    *repository.StoreRepository
}

// NewFollowUpApprovalService 创建实例
func NewFollowUpApprovalService(db *gorm.DB, approvalRepo *repository.FollowUpApprovalRepository, customerRepo *repository.CustomerRepository, userRepo *repository.UserRepository, storeRepo *repository.StoreRepository) *FollowUpApprovalService {
	return &FollowUpApprovalService{
		db:           db,
		approvalRepo: approvalRepo,
		customerRepo: customerRepo,
		userRepo:     userRepo,
		storeRepo:    storeRepo,
	}
}

// Create 创建申请
func (s *FollowUpApprovalService) Create(req *CreateFollowUpApprovalRequest, applicantID, storeID int64) (*models.FollowUpApproval, error) {
	// 检查客户是否存在
	customer, err := s.customerRepo.FindByID(req.CustomerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.NotFound, Message: "客户不存在"}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	// 检查是否已有待审批申请
	existing, _ := s.approvalRepo.FindPendingByCustomer(req.CustomerID)
	if existing != nil {
		return nil, &AppError{Code: apperrors.ErrDuplicateKey, Message: "该客户已有待审批的跟进申请"}
	}

	// 确定审批人
	approverID, err := s.determineApprover(customer, storeID)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: err.Error()}
	}

	approval := &models.FollowUpApproval{
		StoreID:     storeID,
		CustomerID:  req.CustomerID,
		ApplicantID: applicantID,
		Status:      0, // 待审批
		Remark:      req.Remark,
	}

	// approverID 为 0 时设为 nil，避免关联查询异常
	if approverID > 0 {
		approval.ApproverID = &approverID
	}

	if err := s.approvalRepo.Create(approval); err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "创建申请失败"}
	}

	return approval, nil
}

// determineApprover 确定审批人：原业务员的主管或店长
func (s *FollowUpApprovalService) determineApprover(customer *models.Customer, storeID int64) (int64, error) {
	// 获取原业务员信息
	if customer.CreatedBy != nil && *customer.CreatedBy > 0 {
		salesman, err := s.userRepo.FindByID(*customer.CreatedBy)
		if err == nil && salesman.ParentID != nil && *salesman.ParentID > 0 {
			// 有主管，由主管审批
			return *salesman.ParentID, nil
		}
	}

	// 无主管或原业务员不存在，由店长审批
	if storeID > 0 {
		store, err := s.storeRepo.FindByID(storeID)
		if err == nil && store.ManagerID != nil && *store.ManagerID > 0 {
			return *store.ManagerID, nil
		}
	}

	// 未找到审批人，返回0表示待分配
	return 0, nil
}

// GetStatus 查询审批状态
func (s *FollowUpApprovalService) GetStatus(id int64) (*models.FollowUpApproval, error) {
	approval, err := s.approvalRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.NotFound, Message: "审批记录不存在"}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}
	return approval, nil
}

// Approve 审批通过（事务保证原子性）
func (s *FollowUpApprovalService) Approve(id, approverID int64) error {
	approval, err := s.approvalRepo.FindByID(id)
	if err != nil {
		return &AppError{Code: apperrors.NotFound, Message: "审批记录不存在"}
	}

	if approval.Status != 0 {
		return &AppError{Code: apperrors.BadRequest, Message: "该申请已处理"}
	}

	// 使用事务保证原子性
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 更新审批状态
		if err := tx.Model(&models.FollowUpApproval{}).Where("id = ?", id).
			Updates(map[string]interface{}{
				"status":       1,
				"approver_id":  approverID,
				"approved_at":  gorm.Expr("NOW()"),
				"reject_reason": "",
			}).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "审批失败"}
		}

		// 切换客户归属：更新负责业务员
		if err := tx.Model(&models.Customer{}).Where("id = ?", approval.CustomerID).
			Update("salesman_id", approval.ApplicantID).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "切换客户归属失败"}
		}

		return nil
	})
}

// Reject 审批拒绝
func (s *FollowUpApprovalService) Reject(id, approverID int64, reason string) error {
	approval, err := s.approvalRepo.FindByID(id)
	if err != nil {
		return &AppError{Code: apperrors.NotFound, Message: "审批记录不存在"}
	}

	if approval.Status != 0 {
		return &AppError{Code: apperrors.BadRequest, Message: "该申请已处理"}
	}

	if err := s.approvalRepo.UpdateStatus(id, 2, approverID, reason); err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "审批失败"}
	}

	return nil
}

// Cancel 撤回申请
func (s *FollowUpApprovalService) Cancel(id, applicantID int64) error {
	approval, err := s.approvalRepo.FindByID(id)
	if err != nil {
		return &AppError{Code: apperrors.NotFound, Message: "审批记录不存在"}
	}

	if approval.ApplicantID != applicantID {
		return &AppError{Code: apperrors.Forbidden, Message: "只能撤回自己的申请"}
	}

	if approval.Status != 0 {
		return &AppError{Code: apperrors.BadRequest, Message: "只能撤回待审批的申请"}
	}

	// 逻辑删除
	if err := s.db.Delete(approval).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "撤回申请失败"}
	}

	return nil
}

// ListByApplicant 查询申请人的申请列表
func (s *FollowUpApprovalService) ListByApplicant(applicantID int64, page, pageSize int) ([]models.FollowUpApproval, int64, error) {
	return s.approvalRepo.ListByApplicant(applicantID, page, pageSize)
}

// ListByApprover 查询审批人的待审批列表
func (s *FollowUpApprovalService) ListByApprover(approverID int64, page, pageSize int) ([]models.FollowUpApproval, int64, error) {
	var list []models.FollowUpApproval
	var total int64

	db := s.db.Model(&models.FollowUpApproval{}).Where("status = 0")

	// 查询分配给该审批人的，或未分配审批人的（approver_id IS NULL OR approver_id = 0）
	db = db.Where("(approver_id = ? OR approver_id IS NULL OR approver_id = 0)", approverID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Preload("Customer").Preload("Applicant").
		Offset(offset).Limit(pageSize).
		Order("id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
