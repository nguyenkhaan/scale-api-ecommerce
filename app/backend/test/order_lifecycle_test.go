package backend_test

import (
	"errors"
	"sync"
	"testing"

	"cloudian/cloudian-scale/internal/module/order"
	"gorm.io/gorm"
)

// lifecycleFixture seeds one customer with a pending order of two flash sale units and one regular unit.
type lifecycleFixture struct {
	db                        *gorm.DB
	userID, orderID           string
	saleSKU, regularSKU, sale string
	saleBefore, regularBefore lifecycleCounters
}

func newLifecycleFixture(t *testing.T, database *gorm.DB, expiresInSecs float64) lifecycleFixture {
	t.Helper()
	fixture := lifecycleFixture{db: database}
	fixture.userID = lifecycleSeedUser(t, database, "buyer@example.com")
	fixture.saleSKU = cartSeedSKU(t, database, cartSKUOptions{OnHand: 10})
	fixture.regularSKU = cartSeedSKU(t, database, cartSKUOptions{OnHand: 10})
	fixture.sale = cartSeedSale(t, database, fixture.saleSKU, cartSaleOptions{QuotaTotal: 5, PurchaseLimit: 3})
	fixture.orderID = lifecycleSeedPendingOrder(t, database, fixture.userID, expiresInSecs,
		lifecycleLine{SKUID: fixture.saleSKU, Quantity: 2, UnitPrice: 40_000, FlashSaleItemID: fixture.sale},
		lifecycleLine{SKUID: fixture.regularSKU, Quantity: 1, UnitPrice: 100_000},
	)
	fixture.saleBefore = lifecycleReadCounters(t, database, fixture.saleSKU, fixture.sale, fixture.userID)
	fixture.regularBefore = lifecycleReadCounters(t, database, fixture.regularSKU, "", fixture.userID)
	return fixture
}

func (fixture lifecycleFixture) finish(t *testing.T, outcome order.Outcome) (order.Transition, error) {
	t.Helper()
	var transition order.Transition
	err := fixture.db.Transaction(func(tx *gorm.DB) error {
		var err error
		transition, err = order.Finish(tx, fixture.orderID, outcome)
		return err
	})
	return transition, err
}

func (fixture lifecycleFixture) orderStatus(t *testing.T) string {
	t.Helper()
	var status string
	fixture.db.Raw("SELECT status FROM orders WHERE id = ?", fixture.orderID).Scan(&status)
	return status
}

func TestFinishConfirmedSpendsStockAndQuotaOnce(t *testing.T) {
	fixture := newLifecycleFixture(t, openIntegrationDatabase(t).ORM(), 300)

	first, err := fixture.finish(t, order.OutcomeConfirmed)
	if err != nil || !first.Applied || first.Status != "confirmed" {
		t.Fatalf("first confirm = %+v, %v; want applied confirmed", first, err)
	}
	second, err := fixture.finish(t, order.OutcomeConfirmed)
	if err != nil || second.Applied || second.Status != "confirmed" {
		t.Fatalf("repeated confirm = %+v, %v; want no-op confirmed", second, err)
	}

	sale := lifecycleReadCounters(t, fixture.db, fixture.saleSKU, fixture.sale, fixture.userID)
	want := lifecycleCounters{
		OnHand: fixture.saleBefore.OnHand - 2, Reserved: fixture.saleBefore.Reserved - 2,
		QuotaReserved: fixture.saleBefore.QuotaReserved - 2, QuotaSold: fixture.saleBefore.QuotaSold + 2,
		UsageReserved: fixture.saleBefore.UsageReserved - 2, UsageSold: fixture.saleBefore.UsageSold + 2,
	}
	if sale != want {
		t.Errorf("flash sale counters = %+v, want %+v", sale, want)
	}
	regular := lifecycleReadCounters(t, fixture.db, fixture.regularSKU, "", fixture.userID)
	if regular.OnHand != fixture.regularBefore.OnHand-1 || regular.Reserved != fixture.regularBefore.Reserved-1 {
		t.Errorf("regular counters = %+v, want one unit sold from %+v", regular, fixture.regularBefore)
	}
	lifecycleRequireInvariants(t, fixture.db)
}

func TestFinishReleaseOutcomesOnlyReturnTheHold(t *testing.T) {
	for _, outcome := range []order.Outcome{order.OutcomeCancelled, order.OutcomePaymentFailed} {
		t.Run(string(outcome), func(t *testing.T) {
			fixture := newLifecycleFixture(t, openIntegrationDatabase(t).ORM(), 300)

			if transition, err := fixture.finish(t, outcome); err != nil || !transition.Applied || transition.Status != string(outcome) {
				t.Fatalf("finish = %+v, %v", transition, err)
			}
			if transition, _ := fixture.finish(t, outcome); transition.Applied {
				t.Fatal("repeated release was applied twice")
			}

			sale := lifecycleReadCounters(t, fixture.db, fixture.saleSKU, fixture.sale, fixture.userID)
			want := fixture.saleBefore
			want.Reserved, want.QuotaReserved, want.UsageReserved = want.Reserved-2, want.QuotaReserved-2, want.UsageReserved-2
			if sale != want {
				t.Errorf("counters after %s = %+v, want %+v (on_hand and sold unchanged)", outcome, sale, want)
			}
			lifecycleRequireInvariants(t, fixture.db)
		})
	}
}

