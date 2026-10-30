package model

import "time"

const (
	FlashSaleStatusDraft     = "draft"
	FlashSaleStatusScheduled = "scheduled"
	FlashSaleStatusStopped   = "stopped"
)

type FlashSaleCampaign struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"type:text;not null" json:"name"`
	Status    string    `gorm:"type:text;not null;default:draft" json:"status"`
	StartsAt  time.Time `gorm:"type:timestamptz;not null" json:"starts_at"`
	EndsAt    time.Time `gorm:"type:timestamptz;not null" json:"ends_at"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamptz;not null;autoUpdateTime" json:"updated_at"`
}

func (FlashSaleCampaign) TableName() string {
	return "flash_sale_campaigns"
}
