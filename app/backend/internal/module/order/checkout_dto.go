package order

import (
	"encoding/json"
	"time"
)

type PlaceOrderLine struct {
	CartItemID      string  `json:"cart_item_id"`
	CartItemVersion int64   `json:"cart_item_version"`
	SKUID           string  `json:"sku_id"`
	Quantity        int     `json:"quantity"`
	UnitPrice       int64   `json:"unit_price"`
	Subtotal        int64   `json:"subtotal"`
	PriceSource     string  `json:"price_source"`
	FlashSaleItemID *string `json:"flash_sale_item_id"`
}

type RecipientRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

type PlaceOrderRequest struct {
	Lines         []PlaceOrderLine `json:"lines"`
	Currency      string           `json:"currency"`
	ExpectedTotal int64            `json:"expected_total"`
	Recipient     RecipientRequest `json:"recipient"`
}

type IdempotencyResult struct {
	IdempotencyKey string          `json:"idempotency_key"`
	ResponseStatus int             `json:"response_status"`
	ResponseBody   json.RawMessage `json:"response_body"`
	CompletedAt    time.Time       `json:"completed_at"`
}
