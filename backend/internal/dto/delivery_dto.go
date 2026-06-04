package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

// CreateDeliveryRequest 创建送货出库请求
type CreateDeliveryRequest struct {
	OrderID         uint64                `json:"order_id" binding:"required" example:"1"`           // 订单ID
	WarehouseID     uint64                `json:"warehouse_id" example:"1"`                         // 出库仓库ID（打印模式可不传）
	DeliveryTime    time.Time             `json:"delivery_time" example:"2026-01-15T10:00:00+08:00"` // 送货时间
	DeliveryType    uint8                 `json:"delivery_type" example:"1"`                         // 送货类型:1-自送,2-物流,3-快递
	LogisticsNo     string                `json:"logistics_no" example:"SF1234567890"`               // 物流单号
	ReceiverName    string                `json:"receiver_name" example:"张三"`                       // 收货人姓名
	ReceiverPhone   string                `json:"receiver_phone" example:"13800138000"`              // 收货人电话
	ReceiverAddress string                `json:"receiver_address" example:"北京市朝阳区xxx街道"`        // 收货地址
	Remark          string                `json:"remark" example:"请小心搬运"`                         // 备注
	Items           []DeliveryItemRequest `json:"items"`                                            // 送货商品明细（打印模式可不传）
	PrintMode       bool                  `json:"print_mode" example:"false"`                        // 打印模式：true=只创建记录不扣减库存
}

// DeliveryItemRequest 送货商品明细请求
type DeliveryItemRequest struct {
	OrderItemID uint64 `json:"order_item_id" binding:"required" example:"1"` // 订单明细ID
	Quantity    int    `json:"quantity" binding:"required,min=1" example:"2"` // 数量
}

// DeliveryListRequest 送货出库列表请求
type DeliveryListRequest struct {
	StoreID    uint64    `form:"store_id" example:"1"`                              // 门店ID
	OrderID    uint64    `form:"order_id" example:"1"`                              // 订单ID
	OrderNo    string    `form:"order_no" example:"ORD202601150001"`                // 订单编号
	OperatorID uint64    `form:"operator_id" example:"1"`                           // 操作人ID
	Status     uint8     `form:"status" example:"1"`                                // 状态:0-作废,1-正常
	StartTime  time.Time `form:"start_time" example:"2026-01-01T00:00:00+08:00"`    // 开始时间
	EndTime    time.Time `form:"end_time" example:"2026-01-31T23:59:59+08:00"`      // 结束时间
	Page       int       `form:"page,default=1" example:"1"`                        // 页码
	PageSize   int       `form:"page_size,default=10" example:"10"`                 // 每页数量
}

// CancelDeliveryRequest 作废送货出库请求
type CancelDeliveryRequest struct {
	Remark string `json:"remark" example:"客户取消订单"` // 作废原因
}

// DeliveryListQuery 送货出库列表查询参数（内部使用）
type DeliveryListQuery struct {
	StoreID    uint64
	OrderID    uint64
	OrderNo    string
	OperatorID uint64
	Status     uint8
	StartTime  *time.Time
	EndTime    *time.Time
	Page       int
	PageSize   int
}

// PendingDeliveryOrderQuery 待送货订单查询参数
type PendingDeliveryOrderQuery struct {
	StoreID      uint64
	OrderNo      string
	CustomerName string
	SalesmanID   uint64
	Page         int
	PageSize     int
}

// DeliveryDTO 送货出库数据传输对象
type DeliveryDTO struct {
	ID               uint64          `json:"id"`
	StoreID          uint64          `json:"store_id"`
	OrderID          uint64          `json:"order_id"`
	OrderNo          string          `json:"order_no"`
	DeliveryNo       string          `json:"delivery_no"`
	WarehouseID      uint64          `json:"warehouse_id"`
	WarehouseName    string          `json:"warehouse_name"`
	OperatorID       uint64          `json:"operator_id"`
	OperatorName     string          `json:"operator_name"`
	DeliveryTime     time.Time       `json:"delivery_time"`
	DeliveryType     uint8           `json:"delivery_type"`
	DeliveryTypeName string          `json:"delivery_type_name"`
	LogisticsNo      string          `json:"logistics_no"`
	ReceiverName     string          `json:"receiver_name"`
	ReceiverPhone    string          `json:"receiver_phone"`
	ReceiverAddress  string          `json:"receiver_address"`
	Remark           string          `json:"remark"`
	TotalQuantity    int             `json:"total_quantity"`
	TotalAmount      decimal.Decimal `json:"total_amount"`
	Status           uint8           `json:"status"`
	StatusName       string          `json:"status_name"`
	Items            []DeliveryItemDTO `json:"items"`
	CreatedAt        time.Time       `json:"created_at"`
}

// DeliveryItemDTO 送货出库明细数据传输对象
type DeliveryItemDTO struct {
	ID          uint64          `json:"id"`
	DeliveryID  uint64          `json:"delivery_id"`
	OrderItemID uint64          `json:"order_item_id"`
	SkuID       uint64          `json:"sku_id"`
	ProductName string          `json:"product_name"`
	SkuName     string          `json:"sku_name"`
	SkuCode     string          `json:"sku_code"`
	Quantity    int             `json:"quantity"`
	BatchID     *uint64         `json:"batch_id"`
	UnitCost    decimal.Decimal `json:"unit_cost"`
	TotalCost   decimal.Decimal `json:"total_cost"`
}

// PendingDeliveryOrderDTO 待送货订单数据传输对象
type PendingDeliveryOrderDTO struct {
	OrderID         uint64          `json:"order_id"`
	OrderNo         string          `json:"order_no"`
	CustomerName    string          `json:"customer_name"`
	CustomerPhone   string          `json:"customer_phone"`
	CustomerAddress string          `json:"customer_address"`
	SalesmanID      uint64          `json:"salesman_id"`
	SalesmanName    string          `json:"salesman_name"`
	TotalAmount     decimal.Decimal `json:"total_amount"`
	TotalQuantity   int             `json:"total_quantity"`
	OrderTime       time.Time       `json:"order_time"`
	Items           []PendingDeliveryOrderItemDTO `json:"items"`
}

// PendingDeliveryOrderItemDTO 待送货订单商品明细
type PendingDeliveryOrderItemDTO struct {
	OrderItemID uint64          `json:"order_item_id"`
	SkuID       uint64          `json:"sku_id"`
	ProductName string          `json:"product_name"`
	SkuName     string          `json:"sku_name"`
	SkuCode     string          `json:"sku_code"`
	Quantity    int             `json:"quantity"`
	SalePrice   decimal.Decimal `json:"sale_price"`
}
