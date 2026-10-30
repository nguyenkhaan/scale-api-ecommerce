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
	"cloudian/cloudian-scale/internal/model"
	"cloudian/cloudian-scale/internal/module/auth"
	"github.com/gin-gonic/gin"
)

type ProductService interface {
	GetPublicProduct(context.Context, string) (ProductResponse, error)
	ListPublicProducts(context.Context, ListPublicProductsQuery) (ProductPageResponse, error)
	GetAdminProduct(context.Context, string) (ProductResponse, error)
	ListAdminProducts(context.Context, AdminProductListQuery) (ProductPageResponse, error)
	CreateProduct(context.Context, string, CreateProductRequest) (ProductResponse, error)
	CreateSKU(context.Context, string, string, CreateSKUInput) (SKUResponse, error)
	UpdateProduct(context.Context, string, UpdateProductRequest) (ProductResponse, error)
	UpdateSKU(context.Context, string, UpdateSKURequest) (SKUResponse, error)
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

func RegisterAdminRoutes(
	router *gin.RouterGroup,
	controller *Controller,
	requireSession gin.HandlerFunc,
	requireAdmin gin.HandlerFunc,
) {
	group := router.Group("/v1/admin/products", requireSession, requireAdmin)
	group.GET("", controller.ListAdminProducts)
	group.POST("", controller.CreateProduct)
	group.GET("/:productId", controller.GetAdminProduct)
	group.PATCH("/:productId", controller.UpdateProduct)
	group.POST("/:productId/skus", controller.CreateSKU)
	skus := router.Group("/v1/admin/skus", requireSession, requireAdmin)
	skus.PATCH("/:skuId", controller.UpdateSKU)
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

func (controller *Controller) ListAdminProducts(context *gin.Context) {
	query, violations := decodeAdminProductListQuery(context)
	if len(violations) > 0 {
		problem.AbortValidation(context, violations)
		return
	}

	page, err := controller.service.ListAdminProducts(context.Request.Context(), query)
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

func decodeAdminProductListQuery(context *gin.Context) (AdminProductListQuery, []problem.Violation) {
	query := AdminProductListQuery{
		Cursor: context.Query("cursor"),
		Limit:  20,
	}
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
	if rawStatus, exists := context.GetQuery("status"); exists {
		status := strings.TrimSpace(rawStatus)
		if status != model.ProductStatusActive && status != model.ProductStatusInactive {
			violations = append(violations, problem.Violation{Field: "status", Reason: "must be active or inactive"})
		} else {
			query.Status = &status
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

func (controller *Controller) GetAdminProduct(context *gin.Context) {
	productID := context.Param("productId")
	if _, err := uuid.Parse(productID); err != nil {
		problem.AbortValidation(context, []problem.Violation{{
			Field:  "productId",
			Reason: "must be a valid UUID",
		}})
		return
	}

	product, err := controller.service.GetAdminProduct(context.Request.Context(), productID)
	switch {
	case errors.Is(err, ErrProductNotFound):
		problem.Abort(context, http.StatusNotFound, problem.CodeResourceNotFound, "Product was not found.")
	case err != nil:
		problem.AbortInternal(context, err)
	default:
		context.JSON(http.StatusOK, product)
	}
}

func (controller *Controller) CreateProduct(context *gin.Context) {
	request, violations := decodeCreateProduct(context)
	if len(violations) > 0 {
		problem.AbortValidation(context, violations)
		return
	}
	actorUserID, ok := adminActorID(context)
	if !ok {
		return
	}

	product, err := controller.service.CreateProduct(context.Request.Context(), actorUserID, request)
	writeCatalogMutationError(context, err)
	if err == nil {
		context.JSON(http.StatusCreated, product)
	}
}

func (controller *Controller) CreateSKU(context *gin.Context) {
	productID := context.Param("productId")
	if _, err := uuid.Parse(productID); err != nil {
		problem.AbortValidation(context, []problem.Violation{{
			Field:  "productId",
			Reason: "must be a valid UUID",
		}})
		return
	}
	request, violations := decodeCreateSKU(context)
	if len(violations) > 0 {
		problem.AbortValidation(context, violations)
		return
	}
	actorUserID, ok := adminActorID(context)
	if !ok {
		return
	}

	sku, err := controller.service.CreateSKU(context.Request.Context(), actorUserID, productID, request)
	writeCatalogMutationError(context, err)
	if err == nil {
		context.JSON(http.StatusCreated, sku)
	}
}

func (controller *Controller) UpdateProduct(context *gin.Context) {
	productID := context.Param("productId")
	if _, err := uuid.Parse(productID); err != nil {
		problem.AbortValidation(context, []problem.Violation{{
			Field:  "productId",
			Reason: "must be a valid UUID",
		}})
		return
	}
	request, violations := decodeUpdateProduct(context)
	if len(violations) > 0 {
		problem.AbortValidation(context, violations)
		return
	}

	product, err := controller.service.UpdateProduct(context.Request.Context(), productID, request)
	writeCatalogMutationError(context, err)
	if err == nil {
		context.JSON(http.StatusOK, product)
	}
}

func (controller *Controller) UpdateSKU(context *gin.Context) {
	skuID := context.Param("skuId")
	if _, err := uuid.Parse(skuID); err != nil {
		problem.AbortValidation(context, []problem.Violation{{
			Field:  "skuId",
			Reason: "must be a valid UUID",
		}})
		return
	}
	request, violations := decodeUpdateSKU(context)
	if len(violations) > 0 {
		problem.AbortValidation(context, violations)
		return
	}

	sku, err := controller.service.UpdateSKU(context.Request.Context(), skuID, request)
	writeCatalogMutationError(context, err)
	if err == nil {
		context.JSON(http.StatusOK, sku)
	}
}

func adminActorID(context *gin.Context) (string, bool) {
	identity, ok := auth.CurrentIdentity(context)
	if !ok {
		problem.Abort(context, http.StatusUnauthorized, problem.CodeUnauthenticated, "A valid session is required.")
		return "", false
	}
	return identity.UserID, true
}

func writeCatalogMutationError(context *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrProductNotFound), errors.Is(err, ErrSKUNotFound):
		problem.Abort(context, http.StatusNotFound, problem.CodeResourceNotFound, "Catalog resource was not found.")
	case errors.Is(err, ErrCategoryNotFound):
		problem.AbortValidation(context, []problem.Violation{{Field: "category_id", Reason: "does not exist"}})
	case errors.Is(err, ErrSKUCodeTaken):
		problem.Abort(context, http.StatusConflict, problem.CodeSKUAlreadyExists, "SKU code already exists.")
	case err != nil:
		problem.AbortInternal(context, err)
	}
}
