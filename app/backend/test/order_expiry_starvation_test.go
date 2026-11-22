package backend_test

import (
	"context"
	"testing"

	"cloudian/cloudian-scale/internal/module/order"
)

// A permanently failing order at the head of the queue must not stop healthy orders behind it from expiring.
func TestOrderExpiryNotStarvedByCorruptHead(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	customer := lifecycleSeedUser(t, database, "expiry-starve@example.com")
	corruptID, _ := orderAPISeedOrder(t, database, customer, -30, 1)
	cartExec(t, database, `UPDATE inventories SET reserved = 0 WHERE sku_id IN (
		SELECT sku_id FROM order_items WHERE order_id = ?)`, corruptID)
	cartExec(t, database, `DELETE FROM inventory_reservations WHERE order_item_id IN (
		SELECT id FROM order_items WHERE order_id = ?)`, corruptID)
	healthyID, skuID := orderAPISeedOrder(t, database, customer, -10, 2)

	stats, err := order.ExpireDue(context.Background(), database, 1)
	if err != nil {
		t.Fatalf("ExpireDue: %v", err)
	}
	if stats.Expired != 1 || stats.Errors != 1 || stats.Scanned != 2 || stats.Remaining != 1 {
		t.Errorf("stats = %+v, want expired 1, errors 1, scanned 2, remaining 1 (the corrupt order)", stats)
	}
	if got := orderAPIStatus(t, database, healthyID); got != "expired" {
		t.Errorf("healthy order = %s, want expired despite the corrupt order ahead of it", got)
	}
	if got := orderAPIStatus(t, database, corruptID); got != "pending_payment" {
		t.Errorf("corrupt order = %s, want untouched", got)
	}
	if reserved := orderAPIReserved(t, database, skuID); reserved != 0 {
		t.Errorf("reserved = %d, want released", reserved)
	}
}
