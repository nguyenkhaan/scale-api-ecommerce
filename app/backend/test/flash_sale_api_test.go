package backend_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/httpserver"
	"cloudian/cloudian-scale/internal/model"
	"cloudian/cloudian-scale/internal/module/auth"
	"cloudian/cloudian-scale/internal/module/cart"
	"cloudian/cloudian-scale/internal/module/catalog"
	"cloudian/cloudian-scale/internal/module/flashsale"
	"cloudian/cloudian-scale/internal/module/inventory"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func newPhase3Router(t *testing.T, database *gorm.DB) http.Handler {
	t.Helper()
	authService, err := auth.NewService(database, bcrypt.MinCost)
	if err != nil {
		t.Fatalf("create auth service: %v", err)
	}
	catalogService, err := catalog.NewService(database)
	if err != nil {
		t.Fatalf("create catalog service: %v", err)
	}
	inventoryService, err := inventory.NewService(database)
	if err != nil {
		t.Fatalf("create inventory service: %v", err)
	}
	flashSaleService, err := flashsale.NewService(database)
	if err != nil {
		t.Fatalf("create flashsale service: %v", err)
	}
	cartService, err := cart.NewService(database)
	if err != nil {
		t.Fatalf("create cart service: %v", err)
	}

	return httpserver.NewRouter(httpserver.Dependencies{
		Database:         databasePingerStub{},
		AuthService:      authService,
		CatalogService:   catalogService,
		InventoryService: inventoryService,
		FlashSaleService: flashSaleService,
		CartService:      cartService,
		SecureCookies:    true,
	})
}

func newPhase3CustomerClient(t *testing.T, router http.Handler, database *gorm.DB, email string) *testClient {
	t.Helper()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("customer password long enough"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	customer := model.User{
		ID:              fmt.Sprintf("51000000-0000-4000-8000-%012d", time.Now().UnixNano()%1000000000),
		Email:           email,
		EmailNormalized: strings.ToLower(email),
		PasswordHash:    string(passwordHash),
		Role:            model.RoleCustomer,
		Status:          model.UserStatusActive,
	}
	if err := database.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	client := newTestClient(router)
	if response := client.send(t, http.MethodPost, "/api/v1/auth/login", credentials(customer.Email, "customer password long enough")); response.Code != http.StatusOK {
		t.Fatalf("customer login status = %d; body = %s", response.Code, response.Body.String())
	}
	return client
}

func TestAdminFlashSaleDraftAndScheduleLifecycle(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	router := newPhase3Router(t, testDatabase.ORM())
	admin := newCatalogAdminClient(t, router, testDatabase.ORM())

	// Seed product and SKU
	_, productID := seedCatalogProduct(t, testDatabase.ORM(), true, model.ProductStatusActive)
	skuID := seedCatalogSKU(t, testDatabase.ORM(), productID, "FS-SKU-1", model.SKUStatusActive, 100_000, 5, 20)

	startsAt := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
	endsAt := time.Now().Add(4 * time.Hour).UTC().Format(time.RFC3339)

	// 1. Create draft
	createPayload := map[string]any{
		"name":      "Mega 11.11 Flash Sale",
		"starts_at": startsAt,
		"ends_at":   endsAt,
		"items": []map[string]any{
			{
				"sku_id":         skuID,
				"sale_price":     60000,
				"quota_total":    10,
				"purchase_limit": 2,
			},
		},
	}

	createRes := admin.send(t, http.MethodPost, "/api/v1/admin/flash-sales", createPayload, admin.csrfHeader()...)
	if createRes.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201; body = %s", createRes.Code, createRes.Body.String())
	}
	created := decodeJSON[flashsale.FlashSaleResponse](t, createRes)
	if created.Status != "draft" {
		t.Errorf("created status = %q, want draft", created.Status)
	}
	if len(created.Items) != 1 || created.Items[0].SalePrice != 60000 {
		t.Errorf("created items mismatch: %+v", created.Items)
	}

	// 2. Update draft
	newName := "Mega 11.11 Flash Sale Updated"
	updatePayload := map[string]any{
		"name": newName,
	}
	updateRes := admin.send(t, http.MethodPatch, "/api/v1/admin/flash-sales/"+created.ID, updatePayload, admin.csrfHeader()...)
	if updateRes.Code != http.StatusOK {
		t.Fatalf("update status = %d, want 200; body = %s", updateRes.Code, updateRes.Body.String())
	}
	updated := decodeJSON[flashsale.FlashSaleResponse](t, updateRes)
	if updated.Name != newName {
		t.Errorf("updated name = %q, want %q", updated.Name, newName)
	}

	// 3. Schedule campaign
	schedRes := admin.send(t, http.MethodPost, "/api/v1/admin/flash-sales/"+created.ID+"/schedule", nil, admin.csrfHeader()...)
	if schedRes.Code != http.StatusOK {
		t.Fatalf("schedule status = %d, want 200; body = %s", schedRes.Code, schedRes.Body.String())
	}
	scheduled := decodeJSON[flashsale.FlashSaleResponse](t, schedRes)
	if scheduled.Status != "scheduled" {
		t.Errorf("scheduled status = %q, want scheduled", scheduled.Status)
	}

	// 4. Updating scheduled campaign should fail with 409 CONFLICT
	failUpdateRes := admin.send(t, http.MethodPatch, "/api/v1/admin/flash-sales/"+created.ID, updatePayload, admin.csrfHeader()...)
	requireProblem(t, failUpdateRes, http.StatusConflict, "CONFLICT")

	// 5. Cancel schedule before start -> reverts to draft
	cancelRes := admin.send(t, http.MethodPost, "/api/v1/admin/flash-sales/"+created.ID+"/cancel-schedule", nil, admin.csrfHeader()...)
	if cancelRes.Code != http.StatusOK {
		t.Fatalf("cancel schedule status = %d, want 200; body = %s", cancelRes.Code, cancelRes.Body.String())
	}
	cancelled := decodeJSON[flashsale.FlashSaleResponse](t, cancelRes)
	if cancelled.Status != "draft" {
		t.Errorf("cancelled status = %q, want draft", cancelled.Status)
	}

	// Re-schedule
	schedRes2 := admin.send(t, http.MethodPost, "/api/v1/admin/flash-sales/"+created.ID+"/schedule", nil, admin.csrfHeader()...)
	if schedRes2.Code != http.StatusOK {
		t.Fatalf("re-schedule status = %d, want 200; body = %s", schedRes2.Code, schedRes2.Body.String())
	}

	// 6. Stop campaign -> sets status to stopped
	stopRes := admin.send(t, http.MethodPost, "/api/v1/admin/flash-sales/"+created.ID+"/stop", nil, admin.csrfHeader()...)
	if stopRes.Code != http.StatusOK {
		t.Fatalf("stop status = %d, want 200; body = %s", stopRes.Code, stopRes.Body.String())
	}
	stopped := decodeJSON[flashsale.FlashSaleResponse](t, stopRes)
	if stopped.Status != "stopped" {
		t.Errorf("stopped status = %q, want stopped", stopped.Status)
	}
}

