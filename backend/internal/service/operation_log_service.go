package service

import (
	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	"furniture-commission/internal/repository"
)

// OperationLogService 操作日志服务
type OperationLogService struct {
	logRepo *repository.OperationLogRepository
}

// NewOperationLogService 创建操作日志服务实例
func NewOperationLogService(logRepo *repository.OperationLogRepository) *OperationLogService {
	return &OperationLogService{logRepo: logRepo}
}

// ListOperationLogRequest 操作日志查询请求
type ListOperationLogRequest struct {
	Keyword string `form:"keyword" example:"创建订单"`
	Page int `form:"page" example:1`
	PageSize int `form:"pageSize" example:20`
}

// List 操作日志列表
func (s *OperationLogService) List(req *ListOperationLogRequest) (*PageResult, error) {
	logs, total, err := s.logRepo.List(req.Keyword, req.Page, req.PageSize)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询操作日志失败"}
	}

	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	return &PageResult{
		List:     logs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Create 创建操作日志
func (s *OperationLogService) Create(log *models.OperationLog) error {
	return s.logRepo.Create(log)
}
