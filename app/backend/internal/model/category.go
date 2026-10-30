package model

import "time"

type Category struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"type:text;not null" json:"name"`
	Slug      string    `gorm:"type:text;not null;uniqueIndex" json:"slug"`
	IsVisible *bool     `gorm:"type:boolean;not null;default:true" json:"is_visible"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamptz;not null;autoUpdateTime" json:"updated_at"`
}

func (Category) TableName() string {
	return "categories"
}
