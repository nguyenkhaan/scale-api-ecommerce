package flashsale

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"uuid"

	"cloudian/cloudian-scale/internal/httpserver/problem"
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
	requireAdmin gin.HandlerFunc,
) {
	// Public routes
	public := api.Group("/v1/flash-sales")
	public.GET("", controller.ListPublic)
	public.GET("/:flashSaleId", controller.GetPublic)

	// Admin routes
	admin := api.Group("/v1/admin/flash-sales", requireSession, requireAdmin)
	admin.GET("", controller.ListAdmin)
	admin.POST("", controller.Create)
	admin.GET("/:flashSaleId", controller.GetAdmin)
	admin.PATCH("/:flashSaleId", controller.Update)
	admin.POST("/:flashSaleId/schedule", controller.Schedule)
	admin.POST("/:flashSaleId/cancel-schedule", controller.CancelSchedule)
	admin.POST("/:flashSaleId/stop", controller.Stop)
}

func (c *Controller) ListPublic(ctx *gin.Context) {
	query, violations := decodeListQuery(ctx, true)
	if len(violations) > 0 {
		problem.AbortValidation(ctx, violations)
		return
	}

	page, err := c.service.ListPublicFlashSales(ctx.Request.Context(), query)
	if errors.Is(err, ErrInvalidCursor) {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "cursor", Reason: "is invalid or does not match the filters"}})
		return
	}
	if err != nil {
		problem.AbortInternal(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, page)
}

func (c *Controller) GetPublic(ctx *gin.Context) {
	id := ctx.Param("flashSaleId")
	if _, err := uuid.Parse(id); err != nil {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "flashSaleId", Reason: "must be a valid UUID"}})
		return
	}

	sale, err := c.service.GetPublicFlashSale(ctx.Request.Context(), id)
	if errors.Is(err, ErrNotFound) {
		problem.Abort(ctx, http.StatusNotFound, problem.CodeResourceNotFound, "Flash sale was not found.")
		return
	}
	if err != nil {
		problem.AbortInternal(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, sale)
}

func (c *Controller) ListAdmin(ctx *gin.Context) {
	query, violations := decodeListQuery(ctx, false)
	if len(violations) > 0 {
		problem.AbortValidation(ctx, violations)
		return
	}

	page, err := c.service.ListAdminFlashSales(ctx.Request.Context(), query)
	if errors.Is(err, ErrInvalidCursor) {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "cursor", Reason: "is invalid or does not match the filters"}})
		return
	}
	if err != nil {
		problem.AbortInternal(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, page)
}

func (c *Controller) Create(ctx *gin.Context) {
	var req CreateFlashSaleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "body", Reason: "malformed JSON request body"}})
		return
	}

	var violations []problem.Violation
	if strings.TrimSpace(req.Name) == "" {
		violations = append(violations, problem.Violation{Field: "name", Reason: "is required"})
	}
	if len(req.Name) > 200 {
		violations = append(violations, problem.Violation{Field: "name", Reason: "must be at most 200 characters"})
	}
	if req.StartsAt == "" {
		violations = append(violations, problem.Violation{Field: "starts_at", Reason: "is required"})
	}
	if req.EndsAt == "" {
		violations = append(violations, problem.Violation{Field: "ends_at", Reason: "is required"})
	}
	if len(req.Items) == 0 {
		violations = append(violations, problem.Violation{Field: "items", Reason: "must have at least one item"})
	}
	for i, it := range req.Items {
		if _, err := uuid.Parse(it.SKUID); err != nil {
			violations = append(violations, problem.Violation{Field: "items[" + strconv.Itoa(i) + "].sku_id", Reason: "must be a valid UUID"})
		}
		if it.SalePrice <= 0 {
			violations = append(violations, problem.Violation{Field: "items[" + strconv.Itoa(i) + "].sale_price", Reason: "must be greater than 0"})
		}
		if it.QuotaTotal <= 0 {
			violations = append(violations, problem.Violation{Field: "items[" + strconv.Itoa(i) + "].quota_total", Reason: "must be greater than 0"})
		}
		if it.PurchaseLimit <= 0 {
			violations = append(violations, problem.Violation{Field: "items[" + strconv.Itoa(i) + "].purchase_limit", Reason: "must be greater than 0"})
		}
	}
	if len(violations) > 0 {
		problem.AbortValidation(ctx, violations)
		return
	}

	res, err := c.service.CreateDraft(ctx.Request.Context(), req)
	if errors.Is(err, ErrInvalidTimeRange) {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "ends_at", Reason: "must be strictly after starts_at"}})
		return
	}
	if errors.Is(err, ErrDuplicateSKU) {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "items", Reason: "contains duplicate sku_id"}})
		return
	}
	if errors.Is(err, ErrSKUNotActive) {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "items", Reason: "one or more SKUs do not exist or are not active"}})
		return
	}
	if errors.Is(err, ErrPriceNotDiscounted) {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "items", Reason: "sale_price must be strictly less than sku regular_price"}})
		return
	}
	if err != nil {
		problem.AbortInternal(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, res)
}

