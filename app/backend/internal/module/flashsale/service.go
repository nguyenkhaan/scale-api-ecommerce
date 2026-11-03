package flashsale

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"cloudian/cloudian-scale/internal/model"
	"gorm.io/gorm"
	"uuid"
)

var (
	ErrNotFound          = errors.New("flash sale campaign not found")
	ErrNotDraft          = errors.New("campaign can only be modified in draft status")
	ErrNotScheduled      = errors.New("campaign is not scheduled")
	ErrAlreadyStarted    = errors.New("campaign has already started or ended")
	ErrOverlappingSKU    = errors.New("sku overlaps with an existing scheduled or active campaign")
	ErrInvalidTimeRange  = errors.New("ends_at must be strictly after starts_at")
	ErrEmptyItems        = errors.New("campaign must have at least one item")
	ErrDuplicateSKU      = errors.New("duplicate sku in campaign items")
	ErrSKUNotActive      = errors.New("sku does not exist or is not active")
	ErrPriceNotDiscounted = errors.New("sale price must be less than regular price")
)

type Service struct {
	database *gorm.DB
	nowFunc  func() time.Time
}

func NewService(database *gorm.DB) (*Service, error) {
	if database == nil {
		return nil, errors.New("flash sale database is required")
	}
	return &Service{
		database: database,
		nowFunc:  time.Now,
	}, nil
}

func (s *Service) WithNowFunc(fn func() time.Time) *Service {
	s.nowFunc = fn
	return s
}

func (s *Service) now() time.Time {
	if s.nowFunc != nil {
		return s.nowFunc()
	}
	return time.Now()
}

func (s *Service) CreateDraft(ctx context.Context, req CreateFlashSaleRequest) (FlashSaleResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > 200 {
		return FlashSaleResponse{}, errors.New("name must be between 1 and 200 characters")
	}

	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		return FlashSaleResponse{}, fmt.Errorf("parse starts_at: %w", err)
	}
	endsAt, err := time.Parse(time.RFC3339, req.EndsAt)
	if err != nil {
		return FlashSaleResponse{}, fmt.Errorf("parse ends_at: %w", err)
	}
	if !endsAt.After(startsAt) {
		return FlashSaleResponse{}, ErrInvalidTimeRange
	}
	if len(req.Items) == 0 {
		return FlashSaleResponse{}, ErrEmptyItems
	}

	skuIDs := make([]string, len(req.Items))
	skuSeen := make(map[string]bool)
	for i, item := range req.Items {
		if skuSeen[item.SKUID] {
			return FlashSaleResponse{}, ErrDuplicateSKU
		}
		skuSeen[item.SKUID] = true
		skuIDs[i] = item.SKUID

		if item.SalePrice <= 0 || item.QuotaTotal <= 0 || item.PurchaseLimit <= 0 {
			return FlashSaleResponse{}, errors.New("item sale_price, quota_total, and purchase_limit must be positive")
		}
	}

	db := s.database.WithContext(ctx)

	// Validate SKUs exist, are active, and sale_price < regular_price
	var skus []model.SKU
	if err := db.Where("id IN ?", skuIDs).Find(&skus).Error; err != nil {
		return FlashSaleResponse{}, fmt.Errorf("query skus: %w", err)
	}
	if len(skus) != len(skuIDs) {
		return FlashSaleResponse{}, ErrSKUNotActive
	}
	skuMap := make(map[string]model.SKU)
	for _, sku := range skus {
		if sku.Status != model.SKUStatusActive {
			return FlashSaleResponse{}, ErrSKUNotActive
		}
		skuMap[sku.ID] = sku
	}
	for _, item := range req.Items {
		sku := skuMap[item.SKUID]
		if item.SalePrice >= sku.RegularPrice {
			return FlashSaleResponse{}, ErrPriceNotDiscounted
		}
	}

	campaign := model.FlashSaleCampaign{
		ID:        uuid.NewV4().String(),
		Name:      name,
		Status:    "draft",
		StartsAt:  startsAt,
		EndsAt:    endsAt,
		CreatedAt: s.now(),
		UpdatedAt: s.now(),
	}

	items := make([]model.FlashSaleItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = model.FlashSaleItem{
			ID:            uuid.NewV4().String(),
			CampaignID:    campaign.ID,
			SKUID:         item.SKUID,
			SalePrice:     item.SalePrice,
			QuotaTotal:    item.QuotaTotal,
			QuotaReserved: 0,
			QuotaSold:     0,
			PurchaseLimit: item.PurchaseLimit,
			CreatedAt:     s.now(),
			UpdatedAt:     s.now(),
		}
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&campaign).Error; err != nil {
			return err
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return FlashSaleResponse{}, fmt.Errorf("create draft campaign: %w", err)
	}

	return newFlashSaleResponse(campaign, items, s.now()), nil
}

