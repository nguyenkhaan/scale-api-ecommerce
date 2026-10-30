package model

import "time"

const (
	ProductStatusActive   = "active"
	ProductStatusInactive = "inactive"
)

type Product struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CategoryID  *string   `gorm:"type:uuid" json:"category_id"`
	Name        string    `gorm:"type:text;not null" json:"name"`
	Description string    `gorm:"type:text;not null;default:''" json:"description"`
	ImageURL    *string   `gorm:"type:text" json:"image_url"`
	Status      string    `gorm:"type:text;not null;default:active" json:"status"`
	CreatedAt   time.Time `gorm:"type:timestamptz;not null;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"type:timestamptz;not null;autoUpdateTime" json:"updated_at"`
}

func (Product) TableName() string {
	return "products"
}
