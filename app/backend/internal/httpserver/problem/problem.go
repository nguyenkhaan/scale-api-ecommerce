package problem

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)
//exception handler define for gin 
const (
	ContentType = "application/problem+json"
	CodeNotFound 			   = "NOT_FOUND"
	CodeValidation             = "VALIDATION_ERROR"
	CodeUnauthenticated        = "UNAUTHENTICATED"
	CodeForbidden              = "FORBIDDEN"
	CodeEmailAlreadyRegistered = "EMAIL_ALREADY_REGISTERED"
	//email sending 
	CodeRateLimited            = "RATE_LIMITED"
	CodeInternal               = "INTERNAL_ERROR"
	//using for inventory
	CodeInventoryBelowReserved = "INVENTORY_BELOW_RESERVED"

	// Set on the response by the request ID middleware before any handler runs.
	requestIDHeader = "X-Request-ID"
)

type Violation struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

// Problem is the RFC 9457 error envelope shared by every API error response.
type Problem struct {
	Type       string      `json:"type"`
	Title      string      `json:"title"`
	Status     int         `json:"status"`
	Code       string      `json:"code"`
	Detail     string      `json:"detail,omitempty"`
	RequestID  string      `json:"request_id"`
	Violations []Violation `json:"violations,omitempty"`
}

func Abort(context *gin.Context, status int, code, detail string) {
	write(context, Problem{Status: status, Code: code, Detail: detail})
}

func AbortValidation(context *gin.Context, violations []Violation) {
	write(context, Problem{
		Status:     http.StatusUnprocessableEntity,
		Code:       CodeValidation,
		Detail:     "The request contains invalid fields.",
		Violations: violations,
	})
}

// AbortInternal logs the cause server-side and returns a generic 500 so internals never reach the client.
func AbortInternal(context *gin.Context, err error) {
	log.Printf("request_id=%s internal error: %v", context.Writer.Header().Get(requestIDHeader), err)
	write(context, Problem{Status: http.StatusInternalServerError, Code: CodeInternal})
}

func write(context *gin.Context, body Problem) {
	body.Type = "about:blank"
	body.Title = http.StatusText(body.Status)
	body.RequestID = context.Writer.Header().Get(requestIDHeader)
	context.Header("Content-Type", ContentType)
	context.AbortWithStatusJSON(body.Status, body)
}
