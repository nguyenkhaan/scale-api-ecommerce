package flashsale

import (
	"errors"
	"time"

	"cloudian/cloudian-scale/internal/model"
	"gorm.io/gorm"
)

// Service owns flash sale campaign configuration and lifecycle. Every time-based decision uses the
// database clock; PostgreSQL is the only authority for state transitions.
type Service struct {
	database *gorm.DB
}

func NewService(database *gorm.DB) (*Service, error) {
	if database == nil {
		return nil, errors.New("flash sale database is required")
	}
	return &Service{database: database}, nil
}

// campaignRow is a campaign plus the database clock reading of the statement that loaded it.
type campaignRow struct {
	model.FlashSaleCampaign
	DBNow time.Time `gorm:"column:db_now"`
}

func (r campaignRow) effectiveStatus() string {
	return DeriveEffectiveStatus(r.Status, r.StartsAt, r.EndsAt, r.DBNow)
}

const campaignWithNow = "flash_sale_campaigns.*, now() AS db_now"

// lockCampaign loads and row-locks a campaign inside tx.
func lockCampaign(tx *gorm.DB, id string) (campaignRow, error) {
	var rows []campaignRow
	err := tx.Raw(`SELECT `+campaignWithNow+` FROM flash_sale_campaigns WHERE id = ? FOR UPDATE`, id).Scan(&rows).Error
	if err != nil {
		return campaignRow{}, err
	}
	if len(rows) == 0 {
		return campaignRow{}, ErrNotFound
	}
	return rows[0], nil
}

// loadCampaign loads one campaign without locking.
func loadCampaign(db *gorm.DB, id string) (campaignRow, error) {
	var rows []campaignRow
	err := db.Raw(`SELECT `+campaignWithNow+` FROM flash_sale_campaigns WHERE id = ?`, id).Scan(&rows).Error
	if err != nil {
		return campaignRow{}, err
	}
	if len(rows) == 0 {
		return campaignRow{}, ErrNotFound
	}
	return rows[0], nil
}
