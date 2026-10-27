package backend_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"net/http"
	"strings"
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/auth"
	"cloudian/cloudian-scale/internal/httpserver/problem"
	"golang.org/x/crypto/bcrypt"
)

func TestRegisterCreatesCustomerSessionWithSecureCookies(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	client := newTestClient(newAuthRouter(t, testDatabase.ORM()))

	response := client.send(t, http.MethodPost, "/api/v1/auth/register", credentials("  Lan.Tran@Example.com ", testPassword))

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", response.Code, response.Body.String())
	}
	result := decodeJSON[auth.AuthResultResponse](t, response)
	if result.User.Role != "customer" || result.User.Email != "Lan.Tran@Example.com" {
		t.Errorf("user = %+v, want trimmed customer", result.User)
	}
	if len(result.CSRFToken) < 32 {
		t.Errorf("csrf_token length = %d, want at least 32", len(result.CSRFToken))
	}
	expiresAt, err := time.Parse(time.RFC3339, result.ExpiresAt)
	if err != nil || time.Until(expiresAt) < 6*24*time.Hour || time.Until(expiresAt) > 8*24*time.Hour {
		t.Errorf("expires_at = %q, want about seven days ahead", result.ExpiresAt)
	}
	if strings.Contains(response.Body.String(), testPassword) {
		t.Error("register response contains the password")
	}

	cookies := map[string]*http.Cookie{}
	for _, cookie := range response.Result().Cookies() {
		cookies[cookie.Name] = cookie
	}
	session, csrf := cookies[auth.SessionCookieName], cookies[auth.CSRFCookieName]
	if session == nil || !session.HttpOnly || !session.Secure || session.SameSite != http.SameSiteLaxMode ||
		session.Path != "/" || session.MaxAge != 7*24*60*60 {
		t.Errorf("session cookie = %+v, want HttpOnly Secure Lax for seven days", session)
	}
	if csrf == nil || csrf.HttpOnly || !csrf.Secure || csrf.SameSite != http.SameSiteStrictMode || csrf.Value != result.CSRFToken {
		t.Errorf("csrf cookie = %+v, want readable Secure Strict cookie equal to csrf_token", csrf)
	}

	var stored struct {
		EmailNormalized string
		PasswordHash    string
	}
	testDatabase.ORM().Raw("SELECT email_normalized, password_hash FROM users").Scan(&stored)
	if stored.EmailNormalized != "lan.tran@example.com" {
		t.Errorf("email_normalized = %q", stored.EmailNormalized)
	}
	if bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte(testPassword)) != nil {
		t.Error("stored password_hash is not a bcrypt hash of the password")
	}
	// Pins the storage contract: only purpose-labelled HMACs of the cookie values are persisted.
	var matchingSessions int64
	testDatabase.ORM().Raw(
		"SELECT count(*) FROM sessions WHERE token_hash = ? AND csrf_secret_hash = ?",
		labelledHash("salescale/session-token/v1", session.Value), labelledHash("salescale/csrf-token/v1", csrf.Value),
	).Scan(&matchingSessions)
	if matchingSessions != 1 {
		t.Errorf("sessions matching HMAC(cookie) = %d, want 1", matchingSessions)
	}

	me := client.send(t, http.MethodGet, "/api/v1/auth/me", nil)
	if user := decodeJSON[auth.UserResponse](t, me); me.Code != http.StatusOK || user != result.User {
		t.Errorf("GET /me = %d %+v, want %+v", me.Code, user, result.User)
	}
}

func TestRegisterRejectsDuplicateEmailIgnoringCase(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	router := newAuthRouter(t, testDatabase.ORM())
	newTestClient(router).send(t, http.MethodPost, "/api/v1/auth/register", credentials("lan@example.com", testPassword))

	response := newTestClient(router).send(t, http.MethodPost, "/api/v1/auth/register", credentials(" LAN@example.com", testPassword))

	requireProblem(t, response, http.StatusConflict, problem.CodeEmailAlreadyRegistered)
	if setCookie := response.Header().Get("Set-Cookie"); setCookie != "" {
		t.Errorf("rejected register set a cookie: %q", setCookie)
	}
}

func TestLoginDoesNotRevealWhetherEmailExists(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	router := newAuthRouter(t, testDatabase.ORM())
	newTestClient(router).send(t, http.MethodPost, "/api/v1/auth/register", credentials("lan@example.com", testPassword))

	wrongPassword := requireProblem(t, newTestClient(router).send(t, http.MethodPost, "/api/v1/auth/login",
		credentials("lan@example.com", "wrong password")), http.StatusUnauthorized, problem.CodeUnauthenticated)
	unknownEmail := requireProblem(t, newTestClient(router).send(t, http.MethodPost, "/api/v1/auth/login",
		credentials("nobody@example.com", "wrong password")), http.StatusUnauthorized, problem.CodeUnauthenticated)
	if wrongPassword.Detail != unknownEmail.Detail {
		t.Errorf("details differ: %q vs %q", wrongPassword.Detail, unknownEmail.Detail)
	}

	client := newTestClient(router)
	if response := client.send(t, http.MethodPost, "/api/v1/auth/login", credentials("LAN@example.com", testPassword)); response.Code != http.StatusOK {
		t.Fatalf("login status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	if client.cookies[auth.SessionCookieName] == "" {
		t.Error("login did not issue a session cookie")
	}
}

func TestDisabledUserCannotLogIn(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	router := newAuthRouter(t, testDatabase.ORM())
	newTestClient(router).send(t, http.MethodPost, "/api/v1/auth/register", credentials("lan@example.com", testPassword))
	testDatabase.ORM().Exec("UPDATE users SET status = 'disabled'")

	response := newTestClient(router).send(t, http.MethodPost, "/api/v1/auth/login", credentials("lan@example.com", testPassword))

	requireProblem(t, response, http.StatusUnauthorized, problem.CodeUnauthenticated)
	if setCookie := response.Header().Get("Set-Cookie"); setCookie != "" {
		t.Errorf("disabled user received a cookie: %q", setCookie)
	}
}

func labelledHash(label, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(label))
	mac.Write([]byte(secret))
	return mac.Sum(nil)
}
