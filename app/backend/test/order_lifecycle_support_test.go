package backend_test

import (
	"fmt"
	"sync/atomic"
	"testing"

	"gorm.io/gorm"
)

var lifecycleSeedCounter atomic.Int64

// lifecycleLine is one order line of a seeded pending order. FlashSaleItemID empty means regular price.
type lifecycleLine struct {
	SKUID           string
	Quantity        int
	UnitPrice       int64
	FlashSaleItemID string
	CartItemID      string
	CartItemVersion int64
}

func lifecycleSeedUser(t *testing.T, database *gorm.DB, email string) string {
	t.Helper()
	var userID string
	if err := database.Raw(
		`INSERT INTO users (id, email, email_normalized, password_hash) VALUES (gen_random_uuid(), ?, lower(?), 'x')
		 RETURNING id`, email, email,
	).Scan(&userID).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return userID
}

// lifecycleSeedPendingOrder commits exactly what a successful checkout leaves behind: the order, item
// snapshots, one active reservation per line and the matching increments of inventories.reserved,
// flash_sale_items.quota_reserved and flash_sale_customer_usage.reserved_quantity.
// expiresInSecs may be negative to create an order that is already past its deadline.
func lifecycleSeedPendingOrder(t *testing.T, database *gorm.DB, userID string, expiresInSecs float64, lines ...lifecycleLine) string {
	t.Helper()
	var total int64
	for _, line := range lines {
		total += int64(line.Quantity) * line.UnitPrice
	}
	var orderID string
	err := database.Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(
			`INSERT INTO orders (user_id, order_number, total_amount, recipient_name, recipient_phone, recipient_address,
				created_at, expires_at)
			 VALUES (?, ?, ?, 'Lan', '0901234567', '1 Vo Van Ngan',
				now() + (? - 300) * interval '1 second', now() + ? * interval '1 second')
			 RETURNING id`,
			userID, fmt.Sprintf("LC-%06d", lifecycleSeedCounter.Add(1)), total, expiresInSecs, expiresInSecs,
		).Scan(&orderID).Error; err != nil {
			return fmt.Errorf("insert order: %w", err)
		}
		for _, line := range lines {
			priceSource, flashItem := "regular", any(nil)
			if line.FlashSaleItemID != "" {
				priceSource, flashItem = "flash_sale", line.FlashSaleItemID
			}
			cartItem, cartVersion := any(nil), any(nil)
			if line.CartItemID != "" {
				cartItem, cartVersion = line.CartItemID, line.CartItemVersion
			}
			var itemID string
			if err := tx.Raw(
				`INSERT INTO order_items (order_id, sku_id, sku_code, product_name, quantity, unit_price, subtotal,
					price_source, flash_sale_item_id, cart_item_id, cart_item_version)
				 SELECT ?, s.id, s.code, p.name, ?, ?, ?, ?, ?, ?, ? FROM skus s JOIN products p ON p.id = s.product_id
				 WHERE s.id = ? RETURNING id`,
				orderID, line.Quantity, line.UnitPrice, int64(line.Quantity)*line.UnitPrice,
				priceSource, flashItem, cartItem, cartVersion, line.SKUID,
			).Scan(&itemID).Error; err != nil {
				return fmt.Errorf("insert order item: %w", err)
			}
			statements := []struct {
				sql  string
				args []any
			}{
				{`INSERT INTO inventory_reservations (order_item_id, sku_id, quantity) VALUES (?, ?, ?)`,
					[]any{itemID, line.SKUID, line.Quantity}},
				{`UPDATE inventories SET reserved = reserved + ? WHERE sku_id = ?`, []any{line.Quantity, line.SKUID}},
			}
			if line.FlashSaleItemID != "" {
				statements = append(statements,
					struct {
						sql  string
						args []any
					}{`UPDATE flash_sale_items SET quota_reserved = quota_reserved + ? WHERE id = ?`,
						[]any{line.Quantity, line.FlashSaleItemID}},
					struct {
						sql  string
						args []any
					}{`INSERT INTO flash_sale_customer_usage (flash_sale_item_id, user_id, reserved_quantity) VALUES (?, ?, ?)
						ON CONFLICT (flash_sale_item_id, user_id)
						DO UPDATE SET reserved_quantity = flash_sale_customer_usage.reserved_quantity + EXCLUDED.reserved_quantity`,
						[]any{line.FlashSaleItemID, userID, line.Quantity}},
				)
			}
			for _, statement := range statements {
				if err := tx.Exec(statement.sql, statement.args...).Error; err != nil {
					return fmt.Errorf("seed reservation effects: %w", err)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed pending order: %v", err)
	}
	return orderID
}

// lifecycleRequireInvariants runs the INV oracles for reservations and counters (INV-02/03/08/09/10/11).
// Every query returns offending rows; any row fails the test.
func lifecycleRequireInvariants(t *testing.T, database *gorm.DB) {
	t.Helper()
	oracles := map[string]string{
		"INV-02 inventories.reserved = active reservations": `
			SELECT i.sku_id FROM inventories i
			LEFT JOIN inventory_reservations r ON r.sku_id = i.sku_id AND r.status = 'active'
			GROUP BY i.sku_id, i.reserved HAVING i.reserved <> COALESCE(SUM(r.quantity), 0)`,
		"INV-03 pending orders hold every line": `
			SELECT o.id FROM orders o JOIN order_items oi ON oi.order_id = o.id
			LEFT JOIN inventory_reservations r ON r.order_item_id = oi.id AND r.status = 'active'
			WHERE o.status = 'pending_payment' GROUP BY o.id HAVING COUNT(*) <> COUNT(r.id)`,
		"INV-08 terminal orders hold nothing": `
			SELECT o.id FROM orders o JOIN order_items oi ON oi.order_id = o.id
			JOIN inventory_reservations r ON r.order_item_id = oi.id
			WHERE (o.status <> 'pending_payment' AND r.status = 'active')
			   OR (o.status = 'confirmed' AND r.status <> 'finalized')
			   OR (o.status IN ('payment_failed', 'cancelled', 'expired') AND r.status <> 'released')`,
		"INV-09 quota_reserved = active flash reservations": `
			SELECT f.id FROM flash_sale_items f
			LEFT JOIN order_items oi ON oi.flash_sale_item_id = f.id
			LEFT JOIN inventory_reservations r ON r.order_item_id = oi.id AND r.status = 'active'
			GROUP BY f.id, f.quota_reserved HAVING f.quota_reserved <> COALESCE(SUM(r.quantity), 0)`,
		"INV-10 usage.reserved_quantity = active flash reservations of that customer": `
			SELECT u.flash_sale_item_id FROM flash_sale_customer_usage u
			LEFT JOIN orders o ON o.user_id = u.user_id
			LEFT JOIN order_items oi ON oi.order_id = o.id AND oi.flash_sale_item_id = u.flash_sale_item_id
			LEFT JOIN inventory_reservations r ON r.order_item_id = oi.id AND r.status = 'active'
			GROUP BY u.flash_sale_item_id, u.user_id, u.reserved_quantity
			HAVING u.reserved_quantity <> COALESCE(SUM(r.quantity), 0)`,
		"INV-11 order total = sum of line snapshots": `
			SELECT o.id FROM orders o JOIN order_items oi ON oi.order_id = o.id
			GROUP BY o.id, o.total_amount HAVING o.total_amount <> SUM(oi.subtotal)`,
	}
	for name, query := range oracles {
		var offending []string
		if err := database.Raw(query).Scan(&offending).Error; err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(offending) > 0 {
			t.Errorf("%s violated by %v", name, offending)
		}
	}
}

type lifecycleCounters struct {
	OnHand, Reserved, QuotaReserved, QuotaSold, UsageReserved, UsageSold int64
}

// lifecycleReadCounters reads the stock, quota and usage counters touched by one flash sale line.
func lifecycleReadCounters(t *testing.T, database *gorm.DB, skuID, flashSaleItemID, userID string) lifecycleCounters {
	t.Helper()
	var counters lifecycleCounters
	if err := database.Raw(
		`SELECT i.on_hand, i.reserved,
			COALESCE(f.quota_reserved, 0) AS quota_reserved, COALESCE(f.quota_sold, 0) AS quota_sold,
			COALESCE(u.reserved_quantity, 0) AS usage_reserved, COALESCE(u.sold_quantity, 0) AS usage_sold
		 FROM inventories i
		 LEFT JOIN flash_sale_items f ON f.id = ?
		 LEFT JOIN flash_sale_customer_usage u ON u.flash_sale_item_id = f.id AND u.user_id = ?
		 WHERE i.sku_id = ?`, nilIfEmpty(flashSaleItemID), userID, skuID,
	).Scan(&counters).Error; err != nil {
		t.Fatalf("read counters: %v", err)
	}
	return counters
}

func nilIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
