package cart

import (
	"context"
	"fmt"

	"cloudian/cloudian-scale/internal/model"
	"gorm.io/gorm"
)

// UpsertCartItem sets the absolute quantity of the user's line for a SKU. The request has already
// passed shape validation; this checks it against current catalog and sale data.
func (s *Service) UpsertCartItem(ctx context.Context, userID, skuID string, req UpsertCartItemRequest) (CartResponse, error) {
	var view CartResponse
	err := s.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Locking the cart row serializes every writer of this cart, so the 20-SKU count and the
		// write below cannot interleave with another writer.
		cart, err := activeCart(tx, userID, true)
		if err != nil {
			return err
		}
		if err := validateTarget(tx, skuID, req); err != nil {
			return err
		}
		if err := writeLine(tx, cart.ID, skuID, req); err != nil {
			return err
		}
		view, err = loadCartView(tx, cart.ID)
		return err
	})
	if err != nil {
		return CartResponse{}, err
	}
	return view, nil
}

// validateTarget checks SKU availability, quantity caps and the flash sale item for an upsert.
func validateTarget(tx *gorm.DB, skuID string, req UpsertCartItemRequest) error {
	skus, err := loadSKUInfos(tx, []string{skuID})
	if err != nil {
		return err
	}
	sku, ok := skus[skuID]
	if !ok || !sku.Available {
		return ErrItemUnavailable
	}
	if req.Quantity > sku.MaxQuantityPerOrder {
		return fieldError("quantity", fmt.Sprintf("must not exceed %d per order", sku.MaxQuantityPerOrder))
	}
	if req.PriceSource != model.PriceSourceFlashSale {
		return nil
	}

	itemID := *req.FlashSaleItemID
	flashItems, err := loadFlashInfos(tx, []string{itemID})
	if err != nil {
		return err
	}
	item, ok := flashItems[itemID]
	if !ok || item.SKUID != skuID {
		return fieldError("flash_sale_item_id", "does not exist for this SKU")
	}
	now, err := dbNow(tx)
	if err != nil {
		return err
	}
	if !item.notEnded(now) {
		return ErrSaleNotActive
	}
	if int64(req.Quantity) > item.PurchaseLimit {
		return fieldError("quantity", fmt.Sprintf("must not exceed the flash sale purchase limit of %d", item.PurchaseLimit))
	}
	return nil
}

// writeLine applies the write as one atomic statement so the version bump cannot be lost.
func writeLine(tx *gorm.DB, cartID, skuID string, req UpsertCartItemRequest) error {
	if req.ExpectedVersion != nil {
		result := tx.Exec(`
			UPDATE cart_items
			SET quantity = ?, intended_price_source = ?, flash_sale_item_id = ?,
			    version = version + 1, updated_at = now()
			WHERE cart_id = ? AND sku_id = ? AND version = ?`,
			req.Quantity, req.PriceSource, req.FlashSaleItemID, cartID, skuID, *req.ExpectedVersion)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrCartItemChanged
		}
		return nil
	}

	var counts struct {
		Total    int64 `gorm:"column:total"`
		Existing int64 `gorm:"column:existing"`
	}
	if err := tx.Raw(`
		SELECT count(*) AS total, count(*) FILTER (WHERE sku_id = ?) AS existing
		FROM cart_items WHERE cart_id = ?`, skuID, cartID).Scan(&counts).Error; err != nil {
		return err
	}
	if counts.Existing == 0 && counts.Total >= MaxCartItems {
		return fieldError("cart", fmt.Sprintf("cannot exceed %d distinct items", MaxCartItems))
	}
	return tx.Exec(`
		INSERT INTO cart_items (cart_id, sku_id, quantity, intended_price_source, flash_sale_item_id)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (cart_id, sku_id) DO UPDATE SET
			quantity = EXCLUDED.quantity,
			intended_price_source = EXCLUDED.intended_price_source,
			flash_sale_item_id = EXCLUDED.flash_sale_item_id,
			version = cart_items.version + 1,
			updated_at = now()`,
		cartID, skuID, req.Quantity, req.PriceSource, req.FlashSaleItemID).Error
}
