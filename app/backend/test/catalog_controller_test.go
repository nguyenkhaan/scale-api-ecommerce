package backend_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cloudian/cloudian-scale/internal/httpserver/middleware"
	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/model"
	"cloudian/cloudian-scale/internal/module/auth"
	"cloudian/cloudian-scale/internal/module/catalog"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestCatalogDetailRejectsInvalidProductIDBeforeService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	service := &catalogServiceStub{}
	catalog.RegisterRoutes(api, catalog.NewController(service))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/products/not-a-uuid", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	requireCatalogProblem(t, response, http.StatusUnprocessableEntity, problem.CodeValidation)
	if service.detailCalls != 0 {
		t.Fatalf("service calls = %d, want 0", service.detailCalls)
	}
}

func TestCatalogDetailMapsMissingProductToStableProblem(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &catalogServiceStub{detailError: catalog.ErrProductNotFound}
	router := gin.New()
	api := router.Group("/api")
	catalog.RegisterRoutes(api, catalog.NewController(service))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/products/51000000-0000-4000-8000-000000000001", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	requireCatalogProblem(t, response, http.StatusNotFound, problem.CodeResourceNotFound)
}

func TestCatalogListNormalizesPublicQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &catalogServiceStub{}
	router := gin.New()
	api := router.Group("/api")
	catalog.RegisterRoutes(api, catalog.NewController(service))

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products?query=%20shoe%20&category_id=51000000-0000-4000-8000-000000000001",
		nil,
	)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	if service.listCalls != 1 {
		t.Fatalf("list calls = %d, want 1", service.listCalls)
	}
	if service.listQuery.Query != "shoe" || service.listQuery.Limit != 20 {
		t.Errorf("list query = %+v, want trimmed query and limit 20", service.listQuery)
	}
	if service.listQuery.CategoryID == nil || *service.listQuery.CategoryID != "51000000-0000-4000-8000-000000000001" {
		t.Errorf("category ID = %v", service.listQuery.CategoryID)
	}
}

func TestCatalogListRejectsInvalidParametersBeforeService(t *testing.T) {
	testCases := []struct {
		name  string
		query string
	}{
		{name: "limit zero", query: "limit=0"},
		{name: "limit too large", query: "limit=101"},
		{name: "limit not integer", query: "limit=many"},
		{name: "invalid category", query: "category_id=nope"},
		{name: "query too long", query: "query=" + strings.Repeat("x", 101)},
		{name: "cursor too long", query: "cursor=" + strings.Repeat("x", 501)},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			service := &catalogServiceStub{}
			router := gin.New()
			api := router.Group("/api")
			catalog.RegisterRoutes(api, catalog.NewController(service))

			request := httptest.NewRequest(http.MethodGet, "/api/v1/products?"+testCase.query, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			requireCatalogProblem(t, response, http.StatusUnprocessableEntity, problem.CodeValidation)
			if service.listCalls != 0 {
				t.Fatalf("list calls = %d, want 0", service.listCalls)
			}
		})
	}
}

func TestCatalogListMapsInvalidCursorToValidationProblem(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &catalogServiceStub{listError: catalog.ErrInvalidCursor}
	router := gin.New()
	api := router.Group("/api")
	catalog.RegisterRoutes(api, catalog.NewController(service))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/products?cursor=opaque", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	requireCatalogProblem(t, response, http.StatusUnprocessableEntity, problem.CodeValidation)
}

func TestCatalogServiceRejectsMalformedCursorBeforeDatabase(t *testing.T) {
	database, err := gorm.Open(
		postgres.New(postgres.Config{DSN: "postgres://unused@127.0.0.1:1/unused"}),
		&gorm.Config{DisableAutomaticPing: true},
	)
	if err != nil {
		t.Fatalf("open placeholder ORM: %v", err)
	}
	service, err := catalog.NewService(database)
	if err != nil {
		t.Fatalf("create catalog service: %v", err)
	}

	_, err = service.ListPublicProducts(context.Background(), catalog.ListPublicProductsQuery{
		Cursor: "not-a-cursor",
		Limit:  20,
	})
	if !errors.Is(err, catalog.ErrInvalidCursor) {
		t.Fatalf("ListPublicProducts() error = %v, want ErrInvalidCursor", err)
	}
}

