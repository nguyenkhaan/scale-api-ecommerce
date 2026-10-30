package backend_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/httpserver"
	"cloudian/cloudian-scale/internal/model"
	"cloudian/cloudian-scale/internal/module/auth"
	"cloudian/cloudian-scale/internal/module/catalog"
	"cloudian/cloudian-scale/internal/module/inventory"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestPublicProductDetailReturnsOnlyActiveCatalogData(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	router := newCatalogRouter(t, testDatabase.ORM())
	categoryID, productID := seedCatalogProduct(t, testDatabase.ORM(), true, model.ProductStatusActive)
	_ = categoryID
	seedCatalogSKU(t, testDatabase.ORM(), productID, "ACTIVE-SKU", model.SKUStatusActive, 100_000, 5, 7)
	seedCatalogSKU(t, testDatabase.ORM(), productID, "INACTIVE-SKU", model.SKUStatusInactive, 120_000, 5, 9)

	response := newTestClient(router).send(t, http.MethodGet, "/api/v1/products/"+productID, nil)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	body := decodeJSON[catalog.ProductResponse](t, response)
	if len(body.SKUs) != 1 || body.SKUs[0].Code != "ACTIVE-SKU" {
		t.Fatalf("SKUs = %+v, want only ACTIVE-SKU", body.SKUs)
	}
	for _, privateField := range []string{"on_hand", "reserved", "available"} {
		if strings.Contains(response.Body.String(), privateField) {
			t.Errorf("public response leaked %q: %s", privateField, response.Body.String())
		}
	}
}

func TestPublicProductDetailHidesInactiveOrInvisibleCatalog(t *testing.T) {
	testCases := []struct {
		name            string
		categoryVisible bool
		productStatus   string
	}{
		{name: "inactive product", categoryVisible: true, productStatus: model.ProductStatusInactive},
		{name: "invisible category", categoryVisible: false, productStatus: model.ProductStatusActive},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testDatabase := openIntegrationDatabase(t)
			router := newCatalogRouter(t, testDatabase.ORM())
			_, productID := seedCatalogProduct(t, testDatabase.ORM(), testCase.categoryVisible, testCase.productStatus)
			seedCatalogSKU(t, testDatabase.ORM(), productID, "SKU-1", model.SKUStatusActive, 100_000, 5, 1)

			response := newTestClient(router).send(t, http.MethodGet, "/api/v1/products/"+productID, nil)
			requireProblem(t, response, http.StatusNotFound, "RESOURCE_NOT_FOUND")
		})
	}
}

func TestPublicProductListFiltersLiteralSearchAndPaginates(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	database := testDatabase.ORM()
	router := newCatalogRouter(t, database)
	visible := true
	category := model.Category{ID: "61000000-0000-4000-8000-000000000010", Name: "Visible", Slug: "visible", IsVisible: &visible}
	if err := database.Create(&category).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}
	createdAt := time.Date(2026, time.October, 1, 12, 0, 0, 0, time.UTC)
	products := []model.Product{
		{ID: "62000000-0000-4000-8000-000000000010", CategoryID: &category.ID, Name: "100%_ cotton", Status: model.ProductStatusActive, CreatedAt: createdAt.Add(-2 * time.Hour), UpdatedAt: createdAt},
		{ID: "62000000-0000-4000-8000-000000000011", CategoryID: &category.ID, Name: "Newest active", Status: model.ProductStatusActive, CreatedAt: createdAt, UpdatedAt: createdAt},
		{ID: "62000000-0000-4000-8000-000000000012", CategoryID: &category.ID, Name: "Hidden product", Status: model.ProductStatusInactive, CreatedAt: createdAt.Add(time.Hour), UpdatedAt: createdAt},
	}
	for index := range products {
		if err := database.Create(&products[index]).Error; err != nil {
			t.Fatalf("create product %d: %v", index, err)
		}
		seedCatalogSKU(t, database, products[index].ID, "LIST-SKU-"+products[index].ID[len(products[index].ID)-2:], model.SKUStatusActive, 100_000, 5, 5)
	}

	client := newTestClient(router)
	literal := client.send(t, http.MethodGet, "/api/v1/products?query=%25_&category_id="+category.ID, nil)
	if literal.Code != http.StatusOK {
		t.Fatalf("literal search status = %d; body = %s", literal.Code, literal.Body.String())
	}
	literalPage := decodeJSON[catalog.ProductPageResponse](t, literal)
	if len(literalPage.Items) != 1 || literalPage.Items[0].Name != "100%_ cotton" {
		t.Fatalf("literal search items = %+v", literalPage.Items)
	}

	first := client.send(t, http.MethodGet, "/api/v1/products?category_id="+category.ID+"&limit=1", nil)
	firstPage := decodeJSON[catalog.ProductPageResponse](t, first)
	if first.Code != http.StatusOK || len(firstPage.Items) != 1 || firstPage.NextCursor == nil {
		t.Fatalf("first page = %d %+v", first.Code, firstPage)
	}
	second := client.send(t, http.MethodGet, "/api/v1/products?category_id="+category.ID+"&limit=1&cursor="+*firstPage.NextCursor, nil)
	secondPage := decodeJSON[catalog.ProductPageResponse](t, second)
	if second.Code != http.StatusOK || len(secondPage.Items) != 1 || secondPage.Items[0].ID == firstPage.Items[0].ID || secondPage.NextCursor != nil {
		t.Fatalf("second page = %d %+v", second.Code, secondPage)
	}
}

