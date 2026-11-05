package backend_test

import (
	"net/http"
	"strings"
	"testing"

	"cloudian/cloudian-scale/internal/module/cart"
)

func TestCartAutoCreationAndItemLimits(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "cart-limits")

	if initial := cartGet(t, customer.testClient); len(initial.Items) != 0 || initial.ID == "" {
		t.Fatalf("initial cart = %+v, want empty cart with id", initial)
	}

	skuID := cartSeedSKU(t, database, cartSKUOptions{Price: 50_000})
	first := cartAdd(t, customer.testClient, skuID, 2)
	if first.Quantity != 2 || first.Version != 1 || first.PriceSource != "regular" || first.FlashSaleItemID != nil {
		t.Errorf("first item = %+v", first)
	}
	if first.ProductName == "" || first.SKUCode == "" || first.ReferenceUnitPrice != 50_000 {
		t.Errorf("item context missing: %+v", first)
	}

	merged := cartAdd(t, customer.testClient, skuID, 5)
	if merged.Quantity != 5 || merged.Version != 2 || merged.ID != first.ID {
		t.Errorf("merged item = %+v, want same line qty=5 version=2", merged)
	}

	for i := 2; i <= cart.MaxCartItems; i++ {
		cartAdd(t, customer.testClient, cartSeedSKU(t, database, cartSKUOptions{}), 1)
	}
	if got := cartGet(t, customer.testClient); len(got.Items) != cart.MaxCartItems {
		t.Fatalf("cart items = %d, want %d", len(got.Items), cart.MaxCartItems)
	}

	extra := cartSeedSKU(t, database, cartSKUOptions{})
	problemBody := requireProblem(t, cartPut(t, customer.testClient, extra, map[string]any{"quantity": 1}), http.StatusUnprocessableEntity, "VALIDATION_ERROR")
	if len(problemBody.Violations) != 1 || problemBody.Violations[0].Field != "cart" {
		t.Errorf("violations = %+v, want field cart", problemBody.Violations)
	}
	// Changing an existing line stays allowed at the limit.
	cartAdd(t, customer.testClient, skuID, 3)
}

func TestCartItemOrderIsStable(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "cart-order")

	skus := make([]string, 12)
	for i := range skus {
		skus[i] = cartSeedSKU(t, database, cartSKUOptions{})
		cartAdd(t, customer.testClient, skus[i], 1)
	}
	cartAdd(t, customer.testClient, skus[0], 4) // updating must not move the line

	view := cartGet(t, customer.testClient)
	for i, item := range view.Items {
		if item.SKUID != skus[i] {
			t.Fatalf("position %d = %s, want %s (insertion order)", i, item.SKUID, skus[i])
		}
	}
}

func TestCartUpsertValidation(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "cart-validation")

	skuID := cartSeedSKU(t, database, cartSKUOptions{MaxQuantity: 3})
	otherSKU := cartSeedSKU(t, database, cartSKUOptions{})
	saleItem := cartSeedSale(t, database, skuID, cartSaleOptions{PurchaseLimit: 2})
	otherSale := cartSeedSale(t, database, otherSKU, cartSaleOptions{})

	cases := []struct {
		name  string
		sku   string
		body  map[string]any
		field string
	}{
		{"quantity above SKU cap", skuID, map[string]any{"quantity": 4}, "quantity"},
		{"quantity below one", skuID, map[string]any{"quantity": 0}, "quantity"},
		{"regular with sale item", skuID, map[string]any{"quantity": 1, "price_source": "regular", "flash_sale_item_id": saleItem}, "flash_sale_item_id"},
		{"flash without sale item", skuID, map[string]any{"quantity": 1, "price_source": "flash_sale"}, "flash_sale_item_id"},
		{"flash quantity above purchase limit", skuID, map[string]any{"quantity": 3, "price_source": "flash_sale", "flash_sale_item_id": saleItem}, "quantity"},
		{"sale item of another SKU", skuID, map[string]any{"quantity": 1, "price_source": "flash_sale", "flash_sale_item_id": otherSale}, "flash_sale_item_id"},
		{"expected version zero", skuID, map[string]any{"quantity": 1, "expected_version": 0}, "expected_version"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := requireProblem(t, cartPut(t, customer.testClient, tc.sku, tc.body), http.StatusUnprocessableEntity, "VALIDATION_ERROR")
			if len(body.Violations) == 0 || body.Violations[0].Field != tc.field {
				t.Errorf("violations = %+v, want first field %s", body.Violations, tc.field)
			}
		})
	}

	unknown := "99999999-9999-4999-8999-999999999999"
	requireProblem(t, cartPut(t, customer.testClient, unknown, map[string]any{"quantity": 1}), http.StatusConflict, "ITEM_UNAVAILABLE")

	// Omitted price_source defaults to regular; a flash line within its limit is accepted.
	if item := cartAdd(t, customer.testClient, skuID, 3); item.PriceSource != "regular" {
		t.Errorf("default price_source = %s, want regular", item.PriceSource)
	}
	response := cartPut(t, customer.testClient, skuID, map[string]any{"quantity": 2, "price_source": "flash_sale", "flash_sale_item_id": strings.ToUpper(saleItem)})
	if response.Code != http.StatusOK {
		t.Fatalf("flash upsert status = %d; body = %s", response.Code, response.Body.String())
	}
	item := cartFindItem(t, decodeJSON[cart.CartResponse](t, response), skuID)
	if item.PriceSource != "flash_sale" || item.ReferenceUnitPrice != 40_000 || item.Version != 2 {
		t.Errorf("flash line = %+v, want flash_sale reference 40000 version 2", item)
	}
}

