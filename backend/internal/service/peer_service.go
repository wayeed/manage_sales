package service

import (
	"errors"

	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	"furniture-commission/internal/repository"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// PeerService 同行服务
type PeerService struct {
	db       *gorm.DB
	peerRepo *repository.PeerRepository
	userRepo *repository.UserRepository
}

// NewPeerService 创建同行服务实例
func NewPeerService(db *gorm.DB, peerRepo *repository.PeerRepository, userRepo *repository.UserRepository) *PeerService {
	return &PeerService{db: db, peerRepo: peerRepo, userRepo: userRepo}
}

// CreatePeerRequest 创建同行请求
type CreatePeerRequest struct {
	PeerName string `json:"peer_name" binding:"required" example:"赵六"`
	Phone string `json:"phone" example:"13600136000"`
	IDCard string `json:"id_card" example:"110101199002021234"`
	Company string `json:"company" example:"某某家具公司"`
	BankAccount string `json:"bank_account" example:"6222021234567890456"`
	BankName string `json:"bank_name" example:"中国建设银行"`
	CommissionRate float64 `json:"commission_rate" example:0.10`
	Remark string `json:"remark" example:"长期合作伙伴"`
}

// UpdatePeerRequest 更新同行请求
type UpdatePeerRequest struct {
	PeerName string `json:"peer_name" example:"赵六"`
	Phone string `json:"phone" example:"13600136000"`
	IDCard string `json:"id_card" example:"110101199002021234"`
	Company string `json:"company" example:"某某家具公司"`
	BankAccount string `json:"bank_account" example:"6222021234567890456"`
	BankName string `json:"bank_name" example:"中国建设银行"`
	CommissionRate float64 `json:"commission_rate" example:0.10`
	Remark string `json:"remark" example:"长期合作伙伴"`
	Status *int8 `json:"status" example:1`
}

// ListPeerRequest 同行列表查询请求
type ListPeerRequest struct {
	StoreID string `form:"store_id" example:"1"`
	Keyword string `form:"keyword" example:"赵六"`
	Page int `form:"page" example:1`
	PageSize int `form:"page_size" example:10`
}

// Create 创建同行
func (s *PeerService) Create(req *CreatePeerRequest, userID int64) (*models.Peer, error) {
	// 获取当前用户的 store_id
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, &AppError{Code: apperrors.ErrUserNotFound, Message: "用户不存在"}
	}

	peer := &models.Peer{
		StoreID:     *user.StoreID,
		PeerName:    req.PeerName,
		Phone:       req.Phone,
		IDCard:      req.IDCard,
		Company:     req.Company,
		BankAccount: req.BankAccount,
		BankName:    req.BankName,
		Remark:      req.Remark,
		Status:      1,
	}
	if req.CommissionRate > 0 {
		rate := decimal.NewFromFloat(req.CommissionRate)
		peer.CommissionRate = &rate
	}

	if err := s.peerRepo.Create(peer); err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "创建同行失败"}
	}
	return peer, nil
}

// List 同行列表
func (s *PeerService) List(req *ListPeerRequest) (*PageResult, error) {
	peers, total, err := s.peerRepo.List(req.StoreID, req.Keyword, req.Page, req.PageSize)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询同行列表失败"}
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
		List:     peers,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Update 更新同行
func (s *PeerService) Update(id int64, req *UpdatePeerRequest) error {
	peer, err := s.peerRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AppError{Code: apperrors.ErrOrderNotFound, Message: "同行不存在"}
		}
		return &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}

	if req.PeerName != "" {
		peer.PeerName = req.PeerName
	}
	if req.Phone != "" {
		peer.Phone = req.Phone
	}
	if req.IDCard != "" {
		peer.IDCard = req.IDCard
	}
	if req.Company != "" {
		peer.Company = req.Company
	}
	if req.BankAccount != "" {
		peer.BankAccount = req.BankAccount
	}
	if req.BankName != "" {
		peer.BankName = req.BankName
	}
	if req.CommissionRate > 0 {
		rate := decimal.NewFromFloat(req.CommissionRate)
		peer.CommissionRate = &rate
	} else if req.CommissionRate == 0 {
		peer.CommissionRate = nil
	}
	peer.Remark = req.Remark
	if req.Status != nil {
		peer.Status = *req.Status
	}

	if err := s.peerRepo.Update(peer); err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "更新同行失败: " + err.Error()}
	}
	return nil
}

// Delete 删除同行
func (s *PeerService) Delete(id int64) error {
	if err := s.peerRepo.Delete(id); err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "删除同行失败"}
	}
	return nil
}

// GetDetail 获取同行详情
func (s *PeerService) GetDetail(id int64) (*models.Peer, error) {
	peer, err := s.peerRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &AppError{Code: apperrors.ErrOrderNotFound, Message: "同行不存在"}
		}
		return nil, &AppError{Code: apperrors.InternalError, Message: "系统错误"}
	}
	return peer, nil
}
