package backend_test

import (
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/module/order"
	"gorm.io/gorm"
)

// lockingFixture seeds a pending order whose single line came from a cart line, so a confirmation touches
// the cart line and the inventory row that a concurrent checkout of the same customer would also lock.
func lockingFixture(t *testing.T, database *gorm.DB, expiresInSecs float64) (orderID, skuID, cartLineID string) {
	t.Helper()
	userID := lifecycleSeedUser(t, database, "locking@example.com")
	skuID = cartSeedSKU(t, database, cartSKUOptions{OnHand: 10})
	var cartID string
	if err := database.Raw("INSERT INTO carts (user_id) VALUES (?) RETURNING id", userID).Scan(&cartID).Error; err != nil {
		t.Fatalf("seed cart: %v", err)
	}
	if err := database.Raw(
		`INSERT INTO cart_items (cart_id, sku_id, quantity, intended_price_source, version)
		 VALUES (?, ?, 1, 'regular', 1) RETURNING id`, cartID, skuID,
	).Scan(&cartLineID).Error; err != nil {
		t.Fatalf("seed cart line: %v", err)
	}
	orderID = lifecycleSeedPendingOrder(t, database, userID, expiresInSecs,
		lifecycleLine{SKUID: skuID, Quantity: 1, UnitPrice: 100_000, CartItemID: cartLineID, CartItemVersion: 1})
	return orderID, skuID, cartLineID
}

func confirmInBackground(database *gorm.DB, orderID string) <-chan order.Transition {
	results := make(chan order.Transition, 1)
	go func() {
		var transition order.Transition
		err := database.Transaction(func(tx *gorm.DB) error {
			var err error
			transition, err = order.Finish(tx, orderID, order.OutcomeConfirmed)
			return err
		})
		if err != nil {
			transition.Status = "error: " + err.Error()
		}
		results <- transition
	}()
	return results
}

// A checkout holds the customer's cart line and then needs the SKU's inventory row while the payment
// success confirms an earlier order of the same cart. Both must finish: no deadlock (SQLSTATE 40P01).
func TestFinishConfirmLocksCartLinesBeforeInventory(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	orderID, skuID, cartLineID := lockingFixture(t, database, 300)

	checkout := database.Begin()
	defer checkout.Rollback()
	if err := checkout.Exec("SELECT id FROM cart_items WHERE id = ? FOR UPDATE", cartLineID).Error; err != nil {
		t.Fatalf("checkout locks cart line: %v", err)
	}
	results := confirmInBackground(database, orderID)
	time.Sleep(300 * time.Millisecond) // let the confirmation reach its first blocking lock

	if err := checkout.Exec("SELECT sku_id FROM inventories WHERE sku_id = ? FOR UPDATE", skuID).Error; err != nil {
		t.Fatalf("checkout then locks inventory: %v (the confirmation locked inventory before the cart line)", err)
	}
	if err := checkout.Commit().Error; err != nil {
		t.Fatalf("checkout commit: %v", err)
	}
	if transition := <-results; transition.Status != "confirmed" {
		t.Fatalf("confirmation = %+v, want confirmed after the checkout released its locks", transition)
	}
	lifecycleRequireInvariants(t, database)
}

// A success that starts before the deadline but waits on a hot inventory row until after it must expire
// the order, and terminal_at must not predate the deadline (BR-41/42).
func TestFinishRechecksDeadlineAfterWaitingForLocks(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	orderID, skuID, _ := lockingFixture(t, database, 2)

	hotRow := database.Begin()
	defer hotRow.Rollback()
	if err := hotRow.Exec("SELECT sku_id FROM inventories WHERE sku_id = ? FOR UPDATE", skuID).Error; err != nil {
		t.Fatalf("hold inventory row: %v", err)
	}
	results := confirmInBackground(database, orderID)
	time.Sleep(3 * time.Second) // the deadline passes while the confirmation waits
	if err := hotRow.Commit().Error; err != nil {
		t.Fatalf("release inventory row: %v", err)
	}

	transition := <-results
	if transition.Status != "expired" || !transition.DeadlinePassed {
		t.Fatalf("confirmation after waiting past the deadline = %+v, want expired", transition)
	}
	var terminalAfterDeadline bool
	if err := database.Raw("SELECT terminal_at >= expires_at FROM orders WHERE id = ?", orderID).Scan(&terminalAfterDeadline).Error; err != nil {
		t.Fatalf("read terminal_at: %v", err)
	}
	if !terminalAfterDeadline {
		t.Error("terminal_at predates expires_at")
	}
	lifecycleRequireInvariants(t, database)
}