func TestAdminCatalogRejectsAnonymousBeforeController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	service := &catalogServiceStub{}
	catalog.RegisterAdminRoutes(
		api,
		catalog.NewController(service),
		middleware.RequireSession((*auth.Service)(nil)),
		middleware.RequireRole(model.RoleAdmin),
	)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/products", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	requireCatalogProblem(t, response, http.StatusUnauthorized, problem.CodeUnauthenticated)
	if service.adminListCalls != 0 {
		t.Fatalf("admin service calls = %d, want 0", service.adminListCalls)
	}
}

func TestAdminCatalogRejectsCustomerBeforeController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	service := &catalogServiceStub{}
	catalog.RegisterAdminRoutes(
		api,
		catalog.NewController(service),
		setCatalogTestIdentity(model.RoleCustomer),
		middleware.RequireRole(model.RoleAdmin),
	)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/products", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	requireCatalogProblem(t, response, http.StatusForbidden, problem.CodeForbidden)
	if service.adminListCalls != 0 {
		t.Fatalf("admin service calls = %d, want 0", service.adminListCalls)
	}
}

func TestAdminCatalogRejectsInvalidListParametersBeforeService(t *testing.T) {
	testCases := []struct {
		name  string
		query string
	}{
		{name: "status", query: "status=hidden"},
		{name: "cursor", query: "cursor=" + strings.Repeat("x", 501)},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			api := router.Group("/api")
			service := &catalogServiceStub{}
			catalog.RegisterAdminRoutes(
				api,
				catalog.NewController(service),
				setCatalogTestIdentity(model.RoleAdmin),
				middleware.RequireRole(model.RoleAdmin),
			)

			request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/products?"+testCase.query, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			requireCatalogProblem(t, response, http.StatusUnprocessableEntity, problem.CodeValidation)
			if service.adminListCalls != 0 {
				t.Fatalf("admin service calls = %d, want 0", service.adminListCalls)
			}
		})
	}
}

func TestCreateProductRejectsInvalidRequestsBeforeService(t *testing.T) {
	validSKU := `{"code":"SKU-1","attributes":{},"regular_price":1000,"max_quantity_per_order":1,"initial_on_hand":0}`
	testCases := []struct {
		name string
		body string
	}{
		{name: "blank name", body: `{"name":" ","skus":[` + validSKU + `]}`},
		{name: "unknown field", body: `{"name":"Product","skus":[` + validSKU + `],"unexpected":true}`},
		{name: "attributes array", body: `{"name":"Product","skus":[{"code":"SKU-1","attributes":[],"regular_price":1000,"max_quantity_per_order":1,"initial_on_hand":0}]}`},
		{name: "non-positive price", body: `{"name":"Product","skus":[{"code":"SKU-1","regular_price":0,"max_quantity_per_order":1,"initial_on_hand":0}]}`},
		{name: "non-positive maximum", body: `{"name":"Product","skus":[{"code":"SKU-1","regular_price":1000,"max_quantity_per_order":0,"initial_on_hand":0}]}`},
		{name: "negative initial on hand", body: `{"name":"Product","skus":[{"code":"SKU-1","regular_price":1000,"max_quantity_per_order":1,"initial_on_hand":-1}]}`},
		{name: "duplicate normalized code", body: `{"name":"Product","skus":[` + validSKU + `,{"code":" SKU-1 ","regular_price":1000,"max_quantity_per_order":1,"initial_on_hand":0}]}`},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			api := router.Group("/api")
			service := &catalogServiceStub{}
			catalog.RegisterAdminRoutes(
				api,
				catalog.NewController(service),
				setCatalogTestIdentity(model.RoleAdmin),
				middleware.RequireRole(model.RoleAdmin),
			)

			request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/products", strings.NewReader(testCase.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			requireCatalogProblem(t, response, http.StatusUnprocessableEntity, problem.CodeValidation)
			if service.createProductCalls != 0 {
				t.Fatalf("create service calls = %d, want 0", service.createProductCalls)
			}
		})
	}
}

