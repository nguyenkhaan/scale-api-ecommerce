package flashsale

import "time"

type CreateFlashSaleItem struct {
	SKUID         string `json:"sku_id"`
	SalePrice     int64  `json:"sale_price"`
	QuotaTotal    int64  `json:"quota_total"`
	PurchaseLimit int    `json:"purchase_limit"`
}

type CreateFlashSaleRequest struct {
	Name     string                `json:"name"`
	StartsAt string                `json:"starts_at"`
	EndsAt   string                `json:"ends_at"`
	Items    []CreateFlashSaleItem `json:"items"`
}

type UpdateFlashSaleRequest struct {
	Name     *string                `json:"name"`
	StartsAt *string                `json:"starts_at"`
	EndsAt   *string                `json:"ends_at"`
	Items    *[]CreateFlashSaleItem `json:"items"`
}

type FlashSaleItemResponse struct {
	ID            string `json:"id"`
	SKUID         string `json:"sku_id"`
	SKUCode       string `json:"sku_code"`
	ProductName   string `json:"product_name"`
	RegularPrice  int64  `json:"regular_price"`
	SalePrice     int64  `json:"sale_price"`
	QuotaTotal    int64  `json:"quota_total"`
	PurchaseLimit int    `json:"purchase_limit"`
}

type FlashSaleResponse struct {
	ID       string                  `json:"id"`
	Name     string                  `json:"name"`
	Status   string                  `json:"status"`
	StartsAt string                  `json:"starts_at"`
	EndsAt   string                  `json:"ends_at"`
	Items    []FlashSaleItemResponse `json:"items"`
}

type FlashSalePageResponse struct {
	Items      []FlashSaleResponse `json:"items"`
	NextCursor *string             `json:"next_cursor"`
}

type ListFlashSalesQuery struct {
	Status *string
	Cursor string
	Limit  int
}

// DeriveEffectiveStatus maps a stored status and the campaign window to the contract status.
// now must come from the database clock.
func DeriveEffectiveStatus(status string, startsAt, endsAt, now time.Time) string {
	if status != "scheduled" {
		return status
	}
	switch {
	case now.Before(startsAt):
		return "scheduled"
	case now.Before(endsAt):
		return "active"
	default:
		return "ended"
	}
}
