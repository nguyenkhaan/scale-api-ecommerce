package backend_test

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"testing"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/module/cart"
	"cloudian/cloudian-scale/internal/module/order"
)

var orderNumberPattern = regexp.MustCompile(`^SS-\d{8}-[0-9A-Z]{8}$`)

func checkoutPlaceOrder(t *testing.T, service *order.CheckoutService, userID, key string, request order.PlaceOrderRequest) (order.PlaceOrderResult, error) {
	t.Helper()
	if violations := order.ValidatePlaceOrder(&request); len(violations) > 0 {
		t.Fatalf("valid checkout request has violations: %+v", violations)
	}
	return service.PlaceOrder(context.Background(), order.PlaceOrderInput{
		UserID: userID, IdempotencyKey: key, RequestID: "request-checkout-test", Request: request,
	})
}

func TestPlaceOrderRegularMultiSKUCreatesPendingOrderWithReservations(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "checkout-regular")
	skuA := cartSeedSKU(t, database, cartSKUOptions{OnHand: 10})
	skuB := cartSeedSKU(t, database, cartSKUOptions{OnHand: 10})
	lineA := cartAdd(t, customer.testClient, skuA, 2)
	lineB := cartAdd(t, customer.testClient, skuB, 1)
	request := checkoutRequestFromPreview(checkoutPreviewRequest(t, customer.testClient, lineA, lineB))

	result, err := checkoutPlaceOrder(t, checkoutNewService(t, database), customer.ID, "regular-order-01", request)
	if err != nil {
		t.Fatalf("place order: %v", err)
	}
	if result.Status != http.StatusCreated || result.Outcome != order.OutcomeCreated {
		t.Fatalf("place result = %+v, want created 201", result)
	}
	var created order.OrderResponse
	if err := json.Unmarshal(result.Body, &created); err != nil {
		t.Fatalf("decode created order: %v", err)
	}
	if created.Status != order.StatusPendingPayment || !created.Payable || !orderNumberPattern.MatchString(created.OrderNumber) {
		t.Errorf("created order = %+v, want payable pending order with generated order number", created)
	}
	if created.TotalAmount != request.ExpectedTotal || created.Recipient.Name != "Nguyễn Văn An" {
		t.Errorf("created snapshot = %+v", created)
	}

	var expirySeconds int
	if err := database.Raw(`SELECT extract(epoch FROM expires_at - created_at)::int FROM orders WHERE id = ?`, created.ID).
		Scan(&expirySeconds).Error; err != nil {
		t.Fatalf("read order deadline: %v", err)
	}
	if expirySeconds != 300 {
		t.Errorf("expires_at - created_at = %d seconds, want 300", expirySeconds)
	}
	var snapshots []struct {
		CartItemID      string `gorm:"column:cart_item_id"`
		CartItemVersion int64  `gorm:"column:cart_item_version"`
		SKUID           string `gorm:"column:sku_id"`
	}
	if err := database.Raw(`SELECT cart_item_id, cart_item_version, sku_id FROM order_items WHERE order_id = ? ORDER BY sku_id`, created.ID).
		Scan(&snapshots).Error; err != nil {
		t.Fatalf("read order snapshots: %v", err)
	}
	wantSnapshots := map[string]int64{lineA.ID: lineA.Version, lineB.ID: lineB.Version}
	if len(snapshots) != 2 {
		t.Fatalf("order item snapshots = %+v, want two", snapshots)
	}
	for _, snapshot := range snapshots {
		if wantSnapshots[snapshot.CartItemID] != snapshot.CartItemVersion {
			t.Errorf("snapshot = %+v, want cart item version %d", snapshot, wantSnapshots[snapshot.CartItemID])
		}
		delete(wantSnapshots, snapshot.CartItemID)
	}
	if len(wantSnapshots) != 0 {
		t.Errorf("missing cart snapshots: %v", wantSnapshots)
	}
	if count := cartCount(t, database, `SELECT count(*) FROM inventory_reservations r JOIN order_items oi ON oi.id = r.order_item_id WHERE oi.order_id = ? AND r.status = 'active'`, created.ID); count != 2 {
		t.Errorf("active reservations = %d, want 2", count)
	}
	for skuID, quantity := range map[string]int64{skuA: 2, skuB: 1} {
		var reserved int64
		if err := database.Raw(`SELECT reserved FROM inventories WHERE sku_id = ?`, skuID).Scan(&reserved).Error; err != nil {
			t.Fatalf("read reserved quantity: %v", err)
		}
		if reserved != quantity {
			t.Errorf("SKU %s reserved = %d, want %d", skuID, reserved, quantity)
		}
	}
	if count := cartCount(t, database, `SELECT count(*) FROM cart_items WHERE id IN ?`, []string{lineA.ID, lineB.ID}); count != 2 {
		t.Errorf("checkout removed cart lines; remaining = %d, want 2", count)
	}
	if count := cartCount(t, database, `SELECT count(*) FROM idempotency_records WHERE user_id = ? AND idempotency_key = 'regular-order-01' AND response_status = 201`, customer.ID); count != 1 {
		t.Errorf("successful idempotency result rows = %d, want 1", count)
	}
	lifecycleRequireInvariants(t, database)
}

