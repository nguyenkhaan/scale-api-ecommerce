package backend_test

import (
	"context"
	"errors"
	"testing"

	"cloudian/cloudian-scale/internal/model"
	"cloudian/cloudian-scale/internal/module/inventory"
)

func TestInventoryServiceAdjustsAtomicallyAndPaginatesLedger(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	database := testDatabase.ORM()
	actor := model.User{
		ID:              "51000000-0000-4000-8000-000000000020",
		Email:           "inventory-admin@example.com",
		EmailNormalized: "inventory-admin@example.com",
		PasswordHash:    "not-used-by-service-test",
		Role:            model.RoleAdmin,
		Status:          model.UserStatusActive,
	}
	if err := database.Create(&actor).Error; err != nil {
		t.Fatalf("create actor: %v", err)
	}
	product := model.Product{Name: "Inventory product", Status: model.ProductStatusActive}
	if err := database.Create(&product).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}
	sku := model.SKU{
		ProductID:           product.ID,
		Code:                "INVENTORY-TEST-SKU",
		Attributes:          []byte(`{}`),
		RegularPrice:        100_000,
		MaxQuantityPerOrder: 5,
		Status:              model.SKUStatusActive,
	}
	if err := database.Create(&sku).Error; err != nil {
		t.Fatalf("create SKU: %v", err)
	}
	if err := database.Create(&model.Inventory{SKUID: sku.ID, OnHand: 10, Reserved: 3}).Error; err != nil {
		t.Fatalf("create inventory: %v", err)
	}

	service, err := inventory.NewService(database)
	if err != nil {
		t.Fatalf("create inventory service: %v", err)
	}
	if _, err := service.Adjust(context.Background(), sku.ID, actor.ID, inventory.AdjustmentRequest{Delta: -8, Reason: "too low"}); !errors.Is(err, inventory.ErrBelowReserved) {
		t.Fatalf("adjust below reserved error = %v, want ErrBelowReserved", err)
	}
	adjusted, err := service.Adjust(context.Background(), sku.ID, actor.ID, inventory.AdjustmentRequest{Delta: -6, Reason: "cycle count"})
	if err != nil {
		t.Fatalf("decrease inventory: %v", err)
	}
	if adjusted.OnHand != 4 || adjusted.Reserved != 3 || adjusted.Available != 1 || adjusted.Version != 1 {
		t.Errorf("adjusted inventory = %+v", adjusted)
	}
	if _, err := service.Adjust(context.Background(), sku.ID, actor.ID, inventory.AdjustmentRequest{Delta: 2, Reason: "receipt"}); err != nil {
		t.Fatalf("increase inventory: %v", err)
	}

	var current model.Inventory
	if err := database.Where("sku_id = ?", sku.ID).Take(&current).Error; err != nil {
		t.Fatalf("reload inventory: %v", err)
	}
	if current.OnHand != 6 || current.Reserved != 3 || current.Version != 2 {
		t.Errorf("persisted inventory = %+v", current)
	}
	if err := database.Exec("UPDATE inventories SET reserved = ? WHERE sku_id = ?", 7, sku.ID).Error; err == nil {
		t.Fatal("database accepted reserved quantity greater than on_hand")
	}

	page, err := service.ListInventories(context.Background(), inventory.ListInventoriesQuery{Limit: 20})
	if err != nil {
		t.Fatalf("ListInventories() error = %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].SKUCode != sku.Code || page.Items[0].Available != 3 {
		t.Errorf("inventory page = %+v", page)
	}

	first, err := service.ListAdjustments(context.Background(), sku.ID, inventory.ListAdjustmentHistoryQuery{Limit: 1})
	if err != nil {
		t.Fatalf("ListAdjustments() first page error = %v", err)
	}
	if len(first.Items) != 1 || first.Items[0].Reason != "receipt" || first.NextCursor == nil {
		t.Fatalf("first adjustment page = %+v", first)
	}
	second, err := service.ListAdjustments(context.Background(), sku.ID, inventory.ListAdjustmentHistoryQuery{Limit: 1, Cursor: *first.NextCursor})
	if err != nil {
		t.Fatalf("ListAdjustments() second page error = %v", err)
	}
	if len(second.Items) != 1 || second.Items[0].Reason != "cycle count" || second.NextCursor != nil {
		t.Errorf("second adjustment page = %+v", second)
	}
}
