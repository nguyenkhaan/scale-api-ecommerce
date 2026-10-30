package catalog

import (
	"encoding/json"

	"cloudian/cloudian-scale/internal/model"
)

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
		response.SKUs = append(response.SKUs, SKUResponse{
			ID:           sku.ID,
			Code:         sku.Code,
		Attributes:   sku.Attributes,
		RegularPrice: sku.RegularPrice,
		Active:       sku.Status == model.SKUStatusActive,
		})
	}

	return response
}
