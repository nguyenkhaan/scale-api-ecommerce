package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"cloudian/cloudian-scale/internal/config"
	"cloudian/cloudian-scale/internal/errors"
	"cloudian/cloudian-scale/internal/response"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newIntegrationRouter dựng router thật với đủ chuỗi middleware; test tích hợp bị bỏ qua khi chạy -short
func newIntegrationRouter(t *testing.T, allowedOrigins ...string) *gin.Engine {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	return SetupRouter(&config.Config{Env: "test", CORSAllowedOrigins: allowedOrigins})
}

func serve(r *gin.Engine, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for key, val := range headers {
		req.Header.Set(key, val)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestIntegrationRoutes(t *testing.T) {
	r := newIntegrationRouter(t)

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantCode   string
	}{
		{"ping", http.MethodGet, "/api/ping", http.StatusOK, ""},
		{"health", http.MethodGet, "/api/health", http.StatusOK, ""},
		{"unknown route", http.MethodGet, "/api/khong-ton-tai", http.StatusNotFound, errors.CodeNotFound},
		{"wrong method", http.MethodPost, "/api/ping", http.StatusMethodNotAllowed, errors.CodeMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := serve(r, tt.method, tt.path, map[string]string{"X-Request-ID": "client-req-id"})

			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if got := w.Header().Get("X-Request-ID"); got != "client-req-id" {
				t.Errorf("expected X-Request-ID header 'client-req-id', got '%s'", got)
			}

			var res response.Response
			if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if res.Meta.RequestID != "client-req-id" {
				t.Errorf("expected request_id 'client-req-id', got '%s'", res.Meta.RequestID)
			}

			if tt.wantCode == "" {
				if !res.Success {
					t.Errorf("expected success true, got false")
				}
				return
			}
			if res.Error == nil || res.Error.Code != tt.wantCode {
				t.Errorf("expected error code %s, got %+v", tt.wantCode, res.Error)
			}
		})
	}
}

func TestIntegrationCORS(t *testing.T) {
	const frontendOrigin = "http://localhost:5173"

	tests := []struct {
		name           string
		allowedOrigins []string
		origin         string
		wantOrigin     string
	}{
		{"allowed origin", []string{frontendOrigin}, frontendOrigin, frontendOrigin},
		{"unknown origin", []string{frontendOrigin}, "http://evil.example", ""},
		{"wildcard does not match every origin", []string{"*"}, "http://evil.example", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newIntegrationRouter(t, tt.allowedOrigins...)
			w := serve(r, http.MethodGet, "/api/ping", map[string]string{"Origin": tt.origin})

			if got := w.Header().Get("Access-Control-Allow-Origin"); got != tt.wantOrigin {
				t.Errorf("expected Access-Control-Allow-Origin '%s', got '%s'", tt.wantOrigin, got)
			}
		})
	}
}
