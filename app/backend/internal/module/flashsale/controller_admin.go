package flashsale

import (
	"net/http"
	"uuid"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"github.com/gin-gonic/gin"
)

func (c *Controller) Create(ctx *gin.Context) {
	var req CreateFlashSaleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		problem.AbortValidation(ctx, []problem.Violation{violation("body", "malformed JSON request body")})
		return
	}
	res, err := c.service.CreateDraft(ctx.Request.Context(), req)
	if err != nil {
		abortWithError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, res)
}

func (c *Controller) Update(ctx *gin.Context) {
	id := ctx.Param("flashSaleId")
	if _, err := uuid.Parse(id); err != nil {
		problem.AbortValidation(ctx, []problem.Violation{violation("flashSaleId", "must be a valid UUID")})
		return
	}
	var req UpdateFlashSaleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		problem.AbortValidation(ctx, []problem.Violation{violation("body", "malformed JSON request body")})
		return
	}
	res, err := c.service.UpdateDraft(ctx.Request.Context(), id, req)
	if err != nil {
		abortWithError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) Schedule(ctx *gin.Context) {
	c.respond(ctx, http.StatusOK, c.service.Schedule)
}

func (c *Controller) CancelSchedule(ctx *gin.Context) {
	c.respond(ctx, http.StatusOK, c.service.CancelSchedule)
}

func (c *Controller) Stop(ctx *gin.Context) {
	c.respond(ctx, http.StatusOK, c.service.Stop)
}
