package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// DeliveryRecord 送货出库记录表
type DeliveryRecord struct {
	ID              uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	StoreID         uint64          `gorm:"column:store_id;not null;index" json:"store_id"`
	OrderID         uint64          `gorm:"column:order_id;not null;index" json:"order_id"`
	OrderNo         string          `gorm:"column:order_no;size:50;not null" json:"order_no"`
	DeliveryNo      string          `gorm:"column:delivery_no;size:50;not null;uniqueIndex:uk_delivery_no" json:"delivery_no"`
	WarehouseID     uint64          `gorm:"column:warehouse_id;not null;index" json:"warehouse_id"`
	OperatorID      uint64          `gorm:"column:operator_id;not null;index" json:"operator_id"`
	OperatorName    string          `gorm:"column:operator_name;size:50" json:"operator_name"`
	DeliveryTime    time.Time       `gorm:"column:delivery_time;not null;index" json:"delivery_time"`
	DeliveryType    uint8           `gorm:"column:delivery_type;default:1" json:"delivery_type"` // 1-自送,2-物流,3-快递
	LogisticsNo     string          `gorm:"column:logistics_no;size:100" json:"logistics_no"`
	ReceiverName    string          `gorm:"column:receiver_name;size:50" json:"receiver_name"`
	ReceiverPhone   string          `gorm:"column:receiver_phone;size:20" json:"receiver_phone"`
	ReceiverAddress string          `gorm:"column:receiver_address;size:255" json:"receiver_address"`
	Remark          string          `gorm:"column:remark;size:500" json:"remark"`
	TotalQuantity   int             `gorm:"column:total_quantity;default:0" json:"total_quantity"`
	TotalAmount     decimal.Decimal `gorm:"column:total_amount;type:decimal(12,2);default:0.00" json:"total_amount"`
	Status          uint8           `gorm:"column:status;default:1" json:"status"` // 0-作废,1-正常
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`

	// 关联关系
	Items     []DeliveryItem `gorm:"foreignKey:DeliveryID" json:"items,omitempty"`
	Order     Order          `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Warehouse Warehouse      `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	Operator  User           `gorm:"foreignKey:OperatorID" json:"operator,omitempty"`
}

// TableName 指定表名
func (DeliveryRecord) TableName() string {
	return "delivery_records"
}

// DeliveryItem 送货出库明细表
type DeliveryItem struct {
	ID          uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	DeliveryID  uint64          `gorm:"column:delivery_id;not null;index" json:"delivery_id"`
	OrderItemID uint64          `gorm:"column:order_item_id;not null" json:"order_item_id"`
	SkuID       uint64          `gorm:"column:sku_id;not null;index" json:"sku_id"`
	ProductName string          `gorm:"column:product_name;size:200" json:"product_name"`
	SkuName     string          `gorm:"column:sku_name;size:200" json:"sku_name"`
	SkuCode     string          `gorm:"column:sku_code;size:50" json:"sku_code"`
	Quantity    int             `gorm:"column:quantity;not null" json:"quantity"`
	BatchID     *uint64         `gorm:"column:batch_id" json:"batch_id"`
	UnitCost    decimal.Decimal `gorm:"column:unit_cost;type:decimal(12,2);default:0.00" json:"unit_cost"`
	TotalCost   decimal.Decimal `gorm:"column:total_cost;type:decimal(12,2);default:0.00" json:"total_cost"`
	CreatedAt   time.Time       `json:"created_at"`

	// 关联关系
	Delivery *DeliveryRecord `gorm:"foreignKey:DeliveryID" json:"delivery,omitempty"`
}

// TableName 指定表名
func (DeliveryItem) TableName() string {
	return "delivery_items"
}

// 送货类型常量
const (
	DeliveryTypeSelf     uint8 = 1 // 自送
	DeliveryTypeLogistics uint8 = 2 // 物流
	DeliveryTypeExpress  uint8 = 3 // 快递
)

// GetDeliveryTypeName 获取送货类型名称
func GetDeliveryTypeName(deliveryType uint8) string {
	switch deliveryType {
	case DeliveryTypeSelf:
		return "自送"
	case DeliveryTypeLogistics:
		return "物流"
	case DeliveryTypeExpress:
		return "快递"
	default:
		return "未知"
	}
}

// 送货记录状态常量
const (
	DeliveryStatusCancelled uint8 = 0 // 作废
	DeliveryStatusNormal    uint8 = 1 // 正常
)

// GetDeliveryStatusName 获取送货记录状态名称
func GetDeliveryStatusName(status uint8) string {
	switch status {
	case DeliveryStatusCancelled:
		return "作废"
	case DeliveryStatusNormal:
		return "正常"
	default:
		return "未知"
	}
}