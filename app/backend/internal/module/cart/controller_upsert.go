package cart

import (
	"net/http"
	"strings"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/model"
	"github.com/gin-gonic/gin"
	"uuid"
)

func (c *Controller) UpsertItem(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return
	}
	skuID, ok := skuIDParam(ctx)
	if !ok {
		return
	}

	var req UpsertCartItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "body", Reason: "malformed JSON request body"}})
		return
	}
	if violations := validateUpsert(&req); len(violations) > 0 {
		problem.AbortValidation(ctx, violations)
		return
	}

	cart, err := c.service.UpsertCartItem(ctx.Request.Context(), userID, skuID, req)
	if err != nil {
		abortServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, cart)
}

// validateUpsert checks the request shape and normalizes defaults (price_source, UUID case).
// Rules that need stored data (stock caps, sale state) run in the service.
func validateUpsert(req *UpsertCartItemRequest) []problem.Violation {
	var violations []problem.Violation
	if req.PriceSource == "" {
		req.PriceSource = model.PriceSourceRegular
	}
	if req.Quantity < 1 {
		violations = append(violations, problem.Violation{Field: "quantity", Reason: "must be at least 1"})
	}
	if req.ExpectedVersion != nil && *req.ExpectedVersion < 1 {
		violations = append(violations, problem.Violation{Field: "expected_version", Reason: "must be at least 1"})
	}

	switch req.PriceSource {
	case model.PriceSourceRegular:
		if req.FlashSaleItemID != nil {
			violations = append(violations, problem.Violation{Field: "flash_sale_item_id", Reason: "must be null when price_source is regular"})
		}
	case model.PriceSourceFlashSale:
		if req.FlashSaleItemID == nil || *req.FlashSaleItemID == "" {
			violations = append(violations, problem.Violation{Field: "flash_sale_item_id", Reason: "is required when price_source is flash_sale"})
		} else if _, err := uuid.Parse(*req.FlashSaleItemID); err != nil {
			violations = append(violations, problem.Violation{Field: "flash_sale_item_id", Reason: "must be a valid UUID"})
		} else {
			lowered := strings.ToLower(*req.FlashSaleItemID)
			req.FlashSaleItemID = &lowered
		}
	default:
		violations = append(violations, problem.Violation{Field: "price_source", Reason: "must be regular or flash_sale"})
	}
	return violations
}
