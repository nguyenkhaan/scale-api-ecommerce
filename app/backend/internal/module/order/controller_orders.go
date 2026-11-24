package order

import (
	"errors"
	"net/http"
	"strconv"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/module/auth"
	"github.com/gin-gonic/gin"
	"uuid"
)

// OrdersController serves the customer order read model and cancellation.
type OrdersController struct {
	queries *OrderQueries
}

func NewOrdersController(queries *OrderQueries) *OrdersController {
	return &OrdersController{queries: queries}
}

func currentUserID(ctx *gin.Context) (string, bool) {
	identity, ok := auth.CurrentIdentity(ctx)
	if !ok {
		problem.Abort(ctx, http.StatusUnauthorized, problem.CodeUnauthenticated, "A valid session is required.")
		return "", false
	}
	return identity.UserID, true
}

// orderIDParam treats a malformed id like an unknown one so ids reveal nothing.
func orderIDParam(ctx *gin.Context) (string, bool) {
	parsed, err := uuid.Parse(ctx.Param("orderId"))
	if err != nil {
		problem.Abort(ctx, http.StatusNotFound, problem.CodeResourceNotFound, "Order not found.")
		return "", false
	}
	// Parse also accepts forms PostgreSQL rejects (urn:uuid:...); only the canonical text reaches SQL.
	return parsed.String(), true
}

func (c *OrdersController) List(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return
	}
	params := ListParams{Cursor: ctx.Query("cursor"), Status: ctx.Query("status"), Limit: DefaultPageLimit}
	if raw := ctx.Query("limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil {
			problem.AbortValidation(ctx, []problem.Violation{{Field: "limit", Reason: "must be an integer between 1 and 100"}})
			return
		}
		params.Limit = limit
	}
	page, err := c.queries.ListOrders(ctx.Request.Context(), userID, params)
	switch {
	case errors.Is(err, ErrInvalidLimit):
		problem.AbortValidation(ctx, []problem.Violation{{Field: "limit", Reason: "must be an integer between 1 and 100"}})
	case errors.Is(err, ErrInvalidStatus):
		problem.AbortValidation(ctx, []problem.Violation{{Field: "status", Reason: "must be a valid order status"}})
	case errors.Is(err, ErrInvalidCursor):
		problem.AbortValidation(ctx, []problem.Violation{{Field: "cursor", Reason: "is invalid for this query"}})
	case err != nil:
		problem.AbortInternal(ctx, err)
	default:
		ctx.JSON(http.StatusOK, page)
	}
}

func (c *OrdersController) Get(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return
	}
	orderID, ok := orderIDParam(ctx)
	if !ok {
		return
	}
	order, err := c.queries.GetOrder(ctx.Request.Context(), userID, orderID)
	if err != nil {
		abortOrderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, order)
}

func (c *OrdersController) Cancel(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return
	}
	orderID, ok := orderIDParam(ctx)
	if !ok {
		return
	}
	order, err := c.queries.CancelOrder(ctx.Request.Context(), userID, orderID)
	if err != nil {
		abortOrderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, order)
}

func abortOrderError(ctx *gin.Context, err error) {
	var conflict *StateConflictError
	switch {
	case errors.Is(err, ErrOrderNotFound):
		problem.Abort(ctx, http.StatusNotFound, problem.CodeResourceNotFound, "Order not found.")
	case errors.As(err, &conflict):
		problem.Abort(ctx, http.StatusConflict, problem.CodeStateTransitionConflict,
			"The order is already "+conflict.Status+" and cannot be cancelled.")
	default:
		problem.AbortInternal(ctx, err)
	}
}
