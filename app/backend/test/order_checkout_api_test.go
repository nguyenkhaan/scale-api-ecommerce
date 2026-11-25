package backend_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cloudian/cloudian-scale/internal/httpserver"
	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/module/auth"
	"cloudian/cloudian-scale/internal/module/cart"
	"cloudian/cloudian-scale/internal/module/order"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type checkoutAPIFixture struct {
	db       *gorm.DB
	router   http.Handler
	customer cartCustomer
	skuID    string
	line     cart.CartItemResponse
	request  order.PlaceOrderRequest
}

func newCheckoutAPIRouter(t *testing.T, database *gorm.DB) http.Handler {
	t.Helper()
	gin.SetMode(gin.TestMode)
	authService, err := auth.NewService(database, bcrypt.MinCost)
	if err != nil {
		t.Fatalf("create auth service: %v", err)
	}
	cartService, err := cart.NewService(database)
	if err != nil {
		t.Fatalf("create cart service: %v", err)
	}
	queries, err := order.NewOrderQueries(database)
	if err != nil {
		t.Fatalf("create order queries: %v", err)
	}
	checkoutService := checkoutNewService(t, database)
	return httpserver.NewRouter(httpserver.Dependencies{
		Database: databasePingerStub{}, AuthService: authService, CartService: cartService,
		OrderQueries: queries, CheckoutService: checkoutService, SecureCookies: true,
	})
}

func newCheckoutAPIFixture(t *testing.T, stock int64, quantity int) checkoutAPIFixture {
	t.Helper()
	database := openIntegrationDatabase(t).ORM()
	router := newCheckoutAPIRouter(t, database)
	customer := cartNewCustomer(t, router, database, "checkout-api")
	skuID := cartSeedSKU(t, database, cartSKUOptions{OnHand: stock})
	line := cartAdd(t, customer.testClient, skuID, quantity)
	request := checkoutRequestFromPreview(checkoutPreviewRequest(t, customer.testClient, line))
	return checkoutAPIFixture{db: database, router: router, customer: customer, skuID: skuID, line: line, request: request}
}

func checkoutPost(t *testing.T, client *testClient, key string, request any, extraHeaders ...string) *httptest.ResponseRecorder {
	t.Helper()
	headers := append(client.csrfHeader(), "Idempotency-Key", key)
	headers = append(headers, extraHeaders...)
	return client.send(t, http.MethodPost, "/api/orders", request, headers...)
}

func TestPlaceOrderHTTPCreatesOrderAndReplays(t *testing.T) {
	fixture := newCheckoutAPIFixture(t, 10, 1)
	first := checkoutPost(t, fixture.customer.testClient, "http-order-key-01", fixture.request)
	if first.Code != http.StatusCreated || first.Header().Get("Idempotency-Replayed") != "false" {
		t.Fatalf("first checkout status/header = %d/%q; body = %s", first.Code, first.Header().Get("Idempotency-Replayed"), first.Body.String())
	}
	var created order.OrderResponse
	if err := json.Unmarshal(first.Body.Bytes(), &created); err != nil || created.ID == "" {
		t.Fatalf("first body = %s; decode error = %v", first.Body.String(), err)
	}
	second := checkoutPost(t, fixture.customer.testClient, "http-order-key-01", fixture.request)
	if second.Code != http.StatusCreated || second.Header().Get("Idempotency-Replayed") != "true" {
		t.Fatalf("replay status/header = %d/%q; body = %s", second.Code, second.Header().Get("Idempotency-Replayed"), second.Body.String())
	}
	if !bytes.Equal(first.Body.Bytes(), second.Body.Bytes()) {
		t.Errorf("replay body differs:\nfirst:  %s\nsecond: %s", first.Body.String(), second.Body.String())
	}
	lifecycleRequireInvariants(t, fixture.db)
}

