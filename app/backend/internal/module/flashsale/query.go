package flashsale

import (
	"context"
	"fmt"
)

const defaultPageSize = 20

// statusConditions translate a contract status filter to SQL evaluated against the database clock.
var statusConditions = map[string]string{
	"draft":     "status = 'draft'",
	"stopped":   "status = 'stopped'",
	"scheduled": "status = 'scheduled' AND starts_at > now()",
	"active":    "status = 'scheduled' AND starts_at <= now() AND ends_at > now()",
	"ended":     "status = 'scheduled' AND ends_at <= now()",
}

// GetPublicFlashSale returns a campaign that customers may see; drafts and stopped campaigns are hidden.
func (s *Service) GetPublicFlashSale(ctx context.Context, campaignID string) (FlashSaleResponse, error) {
	return s.get(ctx, campaignID, true)
}

func (s *Service) GetAdminFlashSale(ctx context.Context, campaignID string) (FlashSaleResponse, error) {
	return s.get(ctx, campaignID, false)
}

func (s *Service) get(ctx context.Context, campaignID string, publicOnly bool) (FlashSaleResponse, error) {
	db := s.database.WithContext(ctx)
	campaign, err := loadCampaign(db, campaignID)
	if err != nil {
		return FlashSaleResponse{}, err
	}
	if publicOnly && campaign.Status != "scheduled" {
		return FlashSaleResponse{}, ErrNotFound
	}
	return buildResponse(db, campaign)
}

func (s *Service) ListPublicFlashSales(ctx context.Context, query ListFlashSalesQuery) (FlashSalePageResponse, error) {
	return s.list(ctx, query, "public")
}

func (s *Service) ListAdminFlashSales(ctx context.Context, query ListFlashSalesQuery) (FlashSalePageResponse, error) {
	return s.list(ctx, query, "admin")
}

// list pages campaigns newest first with a keyset cursor. The public scope only exposes scheduled campaigns.
func (s *Service) list(ctx context.Context, query ListFlashSalesQuery, scope string) (FlashSalePageResponse, error) {
	fingerprint := filterFingerprint(scope, query.Status)
	var cursor *flashSaleCursor
	if query.Cursor != "" {
		decoded, err := decodeCursor(query.Cursor, fingerprint)
		if err != nil {
			return FlashSalePageResponse{}, ErrInvalidCursor
		}
		cursor = &decoded
	}
	limit := query.Limit
	if limit <= 0 {
		limit = defaultPageSize
	}

	db := s.database.WithContext(ctx).Table("flash_sale_campaigns").Select(campaignWithNow)
	if scope == "public" {
		db = db.Where("status = 'scheduled'")
	}
	if query.Status != nil {
		condition, known := statusConditions[*query.Status]
		if !known || (scope == "public" && (*query.Status == "draft" || *query.Status == "stopped")) {
			return FlashSalePageResponse{}, invalid("status", "is not an allowed status value")
		}
		db = db.Where(condition)
	}
	if cursor != nil {
		db = db.Where("(created_at < ? OR (created_at = ? AND id < ?))", cursor.CreatedAt, cursor.CreatedAt, cursor.ID)
	}

	var campaigns []campaignRow
	if err := db.Order("created_at DESC, id DESC").Limit(limit + 1).Find(&campaigns).Error; err != nil {
		return FlashSalePageResponse{}, fmt.Errorf("list %s campaigns: %w", scope, err)
	}

	var nextCursor *string
	if len(campaigns) > limit {
		campaigns = campaigns[:limit]
		last := campaigns[limit-1]
		encoded, err := encodeCursor(last.CreatedAt, last.ID, fingerprint)
		if err != nil {
			return FlashSalePageResponse{}, fmt.Errorf("encode cursor: %w", err)
		}
		nextCursor = &encoded
	}

	items, err := buildResponses(s.database.WithContext(ctx), campaigns)
	if err != nil {
		return FlashSalePageResponse{}, err
	}
	return FlashSalePageResponse{Items: items, NextCursor: nextCursor}, nil
}
