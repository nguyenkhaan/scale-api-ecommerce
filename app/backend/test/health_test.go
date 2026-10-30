package backend_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"cloudian/cloudian-scale/internal/httpserver/middleware"
	"cloudian/cloudian-scale/internal/module/health"
	"github.com/gin-gonic/gin"
)

type databasePingerStub struct {
	err error
}

func (stub databasePingerStub) Ping(context.Context) error {
	return stub.err
}

func TestLivenessEndpoint(t *testing.T) {
	router := newHealthRouter(databasePingerStub{})
	response := performRequest(router, "/api/health/liveness")

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Header().Get(middleware.RequestIdHeader) == "" {
		t.Fatal("response is missing X-Request-ID")
	}
}

func TestReadinessEndpointWhenDatabaseIsAvailable(t *testing.T) {
	router := newHealthRouter(databasePingerStub{})
	response := performRequest(router, "/api/health/ready")

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if body := response.Body.String(); body != `{"status":"service up"}` {
		t.Errorf("body = %q, want service-up response", body)
	}
}

func TestReadinessEndpointWhenDatabaseIsUnavailable(t *testing.T) {
	router := newHealthRouter(databasePingerStub{err: errors.New("database unavailable")})
	response := performRequest(router, "/api/health/ready")

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
	if body := response.Body.String(); body != `{"status":"service down"}` {
		t.Errorf("body = %q, want service-down response", body)
	}
}

func newHealthRouter(database health.DatabasePinger) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware(), gin.Recovery())
	health.RegisterRoutes(router.Group("/api"), database)
	return router
}

func performRequest(router http.Handler, path string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, path, nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
