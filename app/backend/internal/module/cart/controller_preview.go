package cart

import (
	"net/http"
	"strconv"
	"strings"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"github.com/gin-gonic/gin"
	"uuid"
)

func (c *Controller) Preview(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return
	}

	var req CheckoutSelection
	if err := ctx.ShouldBindJSON(&req); err != nil {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "body", Reason: "malformed JSON request body"}})
		return
	}
	if violations := validateSelection(&req); len(violations) > 0 {
		problem.AbortValidation(ctx, violations)
		return
	}

	preview, err := c.service.PreviewCheckout(ctx.Request.Context(), userID, req)
	if err != nil {
		abortServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, preview)
}

// validateSelection enforces 1..20 unique items with versions >= 1 and lower-cases the ids so they
// match the canonical form stored in the database.
func validateSelection(req *CheckoutSelection) []problem.Violation {
	var violations []problem.Violation
	if len(req.Items) == 0 {
		violations = append(violations, problem.Violation{Field: "items", Reason: "must contain at least one item"})
	}
	if len(req.Items) > MaxCartItems {
		violations = append(violations, problem.Violation{Field: "items", Reason: "cannot exceed 20 items"})
	}

	seen := make(map[string]bool, len(req.Items))
	for i := range req.Items {
		prefix := "items[" + strconv.Itoa(i) + "]."
		item := &req.Items[i]
		if _, err := uuid.Parse(item.CartItemID); err != nil {
			violations = append(violations, problem.Violation{Field: prefix + "cart_item_id", Reason: "must be a valid UUID"})
		} else {
			item.CartItemID = strings.ToLower(item.CartItemID)
			if seen[item.CartItemID] {
				violations = append(violations, problem.Violation{Field: prefix + "cart_item_id", Reason: "duplicate cart item in selection"})
			}
			seen[item.CartItemID] = true
		}
		if item.Version < 1 {
			violations = append(violations, problem.Violation{Field: prefix + "version", Reason: "must be at least 1"})
		}
	}
	return violations
}
