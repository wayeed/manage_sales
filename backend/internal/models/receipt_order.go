package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ReceiptOrder 回货单模型
type ReceiptOrder struct {
	ID            int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	StoreID       int64          `gorm:"default:1" json:"store_id"`
	ReceiptNo     string         `gorm:"unique;not null" json:"receipt_no"`
	SupplierID    int64          `json:"supplier_id"`
	SupplierName  string         `json:"supplier_name"`
	Status        int8           `gorm:"default:0" json:"status"`
	TotalAmount   decimal.Decimal        `gorm:"type:decimal(12,2)" json:"total_amount"`
	TotalQuantity int            `gorm:"default:0" json:"total_quantity"`
	Remark        string         `json:"remark"`
	AuditedBy     *int64         `json:"audited_by"`
	AuditedAt     *time.Time     `json:"audited_at"`
	CreatedBy     *int64         `json:"created_by"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	Items []ReceiptItem `gorm:"foreignKey:ReceiptOrderID" json:"items"`
}

// TableName 表名
func (ReceiptOrder) TableName() string {
	return "receipt_orders"
}

// ReceiptItem 回货明细模型
type ReceiptItem struct {
	ID              int64   `gorm:"primaryKey;autoIncrement" json:"id"`
	ReceiptOrderID  int64   `gorm:"column:receipt_order_id" json:"receipt_order_id"`
	PurchaseItemID  *int64  `gorm:"column:purchase_item_id" json:"purchase_item_id"`
	SKUID           int64   `gorm:"column:sku_id" json:"sku_id"`
	ProductName     string  `gorm:"column:product_name" json:"product_name"`
	SKUName         string  `gorm:"column:sku_name" json:"sku_name"`
	SKUCode         string  `gorm:"column:sku_code" json:"sku_code"`
	BrandStyle      string  `gorm:"column:brand_style" json:"brand_style"`
	ShipQuantity    int     `gorm:"column:ship_quantity" json:"ship_quantity"`
	ReceiveQuantity int     `gorm:"column:receive_quantity;default:0" json:"receive_quantity"`
	CostPrice       decimal.Decimal `gorm:"column:cost_price;type:decimal(12,2)" json:"cost_price"`
	Remark          string  `gorm:"column:remark" json:"remark"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName 表名
func (ReceiptItem) TableName() string {
	return "receipt_items"
}

// ReceiptOrderStatus 回货单状态枚举
const (
	ReceiptStatusPending  int8 = iota // 0-待审核
	ReceiptStatusApproved              // 1-已审核
	ReceiptStatusReceived              // 2-已入库
	ReceiptStatusCancelled             // 3-已取消
)
