package backend_test

import (
	"net/http"
	"slices"
	"testing"
)

func TestOrderListPagesNewestFirstForOwnerOnly(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := orderAPINewRouter(t, database)
	owner := cartNewCustomer(t, router, database, "order-owner")
	other := cartNewCustomer(t, router, database, "order-other")

	var seeded []string // oldest first
	for range 5 {
		orderID, _ := orderAPISeedOrder(t, database, owner.ID, 300, 1)
		seeded = append(seeded, orderID)
	}
	foreignOrder, _ := orderAPISeedOrder(t, database, other.ID, 300, 1)
	want := slices.Clone(seeded)
	slices.Reverse(want)

	var got []string
	query := "?limit=2"
	pages := 0
	for {
		page := orderAPIList(t, owner.testClient, query)
		pages++
		for _, item := range page.Items {
			got = append(got, item.ID)
		}
		if page.NextCursor == nil {
			break
		}
		if pages > 5 {
			t.Fatal("pagination did not terminate")
		}
		query = "?limit=2&cursor=" + *page.NextCursor
	}
	if pages != 3 || !slices.Equal(got, want) {
		t.Errorf("pages = %d, ids = %v; want 3 pages in newest-first order %v", pages, got, want)
	}
	if slices.Contains(got, foreignOrder) {
		t.Error("owner list contains another customer's order")
	}
	if theirs := orderAPIList(t, other.testClient, ""); len(theirs.Items) != 1 || theirs.Items[0].ID != foreignOrder {
		t.Errorf("other customer's list = %+v, want only their own order", theirs.Items)
	}
	if empty := orderAPIList(t, owner.testClient, "?status=confirmed"); empty.Items == nil || len(empty.Items) != 0 {
		t.Errorf("confirmed filter = %+v, want empty non-null items", empty.Items)
	}
	lifecycleRequireInvariants(t, database)
}

func TestOrderListStatusFilterAndValidation(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := orderAPINewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "order-filter")

	pendingID, _ := orderAPISeedOrder(t, database, customer.ID, 300, 1)
	cancelledID, _ := orderAPISeedOrder(t, database, customer.ID, 300, 1)
	if status, body := orderAPICancel(t, customer.testClient, cancelledID); status != http.StatusOK {
		t.Fatalf("cancel status = %d; body = %s", status, body)
	}

	page := orderAPIList(t, customer.testClient, "?status=cancelled")
	if len(page.Items) != 1 || page.Items[0].ID != cancelledID || page.Items[0].Status != "cancelled" {
		t.Errorf("cancelled filter = %+v, want only %s", page.Items, cancelledID)
	}
	page = orderAPIList(t, customer.testClient, "?status=pending_payment")
	if len(page.Items) != 1 || page.Items[0].ID != pendingID {
		t.Errorf("pending filter = %+v, want only %s", page.Items, pendingID)
	}

	cursorFromAll := orderAPIList(t, customer.testClient, "?limit=1").NextCursor
	if cursorFromAll == nil {
		t.Fatal("expected a next cursor with two orders and limit 1")
	}
	for name, query := range map[string]string{
		"unknown status":              "?status=shipped",
		"limit zero":                  "?limit=0",
		"limit too large":             "?limit=101",
		"limit not a number":          "?limit=abc",
		"garbage cursor":              "?cursor=!!!",
		"cursor from another filter":  "?status=cancelled&cursor=" + *cursorFromAll,
		"well formed but foreign key": "?cursor=e30",
	} {
		t.Run(name, func(t *testing.T) {
			requireProblem(t, customer.send(t, http.MethodGet, "/api/orders"+query, nil),
				http.StatusUnprocessableEntity, "VALIDATION_ERROR")
		})
	}

	anonymous := newTestClient(router)
	requireProblem(t, anonymous.send(t, http.MethodGet, "/api/orders", nil), http.StatusUnauthorized, "UNAUTHENTICATED")
	requireProblem(t, anonymous.send(t, http.MethodGet, "/api/orders/"+pendingID, nil), http.StatusUnauthorized, "UNAUTHENTICATED")
}

