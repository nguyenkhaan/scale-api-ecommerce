package backend_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/module/cart"
	"cloudian/cloudian-scale/internal/module/order"
	"gorm.io/gorm"
)

type checkoutIdempotencyFixture struct {
	db      *gorm.DB
	router  http.Handler
	service *order.CheckoutService
	userID  string
	skuID   string
	line    cart.CartItemResponse
	request order.PlaceOrderRequest
}

func newCheckoutIdempotencyFixture(t *testing.T, stock int64, quantity int) checkoutIdempotencyFixture {
	t.Helper()
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "checkout-idempotency")
	skuID := cartSeedSKU(t, database, cartSKUOptions{OnHand: stock})
	line := cartAdd(t, customer.testClient, skuID, quantity)
	request := checkoutRequestFromPreview(checkoutPreviewRequest(t, customer.testClient, line))
	return checkoutIdempotencyFixture{
		db: database, router: router, service: checkoutNewService(t, database),
		userID: customer.ID, skuID: skuID, line: line, request: request,
	}
}

func TestPlaceOrderSameKeySamePayloadReplaysWithoutSecondOrder(t *testing.T) {
	fixture := newCheckoutIdempotencyFixture(t, 10, 1)
	first, err := checkoutPlaceOrder(t, fixture.service, fixture.userID, "same-key-order-01", fixture.request)
	if err != nil || first.Status != http.StatusCreated {
		t.Fatalf("first place order = %+v, %v", first, err)
	}
	var firstOrder order.OrderResponse
	if err := json.Unmarshal(first.Body, &firstOrder); err != nil {
		t.Fatalf("decode first order: %v", err)
	}
	beforeReplay := lifecycleReadCounters(t, fixture.db, fixture.skuID, "", fixture.userID)
	second, err := checkoutPlaceOrder(t, fixture.service, fixture.userID, "same-key-order-01", fixture.request)
	if err != nil || !second.Replayed || second.Outcome != order.OutcomeReplayed || second.Status != http.StatusCreated {
		t.Fatalf("second place order = %+v, %v; want replayed 201", second, err)
	}
	if !bytes.Equal(second.Body, first.Body) {
		t.Errorf("replayed body differs from original:\nfirst:  %s\nsecond: %s", first.Body, second.Body)
	}
	var storedExpiry time.Time
	if err := fixture.db.Raw(`SELECT expires_at FROM orders WHERE id = ?`, firstOrder.ID).Scan(&storedExpiry).Error; err != nil {
		t.Fatalf("read stored expiry: %v", err)
	}
	if !storedExpiry.Equal(firstOrder.ExpiresAt) {
		t.Errorf("replay changed expiry from %s to %s", firstOrder.ExpiresAt, storedExpiry)
	}
	if count := cartCount(t, fixture.db, `SELECT count(*) FROM orders WHERE user_id = ?`, fixture.userID); count != 1 {
		t.Errorf("same-key orders = %d, want 1", count)
	}
	if after := lifecycleReadCounters(t, fixture.db, fixture.skuID, "", fixture.userID); after != beforeReplay {
		t.Errorf("replay changed counters from %+v to %+v", beforeReplay, after)
	}
	lifecycleRequireInvariants(t, fixture.db)
}

func TestPlaceOrderSameKeyDifferentPayloadConflicts(t *testing.T) {
	fixture := newCheckoutIdempotencyFixture(t, 10, 1)
	first, err := checkoutPlaceOrder(t, fixture.service, fixture.userID, "different-key-order-01", fixture.request)
	if err != nil || first.Status != http.StatusCreated {
		t.Fatalf("first place order = %+v, %v", first, err)
	}
	changed := fixture.request
	changed.Recipient.Phone = "0907654321"
	if _, err := checkoutPlaceOrder(t, fixture.service, fixture.userID, "different-key-order-01", changed); !errors.Is(err, order.ErrIdempotencyConflict) {
		t.Fatalf("different payload error = %v, want ErrIdempotencyConflict", err)
	}
	if count := cartCount(t, fixture.db, `SELECT count(*) FROM orders WHERE user_id = ?`, fixture.userID); count != 1 {
		t.Errorf("conflicting payload created another order; count = %d", count)
	}
	lifecycleRequireInvariants(t, fixture.db)
}

