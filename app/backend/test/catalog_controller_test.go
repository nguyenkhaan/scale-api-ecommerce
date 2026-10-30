package backend_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/module/catalog"
	"github.com/gin-gonic/gin"
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

type catalogServiceStub struct {
	detail      catalog.ProductResponse
	detailError error
	detailCalls int
	list        catalog.ProductPageResponse
	listError   error
	listCalls   int
	listQuery   catalog.ListPublicProductsQuery
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
