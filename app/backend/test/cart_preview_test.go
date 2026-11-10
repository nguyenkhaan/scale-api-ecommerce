package backend_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/module/cart"
	"gorm.io/gorm"
)

// cartStateDigest fingerprints every table a preview must leave untouched.
func cartStateDigest(t *testing.T, database *gorm.DB) string {
	t.Helper()
	var digest string
	err := database.Raw(`SELECT md5(
		coalesce((SELECT string_agg(x::text, '|' ORDER BY x::text) FROM inventories x), '') ||
		coalesce((SELECT string_agg(x::text, '|' ORDER BY x::text) FROM flash_sale_items x), '') ||
		coalesce((SELECT string_agg(x::text, '|' ORDER BY x::text) FROM flash_sale_customer_usage x), '') ||
		coalesce((SELECT string_agg(x::text, '|' ORDER BY x::text) FROM cart_items x), '') ||
		coalesce((SELECT string_agg(x::text, '|' ORDER BY x::text) FROM carts x), ''))`).Scan(&digest).Error
	if err != nil {
		t.Fatalf("state digest: %v", err)
	}
	return digest
}

func TestCheckoutPreviewQuotesRegularAndFlashLines(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "preview-pricing")

	regularSKU := cartSeedSKU(t, database, cartSKUOptions{Price: 100_000})
	flashSKU := cartSeedSKU(t, database, cartSKUOptions{Price: 90_000})
	saleItem := cartSeedSale(t, database, flashSKU, cartSaleOptions{SalePrice: 40_000, PurchaseLimit: 3})

	regular := cartAdd(t, customer.testClient, regularSKU, 2)
	response := cartPut(t, customer.testClient, flashSKU, map[string]any{"quantity": 3, "price_source": "flash_sale", "flash_sale_item_id": saleItem})
	if response.Code != http.StatusOK {
		t.Fatalf("flash add status = %d; body = %s", response.Code, response.Body.String())
	}
	flash := cartFindItem(t, decodeJSON[cart.CartResponse](t, response), flashSKU)

	before := cartStateDigest(t, database)
	res := cartPreview(t, customer.testClient, flash, regular)
	if res.Code != http.StatusOK {
		t.Fatalf("preview status = %d; body = %s", res.Code, res.Body.String())
	}
	preview := decodeJSON[cart.CheckoutPreviewResponse](t, res)
	if preview.Currency != "VND" || preview.TotalAmount != 3*40_000+2*100_000 {
		t.Errorf("preview = %+v, want VND total 320000", preview)
	}
	if len(preview.Lines) != 2 || preview.Lines[0].CartItemID != flash.ID || preview.Lines[0].UnitPrice != 40_000 ||
		preview.Lines[0].PriceSource != "flash_sale" || preview.Lines[1].UnitPrice != 100_000 || preview.Lines[1].Subtotal != 200_000 {
		t.Errorf("lines = %+v, want selection order flash then regular", preview.Lines)
	}
	if _, err := time.Parse(time.RFC3339, preview.PreviewedAt); err != nil {
		t.Errorf("previewed_at %q: %v", preview.PreviewedAt, err)
	}
	if after := cartStateDigest(t, database); after != before {
		t.Error("preview changed inventory, quota, usage or cart rows")
	}
}

func TestCheckoutPreviewNeverWritesEvenWhenRejected(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "preview-readonly")

	okSKU := cartSeedSKU(t, database, cartSKUOptions{})
	saleSKU := cartSeedSKU(t, database, cartSKUOptions{})
	saleItem := cartSeedSale(t, database, saleSKU, cartSaleOptions{})
	ok := cartAdd(t, customer.testClient, okSKU, 1)
	flashResponse := cartPut(t, customer.testClient, saleSKU, map[string]any{"quantity": 2, "price_source": "flash_sale", "flash_sale_item_id": saleItem})
	flash := cartFindItem(t, decodeJSON[cart.CartResponse](t, flashResponse), saleSKU)
	cartExec(t, database, `UPDATE flash_sale_campaigns SET status = 'stopped'`)
	cartExec(t, database, `INSERT INTO flash_sale_customer_usage (flash_sale_item_id, user_id, reserved_quantity) VALUES (?, ?, 1)`, saleItem, customer.ID)

	before := cartStateDigest(t, database)
	requireProblem(t, cartPreview(t, customer.testClient, ok, flash), http.StatusConflict, "SALE_NOT_ACTIVE")
	if after := cartStateDigest(t, database); after != before {
		t.Error("rejected preview changed stored state")
	}
}

