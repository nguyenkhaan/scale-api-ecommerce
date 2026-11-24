package backend_test

import (
	"net/http"
	"testing"

	"cloudian/cloudian-scale/internal/module/order"
	"cloudian/cloudian-scale/internal/module/payment"
	"gorm.io/gorm"
	"uuid"
)

// event delivers one validly signed event for the session and returns the HTTP status and receipt.
// It never touches testing.T so it is safe to call from goroutines.
func (fixture paymentFixture) event(session payment.Session, eventType, eventID string) (int, payment.EventReceipt) {
	response := fixture.env.deliver(paymentEventBody(session.ID, eventID, eventType, session.Amount, session.Currency))
	var receipt payment.EventReceipt
	if response.Code == http.StatusAccepted {
		receipt = decodeResponseBody[payment.EventReceipt](response)
	}
	return response.Code, receipt
}

// insertSession creates the session directly, for orders that are already past their deadline.
func (fixture paymentFixture) insertSession(t *testing.T) payment.Session {
	t.Helper()
	var session payment.Session
	if err := fixture.env.db.Raw(
		`INSERT INTO payment_sessions (order_id, amount, simulator_reference) VALUES (?, ?, 'SIM-' || gen_random_uuid())
		 RETURNING id, order_id, status, amount, currency, simulator_reference`, fixture.orderID, paymentOrderTotal,
	).Scan(&session).Error; err != nil {
		t.Fatalf("insert session: %v", err)
	}
	return session
}

func (fixture paymentFixture) finish(t *testing.T, outcome order.Outcome) order.Transition {
	t.Helper()
	var transition order.Transition
	if err := fixture.env.db.Transaction(func(tx *gorm.DB) error {
		var err error
		transition, err = order.Finish(tx, fixture.orderID, outcome)
		return err
	}); err != nil {
		t.Fatalf("finish %s: %v", outcome, err)
	}
	return transition
}

func paymentTally(receipts []payment.EventReceipt) map[string]int {
	tally := map[string]int{}
	for _, receipt := range receipts {
		tally[receipt.Result]++
	}
	return tally
}

func paymentRequireAccepted(t *testing.T, code int) {
	t.Helper()
	if code != http.StatusAccepted {
		t.Fatalf("callback status = %d, want 202", code)
	}
}

func TestPaymentSameSuccessEventDeliveredRepeatedlyAppliesOnce(t *testing.T) {
	fixture := newPaymentEnv(t).newFixture(t, 300)
	session := fixture.createSession(t)
	eventID := uuid.New().String()

	receipts := make([]payment.EventReceipt, 6)
	codes := make([]int, 6)
	cartRunParallel(6, func(i int) { codes[i], receipts[i] = fixture.event(session, "succeeded", eventID) })
	for range 2 {
		code, receipt := fixture.event(session, "succeeded", eventID)
		codes, receipts = append(codes, code), append(receipts, receipt)
	}
	for _, code := range codes {
		paymentRequireAccepted(t, code)
	}
	if tally := paymentTally(receipts); tally[payment.ResultApplied] != 1 || tally[payment.ResultDuplicate] != 7 {
		t.Errorf("results = %v, want 1 applied and 7 duplicate", tally)
	}
	for _, receipt := range receipts {
		if receipt.OrderStatus != "confirmed" {
			t.Errorf("receipt order_status = %q, want confirmed", receipt.OrderStatus)
		}
	}
	if status := fixture.orderStatus(t); status != "confirmed" {
		t.Errorf("order status = %q", status)
	}
	if rows := fixture.count(t, `SELECT count(*) FROM payment_events WHERE event_id = ?`, eventID); rows != 1 {
		t.Errorf("stored events = %d, want 1", rows)
	}
	if status := fixture.sessionStatus(t, session.ID); status != "succeeded" {
		t.Errorf("session status = %q, want succeeded", status)
	}
	fixture.requireSold(t)
	lifecycleRequireInvariants(t, fixture.env.db)
}

func TestPaymentDistinctSuccessEventsOfOnePaymentConfirmOnce(t *testing.T) {
	fixture := newPaymentEnv(t).newFixture(t, 300)
	session := fixture.createSession(t)

	receipts := make([]payment.EventReceipt, 6)
	cartRunParallel(6, func(i int) { _, receipts[i] = fixture.event(session, "succeeded", uuid.New().String()) })
	if tally := paymentTally(receipts); tally[payment.ResultApplied] != 1 || tally[payment.ResultDuplicate] != 5 {
		t.Errorf("results = %v, want 1 applied and 5 duplicate", tally)
	}
	if rows := fixture.reconciliations(t, "late_success"); rows != 0 {
		t.Errorf("late_success cases = %d, want 0", rows)
	}
	fixture.requireSold(t)
	lifecycleRequireInvariants(t, fixture.env.db)
}