func (s *Service) UpdateDraft(ctx context.Context, campaignID string, req UpdateFlashSaleRequest) (FlashSaleResponse, error) {
	db := s.database.WithContext(ctx)

	var campaign model.FlashSaleCampaign
	if err := db.Where("id = ?", campaignID).First(&campaign).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return FlashSaleResponse{}, ErrNotFound
		}
		return FlashSaleResponse{}, err
	}

	if campaign.Status != "draft" {
		return FlashSaleResponse{}, ErrNotDraft
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" || len(name) > 200 {
			return FlashSaleResponse{}, errors.New("name must be between 1 and 200 characters")
		}
		campaign.Name = name
	}

	startsAt := campaign.StartsAt
	endsAt := campaign.EndsAt
	if req.StartsAt != nil {
		t, err := time.Parse(time.RFC3339, *req.StartsAt)
		if err != nil {
			return FlashSaleResponse{}, fmt.Errorf("parse starts_at: %w", err)
		}
		startsAt = t
	}
	if req.EndsAt != nil {
		t, err := time.Parse(time.RFC3339, *req.EndsAt)
		if err != nil {
			return FlashSaleResponse{}, fmt.Errorf("parse ends_at: %w", err)
		}
		endsAt = t
	}
	if !endsAt.After(startsAt) {
		return FlashSaleResponse{}, ErrInvalidTimeRange
	}
	campaign.StartsAt = startsAt
	campaign.EndsAt = endsAt
	campaign.UpdatedAt = s.now()

	var newItems []model.FlashSaleItem
	if req.Items != nil {
		if len(*req.Items) == 0 {
			return FlashSaleResponse{}, ErrEmptyItems
		}
		skuIDs := make([]string, len(*req.Items))
		skuSeen := make(map[string]bool)
		for i, item := range *req.Items {
			if skuSeen[item.SKUID] {
				return FlashSaleResponse{}, ErrDuplicateSKU
			}
			skuSeen[item.SKUID] = true
			skuIDs[i] = item.SKUID

			if item.SalePrice <= 0 || item.QuotaTotal <= 0 || item.PurchaseLimit <= 0 {
				return FlashSaleResponse{}, errors.New("item sale_price, quota_total, and purchase_limit must be positive")
			}
		}

		var skus []model.SKU
		if err := db.Where("id IN ?", skuIDs).Find(&skus).Error; err != nil {
			return FlashSaleResponse{}, fmt.Errorf("query skus: %w", err)
		}
		if len(skus) != len(skuIDs) {
			return FlashSaleResponse{}, ErrSKUNotActive
		}
		skuMap := make(map[string]model.SKU)
		for _, sku := range skus {
			if sku.Status != model.SKUStatusActive {
				return FlashSaleResponse{}, ErrSKUNotActive
			}
			skuMap[sku.ID] = sku
		}
		for _, item := range *req.Items {
			sku := skuMap[item.SKUID]
			if item.SalePrice >= sku.RegularPrice {
				return FlashSaleResponse{}, ErrPriceNotDiscounted
			}
		}

		newItems = make([]model.FlashSaleItem, len(*req.Items))
		for i, item := range *req.Items {
			newItems[i] = model.FlashSaleItem{
				ID:            uuid.NewV4().String(),
				CampaignID:    campaign.ID,
				SKUID:         item.SKUID,
				SalePrice:     item.SalePrice,
				QuotaTotal:    item.QuotaTotal,
				QuotaReserved: 0,
				QuotaSold:     0,
				PurchaseLimit: item.PurchaseLimit,
				CreatedAt:     s.now(),
				UpdatedAt:     s.now(),
			}
		}
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&campaign).Error; err != nil {
			return err
		}
		if req.Items != nil {
			if err := tx.Where("campaign_id = ?", campaign.ID).Delete(&model.FlashSaleItem{}).Error; err != nil {
				return err
			}
			if err := tx.Create(&newItems).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return FlashSaleResponse{}, fmt.Errorf("update draft campaign: %w", err)
	}

	if req.Items == nil {
		if err := db.Where("campaign_id = ?", campaign.ID).Find(&newItems).Error; err != nil {
			return FlashSaleResponse{}, err
		}
	}

	return newFlashSaleResponse(campaign, newItems, s.now()), nil
}

