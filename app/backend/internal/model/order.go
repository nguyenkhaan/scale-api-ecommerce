package model

import "time"

const OrderStatusPendingPayment = "pending_payment"

type Order struct {
	ID               string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID           string     `gorm:"type:uuid;not null" json:"user_id"`
	OrderNumber      string     `gorm:"type:text;not null" json:"order_number"`
	Status           string     `gorm:"type:text;not null" json:"status"`
	Currency         string     `gorm:"type:text;not null" json:"currency"`
	TotalAmount      int64      `gorm:"type:bigint;not null" json:"total_amount"`
	RecipientName    string     `gorm:"type:text;not null" json:"recipient_name"`
	RecipientPhone   string     `gorm:"type:text;not null" json:"recipient_phone"`
	RecipientAddress string     `gorm:"type:text;not null" json:"recipient_address"`
	CreatedAt        time.Time  `gorm:"type:timestamptz;not null" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"type:timestamptz;not null" json:"updated_at"`
	ExpiresAt        time.Time  `gorm:"type:timestamptz;not null" json:"expires_at"`
	TerminalAt       *time.Time `gorm:"type:timestamptz" json:"terminal_at"`
}

func (Order) TableName() string { return "orders" }
