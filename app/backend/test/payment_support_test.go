package backend_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/httpserver/middleware"
	"cloudian/cloudian-scale/internal/model"
	"cloudian/cloudian-scale/internal/module/auth"
	"cloudian/cloudian-scale/internal/module/cart"
	"cloudian/cloudian-scale/internal/module/order"
	"cloudian/cloudian-scale/internal/module/payment"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	paymentTestSecret      = "payment-test-secret"
	paymentTestOperatorKey = "payment-test-operator-key"
	paymentCallbackPath    = "/api/payment-callbacks/simulator"
)

var paymentSeedCounter atomic.Int64

// paymentDelivered is what the callback endpoint answered to one simulator delivery.
type paymentDelivered struct {
	Code    int
	Receipt payment.EventReceipt
}

// paymentRouterSender delivers simulator callbacks straight into the router, so the real signature
// verification runs but timing stays deterministic and observable.
type paymentRouterSender struct {
	router    http.Handler
	delivered chan paymentDelivered
}

func (sender *paymentRouterSender) Send(_ context.Context, delivery payment.Delivery) error {
	request := httptest.NewRequest(http.MethodPost, paymentCallbackPath, strings.NewReader(string(delivery.Body)))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(payment.SignatureHeader, delivery.Signature)
	response := httptest.NewRecorder()
	sender.router.ServeHTTP(response, request)
	result := paymentDelivered{Code: response.Code}
	if response.Code == http.StatusAccepted {
		result.Receipt = decodeResponseBody[payment.EventReceipt](response)
	}
	sender.delivered <- result
	return nil
}

// paymentEnv is the auth + payment router over the shared integration database.
type paymentEnv struct {
	db     *gorm.DB
	router http.Handler
	sender *paymentRouterSender
}

func newPaymentEnv(t *testing.T, mutate ...func(*payment.Config)) paymentEnv {
	t.Helper()
	database := openIntegrationDatabase(t).ORM()
	authService, err := auth.NewService(database, bcrypt.MinCost)
	if err != nil {
		t.Fatalf("create auth service: %v", err)
	}
	cfg := payment.Config{SimulatorEnabled: true, SimulatorSecret: paymentTestSecret, OperatorKey: paymentTestOperatorKey}
	for _, change := range mutate {
		change(&cfg)
	}
	sender := &paymentRouterSender{delivered: make(chan paymentDelivered, 64)}
	service, err := payment.NewService(database, cfg, sender)
	if err != nil {
		t.Fatalf("create payment service: %v", err)
	}
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middleware.RequestIDMiddleware())
	api := engine.Group("/api")
	auth.RegisterRoutes(api, auth.NewController(authService, true), middleware.RequireSession(authService))
	payment.RegisterRoutes(api, payment.NewController(service),
		middleware.RequireSession(authService), middleware.RequireRole(model.RoleAdmin))
	sender.router = engine
	return paymentEnv{db: database, router: engine, sender: sender}
}

func newCheckoutPaymentEnv(t *testing.T) paymentEnv {
	t.Helper()
	database := openIntegrationDatabase(t).ORM()
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
	checkoutService, err := order.NewCheckoutService(database, order.NewCheckoutCounters())
	if err != nil {
		t.Fatalf("create checkout service: %v", err)
	}
	cfg := payment.Config{SimulatorEnabled: true, SimulatorSecret: paymentTestSecret, OperatorKey: paymentTestOperatorKey}
	sender := &paymentRouterSender{delivered: make(chan paymentDelivered, 64)}
	service, err := payment.NewService(database, cfg, sender)
	if err != nil {
		t.Fatalf("create payment service: %v", err)
	}
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middleware.RequestIDMiddleware())
	api := engine.Group("/api")
	requireSession := middleware.RequireSession(authService)
	auth.RegisterRoutes(api, auth.NewController(authService, true), requireSession)
	cart.RegisterRoutes(api, cart.NewController(cartService), requireSession)
	order.RegisterOrderRoutes(api, order.NewOrdersController(queries), requireSession)
	order.RegisterCheckoutRoutes(api, order.NewCheckoutController(checkoutService), requireSession)
	payment.RegisterRoutes(api, payment.NewController(service), requireSession,
		middleware.RequireRole(model.RoleAdmin))
	sender.router = engine
	return paymentEnv{db: database, router: engine, sender: sender}
}

