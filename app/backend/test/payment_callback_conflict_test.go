package backend_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/module/payment"
	"uuid"
)

func TestPaymentOversizedCallbackIsRejectedWith413(t *testing.T) {
	fixture := newPaymentEnv(t).newFixture(t, 300)
	session := fixture.createSession(t)
	body := paymentEventBody(session.ID, uuid.New().String(), "succeeded", session.Amount, session.Currency)
	body = body[:len(body)-1] + `,"padding":"` + strings.Repeat("x", 70<<10) + `"}`

	response := fixture.env.deliver(body)
	if response.Code != http.StatusRequestEntityTooLarge || decodeJSON[problem.Problem](t, response).Code != problem.CodeValidation {
		t.Fatalf("oversized body = %d, want 413 VALIDATION_ERROR", response.Code)
	}
	if rows := fixture.count(t, `SELECT count(*) FROM payment_events`); rows != 0 {
		t.Errorf("stored events = %d, want 0", rows)
	}
}

// The same event_id carrying different content contradicts the stored event: conflict, a case for the stored
// event, and no order change.
func TestPaymentReusedEventIDWithDifferentContentIsAConflict(t *testing.T) {
	fixture := newPaymentEnv(t).newFixture(t, 300)
	session := fixture.createSession(t)
	eventID := uuid.New().String()
	if _, receipt := fixture.event(session, "failed", eventID); receipt.Result != payment.ResultApplied {
		t.Fatalf("first event = %+v, want applied", receipt)
	}

	for range 2 { // the second contradiction must not break on the existing case
		code, receipt := fixture.event(session, "succeeded", eventID)
		paymentRequireAccepted(t, code)
		if receipt.Result != payment.ResultConflict || !receipt.ReconciliationRequired || receipt.OrderStatus != "payment_failed" {
			t.Errorf("contradicting event = %+v, want conflict on payment_failed with reconciliation", receipt)
		}
	}
	if rows := fixture.reconciliations(t, "event_payload_conflict"); rows != 1 {
		t.Errorf("event_payload_conflict cases = %d, want 1", rows)
	}
	if status := fixture.orderStatus(t); status != "payment_failed" {
		t.Errorf("order status = %q, want it unchanged", status)
	}
	if rows := fixture.events(t, eventID, true); rows != 1 {
		t.Errorf("verified rows = %d, want 1", rows)
	}
	fixture.requireReleased(t)
	lifecycleRequireInvariants(t, fixture.env.db)
}

func TestPaymentIdenticalRedeliveryWithNewTimestampStaysADuplicate(t *testing.T) {
	fixture := newPaymentEnv(t).newFixture(t, 300)
	session := fixture.createSession(t)
	eventID := uuid.New().String()
	fixture.event(session, "succeeded", eventID)
	time.Sleep(5 * time.Millisecond) // a later occurred_at must not turn a redelivery into a conflict

	if _, receipt := fixture.event(session, "succeeded", eventID); receipt.Result != payment.ResultDuplicate || receipt.ReconciliationRequired {
		t.Errorf("redelivery = %+v, want a plain duplicate", receipt)
	}
	if rows := fixture.reconciliations(t, "event_payload_conflict"); rows != 0 {
		t.Errorf("event_payload_conflict cases = %d, want 0", rows)
	}
}
