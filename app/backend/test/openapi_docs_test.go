package backend_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"cloudian/cloudian-scale/internal/httpserver"
	"github.com/gin-gonic/gin"
)

func TestRegisterOpenAPIDocsServesScalarAndOpenAPISpec(t *testing.T) {
	t.Setenv("GIN_MODE", gin.TestMode)
	specification, err := os.CreateTemp(t.TempDir(), "openapi-*.yaml")
	if err != nil {
		t.Fatalf("create temporary OpenAPI specification: %v", err)
	}
	if _, err := specification.WriteString("openapi: 3.1.0\ninfo:\n  title: test\n  version: test\npaths: {}\n"); err != nil {
		t.Fatalf("write temporary OpenAPI specification: %v", err)
	}
	if err := specification.Close(); err != nil {
		t.Fatalf("close temporary OpenAPI specification: %v", err)
	}

	router := gin.New()
	httpserver.RegisterOpenAPIDocs(router, specification.Name())

	for _, testCase := range []struct {
		path        string
		wantStatus  int
		wantContent string
	}{
		{path: "/docs", wantStatus: http.StatusOK, wantContent: "/docs/openapi.yaml"},
		{path: "/docs/", wantStatus: http.StatusOK, wantContent: "/docs/openapi.yaml"},
		{path: "/docs/index.html", wantStatus: http.StatusOK, wantContent: "/docs/openapi.yaml"},
		{path: "/docs/openapi.yaml", wantStatus: http.StatusOK, wantContent: "openapi: 3.1.0"},
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, testCase.path, nil))
		if response.Code != testCase.wantStatus {
			t.Errorf("GET %s status = %d, want %d", testCase.path, response.Code, testCase.wantStatus)
		}
		if !strings.Contains(response.Body.String(), testCase.wantContent) {
			t.Errorf("GET %s body does not contain %q", testCase.path, testCase.wantContent)
		}
	}
}
