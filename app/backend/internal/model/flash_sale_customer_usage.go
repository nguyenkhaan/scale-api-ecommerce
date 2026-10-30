package model

import "time"

type FlashSaleCustomerUsage struct {
	FlashSaleItemID  string    `gorm:"type:uuid;primaryKey" json:"flash_sale_item_id"`
	UserID           string    `gorm:"type:uuid;primaryKey" json:"user_id"`
	ReservedQuantity int64     `gorm:"type:bigint;not null;default:0" json:"reserved_quantity"`
	SoldQuantity     int64     `gorm:"type:bigint;not null;default:0" json:"sold_quantity"`
	UpdatedAt        time.Time `gorm:"type:timestamptz;not null;autoUpdateTime" json:"updated_at"`
}

func (FlashSaleCustomerUsage) TableName() string {
	return "flash_sale_customer_usage"
}
