package backend_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/module/cart"
	"cloudian/cloudian-scale/internal/module/order"
	"cloudian/cloudian-scale/internal/module/payment"
)

type checkoutPaymentFixture struct {
	env      paymentEnv
	customer cartCustomer
	skuID    string
	saleItem string
	line     cart.CartItemResponse
	request  order.PlaceOrderRequest
}

func newCheckoutPaymentFixture(t *testing.T, stock int64, quantity int, flashSale bool) checkoutPaymentFixture {
	t.Helper()
	env := newCheckoutPaymentEnv(t)
	customer := cartNewCustomer(t, env.router, env.db, "checkout-payment")
	skuID := cartSeedSKU(t, env.db, cartSKUOptions{OnHand: stock})
	var saleItem string
	var line cart.CartItemResponse
	if flashSale {
		saleItem = cartSeedSale(t, env.db, skuID, cartSaleOptions{QuotaTotal: 10, PurchaseLimit: 4})
		response := cartPut(t, customer.testClient, skuID, map[string]any{
			"quantity": quantity, "price_source": "flash_sale", "flash_sale_item_id": saleItem,
		})
		if response.Code != http.StatusOK {
			t.Fatalf("add sale line = %d %s", response.Code, response.Body.String())
		}
		line = cartFindItem(t, decodeJSON[cart.CartResponse](t, response), skuID)
	} else {
		line = cartAdd(t, customer.testClient, skuID, quantity)
	}
	request := checkoutRequestFromPreview(checkoutPreviewRequest(t, customer.testClient, line))
	return checkoutPaymentFixture{env: env, customer: customer, skuID: skuID, saleItem: saleItem, line: line, request: request}
}

func (fixture checkoutPaymentFixture) place(t *testing.T, key string) order.OrderResponse {
	t.Helper()
	response := checkoutPost(t, fixture.customer.testClient, key, fixture.request)
	if response.Code != http.StatusCreated {
		t.Fatalf("place order status = %d; body = %s", response.Code, response.Body.String())
	}
	return decodeJSON[order.OrderResponse](t, response)
}

func (fixture checkoutPaymentFixture) createSession(t *testing.T, orderID string) payment.Session {
	t.Helper()
	response := fixture.customer.send(t, http.MethodPost, "/api/orders/"+orderID+"/payment-session", nil,
		fixture.customer.csrfHeader()...)
	if response.Code != http.StatusCreated {
		t.Fatalf("create payment session = %d; body = %s", response.Code, response.Body.String())
	}
	return decodeJSON[payment.Session](t, response)
}

func (fixture checkoutPaymentFixture) admin(t *testing.T) *testClient {
	t.Helper()
	return fixture.env.newAdmin(t)
}

func checkoutScheduleSuccess(t *testing.T, fixture checkoutPaymentFixture, session payment.Session, admin *testClient) paymentDelivered {
	t.Helper()
	response := admin.send(t, http.MethodPost, "/api/payment-simulator/events", map[string]any{
		"payment_session_id": session.ID, "event_type": "succeeded",
	}, admin.csrfHeader()...)
	if response.Code != http.StatusAccepted {
		t.Fatalf("schedule success event = %d; body = %s", response.Code, response.Body.String())
	}
	delivered := fixture.env.waitDelivered(t, 1, 3*time.Second)
	if delivered[0].Code != http.StatusAccepted || delivered[0].Receipt.Result != payment.ResultApplied {
		t.Fatalf("payment delivery = %+v, want applied callback", delivered[0])
	}
	return delivered[0]
}

