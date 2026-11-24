package backend_test

import (
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/module/order"
	"cloudian/cloudian-scale/internal/module/payment"
	"gorm.io/gorm"
)

func (fixture paymentFixture) postSession(t *testing.T, client *testClient) int {
	t.Helper()
	return client.send(t, http.MethodPost, "/api/orders/"+fixture.orderID+"/payment-session", nil, client.csrfHeader()...).Code
}

func TestPaymentSessionRetryReturnsSameSessionAndKeepsDeadline(t *testing.T) {
	fixture := newPaymentEnv(t).newFixture(t, 300)
	var deadlineBefore time.Time
	fixture.env.db.Raw(`SELECT expires_at FROM orders WHERE id = ?`, fixture.orderID).Scan(&deadlineBefore)

	first := fixture.customer.send(t, http.MethodPost, "/api/orders/"+fixture.orderID+"/payment-session", nil, fixture.customer.csrfHeader()...)
	if first.Code != http.StatusCreated {
		t.Fatalf("first call = %d, want 201; body = %s", first.Code, first.Body.String())
	}
	created := decodeJSON[payment.Session](t, first)
	if created.Amount != paymentOrderTotal || created.Currency != "VND" || created.Status != "pending" ||
		created.OrderID != fixture.orderID || !strings.HasPrefix(created.SimulatorReference, "SIM-") {
		t.Fatalf("created session = %+v", created)
	}

	second := fixture.customer.send(t, http.MethodPost, "/api/orders/"+fixture.orderID+"/payment-session", nil, fixture.customer.csrfHeader()...)
	if second.Code != http.StatusOK {
		t.Fatalf("retry = %d, want 200", second.Code)
	}
	if retried := decodeJSON[payment.Session](t, second); retried != created {
		t.Errorf("retry returned %+v, want the same session %+v", retried, created)
	}
	var deadlineAfter time.Time
	fixture.env.db.Raw(`SELECT expires_at FROM orders WHERE id = ?`, fixture.orderID).Scan(&deadlineAfter)
	if !deadlineAfter.Equal(deadlineBefore) {
		t.Errorf("deadline moved from %s to %s", deadlineBefore, deadlineAfter)
	}
}

func TestPaymentSessionConcurrentFirstCallsCreateOneRow(t *testing.T) {
	fixture := newPaymentEnv(t).newFixture(t, 300)
	var created, existing, other atomic.Int64
	cartRunParallel(5, func(int) {
		switch fixture.postSession(t, fixture.customer.parallel(fixture.env.router)) {
		case http.StatusCreated:
			created.Add(1)
		case http.StatusOK:
			existing.Add(1)
		default:
			other.Add(1)
		}
	})
	if created.Load() != 1 || existing.Load() != 4 || other.Load() != 0 {
		t.Errorf("created/existing/other = %d/%d/%d, want 1/4/0", created.Load(), existing.Load(), other.Load())
	}
	if rows := fixture.count(t, `SELECT count(*) FROM payment_sessions WHERE order_id = ?`, fixture.orderID); rows != 1 {
		t.Errorf("payment_sessions rows = %d, want 1", rows)
	}
}

func TestPaymentSessionOfAnotherUsersOrderIsNotFound(t *testing.T) {
	env := newPaymentEnv(t)
	fixture := env.newFixture(t, 300)
	stranger := cartNewCustomer(t, env.router, env.db, "stranger")

	response := stranger.send(t, http.MethodPost, "/api/orders/"+fixture.orderID+"/payment-session", nil, stranger.csrfHeader()...)
	if response.Code != http.StatusNotFound || decodeJSON[problem.Problem](t, response).Code != problem.CodeResourceNotFound {
		t.Fatalf("stranger = %d %s, want 404 RESOURCE_NOT_FOUND", response.Code, response.Body.String())
	}
	if rows := fixture.count(t, `SELECT count(*) FROM payment_sessions`); rows != 0 {
		t.Errorf("a session was created for a non-owner: %d rows", rows)
	}
}