func TestPlaceOrderMixedRegularAndFlashSaleReservesQuotaAndUsage(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "checkout-mixed")
	regularSKU := cartSeedSKU(t, database, cartSKUOptions{OnHand: 10})
	saleSKU := cartSeedSKU(t, database, cartSKUOptions{OnHand: 10})
	saleItem := cartSeedSale(t, database, saleSKU, cartSaleOptions{QuotaTotal: 8, PurchaseLimit: 3})
	regularLine := cartAdd(t, customer.testClient, regularSKU, 1)
	saleResponse := cartPut(t, customer.testClient, saleSKU, map[string]any{
		"quantity": 2, "price_source": "flash_sale", "flash_sale_item_id": saleItem,
	})
	if saleResponse.Code != http.StatusOK {
		t.Fatalf("add sale line status = %d; body = %s", saleResponse.Code, saleResponse.Body.String())
	}
	saleLine := cartFindItem(t, decodeJSON[cart.CartResponse](t, saleResponse), saleSKU)
	request := checkoutRequestFromPreview(checkoutPreviewRequest(t, customer.testClient, regularLine, saleLine))
	result, err := checkoutPlaceOrder(t, checkoutNewService(t, database), customer.ID, "mixed-order-01", request)
	if err != nil || result.Status != http.StatusCreated {
		t.Fatalf("place mixed order = %+v, %v", result, err)
	}
	var created order.OrderResponse
	if err := json.Unmarshal(result.Body, &created); err != nil {
		t.Fatalf("decode mixed order: %v", err)
	}
	var counters struct {
		Reserved int64 `gorm:"column:quota_reserved"`
		Usage    int64 `gorm:"column:reserved_quantity"`
	}
	if err := database.Raw(`SELECT f.quota_reserved, u.reserved_quantity FROM flash_sale_items f
		JOIN flash_sale_customer_usage u ON u.flash_sale_item_id = f.id WHERE f.id = ? AND u.user_id = ?`,
		saleItem, customer.ID).Scan(&counters).Error; err != nil {
		t.Fatalf("read flash sale counters: %v", err)
	}
	if counters.Reserved != 2 || counters.Usage != 2 {
		t.Errorf("flash counters = %+v, want quota and customer usage reserved 2", counters)
	}
	var item struct {
		PriceSource     string `gorm:"column:price_source"`
		FlashSaleItemID string `gorm:"column:flash_sale_item_id"`
		Quantity        int    `gorm:"column:quantity"`
	}
	if err := database.Raw(`SELECT price_source, flash_sale_item_id, quantity FROM order_items WHERE order_id = ? AND sku_id = ?`, created.ID, saleSKU).
		Scan(&item).Error; err != nil {
		t.Fatalf("read sale order item: %v", err)
	}
	if item.PriceSource != "flash_sale" || item.FlashSaleItemID != saleItem || item.Quantity != 2 {
		t.Errorf("sale order item = %+v, want flash snapshot", item)
	}
	lifecycleRequireInvariants(t, database)
}

