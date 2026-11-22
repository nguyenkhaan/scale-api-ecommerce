package backend_test

import (
	"net/http"
	"testing"

	"cloudian/cloudian-scale/internal/httpserver/middleware"
	"cloudian/cloudian-scale/internal/module/auth"
	"cloudian/cloudian-scale/internal/module/order"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// orderAPINewRouter serves only the auth and customer order endpoints.
func orderAPINewRouter(t *testing.T, database *gorm.DB) http.Handler {
	t.Helper()
	gin.SetMode(gin.TestMode)
	authService, err := auth.NewService(database, bcrypt.MinCost)
	if err != nil {
		t.Fatalf("create auth service: %v", err)
	}
	queries, err := order.NewOrderQueries(database)
	if err != nil {
		t.Fatalf("create order queries: %v", err)
	}
	requireSession := middleware.RequireSession(authService)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware(), gin.Recovery())
	api := router.Group("/api")
	auth.RegisterRoutes(api, auth.NewController(authService, true), requireSession)
	order.RegisterOrderRoutes(api, order.NewOrdersController(queries), requireSession)
	return router
}

// orderAPISeedOrder seeds a pending order with one regular line of the given quantity on a fresh SKU.
func orderAPISeedOrder(t *testing.T, database *gorm.DB, userID string, expiresInSecs float64, quantity int) (orderID, skuID string) {
	t.Helper()
	skuID = cartSeedSKU(t, database, cartSKUOptions{OnHand: 50})
	orderID = lifecycleSeedPendingOrder(t, database, userID, expiresInSecs,
		lifecycleLine{SKUID: skuID, Quantity: quantity, UnitPrice: 25_000})
	return orderID, skuID
}

func orderAPIGet(t *testing.T, client *testClient, path string) order.OrderResponse {
	t.Helper()
	response := client.send(t, http.MethodGet, path, nil)
	if response.Code != http.StatusOK {
		t.Fatalf("GET %s status = %d; body = %s", path, response.Code, response.Body.String())
	}
	return decodeJSON[order.OrderResponse](t, response)
}

func orderAPIList(t *testing.T, client *testClient, query string) order.OrderPage {
	t.Helper()
	response := client.send(t, http.MethodGet, "/api/orders"+query, nil)
	if response.Code != http.StatusOK {
		t.Fatalf("GET /api/orders%s status = %d; body = %s", query, response.Code, response.Body.String())
	}
	return decodeJSON[order.OrderPage](t, response)
}

func orderAPICancel(t *testing.T, client *testClient, orderID string) (status int, body []byte) {
	t.Helper()
	response := client.send(t, http.MethodPost, "/api/orders/"+orderID+"/cancel", nil, client.csrfHeader()...)
	return response.Code, response.Body.Bytes()
}

func orderAPIStatus(t *testing.T, database *gorm.DB, orderID string) string {
	t.Helper()
	var status string
	if err := database.Raw(`SELECT status FROM orders WHERE id = ?`, orderID).Scan(&status).Error; err != nil {
		t.Fatalf("read order status: %v", err)
	}
	return status
}

func orderAPIReserved(t *testing.T, database *gorm.DB, skuID string) int64 {
	t.Helper()
	var reserved int64
	if err := database.Raw(`SELECT reserved FROM inventories WHERE sku_id = ?`, skuID).Scan(&reserved).Error; err != nil {
		t.Fatalf("read reserved: %v", err)
	}
	return reserved
}