func TestCartOptimisticVersionOnSequentialWrites(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "cart-version")
	skuID := cartSeedSKU(t, database, cartSKUOptions{})

	item := cartAdd(t, customer.testClient, skuID, 1)
	requireProblem(t, cartPut(t, customer.testClient, skuID, map[string]any{"quantity": 2, "expected_version": 999}), http.StatusConflict, "CART_ITEM_CHANGED")

	response := cartPut(t, customer.testClient, skuID, map[string]any{"quantity": 3, "expected_version": item.Version})
	if response.Code != http.StatusOK {
		t.Fatalf("versioned update status = %d; body = %s", response.Code, response.Body.String())
	}
	if updated := cartFindItem(t, decodeJSON[cart.CartResponse](t, response), skuID); updated.Quantity != 3 || updated.Version != 2 {
		t.Errorf("updated = %+v, want qty=3 version=2", updated)
	}

	// expected_version for a SKU with no line cannot create one.
	missing := cartSeedSKU(t, database, cartSKUOptions{})
	requireProblem(t, cartPut(t, customer.testClient, missing, map[string]any{"quantity": 1, "expected_version": 1}), http.StatusConflict, "CART_ITEM_CHANGED")
}

func TestCartItemDeletion(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "cart-delete")
	skuID := cartSeedSKU(t, database, cartSKUOptions{})
	keepSKU := cartSeedSKU(t, database, cartSKUOptions{})
	cartAdd(t, customer.testClient, skuID, 1)
	cartAdd(t, customer.testClient, keepSKU, 1)

	deleted := customer.send(t, http.MethodDelete, "/api/v1/cart/items/"+skuID, nil, customer.csrfHeader()...)
	if deleted.Code != http.StatusNoContent || deleted.Body.Len() != 0 {
		t.Fatalf("delete = %d body %q, want 204 with empty body", deleted.Code, deleted.Body.String())
	}
	view := cartGet(t, customer.testClient)
	if len(view.Items) != 1 || view.Items[0].SKUID != keepSKU {
		t.Fatalf("cart after delete = %+v, want only %s", view.Items, keepSKU)
	}

	again := customer.send(t, http.MethodDelete, "/api/v1/cart/items/"+skuID, nil, customer.csrfHeader()...)
	if again.Code != http.StatusNoContent {
		t.Errorf("repeated delete status = %d, want 204", again.Code)
	}
	fresh := cartNewCustomer(t, router, database, "cart-delete-nocart")
	if res := fresh.send(t, http.MethodDelete, "/api/v1/cart/items/"+skuID, nil, fresh.csrfHeader()...); res.Code != http.StatusNoContent {
		t.Errorf("delete without a cart status = %d, want 204", res.Code)
	}
}

func TestCartMutationsRequireCSRFAndSession(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "cart-csrf")
	skuID := cartSeedSKU(t, database, cartSKUOptions{})
	item := cartAdd(t, customer.testClient, skuID, 1)
	preview := map[string]any{"items": []map[string]any{{"cart_item_id": item.ID, "version": item.Version}}}

	requireProblem(t, customer.send(t, http.MethodPut, "/api/v1/cart/items/"+skuID, map[string]any{"quantity": 2}), http.StatusForbidden, "FORBIDDEN")
	requireProblem(t, customer.send(t, http.MethodDelete, "/api/v1/cart/items/"+skuID, nil), http.StatusForbidden, "FORBIDDEN")
	requireProblem(t, customer.send(t, http.MethodPost, "/api/v1/checkout/preview", preview), http.StatusForbidden, "FORBIDDEN")
	if got := cartGet(t, customer.testClient).Items[0]; got.Quantity != 1 || got.Version != 1 {
		t.Errorf("rejected mutation changed the cart: %+v", got)
	}

	anonymous := newTestClient(router)
	requireProblem(t, anonymous.send(t, http.MethodGet, "/api/v1/cart", nil), http.StatusUnauthorized, "UNAUTHENTICATED")
}