func TestCatalogServiceCreatesInitialInventoryAtomicallyAndKeepsSKUCaseSensitive(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	database := testDatabase.ORM()
	service, err := catalog.NewService(database)
	if err != nil {
		t.Fatalf("create catalog service: %v", err)
	}
	actor := model.User{
		ID:              "51000000-0000-4000-8000-000000000010",
		Email:           "catalog-admin@example.com",
		EmailNormalized: "catalog-admin@example.com",
		PasswordHash:    "not-used-by-service-test",
		Role:            model.RoleAdmin,
		Status:          model.UserStatusActive,
	}
	if err := database.Create(&actor).Error; err != nil {
		t.Fatalf("create actor: %v", err)
	}

	created, err := service.CreateProduct(context.Background(), actor.ID, catalog.CreateProductRequest{
		Name: "Case-sensitive SKU",
		SKUs: []catalog.CreateSKUInput{{
			Code:                "CASE-SKU",
			Attributes:          []byte(`{"size":"M"}`),
			RegularPrice:        100_000,
			MaxQuantityPerOrder: 2,
			InitialOnHand:       7,
		}},
	})
	if err != nil {
		t.Fatalf("CreateProduct() error = %v", err)
	}
	var inventoryRow model.Inventory
	if err := database.Where("sku_id = ?", created.SKUs[0].ID).Take(&inventoryRow).Error; err != nil {
		t.Fatalf("find inventory: %v", err)
	}
	if inventoryRow.OnHand != 7 || inventoryRow.Reserved != 0 {
		t.Errorf("initial inventory = %+v", inventoryRow)
	}
	var adjustments []model.InventoryAdjustment
	if err := database.Where("sku_id = ?", created.SKUs[0].ID).Find(&adjustments).Error; err != nil {
		t.Fatalf("find initial adjustment: %v", err)
	}
	if len(adjustments) != 1 || adjustments[0].Delta != 7 || adjustments[0].ActorUserID != actor.ID {
		t.Errorf("initial adjustments = %+v", adjustments)
	}

	if _, err := service.CreateSKU(context.Background(), actor.ID, created.ID, catalog.CreateSKUInput{
		Code: "case-sku", Attributes: []byte(`{}`), RegularPrice: 100_000, MaxQuantityPerOrder: 1, InitialOnHand: 0,
	}); err != nil {
		t.Fatalf("CreateSKU() with different case error = %v", err)
	}
	_, err = service.CreateProduct(context.Background(), actor.ID, catalog.CreateProductRequest{
		Name: "Must rollback",
		SKUs: []catalog.CreateSKUInput{
			{Code: "TEMP-ROLLBACK", Attributes: []byte(`{}`), RegularPrice: 100_000, MaxQuantityPerOrder: 1, InitialOnHand: 1},
			{Code: "CASE-SKU", Attributes: []byte(`{}`), RegularPrice: 100_000, MaxQuantityPerOrder: 1, InitialOnHand: 1},
		},
	})
	if !errors.Is(err, catalog.ErrSKUCodeTaken) {
		t.Fatalf("duplicate product error = %v, want ErrSKUCodeTaken", err)
	}
	var count int64
	if err := database.Model(&model.Product{}).Where("name = ?", "Must rollback").Count(&count).Error; err != nil {
		t.Fatalf("count rolled-back product: %v", err)
	}
	if count != 0 {
		t.Errorf("rolled-back product count = %d, want 0", count)
	}
	if err := database.Model(&model.SKU{}).Where("code = ?", "TEMP-ROLLBACK").Count(&count).Error; err != nil {
		t.Fatalf("count rolled-back SKU: %v", err)
	}
	if count != 0 {
		t.Errorf("rolled-back SKU count = %d, want 0", count)
	}
}

