package cart

import (
	"time"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/model"
)

// selectedLine is one selected cart row joined with the data needed to judge it.
type selectedLine struct {
	ID              string  `gorm:"column:id"`
	SKUID           string  `gorm:"column:sku_id"`
	Quantity        int     `gorm:"column:quantity"`
	Version         int64   `gorm:"column:version"`
	PriceSource     string  `gorm:"column:intended_price_source"`
	FlashSaleItemID *string `gorm:"column:flash_sale_item_id"`
}

func (l selectedLine) isFlashSale() bool {
	return l.PriceSource == model.PriceSourceFlashSale
}

// lineContext is everything evaluateLine reads; flash is nil for regular lines or a vanished item.
type lineContext struct {
	sku   skuInfo
	flash *flashInfo
	used  int64
}

// evaluateLine returns the unit price of a buyable line, or the first failing check in BR-24 order:
// ITEM_UNAVAILABLE, SALE_NOT_ACTIVE, PURCHASE_LIMIT_EXCEEDED, SALE_QUOTA_EXCEEDED, INSUFFICIENT_STOCK.
// A flash sale line never falls back to the regular price.
func evaluateLine(line selectedLine, ctx lineContext, now time.Time) (int64, *problem.LineProblem) {
	qty := int64(line.Quantity)
	fail := func(code string, available int64) (int64, *problem.LineProblem) {
		if available < 0 {
			available = 0
		}
		return 0, &problem.LineProblem{
			SKUID:             line.SKUID,
			RequestedQuantity: line.Quantity,
			Code:              code,
			AvailableQuantity: &available,
		}
	}
	failWithoutAvailability := func(code string) (int64, *problem.LineProblem) {
		return 0, &problem.LineProblem{SKUID: line.SKUID, RequestedQuantity: line.Quantity, Code: code}
	}

	if !ctx.sku.Available {
		return failWithoutAvailability(problem.CodeItemUnavailable)
	}
	flash := ctx.flash
	if line.isFlashSale() && (flash == nil || flash.SKUID != line.SKUID || !flash.acceptingOrders(now)) {
		return failWithoutAvailability(problem.CodeSaleNotActive)
	}

	allowed := int64(ctx.sku.MaxQuantityPerOrder)
	if line.isFlashSale() {
		allowed = min(allowed, flash.PurchaseLimit-ctx.used)
	}
	if qty > allowed {
		return fail(problem.CodePurchaseLimitExceeded, allowed)
	}
	if line.isFlashSale() {
		if quota := flash.QuotaTotal - flash.QuotaReserved - flash.QuotaSold; qty > quota {
			return fail(problem.CodeSaleQuotaExceeded, quota)
		}
	}
	if stock := ctx.sku.OnHand - ctx.sku.Reserved; qty > stock {
		return fail(problem.CodeInsufficientStock, stock)
	}

	if line.isFlashSale() {
		return flash.SalePrice, nil
	}
	return ctx.sku.RegularPrice, nil
}
