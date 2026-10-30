package backend_test

import (
	"net/http"
	"strings"
	"testing"

	"cloudian/cloudian-scale/internal/httpserver"
	"cloudian/cloudian-scale/internal/model"
	"cloudian/cloudian-scale/internal/module/auth"
	"cloudian/cloudian-scale/internal/module/catalog"
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
		Database:       databasePingerStub{},
		AuthService:    authService,
		CatalogService: catalogService,
		SecureCookies:  true,
	})
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
