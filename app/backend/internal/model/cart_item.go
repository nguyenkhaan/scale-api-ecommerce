package model

import "time"

const (
	PriceSourceRegular   = "regular"
	PriceSourceFlashSale = "flash_sale"
)

type CartItem struct {
	ID                  string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CartID              string    `gorm:"type:uuid;not null" json:"cart_id"`
	SKUID               string    `gorm:"column:sku_id;type:uuid;not null" json:"sku_id"`
	Quantity            int       `gorm:"type:integer;not null" json:"quantity"`
	IntendedPriceSource string    `gorm:"type:text;not null" json:"intended_price_source"`
	FlashSaleItemID     *string   `gorm:"type:uuid" json:"flash_sale_item_id"`
	Version             int64     `gorm:"type:bigint;not null;default:1" json:"version"`
	CreatedAt           time.Time `gorm:"type:timestamptz;not null;autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time `gorm:"type:timestamptz;not null;autoUpdateTime" json:"updated_at"`
}

func (CartItem) TableName() string {
	return "cart_items"
}
