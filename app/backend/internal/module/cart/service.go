package cart

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloudian/cloudian-scale/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"uuid"
)

const MaxCartItems = 20

var (
	ErrCartLimitExceeded       = errors.New("cart cannot contain more than 20 distinct items")
	ErrInvalidQuantity         = errors.New("quantity must be greater than zero")
	ErrInvalidPriceSource      = errors.New("price_source must be regular or flash_sale")
	ErrFlashSaleItemRequired   = errors.New("flash_sale_item_id is required for flash_sale price source")
	ErrFlashSaleItemForbidden  = errors.New("flash_sale_item_id must be null for regular price source")
	ErrSKUNotActive            = errors.New("sku does not exist or is not active")
	ErrExceedsMaxQuantity      = errors.New("quantity exceeds maximum allowed per order")
	ErrFlashSaleItemInvalid    = errors.New("flash sale item does not exist or is not applicable")
)

type Service struct {
	database *gorm.DB
	nowFunc  func() time.Time
}

func NewService(database *gorm.DB) (*Service, error) {
	if database == nil {
		return nil, errors.New("cart database is required")
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

func (s *Service) GetCart(ctx context.Context, userID string) (CartResponse, error) {
	db := s.database.WithContext(ctx)

	cart, err := s.getOrCreateActiveCart(db, userID)
	if err != nil {
		return CartResponse{}, fmt.Errorf("get or create cart: %w", err)
	}

	var items []model.CartItem
	if err := db.Where("cart_id = ?", cart.ID).Order("created_at ASC, id ASC").Find(&items).Error; err != nil {
		return CartResponse{}, fmt.Errorf("find cart items: %w", err)
	}

	return newCartResponse(cart, items), nil
}

func (s *Service) UpsertCartItem(ctx context.Context, userID, skuID string, req UpsertCartItemRequest) (CartResponse, error) {
	if req.Quantity <= 0 {
		return CartResponse{}, ErrInvalidQuantity
	}

	if req.PriceSource != model.PriceSourceRegular && req.PriceSource != model.PriceSourceFlashSale {
		return CartResponse{}, ErrInvalidPriceSource
	}

	if req.PriceSource == model.PriceSourceRegular && req.FlashSaleItemID != nil {
		return CartResponse{}, ErrFlashSaleItemForbidden
	}

	if req.PriceSource == model.PriceSourceFlashSale && (req.FlashSaleItemID == nil || *req.FlashSaleItemID == "") {
		return CartResponse{}, ErrFlashSaleItemRequired
	}

	db := s.database.WithContext(ctx)

	// Validate SKU is active
	var sku model.SKU
	if err := db.Where("id = ? AND status = ?", skuID, model.SKUStatusActive).First(&sku).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return CartResponse{}, ErrSKUNotActive
		}
		return CartResponse{}, err
	}

	if req.PriceSource == model.PriceSourceRegular {
		if req.Quantity > sku.MaxQuantityPerOrder {
			return CartResponse{}, ErrExceedsMaxQuantity
		}
	} else {
		// Validate flash sale item
		var fsItem model.FlashSaleItem
		if err := db.Where("id = ? AND sku_id = ?", *req.FlashSaleItemID, skuID).First(&fsItem).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return CartResponse{}, ErrFlashSaleItemInvalid
			}
			return CartResponse{}, err
		}

		var campaign model.FlashSaleCampaign
		if err := db.Where("id = ?", fsItem.CampaignID).First(&campaign).Error; err != nil {
			return CartResponse{}, err
		}
		if campaign.Status == "stopped" || campaign.Status == "draft" {
			return CartResponse{}, ErrFlashSaleItemInvalid
		}
		if req.Quantity > fsItem.PurchaseLimit {
			return CartResponse{}, ErrExceedsMaxQuantity
		}
	}

	var cart model.Cart
	var items []model.CartItem

	err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		cart, err = s.getOrCreateActiveCart(tx, userID)
		if err != nil {
			return err
		}

		var existing model.CartItem
		err = tx.Where("cart_id = ? AND sku_id = ?", cart.ID, skuID).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Check 20 distinct items limit
			var count int64
			if err := tx.Model(&model.CartItem{}).Where("cart_id = ?", cart.ID).Count(&count).Error; err != nil {
				return err
			}
			if count >= MaxCartItems {
				return ErrCartLimitExceeded
			}

			newItem := model.CartItem{
				ID:                  uuid.NewV4().String(),
				CartID:              cart.ID,
				SKUID:               skuID,
				Quantity:            req.Quantity,
				IntendedPriceSource: req.PriceSource,
				FlashSaleItemID:     req.FlashSaleItemID,
				Version:             1,
				CreatedAt:           s.now(),
				UpdatedAt:           s.now(),
			}
			if err := tx.Create(&newItem).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			// Update existing item with incremented version
			existing.Quantity = req.Quantity
			existing.IntendedPriceSource = req.PriceSource
			existing.FlashSaleItemID = req.FlashSaleItemID
			existing.Version = existing.Version + 1
			existing.UpdatedAt = s.now()
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
		}

		return tx.Where("cart_id = ?", cart.ID).Order("created_at ASC, id ASC").Find(&items).Error
	})
	if err != nil {
		return CartResponse{}, err
	}

	return newCartResponse(cart, items), nil
}

func (s *Service) RemoveCartItem(ctx context.Context, userID, skuID string) error {
	db := s.database.WithContext(ctx)

	var cart model.Cart
	if err := db.Where("user_id = ? AND status = 'active'", userID).First(&cart).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	return db.Where("cart_id = ? AND sku_id = ?", cart.ID, skuID).Delete(&model.CartItem{}).Error
}

func (s *Service) getOrCreateActiveCart(tx *gorm.DB, userID string) (model.Cart, error) {
	var cart model.Cart
	err := tx.Where("user_id = ? AND status = 'active'", userID).First(&cart).Error
	if err == nil {
		return cart, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Cart{}, err
	}

	cart = model.Cart{
		ID:        uuid.NewV4().String(),
		UserID:    userID,
		Status:    "active",
		CreatedAt: s.now(),
		UpdatedAt: s.now(),
	}

	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		Where:     clause.Where{Exprs: []clause.Expression{clause.Expr{SQL: "status = 'active'"}}},
		DoNothing: true,
	}).Create(&cart).Error; err != nil {
		return model.Cart{}, err
	}

	if err := tx.Where("user_id = ? AND status = 'active'", userID).First(&cart).Error; err != nil {
		return model.Cart{}, err
	}
	return cart, nil
}
