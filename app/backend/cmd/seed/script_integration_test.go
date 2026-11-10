//go:build integration

package main

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/database"
	"cloudian/cloudian-scale/internal/migration"
	"golang.org/x/crypto/bcrypt"
)

func TestSeedRestoresFixtureAfterDevelopmentChanges(t *testing.T) {
	ctx, database := openSeedIntegrationDatabase(t)
	fixture := loadSeedFixture(t)
	if err := seed(ctx, database.ORM(), fixture); err != nil {
		t.Fatalf("first seed: %v", err)
	}
	if err := seed(ctx, database.ORM(), fixture); err != nil {
		t.Fatalf("second seed: %v", err)
	}

	assertSeedCounts(t, ctx, database)
	assertSeedCredentialsAndSessions(t, ctx, database)

	const replacementCartItemID = "19000000-0000-4000-8000-000000000099"
	if _, err := database.SQL().ExecContext(ctx, `DELETE FROM cart_items WHERE id = ?`, fixture.CartItems[0].ID); err != nil {
		t.Fatalf("remove fixture cart item: %v", err)
	}
	if _, err := database.SQL().ExecContext(ctx, `
		INSERT INTO cart_items (id, cart_id, sku_id, quantity, intended_price_source, version)
		VALUES (?, ?, ?, 3, 'regular', 1)`, replacementCartItemID, fixture.CartItems[0].CartID, fixture.CartItems[0].SKUID); err != nil {
		t.Fatalf("replace fixture cart item: %v", err)
	}
	if _, err := database.SQL().ExecContext(ctx, `UPDATE carts SET status = 'archived' WHERE id = ?`, fixture.Carts[0].ID); err != nil {
		t.Fatalf("archive fixture active cart: %v", err)
	}
	if _, err := database.SQL().ExecContext(ctx, `
		INSERT INTO carts (id, user_id, status) VALUES ('18000000-0000-4000-8000-000000000099', ?, 'active')`, fixture.Carts[0].UserID); err != nil {
		t.Fatalf("create replacement active cart: %v", err)
	}

	if _, err := database.SQL().ExecContext(ctx, `
		UPDATE inventories SET on_hand = on_hand + 5, version = version + 1 WHERE sku_id = ?`, fixture.Inventories[0].SKUID); err != nil {
		t.Fatalf("adjust inventory: %v", err)
	}
	if _, err := database.SQL().ExecContext(ctx, `
		INSERT INTO inventory_adjustments (id, sku_id, actor_user_id, delta, on_hand_before, on_hand_after, reason)
		VALUES ('15000000-0000-4000-8000-000000000099', ?, ?, 5, 40, 45, 'MANUAL_TEST')`, fixture.Inventories[0].SKUID, fixture.Users[0].ID); err != nil {
		t.Fatalf("record inventory adjustment: %v", err)
	}

	if err := seed(ctx, database.ORM(), fixture); err != nil {
		t.Fatalf("reseed after development changes: %v", err)
	}

	var cartItemID, activeCartID, replacementCartStatus string
	if err := database.SQL().QueryRowContext(ctx, `SELECT id FROM cart_items WHERE cart_id = ? AND sku_id = ?`, fixture.CartItems[0].CartID, fixture.CartItems[0].SKUID).Scan(&cartItemID); err != nil {
		t.Fatalf("load restored cart item: %v", err)
	}
	if cartItemID != fixture.CartItems[0].ID {
		t.Errorf("restored cart item id = %q, want %q", cartItemID, fixture.CartItems[0].ID)
	}
	if err := database.SQL().QueryRowContext(ctx, `SELECT id FROM carts WHERE user_id = ? AND status = 'active'`, fixture.Carts[0].UserID).Scan(&activeCartID); err != nil {
		t.Fatalf("load restored active cart: %v", err)
	}
	if activeCartID != fixture.Carts[0].ID {
		t.Errorf("active cart id = %q, want %q", activeCartID, fixture.Carts[0].ID)
	}
	if err := database.SQL().QueryRowContext(ctx, `SELECT status FROM carts WHERE id = '18000000-0000-4000-8000-000000000099'`).Scan(&replacementCartStatus); err != nil {
		t.Fatalf("load archived replacement cart: %v", err)
	}
	if replacementCartStatus != "archived" {
		t.Errorf("replacement cart status = %q, want archived", replacementCartStatus)
	}

	var onHand, ledgerTotal int64
	if err := database.SQL().QueryRowContext(ctx, `SELECT on_hand FROM inventories WHERE sku_id = ?`, fixture.Inventories[0].SKUID).Scan(&onHand); err != nil {
		t.Fatalf("load restored inventory: %v", err)
	}
	if err := database.SQL().QueryRowContext(ctx, `SELECT COALESCE(SUM(delta), 0) FROM inventory_adjustments WHERE sku_id = ?`, fixture.Inventories[0].SKUID).Scan(&ledgerTotal); err != nil {
		t.Fatalf("sum inventory ledger: %v", err)
	}
	if onHand != fixture.Inventories[0].OnHand || ledgerTotal != onHand {
		t.Errorf("inventory on_hand/ledger = %d/%d, want %d/%d", onHand, ledgerTotal, fixture.Inventories[0].OnHand, fixture.Inventories[0].OnHand)
	}
}

