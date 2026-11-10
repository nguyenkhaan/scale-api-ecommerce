package backend_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"cloudian/cloudian-scale/internal/httpserver"
	"cloudian/cloudian-scale/internal/model"
	"cloudian/cloudian-scale/internal/module/auth"
	"cloudian/cloudian-scale/internal/module/cart"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const cartCustomerPassword = "customer password long enough"

var cartSeedCounter atomic.Int64

// cartNewRouter serves only the auth and cart modules so cart tests do not depend on other modules.
func cartNewRouter(t *testing.T, database *gorm.DB) http.Handler {
	t.Helper()
	authService, err := auth.NewService(database, bcrypt.MinCost)
	if err != nil {
		t.Fatalf("create auth service: %v", err)
	}
	cartService, err := cart.NewService(database)
	if err != nil {
		t.Fatalf("create cart service: %v", err)
	}
	return httpserver.NewRouter(httpserver.Dependencies{
		Database:      databasePingerStub{},
		AuthService:   authService,
		CartService:   cartService,
		SecureCookies: true,
	})
}

type cartCustomer struct {
	*testClient
	ID string
}

// cartNewCustomer inserts an active customer and logs in.
func cartNewCustomer(t *testing.T, router http.Handler, database *gorm.DB, name string) cartCustomer {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(cartCustomerPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	email := fmt.Sprintf("%s-%d@example.com", strings.ReplaceAll(name, " ", "-"), cartSeedCounter.Add(1))
	var userID string
	if err := database.Raw(
		`INSERT INTO users (id, email, email_normalized, password_hash, role, status)
		 VALUES (gen_random_uuid(), ?, ?, ?, 'customer', 'active') RETURNING id`,
		email, strings.ToLower(email), string(hash),
	).Scan(&userID).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	client := newTestClient(router)
	if response := client.send(t, http.MethodPost, "/api/auth/login", credentials(email, cartCustomerPassword)); response.Code != http.StatusOK {
		t.Fatalf("customer login status = %d; body = %s", response.Code, response.Body.String())
	}
	return cartCustomer{testClient: client, ID: userID}
}

// parallel returns an independent client sharing this customer's session, safe for another goroutine.
func (customer cartCustomer) parallel(router http.Handler) *testClient {
	return customer.on(router)
}

type cartSKUOptions struct {
	CategoryHidden  bool
	ProductInactive bool
	SKUInactive     bool
	Price           int64
	MaxQuantity     int
	OnHand          int64
}

// cartSeedSKU creates category, product, SKU and inventory with unique identifiers.
func cartSeedSKU(t *testing.T, database *gorm.DB, options cartSKUOptions) string {
	t.Helper()
	if options.Price == 0 {
		options.Price = 100_000
	}
	if options.MaxQuantity == 0 {
		options.MaxQuantity = 10
	}
	if options.OnHand == 0 {
		options.OnHand = 100
	}
	n := cartSeedCounter.Add(1)
	productStatus, skuStatus := model.ProductStatusActive, model.SKUStatusActive
	if options.ProductInactive {
		productStatus = model.ProductStatusInactive
	}
	if options.SKUInactive {
		skuStatus = model.SKUStatusInactive
	}

	var categoryID, productID, skuID string
	if err := database.Raw(`INSERT INTO categories (name, slug, is_visible) VALUES (?, ?, ?) RETURNING id`,
		fmt.Sprintf("Category %d", n), fmt.Sprintf("cart-category-%d", n), !options.CategoryHidden).Scan(&categoryID).Error; err != nil {
		t.Fatalf("seed category: %v", err)
	}
	if err := database.Raw(`INSERT INTO products (category_id, name, status) VALUES (?, ?, ?) RETURNING id`,
		categoryID, fmt.Sprintf("Product %d", n), productStatus).Scan(&productID).Error; err != nil {
		t.Fatalf("seed product: %v", err)
	}
	if err := database.Raw(`INSERT INTO skus (product_id, code, regular_price, max_quantity_per_order, status)
		VALUES (?, ?, ?, ?, ?) RETURNING id`,
		productID, fmt.Sprintf("CART-SKU-%d", n), options.Price, options.MaxQuantity, skuStatus).Scan(&skuID).Error; err != nil {
		t.Fatalf("seed sku: %v", err)
	}
	if err := database.Exec(`INSERT INTO inventories (sku_id, on_hand) VALUES (?, ?)`, skuID, options.OnHand).Error; err != nil {
		t.Fatalf("seed inventory: %v", err)
	}
	return skuID
}

type cartSaleOptions struct {
	Status        string // defaults to scheduled
	StartsInSecs  float64
	EndsInSecs    float64 // defaults to one hour from now
	SalePrice     int64   // defaults to 40_000
	QuotaTotal    int64   // defaults to 10
	QuotaSold     int64
	PurchaseLimit int // defaults to 2
}

// cartSeedSale inserts a campaign and one item by direct SQL, timed against the database clock.
func cartSeedSale(t *testing.T, database *gorm.DB, skuID string, options cartSaleOptions) string {
	t.Helper()
	if options.Status == "" {
		options.Status = model.FlashSaleStatusScheduled
	}
	if options.EndsInSecs == 0 {
		options.EndsInSecs = 3600
	}
	if options.StartsInSecs == 0 {
		options.StartsInSecs = -3600
	}
	if options.SalePrice == 0 {
		options.SalePrice = 40_000
	}
	if options.QuotaTotal == 0 {
		options.QuotaTotal = 10
	}
	if options.PurchaseLimit == 0 {
		options.PurchaseLimit = 2
	}
	var campaignID, itemID string
	if err := database.Raw(
		`INSERT INTO flash_sale_campaigns (name, status, starts_at, ends_at)
		 VALUES (?, ?, now() + ? * interval '1 second', now() + ? * interval '1 second') RETURNING id`,
		fmt.Sprintf("Cart sale %d", cartSeedCounter.Add(1)), options.Status, options.StartsInSecs, options.EndsInSecs,
	).Scan(&campaignID).Error; err != nil {
		t.Fatalf("seed campaign: %v", err)
	}
	if err := database.Raw(
		`INSERT INTO flash_sale_items (campaign_id, sku_id, sale_price, quota_total, quota_sold, purchase_limit)
		 VALUES (?, ?, ?, ?, ?, ?) RETURNING id`,
		campaignID, skuID, options.SalePrice, options.QuotaTotal, options.QuotaSold, options.PurchaseLimit,
	).Scan(&itemID).Error; err != nil {
		t.Fatalf("seed flash sale item: %v", err)
	}
	return itemID
}

func cartExec(t *testing.T, database *gorm.DB, sql string, args ...any) {
	t.Helper()
	if err := database.Exec(sql, args...).Error; err != nil {
		t.Fatalf("exec %q: %v", sql, err)
	}
}

func cartCount(t *testing.T, database *gorm.DB, sql string, args ...any) int64 {
	t.Helper()
	var count int64
	if err := database.Raw(sql, args...).Scan(&count).Error; err != nil {
		t.Fatalf("count %q: %v", sql, err)
	}
	return count
}

func cartPut(t *testing.T, client *testClient, skuID string, body any) *httptest.ResponseRecorder {
	t.Helper()
	return client.send(t, http.MethodPut, "/api/cart/items/"+skuID, body, client.csrfHeader()...)
}

// cartAdd upserts a regular line and returns the resulting item, failing the test on a non-200.
func cartAdd(t *testing.T, client *testClient, skuID string, quantity int) cart.CartItemResponse {
	t.Helper()
	response := cartPut(t, client, skuID, map[string]any{"quantity": quantity})
	if response.Code != http.StatusOK {
		t.Fatalf("add %s status = %d; body = %s", skuID, response.Code, response.Body.String())
	}
	return cartFindItem(t, decodeJSON[cart.CartResponse](t, response), skuID)
}

func cartFindItem(t *testing.T, view cart.CartResponse, skuID string) cart.CartItemResponse {
	t.Helper()
	for _, item := range view.Items {
		if item.SKUID == skuID {
			return item
		}
	}
	t.Fatalf("sku %s not in cart %+v", skuID, view.Items)
	return cart.CartItemResponse{}
}

func cartGet(t *testing.T, client *testClient) cart.CartResponse {
	t.Helper()
	response := client.send(t, http.MethodGet, "/api/cart", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("get cart status = %d; body = %s", response.Code, response.Body.String())
	}
	return decodeJSON[cart.CartResponse](t, response)
}

func cartPreview(t *testing.T, client *testClient, items ...cart.CartItemResponse) *httptest.ResponseRecorder {
	t.Helper()
	selected := make([]map[string]any, len(items))
	for i, item := range items {
		selected[i] = map[string]any{"cart_item_id": item.ID, "version": item.Version}
	}
	return client.send(t, http.MethodPost, "/api/checkout/preview", map[string]any{"items": selected}, client.csrfHeader()...)
}

// cartRunParallel runs fn(i) for i in [0,n) at the same moment and waits for all of them.
func cartRunParallel(n int, fn func(i int)) {
	var ready, done sync.WaitGroup
	start := make(chan struct{})
	ready.Add(n)
	done.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer done.Done()
			ready.Done()
			<-start
			fn(i)
		}(i)
	}
	ready.Wait()
	close(start)
	done.Wait()
}
