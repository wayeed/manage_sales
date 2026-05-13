package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Supplier 供应商模型
type Supplier struct {
	ID	int64	`gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	StoreID	int64	`gorm:"column:store_id;default:1" json:"store_id" example:"1"`
	SupplierCode	string	`gorm:"column:supplier_code;type:varchar(32);uniqueIndex:uk_store_supplier_code" json:"supplier_code" example:"SUP001"`
	SupplierName	string	`gorm:"column:supplier_name;type:varchar(100)" json:"supplier_name" example:"深圳某某科技有限公司"`
	ContactPerson	string	`gorm:"column:contact_person;type:varchar(50)" json:"contact_person" example:"李四"`
	ContactPhone	string	`gorm:"column:contact_phone;type:varchar(20)" json:"contact_phone" example:"13900139000"`
	Address	string	`gorm:"column:address;type:varchar(255)" json:"address" example:"北京市朝阳区建国路88号"`
	BusinessScope	string	`gorm:"column:business_scope;type:varchar(255)" json:"business_scope" example:"电子产品、通讯设备"`
	BankName	string	`gorm:"column:bank_name;type:varchar(100)" json:"bank_name" example:"中国工商银行"`
	BankAccount	string	`gorm:"column:bank_account;type:varchar(50)" json:"bank_account" example:"6222021234567890123"`
	TaxNo	string	`gorm:"column:tax_no;type:varchar(50)" json:"tax_no" example:"91440300MA5XXXXX"`
	TotalPurchaseAmount	decimal.Decimal	`gorm:"column:total_purchase_amount;type:decimal(14,2);default:0.00" json:"total_purchase_amount" example:"500000.00"`
	TotalPurchaseOrders	int	`gorm:"column:total_purchase_orders;default:0" json:"total_purchase_orders" example:"20"`
	LastPurchaseAt	*time.Time	`gorm:"column:last_purchase_at" json:"last_purchase_at" example:"2025-01-15T00:00:00+08:00"`
	Remark	string	`gorm:"column:remark;type:varchar(500)" json:"remark" example:"备注信息"`
	Status	int8	`gorm:"column:status;default:1" json:"status" example:"1"`
	CreatedAt	time.Time	`json:"created_at" example:"2025-01-15T00:00:00+08:00"`
	UpdatedAt	time.Time	`json:"updated_at" example:"2025-01-15T00:00:00+08:00"`
	DeletedAt          gorm.DeletedAt  `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Supplier) TableName() string { return "suppliers" }
