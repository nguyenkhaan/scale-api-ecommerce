package order

import (
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"uuid"
)

const maxPlaceOrderLines = 20

func ValidatePlaceOrder(req *PlaceOrderRequest) []problem.Violation {
	var violations []problem.Violation
	add := func(field, reason string) {
		violations = append(violations, problem.Violation{Field: field, Reason: reason})
	}
	if len(req.Lines) == 0 {
		add("lines", "must contain at least one item")
	}
	if len(req.Lines) > maxPlaceOrderLines {
		add("lines", "cannot exceed 20 items")
	}

	seenCartItems := make(map[string]bool, len(req.Lines))
	seenSKUs := make(map[string]bool, len(req.Lines))
	var total int64
	totalOverflow := false
	for i := range req.Lines {
		line := &req.Lines[i]
		prefix := "lines[" + strconv.Itoa(i) + "]."
		if parsed, err := uuid.Parse(line.CartItemID); err != nil {
			add(prefix+"cart_item_id", "must be a valid UUID")
		} else {
			line.CartItemID = parsed.String()
			if seenCartItems[line.CartItemID] {
				add(prefix+"cart_item_id", "duplicate cart item")
			}
			seenCartItems[line.CartItemID] = true
		}
		if line.CartItemVersion < 1 {
			add(prefix+"cart_item_version", "must be at least 1")
		}
		if parsed, err := uuid.Parse(line.SKUID); err != nil {
			add(prefix+"sku_id", "must be a valid UUID")
		} else {
			line.SKUID = parsed.String()
			if seenSKUs[line.SKUID] {
				add(prefix+"sku_id", "duplicate SKU")
			}
			seenSKUs[line.SKUID] = true
		}
		if line.Quantity < 1 {
			add(prefix+"quantity", "must be at least 1")
		}
		if line.UnitPrice < 1 {
			add(prefix+"unit_price", "must be at least 1")
		}

		validSubtotal := line.Quantity > 0 && line.UnitPrice > 0 && int64(line.Quantity) <= math.MaxInt64/line.UnitPrice
		if !validSubtotal || line.Subtotal != int64(line.Quantity)*line.UnitPrice || line.Subtotal < 1 {
			add(prefix+"subtotal", "must equal quantity multiplied by unit price")
		} else if total > math.MaxInt64-line.Subtotal {
			totalOverflow = true
		} else {
			total += line.Subtotal
		}

		switch line.PriceSource {
		case "regular":
			if line.FlashSaleItemID != nil {
				add(prefix+"flash_sale_item_id", "must be null for regular price")
			}
		case "flash_sale":
			if line.FlashSaleItemID == nil {
				add(prefix+"flash_sale_item_id", "is required for flash-sale price")
			} else if parsed, err := uuid.Parse(*line.FlashSaleItemID); err != nil {
				add(prefix+"flash_sale_item_id", "must be a valid UUID")
			} else {
				id := parsed.String()
				line.FlashSaleItemID = &id
			}
		default:
			add(prefix+"price_source", "must be regular or flash_sale")
		}
	}

	if req.Currency != "VND" {
		add("currency", "must be VND")
	}
	if req.ExpectedTotal < 1 || totalOverflow || req.ExpectedTotal != total {
		add("expected_total", "must equal the sum of line subtotals")
	}

	req.Recipient.Name = strings.TrimSpace(req.Recipient.Name)
	req.Recipient.Phone = strings.TrimSpace(req.Recipient.Phone)
	req.Recipient.Address = strings.TrimSpace(req.Recipient.Address)
	if length := utf8.RuneCountInString(req.Recipient.Name); length < 1 || length > 120 {
		add("recipient.name", "must contain 1 to 120 characters")
	}
	if length := utf8.RuneCountInString(req.Recipient.Phone); length < 8 || length > 20 {
		add("recipient.phone", "must contain 8 to 20 characters")
	}
	if length := utf8.RuneCountInString(req.Recipient.Address); length < 1 || length > 500 {
		add("recipient.address", "must contain 1 to 500 characters")
	}
	return violations
}

func ValidateIdempotencyKey(key string) []problem.Violation {
	length := utf8.RuneCountInString(key)
	if length < 8 || length > 255 {
		return []problem.Violation{{Field: "Idempotency-Key", Reason: "must contain 8 to 255 characters"}}
	}
	return nil
}
