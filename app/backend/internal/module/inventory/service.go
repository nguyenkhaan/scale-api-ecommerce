package inventory

import (
	"cloudian/cloudian-scale/internal/model"
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)
var (
	ErrInventoryNotFound = errors.New("inventory not found")
	ErrBelowReserved     = errors.New("inventory would fall below reserved")
	ErrQuantityOverflow  = errors.New("inventory quantity would overflow")
)
type Service struct {
	database *gorm.DB
}
func NewService(database *gorm.DB) (*Service , error) {
	if database == nil {
		return nil, errors.New("database is required")
	}
	return &Service{database: database}, nil
}

func (service *Service) Adjust(context context.Context , skuID string, actorUserID string, request AdjustmentRequest) (InventoryResponse , error) {
		var updated model.Inventory

	err := service.database.WithContext(context).Transaction(
		func(transaction *gorm.DB) error {
			var current model.Inventory

			err := transaction.
				Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("sku_id = ?", skuID).
				Take(&current).Error

			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInventoryNotFound
			}
			if err != nil {
				return fmt.Errorf("lock inventory: %w", err)
			}

			if request.Delta > 0 &&
				current.OnHand > math.MaxInt64-request.Delta {
				return ErrQuantityOverflow
			}

			nextOnHand := current.OnHand + request.Delta
			if nextOnHand < current.Reserved {
				return ErrBelowReserved
			}
			if current.Version == math.MaxInt64 {
				return fmt.Errorf("inventory version overflow")
			}

			result := transaction.Model(&model.Inventory{}).
				Where("sku_id = ?", skuID).
				Updates(map[string]any{
					"on_hand":   nextOnHand,
					"version":   current.Version + 1,
					"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
				})
			if result.Error != nil {
				return fmt.Errorf("update inventory: %w", result.Error)
			}
			if result.RowsAffected != 1 {
				return ErrInventoryNotFound
			}

			adjustment := model.InventoryAdjustment{
				SKUID:        skuID,
				ActorUserID:  actorUserID,
				Delta:        request.Delta,
				OnHandBefore: current.OnHand,
				OnHandAfter:  nextOnHand,
				Reason:       strings.TrimSpace(request.Reason),
			}
			if err := transaction.Create(&adjustment).Error; err != nil {
				return fmt.Errorf("insert inventory adjustment: %w", err)
			}

			if err := transaction.
				Where("sku_id = ?", skuID).
				Take(&updated).Error; err != nil {
				return fmt.Errorf("reload inventory: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return InventoryResponse{}, err
	}
	return newInventoryResponse(updated), nil
}