func TestAdminFlashSaleAntiOverlapConstraint(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	router := newPhase3Router(t, testDatabase.ORM())
	admin := newCatalogAdminClient(t, router, testDatabase.ORM())

	_, productID := seedCatalogProduct(t, testDatabase.ORM(), true, model.ProductStatusActive)
	skuID := seedCatalogSKU(t, testDatabase.ORM(), productID, "FS-OVERLAP-SKU", model.SKUStatusActive, 100_000, 5, 20)

	baseTime := time.Now().Add(1 * time.Hour)
	// Campaign 1: T+1h to T+3h
	c1Payload := map[string]any{
		"name":      "Campaign 1",
		"starts_at": baseTime.UTC().Format(time.RFC3339),
		"ends_at":   baseTime.Add(2 * time.Hour).UTC().Format(time.RFC3339),
		"items": []map[string]any{
			{"sku_id": skuID, "sale_price": 50000, "quota_total": 5, "purchase_limit": 1},
		},
	}
	res1 := admin.send(t, http.MethodPost, "/api/v1/admin/flash-sales", c1Payload, admin.csrfHeader()...)
	c1 := decodeJSON[flashsale.FlashSaleResponse](t, res1)
	sched1 := admin.send(t, http.MethodPost, "/api/v1/admin/flash-sales/"+c1.ID+"/schedule", nil, admin.csrfHeader()...)
	if sched1.Code != http.StatusOK {
		t.Fatalf("schedule c1 status = %d; body = %s", sched1.Code, sched1.Body.String())
	}

	// Campaign 2: T+2h to T+4h (overlaps with Campaign 1)
	c2Payload := map[string]any{
		"name":      "Campaign 2 (Overlapping)",
		"starts_at": baseTime.Add(1 * time.Hour).UTC().Format(time.RFC3339),
		"ends_at":   baseTime.Add(3 * time.Hour).UTC().Format(time.RFC3339),
		"items": []map[string]any{
			{"sku_id": skuID, "sale_price": 45000, "quota_total": 5, "purchase_limit": 1},
		},
	}
	res2 := admin.send(t, http.MethodPost, "/api/v1/admin/flash-sales", c2Payload, admin.csrfHeader()...)
	c2 := decodeJSON[flashsale.FlashSaleResponse](t, res2)

	// Scheduling Campaign 2 should fail with 409 CONFLICT due to overlapping SKU
	sched2 := admin.send(t, http.MethodPost, "/api/v1/admin/flash-sales/"+c2.ID+"/schedule", nil, admin.csrfHeader()...)
	requireProblem(t, sched2, http.StatusConflict, "CONFLICT")

	// Campaign 3: T+4h to T+6h (no overlap)
	c3Payload := map[string]any{
		"name":      "Campaign 3 (Non-overlapping)",
		"starts_at": baseTime.Add(3 * time.Hour).UTC().Format(time.RFC3339),
		"ends_at":   baseTime.Add(5 * time.Hour).UTC().Format(time.RFC3339),
		"items": []map[string]any{
			{"sku_id": skuID, "sale_price": 45000, "quota_total": 5, "purchase_limit": 1},
		},
	}
	res3 := admin.send(t, http.MethodPost, "/api/v1/admin/flash-sales", c3Payload, admin.csrfHeader()...)
	c3 := decodeJSON[flashsale.FlashSaleResponse](t, res3)
	sched3 := admin.send(t, http.MethodPost, "/api/v1/admin/flash-sales/"+c3.ID+"/schedule", nil, admin.csrfHeader()...)
	if sched3.Code != http.StatusOK {
		t.Fatalf("schedule c3 status = %d, want 200; body = %s", sched3.Code, sched3.Body.String())
	}
}

