package payment

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"cloudian/cloudian-scale/internal/module/order"
	"gorm.io/gorm"
)

// Session is the PaymentSession contract object.
type Session struct {
	ID                 string `gorm:"column:id" json:"id"`
	OrderID            string `gorm:"column:order_id" json:"order_id"`
	Status             string `gorm:"column:status" json:"status"`
	Amount             int64  `gorm:"column:amount" json:"amount"`
	Currency           string `gorm:"column:currency" json:"currency"`
	SimulatorReference string `gorm:"column:simulator_reference" json:"simulator_reference"`
}

const sessionColumns = `id, order_id, status, amount, currency, simulator_reference`

func newSimulatorReference() string {
	buffer := make([]byte, 8)
	// crypto/rand.Read never returns an error since Go 1.24.
	rand.Read(buffer)
	return "SIM-" + strings.ToUpper(hex.EncodeToString(buffer))
}

// CreateOrGetSession returns the order's payment session, creating it on first use. Only the owner of a
// still-payable order gets one; it never changes the order's deadline (BR-26). An overdue pending order is
// expired in the same transaction so the caller sees the truth before the 409.
func (service *Service) CreateOrGetSession(ctx context.Context, userID, orderID string) (Session, bool, error) {
	var session Session
	created, payable := false, true
	err := service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// The row lock makes payability a decision about the order as it is now: a cancel or expiry holding the
		// lock finishes first, and FOR UPDATE (not SHARE) lets the overdue branch call order.Finish safely.
		var current struct {
			Status  string `gorm:"column:status"`
			Overdue bool   `gorm:"column:overdue"`
		}
		result := tx.Raw(`SELECT status, expires_at <= now() AS overdue FROM orders WHERE id = ? AND user_id = ? FOR UPDATE`,
			orderID, userID).Scan(&current)
		if result.Error != nil {
			return fmt.Errorf("read order: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		if current.Status != order.StatusPendingPayment {
			payable = false
			return nil
		}
		if current.Overdue {
			payable = false
			if _, err := order.Finish(tx, orderID, order.OutcomeExpired); err != nil {
				return fmt.Errorf("expire overdue order: %w", err)
			}
			return nil
		}
		// Insert only while the order is still pending and before its deadline; a concurrent first call
		// loses on the unique order_id and re-reads the winner's row.
		var inserted []string
		if err := tx.Raw(
			`INSERT INTO payment_sessions (order_id, amount, simulator_reference)
			 SELECT o.id, o.total_amount, ? FROM orders o
			 WHERE o.id = ? AND o.status = 'pending_payment' AND o.expires_at > now()
			 ON CONFLICT (order_id) DO NOTHING RETURNING id`,
			newSimulatorReference(), orderID,
		).Scan(&inserted).Error; err != nil {
			return fmt.Errorf("insert payment session: %w", err)
		}
		created = len(inserted) == 1
		var fresh Session
		read := tx.Raw(`SELECT `+sessionColumns+` FROM payment_sessions WHERE order_id = ?`, orderID).Scan(&fresh)
		if read.Error != nil {
			return fmt.Errorf("read payment session: %w", read.Error)
		}
		if read.RowsAffected == 0 {
			payable = false
			return nil
		}
		session = fresh
		return nil
	})
	switch {
	case err != nil:
		return Session{}, false, err
	case !payable:
		return Session{}, false, ErrNotPayable
	}
	return session, created, nil
}
