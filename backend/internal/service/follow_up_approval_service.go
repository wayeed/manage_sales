package service

import (
	"errors"
	"fmt"

	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	"furniture-commission/internal/repository"

	"github.com/shopspring/decimal"
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
	approverID, err := s.determineApprover(customer, storeID, applicantID)
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
func (s *FollowUpApprovalService) determineApprover(customer *models.Customer, storeID int64, applicantID int64) (int64, error) {
	// 如果申请人是店长，直接由店长自己审批
	if storeID > 0 {
		store, err := s.storeRepo.FindByID(storeID)
		if err == nil && store.ManagerID != nil && *store.ManagerID == applicantID {
			return applicantID, nil
		}
	}

	// 获取原业务员信息
	if customer.CreatedBy != nil && *customer.CreatedBy > 0 {
		salesman, err := s.userRepo.FindByID(*customer.CreatedBy)
		if err == nil {
			// 如果原业务员有主管
			if salesman.ParentID != nil && *salesman.ParentID > 0 {
				// 如果申请人就是原业务员的主管，自己审批
				if *salesman.ParentID == applicantID {
					return applicantID, nil
				}
				// 否则由原业务员的主管审批
				return *salesman.ParentID, nil
			}
		}
	}

	// 无主管或原业务员不存在，由店长审批
	if storeID > 0 {
		store, err := s.storeRepo.FindByID(storeID)
		if err == nil && store.ManagerID != nil && *store.ManagerID > 0 {
			// 如果申请人就是店长，自己审批
			if *store.ManagerID == applicantID {
				return applicantID, nil
			}
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

		// 仅跟进转交审批（approval_type=1）需要切换客户归属
		// 打印审批（approval_type=2）不涉及客户归属切换
		if approval.ApprovalType == 1 && approval.CustomerID > 0 {
			if err := tx.Model(&models.Customer{}).Where("id = ?", approval.CustomerID).
				Update("salesman_id", approval.ApplicantID).Error; err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "切换客户归属失败"}
			}
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
func (s *FollowUpApprovalService) ListByApplicant(applicantID int64, status *int8, page, pageSize int) ([]models.FollowUpApproval, int64, error) {
	db := s.approvalRepo.DB().Model(&models.FollowUpApproval{}).Where("applicant_id = ?", applicantID)
	if status != nil {
		db = db.Where("status = ?", *status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []models.FollowUpApproval
	offset := (page - 1) * pageSize
	if err := db.Preload("Customer").Preload("Approver").Preload("Order").
		Offset(offset).Limit(pageSize).
		Order("id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
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
	if err := db.Preload("Customer").Preload("Applicant").Preload("Order").
		Offset(offset).Limit(pageSize).
		Order("id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// ListProcessedByApprover 查询审批人已处理的审批列表（已通过+已拒绝）
func (s *FollowUpApprovalService) ListProcessedByApprover(approverID int64, status int8, page, pageSize int) ([]models.FollowUpApproval, int64, error) {
	var list []models.FollowUpApproval
	var total int64

	db := s.db.Model(&models.FollowUpApproval{}).
		Where("approver_id = ? AND status = ?", approverID, status)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Preload("Customer").Preload("Applicant").Preload("Order").
		Offset(offset).Limit(pageSize).
		Order("id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// PrintApprovalResponse 打印审批响应
type PrintApprovalResponse struct {
	CanPrint      bool    `json:"can_print"`       // 是否可以直接打印
	NeedApproval  bool    `json:"need_approval"`   // 是否需要审批
	RemainingRate float64 `json:"remaining_rate"`  // 尾款比例
	ApprovalID    int64   `json:"approval_id"`     // 审批记录ID（如已创建）
	Message       string  `json:"message"`         // 提示信息
}

// CreatePrintApproval 创建送货单打印审批
// 尾款比例 > 20% 需要直属主管审批
func (s *FollowUpApprovalService) CreatePrintApproval(orderID, applicantID, storeID int64) (*PrintApprovalResponse, error) {
	// 1. 查询订单
	var order models.Order
	if err := s.db.First(&order, orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.NotFound, Message: "订单不存在"}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	// 2. 验证订单状态
	if order.OrderStatus != 1 {
		return nil, &AppError{Code: apperrors.BadRequest, Message: "订单状态不允许打印送货单"}
	}
	if order.DeliveryStatus != 0 {
		return nil, &AppError{Code: apperrors.BadRequest, Message: "订单已配送，无需打印送货单"}
	}

	// 3. 计算尾款比例
	finalAmount := order.FinalAmount
	remainingAmount := order.RemainingAmount
	remainingRate := decimal.Zero
	if finalAmount.GreaterThan(decimal.Zero) {
		remainingRate = remainingAmount.Div(finalAmount).Mul(decimal.NewFromInt(100))
	}

	rateFloat, _ := remainingRate.Float64()

	// 4. 尾款 <= 20%，可直接打印
	if rateFloat <= 20 {
		return &PrintApprovalResponse{
			CanPrint:      true,
			NeedApproval:  false,
			RemainingRate: rateFloat,
			Message:       "尾款比例不超过20%，可直接打印",
		}, nil
	}

	// 5. 尾款 > 20%，需要审批
	// 检查是否已有待审批的打印申请
	var existingApproval models.FollowUpApproval
	err := s.db.Where("order_id = ? AND approval_type = 2 AND status = 0", orderID).First(&existingApproval).Error
	if err == nil {
		// 已有待审批申请
		return &PrintApprovalResponse{
			CanPrint:      false,
			NeedApproval:  true,
			RemainingRate: rateFloat,
			ApprovalID:    existingApproval.ID,
			Message:       "已提交审批，等待主管审核",
		}, nil
	}

	// 检查是否已有已通过的审批
	var approvedApproval models.FollowUpApproval
	err = s.db.Where("order_id = ? AND approval_type = 2 AND status = 1", orderID).First(&approvedApproval).Error
	if err == nil {
		// 已审批通过，可以打印
		return &PrintApprovalResponse{
			CanPrint:      true,
			NeedApproval:  false,
			RemainingRate: rateFloat,
			ApprovalID:    approvedApproval.ID,
			Message:       "审批已通过，可以打印",
		}, nil
	}

	// 6. 确定审批人：业务员的直属主管
	approverID := s.determinePrintApprover(order.SalesmanID, storeID)

	// 7. 创建审批记录
	approval := &models.FollowUpApproval{
		StoreID:      storeID,
		CustomerID:   0, // 打印审批不关联客户
		OrderID:      &orderID,
		ApplicantID:  applicantID,
		ApprovalType: 2, // 送货单打印审批
		Status:       0, // 待审批
		Remark:       fmt.Sprintf("订单%s送货单打印申请，尾款比例%.1f%%", order.OrderNo, rateFloat),
	}
	if approverID > 0 {
		approval.ApproverID = &approverID
	}

	if err := s.db.Create(approval).Error; err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "创建审批申请失败"}
	}

	return &PrintApprovalResponse{
		CanPrint:      false,
		NeedApproval:  true,
		RemainingRate: rateFloat,
		ApprovalID:    approval.ID,
		Message:       "尾款比例超过20%，已提交审批申请，等待主管审核",
	}, nil
}

// determinePrintApprover 确定打印审批人：业务员的直属主管
func (s *FollowUpApprovalService) determinePrintApprover(salesmanID, storeID int64) int64 {
	if salesmanID > 0 {
		salesman, err := s.userRepo.FindByID(salesmanID)
		if err == nil && salesman.ParentID != nil && *salesman.ParentID > 0 {
			return *salesman.ParentID
		}
	}
	return 0
}

// GetPrintApprovalStatus 查询订单的打印审批状态
func (s *FollowUpApprovalService) GetPrintApprovalStatus(orderID int64) (*PrintApprovalResponse, error) {
	var order models.Order
	if err := s.db.First(&order, orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.NotFound, Message: "订单不存在"}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	// 计算尾款比例
	finalAmount := order.FinalAmount
	remainingAmount := order.RemainingAmount
	remainingRate := decimal.Zero
	if finalAmount.GreaterThan(decimal.Zero) {
		remainingRate = remainingAmount.Div(finalAmount).Mul(decimal.NewFromInt(100))
	}
	rateFloat, _ := remainingRate.Float64()

	// 查询最新的审批记录
	var approval models.FollowUpApproval
	err := s.db.Where("order_id = ? AND approval_type = 2", orderID).
		Order("id DESC").First(&approval).Error

	if err != nil {
		// 无审批记录
		if rateFloat <= 20 {
			return &PrintApprovalResponse{
				CanPrint:      true,
				NeedApproval:  false,
				RemainingRate: rateFloat,
				Message:       "尾款比例不超过20%，可直接打印",
			}, nil
		}
		return &PrintApprovalResponse{
			CanPrint:      false,
			NeedApproval:  true,
			RemainingRate: rateFloat,
			Message:       "尾款比例超过20%，需要提交审批",
		}, nil
	}

	// 有审批记录
	switch approval.Status {
	case 0: // 待审批
		return &PrintApprovalResponse{
			CanPrint:      false,
			NeedApproval:  true,
			RemainingRate: rateFloat,
			ApprovalID:    approval.ID,
			Message:       "已提交审批，等待主管审核",
		}, nil
	case 1: // 已通过
		return &PrintApprovalResponse{
			CanPrint:      true,
			NeedApproval:  false,
			RemainingRate: rateFloat,
			ApprovalID:    approval.ID,
			Message:       "审批已通过，可以打印",
		}, nil
	case 2: // 已拒绝
		return &PrintApprovalResponse{
			CanPrint:      false,
			NeedApproval:  true,
			RemainingRate: rateFloat,
			ApprovalID:    approval.ID,
			Message:       "审批被拒绝，请重新提交",
		}, nil
	}

	return &PrintApprovalResponse{
		CanPrint:      false,
		NeedApproval:  true,
		RemainingRate: rateFloat,
		Message:       "未知状态",
	}, nil
}
