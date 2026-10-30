package backend_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/database"
	"cloudian/cloudian-scale/internal/httpserver"
	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/migration"
	"cloudian/cloudian-scale/internal/module/auth"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const testPassword = "correct horse battery"

// newAuthRouterWithoutDatabase serves requests that must be rejected before any database access.
func newAuthRouterWithoutDatabase(t *testing.T) *gin.Engine {
	t.Helper()
	orm, err := gorm.Open(
		postgres.New(postgres.Config{DSN: "postgres://unused@127.0.0.1:1/unused"}),
		&gorm.Config{DisableAutomaticPing: true},
	)
	if err != nil {
		t.Fatalf("open placeholder ORM: %v", err)
	}
	return newAuthRouter(t, orm)
}

// openIntegrationDatabase migrates and empties the database named by TEST_DATABASE_URL.
func openIntegrationDatabase(t *testing.T) *database.Database {
	t.Helper()
	connectionString := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if connectionString == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	if connectionString == strings.TrimSpace(os.Getenv("DATABASE_URL")) {
		t.Fatal("TEST_DATABASE_URL must name a dedicated test database because tests truncate it")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	testDatabase, err := database.Open(ctx, connectionString)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { testDatabase.Close() })
	if err := migration.Run(ctx, testDatabase.SQL(), migration.CommandUp); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	if err := testDatabase.ORM().Exec(
		"TRUNCATE inventory_adjustments, inventories, skus, products, categories, flash_sale_customer_usage, flash_sale_items, flash_sale_campaigns, login_throttles, sessions, users",
	).Error; err != nil {
		t.Fatalf("truncate test database: %v", err)
	}
	return testDatabase
}

func newAuthRouter(t *testing.T, orm *gorm.DB) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	service, err := auth.NewService(orm, bcrypt.MinCost)
	if err != nil {
		t.Fatalf("create auth service: %v", err)
	}
	return httpserver.NewRouter(httpserver.Dependencies{
		Database:      databasePingerStub{},
		AuthService:   service,
		SecureCookies: true,
	})
}

// testClient keeps cookies between requests like a browser, so a session can move between API instances.
type testClient struct {
	handler http.Handler
	cookies map[string]string
}

func newTestClient(handler http.Handler) *testClient {
	return &testClient{handler: handler, cookies: map[string]string{}}
}

func (client *testClient) on(handler http.Handler) *testClient {
	copied := newTestClient(handler)
	for name, value := range client.cookies {
		copied.cookies[name] = value
	}
	return copied
}

func (client *testClient) csrfHeader() []string {
	return []string{auth.CSRFHeaderName, client.cookies[auth.CSRFCookieName]}
}

// send issues a request; body may be a raw JSON string or a value to marshal, headers are name/value pairs.
func (client *testClient) send(t *testing.T, method, path string, body any, headers ...string) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	switch value := body.(type) {
	case nil:
	case string:
		reader = strings.NewReader(value)
	default:
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("encode request body: %v", err)
		}
		reader = strings.NewReader(string(encoded))
	}
	request := httptest.NewRequest(method, path, reader)
	request.Header.Set("Content-Type", "application/json")
	for index := 0; index+1 < len(headers); index += 2 {
		request.Header.Set(headers[index], headers[index+1])
	}
	for name, value := range client.cookies {
		request.AddCookie(&http.Cookie{Name: name, Value: value})
	}

	response := httptest.NewRecorder()
	client.handler.ServeHTTP(response, request)
	for _, cookie := range response.Result().Cookies() {
		if cookie.MaxAge < 0 || cookie.Value == "" {
			delete(client.cookies, cookie.Name)
		} else {
			client.cookies[cookie.Name] = cookie.Value
		}
	}
	return response
}

func credentials(email, password string) map[string]string {
	return map[string]string{"email": email, "password": password}
}

func decodeJSON[T any](t *testing.T, response *httptest.ResponseRecorder) T {
	t.Helper()
	var body T
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response %q: %v", response.Body.String(), err)
	}
	return body
}

// requireProblem checks the shared error envelope and returns it for further assertions.
func requireProblem(t *testing.T, response *httptest.ResponseRecorder, wantStatus int, wantCode string) problem.Problem {
	t.Helper()
	if response.Code != wantStatus {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, wantStatus, response.Body.String())
	}
	if contentType := response.Header().Get("Content-Type"); contentType != problem.ContentType {
		t.Errorf("Content-Type = %q, want %q", contentType, problem.ContentType)
	}
	body := decodeJSON[problem.Problem](t, response)
	if body.Code != wantCode || body.Status != wantStatus {
		t.Errorf("problem = %s/%d, want %s/%d", body.Code, body.Status, wantCode, wantStatus)
	}
	if body.RequestID == "" || body.RequestID != response.Header().Get("X-Request-ID") {
		t.Errorf("problem request_id = %q, response header = %q", body.RequestID, response.Header().Get("X-Request-ID"))
	}
	return body
}
