package catalog

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"
	"uuid"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/model"
	"github.com/gin-gonic/gin"
)

const catalogBodyLimit = 64 << 10

type SKUResponse struct {
	ID           string          `json:"id"`
	Code         string          `json:"code"`
	Attributes   json.RawMessage `json:"attributes"`
	RegularPrice int64           `json:"regular_price"`
	Active       bool            `json:"active"`
}

type ProductResponse struct {
	ID          string        `json:"id"`
	CategoryID  *string       `json:"category_id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	ImageURL    *string       `json:"image_url"`
	Active      bool          `json:"active"`
	SKUs        []SKUResponse `json:"skus"`
}

type ProductPageResponse struct {
	Items      []ProductResponse `json:"items"`
	NextCursor *string           `json:"next_cursor"`
}

type ListPublicProductsQuery struct {
	Query      string
	CategoryID *string
	Cursor     string
	Limit      int
}

type AdminProductListQuery struct {
	Status *string
	Cursor string
	Limit  int
}

type CreateProductRequest struct {
	CategoryID  *string          `json:"category_id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	ImageURL    *string          `json:"image_url"`
	SKUs        []CreateSKUInput `json:"skus"`
}

type CreateSKUInput struct {
	Code                string          `json:"code"`
	Attributes          json.RawMessage `json:"attributes"`
	RegularPrice        int64           `json:"regular_price"`
	MaxQuantityPerOrder int             `json:"max_quantity_per_order"`
	InitialOnHand       int64           `json:"initial_on_hand"`
}

type OptionalString struct {
	Present bool
	Value   *string
}

func (value *OptionalString) UnmarshalJSON(data []byte) error {
	value.Present = true
	if bytes.Equal(data, []byte("null")) {
		value.Value = nil
		return nil
	}
	var decoded string
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	value.Value = &decoded
	return nil
}

type OptionalBool struct {
	Present bool
	Value   bool
}

func (value *OptionalBool) UnmarshalJSON(data []byte) error {
	value.Present = true
	if bytes.Equal(data, []byte("null")) {
		return errors.New("boolean cannot be null")
	}
	return json.Unmarshal(data, &value.Value)
}

type OptionalInt64 struct {
	Present bool
	Value   int64
}

func (value *OptionalInt64) UnmarshalJSON(data []byte) error {
	value.Present = true
	if bytes.Equal(data, []byte("null")) {
		return errors.New("integer cannot be null")
	}
	return json.Unmarshal(data, &value.Value)
}

type OptionalInt struct {
	Present bool
	Value   int
}

func (value *OptionalInt) UnmarshalJSON(data []byte) error {
	value.Present = true
	if bytes.Equal(data, []byte("null")) {
		return errors.New("integer cannot be null")
	}
	return json.Unmarshal(data, &value.Value)
}

type UpdateProductRequest struct {
	CategoryID  OptionalString `json:"category_id"`
	Name        OptionalString `json:"name"`
	Description OptionalString `json:"description"`
	ImageURL    OptionalString `json:"image_url"`
	Active      OptionalBool   `json:"active"`
}

type UpdateSKURequest struct {
	RegularPrice        OptionalInt64 `json:"regular_price"`
	MaxQuantityPerOrder OptionalInt   `json:"max_quantity_per_order"`
	Active              OptionalBool  `json:"active"`
}

func decodeCreateProduct(context *gin.Context) (CreateProductRequest, []problem.Violation) {
	var request CreateProductRequest
	if violations := decodeCatalogJSON(context, &request); len(violations) > 0 {
		return request, violations
	}
	return request, request.validate()
}

func decodeCreateSKU(context *gin.Context) (CreateSKUInput, []problem.Violation) {
	var request CreateSKUInput
	if violations := decodeCatalogJSON(context, &request); len(violations) > 0 {
		return request, violations
	}
	return request, request.validate("sku")
}

