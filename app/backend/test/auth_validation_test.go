package backend_test

import (
	"net/http"
	"strings"
	"testing"

	"cloudian/cloudian-scale/internal/httpserver/problem"
)

func TestRegisterRejectsInvalidPayloads(t *testing.T) {
	client := newTestClient(newAuthRouterWithoutDatabase(t))
	testCases := []struct {
		name      string
		body      string
		wantField string
	}{
		{name: "missing email", body: `{"password":"long enough password"}`, wantField: "email"},
		{name: "malformed email", body: `{"email":"not-an-email","password":"long enough password"}`, wantField: "email"},
		{name: "email with display name", body: `{"email":"Lan <lan@example.com>","password":"long enough password"}`, wantField: "email"},
		{name: "non-ASCII email", body: `{"email":"lân@example.com","password":"long enough password"}`, wantField: "email"},
		{name: "short password", body: `{"email":"lan@example.com","password":"short"}`, wantField: "password"},
		{name: "password over 72 bytes", body: `{"email":"lan@example.com","password":"` + strings.Repeat("a", 73) + `"}`, wantField: "password"},
		{name: "client-chosen role", body: `{"email":"lan@example.com","password":"long enough password","role":"admin"}`, wantField: "body"},
		{name: "trailing JSON", body: `{"email":"lan@example.com","password":"long enough password"}{}`, wantField: "body"},
		{name: "not JSON", body: `email=lan@example.com`, wantField: "body"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			response := client.send(t, http.MethodPost, "/api/v1/auth/register", testCase.body)

			body := requireProblem(t, response, http.StatusUnprocessableEntity, problem.CodeValidation)
			if !hasViolation(body.Violations, testCase.wantField) {
				t.Errorf("violations = %+v, want one for %q", body.Violations, testCase.wantField)
			}
			if strings.Contains(response.Body.String(), "long enough password") {
				t.Error("validation response echoes the submitted password")
			}
		})
	}
}

func TestLoginRejectsMissingPassword(t *testing.T) {
	client := newTestClient(newAuthRouterWithoutDatabase(t))

	response := client.send(t, http.MethodPost, "/api/v1/auth/login", credentials("lan@example.com", ""))

	body := requireProblem(t, response, http.StatusUnprocessableEntity, problem.CodeValidation)
	if !hasViolation(body.Violations, "password") {
		t.Errorf("violations = %+v, want one for password", body.Violations)
	}
}

// A cross-site HTML form can post text/plain whose body parses as JSON; only application/json is accepted.
func TestCredentialRoutesRejectNonJSONContentType(t *testing.T) {
	client := newTestClient(newAuthRouterWithoutDatabase(t))
	formShapedBody := `{"email":"attacker@example.com","password":"attacker password="}` + "\r\n"

	for _, path := range []string{"/api/v1/auth/login", "/api/v1/auth/register"} {
		t.Run(path, func(t *testing.T) {
			response := client.send(t, http.MethodPost, path, formShapedBody, "Content-Type", "text/plain")

			body := requireProblem(t, response, http.StatusUnprocessableEntity, problem.CodeValidation)
			if !hasViolation(body.Violations, "Content-Type") {
				t.Errorf("violations = %+v, want one for Content-Type", body.Violations)
			}
			if setCookie := response.Header().Get("Set-Cookie"); setCookie != "" {
				t.Errorf("rejected request set a cookie: %q", setCookie)
			}
		})
	}
}

func TestPrivateAuthRoutesRequireSession(t *testing.T) {
	client := newTestClient(newAuthRouterWithoutDatabase(t))
	testCases := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/v1/auth/me"},
		{method: http.MethodPost, path: "/api/v1/auth/logout"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.method+" "+testCase.path, func(t *testing.T) {
			response := client.send(t, testCase.method, testCase.path, nil)

			requireProblem(t, response, http.StatusUnauthorized, problem.CodeUnauthenticated)
			if cacheControl := response.Header().Get("Cache-Control"); cacheControl != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store", cacheControl)
			}
		})
	}
}

func hasViolation(violations []problem.Violation, field string) bool {
	for _, violation := range violations {
		if violation.Field == field {
			return true
		}
	}
	return false
}
