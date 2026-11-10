package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/database"
	"cloudian/cloudian-scale/internal/migration"
	"golang.org/x/crypto/bcrypt"
)

func TestLoadFixtureContainsCompleteDevelopmentDataset(t *testing.T) {
	fixture, err := loadFixture()
	if err != nil {
		t.Fatalf("loadFixture() error = %v", err)
	}

	if len(fixture.Users) != 2 {
		t.Fatalf("users = %d, want 2", len(fixture.Users))
	}
	usersByEmail := make(map[string]userFixture, len(fixture.Users))
	for _, user := range fixture.Users {
		usersByEmail[user.Email] = user
	}
	for email, role := range map[string]string{
		"admin@gmail.com":    "admin",
		"customer@gmail.com": "customer",
	} {
		user, ok := usersByEmail[email]
		if !ok {
			t.Fatalf("fixture user %q is missing", email)
		}
		if user.Role != role || user.Password != "cloudian" {
			t.Errorf("fixture user %q = %+v, want role %q and demo password", email, user, role)
		}
	}

	for name, count := range map[string]int{
		"sessions":                  len(fixture.Sessions),
		"login_throttles":           len(fixture.LoginThrottles),
		"categories":                len(fixture.Categories),
		"products":                  len(fixture.Products),
		"skus":                      len(fixture.SKUs),
		"inventories":               len(fixture.Inventories),
		"inventory_adjustments":     len(fixture.InventoryAdjustments),
		"flash_sale_campaigns":      len(fixture.FlashSaleCampaigns),
		"flash_sale_items":          len(fixture.FlashSaleItems),
		"flash_sale_customer_usage": len(fixture.FlashSaleCustomerUsage),
		"carts":                     len(fixture.Carts),
		"cart_items":                len(fixture.CartItems),
	} {
		if count != 4 {
			t.Errorf("%s count = %d, want 4", name, count)
		}
	}
}

func TestValidateFixtureAcceptsEmbeddedData(t *testing.T) {
	fixture, err := loadFixture()
	if err != nil {
		t.Fatalf("loadFixture() error = %v", err)
	}
	if err := validateFixture(fixture); err != nil {
		t.Fatalf("validateFixture() error = %v", err)
	}
}

func TestCanSeedRequiresExplicitDevelopmentConfirmation(t *testing.T) {
	for _, testCase := range []struct {
		name         string
		environment  string
		confirmation string
		want         bool
	}{
		{name: "development", environment: "development", confirmation: "true", want: true},
		{name: "dev", environment: "dev", confirmation: "true", want: true},
		{name: "test", environment: "test", confirmation: "true", want: true},
		{name: "production", environment: "production", confirmation: "true", want: false},
		{name: "missing confirmation", environment: "dev", confirmation: "", want: false},
		{name: "wrong confirmation", environment: "dev", confirmation: "yes", want: false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if got := canSeed(testCase.environment, testCase.confirmation); got != testCase.want {
				t.Errorf("canSeed(%q, %q) = %t, want %t", testCase.environment, testCase.confirmation, got, testCase.want)
			}
		})
	}
}

func TestHashFixturePasswordUsesBcrypt(t *testing.T) {
	hash, err := hashFixturePassword("cloudian")
	if err != nil {
		t.Fatalf("hashFixturePassword() error = %v", err)
	}
	if strings.Contains(hash, "cloudian") {
		t.Fatal("bcrypt hash contains the fixture password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("cloudian")); err != nil {
		t.Fatalf("bcrypt hash does not verify the fixture password: %v", err)
	}
}

func TestSeedWritesRepeatableFixture(t *testing.T) {
	connectionString := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if connectionString == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL seed integration tests")
	}
	if connectionString == strings.TrimSpace(os.Getenv("DATABASE_URL")) {
		t.Fatal("TEST_DATABASE_URL must be different from DATABASE_URL because this test truncates it")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	database, err := database.Open(ctx, connectionString)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	defer database.Close()
	if err := migration.Run(ctx, database.SQL(), migration.CommandUp); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	if _, err := database.SQL().ExecContext(ctx, `
		TRUNCATE cart_items, carts, flash_sale_customer_usage, flash_sale_items,
		flash_sale_campaigns, inventory_adjustments, inventories, skus, products,
		categories, login_throttles, sessions, users CASCADE`); err != nil {
		t.Fatalf("truncate test database: %v", err)
	}

	fixture, err := loadFixture()
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	if err := seed(ctx, database.ORM(), fixture); err != nil {
		t.Fatalf("first seed: %v", err)
	}
	if err := seed(ctx, database.ORM(), fixture); err != nil {
		t.Fatalf("second seed: %v", err)
	}

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

func TestLegacySeedCommandsAreRemoved(t *testing.T) {
	for _, path := range []string{
		"../seed-admin",
		"../seed-catalog",
		"../../internal/bootstrap/admin_seed.go",
		"../../internal/bootstrap/catalog_seed.go",
	} {
		if _, err := os.Stat(filepath.Clean(path)); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("legacy seed path %q still exists", path)
		}
	}
}
