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

func (fixture paymentFixture) events(t *testing.T, eventID string, verified bool) int64 {
	t.Helper()
	return fixture.count(t, `SELECT count(*) FROM payment_events WHERE event_id = ? AND source_verified = ?`, eventID, verified)
}

func requireUnauthenticated(t *testing.T, label string, code int, body string) {
	t.Helper()
	if code != http.StatusUnauthorized {
		t.Errorf("%s = %d, want 401; body = %s", label, code, body)
	}
}

func TestPaymentBadSignatureIsStoredUnverifiedAndDoesNotBlockTheGenuineEvent(t *testing.T) {
	fixture := newPaymentEnv(t).newFixture(t, 300)
	session := fixture.createSession(t)
	eventID := uuid.New().String()
	body := paymentEventBody(session.ID, eventID, "succeeded", session.Amount, session.Currency)

	response := fixture.env.callback(body, payment.Sign("not-the-secret", time.Now(), []byte(body)))
	requireUnauthenticated(t, "bad signature", response.Code, response.Body.String())
	if code := decodeJSON[problem.Problem](t, response).Code; code != problem.CodeUnauthenticated {
		t.Errorf("problem code = %q, want UNAUTHENTICATED", code)
	}
	stored := fixture.count(t,
		`SELECT count(*) FROM payment_events WHERE event_id = ? AND NOT source_verified AND result = 'rejected'
		 AND reason_code = 'invalid_signature' AND payment_session_id IS NULL`, eventID)
	if stored != 1 {
		t.Errorf("unverified rejected rows = %d, want 1", stored)
	}
	if status := fixture.orderStatus(t); status != "pending_payment" {
		t.Errorf("order status = %q after a forged callback", status)
	}
	fixture.requireHeld(t)

	code, receipt := fixture.event(session, "succeeded", eventID)
	paymentRequireAccepted(t, code)
	if receipt.Result != payment.ResultApplied {
		t.Errorf("genuine event with the same id = %+v, want applied", receipt)
	}
	if fixture.events(t, eventID, true) != 1 || fixture.events(t, eventID, false) != 1 {
		t.Errorf("want one verified and one unverified row for %s", eventID)
	}
	fixture.requireSold(t)
	lifecycleRequireInvariants(t, fixture.env.db)
}

func TestPaymentRejectsStaleMissingAndTamperedSignatures(t *testing.T) {
	fixture := newPaymentEnv(t).newFixture(t, 300)
	session := fixture.createSession(t)
	body := paymentEventBody(session.ID, uuid.New().String(), "succeeded", session.Amount, session.Currency)
	tampered := paymentEventBody(session.ID, uuid.New().String(), "succeeded", 1, session.Currency)

	for name, header := range map[string]string{
		"stale timestamp":  payment.Sign(paymentTestSecret, time.Now().Add(-400*time.Second), []byte(body)),
		"future timestamp": payment.Sign(paymentTestSecret, time.Now().Add(400*time.Second), []byte(body)),
		"missing header":   "",
		"malformed header": "v1=zz",
	} {
		response := fixture.env.callback(body, header)
		requireUnauthenticated(t, name, response.Code, response.Body.String())
	}
	response := fixture.env.callback(tampered, payment.Sign(paymentTestSecret, time.Now(), []byte(body)))
	requireUnauthenticated(t, "signature of a different body", response.Code, response.Body.String())

	if status := fixture.orderStatus(t); status != "pending_payment" {
		t.Errorf("order status = %q after rejected deliveries", status)
	}
	if verified := fixture.count(t, `SELECT count(*) FROM payment_events WHERE source_verified`); verified != 0 {
		t.Errorf("verified events = %d, want 0", verified)
	}
	fixture.requireHeld(t)
}