func TestPlaceOrderRejectsWholeOrderWhenLaterSKUIsShort(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "checkout-short")
	regularSKU := cartSeedSKU(t, database, cartSKUOptions{OnHand: 10})
	saleSKU := cartSeedSKU(t, database, cartSKUOptions{OnHand: 2})
	saleItem := cartSeedSale(t, database, saleSKU, cartSaleOptions{QuotaTotal: 8, PurchaseLimit: 3})
	regularLine := cartAdd(t, customer.testClient, regularSKU, 1)
	saleResponse := cartPut(t, customer.testClient, saleSKU, map[string]any{
		"quantity": 2, "price_source": "flash_sale", "flash_sale_item_id": saleItem,
	})
	saleLine := cartFindItem(t, decodeJSON[cart.CartResponse](t, saleResponse), saleSKU)
	request := checkoutRequestFromPreview(checkoutPreviewRequest(t, customer.testClient, regularLine, saleLine))
	cartExec(t, database, `UPDATE inventories SET on_hand = 1 WHERE sku_id = ?`, saleSKU)
	beforeA := lifecycleReadCounters(t, database, regularSKU, "", customer.ID)
	beforeB := lifecycleReadCounters(t, database, saleSKU, saleItem, customer.ID)

	result, err := checkoutPlaceOrder(t, checkoutNewService(t, database), customer.ID, "short-order-01", request)
	if err != nil {
		t.Fatalf("place rejected order: %v", err)
	}
	if result.Status != http.StatusConflict || result.Outcome != order.OutcomeRejected {
		t.Fatalf("rejection = %+v, want 409 rejected", result)
	}
	var rejection problem.Problem
	if err := json.Unmarshal(result.Body, &rejection); err != nil {
		t.Fatalf("decode rejection: %v", err)
	}
	if rejection.Code != problem.CodeInsufficientStock || len(rejection.LineErrors) != 1 || rejection.LineErrors[0].SKUID != saleSKU {
		t.Errorf("rejection body = %+v, want insufficient stock on later sale SKU", rejection)
	}
	if count := cartCount(t, database, `SELECT count(*) FROM orders WHERE user_id = ?`, customer.ID); count != 0 {
		t.Errorf("rejected checkout created %d orders", count)
	}
	if after := lifecycleReadCounters(t, database, regularSKU, "", customer.ID); after != beforeA {
		t.Errorf("regular counters changed from %+v to %+v", beforeA, after)
	}
	if after := lifecycleReadCounters(t, database, saleSKU, saleItem, customer.ID); after != beforeB {
		t.Errorf("sale counters changed from %+v to %+v", beforeB, after)
	}
	if count := cartCount(t, database, `SELECT count(*) FROM flash_sale_customer_usage WHERE user_id = ? AND flash_sale_item_id = ?`, customer.ID, saleItem); count != 0 {
		t.Errorf("rejected checkout left %d usage rows, want none", count)
	}
	if count := cartCount(t, database, `SELECT count(*) FROM idempotency_records WHERE user_id = ? AND idempotency_key = 'short-order-01' AND response_status = 409`, customer.ID); count != 1 {
		t.Errorf("stored rejection rows = %d, want 1", count)
	}
	lifecycleRequireInvariants(t, database)
}