func TestPlaceOrderReplaysStoredRejectionAfterStockReturns(t *testing.T) {
	fixture := newCheckoutIdempotencyFixture(t, 2, 2)
	cartExec(t, fixture.db, `UPDATE inventories SET on_hand = 1 WHERE sku_id = ?`, fixture.skuID)
	first, err := checkoutPlaceOrder(t, fixture.service, fixture.userID, "rejected-key-order-01", fixture.request)
	if err != nil || first.Status != http.StatusConflict {
		t.Fatalf("first rejected place order = %+v, %v", first, err)
	}
	var firstProblem problem.Problem
	if err := json.Unmarshal(first.Body, &firstProblem); err != nil || firstProblem.Code != problem.CodeInsufficientStock {
		t.Fatalf("first rejection = %+v, decode error %v; want INSUFFICIENT_STOCK", firstProblem, err)
	}
	cartExec(t, fixture.db, `UPDATE inventories SET on_hand = 2 WHERE sku_id = ?`, fixture.skuID)
	second, err := checkoutPlaceOrder(t, fixture.service, fixture.userID, "rejected-key-order-01", fixture.request)
	if err != nil || !second.Replayed || second.Status != http.StatusConflict || !bytes.Equal(second.Body, first.Body) {
		t.Fatalf("second rejection = %+v, %v; want byte-identical stored 409", second, err)
	}
	if count := cartCount(t, fixture.db, `SELECT count(*) FROM orders WHERE user_id = ?`, fixture.userID); count != 0 {
		t.Errorf("replayed rejection created %d orders", count)
	}
	lifecycleRequireInvariants(t, fixture.db)
}

