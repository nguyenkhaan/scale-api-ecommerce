package backend_test

import (
	"net/http"
	"testing"
)

func TestCartRejectsUnavailableSKUsOnAdd(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "cart-unavailable")

	cases := map[string]cartSKUOptions{
		"inactive product":  {ProductInactive: true},
		"hidden category":   {CategoryHidden: true},
		"inactive SKU":      {SKUInactive: true},
		"hidden and ending": {CategoryHidden: true, ProductInactive: true},
	}
	for name, options := range cases {
		t.Run(name, func(t *testing.T) {
			skuID := cartSeedSKU(t, database, options)
			requireProblem(t, cartPut(t, customer.testClient, skuID, map[string]any{"quantity": 1}), http.StatusConflict, "ITEM_UNAVAILABLE")
		})
	}
	if view := cartGet(t, customer.testClient); len(view.Items) != 0 {
		t.Errorf("rejected adds left lines: %+v", view.Items)
	}
}

func TestCartFlashSaleStateOnAdd(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "cart-sale-state")

	flashBody := func(itemID string, quantity int) map[string]any {
		return map[string]any{"quantity": quantity, "price_source": "flash_sale", "flash_sale_item_id": itemID}
	}
	inactive := map[string]cartSaleOptions{
		"ended campaign":   {StartsInSecs: -7200, EndsInSecs: -3600},
		"stopped campaign": {Status: "stopped"},
		"draft campaign":   {Status: "draft"},
	}
	for name, options := range inactive {
		t.Run(name, func(t *testing.T) {
			skuID := cartSeedSKU(t, database, cartSKUOptions{})
			itemID := cartSeedSale(t, database, skuID, options)
			requireProblem(t, cartPut(t, customer.testClient, skuID, flashBody(itemID, 1)), http.StatusConflict, "SALE_NOT_ACTIVE")
		})
	}

	t.Run("upcoming campaign is allowed", func(t *testing.T) {
		skuID := cartSeedSKU(t, database, cartSKUOptions{})
		itemID := cartSeedSale(t, database, skuID, cartSaleOptions{StartsInSecs: 1800, EndsInSecs: 7200})
		if response := cartPut(t, customer.testClient, skuID, flashBody(itemID, 1)); response.Code != http.StatusOK {
			t.Fatalf("status = %d; body = %s", response.Code, response.Body.String())
		}
	})

	t.Run("quantity above purchase limit", func(t *testing.T) {
		skuID := cartSeedSKU(t, database, cartSKUOptions{})
		itemID := cartSeedSale(t, database, skuID, cartSaleOptions{PurchaseLimit: 2})
		body := requireProblem(t, cartPut(t, customer.testClient, skuID, flashBody(itemID, 3)), http.StatusUnprocessableEntity, "VALIDATION_ERROR")
		if len(body.Violations) != 1 || body.Violations[0].Field != "quantity" {
			t.Errorf("violations = %+v, want quantity", body.Violations)
		}
	})
}
