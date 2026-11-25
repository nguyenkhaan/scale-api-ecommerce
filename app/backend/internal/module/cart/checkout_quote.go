package cart

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"gorm.io/gorm"
)

type QuotedLine struct {
	CheckoutLineResponse
	SKUCode       string
	ProductName   string
	SKUAttributes json.RawMessage
}

type LockedQuote struct {
	Now         time.Time
	Lines       []QuotedLine
	TotalAmount int64
}

// LockAndQuote recalculates the selected cart lines while holding the shared rows checkout must reserve.
// The caller owns the transaction and decides whether to commit or roll back the locks and usage rows.
func LockAndQuote(tx *gorm.DB, userID string, selection CheckoutSelection) (LockedQuote, error) {
	now, err := dbNow(tx)
	if err != nil {
		return LockedQuote{}, fmt.Errorf("read checkout time: %w", err)
	}
	lines, err := loadSelectedLinesForUpdate(tx, userID, selection)
	if err != nil {
		return LockedQuote{}, err
	}
	if err := lockQuoteResources(tx, userID, lines); err != nil {
		return LockedQuote{}, err
	}
	contexts, err := loadLineContexts(tx, userID, lines)
	if err != nil {
		return LockedQuote{}, fmt.Errorf("load locked checkout quote: %w", err)
	}
	quoted, total, lineErrors := quoteLines(lines, contexts, now)
	if len(lineErrors) > 0 {
		return LockedQuote{}, &LineErrors{Lines: lineErrors}
	}
	return LockedQuote{Now: now, Lines: quoted, TotalAmount: total}, nil
}

func lockQuoteResources(tx *gorm.DB, userID string, lines []selectedLine) error {
	skuIDs := make([]string, 0, len(lines))
	flashItemSet := make(map[string]struct{}, len(lines))
	for _, line := range lines {
		skuIDs = append(skuIDs, line.SKUID)
		if line.isFlashSale() && line.FlashSaleItemID != nil {
			flashItemSet[*line.FlashSaleItemID] = struct{}{}
		}
	}
	sort.Strings(skuIDs)
	if len(skuIDs) > 0 {
		var locked []string
		if err := tx.Raw(`SELECT sku_id FROM inventories WHERE sku_id IN ? ORDER BY sku_id FOR UPDATE`, skuIDs).
			Scan(&locked).Error; err != nil {
			return fmt.Errorf("lock checkout inventory: %w", err)
		}
	}
	flashItemIDs := make([]string, 0, len(flashItemSet))
	for id := range flashItemSet {
		flashItemIDs = append(flashItemIDs, id)
	}
	sort.Strings(flashItemIDs)
	if len(flashItemIDs) == 0 {
		return nil
	}
	var lockedFlashItems []string
	if err := tx.Raw(`SELECT id FROM flash_sale_items WHERE id IN ? ORDER BY id FOR UPDATE`, flashItemIDs).
		Scan(&lockedFlashItems).Error; err != nil {
		return fmt.Errorf("lock checkout flash-sale items: %w", err)
	}
	placeholders := make([]string, len(flashItemIDs))
	args := make([]any, 0, len(flashItemIDs)+1)
	for i, id := range flashItemIDs {
		placeholders[i] = "?::uuid"
		args = append(args, id)
	}
	args = append(args, userID)
	query := `INSERT INTO flash_sale_customer_usage (flash_sale_item_id, user_id)
		SELECT unnest(ARRAY[` + strings.Join(placeholders, ",") + `]), ? ORDER BY 1 ON CONFLICT DO NOTHING`
	if err := tx.Exec(query, args...).Error; err != nil {
		return fmt.Errorf("ensure checkout customer usage: %w", err)
	}
	var lockedUsage []string
	if err := tx.Raw(`SELECT flash_sale_item_id FROM flash_sale_customer_usage
		WHERE user_id = ? AND flash_sale_item_id IN ? ORDER BY flash_sale_item_id FOR UPDATE`, userID, flashItemIDs).
		Scan(&lockedUsage).Error; err != nil {
		return fmt.Errorf("lock checkout customer usage: %w", err)
	}
	return nil
}

func quoteLines(lines []selectedLine, contexts map[string]lineContext, now time.Time) ([]QuotedLine, int64, []problem.LineProblem) {
	quoted := make([]QuotedLine, 0, len(lines))
	var total int64
	var lineErrors []problem.LineProblem
	for _, line := range lines {
		context := contexts[line.ID]
		unitPrice, lineProblem := evaluateLine(line, context, now)
		if lineProblem != nil {
			lineErrors = append(lineErrors, *lineProblem)
			continue
		}
		subtotal := unitPrice * int64(line.Quantity)
		total += subtotal
		quoted = append(quoted, QuotedLine{
			CheckoutLineResponse: CheckoutLineResponse{
				CartItemID: line.ID, CartItemVersion: line.Version, SKUID: line.SKUID,
				Quantity: line.Quantity, UnitPrice: unitPrice, Subtotal: subtotal,
				PriceSource: line.PriceSource, FlashSaleItemID: line.FlashSaleItemID,
			},
			SKUCode: context.sku.Code, ProductName: context.sku.ProductName, SKUAttributes: context.sku.Attributes,
		})
	}
	return quoted, total, lineErrors
}
