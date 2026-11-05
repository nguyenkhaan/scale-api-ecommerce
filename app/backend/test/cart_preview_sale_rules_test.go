package backend_test

import (
	"net/http"
	"testing"

	"cloudian/cloudian-scale/internal/module/cart"
	"gorm.io/gorm"
)

// cartFlashLine puts a flash sale line into the customer's cart and returns it.
func cartFlashLine(t *testing.T, client *testClient, skuID, saleItemID string, quantity int) cart.CartItemResponse {
	t.Helper()
	response := cartPut(t, client, skuID, map[string]any{"quantity": quantity, "price_source": "flash_sale", "flash_sale_item_id": saleItemID})
	if response.Code != http.StatusOK {
		t.Fatalf("flash add status = %d; body = %s", response.Code, response.Body.String())
	}
	return cartFindItem(t, decodeJSON[cart.CartResponse](t, response), skuID)
}

func cartExpectLineError(t *testing.T, client *testClient, item cart.CartItemResponse, wantCode string, wantAvailable *int64) {
	t.Helper()
	body := requireProblem(t, cartPreview(t, client, item), http.StatusConflict, wantCode)
	if len(body.LineErrors) != 1 {
		t.Fatalf("line_errors = %+v, want exactly one", body.LineErrors)
	}
	line := body.LineErrors[0]
	if line.SKUID != item.SKUID || line.Code != wantCode || line.RequestedQuantity != item.Quantity {
		t.Errorf("line = %+v, want %s for sku %s requested %d", line, wantCode, item.SKUID, item.Quantity)
	}
	switch {
	case wantAvailable == nil && line.AvailableQuantity != nil:
		t.Errorf("available_quantity = %d, want absent", *line.AvailableQuantity)
	case wantAvailable != nil && (line.AvailableQuantity == nil || *line.AvailableQuantity != *wantAvailable):
		t.Errorf("available_quantity = %v, want %d", line.AvailableQuantity, *wantAvailable)
	}
}

func cartInt(v int64) *int64 { return &v }

func cartSetup(t *testing.T, name string, sale cartSaleOptions, onHand int64) (*gorm.DB, cartCustomer, string, string) {
	t.Helper()
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, name)
	skuID := cartSeedSKU(t, database, cartSKUOptions{OnHand: onHand})
	return database, customer, skuID, cartSeedSale(t, database, skuID, sale)
}

func TestCheckoutPreviewSaleQuotaExhaustedWhileStockAvailable(t *testing.T) {
	_, customer, skuID, saleItem := cartSetup(t, "preview-quota", cartSaleOptions{QuotaTotal: 5, QuotaSold: 4, PurchaseLimit: 5}, 100)
	line := cartFlashLine(t, customer.testClient, skuID, saleItem, 2)

	cartExpectLineError(t, customer.testClient, line, "SALE_QUOTA_EXCEEDED", cartInt(1))
}

func TestCheckoutPreviewCustomerPurchaseLimitCountsReservedAndSold(t *testing.T) {
	database, customer, skuID, saleItem := cartSetup(t, "preview-limit", cartSaleOptions{QuotaTotal: 50, PurchaseLimit: 5}, 100)
	line := cartFlashLine(t, customer.testClient, skuID, saleItem, 2)
	cartExec(t, database, `INSERT INTO flash_sale_customer_usage (flash_sale_item_id, user_id, reserved_quantity, sold_quantity) VALUES (?, ?, 2, 2)`, saleItem, customer.ID)

	cartExpectLineError(t, customer.testClient, line, "PURCHASE_LIMIT_EXCEEDED", cartInt(1))

	cartExec(t, database, `UPDATE flash_sale_customer_usage SET reserved_quantity = 3, sold_quantity = 2`)
	cartExpectLineError(t, customer.testClient, line, "PURCHASE_LIMIT_EXCEEDED", cartInt(0))
}

func TestCheckoutPreviewOtherCustomersUsageDoesNotCount(t *testing.T) {
	database, customer, skuID, saleItem := cartSetup(t, "preview-usage-owner", cartSaleOptions{QuotaTotal: 50, PurchaseLimit: 2}, 100)
	other := cartNewCustomer(t, cartNewRouter(t, database), database, "preview-usage-other")
	line := cartFlashLine(t, customer.testClient, skuID, saleItem, 2)
	cartExec(t, database, `INSERT INTO flash_sale_customer_usage (flash_sale_item_id, user_id, sold_quantity) VALUES (?, ?, 2)`, saleItem, other.ID)

	if res := cartPreview(t, customer.testClient, line); res.Code != http.StatusOK {
		t.Fatalf("status = %d; body = %s", res.Code, res.Body.String())
	}
}

func TestCheckoutPreviewSaleEndedOrStoppedNeverFallsBackToRegularPrice(t *testing.T) {
	mutations := map[string]string{
		"stopped": `UPDATE flash_sale_campaigns SET status = 'stopped'`,
		"ended":   `UPDATE flash_sale_campaigns SET starts_at = now() - interval '2 hours', ends_at = now() - interval '1 minute'`,
		"not yet": `UPDATE flash_sale_campaigns SET starts_at = now() + interval '1 hour', ends_at = now() + interval '2 hours'`,
		"draft":   `UPDATE flash_sale_campaigns SET status = 'draft'`,
	}
	for name, mutation := range mutations {
		t.Run(name, func(t *testing.T) {
			database, customer, skuID, saleItem := cartSetup(t, "preview-"+name, cartSaleOptions{}, 100)
			line := cartFlashLine(t, customer.testClient, skuID, saleItem, 1)
			cartExec(t, database, mutation)

			cartExpectLineError(t, customer.testClient, line, "SALE_NOT_ACTIVE", nil)
		})
	}
}

func TestCheckoutPreviewInsufficientStock(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "preview-stock")
	skuID := cartSeedSKU(t, database, cartSKUOptions{OnHand: 5})
	line := cartAdd(t, customer.testClient, skuID, 4)
	cartExec(t, database, `UPDATE inventories SET reserved = 3 WHERE sku_id = ?`, skuID)

	cartExpectLineError(t, customer.testClient, line, "INSUFFICIENT_STOCK", cartInt(2))
}

func TestCheckoutPreviewFlashLineNeedsQuotaAndStockAndReportsQuotaFirst(t *testing.T) {
	database, customer, skuID, saleItem := cartSetup(t, "preview-priority", cartSaleOptions{QuotaTotal: 5, QuotaSold: 4, PurchaseLimit: 5}, 3)
	line := cartFlashLine(t, customer.testClient, skuID, saleItem, 2)
	cartExec(t, database, `UPDATE inventories SET reserved = 3 WHERE sku_id = ?`, skuID)

	// Both quota (1 left) and stock (0 left) fail; the sale-specific reason wins.
	cartExpectLineError(t, customer.testClient, line, "SALE_QUOTA_EXCEEDED", cartInt(1))
}

func TestCheckoutPreviewRegularQuantityAboveSKUCapAfterCapLowered(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "preview-cap")
	skuID := cartSeedSKU(t, database, cartSKUOptions{MaxQuantity: 10})
	line := cartAdd(t, customer.testClient, skuID, 6)
	cartExec(t, database, `UPDATE skus SET max_quantity_per_order = 4 WHERE id = ?`, skuID)

	cartExpectLineError(t, customer.testClient, line, "PURCHASE_LIMIT_EXCEEDED", cartInt(4))
}
