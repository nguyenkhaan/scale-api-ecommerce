package model

import "time"

type FlashSaleItem struct {
	ID            string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CampaignID    string    `gorm:"type:uuid;not null" json:"campaign_id"`
	SKUID         string    `gorm:"column:sku_id;type:uuid;not null" json:"sku_id"`
	SalePrice     int64     `gorm:"type:bigint;not null" json:"sale_price"`
	QuotaTotal    int64     `gorm:"type:bigint;not null" json:"quota_total"`
	QuotaReserved int64     `gorm:"type:bigint;not null;default:0" json:"quota_reserved"`
	QuotaSold     int64     `gorm:"type:bigint;not null;default:0" json:"quota_sold"`
	PurchaseLimit int       `gorm:"type:integer;not null" json:"purchase_limit"`
	CreatedAt     time.Time `gorm:"type:timestamptz;not null;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:timestamptz;not null;autoUpdateTime" json:"updated_at"`
}

func (FlashSaleItem) TableName() string {
	return "flash_sale_items"
}