func decodeUpdateProduct(context *gin.Context) (UpdateProductRequest, []problem.Violation) {
	var request UpdateProductRequest
	if violations := decodeCatalogJSON(context, &request); len(violations) > 0 {
		return request, violations
	}
	return request, request.validate()
}

func decodeUpdateSKU(context *gin.Context) (UpdateSKURequest, []problem.Violation) {
	var request UpdateSKURequest
	if violations := decodeCatalogJSON(context, &request); len(violations) > 0 {
		return request, violations
	}
	return request, request.validate()
}

func decodeCatalogJSON(context *gin.Context, destination any) []problem.Violation {
	mediaType, _, err := mime.ParseMediaType(context.GetHeader("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return []problem.Violation{{Field: "Content-Type", Reason: "must be application/json"}}
	}

	decoder := json.NewDecoder(http.MaxBytesReader(context.Writer, context.Request.Body, catalogBodyLimit))
	var raw json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		return invalidCatalogBody()
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return invalidCatalogBody()
	}
	if len(bytes.TrimSpace(raw)) == 0 || bytes.TrimSpace(raw)[0] != '{' {
		return invalidCatalogBody()
	}
	strictDecoder := json.NewDecoder(bytes.NewReader(raw))
	strictDecoder.DisallowUnknownFields()
	if err := strictDecoder.Decode(destination); err != nil {
		return invalidCatalogBody()
	}
	return nil
}

func invalidCatalogBody() []problem.Violation {
	return []problem.Violation{{Field: "body", Reason: "must be one JSON object with only documented fields"}}
}

func (request *CreateProductRequest) validate() []problem.Violation {
	violations := validateCategoryID(request.CategoryID, "category_id")
	violations = append(violations, validateRequiredText(&request.Name, "name", 200)...)
	if utf8.RuneCountInString(request.Description) > 5000 {
		violations = append(violations, problem.Violation{Field: "description", Reason: "must be at most 5000 characters"})
	}
	violations = append(violations, validateOptionalURL(request.ImageURL, "image_url")...)
	if len(request.SKUs) == 0 {
		violations = append(violations, problem.Violation{Field: "skus", Reason: "must contain at least one SKU"})
	}
	if len(request.SKUs) > 100 {
		violations = append(violations, problem.Violation{Field: "skus", Reason: "must contain at most 100 SKUs"})
	}
	codes := make(map[string]struct{}, len(request.SKUs))
	for index := range request.SKUs {
		field := "skus[" + strconv.Itoa(index) + "]"
		violations = append(violations, request.SKUs[index].validate(field)...)
		if request.SKUs[index].Code == "" {
			continue
		}
		if _, exists := codes[request.SKUs[index].Code]; exists {
			violations = append(violations, problem.Violation{Field: field + ".code", Reason: "must be unique within the request"})
			continue
		}
		codes[request.SKUs[index].Code] = struct{}{}
	}
	return violations
}

func (request *CreateSKUInput) validate(field string) []problem.Violation {
	var violations []problem.Violation
	violations = append(violations, validateRequiredText(&request.Code, field+".code", 100)...)
	if request.RegularPrice < 1 {
		violations = append(violations, problem.Violation{Field: field + ".regular_price", Reason: "must be a positive integer"})
	}
	if request.MaxQuantityPerOrder < 1 {
		violations = append(violations, problem.Violation{Field: field + ".max_quantity_per_order", Reason: "must be a positive integer"})
	}
	if request.InitialOnHand < 0 {
		violations = append(violations, problem.Violation{Field: field + ".initial_on_hand", Reason: "must be zero or greater"})
	}
	if len(request.Attributes) == 0 {
		request.Attributes = json.RawMessage(`{}`)
	}
	if !isJSONObject(request.Attributes) {
		violations = append(violations, problem.Violation{Field: field + ".attributes", Reason: "must be a JSON object"})
	}
	return violations
}

