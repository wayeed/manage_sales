package repository

import (
	"furniture-commission/internal/models"

	"gorm.io/gorm"
)

// PeerRepository 同行Repository
type PeerRepository struct {
	db *gorm.DB
}

// NewPeerRepository 创建同行Repository实例
func NewPeerRepository(db *gorm.DB) *PeerRepository {
	return &PeerRepository{db: db}
}

// List 带条件分页查询同行列表
func (r *PeerRepository) List(storeID, keyword string, page, pageSize int) ([]models.Peer, int64, error) {
	db := r.db.Model(&models.Peer{})

	if storeID != "" {
		db = db.Where("store_id = ?", storeID)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("peer_name LIKE ? OR phone LIKE ? OR company LIKE ?", like, like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var peers []models.Peer
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").
		Find(&peers).Error
	if err != nil {
		return nil, 0, err
	}

	return peers, total, nil
}

// FindByID 根据ID查找同行
func (r *PeerRepository) FindByID(id int64) (*models.Peer, error) {
	var peer models.Peer
	err := r.db.First(&peer, id).Error
	if err != nil {
		return nil, err
	}
	return &peer, nil
}

// FindByPhone 根据手机号查找同行
func (r *PeerRepository) FindByPhone(storeID int64, phone string) (*models.Peer, error) {
	var peer models.Peer
	err := r.db.Where("store_id = ? AND phone = ?", storeID, phone).First(&peer).Error
	if err != nil {
		return nil, err
	}
	return &peer, nil
}

// Create 创建同行
func (r *PeerRepository) Create(peer *models.Peer) error {
	return r.db.Create(peer).Error
}

// Update 更新同行
func (r *PeerRepository) Update(peer *models.Peer) error {
	return r.db.Save(peer).Error
}

// Delete 删除同行
func (r *PeerRepository) Delete(id int64) error {
	return r.db.Delete(&models.Peer{}, id).Error
}

// UpdateStats 更新同行累计统计
func (r *PeerRepository) UpdateStats(id int64, totalOrders int, totalAmount, totalProfit float64) error {
	return r.db.Model(&models.Peer{}).Where("id = ?", id).Updates(map[string]interface{}{
		"total_orders":  gorm.Expr("total_orders + ?", totalOrders),
		"total_amount":  gorm.Expr("total_amount + ?", totalAmount),
		"total_profit":  gorm.Expr("total_profit + ?", totalProfit),
	}).Error
}
