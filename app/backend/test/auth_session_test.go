package backend_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"cloudian/cloudian-scale/internal/auth"
	"cloudian/cloudian-scale/internal/bootstrap"
	"cloudian/cloudian-scale/internal/httpserver/middleware"
	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/model"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func TestLogoutRequiresCSRFAndRevokesSessionServerSide(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	client := newTestClient(newAuthRouter(t, testDatabase.ORM()))
	client.send(t, http.MethodPost, "/api/v1/auth/register", credentials("lan@example.com", testPassword))
	staleClient := client.on(client.handler)
	sessionToken := client.cookies[auth.SessionCookieName]

	for name, headers := range map[string][]string{
		"missing token":         nil,
		"wrong token":           {auth.CSRFHeaderName, "not-the-csrf-token-not-the-csrf-token"},
		"session token as CSRF": {auth.CSRFHeaderName, sessionToken},
	} {
		response := client.send(t, http.MethodPost, "/api/v1/auth/logout", nil, headers...)
		if response.Code != http.StatusForbidden {
			t.Errorf("%s: status = %d, want 403", name, response.Code)
		}
	}
	if response := client.send(t, http.MethodGet, "/api/v1/auth/me", nil); response.Code != http.StatusOK {
		t.Fatalf("session stopped working after rejected logouts: %d", response.Code)
	}

	logout := client.send(t, http.MethodPost, "/api/v1/auth/logout", nil, client.csrfHeader()...)
	if logout.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d, want 204; body = %s", logout.Code, logout.Body.String())
	}
	if len(client.cookies) != 0 {
		t.Errorf("cookies after logout = %v, want both cleared", client.cookies)
	}
	// A copy of the old cookie must stop working too: logout revokes the session in PostgreSQL.
	requireProblem(t, staleClient.send(t, http.MethodGet, "/api/v1/auth/me", nil), http.StatusUnauthorized, problem.CodeUnauthenticated)
}

func TestSessionIsSharedAcrossInstancesAndLogoutAppliesToAll(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	instanceA, instanceB := newAuthRouter(t, testDatabase.ORM()), newAuthRouter(t, testDatabase.ORM())
	clientA := newTestClient(instanceA)
	clientA.send(t, http.MethodPost, "/api/v1/auth/register", credentials("lan@example.com", testPassword))
	clientB := clientA.on(instanceB)

	if response := clientB.send(t, http.MethodGet, "/api/v1/auth/me", nil); response.Code != http.StatusOK {
		t.Fatalf("instance B GET /me = %d, want 200", response.Code)
	}
	if response := clientB.send(t, http.MethodPost, "/api/v1/auth/logout", nil, clientB.csrfHeader()...); response.Code != http.StatusNoContent {
		t.Fatalf("instance B logout = %d, want 204", response.Code)
	}
	requireProblem(t, clientA.send(t, http.MethodGet, "/api/v1/auth/me", nil), http.StatusUnauthorized, problem.CodeUnauthenticated)
}

func TestExpiredOrDisabledSessionsAreRejected(t *testing.T) {
	testCases := map[string]string{
		"expired session": "UPDATE sessions SET expires_at = now() - interval '1 second'",
		"disabled user":   "UPDATE users SET status = 'disabled'",
	}
	for name, statement := range testCases {
		t.Run(name, func(t *testing.T) {
			testDatabase := openIntegrationDatabase(t)
			client := newTestClient(newAuthRouter(t, testDatabase.ORM()))
			client.send(t, http.MethodPost, "/api/v1/auth/register", credentials("lan@example.com", testPassword))

			if err := testDatabase.ORM().Exec(statement).Error; err != nil {
				t.Fatalf("apply %q: %v", statement, err)
			}

			requireProblem(t, client.send(t, http.MethodGet, "/api/v1/auth/me", nil), http.StatusUnauthorized, problem.CodeUnauthenticated)
		})
	}
}

// Covers the admin-route half of TEST-24; ownership checks arrive with the first customer-owned resources.
func TestCustomerCannotUseAdminRoute(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	service, err := auth.NewService(testDatabase.ORM(), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("create auth service: %v", err)
	}
	router := newAuthRouter(t, testDatabase.ORM())
	router.GET("/api/v1/admin/probe", middleware.RequireSession(service), middleware.RequireRole(model.RoleAdmin),
		func(context *gin.Context) { context.JSON(http.StatusOK, gin.H{"secret": "admin-only"}) })

	customer := newTestClient(router)
	customer.send(t, http.MethodPost, "/api/v1/auth/register", credentials("customer@example.com", testPassword))
	denied := customer.send(t, http.MethodGet, "/api/v1/admin/probe", nil)
	requireProblem(t, denied, http.StatusForbidden, problem.CodeForbidden)
	if strings.Contains(denied.Body.String(), "admin-only") {
		t.Error("forbidden response leaked admin data")
	}

	t.Setenv("ENV", "test")
	t.Setenv("ALLOW_ADMIN_SEED", "true")
	t.Setenv("ADMIN_EMAIL", "admin@example.com")
	t.Setenv("ADMIN_PASSWORD", "admin password long enough")
	seedConfig, err := bootstrap.LoadAdminSeedConfig()
	if err != nil {
		t.Fatalf("load admin seed config: %v", err)
	}
	if err := bootstrap.SeedAdmin(context.Background(), testDatabase.SQL(), seedConfig); err != nil {
		t.Fatalf("seed admin: %v", err)
	}

	admin := newTestClient(router)
	login := admin.send(t, http.MethodPost, "/api/v1/auth/login", credentials("admin@example.com", "admin password long enough"))
	if result := decodeJSON[auth.AuthResultResponse](t, login); login.Code != http.StatusOK || result.User.Role != model.RoleAdmin {
		t.Fatalf("seeded admin login = %d %+v, want 200 admin", login.Code, result.User)
	}
	if response := admin.send(t, http.MethodGet, "/api/v1/admin/probe", nil); response.Code != http.StatusOK {
		t.Errorf("admin probe = %d, want 200", response.Code)
	}
}
