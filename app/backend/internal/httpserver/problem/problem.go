package problem

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	ContentType                = "application/problem+json"
	CodeResourceNotFound       = "RESOURCE_NOT_FOUND"
	CodeNotFound               = CodeResourceNotFound
	CodeValidation             = "VALIDATION_ERROR"
	CodeUnauthenticated        = "UNAUTHENTICATED"
	CodeForbidden              = "FORBIDDEN"
	CodeEmailAlreadyRegistered = "EMAIL_ALREADY_REGISTERED"
	CodeRateLimited            = "RATE_LIMITED"
	CodeInternal               = "INTERNAL_ERROR"
	CodeInventoryBelowReserved = "INVENTORY_BELOW_RESERVED"
	CodeSKUAlreadyExists       = "SKU_CODE_ALREADY_EXISTS"

	CodeStateTransitionConflict = "STATE_TRANSITION_CONFLICT"
	CodeSaleScheduleOverlap     = "SALE_SCHEDULE_OVERLAP"
	CodeCartItemChanged         = "CART_ITEM_CHANGED"

	// Line-level checkout codes, highest priority first (BR-24).
	CodeItemUnavailable       = "ITEM_UNAVAILABLE"
	CodeSaleNotActive         = "SALE_NOT_ACTIVE"
	CodePriceChanged          = "PRICE_CHANGED"
	CodePurchaseLimitExceeded = "PURCHASE_LIMIT_EXCEEDED"
	CodeSaleQuotaExceeded     = "SALE_QUOTA_EXCEEDED"
	CodeInsufficientStock     = "INSUFFICIENT_STOCK"

	// Set on the response by the request ID middleware before any handler runs.
	requestIDHeader = "X-Request-ID"
)

type Violation struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

// LineProblem explains why one requested SKU line cannot be bought as confirmed.
type LineProblem struct {
	SKUID             string `json:"sku_id"`
	RequestedQuantity int    `json:"requested_quantity"`
	Code              string `json:"code"`
	AvailableQuantity *int64 `json:"available_quantity,omitempty"`
	CurrentUnitPrice  *int64 `json:"current_unit_price,omitempty"`
}

// Problem is the RFC 9457 error envelope shared by every API error response.
type Problem struct {
	Type       string        `json:"type"`
	Title      string        `json:"title"`
	Status     int           `json:"status"`
	Code       string        `json:"code"`
	Detail     string        `json:"detail,omitempty"`
	RequestID  string        `json:"request_id"`
	Violations []Violation   `json:"violations,omitempty"`
	LineErrors []LineProblem `json:"line_errors,omitempty"`
}

func Abort(context *gin.Context, status int, code, detail string) {
	write(context, Problem{Status: status, Code: code, Detail: detail})
}

// AbortLineErrors rejects a whole multi-line request (409) and reports every failing line.
// The top-level code is the most severe line code so clients that only read `code` still act correctly.
func AbortLineErrors(context *gin.Context, detail string, lineErrors []LineProblem) {
	write(context, Problem{
		Status:     http.StatusConflict,
		Code:       mostSevereLineCode(lineErrors),
		Detail:     detail,
		LineErrors: lineErrors,
	})
}

var lineCodeSeverity = map[string]int{
	CodeItemUnavailable:       0,
	CodeSaleNotActive:         1,
	CodePriceChanged:          2,
	CodePurchaseLimitExceeded: 3,
	CodeSaleQuotaExceeded:     4,
	CodeInsufficientStock:     5,
}

func mostSevereLineCode(lineErrors []LineProblem) string {
	best := ""
	for _, line := range lineErrors {
		if best == "" || lineCodeSeverity[line.Code] < lineCodeSeverity[best] {
			best = line.Code
		}
	}
	return best
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
