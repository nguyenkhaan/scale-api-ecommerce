package payment

import (
	"context"
	"fmt"

	"cloudian/cloudian-scale/internal/module/order"
	"gorm.io/gorm"
)

const (
	ResultApplied   = "applied"
	ResultDuplicate = "duplicate"
	ResultRejected  = "rejected"
	ResultLate      = "late"
	ResultConflict  = "conflict"

	reasonInvalidSignature = "invalid_signature"
	reasonUnknownSession   = "unknown_session"
	reasonAmountMismatch   = "amount_mismatch"
	reasonCurrencyMismatch = "currency_mismatch"
	reasonLateSuccess      = "late_success"
	reasonFailureAfterOK   = "failure_after_success"
)

// verdict is what one verified event did and whether the case needs a person to look at it.
type verdict struct {
	result      string
	reason      string
	reconcile   string
	orderStatus string
	// sessionStatus is the status to move the payment session to; empty leaves it unchanged.
	sessionStatus string
}

// RecordRejectedDelivery stores a delivery whose signature failed. It is kept for audit as unverified, so it
// can never block the genuine event that carries the same event_id later. The claimed session id is untrusted,
// so it is not linked: the audit trail is by event_id.
func (service *Service) RecordRejectedDelivery(ctx context.Context, cb callback, payloadHash []byte) error {
	err := service.db.WithContext(ctx).Exec(
		`INSERT INTO payment_events (event_id, payment_session_id, event_type, amount, currency, payload_hash,
			source_verified, result, reason_code, occurred_at)
		 VALUES (?, NULL, ?, ?, ?, ?, false, 'rejected', ?, ?)`,
		cb.EventID, cb.EventType, cb.Amount, cb.Currency, payloadHash, reasonInvalidSignature, cb.OccurredAt,
	).Error
	if err != nil {
		return fmt.Errorf("store rejected delivery: %w", err)
	}
	return nil
}

// ApplyCallback handles one signature-verified event in a single transaction: the event row, the order
// transition, the reconciliation case and the session status commit together, or not at all (BR-49).
func (service *Service) ApplyCallback(ctx context.Context, cb callback, payloadHash []byte) (EventReceipt, error) {
	var receipt EventReceipt
	err := service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var session Session
		found := false
		read := tx.Raw(`SELECT `+sessionColumns+` FROM payment_sessions WHERE id = ?`, cb.SessionID).Scan(&session)
		if read.Error != nil {
			return fmt.Errorf("read payment session: %w", read.Error)
		}
		found = read.RowsAffected > 0

		var sessionRef any
		if found {
			sessionRef = session.ID
		}
		var inserted []string
		if err := tx.Raw(
			`INSERT INTO payment_events (event_id, payment_session_id, event_type, amount, currency, payload_hash,
				source_verified, result, occurred_at)
			 VALUES (?, ?, ?, ?, ?, ?, true, 'applied', ?)
			 ON CONFLICT (event_id) WHERE source_verified DO NOTHING RETURNING id`,
			cb.EventID, sessionRef, cb.EventType, cb.Amount, cb.Currency, payloadHash, cb.OccurredAt,
		).Scan(&inserted).Error; err != nil {
			return fmt.Errorf("insert payment event: %w", err)
		}
		receipt = EventReceipt{EventID: cb.EventID}
		if len(inserted) == 0 {
			return resolveDuplicate(tx, cb, payloadHash, &receipt)
		}

		decided, err := decide(tx, cb, session, found)
		if err != nil {
			return err
		}
		if decided.reconcile != "" {
			if err := tx.Exec(
				`INSERT INTO payment_reconciliations (payment_event_id, order_id, reason) VALUES (?, ?, ?)`,
				inserted[0], session.OrderID, decided.reconcile,
			).Error; err != nil {
				return fmt.Errorf("insert reconciliation: %w", err)
			}
		}
		// The session is locked last (order, reservations, ... then session) to keep one lock order.
		if err := updateSessionStatus(tx, session.ID, decided.sessionStatus); err != nil {
			return err
		}
		if err := tx.Exec(`UPDATE payment_events SET result = ?, reason_code = NULLIF(?, '') WHERE id = ?`,
			decided.result, decided.reason, inserted[0]).Error; err != nil {
			return fmt.Errorf("finalize payment event: %w", err)
		}
		receipt.Result, receipt.OrderStatus = decided.result, decided.orderStatus
		receipt.ReconciliationRequired = decided.reconcile != ""
		return nil
	})
	return receipt, err
}

func decide(tx *gorm.DB, cb callback, session Session, found bool) (verdict, error) {
	if !found {
		return verdict{result: ResultRejected, reason: reasonUnknownSession}, nil
	}
	if reason := mismatchReason(cb, session); reason != "" {
		var status string
		if err := tx.Raw(`SELECT status FROM orders WHERE id = ?`, session.OrderID).Scan(&status).Error; err != nil {
			return verdict{}, fmt.Errorf("read order status: %w", err)
		}
		return verdict{result: ResultRejected, reason: reason, reconcile: reason, orderStatus: status}, nil
	}
	outcome := order.OutcomePaymentFailed
	if cb.EventType == EventSucceeded {
		outcome = order.OutcomeConfirmed
	}
	transition, err := order.Finish(tx, session.OrderID, outcome)
	if err != nil {
		return verdict{}, fmt.Errorf("finish order: %w", err)
	}
	decided := classify(cb.EventType, transition)
	decided.orderStatus = transition.Status
	return decided, nil
}

func mismatchReason(cb callback, session Session) string {
	switch {
	case cb.Amount != session.Amount:
		return reasonAmountMismatch
	case cb.Currency != session.Currency:
		return reasonCurrencyMismatch
	}
	return ""
}

// classify maps an event type and what Finish did to the stored result (PRD BR-43, BR-47). A confirmed order
// is never undone by a failure and a terminal order is never revived by a success.
func classify(eventType string, transition order.Transition) verdict {
	if eventType == EventSucceeded {
		decided := verdict{sessionStatus: EventSucceeded}
		switch {
		case transition.Applied && transition.Status == string(order.OutcomeConfirmed):
			decided.result = ResultApplied
		case !transition.Applied && transition.Status == string(order.OutcomeConfirmed):
			decided.result, decided.reason = ResultDuplicate, "order_already_confirmed"
		default:
			decided.result, decided.reason, decided.reconcile = ResultLate, reasonLateSuccess, reasonLateSuccess
		}
		return decided
	}
	decided := verdict{sessionStatus: EventFailed}
	switch {
	case transition.Applied && transition.Status == string(order.OutcomePaymentFailed):
		decided.result = ResultApplied
	case !transition.Applied && transition.Status == string(order.OutcomePaymentFailed):
		decided.result, decided.reason = ResultDuplicate, "order_already_failed"
	case !transition.Applied && transition.Status == string(order.OutcomeConfirmed):
		decided.result, decided.reason, decided.reconcile = ResultConflict, reasonFailureAfterOK, reasonFailureAfterOK
	default:
		decided.result, decided.reason = ResultLate, "order_already_terminal"
	}
	return decided
}

func updateSessionStatus(tx *gorm.DB, sessionID, status string) error {
	if status == "" {
		return nil
	}
	// A verified failure never downgrades a session that already succeeded.
	err := tx.Exec(
		`UPDATE payment_sessions SET status = ?, updated_at = now() WHERE id = ? AND (? = 'succeeded' OR status <> 'succeeded')`,
		status, sessionID, status,
	).Error
	if err != nil {
		return fmt.Errorf("update payment session: %w", err)
	}
	return nil
}