func TestPlaceOrderPriceChangedReturnsCurrentPriceWithoutReserving(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "checkout-price")
	skuID := cartSeedSKU(t, database, cartSKUOptions{Price: 100_000, OnHand: 10})
	line := cartAdd(t, customer.testClient, skuID, 1)
	request := checkoutRequestFromPreview(checkoutPreviewRequest(t, customer.testClient, line))
	cartExec(t, database, `UPDATE skus SET regular_price = 510000 WHERE id = ?`, skuID)

	result, err := checkoutPlaceOrder(t, checkoutNewService(t, database), customer.ID, "price-order-01", request)
	if err != nil || result.Status != http.StatusConflict {
		t.Fatalf("price change result = %+v, %v; want 409", result, err)
	}
	var rejection problem.Problem
	if err := json.Unmarshal(result.Body, &rejection); err != nil {
		t.Fatalf("decode price rejection: %v", err)
	}
	if rejection.Code != problem.CodePriceChanged || len(rejection.LineErrors) != 1 || rejection.LineErrors[0].CurrentUnitPrice == nil || *rejection.LineErrors[0].CurrentUnitPrice != 510_000 {
		t.Errorf("price rejection = %+v, want current_unit_price=510000", rejection)
	}
	if count := cartCount(t, database, `SELECT count(*) FROM inventory_reservations r JOIN order_items oi ON oi.id = r.order_item_id JOIN orders o ON o.id = oi.order_id WHERE o.user_id = ?`, customer.ID); count != 0 {
		t.Errorf("price rejection left %d reservations", count)
	}
	lifecycleRequireInvariants(t, database)
}

func TestPlaceOrderSaleQuotaExhaustedRejectsWithoutRegularFallback(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "checkout-quota")
	skuID := cartSeedSKU(t, database, cartSKUOptions{Price: 100_000, OnHand: 10})
	saleItem := cartSeedSale(t, database, skuID, cartSaleOptions{SalePrice: 40_000, QuotaTotal: 10, PurchaseLimit: 3})
	lineResponse := cartPut(t, customer.testClient, skuID, map[string]any{
		"quantity": 2, "price_source": "flash_sale", "flash_sale_item_id": saleItem,
	})
	line := cartFindItem(t, decodeJSON[cart.CartResponse](t, lineResponse), skuID)
	request := checkoutRequestFromPreview(checkoutPreviewRequest(t, customer.testClient, line))
	cartExec(t, database, `UPDATE flash_sale_items SET quota_sold = 10 WHERE id = ?`, saleItem)

	result, err := checkoutPlaceOrder(t, checkoutNewService(t, database), customer.ID, "quota-order-01", request)
	if err != nil || result.Status != http.StatusConflict {
		t.Fatalf("quota result = %+v, %v; want 409", result, err)
	}
	var rejection problem.Problem
	if err := json.Unmarshal(result.Body, &rejection); err != nil {
		t.Fatalf("decode quota rejection: %v", err)
	}
	if rejection.Code != problem.CodeSaleQuotaExceeded || len(rejection.LineErrors) != 1 || rejection.LineErrors[0].SKUID != skuID {
		t.Errorf("quota rejection = %+v, want SALE_QUOTA_EXCEEDED", rejection)
	}
	if count := cartCount(t, database, `SELECT count(*) FROM orders WHERE user_id = ?`, customer.ID); count != 0 {
		t.Errorf("quota rejection created %d orders", count)
	}
	lifecycleRequireInvariants(t, database)
}

func TestPlaceOrderStaleCartVersionReturnsCartItemChanged(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "checkout-stale")
	skuID := cartSeedSKU(t, database, cartSKUOptions{OnHand: 10})
	line := cartAdd(t, customer.testClient, skuID, 1)
	request := checkoutRequestFromPreview(checkoutPreviewRequest(t, customer.testClient, line))
	cartAdd(t, customer.testClient, skuID, 2)

	result, err := checkoutPlaceOrder(t, checkoutNewService(t, database), customer.ID, "stale-order-01", request)
	if err != nil || result.Status != http.StatusConflict {
		t.Fatalf("stale cart result = %+v, %v; want 409", result, err)
	}
	var rejection problem.Problem
	if err := json.Unmarshal(result.Body, &rejection); err != nil {
		t.Fatalf("decode stale cart rejection: %v", err)
	}
	if rejection.Code != problem.CodeCartItemChanged {
		t.Errorf("stale cart code = %q, want CART_ITEM_CHANGED", rejection.Code)
	}
	if count := cartCount(t, database, `SELECT count(*) FROM orders WHERE user_id = ?`, customer.ID); count != 0 {
		t.Errorf("stale cart rejection created %d orders", count)
	}
	lifecycleRequireInvariants(t, database)
}