func (env paymentEnv) newAdmin(t *testing.T) *testClient {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(cartCustomerPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	email := fmt.Sprintf("payment-admin-%d@example.com", paymentSeedCounter.Add(1))
	if err := env.db.Exec(
		`INSERT INTO users (id, email, email_normalized, password_hash, role, status)
		 VALUES (gen_random_uuid(), ?, ?, ?, 'admin', 'active')`, email, email, string(hash),
	).Error; err != nil {
		t.Fatalf("create admin: %v", err)
	}
	client := newTestClient(env.router)
	if response := client.send(t, http.MethodPost, "/api/auth/login", credentials(email, cartCustomerPassword)); response.Code != http.StatusOK {
		t.Fatalf("admin login status = %d; body = %s", response.Code, response.Body.String())
	}
	return client
}

// paymentFixture is one customer with a pending order (two flash sale units and one regular unit, 180000 VND).
type paymentFixture struct {
	env                       paymentEnv
	customer                  cartCustomer
	orderID                   string
	saleSKU, regularSKU, sale string
	saleBefore, regularBefore lifecycleCounters
}

const paymentOrderTotal = int64(180_000)

func (env paymentEnv) newFixture(t *testing.T, expiresInSecs float64) paymentFixture {
	t.Helper()
	fixture := paymentFixture{env: env}
	fixture.customer = cartNewCustomer(t, env.router, env.db, "payer")
	fixture.saleSKU = cartSeedSKU(t, env.db, cartSKUOptions{OnHand: 10})
	fixture.regularSKU = cartSeedSKU(t, env.db, cartSKUOptions{OnHand: 10})
	fixture.sale = cartSeedSale(t, env.db, fixture.saleSKU, cartSaleOptions{QuotaTotal: 5, PurchaseLimit: 3})
	fixture.orderID = lifecycleSeedPendingOrder(t, env.db, fixture.customer.ID, expiresInSecs,
		lifecycleLine{SKUID: fixture.saleSKU, Quantity: 2, UnitPrice: 40_000, FlashSaleItemID: fixture.sale},
		lifecycleLine{SKUID: fixture.regularSKU, Quantity: 1, UnitPrice: 100_000},
	)
	fixture.saleBefore = lifecycleReadCounters(t, env.db, fixture.saleSKU, fixture.sale, fixture.customer.ID)
	fixture.regularBefore = lifecycleReadCounters(t, env.db, fixture.regularSKU, "", fixture.customer.ID)
	return fixture
}

func (fixture paymentFixture) saleCounters(t *testing.T) lifecycleCounters {
	t.Helper()
	return lifecycleReadCounters(t, fixture.env.db, fixture.saleSKU, fixture.sale, fixture.customer.ID)
}

// requireHeld asserts stock, quota and usage are exactly as seeded (hold still in place).
func (fixture paymentFixture) requireHeld(t *testing.T) {
	t.Helper()
	if now := fixture.saleCounters(t); now != fixture.saleBefore {
		t.Errorf("flash sale counters = %+v, want unchanged %+v", now, fixture.saleBefore)
	}
}

// requireSold asserts the order's stock and quota were spent exactly once.
func (fixture paymentFixture) requireSold(t *testing.T) {
	t.Helper()
	before := fixture.saleBefore
	want := lifecycleCounters{
		OnHand: before.OnHand - 2, Reserved: before.Reserved - 2,
		QuotaReserved: before.QuotaReserved - 2, QuotaSold: before.QuotaSold + 2,
		UsageReserved: before.UsageReserved - 2, UsageSold: before.UsageSold + 2,
	}
	if now := fixture.saleCounters(t); now != want {
		t.Errorf("flash sale counters = %+v, want sold exactly once %+v", now, want)
	}
}

// requireReleased asserts the hold was returned exactly once and nothing was sold.
func (fixture paymentFixture) requireReleased(t *testing.T) {
	t.Helper()
	want := fixture.saleBefore
	want.Reserved, want.QuotaReserved, want.UsageReserved = want.Reserved-2, want.QuotaReserved-2, want.UsageReserved-2
	if now := fixture.saleCounters(t); now != want {
		t.Errorf("flash sale counters = %+v, want released exactly once %+v", now, want)
	}
}

func (fixture paymentFixture) orderStatus(t *testing.T) string {
	t.Helper()
	var status string
	fixture.env.db.Raw(`SELECT status FROM orders WHERE id = ?`, fixture.orderID).Scan(&status)
	return status
}

func (fixture paymentFixture) createSession(t *testing.T) payment.Session {
	t.Helper()
	response := fixture.customer.send(t, http.MethodPost, "/api/orders/"+fixture.orderID+"/payment-session", nil,
		fixture.customer.csrfHeader()...)
	if response.Code != http.StatusCreated && response.Code != http.StatusOK {
		t.Fatalf("create payment session = %d; body = %s", response.Code, response.Body.String())
	}
	return decodeJSON[payment.Session](t, response)
}

func (fixture paymentFixture) count(t *testing.T, sql string, args ...any) int64 {
	t.Helper()
	return cartCount(t, fixture.env.db, sql, args...)
}

func (fixture paymentFixture) reconciliations(t *testing.T, reason string) int64 {
	t.Helper()
	return fixture.count(t, `SELECT count(*) FROM payment_reconciliations WHERE order_id = ? AND reason = ?`, fixture.orderID, reason)
}

func (fixture paymentFixture) sessionStatus(t *testing.T, sessionID string) string {
	t.Helper()
	var status string
	fixture.env.db.Raw(`SELECT status FROM payment_sessions WHERE id = ?`, sessionID).Scan(&status)
	return status
}

func paymentEventBody(sessionID, eventID, eventType string, amount int64, currency string) string {
	return fmt.Sprintf(
		`{"event_id":%q,"payment_session_id":%q,"event_type":%q,"amount":%d,"currency":%q,"occurred_at":%q}`,
		eventID, sessionID, eventType, amount, currency, time.Now().UTC().Format(time.RFC3339Nano))
}

// callback posts a raw body with an explicit signature header value.
func (env paymentEnv) callback(body, signature string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, paymentCallbackPath, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if signature != "" {
		request.Header.Set(payment.SignatureHeader, signature)
	}
	response := httptest.NewRecorder()
	env.router.ServeHTTP(response, request)
	return response
}

// deliver signs the body correctly at the current time and posts it.
func (env paymentEnv) deliver(body string) *httptest.ResponseRecorder {
	return env.callback(body, payment.Sign(paymentTestSecret, time.Now(), []byte(body)))
}

func decodeResponseBody[T any](response *httptest.ResponseRecorder) T {
	var out T
	if err := json.Unmarshal(response.Body.Bytes(), &out); err != nil {
		panic(fmt.Sprintf("decode %q: %v", response.Body.String(), err))
	}
	return out
}

// waitDelivered waits for n simulator deliveries to reach the callback endpoint.
func (env paymentEnv) waitDelivered(t *testing.T, n int, within time.Duration) []paymentDelivered {
	t.Helper()
	var got []paymentDelivered
	deadline := time.After(within)
	for len(got) < n {
		select {
		case delivery := <-env.sender.delivered:
			got = append(got, delivery)
		case <-deadline:
			t.Fatalf("received %d of %d simulator deliveries within %s", len(got), n, within)
		}
	}
	return got
}
