package backend_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cloudian/cloudian-scale/internal/httpserver/middleware"
	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/model"
	"cloudian/cloudian-scale/internal/module/auth"
	"cloudian/cloudian-scale/internal/module/inventory"
	"github.com/gin-gonic/gin"
)

func TestInventoryAdjustmentUsesAdminInventoryPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	service := &inventoryServiceStub{adjustment: inventory.InventoryResponse{SKUID: "63000000-0000-4000-8000-000000000001", OnHand: 8, Available: 8, Version: 1}}
	inventory.RegisterRoutes(api, inventory.NewController(service), setInventoryTestIdentity(model.RoleAdmin), middleware.RequireRole(model.RoleAdmin))

	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/inventories/63000000-0000-4000-8000-000000000001/adjustments", strings.NewReader(`{"delta":8,"reason":"initial count"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	if service.adjustCalls != 1 {
		t.Fatalf("adjust calls = %d, want 1", service.adjustCalls)
	}

	wrongPath := httptest.NewRequest(http.MethodPost, "/api/v1/products/63000000-0000-4000-8000-000000000001/adjustments", strings.NewReader(`{"delta":8,"reason":"initial count"}`))
	wrongPath.Header.Set("Content-Type", "application/json")
	wrongResponse := httptest.NewRecorder()
	router.ServeHTTP(wrongResponse, wrongPath)
	if wrongResponse.Code != http.StatusNotFound {
		t.Fatalf("wrong path status = %d, want 404", wrongResponse.Code)
	}
}

func TestInventoryAdjustmentRejectsInvalidInputBeforeService(t *testing.T) {
	testCases := []struct {
		name string
		path string
		body string
	}{
		{name: "invalid SKU ID", path: "/api/v1/admin/inventories/not-a-uuid/adjustments", body: `{"delta":1,"reason":"count"}`},
		{name: "zero delta", path: "/api/v1/admin/inventories/63000000-0000-4000-8000-000000000001/adjustments", body: `{"delta":0,"reason":"count"}`},
		{name: "blank reason", path: "/api/v1/admin/inventories/63000000-0000-4000-8000-000000000001/adjustments", body: `{"delta":1,"reason":" "}`},
		{name: "unknown field", path: "/api/v1/admin/inventories/63000000-0000-4000-8000-000000000001/adjustments", body: `{"delta":1,"reason":"count","unexpected":true}`},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			api := router.Group("/api")
			service := &inventoryServiceStub{}
			inventory.RegisterRoutes(api, inventory.NewController(service), setInventoryTestIdentity(model.RoleAdmin), middleware.RequireRole(model.RoleAdmin))

			request := httptest.NewRequest(http.MethodPost, testCase.path, strings.NewReader(testCase.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			requireInventoryProblem(t, response, http.StatusUnprocessableEntity, problem.CodeValidation)
			if service.adjustCalls != 0 {
				t.Fatalf("adjust calls = %d, want 0", service.adjustCalls)
			}
		})
	}
}

func TestInventoryAdjustmentMapsStableErrors(t *testing.T) {
	testCases := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{name: "missing inventory", err: inventory.ErrInventoryNotFound, status: http.StatusNotFound, code: problem.CodeResourceNotFound},
		{name: "below reserved", err: inventory.ErrBelowReserved, status: http.StatusConflict, code: problem.CodeInventoryBelowReserved},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			api := router.Group("/api")
			service := &inventoryServiceStub{adjustmentError: testCase.err}
			inventory.RegisterRoutes(api, inventory.NewController(service), setInventoryTestIdentity(model.RoleAdmin), middleware.RequireRole(model.RoleAdmin))

			request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/inventories/63000000-0000-4000-8000-000000000001/adjustments", strings.NewReader(`{"delta":-1,"reason":"count"}`))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			requireInventoryProblem(t, response, testCase.status, testCase.code)
		})
	}
}

func TestInventoryListNormalizesPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	service := &inventoryServiceStub{}
	inventory.RegisterRoutes(api, inventory.NewController(service), setInventoryTestIdentity(model.RoleAdmin), middleware.RequireRole(model.RoleAdmin))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/inventories", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	if service.listCalls != 1 || service.listQuery.Limit != 20 {
		t.Fatalf("list calls/query = %d/%+v, want 1/default limit", service.listCalls, service.listQuery)
	}
}

