package backend_test

import (
	"testing"
)

func TestCartMigrationEnforcesConstraints(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	database := testDatabase.ORM()

	userID := "41000000-0000-4000-8000-000000000001"
	if err := database.Exec(
		`INSERT INTO users (id, email, email_normalized, password_hash, role, status)
		 VALUES (?, ?, ?, ?, 'customer', 'active')`,
		userID, "cart-user@example.com", "cart-user@example.com", "hash",
	).Error; err != nil {
		t.Fatalf("insert user: %v", err)
	}

	cartID := "42000000-0000-4000-8000-000000000001"
	if err := database.Exec(
		`INSERT INTO carts (id, user_id, status) VALUES (?, ?, 'active')`,
		cartID, userID,
	).Error; err != nil {
		t.Fatalf("insert cart: %v", err)
	}

	t.Run("only one active cart per user", func(t *testing.T) {
		err := database.Exec(
			`INSERT INTO carts (id, user_id, status) VALUES (?, ?, 'active')`,
			"42000000-0000-4000-8000-000000000002", userID,
		).Error
		if err == nil {
			t.Fatal("expected error for duplicate active cart, got nil")
		}
	})

	t.Run("archived carts do not conflict", func(t *testing.T) {
		err := database.Exec(
			`INSERT INTO carts (id, user_id, status) VALUES (?, ?, 'archived')`,
			"42000000-0000-4000-8000-000000000003", userID,
		).Error
		if err != nil {
			t.Fatalf("insert archived cart: %v", err)
		}
	})

	productID := "43000000-0000-4000-8000-000000000001"
	skuID := "44000000-0000-4000-8000-000000000001"
	if err := database.Exec(
		`INSERT INTO products (id, name, status) VALUES (?, 'Cart Product', 'active')`,
		productID,
	).Error; err != nil {
		t.Fatalf("insert product: %v", err)
	}
	if err := database.Exec(
		`INSERT INTO skus (id, product_id, code, regular_price, max_quantity_per_order, status)
		 VALUES (?, ?, 'CART-SKU-1', 100000, 5, 'active')`,
		skuID, productID,
	).Error; err != nil {
		t.Fatalf("insert sku: %v", err)
	}

	itemID := "45000000-0000-4000-8000-000000000001"
	if err := database.Exec(
		`INSERT INTO cart_items (id, cart_id, sku_id, quantity, intended_price_source, version)
		 VALUES (?, ?, ?, 2, 'regular', 1)`,
		itemID, cartID, skuID,
	).Error; err != nil {
		t.Fatalf("insert cart item: %v", err)
	}

	t.Run("cart item unique per sku in cart", func(t *testing.T) {
		err := database.Exec(
			`INSERT INTO cart_items (id, cart_id, sku_id, quantity, intended_price_source, version)
			 VALUES (?, ?, ?, 1, 'regular', 1)`,
			"45000000-0000-4000-8000-000000000002", cartID, skuID,
		).Error
		if err == nil {
			t.Fatal("expected unique constraint violation on (cart_id, sku_id)")
		}
	})

	t.Run("quantity must be positive", func(t *testing.T) {
		skuID2 := "44000000-0000-4000-8000-000000000002"
		if err := database.Exec(
			`INSERT INTO skus (id, product_id, code, regular_price, max_quantity_per_order, status)
			 VALUES (?, ?, 'CART-SKU-2', 100000, 5, 'active')`,
			skuID2, productID,
		).Error; err != nil {
			t.Fatalf("insert sku2: %v", err)
		}
		err := database.Exec(
			`INSERT INTO cart_items (id, cart_id, sku_id, quantity, intended_price_source, version)
			 VALUES (?, ?, ?, 0, 'regular', 1)`,
			"45000000-0000-4000-8000-000000000003", cartID, skuID2,
		).Error
		if err == nil {
			t.Fatal("expected check constraint violation on quantity > 0")
		}
	})

	t.Run("regular price source rejects flash sale item id", func(t *testing.T) {
		campaignID := "46000000-0000-4000-8000-000000000001"
		fsItemID := "47000000-0000-4000-8000-000000000001"
		if err := database.Exec(
			`INSERT INTO flash_sale_campaigns (id, name, starts_at, ends_at, status)
			 VALUES (?, 'Campaign', NOW(), NOW() + INTERVAL '1 hour', 'draft')`,
			campaignID,
		).Error; err != nil {
			t.Fatalf("insert campaign: %v", err)
		}
		if err := database.Exec(
			`INSERT INTO flash_sale_items (id, campaign_id, sku_id, sale_price, quota_total, purchase_limit)
			 VALUES (?, ?, ?, 50000, 10, 2)`,
			fsItemID, campaignID, skuID,
		).Error; err != nil {
			t.Fatalf("insert flash sale item: %v", err)
		}

		skuID3 := "44000000-0000-4000-8000-000000000003"
		if err := database.Exec(
			`INSERT INTO skus (id, product_id, code, regular_price, max_quantity_per_order, status)
			 VALUES (?, ?, 'CART-SKU-3', 100000, 5, 'active')`,
			skuID3, productID,
		).Error; err != nil {
			t.Fatalf("insert sku3: %v", err)
		}

		err := database.Exec(
			`INSERT INTO cart_items (id, cart_id, sku_id, quantity, intended_price_source, flash_sale_item_id, version)
			 VALUES (?, ?, ?, 1, 'regular', ?, 1)`,
			"45000000-0000-4000-8000-000000000004", cartID, skuID3, fsItemID,
		).Error
		if err == nil {
			t.Fatal("expected constraint violation: regular price cannot have flash_sale_item_id")
		}
	})
}
