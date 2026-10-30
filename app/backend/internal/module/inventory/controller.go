package inventory

import (
	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/module/auth"
	"errors"
	"net/http"
	"uuid"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

func RegisterRoutes(
	router *gin.RouterGroup, 
	controller *Controller, 
	requireSession gin.HandlerFunc , 
	requireAdmin gin.HandlerFunc) {
		group := router.Group("/v1/products")
		group.POST(
			"/:skuId/adjustments",
			requireSession,
			requireAdmin,
			controller.Adjust,
		)
}

func (controller *Controller) Adjust(context *gin.Context) {
	skuID := context.Param("skuId")
	if _, err := uuid.Parse(skuID); err != nil {
		problem.AbortValidation(context, []problem.Violation{{
			Field:  "skuId",
			Reason: "must be a valid UUID",
		}})
		return
	}

	request, violations := decodeAdjustment(context)
	if len(violations) > 0 {
		problem.AbortValidation(context, violations)
		return
	}

	identity, ok := auth.CurrentIdentity(context)
	if !ok {
		problem.Abort(
			context,
			http.StatusUnauthorized,
			problem.CodeUnauthenticated,
			"A valid session is required.",
		)
		return
	}

	response, err := controller.service.Adjust(
		context.Request.Context(),
		skuID,
		identity.UserID,
		request,
	)

	switch {
	case errors.Is(err, ErrInventoryNotFound):
		problem.Abort(
			context,
			http.StatusNotFound,
			problem.CodeNotFound,
			"Inventory was not found.",
		)
	case errors.Is(err, ErrBelowReserved):
		problem.Abort(
			context,
			http.StatusConflict,
			problem.CodeInventoryBelowReserved,
			"On-hand quantity cannot be lower than reserved quantity.",
		)
	case errors.Is(err, ErrQuantityOverflow):
		problem.AbortValidation(context, []problem.Violation{{
			Field:  "delta",
			Reason: "would exceed the supported inventory range",
		}})
	case err != nil:
		problem.AbortInternal(context, err)
	default:
		context.JSON(http.StatusOK, response)
	}
}