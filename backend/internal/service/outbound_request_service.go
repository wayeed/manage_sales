package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	"furniture-commission/internal/repository"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// OutboundRequestService 出库申请服务
type OutboundRequestService struct {
	db        *gorm.DB
	orderRepo *repository.OrderRepository
}

// NewOutboundRequestService 创建出库申请服务实例
func NewOutboundRequestService(db *gorm.DB, orderRepo *repository.OrderRepository) *OutboundRequestService {
	return &OutboundRequestService{
		db:        db,
		orderRepo: orderRepo,
	}
}

// CreateRequest 创建出库申请
func (s *OutboundRequestService) CreateRequest(orderID int64, applicantID int64) (*models.OutboundRequest, error) {
	// 查询订单
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.ErrOrderNotFound, Message: "订单不存在"}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询订单失败"}
	}

	// 校验订单状态：order_status=1（已生效）且 delivery_status=0（未配送）
	if order.OrderStatus != 1 {
		return nil, &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "订单未审核通过，无法申请出库"}
	}
	if order.DeliveryStatus != 0 {
		return nil, &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "订单已配送或配送中，无法申请出库"}
	}

	// 校验库存状态：stock_status=0（全部有库存）
	if order.StockStatus != 0 {
		return nil, &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "订单库存不足，无法申请出库"}
	}

	// 查询是否已有出库申请（status != 3 的记录）
	var existingReq models.OutboundRequest
	if err := s.db.Where("order_id = ? AND status != ?", orderID, 3).First(&existingReq).Error; err == nil {
		return nil, &AppError{Code: apperrors.BadRequest, Message: "已有待审批的出库申请"}
	}

	// 查询业务员信息获取 applicant_name 和 parent_id（直属主管）
	var applicant models.User
	if err := s.db.First(&applicant, applicantID).Error; err != nil {
		return nil, &AppError{Code: apperrors.ErrUserNotFound, Message: "申请人不存在"}
	}

	applicantName := applicant.RealName
	if applicantName == "" {
		applicantName = applicant.Username
	}

	// 计算尾款比例
	remainingAmount := order.RemainingAmount
	finalAmount := order.FinalAmount
	var remainingRate float64
	if finalAmount.GreaterThan(decimal.Zero) {
		rate := remainingAmount.Mul(decimal.NewFromInt(100)).Div(finalAmount)
		remainingRate, _ = rate.Float64()
	}

	// 如果尾款比例 > 20%，自动生成备注
	remark := ""
	if remainingRate > 20 && remainingAmount.GreaterThan(decimal.Zero) {
		remark = fmt.Sprintf("此订单尾款还有%s元，送货后由业务员%s负责收回尾款", remainingAmount.String(), applicantName)
	}

	// 创建 OutboundRequest 记录
	// 如果已存在记录（被驳回的），则重置为重新申请
	var oldReq models.OutboundRequest
	err = s.db.Where("order_id = ?", orderID).First(&oldReq).Error
	if err == nil {
		// 已存在记录，重置为重新申请
		now := time.Now()
		if err := s.db.Model(&oldReq).Updates(map[string]interface{}{
			"status":           1,
			"applicant_id":     applicantID,
			"applicant_name":   applicantName,
			"remaining_amount": remainingAmount,
			"remaining_rate":   remainingRate,
			"remark":           remark,
			"supervisor_id":    applicant.ParentID,
			"supervisor_name":  nil,
			"supervisor_at":    nil,
			"supervisor_remark": nil,
			"finance_id":       nil,
			"finance_name":     nil,
			"finance_at":       nil,
			"finance_remark":   nil,
			"updated_at":       now,
		}).Error; err != nil {
			return nil, &AppError{Code: apperrors.InternalError, Message: "创建出库申请失败"}
		}
		return &oldReq, nil
	}

	outboundReq := &models.OutboundRequest{
		OrderID:         orderID,
		ApplicantID:     applicantID,
		ApplicantName:   applicantName,
		Status:          1, // 主管待审批
		RemainingAmount: remainingAmount,
		RemainingRate:   remainingRate,
		Remark:          remark,
		SupervisorID:    applicant.ParentID,
	}

	if err := s.db.Create(outboundReq).Error; err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "创建出库申请失败"}
	}

	return outboundReq, nil
}

