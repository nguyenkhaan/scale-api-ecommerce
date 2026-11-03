package model

import "time"

const (
	CartStatusActive   = "active"
	CartStatusArchived = "archived"
)

type Cart struct {
	ID        string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    string     `gorm:"type:uuid;not null" json:"user_id"`
	Status    string     `gorm:"type:text;not null;default:active" json:"status"`
	CreatedAt time.Time  `gorm:"type:timestamptz;not null;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;not null;autoUpdateTime" json:"updated_at"`
	Items     []CartItem `gorm:"foreignKey:CartID" json:"items,omitempty"`
}

func (Cart) TableName() string {
	return "carts"
}
