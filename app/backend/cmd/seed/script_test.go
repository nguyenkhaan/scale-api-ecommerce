package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

func TestValidateFixtureRejectsDuplicateSessionIDs(t *testing.T) {
	fixture, err := loadFixture()
	if err != nil {
		t.Fatalf("loadFixture() error = %v", err)
	}
	fixture.Sessions[1].ID = fixture.Sessions[0].ID

	if err := validateFixture(fixture); err == nil {
		t.Fatal("validateFixture() accepted duplicate session IDs")
	}
}

func TestValidateFixtureRejectsUsageThatDisagreesWithItemCounters(t *testing.T) {
	fixture, err := loadFixture()
	if err != nil {
		t.Fatalf("loadFixture() error = %v", err)
	}
	fixture.FlashSaleCustomerUsage[0].SoldQuantity = 1

	if err := validateFixture(fixture); err == nil {
		t.Fatal("validateFixture() accepted usage that disagrees with item counters")
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
