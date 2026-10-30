package inventory

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/module/auth"
	"uuid"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	service InventoryService
}

type InventoryService interface {
	Adjust(context.Context, string, string, AdjustmentRequest) (InventoryResponse, error)
	ListInventories(context.Context, ListInventoriesQuery) (InventoryPageResponse, error)
	ListAdjustments(context.Context, string, ListAdjustmentHistoryQuery) (InventoryAdjustmentPageResponse, error)
}

func NewController(service InventoryService) *Controller {
	return &Controller{service: service}
}

func RegisterRoutes(
	router *gin.RouterGroup,
	controller *Controller,
	requireSession gin.HandlerFunc,
	requireAdmin gin.HandlerFunc,
) {
	group := router.Group("/v1/admin/inventories", requireSession, requireAdmin)
	group.GET("", controller.List)
	group.GET("/:skuId/adjustments", controller.ListAdjustments)
	group.POST("/:skuId/adjustments", controller.Adjust)
}

func (controller *Controller) List(context *gin.Context) {
	query, violations := decodeInventoryListQuery(context)
	if len(violations) > 0 {
		problem.AbortValidation(context, violations)
		return
	}
	page, err := controller.service.ListInventories(context.Request.Context(), query)
	switch {
	case errors.Is(err, ErrInvalidCursor):
		problem.AbortValidation(context, []problem.Violation{{
			Field:  "cursor",
			Reason: "is invalid",
		}})
	case err != nil:
		problem.AbortInternal(context, err)
	default:
		context.JSON(http.StatusOK, page)
	}
}

func decodeInventoryListQuery(context *gin.Context) (ListInventoriesQuery, []problem.Violation) {
	query := ListInventoriesQuery{Cursor: context.Query("cursor"), Limit: 20}
	var violations []problem.Violation
	if len(query.Cursor) > 500 {
		violations = append(violations, problem.Violation{Field: "cursor", Reason: "must be at most 500 characters"})
	}
	if rawLimit, exists := context.GetQuery("limit"); exists {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil || limit < 1 || limit > 100 {
			violations = append(violations, problem.Violation{Field: "limit", Reason: "must be an integer between 1 and 100"})
		} else {
			query.Limit = limit
		}
	}
	return query, violations
}

func (controller *Controller) ListAdjustments(context *gin.Context) {
	skuID := context.Param("skuId")
	if _, err := uuid.Parse(skuID); err != nil {
		problem.AbortValidation(context, []problem.Violation{{
			Field:  "skuId",
			Reason: "must be a valid UUID",
		}})
		return
	}
	query, violations := decodeAdjustmentHistoryQuery(context)
	if len(violations) > 0 {
		problem.AbortValidation(context, violations)
		return
	}
	page, err := controller.service.ListAdjustments(context.Request.Context(), skuID, query)
	switch {
	case errors.Is(err, ErrInvalidCursor):
		problem.AbortValidation(context, []problem.Violation{{Field: "cursor", Reason: "is invalid"}})
	case errors.Is(err, ErrInventoryNotFound):
		problem.Abort(context, http.StatusNotFound, problem.CodeResourceNotFound, "Inventory was not found.")
	case err != nil:
		problem.AbortInternal(context, err)
	default:
		context.JSON(http.StatusOK, page)
	}
}

func decodeAdjustmentHistoryQuery(context *gin.Context) (ListAdjustmentHistoryQuery, []problem.Violation) {
	query := ListAdjustmentHistoryQuery{Cursor: context.Query("cursor"), Limit: 20}
	var violations []problem.Violation
	if len(query.Cursor) > 500 {
		violations = append(violations, problem.Violation{Field: "cursor", Reason: "must be at most 500 characters"})
	}
	if rawLimit, exists := context.GetQuery("limit"); exists {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil || limit < 1 || limit > 100 {
			violations = append(violations, problem.Violation{Field: "limit", Reason: "must be an integer between 1 and 100"})
		} else {
			query.Limit = limit
		}
	}
	return query, violations
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
		problem.Abort(context, http.StatusNotFound, problem.CodeResourceNotFound, "Inventory was not found.")
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
