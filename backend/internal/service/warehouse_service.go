package service

import (
	"furniture-commission/internal/models"
	apperrors "furniture-commission/internal/pkg/errors"
	"furniture-commission/internal/repository"
)

// WarehouseService 仓库服务
type WarehouseService struct {
	warehouseRepo *repository.WarehouseRepository
}

// NewWarehouseService 创建仓库服务实例
func NewWarehouseService(warehouseRepo *repository.WarehouseRepository) *WarehouseService {
	return &WarehouseService{warehouseRepo: warehouseRepo}
}

// List 获取仓库列表
func (s *WarehouseService) List(storeID int64) ([]models.Warehouse, error) {
	warehouses, err := s.warehouseRepo.List(storeID)
	if err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "查询仓库列表失败"}
	}
	return warehouses, nil
}

// CreateWarehouseRequest 创建仓库请求
type CreateWarehouseRequest struct {
	WarehouseCode string `json:"warehouse_code" binding:"required" example:"WH001"`
	WarehouseName string `json:"warehouse_name" binding:"required" example:"总仓"`
	WarehouseType int8 `json:"warehouse_type" binding:"required" example:1`
	StoreID int64 `json:"store_id" example:1`
	Address string `json:"address" example:"北京市大兴区某某仓库区"`
	ContactPerson string `json:"contact_name" example:"周八"`
	ContactPhone string `json:"contact_phone" example:"13400134000"`
	Status int8 `json:"status" example:1`
}

// UpdateWarehouseRequest 更新仓库请求
type UpdateWarehouseRequest struct {
	WarehouseCode string `json:"warehouse_code" example:"WH001"`
	WarehouseName string `json:"warehouse_name" example:"总仓"`
	WarehouseType int8 `json:"warehouse_type" example:1`
	StoreID int64 `json:"store_id" example:1`
	Address string `json:"address" example:"北京市大兴区某某仓库区"`
	ContactPerson string `json:"contact_name" example:"周八"`
	ContactPhone string `json:"contact_phone" example:"13400134000"`
	Status int8 `json:"status" example:1`
}

// Create 创建仓库
func (s *WarehouseService) Create(req *CreateWarehouseRequest) (*models.Warehouse, error) {
	warehouse := &models.Warehouse{
		WarehouseCode: req.WarehouseCode,
		WarehouseName: req.WarehouseName,
		WarehouseType: req.WarehouseType,
		StoreID:       req.StoreID,
		Address:       req.Address,
		ContactPerson:   req.ContactPerson,
		ContactPhone:  req.ContactPhone,
		Status:        req.Status,
	}
	if err := s.warehouseRepo.Create(warehouse); err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "创建仓库失败"}
	}
	return warehouse, nil
}

// Update 更新仓库
func (s *WarehouseService) Update(id int64, req *UpdateWarehouseRequest) (*models.Warehouse, error) {
	warehouse, err := s.warehouseRepo.FindByID(id)
	if err != nil {
		return nil, &AppError{Code: apperrors.NotFound, Message: "仓库不存在"}
	}
	if req.WarehouseCode != "" {
		warehouse.WarehouseCode = req.WarehouseCode
	}
	if req.WarehouseName != "" {
		warehouse.WarehouseName = req.WarehouseName
	}
	if req.WarehouseType > 0 {
		warehouse.WarehouseType = req.WarehouseType
	}
	if req.StoreID > 0 {
		warehouse.StoreID = req.StoreID
	}
	warehouse.Address = req.Address
	warehouse.ContactPerson = req.ContactPerson
	warehouse.ContactPhone = req.ContactPhone
	if req.Status >= 0 {
		warehouse.Status = req.Status
	}
	if err := s.warehouseRepo.Update(warehouse); err != nil {
		return nil, &AppError{Code: apperrors.InternalError, Message: "更新仓库失败"}
	}
	return warehouse, nil
}

// Delete 删除仓库
func (s *WarehouseService) Delete(id int64) error {
	if err := s.warehouseRepo.Delete(id); err != nil {
		return &AppError{Code: apperrors.InternalError, Message: "删除仓库失败"}
	}
	return nil
}

// GetByID 根据ID获取仓库
func (s *WarehouseService) GetByID(id int64) (*models.Warehouse, error) {
	warehouse, err := s.warehouseRepo.FindByID(id)
	if err != nil {
		return nil, &AppError{Code: apperrors.NotFound, Message: "仓库不存在"}
	}
	return warehouse, nil
}