func TestFinishOnOverdueOrderExpiresInsteadOfConfirming(t *testing.T) {
	fixture := newLifecycleFixture(t, openIntegrationDatabase(t).ORM(), -1)

	transition, err := fixture.finish(t, order.OutcomeConfirmed)
	if err != nil || !transition.Applied || transition.Status != "expired" || !transition.DeadlinePassed {
		t.Fatalf("confirm on overdue order = %+v, %v; want applied expired", transition, err)
	}
	sale := lifecycleReadCounters(t, fixture.db, fixture.saleSKU, fixture.sale, fixture.userID)
	if sale.OnHand != fixture.saleBefore.OnHand || sale.QuotaSold != fixture.saleBefore.QuotaSold {
		t.Errorf("overdue order sold stock: %+v (before %+v)", sale, fixture.saleBefore)
	}
	lifecycleRequireInvariants(t, fixture.db)
}

func TestFinishExpiredBeforeDeadlineIsRefused(t *testing.T) {
	fixture := newLifecycleFixture(t, openIntegrationDatabase(t).ORM(), 300)

	if _, err := fixture.finish(t, order.OutcomeExpired); !errors.Is(err, order.ErrOrderNotDue) {
		t.Fatalf("early expiry error = %v, want ErrOrderNotDue", err)
	}
	if status := fixture.orderStatus(t); status != "pending_payment" {
		t.Errorf("status = %s, want pending_payment", status)
	}
}

// TEST-15 core: payment success racing a cancellation ends in exactly one terminal state.
func TestFinishConcurrentConfirmAndCancelHaveOneWinner(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	for round := 0; round < 10; round++ {
		fixture := newLifecycleFixture(t, database, 300)
		var waitGroup sync.WaitGroup
		results := make(chan order.Transition, 4)
		for _, outcome := range []order.Outcome{order.OutcomeConfirmed, order.OutcomeCancelled, order.OutcomeConfirmed, order.OutcomeCancelled} {
			waitGroup.Go(func() {
				var transition order.Transition
				if err := database.Transaction(func(tx *gorm.DB) error {
					var err error
					transition, err = order.Finish(tx, fixture.orderID, outcome)
					return err
				}); err != nil {
					t.Errorf("round %d finish %s: %v", round, outcome, err)
				}
				results <- transition
			})
		}
		waitGroup.Wait()
		close(results)
		applied := 0
		for transition := range results {
			if transition.Applied {
				applied++
			}
		}
		if applied != 1 {
			t.Fatalf("round %d: %d transitions applied, want exactly 1", round, applied)
		}
		lifecycleRequireInvariants(t, database)
		database.Exec("TRUNCATE cart_items, inventory_reservations, order_items, orders, flash_sale_customer_usage, flash_sale_items, flash_sale_campaigns, inventories, skus, products, categories, users CASCADE")
	}
}

// TEST-25 core: a confirmed order removes only cart lines still at the ordered version.
func TestFinishConfirmedRemovesOnlyUnchangedCartLines(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	userID := lifecycleSeedUser(t, database, "cart-owner@example.com")
	unchangedSKU := cartSeedSKU(t, database, cartSKUOptions{})
	editedSKU := cartSeedSKU(t, database, cartSKUOptions{})
	var cartID string
	if err := database.Raw("INSERT INTO carts (user_id) VALUES (?) RETURNING id", userID).Scan(&cartID).Error; err != nil {
		t.Fatalf("seed cart: %v", err)
	}
	addLine := func(skuID string) string {
		var lineID string
		if err := database.Raw(
			`INSERT INTO cart_items (cart_id, sku_id, quantity, intended_price_source, version)
			 VALUES (?, ?, 1, 'regular', 1) RETURNING id`, cartID, skuID,
		).Scan(&lineID).Error; err != nil {
			t.Fatalf("seed cart line: %v", err)
		}
		return lineID
	}
	unchangedLine, editedLine := addLine(unchangedSKU), addLine(editedSKU)
	orderID := lifecycleSeedPendingOrder(t, database, userID, 300,
		lifecycleLine{SKUID: unchangedSKU, Quantity: 1, UnitPrice: 100_000, CartItemID: unchangedLine, CartItemVersion: 1},
		lifecycleLine{SKUID: editedSKU, Quantity: 1, UnitPrice: 100_000, CartItemID: editedLine, CartItemVersion: 1},
	)
	// The customer edits one line after ordering; its version moves on.
	database.Exec("UPDATE cart_items SET quantity = 3, version = 2 WHERE id = ?", editedLine)

	if err := database.Transaction(func(tx *gorm.DB) error {
		_, err := order.Finish(tx, orderID, order.OutcomeConfirmed)
		return err
	}); err != nil {
		t.Fatalf("confirm: %v", err)
	}

	var remaining []string
	database.Raw("SELECT id FROM cart_items WHERE cart_id = ? ORDER BY id", cartID).Scan(&remaining)
	if len(remaining) != 1 || remaining[0] != editedLine {
		t.Errorf("cart lines after confirm = %v, want only the edited line %s", remaining, editedLine)
	}
}
