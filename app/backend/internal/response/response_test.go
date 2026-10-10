package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"cloudian/cloudian-scale/internal/errors"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestSuccessResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("request_id", "test-uuid-1234")

	data := map[string]string{"foo": "bar"}
	Success(c, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var res Response
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if !res.Success {
		t.Errorf("expected success true, got false")
	}
	if res.Meta.RequestID != "test-uuid-1234" {
		t.Errorf("expected request_id 'test-uuid-1234', got '%s'", res.Meta.RequestID)
	}
	if res.Meta.Timestamp == "" {
		t.Errorf("expected timestamp not to be empty")
	}
}

func TestErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("request_id", "test-err-req-id")

	appErr := errors.ErrNotFound("Order")
	Error(c, appErr)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var res Response
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if res.Success {
		t.Errorf("expected success false, got true")
	}
	if res.Error == nil {
		t.Fatalf("expected error object, got nil")
	}
	if res.Error.Code != errors.CodeNotFound {
		t.Errorf("expected code %s, got %s", errors.CodeNotFound, res.Error.Code)
	}
	if res.Meta.RequestID != "test-err-req-id" {
		t.Errorf("expected request_id 'test-err-req-id', got '%s'", res.Meta.RequestID)
	}
}

func TestSuccessWithPagination(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("request_id", "page-req-id")

	items := []string{"item1", "item2"}
	pagination := PaginationInfo{
		Page:       1,
		Limit:      10,
		TotalItems: 2,
		TotalPages: 1,
	}

	SuccessWithPagination(c, http.StatusOK, items, pagination)

	var res Response
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}

	if res.Meta.Pagination == nil {
		t.Fatalf("expected pagination meta, got nil")
	}
	if res.Meta.Pagination.TotalItems != 2 {
		t.Errorf("expected total_items 2, got %d", res.Meta.Pagination.TotalItems)
	}
}
