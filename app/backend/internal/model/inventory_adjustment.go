package model

import "time"

const InventoryAdjustmentReasonInitialStock = "INITIAL_STOCK"

type InventoryAdjustment struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SKUID        string    `gorm:"column:sku_id;type:uuid;not null" json:"sku_id"`
	ActorUserID  string    `gorm:"type:uuid;not null" json:"actor_user_id"`
	Delta        int64     `gorm:"type:bigint;not null" json:"delta"`
	OnHandBefore int64     `gorm:"type:bigint;not null" json:"on_hand_before"`
	OnHandAfter  int64     `gorm:"type:bigint;not null" json:"on_hand_after"`
	Reason       string    `gorm:"type:text;not null" json:"reason"`
	CreatedAt    time.Time `gorm:"type:timestamptz;not null;autoCreateTime" json:"created_at"`
}

func (InventoryAdjustment) TableName() string {
	return "inventory_adjustments"
}