func TestCheckoutRacesPaymentConfirmationOnSameCartLine(t *testing.T) {
	fixture := newCheckoutPaymentFixture(t, 10, 1, false)
	firstOrder := fixture.place(t, "payment-race-order-01")
	session := fixture.createSession(t, firstOrder.ID)
	admin := fixture.admin(t)
	responses := make([]int, 2)
	var scheduleBody string
	cartRunParallel(2, func(i int) {
		if i == 0 {
			response := admin.send(t, http.MethodPost, "/api/payment-simulator/events", map[string]any{
				"payment_session_id": session.ID, "event_type": "succeeded",
			}, admin.csrfHeader()...)
			responses[i] = response.Code
			scheduleBody = response.Body.String()
			return
		}
		responses[i] = checkoutPost(t, fixture.customer.parallel(fixture.env.router), "payment-race-order-02", fixture.request).Code
	})
	if responses[0] != http.StatusAccepted {
		t.Fatalf("simulator scheduling = %d %s", responses[0], scheduleBody)
	}
	if responses[1] != http.StatusCreated && responses[1] != http.StatusConflict && responses[1] != http.StatusUnprocessableEntity {
		t.Errorf("second checkout status = %d, want a business result without 5xx", responses[1])
	}
	delivered := fixture.env.waitDelivered(t, 1, 3*time.Second)
	if delivered[0].Code != http.StatusAccepted {
		t.Errorf("callback status = %d, want 202", delivered[0].Code)
	}
	if status := orderAPIStatus(t, fixture.env.db, firstOrder.ID); status != "confirmed" {
		t.Errorf("first order status = %s, want confirmed", status)
	}
	lifecycleRequireInvariants(t, fixture.env.db)
}

func TestCheckoutRacesExpiryWorker(t *testing.T) {
	fixture := newCheckoutPaymentFixture(t, 10, 1, false)
	firstOrder := fixture.place(t, "expiry-race-order-01")
	cartExec(t, fixture.env.db, `UPDATE orders SET expires_at = now() + interval '1 second' WHERE id = ?`, firstOrder.ID)
	time.Sleep(1100 * time.Millisecond)
	var expiryStats order.ExpiryStats
	var expiryErr error
	var checkoutResponseCode int
	cartRunParallel(2, func(i int) {
		if i == 0 {
			expiryStats, expiryErr = order.ExpireDue(context.Background(), fixture.env.db, 10)
			return
		}
		checkoutResponseCode = checkoutPost(t, fixture.customer.parallel(fixture.env.router), "expiry-race-order-02", fixture.request).Code
	})
	if expiryErr != nil || expiryStats.Errors != 0 {
		t.Fatalf("expiry stats = %+v, err = %v", expiryStats, expiryErr)
	}
	if checkoutResponseCode != http.StatusCreated && checkoutResponseCode != http.StatusConflict {
		t.Errorf("new checkout status = %d, want success or business conflict", checkoutResponseCode)
	}
	if status := orderAPIStatus(t, fixture.env.db, firstOrder.ID); status != "expired" {
		t.Errorf("old order status = %s, want expired", status)
	}
	lifecycleRequireInvariants(t, fixture.env.db)
}

func TestCheckoutOrderFlowsIntoPaymentSuccess(t *testing.T) {
	fixture := newCheckoutPaymentFixture(t, 10, 2, false)
	before := lifecycleReadCounters(t, fixture.env.db, fixture.skuID, "", fixture.customer.ID)
	created := fixture.place(t, "payment-flow-order-01")
	session := fixture.createSession(t, created.ID)
	delivery := checkoutScheduleSuccess(t, fixture, session, fixture.admin(t))
	if status := orderAPIStatus(t, fixture.env.db, created.ID); status != "confirmed" {
		t.Errorf("order status = %s, want confirmed", status)
	}
	after := lifecycleReadCounters(t, fixture.env.db, fixture.skuID, "", fixture.customer.ID)
	if after.OnHand != before.OnHand-2 || after.Reserved != before.Reserved {
		t.Errorf("inventory after success = %+v, before = %+v", after, before)
	}
	if rows := cartCount(t, fixture.env.db, `SELECT count(*) FROM inventory_reservations r JOIN order_items oi ON oi.id = r.order_item_id WHERE oi.order_id = ? AND r.status = 'finalized'`, created.ID); rows != 1 {
		t.Errorf("finalized reservations = %d, want 1", rows)
	}
	if rows := cartCount(t, fixture.env.db, `SELECT count(*) FROM cart_items WHERE id = ?`, fixture.line.ID); rows != 0 {
		t.Errorf("confirmed cart line remains: rows=%d", rows)
	}
	if delivery.Receipt.OrderStatus != "confirmed" {
		t.Errorf("payment receipt = %+v, want confirmed order", delivery.Receipt)
	}
	lifecycleRequireInvariants(t, fixture.env.db)
}

