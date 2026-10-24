package backend_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/httpserver/middleware"
	"github.com/gin-gonic/gin"
)

var uuidV4Pattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestRequestIDMiddlewareGeneratesMissingID(t *testing.T) {
	response, capturedRequestID := serveRequestWithID(t, "")

	requestID := response.Header().Get(middleware.RequestIdHeader)
	if !uuidV4Pattern.MatchString(requestID) {
		t.Errorf("response request ID %q is not a UUID v4", requestID)
	}
	if capturedRequestID != requestID {
		t.Errorf("context request ID = %q, response header = %q", capturedRequestID, requestID)
	}
}

func TestRequestIDMiddlewarePreservesValidID(t *testing.T) {
	const incomingRequestID = "checkout.request-123"

	response, capturedRequestID := serveRequestWithID(t, incomingRequestID)

	if got := response.Header().Get(middleware.RequestIdHeader); got != incomingRequestID {
		t.Errorf("response request ID = %q, want %q", got, incomingRequestID)
	}
	if capturedRequestID != incomingRequestID {
		t.Errorf("context request ID = %q, want %q", capturedRequestID, incomingRequestID)
	}
}

func TestRequestIDMiddlewareReplacesUnsafeID(t *testing.T) {
	const unsafeRequestID = "secret value with spaces"

	response, _ := serveRequestWithID(t, unsafeRequestID)

	requestID := response.Header().Get(middleware.RequestIdHeader)
	if requestID == unsafeRequestID {
		t.Fatal("unsafe request ID was copied to the response")
	}
	if !uuidV4Pattern.MatchString(requestID) {
		t.Errorf("replacement request ID %q is not a UUID v4", requestID)
	}
}

func TestRequestIDMiddlewareAcceptsMaximumLength(t *testing.T) {
	incomingRequestID := strings.Repeat("a", 128)

	response, _ := serveRequestWithID(t, incomingRequestID)

	if got := response.Header().Get(middleware.RequestIdHeader); got != incomingRequestID {
		t.Errorf("response request ID length = %d, want the valid 128-character ID", len(got))
	}
}

func TestRequestLogFormatterIncludesIDWithoutSecrets(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/health/ready?token=do-not-log", nil)
	request.Header.Set("Authorization", "Bearer do-not-log")
	logLine := middleware.RequestLogFormatter(gin.LogFormatterParams{
		Request:    request,
		TimeStamp:  time.Date(2026, time.October, 24, 10, 0, 0, 0, time.UTC),
		StatusCode: http.StatusOK,
		Latency:    10 * time.Millisecond,
		Method:     http.MethodGet,
		Path:       "/api/health/ready",
		Keys: map[any]any{
			middleware.RequestIdKey: "request-123",
		},
	})

	if !strings.Contains(logLine, "request_id=request-123") {
		t.Errorf("log line does not contain request ID: %q", logLine)
	}
	if strings.Contains(logLine, "do-not-log") {
		t.Errorf("log line contains a query or header secret: %q", logLine)
	}
}

func serveRequestWithID(t *testing.T, incomingRequestID string) (*httptest.ResponseRecorder, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	var capturedRequestID string
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())
	router.GET("/test", func(context *gin.Context) {
		requestID, exists := context.Get(middleware.RequestIdKey)
		if !exists {
			t.Error("request ID is missing from Gin context")
			context.Status(http.StatusInternalServerError)
			return
		}
		capturedRequestID, _ = requestID.(string)
		context.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	if incomingRequestID != "" {
		request.Header.Set(middleware.RequestIdHeader, incomingRequestID)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	return response, capturedRequestID
}