func TestOrderDetailSnapshotAndOwnership(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := orderAPINewRouter(t, database)
	owner := cartNewCustomer(t, router, database, "detail-owner")
	other := cartNewCustomer(t, router, database, "detail-other")

	saleSKU := cartSeedSKU(t, database, cartSKUOptions{OnHand: 20})
	regularSKU := cartSeedSKU(t, database, cartSKUOptions{OnHand: 20})
	sale := cartSeedSale(t, database, saleSKU, cartSaleOptions{QuotaTotal: 5, PurchaseLimit: 3})
	orderID := lifecycleSeedPendingOrder(t, database, owner.ID, 300,
		lifecycleLine{SKUID: saleSKU, Quantity: 2, UnitPrice: 40_000, FlashSaleItemID: sale},
		lifecycleLine{SKUID: regularSKU, Quantity: 1, UnitPrice: 100_000},
	)
	if err := database.Exec(`UPDATE order_items SET sku_attributes = '{"color":"red"}' WHERE order_id = ? AND sku_id = ?`,
		orderID, regularSKU).Error; err != nil {
		t.Fatalf("set attributes: %v", err)
	}

	got := orderAPIGet(t, owner.testClient, "/api/orders/"+orderID)
	if got.ID != orderID || got.Status != "pending_payment" || got.Currency != "VND" || got.TotalAmount != 180_000 ||
		got.OrderNumber == "" || got.CreatedAt.IsZero() || !got.ExpiresAt.After(got.CreatedAt) {
		t.Errorf("order header = %+v", got)
	}
	if got.Recipient.Name != "Lan" || got.Recipient.Phone != "0901234567" || got.Recipient.Address == "" {
		t.Errorf("recipient = %+v", got.Recipient)
	}
	if len(got.Items) != 2 {
		t.Fatalf("items = %+v, want 2 lines", got.Items)
	}
	bySKU := map[string]int{got.Items[0].SKUID: 0, got.Items[1].SKUID: 1}
	flash, regular := got.Items[bySKU[saleSKU]], got.Items[bySKU[regularSKU]]
	if flash.PriceSource != "flash_sale" || flash.Quantity != 2 || flash.UnitPrice != 40_000 || flash.Subtotal != 80_000 ||
		flash.SKUCode == "" || flash.ProductName == "" || string(flash.SKUAttributes) != "{}" {
		t.Errorf("flash line = %+v", flash)
	}
	if regular.PriceSource != "regular" || regular.Subtotal != 100_000 || string(regular.SKUAttributes) != `{"color":"red"}` {
		t.Errorf("regular line = %+v attrs=%s", regular, regular.SKUAttributes)
	}

	for name, id := range map[string]string{
		"another customer's order": orderID,
		"unknown id":               "11111111-2222-4333-8444-555555555555",
		"malformed id":             "not-a-uuid",
	} {
		t.Run(name, func(t *testing.T) {
			requireProblem(t, other.send(t, http.MethodGet, "/api/orders/"+id, nil), http.StatusNotFound, "RESOURCE_NOT_FOUND")
		})
	}
}

func TestOrderPayableUsesDatabaseDeadline(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := orderAPINewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "payable")

	live, _ := orderAPISeedOrder(t, database, customer.ID, 300, 1)
	overdue, _ := orderAPISeedOrder(t, database, customer.ID, -5, 1) // worker has not run yet

	if got := orderAPIGet(t, customer.testClient, "/api/orders/"+live); !got.Payable || got.Status != "pending_payment" {
		t.Errorf("live order = status %s payable %v, want pending and payable", got.Status, got.Payable)
	}
	got := orderAPIGet(t, customer.testClient, "/api/orders/"+overdue)
	if got.Payable || got.Status != "pending_payment" {
		t.Errorf("overdue order = status %s payable %v, want still pending but not payable", got.Status, got.Payable)
	}
	inList := orderAPIList(t, customer.testClient, "")
	for _, item := range inList.Items {
		if item.ID == overdue && item.Payable {
			t.Error("list reports an overdue order as payable")
		}
	}
	lifecycleRequireInvariants(t, database)
}

func TestOrderPaymentProjection(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := orderAPINewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "projection")
	orderID, _ := orderAPISeedOrder(t, database, customer.ID, 300, 1)

	got := orderAPIGet(t, customer.testClient, "/api/orders/"+orderID)
	if got.PaymentStatus != nil || got.ReconciliationRequired {
		t.Errorf("fresh order payment_status=%v reconciliation_required=%v, want null/false", got.PaymentStatus, got.ReconciliationRequired)
	}

	cartExec(t, database, `INSERT INTO payment_sessions (order_id, status, amount, simulator_reference)
		VALUES (?, 'unknown', 25000, 'sim-projection-1')`, orderID)
	got = orderAPIGet(t, customer.testClient, "/api/orders/"+orderID)
	if got.PaymentStatus == nil || *got.PaymentStatus != "unknown" || got.ReconciliationRequired {
		t.Errorf("with session payment_status=%v reconciliation_required=%v, want unknown/false", got.PaymentStatus, got.ReconciliationRequired)
	}

	cartExec(t, database, `WITH event AS (
			INSERT INTO payment_events (event_id, payment_session_id, event_type, amount, currency, payload_hash,
				source_verified, result)
			SELECT gen_random_uuid(), ps.id, 'succeeded', 25000, 'VND', decode(repeat('00', 32), 'hex'), true, 'late'
			FROM payment_sessions ps WHERE ps.order_id = ? RETURNING id)
		INSERT INTO payment_reconciliations (payment_event_id, order_id, reason) SELECT id, ?, 'late success' FROM event`,
		orderID, orderID)
	got = orderAPIGet(t, customer.testClient, "/api/orders/"+orderID)
	if !got.ReconciliationRequired {
		t.Error("open reconciliation case not reported")
	}
	cartExec(t, database, `UPDATE payment_reconciliations SET status = 'reviewed', reviewed_at = now() WHERE order_id = ?`, orderID)
	if got = orderAPIGet(t, customer.testClient, "/api/orders/"+orderID); got.ReconciliationRequired {
		t.Error("reviewed reconciliation case still reported as required")
	}
}