func TestAdminCatalogMutationRequiresCSRFAndHidesInactiveProduct(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	database := testDatabase.ORM()
	router := newCatalogRouter(t, database)
	admin := newCatalogAdminClient(t, router, database)
	body := map[string]any{
		"name": "HTTP product",
		"skus": []map[string]any{{
			"code": "HTTP-SKU", "regular_price": 100_000, "max_quantity_per_order": 2, "initial_on_hand": 3,
		}},
	}

	missingCSRF := admin.send(t, http.MethodPost, "/api/v1/admin/products", body)
	requireProblem(t, missingCSRF, http.StatusForbidden, "FORBIDDEN")
	createdResponse := admin.send(t, http.MethodPost, "/api/v1/admin/products", body, admin.csrfHeader()...)
	if createdResponse.Code != http.StatusCreated {
		t.Fatalf("create product status = %d; body = %s", createdResponse.Code, createdResponse.Body.String())
	}
	created := decodeJSON[catalog.ProductResponse](t, createdResponse)
	if len(created.SKUs) != 1 || created.SKUs[0].Code != "HTTP-SKU" {
		t.Fatalf("created product = %+v", created)
	}

	public := newTestClient(router).send(t, http.MethodGet, "/api/v1/products/"+created.ID, nil)
	if public.Code != http.StatusOK || strings.Contains(public.Body.String(), "on_hand") || strings.Contains(public.Body.String(), "reserved") {
		t.Fatalf("public response = %d %s", public.Code, public.Body.String())
	}

	hidden := admin.send(t, http.MethodPatch, "/api/v1/admin/products/"+created.ID, map[string]any{"active": false}, admin.csrfHeader()...)
	if hidden.Code != http.StatusOK {
		t.Fatalf("hide product status = %d; body = %s", hidden.Code, hidden.Body.String())
	}
	requireProblem(t, newTestClient(router).send(t, http.MethodGet, "/api/v1/products/"+created.ID, nil), http.StatusNotFound, "RESOURCE_NOT_FOUND")
	if adminDetail := admin.send(t, http.MethodGet, "/api/v1/admin/products/"+created.ID, nil); adminDetail.Code != http.StatusOK {
		t.Errorf("admin cannot read inactive product: %d %s", adminDetail.Code, adminDetail.Body.String())
	}
}

func newCatalogRouter(t *testing.T, database *gorm.DB) http.Handler {
	t.Helper()
	authService, err := auth.NewService(database, bcrypt.MinCost)
	if err != nil {
		t.Fatalf("create auth service: %v", err)
	}
	catalogService, err := catalog.NewService(database)
	if err != nil {
		t.Fatalf("create catalog service: %v", err)
	}
	return httpserver.NewRouter(httpserver.Dependencies{
		Database:         databasePingerStub{},
		AuthService:      authService,
		CatalogService:   catalogService,
		InventoryService: mustInventoryService(t, database),
		SecureCookies:    true,
	})
}

func mustInventoryService(t *testing.T, database *gorm.DB) *inventory.Service {
	t.Helper()
	service, err := inventory.NewService(database)
	if err != nil {
		t.Fatalf("create inventory service: %v", err)
	}
	return service
}

func newCatalogAdminClient(t *testing.T, router http.Handler, database *gorm.DB) *testClient {
	t.Helper()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("admin password long enough"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash admin password: %v", err)
	}
	admin := model.User{
		ID:              "51000000-0000-4000-8000-000000000030",
		Email:           "http-admin@example.com",
		EmailNormalized: "http-admin@example.com",
		PasswordHash:    string(passwordHash),
		Role:            model.RoleAdmin,
		Status:          model.UserStatusActive,
	}
	if err := database.Create(&admin).Error; err != nil {
		t.Fatalf("create admin: %v", err)
	}
	client := newTestClient(router)
	if response := client.send(t, http.MethodPost, "/api/v1/auth/login", credentials(admin.Email, "admin password long enough")); response.Code != http.StatusOK {
		t.Fatalf("admin login status = %d; body = %s", response.Code, response.Body.String())
	}
	return client
}

func seedCatalogProduct(t *testing.T, database *gorm.DB, categoryVisible bool, status string) (string, string) {
	t.Helper()
	categoryID := "61000000-0000-4000-8000-000000000001"
	productID := "62000000-0000-4000-8000-000000000001"
	if err := database.Create(&model.Category{
		ID: categoryID, Name: "Catalog", Slug: "catalog", IsVisible: &categoryVisible,
	}).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}
	if err := database.Create(&model.Product{
		ID: productID, CategoryID: &categoryID, Name: "Product", Status: status,
	}).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}
	return categoryID, productID
}

func seedCatalogSKU(
	t *testing.T,
	database *gorm.DB,
	productID, code, status string,
	price int64,
	maxQuantity int,
	onHand int64,
) string {
	t.Helper()
	sku := model.SKU{
		ProductID:           productID,
		Code:                code,
		Attributes:          []byte(`{}`),
		RegularPrice:        price,
		MaxQuantityPerOrder: maxQuantity,
		Status:              status,
	}
	if err := database.Create(&sku).Error; err != nil {
		t.Fatalf("create SKU: %v", err)
	}
	if err := database.Create(&model.Inventory{SKUID: sku.ID, OnHand: onHand}).Error; err != nil {
		t.Fatalf("create inventory: %v", err)
	}
	return sku.ID
}
