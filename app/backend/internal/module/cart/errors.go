package cart

import (
	"errors"
	"strings"

	"cloudian/cloudian-scale/internal/httpserver/problem"
)

var (
	// ErrItemUnavailable means the SKU, its product or its category cannot be bought (BR-03).
	ErrItemUnavailable = errors.New("sku is not available for purchase")
	// ErrSaleNotActive means the flash sale campaign has ended, was stopped or is not scheduled.
	ErrSaleNotActive = errors.New("flash sale is not active")
	// ErrCartItemChanged means the caller's expected cart item version is stale.
	ErrCartItemChanged = errors.New("cart item changed")
)

// ValidationError carries field violations that map to a 422 response.
type ValidationError struct {
	Violations []problem.Violation
}

func (e *ValidationError) Error() string {
	reasons := make([]string, len(e.Violations))
	for i, violation := range e.Violations {
		reasons[i] = violation.Field + ": " + violation.Reason
	}
	return "validation failed: " + strings.Join(reasons, "; ")
}

func fieldError(field, reason string) *ValidationError {
	return &ValidationError{Violations: []problem.Violation{{Field: field, Reason: reason}}}
}

// LineErrors carries every failing checkout line; it maps to a 409 response.
type LineErrors struct {
	Lines []problem.LineProblem
}

func (e *LineErrors) Error() string {
	return "checkout contains lines that cannot be bought"
}
