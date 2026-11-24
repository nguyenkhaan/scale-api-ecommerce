package payment

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/module/auth"
	"github.com/gin-gonic/gin"
)

const (
	// maxCallbackBytes bounds the raw body read before the signature can be checked.
	maxCallbackBytes   = 64 << 10
	operatorKeyHeader  = "X-Operator-Key"
	operatorContextKey = "payment.operator"
)

var errTrailingData = errors.New("unexpected data after JSON value")

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

// RegisterRoutes mounts payment endpoints under the /api group. Without a simulator secret the callback and
// simulator routes are not mounted at all. Simulator control and the admin listing are 404 when the simulator
// is disabled, and otherwise need the operator key or an admin session.
func RegisterRoutes(api *gin.RouterGroup, controller *Controller, requireSession, requireAdmin gin.HandlerFunc) {
	api.POST("/orders/:orderId/payment-session", requireSession, controller.CreateSession)
	if controller.service.cfg.SimulatorSecret == "" {
		log.Printf("payment: PAYMENT_SIMULATOR_SECRET is empty; callback and simulator routes are not mounted")
		return
	}
	api.POST("/payment-callbacks/simulator", controller.ReceiveCallback)

	control := api.Group("/payment-simulator",
		controller.simulatorGate, skipForOperator(requireSession), skipForOperator(requireAdmin))
	control.POST("/events", controller.ScheduleEvent)
	control.GET("/sessions", controller.ListSessions)
}

func (c *Controller) CreateSession(ctx *gin.Context) {
	identity, ok := auth.CurrentIdentity(ctx)
	if !ok {
		problem.Abort(ctx, http.StatusUnauthorized, problem.CodeUnauthenticated, "A valid session is required.")
		return
	}
	orderID, ok := canonicalUUID(ctx.Param("orderId"))
	if !ok {
		problem.Abort(ctx, http.StatusNotFound, problem.CodeResourceNotFound, "Order not found.")
		return
	}
	session, created, err := c.service.CreateOrGetSession(ctx.Request.Context(), identity.UserID, orderID)
	switch {
	case errors.Is(err, ErrNotFound):
		problem.Abort(ctx, http.StatusNotFound, problem.CodeResourceNotFound, "Order not found.")
	case errors.Is(err, ErrNotPayable):
		problem.Abort(ctx, http.StatusConflict, problem.CodeOrderNotPayable, "The order can no longer be paid.")
	case err != nil:
		problem.AbortInternal(ctx, err)
	case created:
		ctx.JSON(http.StatusCreated, session)
	default:
		ctx.JSON(http.StatusOK, session)
	}
}

// ReceiveCallback authenticates by signature over the exact raw body; there is no session or CSRF here.
// The signature is checked before any field validation, so an unauthenticated caller learns nothing about
// the payload rules.
func (c *Controller) ReceiveCallback(ctx *gin.Context) {
	body, err := io.ReadAll(http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxCallbackBytes))
	var tooLarge *http.MaxBytesError
	switch {
	case errors.As(err, &tooLarge):
		problem.Abort(ctx, http.StatusRequestEntityTooLarge, problem.CodeValidation, "The callback body is too large.")
		return
	case err != nil:
		problem.AbortValidation(ctx, []problem.Violation{{Field: "body", Reason: "could not be read"}})
		return
	}
	parsed, violations := parseCallback(body)
	if !verifySignature(c.service.cfg.SimulatorSecret, ctx.GetHeader(SignatureHeader), body, time.Now()) {
		if len(violations) == 0 {
			if err := c.service.RecordRejectedDelivery(ctx.Request.Context(), parsed, parsed.contentHash()); err != nil {
				problem.AbortInternal(ctx, err)
				return
			}
		}
		problem.Abort(ctx, http.StatusUnauthorized, problem.CodeUnauthenticated, "The callback signature is invalid.")
		return
	}
	if len(violations) > 0 {
		problem.AbortValidation(ctx, violations)
		return
	}
	receipt, err := c.service.ApplyCallback(ctx.Request.Context(), parsed, parsed.contentHash())
	if err != nil {
		// Nothing was committed, so a 5xx makes the sender retry (BR-49).
		problem.AbortInternal(ctx, err)
		return
	}
	ctx.JSON(http.StatusAccepted, receipt)
}

func (c *Controller) ScheduleEvent(ctx *gin.Context) {
	var request ScheduleRequest
	body, err := io.ReadAll(http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxCallbackBytes))
	if err == nil {
		err = decodeStrict(body, &request)
	}
	if err != nil {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "body", Reason: "must be a valid SchedulePaymentEventRequest JSON object"}})
		return
	}
	plan, violations := validateSchedule(request)
	if len(violations) > 0 {
		problem.AbortValidation(ctx, violations)
		return
	}
	scheduled, err := c.service.ScheduleEvent(ctx.Request.Context(), plan)
	switch {
	case errors.Is(err, ErrNotFound):
		problem.AbortValidation(ctx, []problem.Violation{{Field: "payment_session_id", Reason: "does not exist"}})
	case err != nil:
		problem.AbortInternal(ctx, err)
	default:
		ctx.JSON(http.StatusAccepted, scheduled)
	}
}

func (c *Controller) ListSessions(ctx *gin.Context) {
	status := strings.TrimSpace(ctx.Query("status"))
	limit := defaultSessionListLimit
	var violations []problem.Violation
	if status != "" && !sessionStatuses[status] {
		violations = append(violations, problem.Violation{Field: "status", Reason: "must be pending, unknown, succeeded or failed"})
	}
	if raw := ctx.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > maxSessionListLimit {
			violations = append(violations, problem.Violation{Field: "limit", Reason: "must be between 1 and 200"})
		} else {
			limit = parsed
		}
	}
	if len(violations) > 0 {
		problem.AbortValidation(ctx, violations)
		return
	}
	items, err := c.service.ListSimulatorSessions(ctx.Request.Context(), status, limit)
	if err != nil {
		problem.AbortInternal(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"items": items})
}

// simulatorGate hides the simulator when disabled and recognises the operator key. A presented key that does
// not match is refused outright rather than falling back to a session.
func (c *Controller) simulatorGate(ctx *gin.Context) {
	if !c.service.SimulatorEnabled() {
		problem.Abort(ctx, http.StatusNotFound, problem.CodeResourceNotFound, "Resource not found.")
		return
	}
	presented := ctx.GetHeader(operatorKeyHeader)
	if presented == "" {
		return
	}
	if !operatorKeyMatches(c.service.cfg.OperatorKey, presented) {
		problem.Abort(ctx, http.StatusForbidden, problem.CodeForbidden, "The operator key is not valid.")
		return
	}
	ctx.Set(operatorContextKey, true)
}

// operatorKeyMatches compares digests in constant time; an unset key never matches.
func operatorKeyMatches(configured, presented string) bool {
	if configured == "" {
		return false
	}
	want, got := sha256.Sum256([]byte(configured)), sha256.Sum256([]byte(presented))
	return subtle.ConstantTimeCompare(want[:], got[:]) == 1
}

// skipForOperator runs the session/role middleware only for callers that did not present a valid operator key.
func skipForOperator(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.GetBool(operatorContextKey) {
			ctx.Next()
			return
		}
		handler(ctx)
	}
}
