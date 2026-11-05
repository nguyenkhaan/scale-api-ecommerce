package flashsale

import (
	"errors"
	"net/http"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"github.com/gin-gonic/gin"
)

// abortWithError maps service errors to the contract's problem responses; anything unknown is a 500.
func abortWithError(ctx *gin.Context, err error) {
	var validation *ValidationError
	var conflict *ConflictError
	switch {
	case errors.Is(err, ErrNotFound):
		problem.Abort(ctx, http.StatusNotFound, problem.CodeResourceNotFound, "Flash sale was not found.")
	case errors.Is(err, ErrInvalidCursor):
		problem.AbortValidation(ctx, []problem.Violation{violation("cursor", "is invalid or does not match the filters")})
	case errors.As(err, &validation):
		problem.AbortValidation(ctx, validation.Violations)
	case errors.As(err, &conflict):
		problem.Abort(ctx, http.StatusConflict, conflict.Code, conflict.Detail)
	default:
		problem.AbortInternal(ctx, err)
	}
}