func (s *Service) Schedule(ctx context.Context, campaignID string) (FlashSaleResponse, error) {
	db := s.database.WithContext(ctx)

	var campaign model.FlashSaleCampaign
	if err := db.Where("id = ?", campaignID).First(&campaign).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return FlashSaleResponse{}, ErrNotFound
		}
		return FlashSaleResponse{}, err
	}

	if campaign.Status != "draft" {
		return FlashSaleResponse{}, ErrNotDraft
	}

	var items []model.FlashSaleItem
	if err := db.Where("campaign_id = ?", campaign.ID).Find(&items).Error; err != nil {
		return FlashSaleResponse{}, err
	}
	if len(items) == 0 {
		return FlashSaleResponse{}, ErrEmptyItems
	}

	// Anti-overlapping SKU check (FR-015):
	// Check if any other scheduled campaign overlaps in time and shares any SKU
	skuIDs := make([]string, len(items))
	for i, item := range items {
		skuIDs[i] = item.SKUID
	}

	var count int64
	err := db.Model(&model.FlashSaleItem{}).
		Joins("JOIN flash_sale_campaigns ON flash_sale_campaigns.id = flash_sale_items.campaign_id").
		Where("flash_sale_items.campaign_id != ?", campaign.ID).
		Where("flash_sale_campaigns.status = 'scheduled'").
		Where("flash_sale_campaigns.starts_at < ? AND flash_sale_campaigns.ends_at > ?", campaign.EndsAt, campaign.StartsAt).
		Where("flash_sale_items.sku_id IN ?", skuIDs).
		Count(&count).Error
	if err != nil {
		return FlashSaleResponse{}, fmt.Errorf("check overlapping skus: %w", err)
	}
	if count > 0 {
		return FlashSaleResponse{}, ErrOverlappingSKU
	}

	campaign.Status = "scheduled"
	campaign.UpdatedAt = s.now()
	if err := db.Save(&campaign).Error; err != nil {
		return FlashSaleResponse{}, fmt.Errorf("schedule campaign: %w", err)
	}

	return newFlashSaleResponse(campaign, items, s.now()), nil
}

func (s *Service) CancelSchedule(ctx context.Context, campaignID string) (FlashSaleResponse, error) {
	db := s.database.WithContext(ctx)

	var campaign model.FlashSaleCampaign
	if err := db.Where("id = ?", campaignID).First(&campaign).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return FlashSaleResponse{}, ErrNotFound
		}
		return FlashSaleResponse{}, err
	}

	if campaign.Status != "scheduled" {
		return FlashSaleResponse{}, ErrNotScheduled
	}

	now := s.now()
	if !now.Before(campaign.StartsAt) {
		return FlashSaleResponse{}, ErrAlreadyStarted
	}

	campaign.Status = "draft"
	campaign.UpdatedAt = now
	if err := db.Save(&campaign).Error; err != nil {
		return FlashSaleResponse{}, fmt.Errorf("cancel schedule: %w", err)
	}

	var items []model.FlashSaleItem
	if err := db.Where("campaign_id = ?", campaign.ID).Find(&items).Error; err != nil {
		return FlashSaleResponse{}, err
	}

	return newFlashSaleResponse(campaign, items, now), nil
}

