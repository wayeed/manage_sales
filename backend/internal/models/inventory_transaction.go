package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// InventoryTransaction 库存流水模型
type InventoryTransaction struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	StoreID	int64	`gorm:"column:store_id;default:1" json:"store_id" example:"1"`
	WarehouseID	*int64	`gorm:"column:warehouse_id" json:"warehouse_id" example:"1"`
	TransactionType	int8	`gorm:"column:transaction_type" json:"transaction_type" example:"1"`
	BizType	int8	`gorm:"column:biz_type" json:"biz_type" example:"order"`
	BizID	*int64	`gorm:"column:biz_id" json:"biz_id" example:"1"`
	BatchID	*int64	`gorm:"column:batch_id" json:"batch_id" example:"1"`
	RelatedOrderID	*int64	`gorm:"column:related_order_id" json:"related_order_id" example:"1"`
	RelatedPurchaseID	*int64	`gorm:"column:related_purchase_id" json:"related_purchase_id" example:"1"`
	Quantity	int	`gorm:"column:quantity" json:"quantity" example:"10"`
	BeforeStock	int	`gorm:"column:before_stock" json:"before_stock" example:"100"`
	AfterStock	int	`gorm:"column:after_stock" json:"after_stock" example:"90"`
	UnitCost	decimal.Decimal	`gorm:"column:unit_cost;type:decimal(12,2);default:0.00" json:"unit_cost" example:"150.00"`
	TotalCost	decimal.Decimal	`gorm:"column:total_cost;type:decimal(12,2);default:0.00" json:"total_cost" example:"3200.00"`
	Remark	string	`gorm:"column:remark;type:varchar(500)" json:"remark" example:"备注信息"`
	CreatedBy	*int64	`gorm:"column:created_by" json:"created_by" example:"1"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`

	// 关联
	SKU       *ProductSKU `gorm:"foreignKey:BizID;references:ID" json:"sku,omitempty"`
	Warehouse *Warehouse  `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	Creator   *User       `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// 库存交易类型常量
const (
	TransactionTypePurchaseIn  int8 = 1  // 采购入库
	TransactionTypeSaleOut     int8 = 2  // 销售出库
	TransactionTypeTransferOut int8 = 3  // 调拨出库
	TransactionTypeTransferIn  int8 = 4  // 调拨入库
	TransactionTypeProfit      int8 = 5  // 盘盈
	TransactionTypeLoss        int8 = 6  // 盘亏
	TransactionTypeGiftOut     int8 = 7  // 礼品出库
	TransactionTypeGiftIn      int8 = 8  // 礼品入库
	TransactionTypeLock        int8 = 9  // 库存锁定（订单创建时）
	TransactionTypeUnlock      int8 = 10 // 库存解锁（订单取消/驳回时）
	TransactionTypeLockToOut   int8 = 11 // 销售锁定转出库（送货出库时）
	TransactionTypeReturnIn    int8 = 12 // 退货入库
)

// GetTransactionTypeName 获取交易类型名称
func GetTransactionTypeName(transactionType int8) string {
	switch transactionType {
	case TransactionTypePurchaseIn:
		return "采购入库"
	case TransactionTypeSaleOut:
		return "销售出库"
	case TransactionTypeTransferOut:
		return "调拨出库"
	case TransactionTypeTransferIn:
		return "调拨入库"
	case TransactionTypeProfit:
		return "盘盈"
	case TransactionTypeLoss:
		return "盘亏"
	case TransactionTypeGiftOut:
		return "礼品出库"
	case TransactionTypeGiftIn:
		return "礼品入库"
	case TransactionTypeLock:
		return "库存锁定"
	case TransactionTypeUnlock:
		return "库存解锁"
	case TransactionTypeLockToOut:
		return "销售锁定转出库"
	case TransactionTypeReturnIn:
		return "退货入库"
	default:
		return "未知"
	}
}

// TableName 指定表名
func (InventoryTransaction) TableName() string { return "inventory_transactions" }
