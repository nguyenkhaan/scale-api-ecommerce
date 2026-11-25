package backend_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cloudian/cloudian-scale/internal/httpserver/middleware"
	"github.com/gin-gonic/gin"
)

func TestCORSMiddlewareAllowsPhaseThreeMutationPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.CORSMiddleware())
	router.PATCH("/api/test", func(context *gin.Context) { context.Status(http.StatusNoContent) })
	router.DELETE("/api/test", func(context *gin.Context) { context.Status(http.StatusNoContent) })

	request := httptest.NewRequest(http.MethodOptions, "/api/test", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("Access-Control-Request-Method", http.MethodPatch)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("allow origin = %q, want localhost frontend", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("allow credentials = %q, want true", got)
	}
	methods := response.Header().Get("Access-Control-Allow-Methods")
	for _, method := range []string{http.MethodPatch, http.MethodDelete, http.MethodOptions} {
		if !strings.Contains(methods, method) {
			t.Errorf("allow methods %q does not include %s", methods, method)
		}
	}
}

func TestCORSAllowsIdempotencyHeaders(t *testing.T) {
	fixture := newCheckoutAPIFixture(t, 10, 1)
	preflight := httptest.NewRequest(http.MethodOptions, "/api/orders", nil)
	preflight.Header.Set("Origin", "http://localhost:5173")
	preflight.Header.Set("Access-Control-Request-Method", http.MethodPost)
	preflight.Header.Set("Access-Control-Request-Headers", "content-type,idempotency-key,x-csrf-token")
	preflightResponse := httptest.NewRecorder()
	fixture.router.ServeHTTP(preflightResponse, preflight)
	if preflightResponse.Code != http.StatusOK || !strings.Contains(strings.ToLower(preflightResponse.Header().Get("Access-Control-Allow-Headers")), "idempotency-key") {
		t.Fatalf("preflight status/headers = %d/%q", preflightResponse.Code, preflightResponse.Header().Get("Access-Control-Allow-Headers"))
	}
	headers := append(fixture.customer.csrfHeader(), "Idempotency-Key", "cors-checkout-key-01", "Origin", "http://localhost:5173")
	response := fixture.customer.send(t, http.MethodPost, "/api/orders", fixture.request, headers...)
	if response.Code != http.StatusCreated {
		t.Fatalf("checkout status = %d; body = %s", response.Code, response.Body.String())
	}
	exposed := response.Header().Get("Access-Control-Expose-Headers")
	for _, name := range []string{"Idempotency-Replayed", "Retry-After", "X-Request-ID"} {
		if !strings.Contains(strings.ToLower(exposed), strings.ToLower(name)) {
			t.Errorf("exposed headers %q omit %s", exposed, name)
		}
	}
	lifecycleRequireInvariants(t, fixture.db)
}