func TestPlaceOrderHTTPRejectsInvalidInput(t *testing.T) {
	fixture := newCheckoutAPIFixture(t, 10, 1)
	missingHeader := fixture.customer.send(t, http.MethodPost, "/api/orders", fixture.request, fixture.customer.csrfHeader()...)
	missingProblem := requireProblem(t, missingHeader, http.StatusUnprocessableEntity, problem.CodeValidation)
	if len(missingProblem.Violations) != 1 || missingProblem.Violations[0].Field != "Idempotency-Key" {
		t.Errorf("missing key violations = %+v", missingProblem.Violations)
	}
	unknownField := map[string]any{
		"lines": fixture.request.Lines, "currency": fixture.request.Currency,
		"expected_total": fixture.request.ExpectedTotal, "recipient": fixture.request.Recipient, "total": 1,
	}
	if response := checkoutPost(t, fixture.customer.testClient, "unknown-field-key-01", unknownField); response.Code != http.StatusUnprocessableEntity {
		t.Errorf("unknown field status = %d; body = %s", response.Code, response.Body.String())
	}
	if response := checkoutPost(t, fixture.customer.testClient, "malformed-json-key-01", "{"); response.Code != http.StatusUnprocessableEntity {
		t.Errorf("malformed JSON status = %d; body = %s", response.Code, response.Body.String())
	}
	unauthenticated := newTestClient(fixture.router).send(t, http.MethodPost, "/api/orders", fixture.request,
		"Idempotency-Key", "unauthenticated-key-01")
	if unauthenticated.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated status = %d; body = %s", unauthenticated.Code, unauthenticated.Body.String())
	}
	missingCSRF := fixture.customer.send(t, http.MethodPost, "/api/orders", fixture.request, "Idempotency-Key", "missing-csrf-key-01")
	if missingCSRF.Code != http.StatusForbidden {
		t.Errorf("missing CSRF status = %d; body = %s", missingCSRF.Code, missingCSRF.Body.String())
	}
	lifecycleRequireInvariants(t, fixture.db)
}

func TestPlaceOrderHTTPStoredRejectionUsesProblemContentType(t *testing.T) {
	fixture := newCheckoutAPIFixture(t, 2, 2)
	cartExec(t, fixture.db, `UPDATE inventories SET on_hand = 1 WHERE sku_id = ?`, fixture.skuID)
	response := checkoutPost(t, fixture.customer.testClient, "http-rejection-key-01", fixture.request)
	if response.Code != http.StatusConflict || !strings.HasPrefix(response.Header().Get("Content-Type"), problem.ContentType) {
		t.Fatalf("stored rejection status/content-type = %d/%q; body = %s", response.Code, response.Header().Get("Content-Type"), response.Body.String())
	}
	body := requireProblem(t, response, http.StatusConflict, problem.CodeInsufficientStock)
	if len(body.LineErrors) != 1 || body.LineErrors[0].SKUID != fixture.skuID {
		t.Errorf("line errors = %+v, want stock rejection for %s", body.LineErrors, fixture.skuID)
	}
	lifecycleRequireInvariants(t, fixture.db)
}

func TestPlaceOrderHTTPConflictAndInProgress(t *testing.T) {
	fixture := newCheckoutAPIFixture(t, 10, 1)
	const key = "http-conflict-key-01"
	if response := checkoutPost(t, fixture.customer.testClient, key, fixture.request); response.Code != http.StatusCreated {
		t.Fatalf("initial request status = %d; body = %s", response.Code, response.Body.String())
	}
	changed := fixture.request
	changed.Recipient.Phone = "0907654321"
	conflict := checkoutPost(t, fixture.customer.testClient, key, changed)
	requireProblem(t, conflict, http.StatusConflict, problem.CodeIdempotencyConflict)

	inProgressKey := "http-progress-key-01"
	hash := order.CanonicalRequestHash(fixture.request)
	holder := fixture.db.Begin()
	if holder.Error != nil {
		t.Fatalf("begin idempotency holder: %v", holder.Error)
	}
	if err := holder.Exec(`INSERT INTO idempotency_records (user_id, operation, idempotency_key, request_hash)
		VALUES (?, 'place_order', ?, ?)`, fixture.customer.ID, inProgressKey, hash[:]).Error; err != nil {
		_ = holder.Rollback().Error
		t.Fatalf("insert held key: %v", err)
	}
	response := checkoutPost(t, fixture.customer.testClient, inProgressKey, fixture.request)
	if err := holder.Rollback().Error; err != nil {
		t.Fatalf("release held key: %v", err)
	}
	requireProblem(t, response, http.StatusConflict, problem.CodeIdempotencyInProgress)
	if response.Header().Get("Retry-After") != "1" {
		t.Errorf("Retry-After = %q, want 1", response.Header().Get("Retry-After"))
	}
	lifecycleRequireInvariants(t, fixture.db)
}