func TestPaymentSessionRequiresSessionAndCSRF(t *testing.T) {
	fixture := newPaymentEnv(t).newFixture(t, 300)
	anonymous := newTestClient(fixture.env.router)
	if code := anonymous.send(t, http.MethodPost, "/api/orders/"+fixture.orderID+"/payment-session", nil).Code; code != http.StatusUnauthorized {
		t.Errorf("anonymous = %d, want 401", code)
	}
	if code := fixture.customer.send(t, http.MethodPost, "/api/orders/"+fixture.orderID+"/payment-session", nil).Code; code != http.StatusForbidden {
		t.Errorf("without CSRF = %d, want 403", code)
	}
}

func TestPaymentSessionOnOverdueOrderExpiresItAndReturnsConflict(t *testing.T) {
	fixture := newPaymentEnv(t).newFixture(t, -1)

	response := fixture.customer.send(t, http.MethodPost, "/api/orders/"+fixture.orderID+"/payment-session", nil, fixture.customer.csrfHeader()...)
	if response.Code != http.StatusConflict || decodeJSON[problem.Problem](t, response).Code != problem.CodeOrderNotPayable {
		t.Fatalf("overdue = %d %s, want 409 ORDER_NOT_PAYABLE", response.Code, response.Body.String())
	}
	if status := fixture.orderStatus(t); status != "expired" {
		t.Errorf("order status = %q, want expired", status)
	}
	if rows := fixture.count(t, `SELECT count(*) FROM payment_sessions`); rows != 0 {
		t.Errorf("a session exists for an expired order: %d rows", rows)
	}
	fixture.requireReleased(t)
	lifecycleRequireInvariants(t, fixture.env.db)
}

func TestPaymentSessionOnTerminalOrderIsNotPayable(t *testing.T) {
	fixture := newPaymentEnv(t).newFixture(t, 300)
	session := fixture.createSession(t)
	if err := fixture.env.db.Transaction(func(tx *gorm.DB) error {
		_, err := order.Finish(tx, fixture.orderID, order.OutcomeCancelled)
		return err
	}); err != nil {
		t.Fatalf("cancel order: %v", err)
	}

	response := fixture.customer.send(t, http.MethodPost, "/api/orders/"+fixture.orderID+"/payment-session", nil, fixture.customer.csrfHeader()...)
	if response.Code != http.StatusConflict || decodeJSON[problem.Problem](t, response).Code != problem.CodeOrderNotPayable {
		t.Fatalf("cancelled order = %d %s, want 409 ORDER_NOT_PAYABLE", response.Code, response.Body.String())
	}
	if status := fixture.sessionStatus(t, session.ID); status != "pending" {
		t.Errorf("session status = %q, want pending", status)
	}
}

// A cancel that holds the order lock must finish before the session decides payability; otherwise the
// request would answer 201 for an order that ends cancelled.
func TestPaymentSessionWaitsForOrderLockAndThenRefusesCancelledOrder(t *testing.T) {
	fixture := newPaymentEnv(t).newFixture(t, 300)
	cancelTx := fixture.env.db.Begin()
	if _, err := order.Finish(cancelTx, fixture.orderID, order.OutcomeCancelled); err != nil {
		cancelTx.Rollback()
		t.Fatalf("cancel order: %v", err)
	}

	done := make(chan int, 1)
	go func() { done <- fixture.postSession(t, fixture.customer.parallel(fixture.env.router)) }()
	select {
	case code := <-done:
		cancelTx.Rollback()
		t.Fatalf("session request answered %d while the order lock was held", code)
	case <-time.After(400 * time.Millisecond):
	}
	if err := cancelTx.Commit().Error; err != nil {
		t.Fatalf("commit cancel: %v", err)
	}
	select {
	case code := <-done:
		if code != http.StatusConflict {
			t.Errorf("session request = %d, want 409 ORDER_NOT_PAYABLE", code)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("session request did not finish after the lock was released")
	}
	if rows := fixture.count(t, `SELECT count(*) FROM payment_sessions`); rows != 0 {
		t.Errorf("payment_sessions rows = %d, want 0 for an order that ended cancelled", rows)
	}
}
