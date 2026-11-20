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