func (s *Service) Stop(ctx context.Context, campaignID string) (FlashSaleResponse, error) {
	db := s.database.WithContext(ctx)

	var campaign model.FlashSaleCampaign
	if err := db.Where("id = ?", campaignID).First(&campaign).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return FlashSaleResponse{}, ErrNotFound
		}
		return FlashSaleResponse{}, err
	}

	if campaign.Status != "scheduled" {
		return FlashSaleResponse{}, ErrNotScheduled
	}

	now := s.now()
	campaign.Status = "stopped"
	campaign.UpdatedAt = now
	if err := db.Save(&campaign).Error; err != nil {
		return FlashSaleResponse{}, fmt.Errorf("stop campaign: %w", err)
	}

	var items []model.FlashSaleItem
	if err := db.Where("campaign_id = ?", campaign.ID).Find(&items).Error; err != nil {
		return FlashSaleResponse{}, err
	}

	return newFlashSaleResponse(campaign, items, now), nil
}

func (s *Service) GetPublicFlashSale(ctx context.Context, campaignID string) (FlashSaleResponse, error) {
	db := s.database.WithContext(ctx)

	var campaign model.FlashSaleCampaign
	if err := db.Where("id = ?", campaignID).First(&campaign).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return FlashSaleResponse{}, ErrNotFound
		}
		return FlashSaleResponse{}, err
	}

	if campaign.Status == "draft" || campaign.Status == "stopped" {
		return FlashSaleResponse{}, ErrNotFound
	}

	var items []model.FlashSaleItem
	if err := db.Where("campaign_id = ?", campaign.ID).Find(&items).Error; err != nil {
		return FlashSaleResponse{}, err
	}

	return newFlashSaleResponse(campaign, items, s.now()), nil
}

func (s *Service) GetAdminFlashSale(ctx context.Context, campaignID string) (FlashSaleResponse, error) {
	db := s.database.WithContext(ctx)

	var campaign model.FlashSaleCampaign
	if err := db.Where("id = ?", campaignID).First(&campaign).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return FlashSaleResponse{}, ErrNotFound
		}
		return FlashSaleResponse{}, err
	}

	var items []model.FlashSaleItem
	if err := db.Where("campaign_id = ?", campaign.ID).Find(&items).Error; err != nil {
		return FlashSaleResponse{}, err
	}

	return newFlashSaleResponse(campaign, items, s.now()), nil
}

func (s *Service) ListPublicFlashSales(ctx context.Context, query ListFlashSalesQuery) (FlashSalePageResponse, error) {
	fingerprint := filterFingerprint("public", query.Status)
	var cursor *flashSaleCursor
	if query.Cursor != "" {
		c, err := decodeCursor(query.Cursor, fingerprint)
		if err != nil {
			return FlashSalePageResponse{}, ErrInvalidCursor
		}
		cursor = &c
	}

	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}

	now := s.now()
	db := s.database.WithContext(ctx).Model(&model.FlashSaleCampaign{}).
		Where("status = 'scheduled'")

	if query.Status != nil {
		switch *query.Status {
		case "scheduled":
			db = db.Where("starts_at > ?", now)
		case "active":
			db = db.Where("starts_at <= ? AND ends_at > ?", now, now)
		case "ended":
			db = db.Where("ends_at <= ?", now)
		default:
			return FlashSalePageResponse{}, errors.New("invalid status filter for public flash sales")
		}
	}

	if cursor != nil {
		db = db.Where("(created_at < ? OR (created_at = ? AND id < ?))", cursor.CreatedAt, cursor.CreatedAt, cursor.ID)
	}

	db = db.Order("created_at DESC, id DESC").Limit(limit + 1)

	var campaigns []model.FlashSaleCampaign
	if err := db.Find(&campaigns).Error; err != nil {
		return FlashSalePageResponse{}, fmt.Errorf("list public campaigns: %w", err)
	}

	var nextCursor *string
	if len(campaigns) > limit {
		last := campaigns[limit-1]
		nc, err := encodeCursor(last.CreatedAt, last.ID, fingerprint)
		if err != nil {
			return FlashSalePageResponse{}, fmt.Errorf("encode cursor: %w", err)
		}
		nextCursor = &nc
		campaigns = campaigns[:limit]
	}

	if len(campaigns) == 0 {
		return FlashSalePageResponse{Items: []FlashSaleResponse{}, NextCursor: nil}, nil
	}

	campaignIDs := make([]string, len(campaigns))
	for i, c := range campaigns {
		campaignIDs[i] = c.ID
	}

	var allItems []model.FlashSaleItem
	if err := s.database.WithContext(ctx).Where("campaign_id IN ?", campaignIDs).Find(&allItems).Error; err != nil {
		return FlashSalePageResponse{}, fmt.Errorf("query campaign items: %w", err)
	}

	itemsByCampaign := make(map[string][]model.FlashSaleItem)
	for _, it := range allItems {
		itemsByCampaign[it.CampaignID] = append(itemsByCampaign[it.CampaignID], it)
	}

	items := make([]FlashSaleResponse, len(campaigns))
	for i, c := range campaigns {
		items[i] = newFlashSaleResponse(c, itemsByCampaign[c.ID], now)
	}

	return FlashSalePageResponse{Items: items, NextCursor: nextCursor}, nil
}

