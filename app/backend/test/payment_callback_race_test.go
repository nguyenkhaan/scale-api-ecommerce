package backend_test

import (
	"gorm.io/gorm"
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/module/order"
	"cloudian/cloudian-scale/internal/module/payment"
	"uuid"
)

// A success callback and a customer cancel compete for the same pending order: exactly one wins, and the
// counters show either a sale or a release, never both.
func TestPaymentSuccessRacingCancelEndsInExactlyOneState(t *testing.T) {
	env := newPaymentEnv(t)
	seen := map[string]int{}
	for round := 0; round < 16; round++ {
		fixture := env.newFixture(t, 300)
		session := fixture.createSession(t)
		var receipt payment.EventReceipt
		var cancel order.Transition
		cartRunParallel(2, func(i int) {
			if i == 0 {
				_, receipt = fixture.event(session, "succeeded", uuid.New().String())
				return
			}
			time.Sleep(time.Duration(round) * 600 * time.Microsecond) // sweeps the cancel across the callback's lifetime
			cancel, _ = paymentFinishUnlocked(fixture, order.OutcomeCancelled)
		})

		status := fixture.orderStatus(t)
		seen[status]++
		switch status {
		case "confirmed":
			if receipt.Result != payment.ResultApplied || cancel.Applied {
				t.Errorf("round %d: confirmed but receipt=%+v cancel=%+v", round, receipt, cancel)
			}
			fixture.requireSold(t)
		case "cancelled":
			if receipt.Result != payment.ResultLate || !cancel.Applied {
				t.Errorf("round %d: cancelled but receipt=%+v cancel=%+v", round, receipt, cancel)
			}
			if rows := fixture.reconciliations(t, "late_success"); rows != 1 {
				t.Errorf("round %d: late_success cases = %d, want 1", round, rows)
			}
			fixture.requireReleased(t)
		default:
			t.Fatalf("round %d: order ended %q", round, status)
		}
		lifecycleRequireInvariants(t, env.db)
	}
	t.Logf("race outcomes: %v", seen)
}

// A success callback and the expiry path both see an overdue order: it expires once, nothing is sold.
func TestPaymentSuccessRacingExpiryNeverSells(t *testing.T) {
	env := newPaymentEnv(t)
	for round := 0; round < 8; round++ {
		fixture := env.newFixture(t, -1)
		session := fixture.insertSession(t)
		var receipt payment.EventReceipt
		var expiry order.Transition
		cartRunParallel(2, func(i int) {
			if i == 0 {
				_, receipt = fixture.event(session, "succeeded", uuid.New().String())
				return
			}
			expiry, _ = paymentFinishUnlocked(fixture, order.OutcomeExpired)
		})

		if status := fixture.orderStatus(t); status != "expired" {
			t.Fatalf("round %d: order ended %q, want expired", round, status)
		}
		if receipt.Result != payment.ResultLate || expiry.Status != "expired" {
			t.Errorf("round %d: receipt=%+v expiry=%+v", round, receipt, expiry)
		}
		if rows := fixture.reconciliations(t, "late_success"); rows != 1 {
			t.Errorf("round %d: late_success cases = %d, want 1", round, rows)
		}
		fixture.requireReleased(t)
		lifecycleRequireInvariants(t, env.db)
	}
}

// paymentFinishUnlocked is goroutine-safe: it reports errors through the returned value, not testing.T.
func paymentFinishUnlocked(fixture paymentFixture, outcome order.Outcome) (order.Transition, error) {
	var transition order.Transition
	err := fixture.env.db.Transaction(func(tx *gorm.DB) error {
		var err error
		transition, err = order.Finish(tx, fixture.orderID, outcome)
		return err
	})
	return transition, err
}
