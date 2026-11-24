package backend_test

import (
	"context"
	"sync"
	"testing"

	"cloudian/cloudian-scale/internal/module/order"
)

// TEST-13: with no payment event an order stays pending until its deadline, then the worker expires it.
func TestOrderExpiryWaitsForDeadline(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	customer := lifecycleSeedUser(t, database, "expiry-wait@example.com")
	orderID, skuID := orderAPISeedOrder(t, database, customer, 300, 2)

	stats, err := order.ExpireDue(context.Background(), database, 10)
	if err != nil || stats.Scanned != 0 || stats.Expired != 0 || stats.Remaining != 0 {
		t.Fatalf("before the deadline stats = %+v, err = %v; want nothing due", stats, err)
	}
	if got := orderAPIStatus(t, database, orderID); got != "pending_payment" {
		t.Fatalf("status before deadline = %s, want pending_payment (never failed without a payment event)", got)
	}

	cartExec(t, database, `UPDATE orders SET expires_at = now() - interval '3 seconds',
		created_at = now() - interval '303 seconds' WHERE id = ?`, orderID)
	stats, err = order.ExpireDue(context.Background(), database, 10)
	if err != nil || stats.Scanned != 1 || stats.Expired != 1 || stats.Errors != 0 || stats.LostRace != 0 || stats.Remaining != 0 {
		t.Fatalf("after the deadline stats = %+v, err = %v; want one expired", stats, err)
	}
	if stats.MaxLatenessSeconds < 3 || stats.MaxLatenessSeconds > 60 {
		t.Errorf("max lateness = %v, want about 3 seconds", stats.MaxLatenessSeconds)
	}
	if got := orderAPIStatus(t, database, orderID); got != "expired" {
		t.Errorf("status after deadline = %s, want expired", got)
	}
	if reserved := orderAPIReserved(t, database, skuID); reserved != 0 {
		t.Errorf("reserved = %d, want released", reserved)
	}
	again, err := order.ExpireDue(context.Background(), database, 10)
	if err != nil || again.Scanned != 0 || again.Expired != 0 {
		t.Errorf("second pass stats = %+v, err = %v; want a no-op", again, err)
	}
	lifecycleRequireInvariants(t, database)
}

// TEST-19: concurrent workers expire every due order exactly once and a restart finishes the backlog.
func TestOrderExpiryConcurrentWorkersAndRestart(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	customer := lifecycleSeedUser(t, database, "expiry-bulk@example.com")
	const due = 36

	skuID := cartSeedSKU(t, database, cartSKUOptions{OnHand: 500})
	saleSKU := cartSeedSKU(t, database, cartSKUOptions{OnHand: 500})
	sale := cartSeedSale(t, database, saleSKU, cartSaleOptions{QuotaTotal: 100, PurchaseLimit: 100})
	var dueIDs []string
	for range due {
		dueIDs = append(dueIDs, lifecycleSeedPendingOrder(t, database, customer, -10,
			lifecycleLine{SKUID: skuID, Quantity: 2, UnitPrice: 10_000},
			lifecycleLine{SKUID: saleSKU, Quantity: 1, UnitPrice: 40_000, FlashSaleItemID: sale},
		))
	}
	var liveIDs []string
	for range 4 {
		liveIDs = append(liveIDs, lifecycleSeedPendingOrder(t, database, customer, 300,
			lifecycleLine{SKUID: skuID, Quantity: 1, UnitPrice: 10_000}))
	}
	if reserved := orderAPIReserved(t, database, skuID); reserved != due*2+4 {
		t.Fatalf("seeded reserved = %d, want %d", reserved, due*2+4)
	}

	results := make([]order.ExpiryStats, 3)
	errs := make([]error, 3)
	var wg sync.WaitGroup
	for i := range results {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i], errs[i] = order.ExpireDue(context.Background(), database, 100)
		}()
	}
	wg.Wait()

	expired := 0
	for i, stats := range results {
		if errs[i] != nil || stats.Errors != 0 {
			t.Errorf("worker %d stats = %+v, err = %v; want no errors", i, stats, errs[i])
		}
		expired += stats.Expired
	}
	if expired != due {
		t.Errorf("sum of expired = %d, want exactly %d", expired, due)
	}
	for _, id := range dueIDs {
		if got := orderAPIStatus(t, database, id); got != "expired" {
			t.Errorf("due order %s = %s, want expired", id, got)
		}
	}
	for _, id := range liveIDs {
		if got := orderAPIStatus(t, database, id); got != "pending_payment" {
			t.Errorf("non-due order %s = %s, want untouched", id, got)
		}
	}
	if reserved := orderAPIReserved(t, database, skuID); reserved != 4 {
		t.Errorf("reserved = %d, want only the 4 live holds left", reserved)
	}
	if counters := lifecycleReadCounters(t, database, saleSKU, sale, customer); counters.QuotaReserved != 0 ||
		counters.UsageReserved != 0 || counters.QuotaSold != 0 || counters.Reserved != 0 {
		t.Errorf("flash sale counters = %+v, want all holds released and nothing sold", counters)
	}
	lifecycleRequireInvariants(t, database)
}