func TestPlaceOrderSameKeyAcrossTwoRoutersCreatesOneOrder(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	firstRouter := newCheckoutAPIRouter(t, database)
	secondRouter := newCheckoutAPIRouter(t, database)
	customer := cartNewCustomer(t, firstRouter, database, "checkout-two-routers")
	skuID := cartSeedSKU(t, database, cartSKUOptions{OnHand: 10})
	line := cartAdd(t, customer.testClient, skuID, 1)
	request := checkoutRequestFromPreview(checkoutPreviewRequest(t, customer.testClient, line))
	clients := []*testClient{customer.testClient, customer.parallel(secondRouter)}
	responses := make([]*httptest.ResponseRecorder, len(clients))
	const key = "two-router-key-01"
	cartRunParallel(len(clients), func(i int) {
		responses[i] = checkoutPost(t, clients[i], key, request)
	})
	for i, response := range responses {
		if response == nil {
			t.Fatalf("router %d produced no response", i)
		}
		if response.Code == http.StatusCreated {
			if replayed := response.Header().Get("Idempotency-Replayed"); replayed != "true" && replayed != "false" {
				t.Errorf("router %d replay header = %q", i, replayed)
			}
			continue
		}
		body := requireProblem(t, response, http.StatusConflict, problem.CodeIdempotencyInProgress)
		_ = body
	}
	if count := cartCount(t, database, `SELECT count(*) FROM orders WHERE user_id = ?`, customer.ID); count != 1 {
		t.Errorf("two-router orders = %d, want 1", count)
	}
	if count := cartCount(t, database, `SELECT count(*) FROM idempotency_records WHERE user_id = ? AND idempotency_key = ?`, customer.ID, key); count != 1 {
		t.Errorf("two-router idempotency rows = %d, want 1", count)
	}
	lifecycleRequireInvariants(t, database)
}

func TestGetOrderAttemptHTTP(t *testing.T) {
	fixture := newCheckoutAPIFixture(t, 10, 1)
	const key = "http-lookup-key-01"
	created := checkoutPost(t, fixture.customer.testClient, key, fixture.request)
	if created.Code != http.StatusCreated {
		t.Fatalf("place order status = %d; body = %s", created.Code, created.Body.String())
	}
	response := fixture.customer.send(t, http.MethodGet, "/api/idempotency/orders/"+key, nil)
	if response.Code != http.StatusOK {
		t.Fatalf("lookup status = %d; body = %s", response.Code, response.Body.String())
	}
	attempt := decodeJSON[order.IdempotencyResult](t, response)
	var lookedUp order.OrderResponse
	if err := json.Unmarshal(attempt.ResponseBody, &lookedUp); err != nil {
		t.Fatalf("decode attempt response body: %v", err)
	}
	var createdOrder order.OrderResponse
	if err := json.Unmarshal(created.Body.Bytes(), &createdOrder); err != nil {
		t.Fatalf("decode created body: %v", err)
	}
	if attempt.IdempotencyKey != key || attempt.ResponseStatus != http.StatusCreated || lookedUp.ID != createdOrder.ID || attempt.CompletedAt.IsZero() {
		t.Errorf("lookup attempt = %+v; looked up order = %+v", attempt, lookedUp)
	}
	other := cartNewCustomer(t, fixture.router, fixture.db, "checkout-api-stranger")
	missing := other.send(t, http.MethodGet, "/api/idempotency/orders/"+key, nil)
	requireProblem(t, missing, http.StatusNotFound, problem.CodeResourceNotFound)
	lifecycleRequireInvariants(t, fixture.db)
}
