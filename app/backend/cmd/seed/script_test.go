package main

import "testing"

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
