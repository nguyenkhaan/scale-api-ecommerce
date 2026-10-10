package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"cloudian/cloudian-scale/internal/config"
	"cloudian/cloudian-scale/internal/response"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHealthCheck(t *testing.T) {
	cfg := &config.Config{
		Env: "test",
	}
	h := NewHealthHandler(cfg)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("request_id", "test-health-req-id")

	h.HealthCheck(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", w.Code)
	}

	var res response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !res.Success {
		t.Errorf("expected success true, got false")
	}

	dataMap, ok := res.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be map, got %T", res.Data)
	}

	if dataMap["status"] != "UP" {
		t.Errorf("expected status 'UP', got %v", dataMap["status"])
	}
	if dataMap["environment"] != "test" {
		t.Errorf("expected environment 'test', got %v", dataMap["environment"])
	}
}

func TestPing(t *testing.T) {
	cfg := &config.Config{}
	h := NewHealthHandler(cfg)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("request_id", "test-ping-req-id")

	h.Ping(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", w.Code)
	}

	var res response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !res.Success {
		t.Errorf("expected success true, got false")
	}
}