func TestCheckoutPreviewReportsAllFailingLinesWithMostSevereCode(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "preview-multi")

	healthySKU := cartSeedSKU(t, database, cartSKUOptions{})
	hiddenSKU := cartSeedSKU(t, database, cartSKUOptions{})
	lowStockSKU := cartSeedSKU(t, database, cartSKUOptions{OnHand: 5})
	healthy := cartAdd(t, customer.testClient, healthySKU, 1)
	hidden := cartAdd(t, customer.testClient, hiddenSKU, 2)
	lowStock := cartAdd(t, customer.testClient, lowStockSKU, 4)
	cartExec(t, database, `UPDATE categories SET is_visible = false WHERE id = (SELECT p.category_id FROM products p JOIN skus s ON s.product_id = p.id WHERE s.id = ?)`, hiddenSKU)
	cartExec(t, database, `UPDATE inventories SET reserved = 3 WHERE sku_id = ?`, lowStockSKU)

	body := requireProblem(t, cartPreview(t, customer.testClient, lowStock, healthy, hidden), http.StatusConflict, "ITEM_UNAVAILABLE")
	if len(body.LineErrors) != 2 {
		t.Fatalf("line_errors = %+v, want the two failing lines only", body.LineErrors)
	}
	stock, unavailable := body.LineErrors[0], body.LineErrors[1]
	if stock.SKUID != lowStockSKU || stock.Code != "INSUFFICIENT_STOCK" || stock.RequestedQuantity != 4 || stock.AvailableQuantity == nil || *stock.AvailableQuantity != 2 {
		t.Errorf("stock line = %+v, want INSUFFICIENT_STOCK requested 4 available 2", stock)
	}
	if unavailable.SKUID != hiddenSKU || unavailable.Code != "ITEM_UNAVAILABLE" || unavailable.RequestedQuantity != 2 {
		t.Errorf("hidden line = %+v, want ITEM_UNAVAILABLE requested 2", unavailable)
	}
}

func TestCheckoutPreviewLineBecomesUnavailableAfterAdd(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "preview-hidden")

	cases := map[string]string{
		"hidden category":  `UPDATE categories SET is_visible = false WHERE id = (SELECT p.category_id FROM products p JOIN skus s ON s.product_id = p.id WHERE s.id = ?)`,
		"inactive product": `UPDATE products SET status = 'inactive' WHERE id = (SELECT product_id FROM skus WHERE id = ?)`,
		"inactive SKU":     `UPDATE skus SET status = 'inactive' WHERE id = ?`,
	}
	for name, mutation := range cases {
		t.Run(name, func(t *testing.T) {
			skuID := cartSeedSKU(t, database, cartSKUOptions{})
			item := cartAdd(t, customer.testClient, skuID, 1)
			cartExec(t, database, mutation, skuID)
			body := requireProblem(t, cartPreview(t, customer.testClient, item), http.StatusConflict, "ITEM_UNAVAILABLE")
			if len(body.LineErrors) != 1 || body.LineErrors[0].Code != "ITEM_UNAVAILABLE" || body.LineErrors[0].SKUID != skuID {
				t.Errorf("line_errors = %+v", body.LineErrors)
			}
		})
	}
}

func TestCheckoutPreviewSelectionValidationAndStaleVersion(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "preview-selection")
	skuID := cartSeedSKU(t, database, cartSKUOptions{})
	item := cartAdd(t, customer.testClient, skuID, 1)

	post := func(body string) int {
		return customer.send(t, http.MethodPost, "/api/checkout/preview", body, customer.csrfHeader()...).Code
	}
	if code := post(`{"items":[{"cart_item_id":"` + item.ID + `","version":0}]}`); code != http.StatusUnprocessableEntity {
		t.Errorf("version 0 status = %d, want 422", code)
	}
	duplicate := `{"items":[{"cart_item_id":"` + item.ID + `","version":1},{"cart_item_id":"` + strings.ToUpper(item.ID) + `","version":1}]}`
	if code := post(duplicate); code != http.StatusUnprocessableEntity {
		t.Errorf("duplicate selection status = %d, want 422", code)
	}
	var tooMany []string
	for i := 0; i < cart.MaxCartItems+1; i++ {
		tooMany = append(tooMany, `{"cart_item_id":"00000000-0000-4000-8000-0000000000`+fmt.Sprintf("%02d", i)+`","version":1}`)
	}
	if code := post(`{"items":[` + strings.Join(tooMany, ",") + `]}`); code != http.StatusUnprocessableEntity {
		t.Errorf("21 items status = %d, want 422", code)
	}

	// A write after the customer loaded the cart makes the confirmed version stale.
	cartAdd(t, customer.testClient, skuID, 2)
	requireProblem(t, cartPreview(t, customer.testClient, item), http.StatusConflict, "CART_ITEM_CHANGED")
}

func TestCheckoutPreviewDoesNotExposeOtherCustomersCarts(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	owner := cartNewCustomer(t, router, database, "preview-owner")
	stranger := cartNewCustomer(t, router, database, "preview-stranger")
	skuID := cartSeedSKU(t, database, cartSKUOptions{})
	item := cartAdd(t, owner.testClient, skuID, 1)

	res := cartPreview(t, stranger.testClient, item)
	body := requireProblem(t, res, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
	if len(body.Violations) != 1 || body.Violations[0].Field != "items[0].cart_item_id" {
		t.Errorf("violations = %+v, want items[0].cart_item_id", body.Violations)
	}
	for _, leaked := range []string{skuID, item.SKUCode, item.ProductName} {
		if strings.Contains(res.Body.String(), leaked) {
			t.Errorf("response leaks %q: %s", leaked, res.Body.String())
		}
	}

	del := stranger.send(t, http.MethodDelete, "/api/cart/items/"+skuID, nil, stranger.csrfHeader()...)
	if del.Code != http.StatusNoContent {
		t.Errorf("stranger delete status = %d, want 204", del.Code)
	}
	if view := cartGet(t, owner.testClient); len(view.Items) != 1 || view.Items[0].ID != item.ID {
		t.Errorf("owner's cart changed by another customer: %+v", view.Items)
	}
}
