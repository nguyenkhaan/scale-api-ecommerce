package cart

import (
	"context"
	"errors"
	"time"

	"cloudian/cloudian-scale/internal/model"
	"gorm.io/gorm"
)

var (
	ErrCartItemNotFound     = errors.New("cart item was not found in active cart")
	ErrCartItemVersionMismatch = errors.New("cart item version mismatch; item was modified")
	ErrFlashSaleNotActive   = errors.New("flash sale campaign is not currently active")
	ErrQuotaExhausted       = errors.New("flash sale quota is exhausted")
	ErrPurchaseLimitExceeded = errors.New("flash sale purchase limit exceeded")
)

type SelectedCartItem struct {
	CartItemID string `json:"cart_item_id"`
	Version    int64  `json:"version"`
}

type CheckoutSelection struct {
	Items []SelectedCartItem `json:"items"`
}

type CheckoutLineResponse struct {
	CartItemID      string  `json:"cart_item_id"`
	CartItemVersion int64   `json:"cart_item_version"`
	SKUID           string  `json:"sku_id"`
	Quantity        int     `json:"quantity"`
	UnitPrice       int64   `json:"unit_price"`
	Subtotal        int64   `json:"subtotal"`
	PriceSource     string  `json:"price_source"`
	FlashSaleItemID *string `json:"flash_sale_item_id"`
}

type CheckoutPreviewResponse struct {
	Currency    string                 `json:"currency"`
	TotalAmount int64                  `json:"total_amount"`
	Lines       []CheckoutLineResponse `json:"lines"`
	PreviewedAt string                 `json:"previewed_at"`
}

func (s *Service) PreviewCheckout(ctx context.Context, userID string, selection CheckoutSelection) (CheckoutPreviewResponse, error) {
	if len(selection.Items) == 0 {
		return CheckoutPreviewResponse{}, errors.New("items must contain at least one item")
	}
	if len(selection.Items) > MaxCartItems {
		return CheckoutPreviewResponse{}, errors.New("items cannot exceed 20 items")
	}

	db := s.database.WithContext(ctx)

	var cart model.Cart
	if err := db.Where("user_id = ? AND status = 'active'", userID).First(&cart).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return CheckoutPreviewResponse{}, ErrCartItemNotFound
		}
		return CheckoutPreviewResponse{}, err
	}

	var cartItems []model.CartItem
	if err := db.Where("cart_id = ?", cart.ID).Find(&cartItems).Error; err != nil {
		return CheckoutPreviewResponse{}, err
	}

	cartItemMap := make(map[string]model.CartItem)
	for _, it := range cartItems {
		cartItemMap[it.ID] = it
	}

	now := s.now()
	var lines []CheckoutLineResponse
	var totalAmount int64

	for _, sel := range selection.Items {
		item, exists := cartItemMap[sel.CartItemID]
		if !exists {
			return CheckoutPreviewResponse{}, ErrCartItemNotFound
		}
		if item.Version != sel.Version {
			return CheckoutPreviewResponse{}, ErrCartItemVersionMismatch
		}

		var sku model.SKU
		if err := db.Where("id = ? AND status = ?", item.SKUID, model.SKUStatusActive).First(&sku).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return CheckoutPreviewResponse{}, ErrSKUNotActive
			}
			return CheckoutPreviewResponse{}, err
		}

		var unitPrice int64
		if item.IntendedPriceSource == model.PriceSourceRegular {
			if item.Quantity > sku.MaxQuantityPerOrder {
				return CheckoutPreviewResponse{}, ErrExceedsMaxQuantity
			}
			unitPrice = sku.RegularPrice
		} else {
			if item.FlashSaleItemID == nil {
				return CheckoutPreviewResponse{}, ErrFlashSaleItemInvalid
			}

			var fsItem model.FlashSaleItem
			if err := db.Where("id = ? AND sku_id = ?", *item.FlashSaleItemID, item.SKUID).First(&fsItem).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return CheckoutPreviewResponse{}, ErrFlashSaleItemInvalid
				}
				return CheckoutPreviewResponse{}, err
			}

			var campaign model.FlashSaleCampaign
			if err := db.Where("id = ?", fsItem.CampaignID).First(&campaign).Error; err != nil {
				return CheckoutPreviewResponse{}, err
			}

			// Must be scheduled and within active time window
			if campaign.Status != "scheduled" || now.Before(campaign.StartsAt) || !now.Before(campaign.EndsAt) {
				return CheckoutPreviewResponse{}, ErrFlashSaleNotActive
			}

			// Check available quota
			availableQuota := fsItem.QuotaTotal - fsItem.QuotaReserved - fsItem.QuotaSold
			if int64(item.Quantity) > availableQuota {
				return CheckoutPreviewResponse{}, ErrQuotaExhausted
			}

			// Check customer usage purchase limit
			var usage model.FlashSaleCustomerUsage
			err := db.Where("flash_sale_item_id = ? AND user_id = ?", fsItem.ID, userID).First(&usage).Error
			usedQuantity := int64(0)
			if err == nil {
				usedQuantity = usage.ReservedQuantity + usage.SoldQuantity
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return CheckoutPreviewResponse{}, err
			}

			remainingLimit := int64(fsItem.PurchaseLimit) - usedQuantity
			if int64(item.Quantity) > remainingLimit {
				return CheckoutPreviewResponse{}, ErrPurchaseLimitExceeded
			}

			unitPrice = fsItem.SalePrice
		}

		subtotal := int64(item.Quantity) * unitPrice
		totalAmount += subtotal

		lines = append(lines, CheckoutLineResponse{
			CartItemID:      item.ID,
			CartItemVersion: item.Version,
			SKUID:           item.SKUID,
			Quantity:        item.Quantity,
			UnitPrice:       unitPrice,
			Subtotal:        subtotal,
			PriceSource:     item.IntendedPriceSource,
			FlashSaleItemID: item.FlashSaleItemID,
		})
	}

	return CheckoutPreviewResponse{
		Currency:    "VND",
		TotalAmount: totalAmount,
		Lines:       lines,
		PreviewedAt: now.UTC().Format(time.RFC3339),
	}, nil
}
