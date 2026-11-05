package flashsale

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// Schedule moves a draft to scheduled. Everything runs in one transaction: the campaign row is locked,
// then a per-SKU advisory lock is taken in ascending sku_id order, so concurrent schedules (and draft
// edits, which lock the same row) of campaigns that share a SKU are serialized before the overlap check.
func (s *Service) Schedule(ctx context.Context, campaignID string) (FlashSaleResponse, error) {
	var response FlashSaleResponse
	err := s.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		campaign, err := lockCampaign(tx, campaignID)
		if err != nil {
			return err
		}
		if campaign.Status != "draft" {
			return stateConflict("Only a draft campaign can be scheduled.")
		}
		if !campaign.EndsAt.After(campaign.DBNow) {
			return stateConflict("The campaign window has already ended.")
		}

		items, err := loadScheduleItems(tx, campaign.ID)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return invalid("items", "must contain at least one item")
		}
		skuIDs := make([]string, len(items))
		for i, item := range items {
			skuIDs[i] = item.SKUID
			if err := tx.Exec(`SELECT pg_advisory_xact_lock(hashtextextended(?::text, 0))`, item.SKUID).Error; err != nil {
				return fmt.Errorf("lock sku %s: %w", item.SKUID, err)
			}
		}
		if err := validateSKUs(tx, items); err != nil {
			return err
		}

		var overlapping bool
		err = tx.Raw(`
			SELECT EXISTS (
				SELECT 1 FROM flash_sale_items i
				JOIN flash_sale_campaigns c ON c.id = i.campaign_id
				WHERE i.campaign_id <> ?
				  AND c.status IN ('scheduled', 'stopped')
				  AND c.starts_at < ? AND c.ends_at > ?
				  AND i.sku_id IN ?)`,
			campaign.ID, campaign.EndsAt, campaign.StartsAt, skuIDs).Scan(&overlapping).Error
		if err != nil {
			return fmt.Errorf("check overlapping campaigns: %w", err)
		}
		if overlapping {
			return scheduleOverlap("A SKU already belongs to a scheduled or stopped campaign with an overlapping time window.")
		}

		result := tx.Exec(`UPDATE flash_sale_campaigns SET status = 'scheduled', updated_at = now()
			WHERE id = ? AND status = 'draft'`, campaign.ID)
		if result.Error != nil {
			return fmt.Errorf("schedule campaign: %w", result.Error)
		}
		if result.RowsAffected != 1 {
			return stateConflict("Only a draft campaign can be scheduled.")
		}
		campaign, err = loadCampaign(tx, campaign.ID)
		if err != nil {
			return err
		}
		response, err = buildResponse(tx, campaign)
		return err
	})
	if err != nil {
		return FlashSaleResponse{}, err
	}
	return response, nil
}

// loadScheduleItems reads a campaign's items in ascending sku_id order, the global lock order.
func loadScheduleItems(tx *gorm.DB, campaignID string) ([]CreateFlashSaleItem, error) {
	var rows []struct {
		SKUID         string `gorm:"column:sku_id"`
		SalePrice     int64  `gorm:"column:sale_price"`
		QuotaTotal    int64  `gorm:"column:quota_total"`
		PurchaseLimit int    `gorm:"column:purchase_limit"`
	}
	err := tx.Raw(`SELECT sku_id::text AS sku_id, sale_price, quota_total, purchase_limit
		FROM flash_sale_items WHERE campaign_id = ? ORDER BY sku_id`, campaignID).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("load campaign items: %w", err)
	}
	items := make([]CreateFlashSaleItem, len(rows))
	for i, row := range rows {
		items[i] = CreateFlashSaleItem{SKUID: row.SKUID, SalePrice: row.SalePrice, QuotaTotal: row.QuotaTotal, PurchaseLimit: row.PurchaseLimit}
	}
	return items, nil
}

// CancelSchedule returns a scheduled campaign that has not started (by database time) to draft.
func (s *Service) CancelSchedule(ctx context.Context, campaignID string) (FlashSaleResponse, error) {
	return s.guardedTransition(ctx, campaignID, `
		UPDATE flash_sale_campaigns SET status = 'draft', updated_at = now()
		WHERE id = ? AND status = 'scheduled' AND starts_at > now()
		RETURNING `+campaignWithNow, "cancel the schedule of")
}

// Stop halts a running campaign. Upcoming campaigns must be cancelled instead and ended ones stay as is.
func (s *Service) Stop(ctx context.Context, campaignID string) (FlashSaleResponse, error) {
	return s.guardedTransition(ctx, campaignID, `
		UPDATE flash_sale_campaigns SET status = 'stopped', updated_at = now()
		WHERE id = ? AND status = 'scheduled' AND starts_at <= now() AND ends_at > now()
		RETURNING `+campaignWithNow, "stop")
}

// guardedTransition runs a single guarded UPDATE ... RETURNING. When no row matches, it reports 404 for
// a missing campaign and a state conflict otherwise.
func (s *Service) guardedTransition(ctx context.Context, campaignID, statement, action string) (FlashSaleResponse, error) {
	db := s.database.WithContext(ctx)
	var rows []campaignRow
	if err := db.Raw(statement, campaignID).Scan(&rows).Error; err != nil {
		return FlashSaleResponse{}, fmt.Errorf("%s: %w", action, err)
	}
	if len(rows) == 0 {
		current, err := loadCampaign(db, campaignID)
		if err != nil {
			return FlashSaleResponse{}, err
		}
		return FlashSaleResponse{}, stateConflict(fmt.Sprintf("Cannot %s a campaign that is %s.", action, current.effectiveStatus()))
	}
	return buildResponse(db, rows[0])
}