func TestPaymentRepeatedFailureReleasesOnce(t *testing.T) {
	fixture := newPaymentEnv(t).newFixture(t, 300)
	session := fixture.createSession(t)
	eventID := uuid.New().String()

	receipts := make([]payment.EventReceipt, 5)
	cartRunParallel(5, func(i int) { _, receipts[i] = fixture.event(session, "failed", eventID) })
	_, other := fixture.event(session, "failed", uuid.New().String())
	receipts = append(receipts, other)
	if tally := paymentTally(receipts); tally[payment.ResultApplied] != 1 || tally[payment.ResultDuplicate] != 5 {
		t.Errorf("results = %v, want 1 applied and 5 duplicate", tally)
	}
	if status := fixture.orderStatus(t); status != "payment_failed" {
		t.Errorf("order status = %q, want payment_failed", status)
	}
	if status := fixture.sessionStatus(t, session.ID); status != "failed" {
		t.Errorf("session status = %q, want failed", status)
	}
	fixture.requireReleased(t)
	lifecycleRequireInvariants(t, fixture.env.db)
}

func TestPaymentSuccessAfterTerminalOrderIsLateAndNeverRevives(t *testing.T) {
	cases := map[string]struct {
		expiresIn float64
		prepare   func(t *testing.T, fixture paymentFixture)
		wantOrder string
	}{
		"cancelled":                           {300, func(t *testing.T, f paymentFixture) { f.finish(t, order.OutcomeCancelled) }, "cancelled"},
		"expired by the worker":               {-1, func(t *testing.T, f paymentFixture) { f.finish(t, order.OutcomeExpired) }, "expired"},
		"deadline passed before the callback": {-1, func(*testing.T, paymentFixture) {}, "expired"},
		"payment already failed":              {300, func(t *testing.T, f paymentFixture) { f.finish(t, order.OutcomePaymentFailed) }, "payment_failed"},
	}
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			fixture := newPaymentEnv(t).newFixture(t, test.expiresIn)
			session := fixture.insertSession(t)
			test.prepare(t, fixture)

			code, receipt := fixture.event(session, "succeeded", uuid.New().String())
			paymentRequireAccepted(t, code)
			if receipt.Result != payment.ResultLate || receipt.OrderStatus != test.wantOrder || !receipt.ReconciliationRequired {
				t.Errorf("receipt = %+v, want late on %s with reconciliation", receipt, test.wantOrder)
			}
			if rows := fixture.reconciliations(t, "late_success"); rows != 1 {
				t.Errorf("late_success cases = %d, want 1", rows)
			}
			if status := fixture.orderStatus(t); status != test.wantOrder {
				t.Errorf("order status = %q, want it to stay %s", status, test.wantOrder)
			}
			if status := fixture.sessionStatus(t, session.ID); status != "succeeded" {
				t.Errorf("session status = %q, want succeeded (money was taken)", status)
			}
			fixture.requireReleased(t)
			lifecycleRequireInvariants(t, fixture.env.db)
		})
	}
}

func TestPaymentFailureAfterConfirmedIsAConflictAndKeepsTheOrder(t *testing.T) {
	fixture := newPaymentEnv(t).newFixture(t, 300)
	session := fixture.createSession(t)
	if _, receipt := fixture.event(session, "succeeded", uuid.New().String()); receipt.Result != payment.ResultApplied {
		t.Fatalf("success receipt = %+v", receipt)
	}

	code, receipt := fixture.event(session, "failed", uuid.New().String())
	paymentRequireAccepted(t, code)
	if receipt.Result != payment.ResultConflict || receipt.OrderStatus != "confirmed" || !receipt.ReconciliationRequired {
		t.Errorf("receipt = %+v, want conflict on confirmed with reconciliation", receipt)
	}
	if rows := fixture.reconciliations(t, "failure_after_success"); rows != 1 {
		t.Errorf("failure_after_success cases = %d, want 1", rows)
	}
	if status := fixture.sessionStatus(t, session.ID); status != "succeeded" {
		t.Errorf("session status = %q, want it to stay succeeded", status)
	}
	fixture.requireSold(t)
	lifecycleRequireInvariants(t, fixture.env.db)
}
