package flashsale

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// CreateDraft validates the request and stores a draft campaign with its items in one transaction.
func (s *Service) CreateDraft(ctx context.Context, req CreateFlashSaleRequest) (FlashSaleResponse, error) {
	input, violations := validateCreate(req)
	if len(violations) > 0 {
		return FlashSaleResponse{}, &ValidationError{Violations: violations}
	}

	var response FlashSaleResponse
	err := s.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := validateSKUs(tx, input.Items); err != nil {
			return err
		}
		var ids []string
		err := tx.Raw(`INSERT INTO flash_sale_campaigns (name, starts_at, ends_at) VALUES (?, ?, ?) RETURNING id`,
			input.Name, input.StartsAt, input.EndsAt).Scan(&ids).Error
		if err != nil {
			return fmt.Errorf("insert campaign: %w", err)
		}
		if len(ids) != 1 {
			return errors.New("insert campaign returned no id")
		}
		if err := insertItems(tx, ids[0], input.Items); err != nil {
			return fmt.Errorf("insert items: %w", err)
		}
		campaign, err := loadCampaign(tx, ids[0])
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

// UpdateDraft applies a partial update to a draft campaign. The campaign row is locked first, so the
// update can never interleave with Schedule on the same campaign, and items are changed in place.
func (s *Service) UpdateDraft(ctx context.Context, campaignID string, req UpdateFlashSaleRequest) (FlashSaleResponse, error) {
	patch, violations := validateUpdate(req)
	if len(violations) > 0 {
		return FlashSaleResponse{}, &ValidationError{Violations: violations}
	}

	var response FlashSaleResponse
	err := s.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		campaign, err := lockCampaign(tx, campaignID)
		if err != nil {
			return err
		}
		if campaign.Status != "draft" {
			return stateConflict("Campaign can only be modified while it is a draft; cancel the schedule first.")
		}
		if patch.Name != nil {
			campaign.Name = *patch.Name
		}
		if patch.StartsAt != nil {
			campaign.StartsAt = *patch.StartsAt
		}
		if patch.EndsAt != nil {
			campaign.EndsAt = *patch.EndsAt
		}
		if !campaign.EndsAt.After(campaign.StartsAt) {
			return invalid("ends_at", "must be strictly after starts_at")
		}
		if patch.Items != nil {
			if err := validateSKUs(tx, *patch.Items); err != nil {
				return err
			}
			if err := syncItems(tx, campaign.ID, *patch.Items); err != nil {
				return err
			}
		}
		err = tx.Exec(`UPDATE flash_sale_campaigns SET name = ?, starts_at = ?, ends_at = ?, updated_at = now()
			WHERE id = ? AND status = 'draft'`, campaign.Name, campaign.StartsAt, campaign.EndsAt, campaign.ID).Error
		if err != nil {
			return fmt.Errorf("update campaign: %w", err)
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