func TestPublicFlashSaleListingAndDetail(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	router := newPhase3Router(t, testDatabase.ORM())
	admin := newCatalogAdminClient(t, router, testDatabase.ORM())
	public := newTestClient(router)

	_, productID := seedCatalogProduct(t, testDatabase.ORM(), true, model.ProductStatusActive)
	skuID := seedCatalogSKU(t, testDatabase.ORM(), productID, "FS-PUB-SKU", model.SKUStatusActive, 100_000, 5, 20)

	// 1. Create a draft campaign -> should NOT appear in public listing
	draftPayload := map[string]any{
		"name":      "Draft Campaign",
		"starts_at": time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
		"ends_at":   time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339),
		"items": []map[string]any{
			{"sku_id": skuID, "sale_price": 50000, "quota_total": 5, "purchase_limit": 1},
		},
	}
	draftRes := admin.send(t, http.MethodPost, "/api/v1/admin/flash-sales", draftPayload, admin.csrfHeader()...)
	draftCampaign := decodeJSON[flashsale.FlashSaleResponse](t, draftRes)

	pubListBefore := public.send(t, http.MethodGet, "/api/v1/flash-sales", nil)
	if pubListBefore.Code != http.StatusOK {
		t.Fatalf("public list status = %d; body = %s", pubListBefore.Code, pubListBefore.Body.String())
	}
	bodyBefore := decodeJSON[flashsale.FlashSalePageResponse](t, pubListBefore)
	if len(bodyBefore.Items) != 0 {
		t.Errorf("public list contains draft campaign: %+v", bodyBefore.Items)
	}

	// 2. Schedule it -> should appear in public listing
	admin.send(t, http.MethodPost, "/api/v1/admin/flash-sales/"+draftCampaign.ID+"/schedule", nil, admin.csrfHeader()...)

	pubListAfter := public.send(t, http.MethodGet, "/api/v1/flash-sales", nil)
	if pubListAfter.Code != http.StatusOK {
		t.Fatalf("public list after schedule status = %d; body = %s", pubListAfter.Code, pubListAfter.Body.String())
	}
	bodyAfter := decodeJSON[flashsale.FlashSalePageResponse](t, pubListAfter)
	if len(bodyAfter.Items) != 1 {
		t.Fatalf("public list items len = %d, want 1", len(bodyAfter.Items))
	}

	// Verify no internal inventory fields leak
	for _, privateField := range []string{"on_hand", "reserved"} {
		if strings.Contains(pubListAfter.Body.String(), privateField) {
			t.Errorf("public response leaked %q: %s", privateField, pubListAfter.Body.String())
		}
	}
}
