package cart

import (
	"errors"
	"net/http"
	"strings"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/module/auth"
	"github.com/gin-gonic/gin"
	"uuid"
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
	cartGroup := api.Group("/cart", requireSession)
	cartGroup.GET("", controller.GetCart)
	cartGroup.PUT("/items/:skuId", controller.UpsertItem)
	cartGroup.DELETE("/items/:skuId", controller.RemoveItem)

	checkoutGroup := api.Group("/checkout", requireSession)
	checkoutGroup.POST("/preview", controller.Preview)
}

// currentUserID aborts with 401 when the request has no session identity.
func currentUserID(ctx *gin.Context) (string, bool) {
	identity, ok := auth.CurrentIdentity(ctx)
	if !ok {
		problem.Abort(ctx, http.StatusUnauthorized, problem.CodeUnauthenticated, "A valid session is required.")
		return "", false
	}
	return identity.UserID, true
}

// skuIDParam validates and canonicalizes the {skuId} path parameter, aborting with 422 when invalid.
func skuIDParam(ctx *gin.Context) (string, bool) {
	skuID := ctx.Param("skuId")
	if _, err := uuid.Parse(skuID); err != nil {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "skuId", Reason: "must be a valid UUID"}})
		return "", false
	}
	return strings.ToLower(skuID), true
}

// abortServiceError maps the service's business errors to contract responses.
func abortServiceError(ctx *gin.Context, err error) {
	var validation *ValidationError
	var lineErrors *LineErrors
	switch {
	case errors.As(err, &validation):
		problem.AbortValidation(ctx, validation.Violations)
	case errors.As(err, &lineErrors):
		problem.AbortLineErrors(ctx, "One or more selected items cannot be purchased as requested.", lineErrors.Lines)
	case errors.Is(err, ErrItemUnavailable):
		problem.Abort(ctx, http.StatusConflict, problem.CodeItemUnavailable, "The item is not available for purchase.")
	case errors.Is(err, ErrSaleNotActive):
		problem.Abort(ctx, http.StatusConflict, problem.CodeSaleNotActive, "The flash sale is not active.")
	case errors.Is(err, ErrCartItemChanged):
		problem.Abort(ctx, http.StatusConflict, problem.CodeCartItemChanged, "Cart items changed; reload the cart and retry.")
	default:
		problem.AbortInternal(ctx, err)
	}
}

func (c *Controller) GetCart(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return
	}
	cart, err := c.service.GetCart(ctx.Request.Context(), userID)
	if err != nil {
		problem.AbortInternal(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, cart)
}

func (c *Controller) RemoveItem(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return
	}
	skuID, ok := skuIDParam(ctx)
	if !ok {
		return
	}
	if err := c.service.RemoveCartItem(ctx.Request.Context(), userID, skuID); err != nil {
		problem.AbortInternal(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}