func TestSeedRollsBackOnLateWriteFailure(t *testing.T) {
	ctx, database := openSeedIntegrationDatabase(t)
	fixture := loadSeedFixture(t)
	if _, err := database.SQL().ExecContext(ctx, `
		INSERT INTO categories (id, name, slug, is_visible)
		VALUES ('12000000-0000-4000-8000-000000000099', 'Conflicting category', ?, true)`, fixture.Categories[0].Slug); err != nil {
		t.Fatalf("insert late-write conflict: %v", err)
	}

	if err := seed(ctx, database.ORM(), fixture); err == nil {
		t.Fatal("seed() succeeded despite a late category conflict")
	}

	var users int
	if err := database.SQL().QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE email_normalized IN ('admin@gmail.com', 'customer@gmail.com')`).Scan(&users); err != nil {
		t.Fatalf("count rolled-back users: %v", err)
	}
	if users != 0 {
		t.Errorf("users persisted after failed transaction = %d, want 0", users)
	}
}

func openSeedIntegrationDatabase(t *testing.T) (context.Context, *database.Database) {
	t.Helper()
	connectionString := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if connectionString == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL seed integration tests")
	}
	if connectionString == strings.TrimSpace(os.Getenv("DATABASE_URL")) {
		t.Fatal("TEST_DATABASE_URL must be different from DATABASE_URL because this test truncates it")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	database, err := database.Open(ctx, connectionString)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	if err := migration.Run(ctx, database.SQL(), migration.CommandUp); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	if _, err := database.SQL().ExecContext(ctx, `
		TRUNCATE cart_items, carts, flash_sale_customer_usage, flash_sale_items,
		flash_sale_campaigns, inventory_adjustments, inventories, skus, products,
		categories, login_throttles, sessions, users CASCADE`); err != nil {
		t.Fatalf("truncate test database: %v", err)
	}
	return ctx, database
}

func loadSeedFixture(t *testing.T) fixture {
	t.Helper()
	fixture, err := loadFixture()
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	return fixture
}

func assertSeedCounts(t *testing.T, ctx context.Context, database *database.Database) {
	t.Helper()
	for table, want := range map[string]int{
		"users": 2, "sessions": 4, "login_throttles": 4, "categories": 4,
		"products": 4, "skus": 4, "inventories": 4, "inventory_adjustments": 4,
		"flash_sale_campaigns": 4, "flash_sale_items": 4, "flash_sale_customer_usage": 4,
		"carts": 4, "cart_items": 4,
	} {
		var got int
		if err := database.SQL().QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&got); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if got != want {
			t.Errorf("%s count = %d, want %d", table, got, want)
		}
	}
}

func assertSeedCredentialsAndSessions(t *testing.T, ctx context.Context, database *database.Database) {
	t.Helper()
	var passwordHash string
	if err := database.SQL().QueryRowContext(ctx, `SELECT password_hash FROM users WHERE email_normalized = 'admin@gmail.com'`).Scan(&passwordHash); err != nil {
		t.Fatalf("load admin password hash: %v", err)
	}
	if strings.Contains(passwordHash, "cloudian") || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("cloudian")) != nil {
		t.Error("admin password is not stored as a valid bcrypt hash")
	}

	var invalidSessions int
	if err := database.SQL().QueryRowContext(ctx, `
		SELECT COUNT(*) FROM sessions
		WHERE revoked_at IS NULL OR octet_length(token_hash) <> 32 OR octet_length(csrf_secret_hash) <> 32`).Scan(&invalidSessions); err != nil {
		t.Fatalf("verify seeded sessions: %v", err)
	}
	if invalidSessions != 0 {
		t.Errorf("invalid seeded sessions = %d, want 0", invalidSessions)
	}
}
