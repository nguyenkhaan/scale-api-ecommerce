package backend_test

import (
	"testing"
)

func TestCatalogMigrationEnforcesStatusPriceAndCaseSensitiveSKUUniqueness(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	database := testDatabase.ORM()

	categoryID := "51000000-0000-4000-8000-000000000001"
	productID := "52000000-0000-4000-8000-000000000001"
	if err := database.Exec(
		"INSERT INTO categories (id, name, slug) VALUES (?, ?, ?)",
		categoryID, "Giày", "giay",
	).Error; err != nil {
		t.Fatalf("insert category: %v", err)
	}
	if err := database.Exec(
		"INSERT INTO products (id, category_id, name) VALUES (?, ?, ?)",
		productID, categoryID, "Giày chạy bộ",
	).Error; err != nil {
		t.Fatalf("insert product: %v", err)
	}

	for index, code := range []string{"ABC-01", "abc-01"} {
		if err := database.Exec(
			`INSERT INTO skus (id, product_id, code, regular_price, max_quantity_per_order)
			 VALUES (?, ?, ?, ?, ?)`,
			[]string{
				"53000000-0000-4000-8000-000000000001",
				"53000000-0000-4000-8000-000000000002",
			}[index],
			productID, code, 100_000, 10,
		).Error; err != nil {
			t.Fatalf("insert case-distinct SKU %q: %v", code, err)
		}
	}

	t.Run("exact SKU code is unique", func(t *testing.T) {
		err := database.Exec(
			`INSERT INTO skus (id, product_id, code, regular_price, max_quantity_per_order)
			 VALUES (?, ?, ?, ?, ?)`,
			"53000000-0000-4000-8000-000000000003", productID, "ABC-01", 100_000, 10,
		).Error
		if err == nil {
			t.Fatal("insert exact duplicate SKU code succeeded")
		}
	})

	t.Run("regular price must be positive", func(t *testing.T) {
		err := database.Exec(
			`INSERT INTO skus (id, product_id, code, regular_price, max_quantity_per_order)
			 VALUES (?, ?, ?, ?, ?)`,
			"53000000-0000-4000-8000-000000000004", productID, "ZERO-PRICE", 0, 10,
		).Error
		if err == nil {
			t.Fatal("insert SKU with zero price succeeded")
		}
	})

	t.Run("maximum quantity must be positive", func(t *testing.T) {
		err := database.Exec(
			`INSERT INTO skus (id, product_id, code, regular_price, max_quantity_per_order)
			 VALUES (?, ?, ?, ?, ?)`,
			"53000000-0000-4000-8000-000000000005", productID, "ZERO-LIMIT", 100_000, 0,
		).Error
		if err == nil {
			t.Fatal("insert SKU with zero maximum quantity succeeded")
		}
	})

	t.Run("catalog status must be active or inactive", func(t *testing.T) {
		err := database.Exec(
			"UPDATE products SET status = ? WHERE id = ?",
			"deleted", productID,
		).Error
		if err == nil {
			t.Fatal("update product to unsupported status succeeded")
		}
	})

	t.Run("SKU status must be active or inactive", func(t *testing.T) {
		err := database.Exec(
			"UPDATE skus SET status = ? WHERE id = ?",
			"deleted", "53000000-0000-4000-8000-000000000001",
		).Error
		if err == nil {
			t.Fatal("update SKU to unsupported status succeeded")
		}
	})

	t.Run("referenced product cannot be hard deleted", func(t *testing.T) {
		err := database.Exec("DELETE FROM products WHERE id = ?", productID).Error
		if err == nil {
			t.Fatal("delete product referenced by SKUs succeeded")
		}
	})
}

func TestInventoryMigrationEnforcesCountersAndLedger(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	database := testDatabase.ORM()

	actorID := "54000000-0000-4000-8000-000000000001"
	productID := "52000000-0000-4000-8000-000000000011"
	skuID := "53000000-0000-4000-8000-000000000011"
	if err := database.Exec(
		`INSERT INTO users (id, email, email_normalized, password_hash, role)
		 VALUES (?, ?, ?, ?, ?)`,
		actorID, "inventory-admin@example.com", "inventory-admin@example.com", "test-hash", "admin",
	).Error; err != nil {
		t.Fatalf("insert actor: %v", err)
	}
	if err := database.Exec(
		"INSERT INTO products (id, name) VALUES (?, ?)",
		productID, "Áo khoác",
	).Error; err != nil {
		t.Fatalf("insert product: %v", err)
	}
	if err := database.Exec(
		`INSERT INTO skus (id, product_id, code, regular_price, max_quantity_per_order)
		 VALUES (?, ?, ?, ?, ?)`,
		skuID, productID, "JACKET-BLACK-M", 350_000, 5,
	).Error; err != nil {
		t.Fatalf("insert SKU: %v", err)
	}
	if err := database.Exec(
		"INSERT INTO inventories (sku_id, on_hand, reserved) VALUES (?, ?, ?)",
		skuID, 5, 3,
	).Error; err != nil {
		t.Fatalf("insert inventory: %v", err)
	}

	t.Run("on hand cannot fall below reserved", func(t *testing.T) {
		err := database.Exec(
			"UPDATE inventories SET on_hand = ? WHERE sku_id = ?",
			2, skuID,
		).Error
		if err == nil {
			t.Fatal("update on_hand below reserved succeeded")
		}
	})

	t.Run("inventory counters and version cannot be negative", func(t *testing.T) {
		testCases := []struct {
			name   string
			column string
		}{
			{name: "on hand", column: "on_hand"},
			{name: "reserved", column: "reserved"},
			{name: "version", column: "version"},
		}
		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				err := database.Exec(
					"UPDATE inventories SET "+testCase.column+" = ? WHERE sku_id = ?",
					-1, skuID,
				).Error
				if err == nil {
					t.Fatalf("update %s to a negative value succeeded", testCase.name)
				}
			})
		}
	})

	t.Run("ledger before plus delta must equal after", func(t *testing.T) {
		err := database.Exec(
			`INSERT INTO inventory_adjustments
			 (sku_id, actor_user_id, delta, on_hand_before, on_hand_after, reason)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			skuID, actorID, 2, 5, 8, "Restock",
		).Error
		if err == nil {
			t.Fatal("insert inconsistent adjustment succeeded")
		}
	})

	t.Run("ledger delta cannot be zero", func(t *testing.T) {
		err := database.Exec(
			`INSERT INTO inventory_adjustments
			 (sku_id, actor_user_id, delta, on_hand_before, on_hand_after, reason)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			skuID, actorID, 0, 5, 5, "No change",
		).Error
		if err == nil {
			t.Fatal("insert zero-delta adjustment succeeded")
		}
	})

	t.Run("ledger reason must contain a visible character", func(t *testing.T) {
		for _, reason := range []string{"   ", "\t", "\n", " \t\n "} {
			err := database.Exec(
				`INSERT INTO inventory_adjustments
				 (sku_id, actor_user_id, delta, on_hand_before, on_hand_after, reason)
				 VALUES (?, ?, ?, ?, ?, ?)`,
				skuID, actorID, 2, 5, 7, reason,
			).Error
			if err == nil {
				t.Fatalf("insert whitespace-only adjustment reason %q succeeded", reason)
			}
		}
	})

	if err := database.Exec(
		`INSERT INTO inventory_adjustments
		 (sku_id, actor_user_id, delta, on_hand_before, on_hand_after, reason)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		skuID, actorID, 2, 5, 7, "Restock",
	).Error; err != nil {
		t.Fatalf("insert valid adjustment: %v", err)
	}
}
