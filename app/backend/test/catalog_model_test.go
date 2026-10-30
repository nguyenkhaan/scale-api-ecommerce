package backend_test

import (
	"reflect"
	"sync"
	"testing"

	"cloudian/cloudian-scale/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

func TestCategoryCreatePreservesExplicitInvisible(t *testing.T) {
	database, err := gorm.Open(
		postgres.New(postgres.Config{DSN: "postgres://unused@127.0.0.1:1/unused"}),
		&gorm.Config{DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true},
	)
	if err != nil {
		t.Fatalf("open dry-run database: %v", err)
	}

	category := model.Category{Name: "Ẩn", Slug: "an"}
	visibility := reflect.ValueOf(&category).Elem().FieldByName("IsVisible")
	switch visibility.Kind() {
	case reflect.Bool:
		visibility.SetBool(false)
	case reflect.Pointer:
		hidden := false
		visibility.Set(reflect.ValueOf(&hidden))
	default:
		t.Fatalf("IsVisible kind = %s, want bool or pointer", visibility.Kind())
	}

	if err := database.Create(&category).Error; err != nil {
		t.Fatalf("dry-run create category: %v", err)
	}
	visibility = reflect.ValueOf(&category).Elem().FieldByName("IsVisible")
	if visibility.Kind() == reflect.Bool && visibility.Bool() {
		t.Fatal("explicit false visibility was replaced with true")
	}
	if visibility.Kind() == reflect.Pointer && (!visibility.IsNil() && visibility.Elem().Bool()) {
		t.Fatal("explicit false visibility was replaced with true")
	}
}
