package model

import "time"

const InventoryReservationActive = "active"

type InventoryReservation struct {
	ID          string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrderItemID string     `gorm:"type:uuid;not null" json:"order_item_id"`
	SKUID       string     `gorm:"type:uuid;not null" json:"sku_id"`
	Quantity    int        `gorm:"type:integer;not null" json:"quantity"`
	Status      string     `gorm:"type:text;not null" json:"status"`
	CreatedAt   time.Time  `gorm:"type:timestamptz;not null" json:"created_at"`
	ReleasedAt  *time.Time `gorm:"type:timestamptz" json:"released_at"`
}

func (InventoryReservation) TableName() string { return "inventory_reservations" }
