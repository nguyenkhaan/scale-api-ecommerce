package payment

import (
	"bytes"
	"fmt"

	"gorm.io/gorm"
)

const reasonPayloadConflict = "event_payload_conflict"

// resolveDuplicate answers a verified event whose event_id was already stored. A genuine redelivery carries the
// same content and is a plain duplicate. The same id with different content contradicts the stored event
// (BR-47): it answers conflict and opens a reconciliation case for the stored event, and changes no order.
func resolveDuplicate(tx *gorm.DB, cb callback, payloadHash []byte, receipt *EventReceipt) error {
	var stored []struct {
		ID          string  `gorm:"column:id"`
		PayloadHash []byte  `gorm:"column:payload_hash"`
		OrderID     *string `gorm:"column:order_id"`
		OrderStatus *string `gorm:"column:order_status"`
	}
	err := tx.Raw(
		`SELECT e.id, e.payload_hash, ps.order_id, o.status AS order_status FROM payment_events e
		 LEFT JOIN payment_sessions ps ON ps.id = e.payment_session_id
		 LEFT JOIN orders o ON o.id = ps.order_id
		 WHERE e.event_id = ? AND e.source_verified`, cb.EventID,
	).Scan(&stored).Error
	if err != nil {
		return fmt.Errorf("read stored event: %w", err)
	}
	receipt.Result = ResultDuplicate
	if len(stored) == 0 {
		return nil
	}
	first := stored[0]
	if first.OrderStatus != nil {
		receipt.OrderStatus = *first.OrderStatus
	}
	if bytes.Equal(first.PayloadHash, payloadHash) {
		return nil
	}
	receipt.Result = ResultConflict
	if first.OrderID == nil {
		return nil
	}
	err = tx.Exec(
		`INSERT INTO payment_reconciliations (payment_event_id, order_id, reason) VALUES (?, ?, ?)
		 ON CONFLICT (payment_event_id) DO NOTHING`, first.ID, *first.OrderID, reasonPayloadConflict,
	).Error
	if err != nil {
		return fmt.Errorf("insert reconciliation: %w", err)
	}
	receipt.ReconciliationRequired = true
	return nil
}
