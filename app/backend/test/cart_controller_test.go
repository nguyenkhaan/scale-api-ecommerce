package backend_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cloudian/cloudian-scale/internal/httpserver/middleware"
	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/model"
	"cloudian/cloudian-scale/internal/module/auth"
	"cloudian/cloudian-scale/internal/module/cart"
	"github.com/gin-gonic/gin"
)

func TestCartControllerValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())
	api := router.Group("/api")

	dummyAuth := func(c *gin.Context) {
		auth.SetIdentity(c, auth.Identity{
			UserID: "51000000-0000-4000-8000-000000000001",
			Role:   model.RoleCustomer,
		})
		c.Next()
	}

	cart.RegisterRoutes(api, cart.NewController(&cart.Service{}), dummyAuth)

	testCases := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "invalid sku uuid in path",
			method:     http.MethodPut,
			path:       "/api/cart/items/not-a-uuid",
			body:       `{"quantity":1}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   problem.CodeValidation,
		},
		{
			name:       "zero quantity",
			method:     http.MethodPut,
			path:       "/api/cart/items/53000000-0000-4000-8000-000000000001",
			body:       `{"quantity":0}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   problem.CodeValidation,
		},
		{
			name:       "negative quantity",
			method:     http.MethodPut,
			path:       "/api/cart/items/53000000-0000-4000-8000-000000000001",
			body:       `{"quantity":-5}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   problem.CodeValidation,
		},
		{
			name:       "invalid price source",
			method:     http.MethodPut,
			path:       "/api/cart/items/53000000-0000-4000-8000-000000000001",
			body:       `{"quantity":1,"price_source":"invalid"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   problem.CodeValidation,
		},
		{
			name:       "flash sale price source without flash_sale_item_id",
			method:     http.MethodPut,
			path:       "/api/cart/items/53000000-0000-4000-8000-000000000001",
			body:       `{"quantity":1,"price_source":"flash_sale"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   problem.CodeValidation,
		},
		{
			name:       "preview with empty items",
			method:     http.MethodPost,
			path:       "/api/checkout/preview",
			body:       `{"items":[]}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   problem.CodeValidation,
		},
		{
			name:       "preview with invalid item uuid",
			method:     http.MethodPost,
			path:       "/api/checkout/preview",
			body:       `{"items":[{"cart_item_id":"invalid","version":1}]}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   problem.CodeValidation,
		},
		{
			name:       "regular price source with flash sale item",
			method:     http.MethodPut,
			path:       "/api/cart/items/53000000-0000-4000-8000-000000000001",
			body:       `{"quantity":1,"flash_sale_item_id":"54000000-0000-4000-8000-000000000001"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   problem.CodeValidation,
		},
		{
			name:       "expected version below one",
			method:     http.MethodPut,
			path:       "/api/cart/items/53000000-0000-4000-8000-000000000001",
			body:       `{"quantity":1,"expected_version":0}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   problem.CodeValidation,
		},
		{
			name:       "preview version below one",
			method:     http.MethodPost,
			path:       "/api/checkout/preview",
			body:       `{"items":[{"cart_item_id":"55000000-0000-4000-8000-000000000001","version":0}]}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   problem.CodeValidation,
		},
		{
			name:       "preview duplicate cart item",
			method:     http.MethodPost,
			path:       "/api/checkout/preview",
			body:       `{"items":[{"cart_item_id":"55000000-0000-4000-8000-000000000001","version":1},{"cart_item_id":"55000000-0000-4000-8000-000000000001","version":1}]}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   problem.CodeValidation,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			requireProblem(t, rec, tc.wantStatus, tc.wantCode)
		})
	}
}
