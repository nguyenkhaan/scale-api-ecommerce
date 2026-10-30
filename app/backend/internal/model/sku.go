package model

import (
	"encoding/json"
	"time"
)

const (
	SKUStatusActive   = "active"
	SKUStatusInactive = "inactive"
)

type SKU struct {
	ID                  string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProductID           string          `gorm:"type:uuid;not null" json:"product_id"`
	Code                string          `gorm:"type:text;not null;uniqueIndex" json:"code"`
	Attributes          json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"attributes"`
	RegularPrice        int64           `gorm:"type:bigint;not null" json:"regular_price"`
	MaxQuantityPerOrder int             `gorm:"type:integer;not null" json:"max_quantity_per_order"`
	Status              string          `gorm:"type:text;not null;default:active" json:"status"`
	CreatedAt           time.Time       `gorm:"type:timestamptz;not null;autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time       `gorm:"type:timestamptz;not null;autoUpdateTime" json:"updated_at"`
}

func (SKU) TableName() string {
	return "skus"
}
