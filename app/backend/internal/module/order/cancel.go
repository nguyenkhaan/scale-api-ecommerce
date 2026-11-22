package order

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// StateConflictError means the order already reached a terminal state that forbids the request.
type StateConflictError struct {
	Status string
}

func (e *StateConflictError) Error() string {
	return fmt.Sprintf("order is already %s", e.Status)
}

// CancelOrder cancels the caller's pending order through the guarded Finish transition.
// Repeating a cancel is idempotent; any other terminal state (including an expiry that Finish just
// applied because the deadline passed) is reported as a StateConflictError after that effect commits.
func (q *OrderQueries) CancelOrder(ctx context.Context, userID, orderID string) (OrderResponse, error) {
	var transition Transition
	err := q.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Ownership is checked before Finish so another customer's request can never change the order.
		var owned int64
		if err := tx.Raw(`SELECT count(*) FROM orders WHERE id = ? AND user_id = ?`, orderID, userID).
			Scan(&owned).Error; err != nil {
			return fmt.Errorf("check order owner: %w", err)
		}
		if owned == 0 {
			return ErrOrderNotFound
		}
		var err error
		transition, err = Finish(tx, orderID, OutcomeCancelled)
		return err
	})
	if err != nil {
		return OrderResponse{}, err
	}
	if transition.Status != string(OutcomeCancelled) {
		return OrderResponse{}, &StateConflictError{Status: transition.Status}
	}
	return getOwnedOrder(ctx, q.db, userID, orderID)
}
