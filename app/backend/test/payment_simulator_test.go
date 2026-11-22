package backend_test

import (
	"net/http"
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/module/payment"
	"uuid"
)

const paymentScheduleOK = http.StatusAccepted

func (fixture paymentFixture) schedule(t *testing.T, client *testClient, request map[string]any, headers ...string) (int, payment.ScheduledEvent) {
	t.Helper()
	response := client.send(t, http.MethodPost, "/api/payment-simulator/events", request, headers...)
	var scheduled payment.ScheduledEvent
	if response.Code == paymentScheduleOK {
		scheduled = decodeJSON[payment.ScheduledEvent](t, response)
	}
	return response.Code, scheduled
}

func paymentOperatorHeader() []string { return []string{"X-Operator-Key", paymentTestOperatorKey} }

func TestPaymentSimulatorAccessRules(t *testing.T) {
	env := newPaymentEnv(t)
	fixture := env.newFixture(t, 300)
	session := fixture.createSession(t)
	request := map[string]any{"payment_session_id": session.ID, "event_type": "succeeded", "drop_delivery": true}
	admin := env.newAdmin(t)
	anonymous := newTestClient(env.router)

	if code, scheduled := fixture.schedule(t, admin, request, admin.csrfHeader()...); code != paymentScheduleOK || scheduled.EventID == "" {
		t.Errorf("admin session = %d %+v, want 202", code, scheduled)
	}
	if code, _ := fixture.schedule(t, admin, request); code != http.StatusForbidden {
		t.Errorf("admin without CSRF = %d, want 403", code)
	}
	if code, _ := fixture.schedule(t, anonymous, request, paymentOperatorHeader()...); code != paymentScheduleOK {
		t.Errorf("operator key = %d, want 202", code)
	}
	if code, _ := fixture.schedule(t, anonymous, request, "X-Operator-Key", "wrong"); code != http.StatusForbidden {
		t.Errorf("wrong operator key = %d, want 403", code)
	}
	if code, _ := fixture.schedule(t, anonymous, request); code != http.StatusUnauthorized {
		t.Errorf("anonymous = %d, want 401", code)
	}
	if code, _ := fixture.schedule(t, fixture.customer.testClient, request, fixture.customer.csrfHeader()...); code != http.StatusForbidden {
		t.Errorf("customer = %d, want 403", code)
	}
	if code := fixture.customer.send(t, http.MethodGet, "/api/payment-simulator/sessions", nil).Code; code != http.StatusForbidden {
		t.Errorf("customer listing = %d, want 403", code)
	}
}

func TestPaymentSimulatorDisabledIsNotFound(t *testing.T) {
	env := newPaymentEnv(t, func(cfg *payment.Config) { cfg.SimulatorEnabled = false })
	fixture := env.newFixture(t, 300)
	session := fixture.createSession(t)
	admin := env.newAdmin(t)
	request := map[string]any{"payment_session_id": session.ID, "event_type": "succeeded"}

	if code, _ := fixture.schedule(t, admin, request, admin.csrfHeader()...); code != http.StatusNotFound {
		t.Errorf("admin schedule = %d, want 404", code)
	}
	if code, _ := fixture.schedule(t, newTestClient(env.router), request, paymentOperatorHeader()...); code != http.StatusNotFound {
		t.Errorf("operator schedule = %d, want 404", code)
	}
	if code := admin.send(t, http.MethodGet, "/api/payment-simulator/sessions", nil).Code; code != http.StatusNotFound {
		t.Errorf("listing = %d, want 404", code)
	}
}

