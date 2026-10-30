package model

import "time"

const (
	RoleCustomer = "customer"
	RoleAdmin    = "admin"

	UserStatusActive   = "active"
	UserStatusDisabled = "disabled"
)

type User struct {
	ID              string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email           string
	EmailNormalized string
	PasswordHash    string
	Role            string
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
