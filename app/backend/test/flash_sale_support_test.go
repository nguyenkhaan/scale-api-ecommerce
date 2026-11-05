package backend_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/model"
	"cloudian/cloudian-scale/internal/module/flashsale"
	"gorm.io/gorm"
)

const flashSaleBasePath = "/api/v1/admin/flash-sales"

// flashSaleEnv is an admin API client over a freshly truncated database with one active product.
type flashSaleEnv struct {
	t         *testing.T
	db        *gorm.DB
	router    http.Handler
	admin     *testClient
	productID string
	skuCount  int
}

func newFlashSaleEnv(t *testing.T) *flashSaleEnv {
	t.Helper()
	orm := openIntegrationDatabase(t).ORM()
	router := newPhase3Router(t, orm)
	admin := newCatalogAdminClient(t, router, orm)
	_, productID := seedCatalogProduct(t, orm, true, model.ProductStatusActive)
	return &flashSaleEnv{t: t, db: orm, router: router, admin: admin, productID: productID}
}

// sku seeds an active SKU with the given regular price and returns its id.
func (e *flashSaleEnv) sku(price int64) string {
	e.t.Helper()
	e.skuCount++
	return seedCatalogSKU(e.t, e.db, e.productID, fmt.Sprintf("FS-%03d", e.skuCount), model.SKUStatusActive, price, 5, 20)
}

// call sends an admin request with the CSRF header. Use a per-goroutine client when calling concurrently.
func (e *flashSaleEnv) call(client *testClient, method, path string, body any) *httptest.ResponseRecorder {
	return client.send(e.t, method, path, body, client.csrfHeader()...)
}

func (e *flashSaleEnv) concurrentClient() *testClient { return e.admin.on(e.router) }

func flashSaleItemPayload(skuID string, salePrice int64) map[string]any {
	return map[string]any{"sku_id": skuID, "sale_price": salePrice, "quota_total": 10, "purchase_limit": 2}
}

func flashSalePayload(name string, startsIn, endsIn time.Duration, items ...map[string]any) map[string]any {
	return map[string]any{
		"name":      name,
		"starts_at": time.Now().Add(startsIn).UTC().Format(time.RFC3339),
		"ends_at":   time.Now().Add(endsIn).UTC().Format(time.RFC3339),
		"items":     items,
	}
}

// draft creates a draft with one item per SKU, sale price 1000 and the given window relative to now.
func (e *flashSaleEnv) draft(startsIn, endsIn time.Duration, skuIDs ...string) flashsale.FlashSaleResponse {
	e.t.Helper()
	items := make([]map[string]any, len(skuIDs))
	for i, id := range skuIDs {
		items[i] = flashSaleItemPayload(id, 1000)
	}
	res := e.call(e.admin, http.MethodPost, flashSaleBasePath, flashSalePayload("Campaign", startsIn, endsIn, items...))
	if res.Code != http.StatusCreated {
		e.t.Fatalf("create draft status = %d; body = %s", res.Code, res.Body.String())
	}
	return decodeJSON[flashsale.FlashSaleResponse](e.t, res)
}

func (e *flashSaleEnv) action(id, action string) *httptest.ResponseRecorder {
	return e.call(e.admin, http.MethodPost, flashSaleBasePath+"/"+id+"/"+action, nil)
}

func (e *flashSaleEnv) mustSchedule(id string) {
	e.t.Helper()
	if res := e.action(id, "schedule"); res.Code != http.StatusOK {
		e.t.Fatalf("schedule status = %d; body = %s", res.Code, res.Body.String())
	}
}

func (e *flashSaleEnv) storedStatus(id string) string {
	e.t.Helper()
	var status string
	if err := e.db.Raw("SELECT status FROM flash_sale_campaigns WHERE id = ?", id).Scan(&status).Error; err != nil {
		e.t.Fatalf("read status: %v", err)
	}
	return status
}

func (e *flashSaleEnv) count(query string, args ...any) int {
	e.t.Helper()
	var n int
	if err := e.db.Raw(query, args...).Scan(&n).Error; err != nil {
		e.t.Fatalf("count %q: %v", query, err)
	}
	return n
}

// flashSaleSetWindow moves a campaign window using SQL expressions evaluated by the database clock.
func flashSaleSetWindow(t *testing.T, db *gorm.DB, id, startsAtSQL, endsAtSQL string) {
	t.Helper()
	statement := fmt.Sprintf("UPDATE flash_sale_campaigns SET starts_at = %s, ends_at = %s WHERE id = ?", startsAtSQL, endsAtSQL)
	if err := db.Exec(statement, id).Error; err != nil {
		t.Fatalf("set campaign window: %v", err)
	}
}

func (e *flashSaleEnv) setWindow(id, startsAtSQL, endsAtSQL string) {
	e.t.Helper()
	flashSaleSetWindow(e.t, e.db, id, startsAtSQL, endsAtSQL)
}

func flashSaleViolationFields(p problem.Problem) []string {
	fields := make([]string, len(p.Violations))
	for i, v := range p.Violations {
		fields[i] = v.Field
	}
	return fields
}

func flashSaleHasViolation(p problem.Problem, field string) bool {
	for _, got := range flashSaleViolationFields(p) {
		if got == field {
			return true
		}
	}
	return false
}
