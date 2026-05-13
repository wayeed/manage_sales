package repository

import (
	"furniture-commission/internal/models"

	"gorm.io/gorm"
)

// ReferralRelationRepository 老带新引荐关系Repository
type ReferralRelationRepository struct {
	db *gorm.DB
}

// NewReferralRelationRepository 创建引荐关系Repository实例
func NewReferralRelationRepository(db *gorm.DB) *ReferralRelationRepository {
	return &ReferralRelationRepository{db: db}
}

// Create 创建引荐关系
func (r *ReferralRelationRepository) Create(relation *models.ReferralRelation) error {
	return r.db.Create(relation).Error
}

// FindByID 根据ID查找引荐关系
func (r *ReferralRelationRepository) FindByID(id int64) (*models.ReferralRelation, error) {
	var relation models.ReferralRelation
	err := r.db.Preload("Referrer").Preload("Referred").First(&relation, id).Error
	if err != nil {
		return nil, err
	}
	return &relation, nil
}

// FindActiveByReferredID 查找被引荐人的有效引荐关系
func (r *ReferralRelationRepository) FindActiveByReferredID(referredID int64) (*models.ReferralRelation, error) {
	var relation models.ReferralRelation
	err := r.db.Where("referred_id = ? AND status = 1", referredID).First(&relation).Error
	if err != nil {
		return nil, err
	}
	return &relation, nil
}

// ListWithFilter 带条件分页查询引荐关系列表
func (r *ReferralRelationRepository) ListWithFilter(storeID string, page, pageSize int) ([]models.ReferralRelation, int64, error) {
	db := r.db.Model(&models.ReferralRelation{})

	if storeID != "" {
		db = db.Joins("JOIN users ON users.id = referral_relations.referrer_id").
			Where("users.store_id = ?", storeID)
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

	var relations []models.ReferralRelation
	err := db.Preload("Referrer").Preload("Referred").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").
		Find(&relations).Error
	if err != nil {
		return nil, 0, err
	}

	return relations, total, nil
}

// UpdateFields 更新引荐关系指定字段
func (r *ReferralRelationRepository) UpdateFields(id int64, fields map[string]interface{}) error {
	return r.db.Model(&models.ReferralRelation{}).Where("id = ?", id).Updates(fields).Error
}

// ExistsByReferredID 检查被引荐人是否已有引荐关系
func (r *ReferralRelationRepository) ExistsByReferredID(referredID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.ReferralRelation{}).Where("referred_id = ? AND status = 1", referredID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