func TestCreateProductUsesAuthenticatedActorAndNormalizedInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	service := &catalogServiceStub{
		createdProduct: catalog.ProductResponse{ID: "62000000-0000-4000-8000-000000000001", Name: "Product"},
	}
	catalog.RegisterAdminRoutes(
		api,
		catalog.NewController(service),
		setCatalogTestIdentity(model.RoleAdmin),
		middleware.RequireRole(model.RoleAdmin),
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/products",
		strings.NewReader(`{"name":" Product ","skus":[{"code":" SKU-1 ","regular_price":1000,"max_quantity_per_order":1,"initial_on_hand":0}]}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", response.Code, response.Body.String())
	}
	if service.createProductActor != "51000000-0000-4000-8000-000000000001" {
		t.Errorf("actor = %q", service.createProductActor)
	}
	if service.createdRequest.Name != "Product" || service.createdRequest.SKUs[0].Code != "SKU-1" {
		t.Errorf("normalized request = %+v", service.createdRequest)
	}
}

func TestCreateProductMapsStableServiceErrors(t *testing.T) {
	testCases := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{name: "duplicate SKU", err: catalog.ErrSKUCodeTaken, status: http.StatusConflict, code: problem.CodeSKUAlreadyExists},
		{name: "missing category", err: catalog.ErrCategoryNotFound, status: http.StatusUnprocessableEntity, code: problem.CodeValidation},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			api := router.Group("/api")
			service := &catalogServiceStub{createProductError: testCase.err}
			catalog.RegisterAdminRoutes(api, catalog.NewController(service), setCatalogTestIdentity(model.RoleAdmin), middleware.RequireRole(model.RoleAdmin))

			request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/products", strings.NewReader(`{"name":"Product","skus":[{"code":"SKU-1","regular_price":1000,"max_quantity_per_order":1,"initial_on_hand":0}]}`))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			requireCatalogProblem(t, response, testCase.status, testCase.code)
		})
	}
}

func TestUpdateProductRejectsInvalidPatchesBeforeService(t *testing.T) {
	testCases := []struct {
		name string
		body string
	}{
		{name: "empty", body: `{}`},
		{name: "name null", body: `{"name":null}`},
		{name: "invalid category", body: `{"category_id":"not-a-uuid"}`},
		{name: "active null", body: `{"active":null}`},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			api := router.Group("/api")
			service := &catalogServiceStub{}
			catalog.RegisterAdminRoutes(api, catalog.NewController(service), setCatalogTestIdentity(model.RoleAdmin), middleware.RequireRole(model.RoleAdmin))

			request := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/products/62000000-0000-4000-8000-000000000001", strings.NewReader(testCase.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			requireCatalogProblem(t, response, http.StatusUnprocessableEntity, problem.CodeValidation)
			if service.updateProductCalls != 0 {
				t.Fatalf("update service calls = %d, want 0", service.updateProductCalls)
			}
		})
	}
}

func TestUpdateProductDistinguishesNullFromOmittedFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	service := &catalogServiceStub{updatedProduct: catalog.ProductResponse{ID: "62000000-0000-4000-8000-000000000001"}}
	catalog.RegisterAdminRoutes(api, catalog.NewController(service), setCatalogTestIdentity(model.RoleAdmin), middleware.RequireRole(model.RoleAdmin))

	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/admin/products/62000000-0000-4000-8000-000000000001",
		strings.NewReader(`{"category_id":null,"image_url":null,"active":false}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	updated := service.updatedRequest
	if !updated.CategoryID.Present || updated.CategoryID.Value != nil || !updated.ImageURL.Present || updated.ImageURL.Value != nil {
		t.Errorf("nullable fields = %+v", updated)
	}
	if updated.Name.Present || updated.Description.Present || !updated.Active.Present || updated.Active.Value {
		t.Errorf("field presence = %+v", updated)
	}
}

func TestUpdateSKURejectsEmptyPatchBeforeService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	service := &catalogServiceStub{}
	catalog.RegisterAdminRoutes(api, catalog.NewController(service), setCatalogTestIdentity(model.RoleAdmin), middleware.RequireRole(model.RoleAdmin))

	request := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/skus/63000000-0000-4000-8000-000000000001", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	requireCatalogProblem(t, response, http.StatusUnprocessableEntity, problem.CodeValidation)
	if service.updateSKUCalls != 0 {
		t.Fatalf("update SKU service calls = %d, want 0", service.updateSKUCalls)
	}
}

