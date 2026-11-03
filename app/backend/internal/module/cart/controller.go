package cart

import (
	"errors"
	"net/http"
	"uuid"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/module/auth"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

func RegisterRoutes(
	api *gin.RouterGroup,
	controller *Controller,
	requireSession gin.HandlerFunc,
) {
	cartGroup := api.Group("/v1/cart", requireSession)
	cartGroup.GET("", controller.GetCart)
	cartGroup.PUT("/items/:skuId", controller.UpsertItem)
	cartGroup.DELETE("/items/:skuId", controller.RemoveItem)

	checkoutGroup := api.Group("/v1/checkout", requireSession)
	checkoutGroup.POST("/preview", controller.Preview)
}

func (c *Controller) GetCart(ctx *gin.Context) {
	identity, ok := auth.CurrentIdentity(ctx)
	if !ok {
		problem.Abort(ctx, http.StatusUnauthorized, problem.CodeUnauthenticated, "A valid session is required.")
		return
	}

	cart, err := c.service.GetCart(ctx.Request.Context(), identity.UserID)
	if err != nil {
		problem.AbortInternal(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, cart)
}

func (c *Controller) UpsertItem(ctx *gin.Context) {
	identity, ok := auth.CurrentIdentity(ctx)
	if !ok {
		problem.Abort(ctx, http.StatusUnauthorized, problem.CodeUnauthenticated, "A valid session is required.")
		return
	}

	skuID := ctx.Param("skuId")
	if _, err := uuid.Parse(skuID); err != nil {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "skuId", Reason: "must be a valid UUID"}})
		return
	}

	var req UpsertCartItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "body", Reason: "malformed JSON request body"}})
		return
	}

	var violations []problem.Violation
	if req.Quantity <= 0 {
		violations = append(violations, problem.Violation{Field: "quantity", Reason: "must be greater than 0"})
	}
	if req.PriceSource != "regular" && req.PriceSource != "flash_sale" {
		violations = append(violations, problem.Violation{Field: "price_source", Reason: "must be regular or flash_sale"})
	}
	if req.PriceSource == "flash_sale" {
		if req.FlashSaleItemID == nil || *req.FlashSaleItemID == "" {
			violations = append(violations, problem.Violation{Field: "flash_sale_item_id", Reason: "is required when price_source is flash_sale"})
		} else if _, err := uuid.Parse(*req.FlashSaleItemID); err != nil {
			violations = append(violations, problem.Violation{Field: "flash_sale_item_id", Reason: "must be a valid UUID"})
		}
	} else if req.PriceSource == "regular" && req.FlashSaleItemID != nil {
		violations = append(violations, problem.Violation{Field: "flash_sale_item_id", Reason: "must be null when price_source is regular"})
	}

	if len(violations) > 0 {
		problem.AbortValidation(ctx, violations)
		return
	}

	cart, err := c.service.UpsertCartItem(ctx.Request.Context(), identity.UserID, skuID, req)
	if errors.Is(err, ErrCartLimitExceeded) {
		problem.Abort(ctx, http.StatusConflict, problem.CodeConflict, "Cart cannot exceed 20 distinct items.")
		return
	}
	if errors.Is(err, ErrSKUNotActive) {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "sku_id", Reason: "sku does not exist or is not active"}})
		return
	}
	if errors.Is(err, ErrExceedsMaxQuantity) {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "quantity", Reason: "exceeds maximum allowed quantity"}})
		return
	}
	if errors.Is(err, ErrFlashSaleItemInvalid) {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "flash_sale_item_id", Reason: "flash sale item is invalid or not active"}})
		return
	}
	if err != nil {
		problem.AbortInternal(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, cart)
}

func (c *Controller) RemoveItem(ctx *gin.Context) {
	identity, ok := auth.CurrentIdentity(ctx)
	if !ok {
		problem.Abort(ctx, http.StatusUnauthorized, problem.CodeUnauthenticated, "A valid session is required.")
		return
	}

	skuID := ctx.Param("skuId")
	if _, err := uuid.Parse(skuID); err != nil {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "skuId", Reason: "must be a valid UUID"}})
		return
	}

	if err := c.service.RemoveCartItem(ctx.Request.Context(), identity.UserID, skuID); err != nil {
		problem.AbortInternal(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *Controller) Preview(ctx *gin.Context) {
	identity, ok := auth.CurrentIdentity(ctx)
	if !ok {
		problem.Abort(ctx, http.StatusUnauthorized, problem.CodeUnauthenticated, "A valid session is required.")
		return
	}

	var req CheckoutSelection
	if err := ctx.ShouldBindJSON(&req); err != nil {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "body", Reason: "malformed JSON request body"}})
		return
	}

	var violations []problem.Violation
	if len(req.Items) == 0 {
		violations = append(violations, problem.Violation{Field: "items", Reason: "must contain at least one item"})
	}
	if len(req.Items) > MaxCartItems {
		violations = append(violations, problem.Violation{Field: "items", Reason: "cannot exceed 20 items"})
	}

	seenItems := make(map[string]bool)
	for i, it := range req.Items {
		idx := string(rune('0' + i))
		if _, err := uuid.Parse(it.CartItemID); err != nil {
			violations = append(violations, problem.Violation{Field: "items[" + idx + "].cart_item_id", Reason: "must be a valid UUID"})
		}
		if it.Version < 1 {
			violations = append(violations, problem.Violation{Field: "items[" + idx + "].version", Reason: "must be at least 1"})
		}
		if seenItems[it.CartItemID] {
			violations = append(violations, problem.Violation{Field: "items[" + idx + "].cart_item_id", Reason: "duplicate cart item in selection"})
		}
		seenItems[it.CartItemID] = true
	}

	if len(violations) > 0 {
		problem.AbortValidation(ctx, violations)
		return
	}

	preview, err := c.service.PreviewCheckout(ctx.Request.Context(), identity.UserID, req)
	if errors.Is(err, ErrCartItemNotFound) {
		problem.Abort(ctx, http.StatusConflict, problem.CodeConflict, "One or more cart items do not exist in the active cart.")
		return
	}
	if errors.Is(err, ErrCartItemVersionMismatch) {
		problem.Abort(ctx, http.StatusConflict, problem.CodeConflict, "One or more cart items have been modified; please refresh your cart.")
		return
	}
	if errors.Is(err, ErrSKUNotActive) {
		problem.Abort(ctx, http.StatusConflict, problem.CodeConflict, "One or more items are no longer available for purchase.")
		return
	}
	if errors.Is(err, ErrExceedsMaxQuantity) {
		problem.Abort(ctx, http.StatusConflict, problem.CodeConflict, "Quantity exceeds the maximum allowed for regular purchase.")
		return
	}
	if errors.Is(err, ErrFlashSaleNotActive) {
		problem.Abort(ctx, http.StatusConflict, problem.CodeConflict, "The flash sale for one or more items is not currently active.")
		return
	}
	if errors.Is(err, ErrQuotaExhausted) {
		problem.Abort(ctx, http.StatusConflict, problem.CodeConflict, "The flash sale quota for one or more items has been exhausted.")
		return
	}
	if errors.Is(err, ErrPurchaseLimitExceeded) {
		problem.Abort(ctx, http.StatusConflict, problem.CodeConflict, "The purchase limit for one or more flash sale items has been reached.")
		return
	}
	if errors.Is(err, ErrFlashSaleItemInvalid) {
		problem.Abort(ctx, http.StatusConflict, problem.CodeConflict, "One or more flash sale items are invalid.")
		return
	}
	if err != nil {
		problem.AbortInternal(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, preview)
}
