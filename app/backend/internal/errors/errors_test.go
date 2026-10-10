package errors

import (
	"fmt"
	"net/http"
	"testing"
)

func TestNewAppError(t *testing.T) {
	err := NewAppError(http.StatusBadRequest, CodeBadRequest, "Invalid payload")
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, err.StatusCode)
	}
	if err.Code != CodeBadRequest {
		t.Errorf("expected code %s, got %s", CodeBadRequest, err.Code)
	}
	if err.Message != "Invalid payload" {
		t.Errorf("expected message %s, got %s", "Invalid payload", err.Message)
	}
}

func TestErrNotFound(t *testing.T) {
	err := ErrNotFound("Product")
	if err.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", err.StatusCode)
	}
	if err.Code != CodeNotFound {
		t.Errorf("expected NOT_FOUND, got %s", err.Code)
	}
}

func TestErrInternal(t *testing.T) {
	rootErr := fmt.Errorf("database connection lost")
	err := ErrInternal(rootErr)

	if err.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", err.StatusCode)
	}
	if err.Code != CodeInternalError {
		t.Errorf("expected INTERNAL_SERVER_ERROR, got %s", err.Code)
	}
	if err.Err != rootErr {
		t.Errorf("expected root error to be preserved")
	}
}
