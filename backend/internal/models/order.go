package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Order 订单模型
type Order struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	StoreID	int64	`gorm:"column:store_id;default:1" json:"store_id" example:"1"`
	OrderNo	string	`gorm:"column:order_no;type:varchar(32);uniqueIndex" json:"order_no" example:"ORD20250115001"`
	SalesmanID	int64	`gorm:"column:salesman_id;not null" json:"salesman_id" example:"1"`
	CustomerID	*int64	`gorm:"column:customer_id" json:"customer_id" example:"1"`
	CustomerName	string	`gorm:"column:customer_name;type:varchar(100)" json:"customer_name" example:"王五"`
	CustomerPhone	string	`gorm:"column:customer_phone;type:varchar(20)" json:"customer_phone" example:"13700137000"`
	CustomerAddress	string	`gorm:"column:customer_address;type:varchar(500)" json:"customer_address" example:"上海市浦东新区陆家嘴环路1000号"`
	Source	int8	`gorm:"column:source;default:0" json:"source" example:"1"`
	DeliveryStatus	int8	`gorm:"column:delivery_status;default:0" json:"delivery_status" example:"0"`
	OrderType	int8	`gorm:"column:order_type;default:1" json:"order_type" example:"1"`
	OrderStatus	int8	`gorm:"column:order_status;default:0" json:"order_status" example:"1"`
	PaymentStatus	int8	`gorm:"column:payment_status;default:0" json:"payment_status" example:"0"`
	TotalListPrice	decimal.Decimal	`gorm:"column:total_list_price;type:decimal(12,2);default:0.00" json:"total_list_price" example:"5998.00"`
	TotalSalePrice	decimal.Decimal	`gorm:"column:total_sale_price;type:decimal(12,2);default:0.00" json:"total_sale_price" example:"4998.00"`
	DiscountAmount	decimal.Decimal	`gorm:"column:discount_amount;type:decimal(12,2);default:0.00" json:"discount_amount" example:"1000.00"`
	FinalAmount	decimal.Decimal	`gorm:"column:final_amount;type:decimal(12,2);default:0.00" json:"final_amount" example:"4998.00"`
	TotalCost	decimal.Decimal	`gorm:"column:total_cost;type:decimal(12,2);default:0.00" json:"total_cost" example:"3200.00"`
	GiftCost	decimal.Decimal	`gorm:"column:gift_cost;type:decimal(12,2);default:0.00" json:"gift_cost" example:"50.00"`
	ActualProfit	decimal.Decimal	`gorm:"column:actual_profit;type:decimal(12,2);default:0.00" json:"actual_profit" example:"1748.00"`
	CategoryCount	int	`gorm:"column:category_count;default:0" json:"category_count" example:"2"`
	SKUCount	int	`gorm:"column:sku_count;default:0" json:"sku_count" example:"3"`
	TotalQuantity	int	`gorm:"column:total_quantity;default:0" json:"total_quantity" example:"5"`
	PaidAmount	decimal.Decimal	`gorm:"column:paid_amount;type:decimal(12,2);default:0.00" json:"paid_amount" example:"4998.00"`
	RemainingAmount	decimal.Decimal	`gorm:"column:remaining_amount;type:decimal(12,2);default:0.00" json:"remaining_amount" example:"0.00"`
	IsPeerOrder	int8	`gorm:"column:is_peer_order;default:0" json:"is_peer_order" example:"0"`
	PeerID	*int64	`gorm:"column:peer_id" json:"peer_id" example:"1"`
	IsSpecialApproved	int8	`gorm:"column:is_special_approved;default:0" json:"is_special_approved" example:"0"`
	ApprovalRemark	string	`gorm:"column:approval_remark;type:varchar(500)" json:"approval_remark" example:""`
	EditCount	int8	`gorm:"column:edit_count;default:0" json:"edit_count"` // 修改次数(0=未修改)
	ApprovedBy	*int64	`gorm:"column:approved_by" json:"approved_by" example:"1"`
	ApprovedAt	*time.Time	`gorm:"column:approved_at" json:"approved_at" example:"2025-01-15T14:00:00+08:00"`
	IsReturned	int8	`gorm:"column:is_returned;default:0" json:"is_returned" example:"0"`
	ReturnAmount	decimal.Decimal	`gorm:"column:return_amount;type:decimal(12,2);default:0.00" json:"return_amount" example:"0.00"`
	ReturnProfit	decimal.Decimal	`gorm:"column:return_profit;type:decimal(12,2);default:0.00" json:"return_profit" example:"0.00"`
	Remark	string	`gorm:"column:remark;type:varchar(500)" json:"remark" example:"备注信息"`
	OrderDate	*time.Time	`gorm:"column:order_date" json:"order_date" example:"2025-01-15T00:00:00+08:00"`
	IsDraft	int8	`gorm:"column:is_draft;default:0" json:"is_draft" example:"0"` // 0-正式订单, 1-草稿
	StockStatus	int8	`gorm:"column:stock_status;default:0" json:"stock_status" example:"0"` // 0-全部有库存, 1-部分缺货, 2-全部缺货
	CreatedBy	*int64	`gorm:"column:created_by" json:"created_by" example:"1"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`

	// 非数据库字段（用于返回）
	SalesmanName string `gorm:"-" json:"salesman_name,omitempty"`

	// 关联
	Items     []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	Gifts     []OrderGift `gorm:"foreignKey:OrderID" json:"gifts,omitempty"`
	Payments  []Payment   `gorm:"foreignKey:OrderID" json:"payments,omitempty"`
	Salesman  *User       `gorm:"foreignKey:SalesmanID" json:"salesman,omitempty"`
	Customer  *Customer   `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Peer      *Peer       `gorm:"foreignKey:PeerID" json:"peer,omitempty"`
	OutboundRequest *OutboundRequest `gorm:"foreignKey:OrderID" json:"outbound_request,omitempty"`
	OutboundConfirmed bool         `gorm:"column:outbound_confirmed;default:false" json:"outbound_confirmed"`
}

// TableName 指定表名
func (Order) TableName() string { return "orders" }
