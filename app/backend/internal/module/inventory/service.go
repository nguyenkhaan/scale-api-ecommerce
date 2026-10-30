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

type inventoryRow struct {
	SKUID    string `gorm:"column:sku_id"`
	SKUCode  string `gorm:"column:sku_code"`
	OnHand   int64  `gorm:"column:on_hand"`
	Reserved int64  `gorm:"column:reserved"`
	Version  int64  `gorm:"column:version"`
}

func NewService(database *gorm.DB) (*Service, error) {
	if database == nil {
		return nil, errors.New("database is required")
	}
	return &Service{database: database}, nil
}

func (service *Service) Adjust(context context.Context, skuID string, actorUserID string, request AdjustmentRequest) (InventoryResponse, error) {
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
					"on_hand":    nextOnHand,
					"version":    current.Version + 1,
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

func (service *Service) ListInventories(
	ctx context.Context,
	query ListInventoriesQuery,
) (InventoryPageResponse, error) {
	if query.Limit == 0 {
		query.Limit = 20
	}
	var cursor inventoryCursor
	if query.Cursor != "" {
		decoded, err := decodeInventoryCursor(query.Cursor)
		if err != nil {
			return InventoryPageResponse{}, err
		}
		cursor = decoded
	}

	database := service.database.WithContext(ctx).
		Table("inventories").
		Select("inventories.sku_id, skus.code AS sku_code, inventories.on_hand, inventories.reserved, inventories.version").
		Joins("JOIN skus ON skus.id = inventories.sku_id")
	if query.Cursor != "" {
		database = database.Where("(skus.code, inventories.sku_id) > (?, ?)", cursor.SKUCode, cursor.SKUID)
	}
	var rows []inventoryRow
	if err := database.
		Order("skus.code ASC, inventories.sku_id ASC").
		Limit(query.Limit + 1).
		Scan(&rows).Error; err != nil {
		return InventoryPageResponse{}, fmt.Errorf("list inventories: %w", err)
	}

	page := InventoryPageResponse{Items: make([]InventoryResponse, 0, min(len(rows), query.Limit))}
	if len(rows) > query.Limit {
		rows = rows[:query.Limit]
		last := rows[len(rows)-1]
		nextCursor, err := encodeInventoryCursor(last.SKUCode, last.SKUID)
		if err != nil {
			return InventoryPageResponse{}, fmt.Errorf("encode inventory cursor: %w", err)
		}
		page.NextCursor = &nextCursor
	}
	for _, row := range rows {
		page.Items = append(page.Items, InventoryResponse{
			SKUID:     row.SKUID,
			SKUCode:   row.SKUCode,
			OnHand:    row.OnHand,
			Reserved:  row.Reserved,
			Available: row.OnHand - row.Reserved,
			Version:   row.Version,
		})
	}
	return page, nil
}

func (service *Service) ListAdjustments(
	ctx context.Context,
	skuID string,
	query ListAdjustmentHistoryQuery,
) (InventoryAdjustmentPageResponse, error) {
	if query.Limit == 0 {
		query.Limit = 20
	}
	var cursor adjustmentCursor
	if query.Cursor != "" {
		decoded, err := decodeAdjustmentCursor(query.Cursor, skuID)
		if err != nil {
			return InventoryAdjustmentPageResponse{}, err
		}
		cursor = decoded
	}

	var sku model.SKU
	if err := service.database.WithContext(ctx).Select("id").Where("id = ?", skuID).Take(&sku).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return InventoryAdjustmentPageResponse{}, ErrInventoryNotFound
		}
		return InventoryAdjustmentPageResponse{}, fmt.Errorf("find adjustment SKU: %w", err)
	}

	database := service.database.WithContext(ctx).Where("sku_id = ?", skuID)
	if query.Cursor != "" {
		database = database.Where("(created_at, id) < (?, ?)", cursor.CreatedAt, cursor.ID)
	}
	var adjustments []model.InventoryAdjustment
	if err := database.
		Order("created_at DESC, id DESC").
		Limit(query.Limit + 1).
		Find(&adjustments).Error; err != nil {
		return InventoryAdjustmentPageResponse{}, fmt.Errorf("list inventory adjustments: %w", err)
	}

	page := InventoryAdjustmentPageResponse{Items: make([]InventoryAdjustmentResponse, 0, min(len(adjustments), query.Limit))}
	if len(adjustments) > query.Limit {
		adjustments = adjustments[:query.Limit]
		last := adjustments[len(adjustments)-1]
		nextCursor, err := encodeAdjustmentCursor(last.CreatedAt, last.ID, skuID)
		if err != nil {
			return InventoryAdjustmentPageResponse{}, fmt.Errorf("encode adjustment cursor: %w", err)
		}
		page.NextCursor = &nextCursor
	}
	for _, adjustment := range adjustments {
		page.Items = append(page.Items, newInventoryAdjustmentResponse(adjustment))
	}
	return page, nil
}
