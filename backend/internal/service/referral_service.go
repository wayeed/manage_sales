package service

import (
	"errors"
	"time"

	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	"furniture-commission/internal/repository"

	"gorm.io/gorm"
)

// CreateReferralRequest 创建引荐关系请求
type CreateReferralRequest struct {
	ReferrerID int64 `json:"referrer_id" binding:"required" example:1`
	ReferredID int64 `json:"referred_id" binding:"required" example:2`
	RewardRate float64 `json:"reward_rate" example:0.10`
	Remark string `json:"remark" example:"老带新关系"`
}

// ReferralService 引荐关系服务
type ReferralService struct {
	db         *gorm.DB
	referralRepo *repository.ReferralRelationRepository
	userRepo   *repository.UserRepository
}

// NewReferralService 创建引荐关系服务实例
func NewReferralService(db *gorm.DB, referralRepo *repository.ReferralRelationRepository, userRepo *repository.UserRepository) *ReferralService {
	return &ReferralService{
		db:         db,
		referralRepo: referralRepo,
		userRepo:   userRepo,
	}
}

// List 获取引荐关系列表
func (s *ReferralService) List(page, pageSize int) ([]models.ReferralRelation, int64, error) {
	return s.referralRepo.ListWithFilter("", page, pageSize)
}

// Create 创建引荐关系
func (s *ReferralService) Create(req *CreateReferralRequest) error {
	// 检查引荐人和被引荐人是否存在
	if _, err := s.userRepo.FindByID(req.ReferrerID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "引荐人不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if _, err := s.userRepo.FindByID(req.ReferredID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "被引荐人不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	// 检查是否已存在有效的引荐关系
	exists, err := s.referralRepo.ExistsByReferredID(req.ReferredID)
	if err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}
	if exists {
		return &AppError{Code: apperrors.ErrDuplicateKey, Message: "该员工已被其他员工引荐"}
	}

	// 不能自己引荐自己
	if req.ReferrerID == req.ReferredID {
		return &AppError{Code: apperrors.BadRequest, Message: "不能引荐自己"}
	}

	relation := &models.ReferralRelation{
		ReferrerID: req.ReferrerID,
		ReferredID: req.ReferredID,
		Status:     1, // 生效中
	}

	if err := s.referralRepo.Create(relation); err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "创建引荐关系失败"}
	}

	return nil
}

// Terminate 终止引荐关系
func (s *ReferralService) Terminate(id int64) error {
	relation, err := s.referralRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.NotFound, Message: "引荐关系不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if relation.Status != 1 {
		return &AppError{Code: apperrors.BadRequest, Message: "该引荐关系已终止"}
	}

	now := time.Now()
	if err := s.referralRepo.UpdateFields(id, map[string]interface{}{
		"status":     2,
		"end_date":   now,
	}); err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "终止引荐关系失败"}
	}

	return nil
}