func (c *Controller) GetAdmin(ctx *gin.Context) {
	id := ctx.Param("flashSaleId")
	if _, err := uuid.Parse(id); err != nil {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "flashSaleId", Reason: "must be a valid UUID"}})
		return
	}

	sale, err := c.service.GetAdminFlashSale(ctx.Request.Context(), id)
	if errors.Is(err, ErrNotFound) {
		problem.Abort(ctx, http.StatusNotFound, problem.CodeResourceNotFound, "Flash sale was not found.")
		return
	}
	if err != nil {
		problem.AbortInternal(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, sale)
}

func (c *Controller) Update(ctx *gin.Context) {
	id := ctx.Param("flashSaleId")
	if _, err := uuid.Parse(id); err != nil {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "flashSaleId", Reason: "must be a valid UUID"}})
		return
	}

	var req UpdateFlashSaleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "body", Reason: "malformed JSON request body"}})
		return
	}

	var violations []problem.Violation
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		violations = append(violations, problem.Violation{Field: "name", Reason: "cannot be empty"})
	}
	if req.Items != nil {
		if len(*req.Items) == 0 {
			violations = append(violations, problem.Violation{Field: "items", Reason: "must have at least one item"})
		}
		for i, it := range *req.Items {
			if _, err := uuid.Parse(it.SKUID); err != nil {
				violations = append(violations, problem.Violation{Field: "items[" + strconv.Itoa(i) + "].sku_id", Reason: "must be a valid UUID"})
			}
			if it.SalePrice <= 0 {
				violations = append(violations, problem.Violation{Field: "items[" + strconv.Itoa(i) + "].sale_price", Reason: "must be greater than 0"})
			}
			if it.QuotaTotal <= 0 {
				violations = append(violations, problem.Violation{Field: "items[" + strconv.Itoa(i) + "].quota_total", Reason: "must be greater than 0"})
			}
			if it.PurchaseLimit <= 0 {
				violations = append(violations, problem.Violation{Field: "items[" + strconv.Itoa(i) + "].purchase_limit", Reason: "must be greater than 0"})
			}
		}
	}
	if len(violations) > 0 {
		problem.AbortValidation(ctx, violations)
		return
	}

	res, err := c.service.UpdateDraft(ctx.Request.Context(), id, req)
	if errors.Is(err, ErrNotFound) {
		problem.Abort(ctx, http.StatusNotFound, problem.CodeResourceNotFound, "Flash sale was not found.")
		return
	}
	if errors.Is(err, ErrNotDraft) {
		problem.Abort(ctx, http.StatusConflict, problem.CodeConflict, "Campaign can only be modified while in draft status.")
		return
	}
	if errors.Is(err, ErrInvalidTimeRange) {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "ends_at", Reason: "must be strictly after starts_at"}})
		return
	}
	if errors.Is(err, ErrDuplicateSKU) {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "items", Reason: "contains duplicate sku_id"}})
		return
	}
	if errors.Is(err, ErrSKUNotActive) {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "items", Reason: "one or more SKUs do not exist or are not active"}})
		return
	}
	if errors.Is(err, ErrPriceNotDiscounted) {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "items", Reason: "sale_price must be strictly less than sku regular_price"}})
		return
	}
	if err != nil {
		problem.AbortInternal(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) Schedule(ctx *gin.Context) {
	id := ctx.Param("flashSaleId")
	if _, err := uuid.Parse(id); err != nil {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "flashSaleId", Reason: "must be a valid UUID"}})
		return
	}

	res, err := c.service.Schedule(ctx.Request.Context(), id)
	if errors.Is(err, ErrNotFound) {
		problem.Abort(ctx, http.StatusNotFound, problem.CodeResourceNotFound, "Flash sale was not found.")
		return
	}
	if errors.Is(err, ErrNotDraft) {
		problem.Abort(ctx, http.StatusConflict, problem.CodeConflict, "Campaign must be in draft status to schedule.")
		return
	}
	if errors.Is(err, ErrOverlappingSKU) {
		problem.Abort(ctx, http.StatusConflict, problem.CodeConflict, "One or more SKUs overlap with an existing scheduled or active campaign.")
		return
	}
	if err != nil {
		problem.AbortInternal(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) CancelSchedule(ctx *gin.Context) {
	id := ctx.Param("flashSaleId")
	if _, err := uuid.Parse(id); err != nil {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "flashSaleId", Reason: "must be a valid UUID"}})
		return
	}

	res, err := c.service.CancelSchedule(ctx.Request.Context(), id)
	if errors.Is(err, ErrNotFound) {
		problem.Abort(ctx, http.StatusNotFound, problem.CodeResourceNotFound, "Flash sale was not found.")
		return
	}
	if errors.Is(err, ErrNotScheduled) {
		problem.Abort(ctx, http.StatusConflict, problem.CodeConflict, "Campaign is not currently scheduled.")
		return
	}
	if errors.Is(err, ErrAlreadyStarted) {
		problem.Abort(ctx, http.StatusConflict, problem.CodeConflict, "Cannot cancel schedule for an active or ended campaign.")
		return
	}
	if err != nil {
		problem.AbortInternal(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) Stop(ctx *gin.Context) {
	id := ctx.Param("flashSaleId")
	if _, err := uuid.Parse(id); err != nil {
		problem.AbortValidation(ctx, []problem.Violation{{Field: "flashSaleId", Reason: "must be a valid UUID"}})
		return
	}

	res, err := c.service.Stop(ctx.Request.Context(), id)
	if errors.Is(err, ErrNotFound) {
		problem.Abort(ctx, http.StatusNotFound, problem.CodeResourceNotFound, "Flash sale was not found.")
		return
	}
	if errors.Is(err, ErrNotScheduled) {
		problem.Abort(ctx, http.StatusConflict, problem.CodeConflict, "Campaign is not currently scheduled.")
		return
	}
	if err != nil {
		problem.AbortInternal(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, res)
}

func decodeListQuery(ctx *gin.Context, isPublic bool) (ListFlashSalesQuery, []problem.Violation) {
	query := ListFlashSalesQuery{
		Cursor: ctx.Query("cursor"),
		Limit:  20,
	}
	var violations []problem.Violation

	if rawLimit, exists := ctx.GetQuery("limit"); exists {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil || limit < 1 || limit > 100 {
			violations = append(violations, problem.Violation{Field: "limit", Reason: "must be an integer between 1 and 100"})
		} else {
			query.Limit = limit
		}
	}

	if status, exists := ctx.GetQuery("status"); exists {
		status = strings.TrimSpace(status)
		valid := false
		if isPublic {
			valid = status == "scheduled" || status == "active" || status == "ended"
		} else {
			valid = status == "draft" || status == "scheduled" || status == "active" || status == "stopped" || status == "ended"
		}
		if !valid {
			violations = append(violations, problem.Violation{Field: "status", Reason: "is not an allowed status value"})
		} else {
			query.Status = &status
		}
	}

	return query, violations
}
