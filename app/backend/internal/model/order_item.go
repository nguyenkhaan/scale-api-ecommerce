package model

import "encoding/json"

type OrderItem struct {
	ID              string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrderID         string          `gorm:"type:uuid;not null" json:"order_id"`
	SKUID           string          `gorm:"type:uuid;not null" json:"sku_id"`
	SKUCode         string          `gorm:"type:text;not null" json:"sku_code"`
	ProductName     string          `gorm:"type:text;not null" json:"product_name"`
	SKUAttributes   json.RawMessage `gorm:"type:jsonb;not null" json:"sku_attributes"`
	Quantity        int             `gorm:"type:integer;not null" json:"quantity"`
	UnitPrice       int64           `gorm:"type:bigint;not null" json:"unit_price"`
	Subtotal        int64           `gorm:"type:bigint;not null" json:"subtotal"`
	PriceSource     string          `gorm:"type:text;not null" json:"price_source"`
	FlashSaleItemID *string         `gorm:"type:uuid" json:"flash_sale_item_id"`
}

func (OrderItem) TableName() string { return "order_items" }
