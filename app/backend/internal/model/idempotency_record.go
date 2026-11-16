package model

import "time"

const IdempotencyOperationPlaceOrder = "place_order"

type IdempotencyRecord struct {
	ID             string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID         string     `gorm:"type:uuid;not null" json:"user_id"`
	Operation      string     `gorm:"type:text;not null" json:"operation"`
	IdempotencyKey string     `gorm:"type:text;not null" json:"idempotency_key"`
	RequestHash    []byte     `gorm:"type:bytea;not null" json:"request_hash"`
	ResponseStatus *int       `gorm:"type:integer" json:"response_status"`
	ResponseBody   []byte     `gorm:"type:jsonb" json:"response_body"`
	CompletedAt    *time.Time `gorm:"type:timestamptz" json:"completed_at"`
	CreatedAt      time.Time  `gorm:"type:timestamptz;not null" json:"created_at"`
}

func (IdempotencyRecord) TableName() string { return "idempotency_records" }