func setCatalogTestIdentity(role string) gin.HandlerFunc {
	return func(context *gin.Context) {
		auth.SetIdentity(context, auth.Identity{UserID: "51000000-0000-4000-8000-000000000001", Role: role})
		context.Next()
	}
}

type catalogServiceStub struct {
	detail             catalog.ProductResponse
	detailError        error
	detailCalls        int
	list               catalog.ProductPageResponse
	listError          error
	listCalls          int
	listQuery          catalog.ListPublicProductsQuery
	adminListCalls     int
	createProductCalls int
	createdProduct     catalog.ProductResponse
	createProductError error
	createProductActor string
	createdRequest     catalog.CreateProductRequest
	updateProductCalls int
	updatedProduct     catalog.ProductResponse
	updatedRequest     catalog.UpdateProductRequest
	updateSKUCalls     int
}

func (service *catalogServiceStub) ListAdminProducts(context.Context, catalog.AdminProductListQuery) (catalog.ProductPageResponse, error) {
	service.adminListCalls++
	return catalog.ProductPageResponse{}, nil
}

func (service *catalogServiceStub) GetAdminProduct(context.Context, string) (catalog.ProductResponse, error) {
	return catalog.ProductResponse{}, nil
}

func (service *catalogServiceStub) CreateProduct(_ context.Context, actor string, request catalog.CreateProductRequest) (catalog.ProductResponse, error) {
	service.createProductCalls++
	service.createProductActor = actor
	service.createdRequest = request
	return service.createdProduct, service.createProductError
}

func (service *catalogServiceStub) CreateSKU(context.Context, string, string, catalog.CreateSKUInput) (catalog.SKUResponse, error) {
	return catalog.SKUResponse{}, nil
}

func (service *catalogServiceStub) UpdateProduct(_ context.Context, _ string, request catalog.UpdateProductRequest) (catalog.ProductResponse, error) {
	service.updateProductCalls++
	service.updatedRequest = request
	return service.updatedProduct, nil
}

func (service *catalogServiceStub) UpdateSKU(context.Context, string, catalog.UpdateSKURequest) (catalog.SKUResponse, error) {
	service.updateSKUCalls++
	return catalog.SKUResponse{}, nil
}

func (service *catalogServiceStub) GetPublicProduct(context.Context, string) (catalog.ProductResponse, error) {
	service.detailCalls++
	return service.detail, service.detailError
}

func (service *catalogServiceStub) ListPublicProducts(
	_ context.Context,
	query catalog.ListPublicProductsQuery,
) (catalog.ProductPageResponse, error) {
	service.listCalls++
	service.listQuery = query
	return service.list, service.listError
}

func requireCatalogProblem(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if response.Code != status {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, status, response.Body.String())
	}
	var body problem.Problem
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode problem: %v", err)
	}
	if body.Code != code {
		t.Errorf("problem code = %q, want %q", body.Code, code)
	}
}