func TestOrderExpiryBatchBacklogResumesAfterRestart(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	customer := lifecycleSeedUser(t, database, "expiry-backlog@example.com")
	skuID := cartSeedSKU(t, database, cartSKUOptions{OnHand: 100})
	for range 12 {
		lifecycleSeedPendingOrder(t, database, customer, -10, lifecycleLine{SKUID: skuID, Quantity: 1, UnitPrice: 10_000})
	}

	first, err := order.ExpireDue(context.Background(), database, 5)
	if err != nil || first.Scanned != 5 || first.Expired != 5 || first.Remaining != 7 {
		t.Fatalf("first run = %+v, err = %v; want 5 expired and 7 remaining", first, err)
	}
	second, err := order.ExpireDue(context.Background(), database, 100)
	if err != nil || second.Expired != 7 || second.Remaining != 0 {
		t.Fatalf("restart run = %+v, err = %v; want the remaining 7 expired", second, err)
	}
	if reserved := orderAPIReserved(t, database, skuID); reserved != 0 {
		t.Errorf("reserved = %d, want 0", reserved)
	}
	if _, err := order.ExpireDue(context.Background(), database, 0); err == nil {
		t.Error("batch size 0 accepted, want an error")
	}
	lifecycleRequireInvariants(t, database)
}

func TestOrderExpiryReportsCorruptOrderWithoutBlockingOthers(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	customer := lifecycleSeedUser(t, database, "expiry-corrupt@example.com")
	corruptID, _ := orderAPISeedOrder(t, database, customer, -20, 1) // oldest deadline, processed first
	// Simulate corruption: the order lost its reservation and the stock counter was reconciled by hand.
	cartExec(t, database, `UPDATE inventories SET reserved = 0 WHERE sku_id IN (
		SELECT sku_id FROM order_items WHERE order_id = ?)`, corruptID)
	cartExec(t, database, `DELETE FROM inventory_reservations WHERE order_item_id IN (
		SELECT id FROM order_items WHERE order_id = ?)`, corruptID)
	healthyID, skuID := orderAPISeedOrder(t, database, customer, -10, 3)

	stats, err := order.ExpireDue(context.Background(), database, 10)
	if err != nil {
		t.Fatalf("ExpireDue returned %v, want the batch to survive one bad order", err)
	}
	if stats.Scanned != 2 || stats.Errors != 1 || stats.Expired != 1 || stats.Remaining != 1 {
		t.Errorf("stats = %+v, want scanned 2, errors 1, expired 1, remaining 1", stats)
	}
	if got := orderAPIStatus(t, database, corruptID); got != "pending_payment" {
		t.Errorf("corrupt order = %s, want untouched (its transaction rolled back)", got)
	}
	if got := orderAPIStatus(t, database, healthyID); got != "expired" {
		t.Errorf("healthy order = %s, want expired", got)
	}
	if reserved := orderAPIReserved(t, database, skuID); reserved != 0 {
		t.Errorf("reserved = %d, want released", reserved)
	}
}