func TestPaymentMalformedCallbacksAreValidationErrorsAndNotStored(t *testing.T) {
	fixture := newPaymentEnv(t).newFixture(t, 300)
	session := fixture.createSession(t)
	valid := paymentEventBody(session.ID, uuid.New().String(), "succeeded", session.Amount, session.Currency)
	bodies := map[string]string{
		"not json":      "not json",
		"missing field": `{"event_id":"` + uuid.New().String() + `"}`,
		"unknown field": valid[:len(valid)-1] + `,"extra":1}`,
		"zero amount":   paymentEventBody(session.ID, uuid.New().String(), "succeeded", 0, "VND"),
		"bad type":      paymentEventBody(session.ID, uuid.New().String(), "refunded", session.Amount, "VND"),
		"bad event id":  paymentEventBody(session.ID, "nope", "succeeded", session.Amount, "VND"),
		"bad currency":  paymentEventBody(session.ID, uuid.New().String(), "succeeded", session.Amount, "VN"),
	}
	for name, body := range bodies {
		signed := fixture.env.callback(body, payment.Sign(paymentTestSecret, time.Now(), []byte(body)))
		forged := fixture.env.callback(body, payment.Sign("not-the-secret", time.Now(), []byte(body)))
		if signed.Code != http.StatusUnprocessableEntity {
			t.Errorf("%s (signed) = %d, want 422", name, signed.Code)
		}
		// An invalid signature is always a plain 401: no field-level hints for unauthenticated callers.
		if forged.Code != http.StatusUnauthorized || strings.Contains(forged.Body.String(), "violations") {
			t.Errorf("%s (forged) = %d %s, want 401 without violations", name, forged.Code, forged.Body.String())
		}
	}
	if rows := fixture.count(t, `SELECT count(*) FROM payment_events`); rows != 0 {
		t.Errorf("stored events = %d, want 0 for malformed callbacks", rows)
	}
}

func TestPaymentMismatchedCallbacksAreRejectedAndReconciled(t *testing.T) {
	cases := map[string]struct {
		amount   int64
		currency string
		reason   string
	}{
		"amount":   {paymentOrderTotal + 1, "VND", "amount_mismatch"},
		"currency": {paymentOrderTotal, "USD", "currency_mismatch"},
	}
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			fixture := newPaymentEnv(t).newFixture(t, 300)
			session := fixture.createSession(t)
			body := paymentEventBody(session.ID, uuid.New().String(), "succeeded", test.amount, test.currency)

			response := fixture.env.deliver(body)
			paymentRequireAccepted(t, response.Code)
			receipt := decodeResponseBody[payment.EventReceipt](response)
			if receipt.Result != payment.ResultRejected || !receipt.ReconciliationRequired || receipt.OrderStatus != "pending_payment" {
				t.Errorf("receipt = %+v, want rejected with reconciliation on a pending order", receipt)
			}
			if rows := fixture.reconciliations(t, test.reason); rows != 1 {
				t.Errorf("%s cases = %d, want 1", test.reason, rows)
			}
			if rows := fixture.count(t, `SELECT count(*) FROM payment_events WHERE result = 'rejected' AND reason_code = ?`, test.reason); rows != 1 {
				t.Errorf("stored %s events = %d, want 1", test.reason, rows)
			}
			if status := fixture.orderStatus(t); status != "pending_payment" {
				t.Errorf("order status = %q, a mismatched callback must not confirm", status)
			}
			if status := fixture.sessionStatus(t, session.ID); status != "pending" {
				t.Errorf("session status = %q, want pending", status)
			}
			fixture.requireHeld(t)
			lifecycleRequireInvariants(t, fixture.env.db)
		})
	}
}

func TestPaymentCallbackForUnknownSessionIsRejected(t *testing.T) {
	fixture := newPaymentEnv(t).newFixture(t, 300)
	fixture.createSession(t)
	body := paymentEventBody(uuid.New().String(), uuid.New().String(), "succeeded", paymentOrderTotal, "VND")

	response := fixture.env.deliver(body)
	paymentRequireAccepted(t, response.Code)
	if receipt := decodeResponseBody[payment.EventReceipt](response); receipt.Result != payment.ResultRejected || receipt.OrderStatus != "" {
		t.Errorf("receipt = %+v, want rejected without an order", receipt)
	}
	if rows := fixture.count(t, `SELECT count(*) FROM payment_events WHERE result = 'rejected' AND reason_code = 'unknown_session' AND payment_session_id IS NULL`); rows != 1 {
		t.Errorf("unknown_session events = %d, want 1", rows)
	}
	if status := fixture.orderStatus(t); status != "pending_payment" {
		t.Errorf("order status = %q", status)
	}
	fixture.requireHeld(t)
}