func TestPaymentSimulatorRejectsInvalidRequests(t *testing.T) {
	env := newPaymentEnv(t)
	fixture := env.newFixture(t, 300)
	session := fixture.createSession(t)
	operator := newTestClient(env.router)
	for name, request := range map[string]map[string]any{
		"unknown session":     {"payment_session_id": uuid.New().String(), "event_type": "succeeded"},
		"bad session id":      {"payment_session_id": "x", "event_type": "succeeded"},
		"bad event type":      {"payment_session_id": session.ID, "event_type": "refunded"},
		"delay too long":      {"payment_session_id": session.ID, "event_type": "failed", "delay_ms": 600001},
		"repeat too many":     {"payment_session_id": session.ID, "event_type": "failed", "repeat_count": 11},
		"bad signature":       {"payment_session_id": session.ID, "event_type": "failed", "signature_mode": "maybe"},
		"bad currency":        {"payment_session_id": session.ID, "event_type": "failed", "currency_override": "VN"},
		"unknown property":    {"payment_session_id": session.ID, "event_type": "failed", "oops": true},
		"non-positive amount": {"payment_session_id": session.ID, "event_type": "failed", "amount_override": 0},
	} {
		if code, _ := fixture.schedule(t, operator, request, paymentOperatorHeader()...); code != http.StatusUnprocessableEntity {
			t.Errorf("%s = %d, want 422", name, code)
		}
	}
}

func TestPaymentSimulatorDeliversThroughTheRealCallbackPath(t *testing.T) {
	env := newPaymentEnv(t)
	operator := newTestClient(env.router)

	t.Run("delayed success confirms the order", func(t *testing.T) {
		fixture := env.newFixture(t, 300)
		session := fixture.createSession(t)
		started := time.Now()
		code, scheduled := fixture.schedule(t, operator, map[string]any{
			"payment_session_id": session.ID, "event_type": "succeeded", "delay_ms": 150}, paymentOperatorHeader()...)
		if code != paymentScheduleOK || !scheduled.ScheduledAt.After(started) {
			t.Fatalf("schedule = %d %+v", code, scheduled)
		}
		if status := fixture.orderStatus(t); status != "pending_payment" {
			t.Errorf("order status = %q before the delay elapsed", status)
		}
		delivered := env.waitDelivered(t, 1, 3*time.Second)
		if elapsed := time.Since(started); elapsed < 150*time.Millisecond {
			t.Errorf("delivered after %s, want at least the 150ms delay", elapsed)
		}
		if delivered[0].Receipt.Result != payment.ResultApplied || delivered[0].Receipt.EventID != scheduled.EventID {
			t.Errorf("delivery = %+v", delivered[0])
		}
		fixture.requireSold(t)
		lifecycleRequireInvariants(t, env.db)
	})

	t.Run("repeats become duplicates", func(t *testing.T) {
		fixture := env.newFixture(t, 300)
		session := fixture.createSession(t)
		fixture.schedule(t, operator, map[string]any{
			"payment_session_id": session.ID, "event_type": "succeeded", "repeat_count": 4}, paymentOperatorHeader()...)
		tally := map[string]int{}
		for _, delivery := range env.waitDelivered(t, 4, 3*time.Second) {
			tally[delivery.Receipt.Result]++
		}
		if tally[payment.ResultApplied] != 1 || tally[payment.ResultDuplicate] != 3 {
			t.Errorf("results = %v, want 1 applied and 3 duplicate", tally)
		}
		fixture.requireSold(t)
	})

	t.Run("reused event id is a duplicate", func(t *testing.T) {
		fixture := env.newFixture(t, 300)
		session := fixture.createSession(t)
		eventID := uuid.New().String()
		for range 2 {
			fixture.schedule(t, operator, map[string]any{
				"payment_session_id": session.ID, "event_type": "succeeded", "event_id": eventID}, paymentOperatorHeader()...)
			env.waitDelivered(t, 1, 3*time.Second)
		}
		if rows := fixture.events(t, eventID, true); rows != 1 {
			t.Errorf("verified rows = %d, want 1", rows)
		}
		fixture.requireSold(t)
	})

	t.Run("dropped delivery sends nothing", func(t *testing.T) {
		fixture := env.newFixture(t, 300)
		session := fixture.createSession(t)
		fixture.schedule(t, operator, map[string]any{
			"payment_session_id": session.ID, "event_type": "succeeded", "drop_delivery": true}, paymentOperatorHeader()...)
		select {
		case delivery := <-env.sender.delivered:
			t.Fatalf("a dropped delivery arrived: %+v", delivery)
		case <-time.After(400 * time.Millisecond):
		}
		if status := fixture.orderStatus(t); status != "pending_payment" {
			t.Errorf("order status = %q, want pending_payment", status)
		}
		fixture.requireHeld(t)
	})

	t.Run("invalid signature is rejected", func(t *testing.T) {
		fixture := env.newFixture(t, 300)
		session := fixture.createSession(t)
		_, scheduled := fixture.schedule(t, operator, map[string]any{
			"payment_session_id": session.ID, "event_type": "succeeded", "signature_mode": "invalid"}, paymentOperatorHeader()...)
		if delivered := env.waitDelivered(t, 1, 3*time.Second); delivered[0].Code != http.StatusUnauthorized {
			t.Errorf("delivery answered %d, want 401", delivered[0].Code)
		}
		if rows := fixture.events(t, scheduled.EventID, false); rows != 1 {
			t.Errorf("unverified rows = %d, want 1", rows)
		}
		fixture.requireHeld(t)
	})

	t.Run("amount override is rejected and reconciled", func(t *testing.T) {
		fixture := env.newFixture(t, 300)
		session := fixture.createSession(t)
		fixture.schedule(t, operator, map[string]any{
			"payment_session_id": session.ID, "event_type": "succeeded", "amount_override": 1}, paymentOperatorHeader()...)
		if delivered := env.waitDelivered(t, 1, 3*time.Second); delivered[0].Receipt.Result != payment.ResultRejected {
			t.Errorf("delivery = %+v, want rejected", delivered[0])
		}
		if rows := fixture.reconciliations(t, "amount_mismatch"); rows != 1 {
			t.Errorf("amount_mismatch cases = %d, want 1", rows)
		}
		fixture.requireHeld(t)
	})
}

