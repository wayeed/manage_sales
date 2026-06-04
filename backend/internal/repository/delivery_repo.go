package repository

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"furniture-commission/internal/dto"
	"furniture-commission/internal/models"

	"gorm.io/gorm"
)

// DeliveryRepository 送货出库数据访问接口
type DeliveryRepository interface {
	// Create 创建送货出库记录
	Create(record *models.DeliveryRecord) error
	// CreateItems 批量创建送货出库明细
	CreateItems(items []models.DeliveryItem) error
	// GetByID 根据ID获取送货出库记录
	GetByID(id uint64) (*models.DeliveryRecord, error)
	// GetByOrderID 根据订单ID获取送货出库记录列表
	GetByOrderID(orderID uint64) ([]models.DeliveryRecord, error)
	// GetList 获取送货出库记录列表
	GetList(query *dto.DeliveryListQuery) ([]models.DeliveryRecord, int64, error)
	// Update 更新送货出库记录
	Update(record *models.DeliveryRecord) error
	// Cancel 作废送货出库记录
	Cancel(id uint64) error
	// GetByDeliveryNo 根据送货单号获取记录
	GetByDeliveryNo(deliveryNo string) (*models.DeliveryRecord, error)
	// ExistsByOrderID 检查订单是否已有送货记录
	ExistsByOrderID(orderID uint64) (bool, error)
}

// deliveryRepository 送货出库数据访问实现
type deliveryRepository struct {
	db *gorm.DB
}

// NewDeliveryRepository 创建送货出库数据访问实例
func NewDeliveryRepository(db *gorm.DB) DeliveryRepository {
	return &deliveryRepository{db: db}
}

// Create 创建送货出库记录
func (r *deliveryRepository) Create(record *models.DeliveryRecord) error {
	return r.db.Create(record).Error
}

// CreateItems 批量创建送货出库明细
func (r *deliveryRepository) CreateItems(items []models.DeliveryItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.Create(&items).Error
}

// GetByID 根据ID获取送货出库记录（包含明细）
func (r *deliveryRepository) GetByID(id uint64) (*models.DeliveryRecord, error) {
	var record models.DeliveryRecord
	err := r.db.Preload("Items").
		Preload("Order").
		Preload("Warehouse").
		Preload("Operator").
		First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetByOrderID 根据订单ID获取送货出库记录列表
func (r *deliveryRepository) GetByOrderID(orderID uint64) ([]models.DeliveryRecord, error) {
	var records []models.DeliveryRecord
	err := r.db.Where("order_id = ?", orderID).
		Preload("Items").
		Preload("Operator").
		Order("created_at DESC").
		Find(&records).Error
	return records, err
}

// GetList 获取送货出库记录列表
func (r *deliveryRepository) GetList(query *dto.DeliveryListQuery) ([]models.DeliveryRecord, int64, error) {
	var records []models.DeliveryRecord
	var total int64

	db := r.db.Model(&models.DeliveryRecord{})

	// 构建查询条件
	if query.StoreID > 0 {
		db = db.Where("store_id = ?", query.StoreID)
	}
	if query.OrderID > 0 {
		db = db.Where("order_id = ?", query.OrderID)
	}
	if query.OrderNo != "" {
		db = db.Where("order_no LIKE ?", "%"+query.OrderNo+"%")
	}
	if query.OperatorID > 0 {
		db = db.Where("operator_id = ?", query.OperatorID)
	}
	if query.Status > 0 {
		db = db.Where("status = ?", query.Status)
	}
	if query.StartTime != nil {
		db = db.Where("delivery_time >= ?", query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("delivery_time <= ?", query.EndTime)
	}

	// 查询总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	err := db.Preload("Items").
		Preload("Order").
		Preload("Warehouse").
		Preload("Operator").
		Order("created_at DESC").
		Offset(offset).
		Limit(query.PageSize).
		Find(&records).Error

	return records, total, err
}

// Update 更新送货出库记录
func (r *deliveryRepository) Update(record *models.DeliveryRecord) error {
	return r.db.Save(record).Error
}

// Cancel 作废送货出库记录
func (r *deliveryRepository) Cancel(id uint64) error {
	return r.db.Model(&models.DeliveryRecord{}).
		Where("id = ?", id).
		Update("status", models.DeliveryStatusCancelled).Error
}

// GetByDeliveryNo 根据送货单号获取记录
func (r *deliveryRepository) GetByDeliveryNo(deliveryNo string) (*models.DeliveryRecord, error) {
	var record models.DeliveryRecord
	err := r.db.Where("delivery_no = ?", deliveryNo).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// ExistsByOrderID 检查订单是否已有送货记录
func (r *deliveryRepository) ExistsByOrderID(orderID uint64) (bool, error) {
	var count int64
	err := r.db.Model(&models.DeliveryRecord{}).
		Where("order_id = ? AND status = ?", orderID, models.DeliveryStatusNormal).
		Count(&count).Error
	return count > 0, err
}

// WithTransaction 事务包装器
func (r *deliveryRepository) WithTransaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// GetDB 获取数据库连接
func (r *deliveryRepository) GetDB() *gorm.DB {
	return r.db
}

// DeliveryRepo 送货出库仓库（用于Service层直接调用）
type DeliveryRepo struct {
	db *gorm.DB
}

// NewDeliveryRepo 创建送货出库仓库实例
func NewDeliveryRepo(db *gorm.DB) *DeliveryRepo {
	return &DeliveryRepo{db: db}
}

// CreateInTx 在事务中创建送货记录
func (r *DeliveryRepo) CreateInTx(ctx context.Context, tx *gorm.DB, record *models.DeliveryRecord) error {
	if tx == nil {
		tx = r.db
	}
	return tx.WithContext(ctx).Create(record).Error
}

// CreateItemsInTx 在事务中批量创建送货明细
func (r *DeliveryRepo) CreateItemsInTx(ctx context.Context, tx *gorm.DB, items []models.DeliveryItem) error {
	if tx == nil {
		tx = r.db
	}
	if len(items) == 0 {
		return nil
	}
	return tx.WithContext(ctx).Create(&items).Error
}

// GetByIDWithItems 根据ID获取送货记录及明细
func (r *DeliveryRepo) GetByIDWithItems(id uint64) (*models.DeliveryRecord, error) {
	var record models.DeliveryRecord
	err := r.db.Preload("Items").First(&record, id).Error
	return &record, err
}

// UpdateOrderDeliveryStatus 更新订单配送状态
func (r *DeliveryRepo) UpdateOrderDeliveryStatus(orderID uint64, status uint8) error {
	return r.db.Model(&models.Order{}).
		Where("id = ?", orderID).
		Update("delivery_status", status).Error
}

// GenerateDeliveryNo 生成送货单号
func (r *DeliveryRepo) GenerateDeliveryNo(storeID uint64) string {
	// 格式: DL + 年月日 + 门店ID(3位) + 随机数(4位)
	now := time.Now()
	randomNum := rand.Intn(9000) + 1000
	return fmt.Sprintf("DL%s%03d%d", now.Format("20060102"), storeID%1000, randomNum)
}