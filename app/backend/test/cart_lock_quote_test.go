package backend_test

import (
	"errors"
	"testing"

	"cloudian/cloudian-scale/internal/module/cart"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func lockQuoteSelection(items ...cart.CartItemResponse) cart.CheckoutSelection {
	selection := cart.CheckoutSelection{Items: make([]cart.SelectedCartItem, len(items))}
	for i, item := range items {
		selection.Items[i] = cart.SelectedCartItem{CartItemID: item.ID, Version: item.Version}
	}
	return selection
}

func TestLockAndQuoteMatchesPreviewForMixedLines(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "locked-quote-mixed")
	regularSKU := cartSeedSKU(t, database, cartSKUOptions{Price: 100_000})
	flashSKU := cartSeedSKU(t, database, cartSKUOptions{Price: 90_000})
	saleItem := cartSeedSale(t, database, flashSKU, cartSaleOptions{SalePrice: 40_000, PurchaseLimit: 3})
	regular := cartAdd(t, customer.testClient, regularSKU, 2)
	flashResponse := cartPut(t, customer.testClient, flashSKU, map[string]any{
		"quantity": 3, "price_source": "flash_sale", "flash_sale_item_id": saleItem,
	})
	if flashResponse.Code != 200 {
		t.Fatalf("add flash-sale item status = %d; body = %s", flashResponse.Code, flashResponse.Body.String())
	}
	flash := cartFindItem(t, decodeJSON[cart.CartResponse](t, flashResponse), flashSKU)
	selection := lockQuoteSelection(flash, regular)
	service, err := cart.NewService(database)
	if err != nil {
		t.Fatal(err)
	}
	preview, err := service.PreviewCheckout(t.Context(), customer.ID, selection)
	if err != nil {
		t.Fatalf("preview checkout: %v", err)
	}

	var quote cart.LockedQuote
	if err := database.Transaction(func(tx *gorm.DB) error {
		var err error
		quote, err = cart.LockAndQuote(tx, customer.ID, selection)
		return err
	}); err != nil {
		t.Fatalf("lock and quote: %v", err)
	}
	if quote.TotalAmount != preview.TotalAmount || len(quote.Lines) != len(preview.Lines) {
		t.Fatalf("locked quote = %+v, preview = %+v", quote, preview)
	}
	for i, line := range quote.Lines {
		previewLine := preview.Lines[i]
		flashItemMismatch := (line.FlashSaleItemID == nil) != (previewLine.FlashSaleItemID == nil)
		if line.FlashSaleItemID != nil && previewLine.FlashSaleItemID != nil {
			flashItemMismatch = *line.FlashSaleItemID != *previewLine.FlashSaleItemID
		}
		if line.CartItemID != preview.Lines[i].CartItemID || line.CartItemVersion != preview.Lines[i].CartItemVersion ||
			line.SKUID != preview.Lines[i].SKUID || line.Quantity != preview.Lines[i].Quantity ||
			line.UnitPrice != preview.Lines[i].UnitPrice || line.Subtotal != preview.Lines[i].Subtotal ||
			line.PriceSource != previewLine.PriceSource || flashItemMismatch {
			t.Errorf("locked line %d = %+v, preview line = %+v", i, line, preview.Lines[i])
		}
		if line.SKUCode == "" || line.ProductName == "" || len(line.SKUAttributes) == 0 {
			t.Errorf("locked line %d missing SKU snapshot: %+v", i, line)
		}
	}
}

func TestLockAndQuoteHoldsInventoryAndCartLocks(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "locked-quote-locks")
	skuID := cartSeedSKU(t, database, cartSKUOptions{})
	item := cartAdd(t, customer.testClient, skuID, 1)
	selection := lockQuoteSelection(item)

	err := database.Transaction(func(tx *gorm.DB) error {
		if _, err := cart.LockAndQuote(tx, customer.ID, selection); err != nil {
			return err
		}
		for _, query := range []struct {
			name string
			sql  string
			id   string
		}{{"inventory", `SELECT sku_id FROM inventories WHERE sku_id = ? FOR UPDATE NOWAIT`, skuID},
			{"cart item", `SELECT id FROM cart_items WHERE id = ? FOR UPDATE NOWAIT`, item.ID}} {
			var result string
			err := database.Raw(query.sql, query.id).Scan(&result).Error
			var pgError *pgconn.PgError
			if !errors.As(err, &pgError) || pgError.Code != "55P03" {
				return errors.New(query.name + " lock error = " + errorString(err) + ", want SQLSTATE 55P03")
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLockAndQuoteCreatesZeroUsageRowForFirstFlashSaleLine(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "locked-quote-usage")
	regularSKU := cartSeedSKU(t, database, cartSKUOptions{})
	flashSKU := cartSeedSKU(t, database, cartSKUOptions{})
	saleItem := cartSeedSale(t, database, flashSKU, cartSaleOptions{})
	regular := cartAdd(t, customer.testClient, regularSKU, 1)
	flashResponse := cartPut(t, customer.testClient, flashSKU, map[string]any{
		"quantity": 1, "price_source": "flash_sale", "flash_sale_item_id": saleItem,
	})
	if flashResponse.Code != 200 {
		t.Fatalf("add flash-sale item status = %d; body = %s", flashResponse.Code, flashResponse.Body.String())
	}
	flash := cartFindItem(t, decodeJSON[cart.CartResponse](t, flashResponse), flashSKU)
	selection := lockQuoteSelection(regular, flash)

	rollback := errors.New("rollback usage lock test")
	err := database.Transaction(func(tx *gorm.DB) error {
		if _, err := cart.LockAndQuote(tx, customer.ID, selection); err != nil {
			return err
		}
		var rows []struct {
			FlashSaleItemID string `gorm:"column:flash_sale_item_id"`
			Reserved        int64  `gorm:"column:reserved_quantity"`
			Sold            int64  `gorm:"column:sold_quantity"`
		}
		if err := tx.Raw(`SELECT flash_sale_item_id, reserved_quantity, sold_quantity
			FROM flash_sale_customer_usage WHERE user_id = ?`, customer.ID).Scan(&rows).Error; err != nil {
			return err
		}
		if len(rows) != 1 || rows[0].FlashSaleItemID != saleItem || rows[0].Reserved != 0 || rows[0].Sold != 0 {
			return errors.New("flash-sale usage row = " + errorString(rows) + ", want exactly one zeroed row")
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("usage check transaction error = %v, want test rollback", err)
	}
	if count := cartCount(t, database, `SELECT count(*) FROM flash_sale_customer_usage WHERE user_id = ? AND flash_sale_item_id = ?`, customer.ID, saleItem); count != 0 {
		t.Fatalf("usage rows after rollback = %d, want 0", count)
	}
}

func TestLockAndQuoteStaleVersionReturnsCartItemChanged(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "locked-quote-stale")
	skuID := cartSeedSKU(t, database, cartSKUOptions{})
	item := cartAdd(t, customer.testClient, skuID, 1)
	selection := lockQuoteSelection(item)
	selection.Items[0].Version++

	err := database.Transaction(func(tx *gorm.DB) error {
		_, err := cart.LockAndQuote(tx, customer.ID, selection)
		return err
	})
	if !errors.Is(err, cart.ErrCartItemChanged) {
		t.Fatalf("LockAndQuote error = %v, want ErrCartItemChanged", err)
	}
}

func errorString(value any) string {
	if err, ok := value.(error); ok {
		return err.Error()
	}
	return "unexpected result"
}
