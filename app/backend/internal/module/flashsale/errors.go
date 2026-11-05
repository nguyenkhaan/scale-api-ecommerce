package flashsale

import (
	"errors"
	"strings"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound      = errors.New("flash sale campaign not found")
	ErrInvalidCursor = errors.New("invalid flash sale cursor")
)

// ValidationError carries field-level violations for a 422 response.
type ValidationError struct {
	Violations []problem.Violation
}

func (e *ValidationError) Error() string {
	reasons := make([]string, len(e.Violations))
	for i, v := range e.Violations {
		reasons[i] = v.Field + ": " + v.Reason
	}
	return "flash sale validation failed: " + strings.Join(reasons, "; ")
}

// ConflictError is a 409 with a stable contract code.
type ConflictError struct {
	Code   string
	Detail string
}

func (e *ConflictError) Error() string { return e.Code + ": " + e.Detail }

func stateConflict(detail string) error {
	return &ConflictError{Code: problem.CodeStateTransitionConflict, Detail: detail}
}

func scheduleOverlap(detail string) error {
	return &ConflictError{Code: problem.CodeSaleScheduleOverlap, Detail: detail}
}

func invalid(field, reason string) error {
	return &ValidationError{Violations: []problem.Violation{{Field: field, Reason: reason}}}
}

// PostgreSQL SQLSTATE values that map to business outcomes.
const (
	sqlStateForeignKeyViolation = "23503"
	sqlStateCheckViolation      = "23514"
)

func pgErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}
