package flashsale

import (
	"fmt"
	"strings"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/model"
	"gorm.io/gorm"
)

type skuState struct {
	ID            string `gorm:"column:id"`
	Status        string `gorm:"column:status"`
	RegularPrice  int64  `gorm:"column:regular_price"`
	ProductStatus string `gorm:"column:product_status"`
	CategoryOK    bool   `gorm:"column:category_ok"`
}

// validateSKUs checks, with one query, that every item points at a sellable SKU (active SKU, active
// product, visible or missing category) and that its sale price is below the current regular price.
func validateSKUs(tx *gorm.DB, items []CreateFlashSaleItem) error {
	ids := make([]string, len(items))
	for i, item := range items {
		ids[i] = item.SKUID
	}
	var states []skuState
	err := tx.Raw(`
		SELECT s.id, s.status, s.regular_price, p.status AS product_status,
		       (p.category_id IS NULL OR COALESCE(c.is_visible, false)) AS category_ok
		FROM skus s
		JOIN products p ON p.id = s.product_id
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE s.id IN ?`, ids).Scan(&states).Error
	if err != nil {
		return fmt.Errorf("load skus: %w", err)
	}
	byID := make(map[string]skuState, len(states))
	for _, state := range states {
		byID[state.ID] = state
	}

	var violations []problem.Violation
	for i, item := range items {
		field := fmt.Sprintf("items[%d].", i)
		state, found := byID[item.SKUID]
		if !found || state.Status != model.SKUStatusActive ||
			state.ProductStatus != model.ProductStatusActive || !state.CategoryOK {
			violations = append(violations, violation(field+"sku_id", "must reference an active SKU of an active product in a visible category"))
			continue
		}
		if item.SalePrice >= state.RegularPrice {
			violations = append(violations, violation(field+"sale_price",
				fmt.Sprintf("must be less than the SKU regular price (%d)", state.RegularPrice)))
		}
	}
	if len(violations) > 0 {
		return &ValidationError{Violations: violations}
	}
	return nil
}

// insertItems adds items to a campaign in one statement.
func insertItems(tx *gorm.DB, campaignID string, items []CreateFlashSaleItem) error {
	if len(items) == 0 {
		return nil
	}
	tuples := make([]string, len(items))
	args := make([]any, 0, len(items)*5)
	for i, item := range items {
		tuples[i] = "(?, ?, ?, ?, ?)"
		args = append(args, campaignID, item.SKUID, item.SalePrice, item.QuotaTotal, item.PurchaseLimit)
	}
	return tx.Exec(`INSERT INTO flash_sale_items (campaign_id, sku_id, sale_price, quota_total, purchase_limit) VALUES `+
		strings.Join(tuples, ", "), args...).Error
}

// syncItems makes the campaign's items equal to wanted, in place by sku_id: items of SKUs that stay keep
// their ids, new SKUs are inserted and removed SKUs are deleted. The caller holds the campaign row lock.
func syncItems(tx *gorm.DB, campaignID string, wanted []CreateFlashSaleItem) error {
	var existing []string
	if err := tx.Raw(`SELECT sku_id FROM flash_sale_items WHERE campaign_id = ?`, campaignID).Scan(&existing).Error; err != nil {
		return fmt.Errorf("load current items: %w", err)
	}
	current := make(map[string]bool, len(existing))
	for _, skuID := range existing {
		current[skuID] = true
	}
	keep := make(map[string]bool, len(wanted))
	var kept, added []CreateFlashSaleItem
	for _, item := range wanted {
		keep[item.SKUID] = true
		if current[item.SKUID] {
			kept = append(kept, item)
		} else {
			added = append(added, item)
		}
	}
	var removed []string
	for _, skuID := range existing {
		if !keep[skuID] {
			removed = append(removed, skuID)
		}
	}

	if len(removed) > 0 {
		err := tx.Exec(`DELETE FROM flash_sale_items WHERE campaign_id = ? AND sku_id IN ?`, campaignID, removed).Error
		if err != nil {
			if pgErrorCode(err) == sqlStateForeignKeyViolation {
				return stateConflict("An item cannot be removed because carts or customer usage still reference it.")
			}
			return fmt.Errorf("delete removed items: %w", err)
		}
	}
	if err := updateKeptItems(tx, campaignID, kept); err != nil {
		return err
	}
	if err := insertItems(tx, campaignID, added); err != nil {
		return fmt.Errorf("insert new items: %w", err)
	}
	return nil
}

// updateKeptItems rewrites price, quota and limit of existing items in one statement.
func updateKeptItems(tx *gorm.DB, campaignID string, kept []CreateFlashSaleItem) error {
	if len(kept) == 0 {
		return nil
	}
	tuples := make([]string, len(kept))
	args := make([]any, 0, len(kept)*4+1)
	for i, item := range kept {
		tuples[i] = "(?::uuid, ?::bigint, ?::bigint, ?::integer)"
		args = append(args, item.SKUID, item.SalePrice, item.QuotaTotal, item.PurchaseLimit)
	}
	args = append(args, campaignID)
	err := tx.Exec(`
		UPDATE flash_sale_items AS i
		SET sale_price = v.sale_price, quota_total = v.quota_total,
		    purchase_limit = v.purchase_limit, updated_at = now()
		FROM (VALUES `+strings.Join(tuples, ", ")+`) AS v(sku_id, sale_price, quota_total, purchase_limit)
		WHERE i.sku_id = v.sku_id AND i.campaign_id = ?`, args...).Error
	if err != nil {
		if pgErrorCode(err) == sqlStateCheckViolation {
			return stateConflict("quota_total cannot be lowered below the quota already reserved or sold.")
		}
		return fmt.Errorf("update kept items: %w", err)
	}
	return nil
}
