package backend_test

import (
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestFlashSaleModelsMapToExpectedTablesAndColumns(t *testing.T) {
	testCases := []struct {
		name        string
		model       any
		table       string
		primaryKeys []string
		columns     []string
		autoID      bool
	}{
		{
			name:        "campaign",
			model:       &FlashSaleCampaign{},
			table:       "flash_sale_campaigns",
			primaryKeys: []string{"id"},
			columns:     []string{"id", "name", "status", "starts_at", "ends_at", "created_at", "updated_at"},
			autoID:      true,
		},
		{
			name:        "item",
			model:       &FlashSaleItem{},
			table:       "flash_sale_items",
			primaryKeys: []string{"id"},
			autoID:      true,
			columns: []string{
				"id", "campaign_id", "sku_id", "sale_price", "quota_total", "quota_reserved",
				"quota_sold", "purchase_limit", "created_at", "updated_at",
			},
		},
		{
			name:        "customer usage",
			model:       &FlashSaleCustomerUsage{},
			table:       "flash_sale_customer_usage",
			primaryKeys: []string{"flash_sale_item_id", "user_id"},
			columns: []string{
				"flash_sale_item_id", "user_id", "reserved_quantity", "sold_quantity", "updated_at",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			parsed, err := schema.Parse(testCase.model, &sync.Map{}, schema.NamingStrategy{})
			if err != nil {
				t.Fatalf("parse GORM schema: %v", err)
			}
			if parsed.Table != testCase.table {
				t.Errorf("table = %q, want %q", parsed.Table, testCase.table)
			}

			for _, column := range testCase.columns {
				if parsed.LookUpField(column) == nil {
					t.Errorf("column %q is not mapped", column)
				}
			}
			if len(parsed.PrimaryFields) != len(testCase.primaryKeys) {
				t.Fatalf("primary key count = %d, want %d", len(parsed.PrimaryFields), len(testCase.primaryKeys))
			}
			for index, column := range testCase.primaryKeys {
				if parsed.PrimaryFields[index].DBName != column {
					t.Errorf("primary key %d = %q, want %q", index, parsed.PrimaryFields[index].DBName, column)
				}
			}
			if testCase.autoID {
				id := parsed.LookUpField("id")
				if id == nil || !id.HasDefaultValue || id.DefaultValue != "gen_random_uuid()" {
					t.Errorf("id default = %q, want gen_random_uuid()", id.DefaultValue)
				}
			}
		})
	}
}