func (request *UpdateProductRequest) validate() []problem.Violation {
	if !request.CategoryID.Present && !request.Name.Present && !request.Description.Present && !request.ImageURL.Present && !request.Active.Present {
		return []problem.Violation{{Field: "body", Reason: "must update at least one field"}}
	}
	var violations []problem.Violation
	if request.CategoryID.Present {
		violations = append(violations, validateCategoryID(request.CategoryID.Value, "category_id")...)
	}
	if request.Name.Present {
		if request.Name.Value == nil {
			violations = append(violations, problem.Violation{Field: "name", Reason: "must be a string"})
		} else {
			violations = append(violations, validateRequiredText(request.Name.Value, "name", 200)...)
		}
	}
	if request.Description.Present {
		if request.Description.Value == nil {
			violations = append(violations, problem.Violation{Field: "description", Reason: "must be a string"})
		} else if utf8.RuneCountInString(*request.Description.Value) > 5000 {
			violations = append(violations, problem.Violation{Field: "description", Reason: "must be at most 5000 characters"})
		}
	}
	if request.ImageURL.Present {
		violations = append(violations, validateOptionalURL(request.ImageURL.Value, "image_url")...)
	}
	return violations
}

func (request UpdateSKURequest) validate() []problem.Violation {
	if !request.RegularPrice.Present && !request.MaxQuantityPerOrder.Present && !request.Active.Present {
		return []problem.Violation{{Field: "body", Reason: "must update at least one field"}}
	}
	var violations []problem.Violation
	if request.RegularPrice.Present && request.RegularPrice.Value < 1 {
		violations = append(violations, problem.Violation{Field: "regular_price", Reason: "must be a positive integer"})
	}
	if request.MaxQuantityPerOrder.Present && request.MaxQuantityPerOrder.Value < 1 {
		violations = append(violations, problem.Violation{Field: "max_quantity_per_order", Reason: "must be a positive integer"})
	}
	return violations
}

func validateCategoryID(categoryID *string, field string) []problem.Violation {
	if categoryID == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*categoryID)
	if _, err := uuid.Parse(trimmed); err != nil {
		return []problem.Violation{{Field: field, Reason: "must be a valid UUID"}}
	}
	*categoryID = trimmed
	return nil
}

func validateRequiredText(value *string, field string, maximum int) []problem.Violation {
	*value = strings.TrimSpace(*value)
	switch {
	case *value == "":
		return []problem.Violation{{Field: field, Reason: "is required"}}
	case utf8.RuneCountInString(*value) > maximum:
		return []problem.Violation{{Field: field, Reason: "must be at most " + strconv.Itoa(maximum) + " characters"}}
	default:
		return nil
	}
}

func validateOptionalURL(value *string, field string) []problem.Violation {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	parsed, err := url.ParseRequestURI(trimmed)
	if trimmed == "" || err != nil || !parsed.IsAbs() {
		return []problem.Violation{{Field: field, Reason: "must be an absolute URI"}}
	}
	*value = trimmed
	return nil
}

func isJSONObject(value json.RawMessage) bool {
	trimmed := bytes.TrimSpace(value)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return false
	}
	var object map[string]json.RawMessage
	return json.Unmarshal(trimmed, &object) == nil
}

func newProductResponse(product model.Product, skus []model.SKU) ProductResponse {
	response := ProductResponse{
		ID:          product.ID,
		CategoryID:  product.CategoryID,
		Name:        product.Name,
		Description: product.Description,
		ImageURL:    product.ImageURL,
		Active:      product.Status == model.ProductStatusActive,
		SKUs:        make([]SKUResponse, 0, len(skus)),
	}

	for _, sku := range skus {
		response.SKUs = append(response.SKUs, newSKUResponse(sku))
	}

	return response
}

func newSKUResponse(sku model.SKU) SKUResponse {
	return SKUResponse{
		ID:           sku.ID,
		Code:         sku.Code,
		Attributes:   sku.Attributes,
		RegularPrice: sku.RegularPrice,
		Active:       sku.Status == model.SKUStatusActive,
	}
}
