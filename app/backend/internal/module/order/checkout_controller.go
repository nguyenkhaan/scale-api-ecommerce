package order

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/module/cart"
	"github.com/gin-gonic/gin"
)

const checkoutRequestMaxBytes = 1 << 20

type CheckoutController struct {
	service *CheckoutService
}

func NewCheckoutController(service *CheckoutService) *CheckoutController {
	return &CheckoutController{service: service}
}

func RegisterCheckoutRoutes(api *gin.RouterGroup, controller *CheckoutController, requireSession gin.HandlerFunc) {
	checkout := api.Group("", requireSession)
	checkout.POST("/orders", controller.PlaceOrder)
	checkout.GET("/idempotency/orders/:idempotencyKey", controller.LookupAttempt)
}

func (c *CheckoutController) PlaceOrder(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return
	}
	var request PlaceOrderRequest
	decoder := json.NewDecoder(http.MaxBytesReader(ctx.Writer, ctx.Request.Body, checkoutRequestMaxBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "body", Reason: "must be valid checkout JSON"}})
		return
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "body", Reason: "must contain one JSON object"}})
		return
	}
	violations := ValidateIdempotencyKey(ctx.GetHeader("Idempotency-Key"))
	violations = append(violations, ValidatePlaceOrder(&request)...)
	if len(violations) > 0 {
		problem.AbortValidation(ctx, violations)
		return
	}
	result, err := c.service.PlaceOrder(ctx.Request.Context(), PlaceOrderInput{
		UserID: userID, IdempotencyKey: ctx.GetHeader("Idempotency-Key"),
		RequestID: ctx.Writer.Header().Get("X-Request-ID"), Request: request,
	})
	if err != nil {
		var validationError *cart.ValidationError
		switch {
		case errors.Is(err, ErrIdempotencyConflict):
			problem.Abort(ctx, http.StatusConflict, problem.CodeIdempotencyConflict,
				"This idempotency key was already used for a different checkout request.")
		case errors.Is(err, ErrIdempotencyInProgress):
			ctx.Header("Retry-After", "1")
			problem.Abort(ctx, http.StatusConflict, problem.CodeIdempotencyInProgress,
				"A checkout request with this key is still in progress.")
		case errors.As(err, &validationError):
			problem.AbortValidation(ctx, validationError.Violations)
		default:
			checkoutLogInternalError(ctx)
			problem.Abort(ctx, http.StatusInternalServerError, problem.CodeInternal, "The checkout could not be completed.")
		}
		return
	}
	contentType := "application/json"
	if result.Status == http.StatusConflict {
		contentType = problem.ContentType
	}
	ctx.Header("Idempotency-Replayed", "false")
	if result.Replayed {
		ctx.Header("Idempotency-Replayed", "true")
	}
	problem.WriteRaw(ctx, result.Status, contentType, result.Body, nil)
}

func (c *CheckoutController) LookupAttempt(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return
	}
	key := ctx.Param("idempotencyKey")
	if violations := ValidateIdempotencyKey(key); len(violations) > 0 {
		problem.AbortValidation(ctx, violations)
		return
	}
	attempt, err := c.service.LookupAttempt(ctx.Request.Context(), userID, key)
	if errors.Is(err, ErrAttemptNotFound) {
		problem.Abort(ctx, http.StatusNotFound, problem.CodeResourceNotFound, "Checkout attempt not found.")
		return
	}
	if err != nil {
		checkoutLogInternalError(ctx)
		problem.Abort(ctx, http.StatusInternalServerError, problem.CodeInternal, "The checkout attempt could not be loaded.")
		return
	}
	ctx.JSON(http.StatusOK, attempt)
}

func checkoutLogInternalError(ctx *gin.Context) {
	log.Printf("request_id=%s checkout internal error", ctx.Writer.Header().Get("X-Request-ID"))
}
