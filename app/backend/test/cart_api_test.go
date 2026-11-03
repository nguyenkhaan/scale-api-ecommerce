package backend_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/model"
	"cloudian/cloudian-scale/internal/module/cart"
	"cloudian/cloudian-scale/internal/module/flashsale"
)

func TestCartAutoCreationAndItemLimits(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	router := newPhase3Router(t, testDatabase.ORM())
	customer := newPhase3CustomerClient(t, router, testDatabase.ORM(), "customer-cart-1@example.com")

	// 1. Initial cart is empty
	res := customer.send(t, http.MethodGet, "/api/v1/cart", nil)
	if res.Code != http.StatusOK {
		t.Fatalf("get cart status = %d; body = %s", res.Code, res.Body.String())
	}
	initialCart := decodeJSON[cart.CartResponse](t, res)
	if len(initialCart.Items) != 0 {
		t.Errorf("initial items = %d, want 0", len(initialCart.Items))
	}

	// 2. Add an item
	_, productID := seedCatalogProduct(t, testDatabase.ORM(), true, model.ProductStatusActive)
	skuID1 := seedCatalogSKU(t, testDatabase.ORM(), productID, "CART-SKU-1", model.SKUStatusActive, 50_000, 10, 100)

	putPayload := map[string]any{"quantity": 2}
	putRes := customer.send(t, http.MethodPut, "/api/v1/cart/items/"+skuID1, putPayload, customer.csrfHeader()...)
	if putRes.Code != http.StatusOK {
		t.Fatalf("put item status = %d; body = %s", putRes.Code, putRes.Body.String())
	}
	cartAfter1 := decodeJSON[cart.CartResponse](t, putRes)
	if len(cartAfter1.Items) != 1 {
		t.Fatalf("items count = %d, want 1", len(cartAfter1.Items))
	}
	if cartAfter1.Items[0].SKUID != skuID1 || cartAfter1.Items[0].Quantity != 2 || cartAfter1.Items[0].Version != 1 {
		t.Errorf("item 1 state mismatch: %+v", cartAfter1.Items[0])
	}

	// 3. Duplicate SKU upsert merges / updates quantity
	putPayload2 := map[string]any{"quantity": 5}
	putRes2 := customer.send(t, http.MethodPut, "/api/v1/cart/items/"+skuID1, putPayload2, customer.csrfHeader()...)
	if putRes2.Code != http.StatusOK {
		t.Fatalf("put duplicate item status = %d; body = %s", putRes2.Code, putRes2.Body.String())
	}
	cartAfter2 := decodeJSON[cart.CartResponse](t, putRes2)
	if len(cartAfter2.Items) != 1 {
		t.Fatalf("expected merged SKU, got items count = %d", len(cartAfter2.Items))
	}
	if cartAfter2.Items[0].Quantity != 5 || cartAfter2.Items[0].Version != 2 {
		t.Errorf("expected quantity=5, version=2; got %+v", cartAfter2.Items[0])
	}

	// 4. Test 20 distinct SKU limit (BR-16)
	// Currently 1 SKU is in the cart. Add 19 more to reach 20 distinct SKUs.
	for i := 2; i <= 20; i++ {
		skuIDI := seedCatalogSKU(t, testDatabase.ORM(), productID, fmt.Sprintf("CART-SKU-%d", i), model.SKUStatusActive, 10_000, 10, 100)
		resI := customer.send(t, http.MethodPut, "/api/v1/cart/items/"+skuIDI, map[string]any{"quantity": 1}, customer.csrfHeader()...)
		if resI.Code != http.StatusOK {
			t.Fatalf("failed to add item %d: status=%d, body=%s", i, resI.Code, resI.Body.String())
		}
	}

	// Verify cart now has 20 items
	cart20Res := customer.send(t, http.MethodGet, "/api/v1/cart", nil)
	cart20 := decodeJSON[cart.CartResponse](t, cart20Res)
	if len(cart20.Items) != 20 {
		t.Fatalf("cart items = %d, want 20", len(cart20.Items))
	}

	// 21st distinct SKU must be rejected with 422
	skuID21 := seedCatalogSKU(t, testDatabase.ORM(), productID, "CART-SKU-21", model.SKUStatusActive, 10_000, 10, 100)
	res21 := customer.send(t, http.MethodPut, "/api/v1/cart/items/"+skuID21, map[string]any{"quantity": 1}, customer.csrfHeader()...)
	requireProblem(t, res21, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
}

func TestCartOptimisticConcurrency(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	router := newPhase3Router(t, testDatabase.ORM())
	customer := newPhase3CustomerClient(t, router, testDatabase.ORM(), "customer-cart-version@example.com")

	_, productID := seedCatalogProduct(t, testDatabase.ORM(), true, model.ProductStatusActive)
	skuID := seedCatalogSKU(t, testDatabase.ORM(), productID, "CONCURRENCY-SKU", model.SKUStatusActive, 50_000, 10, 100)

	// Add item -> version 1
	addRes := customer.send(t, http.MethodPut, "/api/v1/cart/items/"+skuID, map[string]any{"quantity": 1}, customer.csrfHeader()...)
	cart1 := decodeJSON[cart.CartResponse](t, addRes)
	item := cart1.Items[0]
	if item.Version != 1 {
		t.Fatalf("initial version = %d, want 1", item.Version)
	}

	// Update with wrong expected_version -> 409 CONFLICT
	wrongVersionPayload := map[string]any{"quantity": 2, "expected_version": 999}
	failRes := customer.send(t, http.MethodPut, "/api/v1/cart/items/"+skuID, wrongVersionPayload, customer.csrfHeader()...)
	requireProblem(t, failRes, http.StatusConflict, "CONFLICT")

	// Update with correct expected_version -> 200 OK, version increments to 2
	correctVersionPayload := map[string]any{"quantity": 3, "expected_version": 1}
	okRes := customer.send(t, http.MethodPut, "/api/v1/cart/items/"+skuID, correctVersionPayload, customer.csrfHeader()...)
	if okRes.Code != http.StatusOK {
		t.Fatalf("update with version status = %d; body = %s", okRes.Code, okRes.Body.String())
	}
	cart2 := decodeJSON[cart.CartResponse](t, okRes)
	if cart2.Items[0].Quantity != 3 || cart2.Items[0].Version != 2 {
		t.Errorf("updated item state = %+v, want qty=3, ver=2", cart2.Items[0])
	}
}

func TestCartItemDeletion(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	router := newPhase3Router(t, testDatabase.ORM())
	customer := newPhase3CustomerClient(t, router, testDatabase.ORM(), "customer-cart-delete@example.com")

	_, productID := seedCatalogProduct(t, testDatabase.ORM(), true, model.ProductStatusActive)
	skuID := seedCatalogSKU(t, testDatabase.ORM(), productID, "DEL-SKU", model.SKUStatusActive, 50_000, 10, 100)

	// Add item
	customer.send(t, http.MethodPut, "/api/v1/cart/items/"+skuID, map[string]any{"quantity": 1}, customer.csrfHeader()...)

	// Delete item
	delRes := customer.send(t, http.MethodDelete, "/api/v1/cart/items/"+skuID, nil, customer.csrfHeader()...)
	if delRes.Code != http.StatusOK {
		t.Fatalf("delete item status = %d; body = %s", delRes.Code, delRes.Body.String())
	}
	cartAfterDel := decodeJSON[cart.CartResponse](t, delRes)
	if len(cartAfterDel.Items) != 0 {
		t.Errorf("cart items after delete = %d, want 0", len(cartAfterDel.Items))
	}
}

func TestCheckoutPreviewPricingAndLimits(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	router := newPhase3Router(t, testDatabase.ORM())
	admin := newCatalogAdminClient(t, router, testDatabase.ORM())
	customer := newPhase3CustomerClient(t, router, testDatabase.ORM(), "customer-checkout-preview@example.com")

	_, productID := seedCatalogProduct(t, testDatabase.ORM(), true, model.ProductStatusActive)
	// Regular price: 100,000 VND
	skuID := seedCatalogSKU(t, testDatabase.ORM(), productID, "PREVIEW-SKU", model.SKUStatusActive, 100_000, 10, 100)

	// Case 1: Preview regular item in cart -> regular price
	addRegular := customer.send(t, http.MethodPut, "/api/v1/cart/items/"+skuID, map[string]any{
		"quantity":     2,
		"price_source": "regular",
	}, customer.csrfHeader()...)
	if addRegular.Code != http.StatusOK {
		t.Fatalf("add regular status = %d; body = %s", addRegular.Code, addRegular.Body.String())
	}
	cart1 := decodeJSON[cart.CartResponse](t, addRegular)
	item1 := cart1.Items[0]

	previewReq1 := map[string]any{
		"items": []map[string]any{
			{"cart_item_id": item1.ID, "version": item1.Version},
		},
	}
	res1 := customer.send(t, http.MethodPost, "/api/v1/checkout/preview", previewReq1, customer.csrfHeader()...)
	if res1.Code != http.StatusOK {
		t.Fatalf("preview status = %d; body = %s", res1.Code, res1.Body.String())
	}
	prev1 := decodeJSON[cart.CheckoutPreviewResponse](t, res1)
	if prev1.TotalAmount != 200_000 {
		t.Errorf("prev1 total = %d, want 200,000", prev1.TotalAmount)
	}
	if len(prev1.Lines) != 1 || prev1.Lines[0].UnitPrice != 100_000 || prev1.Lines[0].PriceSource != "regular" {
		t.Errorf("prev1 lines mismatch: %+v", prev1.Lines)
	}

	// Case 2: Active Flash Sale with quota and purchase limit
	// Sale price: 40,000 VND, Quota: 10, Purchase limit: 2
	startsAt := time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339)
	endsAt := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
	fsPayload := map[string]any{
		"name":      "Active Preview Sale",
		"starts_at": startsAt,
		"ends_at":   endsAt,
		"items": []map[string]any{
			{"sku_id": skuID, "sale_price": 40_000, "quota_total": 10, "purchase_limit": 2},
		},
	}
	fsRes := admin.send(t, http.MethodPost, "/api/v1/admin/flash-sales", fsPayload, admin.csrfHeader()...)
	fs := decodeJSON[flashsale.FlashSaleResponse](t, fsRes)
	// Schedule to activate
	admin.send(t, http.MethodPost, "/api/v1/admin/flash-sales/"+fs.ID+"/schedule", nil, admin.csrfHeader()...)
	fsItemID := fs.Items[0].ID

	// Update cart item to flash sale source with quantity 2
	addFS := customer.send(t, http.MethodPut, "/api/v1/cart/items/"+skuID, map[string]any{
		"quantity":           2,
		"price_source":       "flash_sale",
		"flash_sale_item_id": fsItemID,
		"expected_version":   item1.Version,
	}, customer.csrfHeader()...)
	if addFS.Code != http.StatusOK {
		t.Fatalf("add flash sale status = %d; body = %s", addFS.Code, addFS.Body.String())
	}
	cart2 := decodeJSON[cart.CartResponse](t, addFS)
	item2 := cart2.Items[0]

	// Preview 2 units: should get flash sale price (2 * 40,000 = 80,000)
	previewReq2 := map[string]any{
		"items": []map[string]any{
			{"cart_item_id": item2.ID, "version": item2.Version},
		},
	}
	res2 := customer.send(t, http.MethodPost, "/api/v1/checkout/preview", previewReq2, customer.csrfHeader()...)
	if res2.Code != http.StatusOK {
		t.Fatalf("preview 2 status = %d; body = %s", res2.Code, res2.Body.String())
	}
	prev2 := decodeJSON[cart.CheckoutPreviewResponse](t, res2)
	if prev2.TotalAmount != 80_000 {
		t.Errorf("prev2 total = %d, want 80,000", prev2.TotalAmount)
	}
	if prev2.Lines[0].PriceSource != "flash_sale" || prev2.Lines[0].UnitPrice != 40_000 {
		t.Errorf("prev2 line mismatch: %+v", prev2.Lines[0])
	}

	// Case 3: Exceeds purchase limit of 2 (try quantity 3) -> 409 CONFLICT
	addLimitExceeded := customer.send(t, http.MethodPut, "/api/v1/cart/items/"+skuID, map[string]any{
		"quantity":           3,
		"price_source":       "flash_sale",
		"flash_sale_item_id": fsItemID,
		"expected_version":   item2.Version,
	}, customer.csrfHeader()...)
	cart3 := decodeJSON[cart.CartResponse](t, addLimitExceeded)
	item3 := cart3.Items[0]

	previewReq3 := map[string]any{
		"items": []map[string]any{
			{"cart_item_id": item3.ID, "version": item3.Version},
		},
	}
	res3 := customer.send(t, http.MethodPost, "/api/v1/checkout/preview", previewReq3, customer.csrfHeader()...)
	requireProblem(t, res3, http.StatusConflict, "CONFLICT")

	// Case 4: Verify ZERO inventory lock: on_hand and reserved in inventory table remain untouched
	var inv model.Inventory
	if err := testDatabase.ORM().Where("sku_id = ?", skuID).First(&inv).Error; err != nil {
		t.Fatalf("query inventory: %v", err)
	}
	if inv.OnHand != 100 || inv.Reserved != 0 {
		t.Errorf("inventory mutated during preview: on_hand=%d, reserved=%d (want 100, 0)", inv.OnHand, inv.Reserved)
	}
}
