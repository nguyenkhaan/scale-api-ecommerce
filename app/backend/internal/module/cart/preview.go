package cart

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"gorm.io/gorm"
)

// PreviewCheckout quotes the selected cart lines from current data. It never writes: no inventory,
// quota or usage row is touched and the read-only snapshot keeps all lines consistent with one clock.
func (s *Service) PreviewCheckout(ctx context.Context, userID string, selection CheckoutSelection) (CheckoutPreviewResponse, error) {
	var preview CheckoutPreviewResponse
	err := s.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		preview, err = previewInSnapshot(tx, userID, selection)
		return err
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return CheckoutPreviewResponse{}, err
	}
	return preview, nil
}

func previewInSnapshot(tx *gorm.DB, userID string, selection CheckoutSelection) (CheckoutPreviewResponse, error) {
	now, err := dbNow(tx)
	if err != nil {
		return CheckoutPreviewResponse{}, err
	}
	lines, err := loadSelectedLines(tx, userID, selection)
	if err != nil {
		return CheckoutPreviewResponse{}, err
	}
	contexts, err := loadLineContexts(tx, userID, lines)
	if err != nil {
		return CheckoutPreviewResponse{}, err
	}

	var (
		responses  = make([]CheckoutLineResponse, 0, len(lines))
		lineErrors []problem.LineProblem
		total      int64
	)
	for _, line := range lines {
		unitPrice, lineProblem := evaluateLine(line, contexts[line.ID], now)
		if lineProblem != nil {
			lineErrors = append(lineErrors, *lineProblem)
			continue
		}
		subtotal := unitPrice * int64(line.Quantity)
		total += subtotal
		responses = append(responses, CheckoutLineResponse{
			CartItemID:      line.ID,
			CartItemVersion: line.Version,
			SKUID:           line.SKUID,
			Quantity:        line.Quantity,
			UnitPrice:       unitPrice,
			Subtotal:        subtotal,
			PriceSource:     line.PriceSource,
			FlashSaleItemID: line.FlashSaleItemID,
		})
	}
	if len(lineErrors) > 0 {
		return CheckoutPreviewResponse{}, &LineErrors{Lines: lineErrors}
	}
	return CheckoutPreviewResponse{
		Currency:    "VND",
		TotalAmount: total,
		Lines:       responses,
		PreviewedAt: now.UTC().Format(time.RFC3339),
	}, nil
}

// loadSelectedLines resolves the selection against the user's active cart in one query and returns
// the lines in selection order. Unknown ids yield a 422 that cannot reveal other customers' carts;
// a stale version yields ErrCartItemChanged before any line is judged.
func loadSelectedLines(tx *gorm.DB, userID string, selection CheckoutSelection) ([]selectedLine, error) {
	ids := make([]string, len(selection.Items))
	for i, item := range selection.Items {
		ids[i] = item.CartItemID
	}
	var rows []selectedLine
	err := tx.Raw(`
		SELECT ci.id, ci.sku_id, ci.quantity, ci.version, ci.intended_price_source, ci.flash_sale_item_id
		FROM cart_items ci
		JOIN carts c ON c.id = ci.cart_id AND c.user_id = ? AND c.status = 'active'
		WHERE ci.id IN ?`, userID, ids).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	byID := make(map[string]selectedLine, len(rows))
	for _, row := range rows {
		byID[row.ID] = row
	}

	var violations []problem.Violation
	lines := make([]selectedLine, 0, len(ids))
	stale := false
	for i, item := range selection.Items {
		line, ok := byID[item.CartItemID]
		if !ok {
			violations = append(violations, problem.Violation{
				Field:  "items[" + strconv.Itoa(i) + "].cart_item_id",
				Reason: "not found in your cart",
			})
			continue
		}
		stale = stale || line.Version != item.Version
		lines = append(lines, line)
	}
	if len(violations) > 0 {
		return nil, &ValidationError{Violations: violations}
	}
	if stale {
		return nil, ErrCartItemChanged
	}
	return lines, nil
}

// loadLineContexts batch-loads SKU, stock, flash sale and customer usage data for all lines.
func loadLineContexts(tx *gorm.DB, userID string, lines []selectedLine) (map[string]lineContext, error) {
	skuIDs := make([]string, 0, len(lines))
	var flashIDs []string
	for _, line := range lines {
		skuIDs = append(skuIDs, line.SKUID)
		if line.isFlashSale() && line.FlashSaleItemID != nil {
			flashIDs = append(flashIDs, *line.FlashSaleItemID)
		}
	}
	skus, err := loadSKUInfos(tx, skuIDs)
	if err != nil {
		return nil, err
	}
	flashItems, err := loadFlashInfos(tx, flashIDs)
	if err != nil {
		return nil, err
	}
	used, err := loadUsedQuantities(tx, userID, flashIDs)
	if err != nil {
		return nil, err
	}

	contexts := make(map[string]lineContext, len(lines))
	for _, line := range lines {
		lineCtx := lineContext{sku: skus[line.SKUID]}
		if line.isFlashSale() && line.FlashSaleItemID != nil {
			if item, ok := flashItems[*line.FlashSaleItemID]; ok {
				lineCtx.flash = &item
				lineCtx.used = used[item.ID]
			}
		}
		contexts[line.ID] = lineCtx
	}
	return contexts, nil
}
