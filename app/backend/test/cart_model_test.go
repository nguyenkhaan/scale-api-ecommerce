package backend_test

import (
	"sync"
	"testing"

	"cloudian/cloudian-scale/internal/model"
	"gorm.io/gorm/schema"
)

func TestCartModelsMapToExpectedTablesAndColumns(t *testing.T) {
	testCases := []struct {
		name    string
		model   any
		table   string
		columns []string
	}{
		{
			name:  "cart",
			model: &model.Cart{},
			table: "carts",
			columns: []string{
				"id", "user_id", "status", "created_at", "updated_at",
			},
		},
		{
			name:  "cart_item",
			model: &model.CartItem{},
			table: "cart_items",
			columns: []string{
				"id", "cart_id", "sku_id", "quantity", "intended_price_source",
				"flash_sale_item_id", "version", "created_at", "updated_at",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := schema.Parse(tc.model, &sync.Map{}, schema.NamingStrategy{})
			if err != nil {
				t.Fatalf("Parse(%s) error = %v", tc.name, err)
			}
			if parsed.Table != tc.table {
				t.Errorf("Table = %q, want %q", parsed.Table, tc.table)
			}
			for _, col := range tc.columns {
				if parsed.LookUpField(col) == nil {
					t.Errorf("column %q missing from schema %s", col, tc.name)
				}
			}
		})
	}
}
