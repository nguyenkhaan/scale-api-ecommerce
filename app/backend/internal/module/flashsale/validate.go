package flashsale

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"
	"uuid"

	"cloudian/cloudian-scale/internal/httpserver/problem"
)

const maxNameRunes = 200

// draftInput is a request whose shape has been validated; it is not yet checked against the database.
type draftInput struct {
	Name     string
	StartsAt time.Time
	EndsAt   time.Time
	Items    []CreateFlashSaleItem
}

// draftPatch is a validated partial update; nil fields are unchanged.
type draftPatch struct {
	Name     *string
	StartsAt *time.Time
	EndsAt   *time.Time
	Items    *[]CreateFlashSaleItem
}

func violation(field, reason string) problem.Violation {
	return problem.Violation{Field: field, Reason: reason}
}

func validateCreate(req CreateFlashSaleRequest) (draftInput, []problem.Violation) {
	var violations []problem.Violation
	input := draftInput{}

	name, v := validateName(req.Name)
	input.Name = name
	violations = append(violations, v...)

	startsAt, v := parseTimestamp("starts_at", req.StartsAt)
	violations = append(violations, v...)
	endsAt, v := parseTimestamp("ends_at", req.EndsAt)
	violations = append(violations, v...)
	input.StartsAt, input.EndsAt = startsAt, endsAt
	if !startsAt.IsZero() && !endsAt.IsZero() && !endsAt.After(startsAt) {
		violations = append(violations, violation("ends_at", "must be strictly after starts_at"))
	}

	items, v := validateItems(req.Items)
	input.Items = items
	violations = append(violations, v...)
	return input, violations
}

func validateUpdate(req UpdateFlashSaleRequest) (draftPatch, []problem.Violation) {
	var violations []problem.Violation
	patch := draftPatch{}
	if req.Name == nil && req.StartsAt == nil && req.EndsAt == nil && req.Items == nil {
		return patch, []problem.Violation{violation("body", "must contain at least one field to update")}
	}

	if req.Name != nil {
		name, v := validateName(*req.Name)
		patch.Name = &name
		violations = append(violations, v...)
	}
	if req.StartsAt != nil {
		t, v := parseTimestamp("starts_at", *req.StartsAt)
		patch.StartsAt = &t
		violations = append(violations, v...)
	}
	if req.EndsAt != nil {
		t, v := parseTimestamp("ends_at", *req.EndsAt)
		patch.EndsAt = &t
		violations = append(violations, v...)
	}
	if patch.StartsAt != nil && patch.EndsAt != nil && !patch.StartsAt.IsZero() && !patch.EndsAt.IsZero() &&
		!patch.EndsAt.After(*patch.StartsAt) {
		violations = append(violations, violation("ends_at", "must be strictly after starts_at"))
	}
	if req.Items != nil {
		items, v := validateItems(*req.Items)
		patch.Items = &items
		violations = append(violations, v...)
	}
	return patch, violations
}

// validateName trims the name and counts characters (not bytes) so Vietnamese names get the full limit.
func validateName(raw string) (string, []problem.Violation) {
	name := strings.TrimSpace(raw)
	switch count := utf8.RuneCountInString(name); {
	case count == 0:
		return name, []problem.Violation{violation("name", "is required")}
	case count > maxNameRunes:
		return name, []problem.Violation{violation("name", fmt.Sprintf("must be at most %d characters", maxNameRunes))}
	}
	return name, nil
}

func parseTimestamp(field, raw string) (time.Time, []problem.Violation) {
	if strings.TrimSpace(raw) == "" {
		return time.Time{}, []problem.Violation{violation(field, "is required")}
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, []problem.Violation{violation(field, "must be an RFC 3339 date-time")}
	}
	return parsed.UTC(), nil
}

// validateItems checks item shape and returns items with canonical lower-case SKU IDs.
func validateItems(items []CreateFlashSaleItem) ([]CreateFlashSaleItem, []problem.Violation) {
	if len(items) == 0 {
		return nil, []problem.Violation{violation("items", "must contain at least one item")}
	}
	var violations []problem.Violation
	normalized := make([]CreateFlashSaleItem, len(items))
	firstIndex := make(map[string]int, len(items))
	for i, item := range items {
		prefix := fmt.Sprintf("items[%d].", i)
		normalized[i] = item
		parsed, err := uuid.Parse(item.SKUID)
		if err != nil {
			violations = append(violations, violation(prefix+"sku_id", "must be a valid UUID"))
		} else {
			canonical := parsed.String()
			normalized[i].SKUID = canonical
			if first, seen := firstIndex[canonical]; seen {
				violations = append(violations, violation(prefix+"sku_id", fmt.Sprintf("duplicates items[%d].sku_id", first)))
			} else {
				firstIndex[canonical] = i
			}
		}
		if item.SalePrice <= 0 {
			violations = append(violations, violation(prefix+"sale_price", "must be greater than 0"))
		}
		if item.QuotaTotal <= 0 {
			violations = append(violations, violation(prefix+"quota_total", "must be greater than 0"))
		}
		if item.PurchaseLimit <= 0 || item.PurchaseLimit > math.MaxInt32 {
			violations = append(violations, violation(prefix+"purchase_limit", "must be between 1 and 2147483647"))
		}
	}
	return normalized, violations
}
