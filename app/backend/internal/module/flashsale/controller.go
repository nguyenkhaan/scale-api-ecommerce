package flashsale

import (
	"context"
	"net/http"
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
	public := api.Group("/flash-sales")
	public.GET("", controller.ListPublic)
	public.GET("/:flashSaleId", controller.GetPublic)

	admin := api.Group("/admin/flash-sales", requireSession, requireAdmin)
	admin.GET("", controller.ListAdmin)
	admin.POST("", controller.Create)
	admin.GET("/:flashSaleId", controller.GetAdmin)
	admin.PATCH("/:flashSaleId", controller.Update)
	admin.POST("/:flashSaleId/schedule", controller.Schedule)
	admin.POST("/:flashSaleId/cancel-schedule", controller.CancelSchedule)
	admin.POST("/:flashSaleId/stop", controller.Stop)
}

func (c *Controller) ListPublic(ctx *gin.Context) {
	c.list(ctx, true, c.service.ListPublicFlashSales)
}

func (c *Controller) ListAdmin(ctx *gin.Context) {
	c.list(ctx, false, c.service.ListAdminFlashSales)
}

func (c *Controller) GetPublic(ctx *gin.Context) {
	c.respond(ctx, http.StatusOK, c.service.GetPublicFlashSale)
}

func (c *Controller) GetAdmin(ctx *gin.Context) {
	c.respond(ctx, http.StatusOK, c.service.GetAdminFlashSale)
}

func (c *Controller) list(
	ctx *gin.Context,
	public bool,
	fetch func(context.Context, ListFlashSalesQuery) (FlashSalePageResponse, error),
) {
	query, violations := decodeListQuery(ctx, public)
	if len(violations) > 0 {
		problem.AbortValidation(ctx, violations)
		return
	}
	page, err := fetch(ctx.Request.Context(), query)
	if err != nil {
		abortWithError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, page)
}

// respond validates the path ID, runs an ID-based service call and writes its result.
func (c *Controller) respond(ctx *gin.Context, status int, call func(context.Context, string) (FlashSaleResponse, error)) {
	id := ctx.Param("flashSaleId")
	if _, err := uuid.Parse(id); err != nil {
		problem.AbortValidation(ctx, []problem.Violation{violation("flashSaleId", "must be a valid UUID")})
		return
	}
	sale, err := call(ctx.Request.Context(), id)
	if err != nil {
		abortWithError(ctx, err)
		return
	}
	ctx.JSON(status, sale)
}