func TestPlaceOrderInfrastructureFailureLeavesKeyRetryable(t *testing.T) {
	fixture := newCheckoutIdempotencyFixture(t, 10, 1)
	request := fixture.request
	if violations := order.ValidatePlaceOrder(&request); len(violations) > 0 {
		t.Fatalf("valid request has violations: %+v", violations)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	input := order.PlaceOrderInput{
		UserID: fixture.userID, IdempotencyKey: "retry-after-error-01", RequestID: "cancelled-request", Request: request,
	}
	if _, err := fixture.service.PlaceOrder(ctx, input); err == nil {
		t.Fatal("cancelled request unexpectedly succeeded")
	}
	if count := cartCount(t, fixture.db, `SELECT count(*) FROM idempotency_records WHERE user_id = ? AND idempotency_key = ?`, fixture.userID, input.IdempotencyKey); count != 0 {
		t.Errorf("infrastructure failure left %d idempotency rows, want none", count)
	}
	result, err := checkoutPlaceOrder(t, fixture.service, fixture.userID, input.IdempotencyKey, request)
	if err != nil || result.Status != http.StatusCreated {
		t.Fatalf("retry after infrastructure failure = %+v, %v; want 201", result, err)
	}
	lifecycleRequireInvariants(t, fixture.db)
}

func TestPlaceOrderSameKeyWhileFirstHoldsLockReturnsInProgress(t *testing.T) {
	fixture := newCheckoutIdempotencyFixture(t, 10, 1)
	if violations := order.ValidatePlaceOrder(&fixture.request); len(violations) > 0 {
		t.Fatalf("valid request has violations: %+v", violations)
	}
	key := "locked-key-order-01"
	hash := order.CanonicalRequestHash(fixture.request)
	holder := fixture.db.Begin()
	if holder.Error != nil {
		t.Fatalf("begin idempotency holder: %v", holder.Error)
	}
	if err := holder.Exec(`INSERT INTO idempotency_records (user_id, operation, idempotency_key, request_hash)
		VALUES (?, 'place_order', ?, ?)`, fixture.userID, key, hash[:]).Error; err != nil {
		_ = holder.Rollback().Error
		t.Fatalf("insert held idempotency row: %v", err)
	}
	started := time.Now()
	_, err := checkoutPlaceOrder(t, fixture.service, fixture.userID, key, fixture.request)
	elapsed := time.Since(started)
	if !errors.Is(err, order.ErrIdempotencyInProgress) {
		t.Errorf("same-key locked error = %v, want ErrIdempotencyInProgress", err)
	}
	if elapsed >= 5*time.Second {
		t.Errorf("same-key lock wait took %s, want less than 5 seconds", elapsed)
	}
	if err := holder.Rollback().Error; err != nil {
		t.Fatalf("release held idempotency row: %v", err)
	}
	result, err := checkoutPlaceOrder(t, fixture.service, fixture.userID, key, fixture.request)
	if err != nil || result.Status != http.StatusCreated {
		t.Fatalf("retry after holder rollback = %+v, %v; want 201", result, err)
	}
	lifecycleRequireInvariants(t, fixture.db)
}

func TestLookupAttemptAfterLostResponse(t *testing.T) {
	fixture := newCheckoutIdempotencyFixture(t, 10, 1)
	const key = "lost-response-order-01"
	created, err := checkoutPlaceOrder(t, fixture.service, fixture.userID, key, fixture.request)
	if err != nil || created.Status != http.StatusCreated {
		t.Fatalf("place order = %+v, %v", created, err)
	}
	var expected order.OrderResponse
	if err := json.Unmarshal(created.Body, &expected); err != nil {
		t.Fatalf("decode created order: %v", err)
	}
	attempt, err := fixture.service.LookupAttempt(context.Background(), fixture.userID, key)
	if err != nil {
		t.Fatalf("lookup attempt: %v", err)
	}
	var lookedUp order.OrderResponse
	if err := json.Unmarshal(attempt.ResponseBody, &lookedUp); err != nil {
		t.Fatalf("decode lookup response: %v", err)
	}
	if attempt.ResponseStatus != http.StatusCreated || lookedUp.ID != expected.ID {
		t.Errorf("attempt = %+v; order id = %s, want status 201 and id %s", attempt, lookedUp.ID, expected.ID)
	}
	replayed, err := checkoutPlaceOrder(t, fixture.service, fixture.userID, key, fixture.request)
	if err != nil || !replayed.Replayed || !bytes.Equal(replayed.Body, created.Body) {
		t.Fatalf("replay after lookup = %+v, %v; want same stored response", replayed, err)
	}
	lifecycleRequireInvariants(t, fixture.db)
}

func TestLookupAttemptIsOwnerOnly(t *testing.T) {
	fixture := newCheckoutIdempotencyFixture(t, 10, 1)
	const key = "owner-only-order-01"
	if _, err := checkoutPlaceOrder(t, fixture.service, fixture.userID, key, fixture.request); err != nil {
		t.Fatalf("place order: %v", err)
	}
	other := cartNewCustomer(t, fixture.router, fixture.db, "checkout-idempotency-stranger")
	if _, err := fixture.service.LookupAttempt(context.Background(), other.ID, key); !errors.Is(err, order.ErrAttemptNotFound) {
		t.Errorf("other user's lookup error = %v, want ErrAttemptNotFound", err)
	}
	if _, err := fixture.service.LookupAttempt(context.Background(), fixture.userID, "never-used-order-01"); !errors.Is(err, order.ErrAttemptNotFound) {
		t.Errorf("unknown key lookup error = %v, want ErrAttemptNotFound", err)
	}
	lifecycleRequireInvariants(t, fixture.db)
}