// SupervisorApprove 主管审批通过
func (s *OutboundRequestService) SupervisorApprove(id int64, supervisorID int64, remark string) error {
	// 查询出库申请
	var req models.OutboundRequest
	if err := s.db.First(&req, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "出库申请不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "查询出库申请失败"}
	}

	// 校验 status=1（主管待审批）
	if req.Status != 1 {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "当前状态不允许主管审批"}
	}

	// 查询主管信息
	var supervisor models.User
	if err := s.db.First(&supervisor, supervisorID).Error; err != nil {
		return &AppError{Code: apperrors.ErrUserNotFound, Message: "审批人不存在"}
	}
	supervisorName := supervisor.RealName
	if supervisorName == "" {
		supervisorName = supervisor.Username
	}

	now := time.Now()

	// 更新 status=2（财务待审批）
	if err := s.db.Model(&req).Updates(map[string]interface{}{
		"status":            2,
		"supervisor_id":     supervisorID,
		"supervisor_name":   supervisorName,
		"supervisor_at":     now,
		"supervisor_remark": remark,
	}).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "更新出库申请失败"}
	}

	return nil
}

// FinanceApprove 财务审批通过
func (s *OutboundRequestService) FinanceApprove(id int64, financeID int64, remark string) error {
	// 查询出库申请
	var req models.OutboundRequest
	if err := s.db.First(&req, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "出库申请不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "查询出库申请失败"}
	}

	// 校验 status=2（财务待审批）
	if req.Status != 2 {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "当前状态不允许财务审批"}
	}

	// 查询财务信息
	var finance models.User
	if err := s.db.First(&finance, financeID).Error; err != nil {
		return &AppError{Code: apperrors.ErrUserNotFound, Message: "审批人不存在"}
	}
	financeName := finance.RealName
	if financeName == "" {
		financeName = finance.Username
	}

	now := time.Now()

	// 更新 status=4（已通过）
	if err := s.db.Model(&req).Updates(map[string]interface{}{
		"status":         4,
		"finance_id":     financeID,
		"finance_name":   financeName,
		"finance_at":     now,
		"finance_remark": remark,
	}).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "更新出库申请失败"}
	}

	return nil
}

// Reject 拒绝出库申请
func (s *OutboundRequestService) Reject(id int64, approverID int64, remark string) error {
	// 查询出库申请
	var req models.OutboundRequest
	if err := s.db.First(&req, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "出库申请不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "查询出库申请失败"}
	}

	// 校验 status=1 或 status=2
	if req.Status != 1 && req.Status != 2 {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "当前状态不允许拒绝"}
	}

	// 更新 status=3（已拒绝）
	if err := s.db.Model(&req).Update("status", 3).Error; err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "更新出库申请失败"}
	}

	return nil
}

// GetByOrderID 根据订单ID查询出库申请
func (s *OutboundRequestService) GetByOrderID(orderID int64) (*models.OutboundRequest, error) {
	var req models.OutboundRequest
	if err := s.db.Where("order_id = ?", orderID).Order("id DESC").First(&req).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询出库申请失败"}
	}
	return &req, nil
}

// ListPending 查询待审批的出库申请列表
func (s *OutboundRequestService) ListPending(role string, page, pageSize int) (*PageResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := s.db.Model(&models.OutboundRequest{})

	// 如果传了role参数，只返回该角色待审批的申请
	// 如果没传role参数，返回所有待审批的申请（status=1 或 status=2）
	if role != "" {
		switch strings.ToUpper(role) {
		case "SUPERVISOR":
			query = query.Where("status = ?", 1)
		case "FINANCE":
			query = query.Where("status = ?", 2)
		default:
			return nil, &AppError{Code: apperrors.BadRequest, Message: "无效的角色参数"}
		}
	} else {
		// 不传role参数时，返回所有待审批申请
		query = query.Where("status IN ?", []int{1, 2})
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询出库申请列表失败"}
	}

	var requests []models.OutboundRequest
	err := query.Preload("Order").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").
		Find(&requests).Error
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询出库申请列表失败"}
	}

	return &PageResult{
		List:     requests,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
