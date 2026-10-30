package catalog

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
	"uuid"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"github.com/gin-gonic/gin"
)

type ProductService interface {
	GetPublicProduct(context.Context, string) (ProductResponse, error)
	ListPublicProducts(context.Context, ListPublicProductsQuery) (ProductPageResponse, error)
}

type Controller struct {
	service ProductService
}

func NewController(service ProductService) *Controller {
	return &Controller{service: service}
}

func RegisterRoutes(router *gin.RouterGroup, controller *Controller) {
	group := router.Group("/v1/products")
	group.GET("", controller.ListProducts)
	group.GET("/:productId", controller.GetProduct)
}

func (controller *Controller) ListProducts(context *gin.Context) {
	query, violations := decodePublicListQuery(context)
	if len(violations) > 0 {
		problem.AbortValidation(context, violations)
		return
	}

	page, err := controller.service.ListPublicProducts(context.Request.Context(), query)
	switch {
	case errors.Is(err, ErrInvalidCursor):
		problem.AbortValidation(context, []problem.Violation{{
			Field:  "cursor",
			Reason: "is invalid or does not match the filters",
		}})
	case err != nil:
		problem.AbortInternal(context, err)
	default:
		context.JSON(http.StatusOK, page)
	}
}

func decodePublicListQuery(context *gin.Context) (ListPublicProductsQuery, []problem.Violation) {
	query := ListPublicProductsQuery{
		Query:  strings.TrimSpace(context.Query("query")),
		Cursor: context.Query("cursor"),
		Limit:  20,
	}
	var violations []problem.Violation
	if utf8.RuneCountInString(query.Query) > 100 {
		violations = append(violations, problem.Violation{Field: "query", Reason: "must be at most 100 characters"})
	}
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
	if categoryID, exists := context.GetQuery("category_id"); exists {
		categoryID = strings.TrimSpace(categoryID)
		if _, err := uuid.Parse(categoryID); err != nil {
			violations = append(violations, problem.Violation{Field: "category_id", Reason: "must be a valid UUID"})
		} else {
			query.CategoryID = &categoryID
		}
	}
	return query, violations
}

func (controller *Controller) GetProduct(context *gin.Context) {
	productID := context.Param("productId")
	if _, err := uuid.Parse(productID); err != nil {
		problem.AbortValidation(context, []problem.Violation{{
			Field:  "productId",
			Reason: "must be a valid UUID",
		}})
		return
	}

	product, err := controller.service.GetPublicProduct(context.Request.Context(), productID)
	switch {
	case errors.Is(err, ErrProductNotFound):
		problem.Abort(context, http.StatusNotFound, problem.CodeResourceNotFound, "Product was not found.")
	case err != nil:
		problem.AbortInternal(context, err)
	default:
		context.JSON(http.StatusOK, product)
	}
}
