package model

import (
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestInventoryModelsMapToExpectedTablesAndColumns(t *testing.T) {
	inventory, err := schema.Parse(&Inventory{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse inventory schema: %v", err)
	}
	if inventory.Table != "inventories" {
		t.Errorf("table = %q, want inventories", inventory.Table)
	}
	if len(inventory.PrimaryFields) != 1 || inventory.PrimaryFields[0].DBName != "sku_id" {
		t.Fatalf("primary fields = %v, want [sku_id]", inventory.PrimaryFieldDBNames)
	}
	for _, column := range []string{"sku_id", "on_hand", "reserved", "version", "updated_at"} {
		if inventory.LookUpField(column) == nil {
			t.Errorf("inventory column %q is not mapped", column)
		}
	}
	if inventory.LookUpField("available") != nil {
		t.Error("available must be derived, not persisted")
	}

	adjustment, err := schema.Parse(&InventoryAdjustment{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse inventory adjustment schema: %v", err)
	}
	if adjustment.Table != "inventory_adjustments" {
		t.Errorf("table = %q, want inventory_adjustments", adjustment.Table)
	}
	if len(adjustment.PrimaryFields) != 1 || adjustment.PrimaryFields[0].DBName != "id" {
		t.Fatalf("primary fields = %v, want [id]", adjustment.PrimaryFieldDBNames)
	}
	id := adjustment.LookUpField("id")
	if id == nil {
		t.Fatal("adjustment id is not mapped")
	}
	if !id.HasDefaultValue || id.DefaultValue != "gen_random_uuid()" {
		t.Errorf("id default = %q, want gen_random_uuid()", id.DefaultValue)
	}
	for _, column := range []string{
		"id", "sku_id", "actor_user_id", "delta", "on_hand_before", "on_hand_after", "reason", "created_at",
	} {
		if adjustment.LookUpField(column) == nil {
			t.Errorf("adjustment column %q is not mapped", column)
		}
	}
}