func TestInventoryListRejectsInvalidParametersBeforeService(t *testing.T) {
	testCases := []string{
		"limit=0",
		"limit=101",
		"limit=not-a-number",
		"cursor=" + strings.Repeat("x", 501),
	}
	for _, query := range testCases {
		t.Run(query, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			api := router.Group("/api")
			service := &inventoryServiceStub{}
			inventory.RegisterRoutes(api, inventory.NewController(service), setInventoryTestIdentity(model.RoleAdmin), middleware.RequireRole(model.RoleAdmin))

			request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/inventories?"+query, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			requireInventoryProblem(t, response, http.StatusUnprocessableEntity, problem.CodeValidation)
			if service.listCalls != 0 {
				t.Fatalf("list calls = %d, want 0", service.listCalls)
			}
		})
	}
}

func TestInventoryListMapsMalformedCursorToValidationProblem(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	service := &inventoryServiceStub{listError: inventory.ErrInvalidCursor}
	inventory.RegisterRoutes(api, inventory.NewController(service), setInventoryTestIdentity(model.RoleAdmin), middleware.RequireRole(model.RoleAdmin))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/inventories?cursor=opaque", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	requireInventoryProblem(t, response, http.StatusUnprocessableEntity, problem.CodeValidation)
}

func TestInventoryAdjustmentHistoryUsesCursorPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	service := &inventoryServiceStub{}
	inventory.RegisterRoutes(api, inventory.NewController(service), setInventoryTestIdentity(model.RoleAdmin), middleware.RequireRole(model.RoleAdmin))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/inventories/63000000-0000-4000-8000-000000000001/adjustments?limit=10", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	if service.historyCalls != 1 || service.historySKU != "63000000-0000-4000-8000-000000000001" || service.historyQuery.Limit != 10 {
		t.Fatalf("history calls/SKU/query = %d/%q/%+v", service.historyCalls, service.historySKU, service.historyQuery)
	}
}

func TestInventoryAdjustmentHistoryRejectsBadInputAndMapsErrors(t *testing.T) {
	testCases := []struct {
		name   string
		path   string
		err    error
		status int
		code   string
	}{
		{name: "invalid SKU ID", path: "/api/v1/admin/inventories/not-a-uuid/adjustments", status: http.StatusUnprocessableEntity, code: problem.CodeValidation},
		{name: "malformed cursor", path: "/api/v1/admin/inventories/63000000-0000-4000-8000-000000000001/adjustments?cursor=opaque", err: inventory.ErrInvalidCursor, status: http.StatusUnprocessableEntity, code: problem.CodeValidation},
		{name: "unknown SKU", path: "/api/v1/admin/inventories/63000000-0000-4000-8000-000000000001/adjustments", err: inventory.ErrInventoryNotFound, status: http.StatusNotFound, code: problem.CodeResourceNotFound},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			api := router.Group("/api")
			service := &inventoryServiceStub{historyError: testCase.err}
			inventory.RegisterRoutes(api, inventory.NewController(service), setInventoryTestIdentity(model.RoleAdmin), middleware.RequireRole(model.RoleAdmin))

			request := httptest.NewRequest(http.MethodGet, testCase.path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			requireInventoryProblem(t, response, testCase.status, testCase.code)
			if testCase.name == "invalid SKU ID" && service.historyCalls != 0 {
				t.Fatalf("history calls = %d, want 0", service.historyCalls)
			}
		})
	}
}

func setInventoryTestIdentity(role string) gin.HandlerFunc {
	return func(context *gin.Context) {
		auth.SetIdentity(context, auth.Identity{UserID: "51000000-0000-4000-8000-000000000001", Role: role})
		context.Next()
	}
}

type inventoryServiceStub struct {
	adjustCalls     int
	adjustment      inventory.InventoryResponse
	adjustmentError error
	listCalls       int
	listQuery       inventory.ListInventoriesQuery
	listError       error
	historyCalls    int
	historySKU      string
	historyQuery    inventory.ListAdjustmentHistoryQuery
	historyError    error
}

func (service *inventoryServiceStub) Adjust(_ context.Context, _ string, _ string, _ inventory.AdjustmentRequest) (inventory.InventoryResponse, error) {
	service.adjustCalls++
	return service.adjustment, service.adjustmentError
}

func (service *inventoryServiceStub) ListInventories(_ context.Context, query inventory.ListInventoriesQuery) (inventory.InventoryPageResponse, error) {
	service.listCalls++
	service.listQuery = query
	return inventory.InventoryPageResponse{}, service.listError
}

func (service *inventoryServiceStub) ListAdjustments(_ context.Context, skuID string, query inventory.ListAdjustmentHistoryQuery) (inventory.InventoryAdjustmentPageResponse, error) {
	service.historyCalls++
	service.historySKU = skuID
	service.historyQuery = query
	return inventory.InventoryAdjustmentPageResponse{}, service.historyError
}

func requireInventoryProblem(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
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
