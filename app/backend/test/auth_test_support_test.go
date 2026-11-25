package backend_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/database"
	"cloudian/cloudian-scale/internal/httpserver"
	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/migration"
	"cloudian/cloudian-scale/internal/module/auth"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const testPassword = "correct horse battery"

const integrationDatabaseTruncate = "TRUNCATE cart_items, carts, inventory_adjustments, inventories, skus, products, categories, flash_sale_customer_usage, flash_sale_items, flash_sale_campaigns, login_throttles, sessions, users CASCADE"

var integrationSchemaState struct {
	sync.Mutex
	admin            *database.Database
	schemaName       string
	connectionString string
}

func TestMain(m *testing.M) {
	code := m.Run()
	cleanupIntegrationSchema()
	os.Exit(code)
}

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

// openIntegrationDatabase migrates one isolated schema on DATABASE_URL and empties it per test.
func openIntegrationDatabase(t *testing.T) *database.Database {
	t.Helper()
	connectionString := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if connectionString == "" {
		_ = godotenv.Load("../.env")
		connectionString = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}
	if connectionString == "" {
		t.Skip("set DATABASE_URL or configure it in ../.env to run PostgreSQL integration tests")
	}
	testConnectionString := ensureIntegrationSchema(t, connectionString)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	testDatabase, err := database.Open(ctx, testConnectionString)
	if err != nil {
		t.Fatalf("open isolated integration schema: %v", err)
	}
	t.Cleanup(func() {
		if err := testDatabase.Close(); err != nil {
			t.Errorf("close isolated integration database: %v", err)
		}
	})
	if err := testDatabase.ORM().Exec(integrationDatabaseTruncate).Error; err != nil {
		t.Fatalf("truncate isolated integration schema: %v", err)
	}
	return testDatabase
}

func ensureIntegrationSchema(t *testing.T, connectionString string) string {
	t.Helper()
	integrationSchemaState.Lock()
	defer integrationSchemaState.Unlock()
	if integrationSchemaState.admin != nil {
		return integrationSchemaState.connectionString
	}

	var schemaToken [16]byte
	if _, err := rand.Read(schemaToken[:]); err != nil {
		t.Fatalf("generate integration schema name: %v", err)
	}
	schemaName := "scalingapi_test_" + hex.EncodeToString(schemaToken[:])
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	adminDatabase, err := database.Open(ctx, connectionString)
	if err != nil {
		t.Fatalf("open DATABASE_URL for isolated test schema: %v", err)
	}
	if _, err := adminDatabase.SQL().ExecContext(ctx, `CREATE SCHEMA "`+schemaName+`"`); err != nil {
		adminDatabase.Close()
		t.Fatalf("create isolated integration schema: %v", err)
	}
	testConnectionString, err := integrationSchemaConnectionString(connectionString, schemaName)
	if err != nil {
		_ = adminDatabase.Close()
		t.Fatalf("configure isolated integration schema connection: %v", err)
	}
	testDatabase, err := database.Open(ctx, testConnectionString)
	if err != nil {
		_ = adminDatabase.Close()
		t.Fatalf("open isolated integration schema: %v", err)
	}
	defer testDatabase.Close()
	var currentSchema, searchPath string
	if err := testDatabase.ORM().Raw("SELECT current_schema(), current_setting('search_path')").Row().Scan(&currentSchema, &searchPath); err != nil {
		_ = adminDatabase.Close()
		t.Fatalf("verify isolated integration schema: %v", err)
	}
	if currentSchema != schemaName {
		_ = adminDatabase.Close()
		t.Fatalf("integration connection selected schema %q with search_path %q, want isolated schema", currentSchema, searchPath)
	}
	if err := migration.Run(ctx, testDatabase.SQL(), migration.CommandUp); err != nil {
		_ = adminDatabase.Close()
		t.Fatalf("migrate isolated integration schema: %v", err)
	}
	var migratedTables int64
	if err := testDatabase.ORM().Raw("SELECT count(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = 'orders'", schemaName).Scan(&migratedTables).Error; err != nil || migratedTables != 1 {
		_ = adminDatabase.Close()
		t.Fatalf("isolated integration schema was not migrated (orders tables = %d, err = %v)", migratedTables, err)
	}
	integrationSchemaState.admin = adminDatabase
	integrationSchemaState.schemaName = schemaName
	integrationSchemaState.connectionString = testConnectionString
	return testConnectionString
}

func cleanupIntegrationSchema() {
	integrationSchemaState.Lock()
	defer integrationSchemaState.Unlock()
	if integrationSchemaState.admin == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, _ = integrationSchemaState.admin.SQL().ExecContext(ctx, `DROP SCHEMA IF EXISTS "`+integrationSchemaState.schemaName+`" CASCADE`)
	_ = integrationSchemaState.admin.Close()
}

func integrationSchemaConnectionString(connectionString, schemaName string) (string, error) {
	parsed, err := url.Parse(connectionString)
	if err != nil {
		return "", fmt.Errorf("parse DATABASE_URL: %w", err)
	}
	if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return "", fmt.Errorf("DATABASE_URL must use a PostgreSQL URL")
	}
	// Neon pooler connections ignore startup search_path parameters; a direct connection to the
	// same database is needed to keep each test's schema isolated across the connection pool.
	parsed.Host = strings.Replace(parsed.Host, "-pooler.", ".", 1)
	query := parsed.Query()
	query.Set("options", "-c search_path="+schemaName)
	parsed.RawQuery = strings.ReplaceAll(query.Encode(), "+", "%20")
	return parsed.String(), nil
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