func TestPaymentSimulatorSessionListing(t *testing.T) {
	env := newPaymentEnv(t)
	first := env.newFixture(t, 300)
	firstSession := first.createSession(t)
	second := env.newFixture(t, 300)
	secondSession := second.createSession(t)
	first.event(firstSession, "succeeded", uuid.New().String())
	admin := env.newAdmin(t)

	type listing struct {
		Items []payment.SimulatorSession `json:"items"`
	}
	all := decodeJSON[listing](t, admin.send(t, http.MethodGet, "/api/payment-simulator/sessions", nil))
	if len(all.Items) != 2 || all.Items[0].ID != secondSession.ID || all.Items[1].ID != firstSession.ID {
		t.Fatalf("listing = %+v, want newest first [%s %s]", all.Items, secondSession.ID, firstSession.ID)
	}
	if item := all.Items[0]; item.OrderID != second.orderID || item.OrderNumber == "" || item.Amount != paymentOrderTotal ||
		item.Status != "pending" || item.OrderStatus != "pending_payment" || item.ExpiresAt.IsZero() {
		t.Errorf("item = %+v", item)
	}
	pending := decodeJSON[listing](t, newTestClient(env.router).send(t, http.MethodGet,
		"/api/payment-simulator/sessions?status=pending&limit=1", nil, paymentOperatorHeader()...))
	if len(pending.Items) != 1 || pending.Items[0].ID != secondSession.ID {
		t.Errorf("pending listing = %+v", pending.Items)
	}
	if code := admin.send(t, http.MethodGet, "/api/payment-simulator/sessions?status=bogus", nil).Code; code != http.StatusUnprocessableEntity {
		t.Errorf("bad status filter = %d, want 422", code)
	}
}