func TestPaymentSuccessKeepsCartLineEditedAfterCheckout(t *testing.T) {
	fixture := newCheckoutPaymentFixture(t, 10, 1, false)
	created := fixture.place(t, "edited-cart-order-01")
	edit := cartPut(t, fixture.customer.testClient, fixture.skuID, map[string]any{"quantity": 3})
	if edit.Code != http.StatusOK {
		t.Fatalf("edit cart line = %d %s", edit.Code, edit.Body.String())
	}
	var edited cart.CartItemResponse
	for _, item := range decodeJSON[cart.CartResponse](t, edit).Items {
		if item.ID == fixture.line.ID {
			edited = item
		}
	}
	if edited.ID == "" || edited.Version <= fixture.line.Version || edited.Quantity != 3 {
		t.Fatalf("edited cart line = %+v, want versioned quantity 3", edited)
	}
	session := fixture.createSession(t, created.ID)
	checkoutScheduleSuccess(t, fixture, session, fixture.admin(t))
	var remaining struct {
		Quantity int   `gorm:"column:quantity"`
		Version  int64 `gorm:"column:version"`
	}
	if err := fixture.env.db.Raw(`SELECT quantity, version FROM cart_items WHERE id = ?`, fixture.line.ID).Scan(&remaining).Error; err != nil {
		t.Fatalf("read edited cart line: %v", err)
	}
	if remaining.Quantity != 3 || remaining.Version != edited.Version {
		t.Errorf("remaining cart line = %+v, want edited version %+v", remaining, edited)
	}
	var snapshot struct {
		Quantity        int   `gorm:"column:quantity"`
		CartItemVersion int64 `gorm:"column:cart_item_version"`
	}
	if err := fixture.env.db.Raw(`SELECT quantity, cart_item_version FROM order_items WHERE order_id = ?`, created.ID).Scan(&snapshot).Error; err != nil {
		t.Fatalf("read order snapshot: %v", err)
	}
	if snapshot.Quantity != 1 || snapshot.CartItemVersion != fixture.line.Version {
		t.Errorf("order snapshot = %+v, want original qty/version", snapshot)
	}
	lifecycleRequireInvariants(t, fixture.env.db)
}

func TestCheckoutOrderCancelReleasesEverything(t *testing.T) {
	fixture := newCheckoutPaymentFixture(t, 10, 2, true)
	before := lifecycleReadCounters(t, fixture.env.db, fixture.skuID, fixture.saleItem, fixture.customer.ID)
	created := fixture.place(t, "cancel-checkout-order-01")
	response := fixture.customer.send(t, http.MethodPost, "/api/orders/"+created.ID+"/cancel", nil, fixture.customer.csrfHeader()...)
	if response.Code != http.StatusOK {
		t.Fatalf("cancel checkout order = %d %s", response.Code, response.Body.String())
	}
	after := lifecycleReadCounters(t, fixture.env.db, fixture.skuID, fixture.saleItem, fixture.customer.ID)
	if after != before {
		t.Errorf("counters after cancellation = %+v, want original %+v", after, before)
	}
	if rows := cartCount(t, fixture.env.db, `SELECT count(*) FROM inventory_reservations r JOIN order_items oi ON oi.id = r.order_item_id WHERE oi.order_id = ? AND r.status = 'released'`, created.ID); rows != 1 {
		t.Errorf("released reservations = %d, want 1", rows)
	}
	lifecycleRequireInvariants(t, fixture.env.db)
}
