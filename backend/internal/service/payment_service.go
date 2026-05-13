package service

import (
	"errors"
	"time"

	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	appsnow "furniture-commission/internal/pkg/snowflake"
	"furniture-commission/internal/repository"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// PaymentService 回款服务
type PaymentService struct {
	db               *gorm.DB
	paymentRepo      *repository.PaymentRepository
	orderRepo        *repository.OrderRepository
	commissionSvc    *CommissionService
}

// NewPaymentService 创建回款服务实例
func NewPaymentService(db *gorm.DB, paymentRepo *repository.PaymentRepository, orderRepo *repository.OrderRepository, commissionSvc *CommissionService) *PaymentService {
	return &PaymentService{
		db:            db,
		paymentRepo:   paymentRepo,
		orderRepo:     orderRepo,
		commissionSvc: commissionSvc,
	}
}

// CreatePaymentRequest 创建回款请求
type CreatePaymentRequest struct {
	OrderID int64 `json:"order_id" binding:"required" example:1`
	Amount string `json:"amount" binding:"required" example:"8500.00"`
	PaymentDate string `json:"payment_date" example:"2024-01-20"`
	PaymentMethod int8 `json:"payment_method" example:1`
	Remark string `json:"remark" example:"客户银行转账"`
}

// ListPaymentRequest 回款列表查询请求
type ListPaymentRequest struct {
	OrderID string `form:"order_id" example:"1"`
	Status string `form:"status" example:"0"`
	StartDate string `form:"start_date" example:"2024-01-01"`
	EndDate string `form:"end_date" example:"2024-12-31"`
	Page int `form:"page" example:1`
	PageSize int `form:"page_size" example:10`
}

// CreatePayment 录入回款
func (s *PaymentService) CreatePayment(req *CreatePaymentRequest, createdBy int64) error {
	// 1. 验证订单存在且已生效(order_status=1)
	order, err := s.orderRepo.FindByID(req.OrderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrOrderNotFound, Message: "订单不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if order.OrderStatus != 1 {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "只有已生效的订单才能录入回款"}
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return &AppError{Code: apperrors.BadRequest, Message: "回款金额格式错误"}
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return &AppError{Code: apperrors.BadRequest, Message: "回款金额必须大于0"}
	}

	// 2. 创建回款记录
	paymentNo := "PAY" + appsnow.GenerateOrderNo()

	var paymentDate *time.Time
	if req.PaymentDate != "" {
		pd, err := time.Parse("2006-01-02", req.PaymentDate)
		if err == nil {
			paymentDate = &pd
		}
	}
	if paymentDate == nil {
		now := time.Now()
		paymentDate = &now
	}

	payment := &models.Payment{
		OrderID:       req.OrderID,
		PaymentNo:     paymentNo,
		Amount:        amount,
		PaymentDate:   paymentDate,
		PaymentMethod: req.PaymentMethod,
		Status:        0, // 待审核
		Remark:        req.Remark,
		CreatedBy:     &createdBy,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(payment).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "创建回款记录失败"}
		}
		return nil
	})

	if err != nil {
		if appErr, ok := err.(*AppError); ok {
			return appErr
		}
		return &AppError{Code: apperrors.InternalError, Message: "录入回款失败"}
	}

	return nil
}

// ApprovePayment 审核回款
func (s *PaymentService) ApprovePayment(paymentID int64, approvedBy int64, approved bool) error {
	payment, err := s.paymentRepo.FindByID(paymentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrOrderNotFound, Message: "回款记录不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if payment.Status != 0 {
		return &AppError{Code: apperrors.ErrInvalidOrderStatus, Message: "回款记录已审核"}
	}

	if approved {
		// 审核通过
		err = s.db.Transaction(func(tx *gorm.DB) error {
			now := time.Now()

			// 更新回款状态
			if err := tx.Model(payment).Updates(map[string]interface{}{
				"status":     1, // 已审核
				"audited_by": approvedBy,
				"audited_at": now,
			}).Error; err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "更新回款状态失败"}
			}

			// 更新订单paid_amount
			if err := tx.Model(&models.Order{}).Where("id = ?", payment.OrderID).
				Update("paid_amount", gorm.Expr("paid_amount + ?", payment.Amount)).Error; err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "更新订单回款金额失败"}
			}

			// 重新查询订单判断回款状态
			var order models.Order
			if err := tx.First(&order, payment.OrderID).Error; err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "查询订单失败"}
			}

			// 判断回款状态
			paymentStatus := int8(0)
			if order.PaidAmount.GreaterThanOrEqual(order.FinalAmount) {
				paymentStatus = 2 // 已回款
			} else if order.PaidAmount.GreaterThan(decimal.Zero) {
				paymentStatus = 1 // 部分回款
			}

			remainingAmount := order.FinalAmount.Sub(order.PaidAmount)
			if remainingAmount.IsNegative() {
				remainingAmount = decimal.Zero
			}

			if err := tx.Model(&models.Order{}).Where("id = ?", payment.OrderID).Updates(map[string]interface{}{
			"payment_status":  paymentStatus,
			"remaining_amount": remainingAmount,
		}).Error; err != nil {
			return &AppError{Code: apperrors.InternalError, Message: "更新订单回款状态失败"}
		}

		// 如果回款完成，异步计算提成
		if paymentStatus == 2 {
			go func(orderID int64) {
				if err := s.commissionSvc.CalculateOrderCommission(orderID); err != nil {
					// 记录错误但不影响回款审核结果
					// 实际项目中应该使用日志记录或使用消息队列重试
					_ = err
				}
			}(payment.OrderID)
		}

		return nil
		})
	} else {
		// 审核驳回
		now := time.Now()
		err = s.db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(payment).Updates(map[string]interface{}{
				"status":     2, // 已驳回
				"audited_by": approvedBy,
				"audited_at": now,
			}).Error; err != nil {
				return &AppError{Code: apperrors.InternalError, Message: "更新回款状态失败"}
			}
			return nil
		})
	}

	if err != nil {
		if appErr, ok := err.(*AppError); ok {
			return appErr
		}
		return &AppError{Code: apperrors.InternalError, Message: "审核回款失败"}
	}

	return nil
}

// List 回款列表
func (s *PaymentService) List(req *ListPaymentRequest) (*PageResult, error) {
	payments, total, err := s.paymentRepo.ListWithFilter(
		req.OrderID, req.Status, req.StartDate, req.EndDate,
		req.Page, req.PageSize,
	)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询回款列表失败"}
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
		List:     payments,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetByOrderID 根据订单ID获取回款记录
func (s *PaymentService) GetByOrderID(orderID int64) ([]models.Payment, error) {
	payments, err := s.paymentRepo.FindByOrderID(orderID)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询回款记录失败"}
	}
	return payments, nil
}
