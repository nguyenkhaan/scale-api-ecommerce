package flashsale

import (
	"time"

	"cloudian/cloudian-scale/internal/model"
)

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

func DeriveEffectiveStatus(status string, startsAt, endsAt, now time.Time) string {
	if status == "stopped" {
		return "stopped"
	}
	if status == "draft" {
		return "draft"
	}
	if status == "scheduled" {
		if now.Before(startsAt) {
			return "scheduled"
		}
		if now.Before(endsAt) {
			return "active"
		}
		return "ended"
	}
	return status
}

func deriveEffectiveStatus(campaign model.FlashSaleCampaign, now time.Time) string {
	return DeriveEffectiveStatus(campaign.Status, campaign.StartsAt, campaign.EndsAt, now)
}

func newFlashSaleResponse(campaign model.FlashSaleCampaign, items []model.FlashSaleItem, now time.Time) FlashSaleResponse {
	itemResponses := make([]FlashSaleItemResponse, len(items))
	for i, item := range items {
		itemResponses[i] = FlashSaleItemResponse{
			ID:            item.ID,
			SKUID:         item.SKUID,
			SalePrice:     item.SalePrice,
			QuotaTotal:    item.QuotaTotal,
			PurchaseLimit: item.PurchaseLimit,
		}
	}
	return FlashSaleResponse{
		ID:       campaign.ID,
		Name:     campaign.Name,
		Status:   deriveEffectiveStatus(campaign, now),
		StartsAt: campaign.StartsAt.UTC().Format(time.RFC3339),
		EndsAt:   campaign.EndsAt.UTC().Format(time.RFC3339),
		Items:    itemResponses,
	}
}
