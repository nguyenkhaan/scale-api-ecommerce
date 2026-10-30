package model

import "time"

type Inventory struct {
	SKUID     string    `gorm:"column:sku_id;type:uuid;primaryKey" json:"sku_id"`
	OnHand    int64     `gorm:"type:bigint;not null;default:0" json:"on_hand"`
	Reserved  int64     `gorm:"type:bigint;not null;default:0" json:"reserved"`
	Version   int64     `gorm:"type:bigint;not null;default:0" json:"version"`
	UpdatedAt time.Time `gorm:"type:timestamptz;not null;autoUpdateTime" json:"updated_at"`
}

func (Inventory) TableName() string {
	return "inventories"
}
