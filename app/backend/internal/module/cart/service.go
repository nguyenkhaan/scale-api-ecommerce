package cart

import (
	"context"
	"errors"
	"fmt"

	"cloudian/cloudian-scale/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const MaxCartItems = 20

type Service struct {
	database *gorm.DB
}

func NewService(database *gorm.DB) (*Service, error) {
	if database == nil {
		return nil, errors.New("cart database is required")
	}
	return &Service{database: database}, nil
}

func (s *Service) GetCart(ctx context.Context, userID string) (CartResponse, error) {
	db := s.database.WithContext(ctx)
	cart, err := activeCart(db, userID, false)
	if err != nil {
		return CartResponse{}, fmt.Errorf("get or create cart: %w", err)
	}
	return loadCartView(db, cart.ID)
}

// RemoveCartItem deletes the user's line for a SKU; a missing line or cart is not an error.
func (s *Service) RemoveCartItem(ctx context.Context, userID, skuID string) error {
	return s.database.WithContext(ctx).Exec(`
		DELETE FROM cart_items
		WHERE sku_id = ?
		  AND cart_id IN (SELECT id FROM carts WHERE user_id = ? AND status = 'active')`,
		skuID, userID).Error
}

// activeCart returns the user's active cart, creating it on first use. With lock set the row is
// locked FOR UPDATE until the surrounding transaction ends, which serializes writers of that cart.
func activeCart(db *gorm.DB, userID string, lock bool) (model.Cart, error) {
	cart, found, err := findActiveCart(db, userID, lock)
	if err != nil || found {
		return cart, err
	}
	// Racing creators hit the partial unique index; the loser's insert is a no-op and it re-reads.
	if err := db.Exec(`
		INSERT INTO carts (user_id, status) VALUES (?, 'active')
		ON CONFLICT (user_id) WHERE status = 'active' DO NOTHING`, userID).Error; err != nil {
		return model.Cart{}, err
	}
	cart, found, err = findActiveCart(db, userID, lock)
	if err != nil {
		return model.Cart{}, err
	}
	if !found {
		return model.Cart{}, errors.New("active cart missing after creation")
	}
	return cart, nil
}

func findActiveCart(db *gorm.DB, userID string, lock bool) (model.Cart, bool, error) {
	var cart model.Cart
	query := db
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	result := query.Where("user_id = ? AND status = ?", userID, model.CartStatusActive).Limit(1).Find(&cart)
	return cart, result.RowsAffected > 0, result.Error
}

// loadCartView reads a cart's lines with product and price context in one join query.
func loadCartView(db *gorm.DB, cartID string) (CartResponse, error) {
	items := []CartItemResponse{}
	err := db.Raw(`
		SELECT ci.id, ci.sku_id, ci.quantity, ci.version,
		       ci.intended_price_source AS price_source, ci.flash_sale_item_id,
		       p.name AS product_name, s.code AS sku_code,
		       CASE WHEN ci.intended_price_source = 'flash_sale' AND fi.id IS NOT NULL
		            THEN fi.sale_price ELSE s.regular_price END AS reference_unit_price
		FROM cart_items ci
		JOIN skus s ON s.id = ci.sku_id
		JOIN products p ON p.id = s.product_id
		LEFT JOIN flash_sale_items fi ON fi.id = ci.flash_sale_item_id
		WHERE ci.cart_id = ?
		ORDER BY ci.created_at, ci.id`, cartID).Scan(&items).Error
	if err != nil {
		return CartResponse{}, fmt.Errorf("load cart items: %w", err)
	}
	return CartResponse{ID: cartID, Items: items}, nil
}
