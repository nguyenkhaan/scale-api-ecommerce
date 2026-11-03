package cart

import (
	"cloudian/cloudian-scale/internal/model"
)

type UpsertCartItemRequest struct {
	Quantity         int     `json:"quantity"`
	PriceSource      string  `json:"price_source"`
	FlashSaleItemID *string `json:"flash_sale_item_id"`
}

type CartItemResponse struct {
	ID              string  `json:"id"`
	SKUID           string  `json:"sku_id"`
	Quantity        int     `json:"quantity"`
	Version         int64   `json:"version"`
	PriceSource     string  `json:"price_source"`
	FlashSaleItemID *string `json:"flash_sale_item_id"`
}

type CartResponse struct {
	ID    string             `json:"id"`
	Items []CartItemResponse `json:"items"`
}

func newCartResponse(cart model.Cart, items []model.CartItem) CartResponse {
	itemResponses := make([]CartItemResponse, len(items))
	for i, it := range items {
		itemResponses[i] = CartItemResponse{
			ID:              it.ID,
			SKUID:           it.SKUID,
			Quantity:        it.Quantity,
			Version:         it.Version,
			PriceSource:     it.IntendedPriceSource,
			FlashSaleItemID: it.FlashSaleItemID,
		}
	}
	return CartResponse{
		ID:    cart.ID,
		Items: itemResponses,
	}
}
