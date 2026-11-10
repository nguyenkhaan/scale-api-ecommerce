package backend_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/httpserver/middleware"
	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/module/flashsale"
	"github.com/gin-gonic/gin"
)

func TestFlashSaleControllerRejectsInvalidUUIDBeforeService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())
	api := router.Group("/api")
	flashsale.RegisterRoutes(api, flashsale.NewController(&flashsale.Service{}), nil, nil)

	testCases := []struct {
		name   string
		method string
		path   string
	}{
		{name: "public get invalid uuid", method: http.MethodGet, path: "/api/flash-sales/not-a-uuid"},
		{name: "admin get invalid uuid", method: http.MethodGet, path: "/api/admin/flash-sales/not-a-uuid"},
		{name: "admin update invalid uuid", method: http.MethodPatch, path: "/api/admin/flash-sales/not-a-uuid"},
		{name: "admin schedule invalid uuid", method: http.MethodPost, path: "/api/admin/flash-sales/not-a-uuid/schedule"},
		{name: "admin cancel invalid uuid", method: http.MethodPost, path: "/api/admin/flash-sales/not-a-uuid/cancel-schedule"},
		{name: "admin stop invalid uuid", method: http.MethodPost, path: "/api/admin/flash-sales/not-a-uuid/stop"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			requireProblem(t, rec, http.StatusUnprocessableEntity, problem.CodeValidation)
		})
	}
}

func TestFlashSaleCreateValidationRejectsMalformedBodies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())
	api := router.Group("/api")
	flashsale.RegisterRoutes(api, flashsale.NewController(&flashsale.Service{}), nil, nil)

	testCases := []struct {
		name string
		body string
	}{
		{name: "empty body", body: `{}`},
		{name: "missing items", body: `{"name":"Sale","starts_at":"2026-11-03T10:00:00Z","ends_at":"2026-11-03T12:00:00Z","items":[]}`},
		{name: "invalid item sku uuid", body: `{"name":"Sale","starts_at":"2026-11-03T10:00:00Z","ends_at":"2026-11-03T12:00:00Z","items":[{"sku_id":"invalid","sale_price":1000,"quota_total":10,"purchase_limit":1}]}`},
		{name: "negative sale price", body: `{"name":"Sale","starts_at":"2026-11-03T10:00:00Z","ends_at":"2026-11-03T12:00:00Z","items":[{"sku_id":"53000000-0000-4000-8000-000000000001","sale_price":-10,"quota_total":10,"purchase_limit":1}]}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/admin/flash-sales", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			requireProblem(t, rec, http.StatusUnprocessableEntity, problem.CodeValidation)
		})
	}
}

func TestFlashSaleStatusDerivation(t *testing.T) {
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	testCases := []struct {
		name       string
		status     string
		startsAt   time.Time
		endsAt     time.Time
		wantStatus string
	}{
		{
			name:       "draft remains draft",
			status:     "draft",
			startsAt:   now.Add(-time.Hour),
			endsAt:     now.Add(time.Hour),
			wantStatus: "draft",
		},
		{
			name:       "stopped remains stopped",
			status:     "stopped",
			startsAt:   now.Add(-time.Hour),
			endsAt:     now.Add(time.Hour),
			wantStatus: "stopped",
		},
		{
			name:       "scheduled before start",
			status:     "scheduled",
			startsAt:   now.Add(time.Hour),
			endsAt:     now.Add(2 * time.Hour),
			wantStatus: "scheduled",
		},
		{
			name:       "scheduled during time window is active",
			status:     "scheduled",
			startsAt:   now.Add(-time.Hour),
			endsAt:     now.Add(time.Hour),
			wantStatus: "active",
		},
		{
			name:       "scheduled after end is ended",
			status:     "scheduled",
			startsAt:   now.Add(-2 * time.Hour),
			endsAt:     now.Add(-time.Hour),
			wantStatus: "ended",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := flashsale.DeriveEffectiveStatus(tc.status, tc.startsAt, tc.endsAt, now)
			if got != tc.wantStatus {
				t.Errorf("DeriveEffectiveStatus() = %q, want %q", got, tc.wantStatus)
			}
		})
	}
}
