package payment

import (
	"context"
	"fmt"
	"time"
)

const (
	defaultSessionListLimit = 50
	maxSessionListLimit     = 200
)

// SimulatorSession is one row of the admin simulator listing.
type SimulatorSession struct {
	ID                 string    `gorm:"column:id" json:"id"`
	OrderID            string    `gorm:"column:order_id" json:"order_id"`
	OrderNumber        string    `gorm:"column:order_number" json:"order_number"`
	Amount             int64     `gorm:"column:amount" json:"amount"`
	Currency           string    `gorm:"column:currency" json:"currency"`
	Status             string    `gorm:"column:status" json:"status"`
	SimulatorReference string    `gorm:"column:simulator_reference" json:"simulator_reference"`
	OrderStatus        string    `gorm:"column:order_status" json:"order_status"`
	ExpiresAt          time.Time `gorm:"column:expires_at" json:"expires_at"`
}

var sessionStatuses = map[string]bool{"pending": true, "unknown": true, "succeeded": true, "failed": true}

// ListSimulatorSessions returns payment sessions newest first, optionally filtered by session status.
func (service *Service) ListSimulatorSessions(ctx context.Context, status string, limit int) ([]SimulatorSession, error) {
	items := []SimulatorSession{}
	err := service.db.WithContext(ctx).Raw(
		`SELECT ps.id, ps.order_id, o.order_number, ps.amount, ps.currency, ps.status, ps.simulator_reference,
			o.status AS order_status, o.expires_at
		 FROM payment_sessions ps JOIN orders o ON o.id = ps.order_id
		 WHERE ? = '' OR ps.status = ?
		 ORDER BY ps.created_at DESC, ps.id DESC LIMIT ?`,
		status, status, limit,
	).Scan(&items).Error
	if err != nil {
		return nil, fmt.Errorf("list simulator sessions: %w", err)
	}
	return items, nil
}
