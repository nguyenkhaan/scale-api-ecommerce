package backend_test

import (
	"sync"
	"testing"

	"cloudian/cloudian-scale/internal/model"
	"gorm.io/gorm/schema"
)

func TestCatalogModelsMapToExpectedTablesAndColumns(t *testing.T) {
	testCases := []struct {
		name    string
		model   any
		table   string
		columns []string
	}{
		{
			name:  "category",
			model: &model.Category{},
			table: "categories",
			columns: []string{
				"id", "name", "slug", "is_visible", "created_at", "updated_at",
			},
		},
		{
			name:  "product",
			model: &model.Product{},
			table: "products",
			columns: []string{
				"id", "category_id", "name", "description", "image_url", "status", "created_at", "updated_at",
			},
		},
		{
			name:  "sku",
			model: &model.SKU{},
			table: "skus",
			columns: []string{
				"id", "product_id", "code", "attributes", "regular_price", "max_quantity_per_order", "status", "created_at", "updated_at",
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
			if len(parsed.PrimaryFields) != 1 || parsed.PrimaryFields[0].DBName != "id" {
				t.Fatalf("primary fields = %v, want [id]", parsed.PrimaryFieldDBNames)
			}
			id := parsed.LookUpField("id")
			if id == nil || !id.HasDefaultValue || id.DefaultValue != "gen_random_uuid()" {
				t.Errorf("id default = %q, want gen_random_uuid()", id.DefaultValue)
			}
			for _, column := range testCase.columns {
				if parsed.LookUpField(column) == nil {
					t.Errorf("column %q is not mapped", column)
				}
			}
		})
	}

	product, err := schema.Parse(&model.Product{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse product schema: %v", err)
	}
	for _, column := range []string{"category_id", "image_url"} {
		field := product.LookUpField(column)
		if field == nil || field.NotNull {
			t.Errorf("product %s must be nullable", column)
		}
	}

	sku, err := schema.Parse(&model.SKU{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse SKU schema: %v", err)
	}
	if attributes := sku.LookUpField("attributes"); attributes == nil || attributes.DataType != "jsonb" {
		t.Errorf("attributes data type = %q, want jsonb", attributes.DataType)
	}
}
