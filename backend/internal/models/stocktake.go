package models

import (
	"time"

	"gorm.io/gorm"
)

// Stocktake 盘点单
type Stocktake struct {
	ID            int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	StoreID       int64          `gorm:"column:store_id;default:1" json:"store_id"`
	WarehouseID   int64          `gorm:"column:warehouse_id;not null" json:"warehouse_id"`
	StocktakeNo   string         `gorm:"column:stocktake_no;type:varchar(32);uniqueIndex" json:"stocktake_no"`
	Status        int8           `gorm:"column:status;default:0" json:"status"` // 0-盘点中 1-已提交 2-已审核
	TotalItems    int            `gorm:"column:total_items;default:0" json:"total_items"`
	ProfitItems   int            `gorm:"column:profit_items;default:0" json:"profit_items"`
	LossItems     int            `gorm:"column:loss_items;default:0" json:"loss_items"`
	Remark        string         `gorm:"column:remark;type:varchar(500)" json:"remark"`
	CreatedBy     *int64         `gorm:"column:created_by" json:"created_by"`
	ApprovedBy    *int64         `gorm:"column:approved_by" json:"approved_by"`
	ApprovedAt    *time.Time     `gorm:"column:approved_at" json:"approved_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Warehouse *Warehouse `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	Items     []StocktakeItem `gorm:"foreignKey:StocktakeID" json:"items,omitempty"`
	Creator   *User       `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// StocktakeItem 盘点明细
type StocktakeItem struct {
	ID            int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	StocktakeID   int64          `gorm:"column:stocktake_id;not null;index" json:"stocktake_id"`
	SKUID         int64          `gorm:"column:sku_id;not null" json:"sku_id"`
	SystemStock   int            `gorm:"column:system_stock;default:0" json:"system_stock"`     // 系统库存
	ActualStock   int            `gorm:"column:actual_stock;default:0" json:"actual_stock"`     // 实际盘点数量
	DiffQuantity  int            `gorm:"column:diff_quantity;default:0" json:"diff_quantity"`  // 差异 = 实际 - 系统
	Remark        string         `gorm:"column:remark;type:varchar(255)" json:"remark"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Stocktake *Stocktake  `gorm:"foreignKey:StocktakeID" json:"stocktake,omitempty"`
	SKU       *ProductSKU `gorm:"foreignKey:SKUID" json:"sku,omitempty"`
}

func (Stocktake) TableName() string    { return "stocktakes" }
func (StocktakeItem) TableName() string { return "stocktake_items" }
