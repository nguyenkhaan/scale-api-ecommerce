package backend_test

import (
	"testing"
	"time"
)

func TestFlashSaleMigrationGeneratesCampaignAndItemIDs(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	database := testDatabase.ORM()
	startsAt := time.Date(2026, time.October, 30, 10, 0, 0, 0, time.UTC)

	var campaignID string
	if err := database.Raw(
		`INSERT INTO flash_sale_campaigns (name, starts_at, ends_at)
		 VALUES (?, ?, ?)
		 RETURNING id`,
		"Generated IDs", startsAt, startsAt.Add(time.Hour),
	).Scan(&campaignID).Error; err != nil {
		t.Fatalf("insert campaign without id: %v", err)
	}
	if campaignID == "" {
		t.Fatal("generated campaign id is empty")
	}

	var itemID string
	if err := database.Raw(
		`INSERT INTO flash_sale_items
			(campaign_id, sku_id, sale_price, quota_total, purchase_limit)
		 VALUES (?, ?, ?, ?, ?)
		 RETURNING id`,
		campaignID, "30000000-0000-4000-8000-000000000099", 90_000, 10, 2,
	).Scan(&itemID).Error; err != nil {
		t.Fatalf("insert item without id: %v", err)
	}
	if itemID == "" {
		t.Fatal("generated item id is empty")
	}
}

func TestFlashSaleMigrationEnforcesCampaignAndItemConstraints(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	database := testDatabase.ORM()

	campaignID := "10000000-0000-4000-8000-000000000001"
	itemID := "20000000-0000-4000-8000-000000000001"
	skuID := "30000000-0000-4000-8000-000000000001"
	startsAt := time.Date(2026, time.October, 30, 10, 0, 0, 0, time.UTC)
	endsAt := startsAt.Add(time.Hour)

	if err := database.Exec(
		"INSERT INTO flash_sale_campaigns (id, name, starts_at, ends_at) VALUES (?, ?, ?, ?)",
		campaignID, "Morning sale", startsAt, endsAt,
	).Error; err != nil {
		t.Fatalf("insert valid campaign: %v", err)
	}
	if err := database.Exec(
		`INSERT INTO flash_sale_items
			(id, campaign_id, sku_id, sale_price, quota_total, purchase_limit)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		itemID, campaignID, skuID, 90_000, 10, 2,
	).Error; err != nil {
		t.Fatalf("insert valid flash sale item: %v", err)
	}

	t.Run("campaign end must follow start", func(t *testing.T) {
		err := database.Exec(
			"INSERT INTO flash_sale_campaigns (id, name, starts_at, ends_at) VALUES (?, ?, ?, ?)",
			"10000000-0000-4000-8000-000000000002", "Invalid sale", startsAt, startsAt,
		).Error
		if err == nil {
			t.Fatal("insert campaign with ends_at = starts_at succeeded")
		}
	})

	t.Run("item values must be positive", func(t *testing.T) {
		testCases := []struct {
			name          string
			salePrice     int64
			quotaTotal    int64
			purchaseLimit int
		}{
			{name: "sale price", salePrice: 0, quotaTotal: 10, purchaseLimit: 2},
			{name: "quota", salePrice: 90_000, quotaTotal: 0, purchaseLimit: 2},
			{name: "purchase limit", salePrice: 90_000, quotaTotal: 10, purchaseLimit: 0},
		}
		for index, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				err := database.Exec(
					`INSERT INTO flash_sale_items
						(id, campaign_id, sku_id, sale_price, quota_total, purchase_limit)
					 VALUES (?, ?, ?, ?, ?, ?)`,
					[]string{
						"20000000-0000-4000-8000-000000000002",
						"20000000-0000-4000-8000-000000000003",
						"20000000-0000-4000-8000-000000000004",
					}[index],
					campaignID,
					[]string{
						"30000000-0000-4000-8000-000000000002",
						"30000000-0000-4000-8000-000000000003",
						"30000000-0000-4000-8000-000000000004",
					}[index],
					testCase.salePrice, testCase.quotaTotal, testCase.purchaseLimit,
				).Error
				if err == nil {
					t.Fatalf("insert with invalid %s succeeded", testCase.name)
				}
			})
		}
	})

	t.Run("campaign cannot contain duplicate SKU", func(t *testing.T) {
		err := database.Exec(
			`INSERT INTO flash_sale_items
				(id, campaign_id, sku_id, sale_price, quota_total, purchase_limit)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			"20000000-0000-4000-8000-000000000005", campaignID, skuID, 80_000, 5, 1,
		).Error
		if err == nil {
			t.Fatal("insert duplicate campaign/SKU succeeded")
		}
	})

	t.Run("quota counters stay non-negative and within total", func(t *testing.T) {
		testCases := []struct {
			name     string
			reserved int64
			sold     int64
		}{
			{name: "negative reserved", reserved: -1},
			{name: "negative sold", sold: -1},
			{name: "over quota", reserved: 6, sold: 5},
		}
		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				err := database.Exec(
					"UPDATE flash_sale_items SET quota_reserved = ?, quota_sold = ? WHERE id = ?",
					testCase.reserved, testCase.sold, itemID,
				).Error
				if err == nil {
					t.Fatalf("update with %s succeeded", testCase.name)
				}
			})
		}
	})
}

func TestFlashSaleMigrationEnforcesCustomerUsageConstraints(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	database := testDatabase.ORM()

	userID := "40000000-0000-4000-8000-000000000001"
	campaignID := "10000000-0000-4000-8000-000000000011"
	itemID := "20000000-0000-4000-8000-000000000011"
	startsAt := time.Date(2026, time.October, 30, 10, 0, 0, 0, time.UTC)

	if err := database.Exec(
		`INSERT INTO users (id, email, email_normalized, password_hash)
		 VALUES (?, ?, ?, ?)`,
		userID, "flash@example.com", "flash@example.com", "test-hash",
	).Error; err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if err := database.Exec(
		"INSERT INTO flash_sale_campaigns (id, name, starts_at, ends_at) VALUES (?, ?, ?, ?)",
		campaignID, "Usage sale", startsAt, startsAt.Add(time.Hour),
	).Error; err != nil {
		t.Fatalf("insert campaign: %v", err)
	}
	if err := database.Exec(
		`INSERT INTO flash_sale_items
			(id, campaign_id, sku_id, sale_price, quota_total, purchase_limit)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		itemID, campaignID, "30000000-0000-4000-8000-000000000011", 90_000, 10, 2,
	).Error; err != nil {
		t.Fatalf("insert item: %v", err)
	}

	testCases := []struct {
		name     string
		reserved int64
		sold     int64
	}{
		{name: "negative reserved", reserved: -1},
		{name: "negative sold", sold: -1},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := database.Exec(
				`INSERT INTO flash_sale_customer_usage
					(flash_sale_item_id, user_id, reserved_quantity, sold_quantity)
				 VALUES (?, ?, ?, ?)`,
				itemID, userID, testCase.reserved, testCase.sold,
			).Error
			if err == nil {
				t.Fatalf("insert with %s succeeded", testCase.name)
			}
		})
	}
}