func (s *Service) ListAdminFlashSales(ctx context.Context, query ListFlashSalesQuery) (FlashSalePageResponse, error) {
	fingerprint := filterFingerprint("admin", query.Status)
	var cursor *flashSaleCursor
	if query.Cursor != "" {
		c, err := decodeCursor(query.Cursor, fingerprint)
		if err != nil {
			return FlashSalePageResponse{}, ErrInvalidCursor
		}
		cursor = &c
	}

	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}

	now := s.now()
	db := s.database.WithContext(ctx).Model(&model.FlashSaleCampaign{})

	if query.Status != nil {
		switch *query.Status {
		case "draft":
			db = db.Where("status = 'draft'")
		case "stopped":
			db = db.Where("status = 'stopped'")
		case "scheduled":
			db = db.Where("status = 'scheduled' AND starts_at > ?", now)
		case "active":
			db = db.Where("status = 'scheduled' AND starts_at <= ? AND ends_at > ?", now, now)
		case "ended":
			db = db.Where("status = 'scheduled' AND ends_at <= ?", now)
		default:
			return FlashSalePageResponse{}, errors.New("invalid status filter for admin flash sales")
		}
	}

	if cursor != nil {
		db = db.Where("(created_at < ? OR (created_at = ? AND id < ?))", cursor.CreatedAt, cursor.CreatedAt, cursor.ID)
	}

	db = db.Order("created_at DESC, id DESC").Limit(limit + 1)

	var campaigns []model.FlashSaleCampaign
	if err := db.Find(&campaigns).Error; err != nil {
		return FlashSalePageResponse{}, fmt.Errorf("list admin campaigns: %w", err)
	}

	var nextCursor *string
	if len(campaigns) > limit {
		last := campaigns[limit-1]
		nc, err := encodeCursor(last.CreatedAt, last.ID, fingerprint)
		if err != nil {
			return FlashSalePageResponse{}, fmt.Errorf("encode cursor: %w", err)
		}
		nextCursor = &nc
		campaigns = campaigns[:limit]
	}

	if len(campaigns) == 0 {
		return FlashSalePageResponse{Items: []FlashSaleResponse{}, NextCursor: nil}, nil
	}

	campaignIDs := make([]string, len(campaigns))
	for i, c := range campaigns {
		campaignIDs[i] = c.ID
	}

	var allItems []model.FlashSaleItem
	if err := s.database.WithContext(ctx).Where("campaign_id IN ?", campaignIDs).Find(&allItems).Error; err != nil {
		return FlashSalePageResponse{}, fmt.Errorf("query campaign items: %w", err)
	}

	itemsByCampaign := make(map[string][]model.FlashSaleItem)
	for _, it := range allItems {
		itemsByCampaign[it.CampaignID] = append(itemsByCampaign[it.CampaignID], it)
	}

	items := make([]FlashSaleResponse, len(campaigns))
	for i, c := range campaigns {
		items[i] = newFlashSaleResponse(c, itemsByCampaign[c.ID], now)
	}

	return FlashSalePageResponse{Items: items, NextCursor: nextCursor}, nil
}
