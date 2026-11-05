package backend_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/module/flashsale"
)

func TestFlashSaleCreateAndPatchRejectBadInputWith422(t *testing.T) {
	env := newFlashSaleEnv(t)
	sku := env.sku(100_000)
	item := flashSaleItemPayload(sku, 1000)
	valid := func() map[string]any { return flashSalePayload("Sale", time.Hour, 2*time.Hour, item) }
	with := func(key string, value any) map[string]any {
		body := valid()
		body[key] = value
		return body
	}
	later := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)

	cases := []struct {
		name  string
		body  map[string]any
		field string
	}{
		{"malformed starts_at", with("starts_at", "tomorrow"), "starts_at"},
		{"malformed ends_at", with("ends_at", "2026-13-45T00:00:00Z"), "ends_at"},
		{"date without zone", with("starts_at", "2026-11-03T10:00:00"), "starts_at"},
		{"name too long", with("name", strings.Repeat("a", 201)), "name"},
		{"name of 201 multi-byte runes", with("name", strings.Repeat("Đ", 201)), "name"},
		{"blank name", with("name", "   "), "name"},
		{"ends before starts", with("starts_at", later), "ends_at"},
		{"no items", with("items", []map[string]any{}), "items"},
		{"duplicate sku", with("items", []map[string]any{item, item}), "items[1].sku_id"},
		{"zero sale price", with("items", []map[string]any{flashSaleItemPayload(sku, 0)}), "items[0].sale_price"},
		{"sale price not below regular", with("items", []map[string]any{flashSaleItemPayload(sku, 100_000)}), "items[0].sale_price"},
		{"unknown sku", with("items", []map[string]any{flashSaleItemPayload("00000000-0000-4000-8000-000000000999", 10)}), "items[0].sku_id"},
		{"zero quota", with("items", []map[string]any{{"sku_id": sku, "sale_price": 10, "quota_total": 0, "purchase_limit": 1}}), "items[0].quota_total"},
		{"zero purchase limit", with("items", []map[string]any{{"sku_id": sku, "sale_price": 10, "quota_total": 1, "purchase_limit": 0}}), "items[0].purchase_limit"},
		{"purchase limit overflows integer", with("items", []map[string]any{{"sku_id": sku, "sale_price": 10, "quota_total": 1, "purchase_limit": 4_000_000_000}}), "items[0].purchase_limit"},
	}
	existing := env.draft(time.Hour, 2*time.Hour, env.sku(100_000))
	for _, tc := range cases {
		t.Run("create "+tc.name, func(t *testing.T) {
			body := requireProblem(t, env.call(env.admin, http.MethodPost, flashSaleBasePath, tc.body), http.StatusUnprocessableEntity, problem.CodeValidation)
			if !flashSaleHasViolation(body, tc.field) {
				t.Errorf("violations = %v, want %s", flashSaleViolationFields(body), tc.field)
			}
		})
		t.Run("patch "+tc.name, func(t *testing.T) {
			body := requireProblem(t, env.call(env.admin, http.MethodPatch, flashSaleBasePath+"/"+existing.ID, tc.body), http.StatusUnprocessableEntity, problem.CodeValidation)
			if !flashSaleHasViolation(body, tc.field) {
				t.Errorf("violations = %v, want %s", flashSaleViolationFields(body), tc.field)
			}
		})
	}
	if got := env.count("SELECT count(*) FROM flash_sale_campaigns"); got != 1 {
		t.Errorf("campaigns after rejected requests = %d, want only the pre-existing draft", got)
	}

	requireProblem(t, env.call(env.admin, http.MethodPatch, flashSaleBasePath+"/"+existing.ID, map[string]any{}),
		http.StatusUnprocessableEntity, problem.CodeValidation)
	requireProblem(t, env.call(env.admin, http.MethodPatch, flashSaleBasePath+"/"+existing.ID, map[string]any{"ends_at": time.Now().UTC().Format(time.RFC3339)}),
		http.StatusUnprocessableEntity, problem.CodeValidation)
}

func TestFlashSaleNameLengthCountsRunesNotBytes(t *testing.T) {
	env := newFlashSaleEnv(t)
	sku := env.sku(100_000)

	name := strings.Repeat("Đ", 200)
	payload := flashSalePayload(name, time.Hour, 2*time.Hour, flashSaleItemPayload(sku, 1000))
	res := env.call(env.admin, http.MethodPost, flashSaleBasePath, payload)
	if res.Code != http.StatusCreated {
		t.Fatalf("200-rune Vietnamese name status = %d; body = %s", res.Code, res.Body.String())
	}
	created := decodeJSON[flashsale.FlashSaleResponse](t, res)
	if created.Name != name {
		t.Errorf("stored name differs from the request (len %d vs %d)", len(created.Name), len(name))
	}

	renamed := strings.Repeat("ế", 200)
	patchRes := env.call(env.admin, http.MethodPatch, flashSaleBasePath+"/"+created.ID, map[string]any{"name": renamed})
	if patchRes.Code != http.StatusOK {
		t.Fatalf("patch 200-rune name status = %d; body = %s", patchRes.Code, patchRes.Body.String())
	}
	if got := decodeJSON[flashsale.FlashSaleResponse](t, patchRes).Name; got != renamed {
		t.Errorf("patched name was altered")
	}
	requireProblem(t, env.call(env.admin, http.MethodPatch, flashSaleBasePath+"/"+created.ID, map[string]any{"name": renamed + "ế"}),
		http.StatusUnprocessableEntity, problem.CodeValidation)
}

func TestFlashSalePatchKeepsItemIDsOfRetainedSKUs(t *testing.T) {
	env := newFlashSaleEnv(t)
	skuA, skuB, skuC := env.sku(100_000), env.sku(100_000), env.sku(100_000)
	draft := env.draft(time.Hour, 2*time.Hour, skuA, skuB)
	itemIDs := map[string]string{}
	for _, item := range draft.Items {
		itemIDs[item.SKUID] = item.ID
	}

	res := env.call(env.admin, http.MethodPatch, flashSaleBasePath+"/"+draft.ID, map[string]any{
		"items": []map[string]any{
			{"sku_id": skuA, "sale_price": 2000, "quota_total": 7, "purchase_limit": 3},
			flashSaleItemPayload(skuC, 3000),
		},
	})
	if res.Code != http.StatusOK {
		t.Fatalf("patch status = %d; body = %s", res.Code, res.Body.String())
	}
	patched := decodeJSON[flashsale.FlashSaleResponse](t, res)
	if len(patched.Items) != 2 {
		t.Fatalf("patched items = %+v, want 2", patched.Items)
	}
	for _, item := range patched.Items {
		switch item.SKUID {
		case skuA:
			if item.ID != itemIDs[skuA] || item.SalePrice != 2000 || item.QuotaTotal != 7 || item.PurchaseLimit != 3 {
				t.Errorf("retained item = %+v, want id %s with updated values", item, itemIDs[skuA])
			}
		case skuC:
			if item.SalePrice != 3000 {
				t.Errorf("new item = %+v", item)
			}
		default:
			t.Errorf("unexpected item %+v", item)
		}
	}
	if got := env.count("SELECT count(*) FROM flash_sale_items WHERE id = ?", itemIDs[skuB]); got != 0 {
		t.Errorf("removed item still stored")
	}
}

func TestFlashSalePatchRemovingCartReferencedItemIsConflictNotServerError(t *testing.T) {
	env := newFlashSaleEnv(t)
	skuA, skuB := env.sku(100_000), env.sku(100_000)
	draft := env.draft(time.Hour, 2*time.Hour, skuA, skuB)
	var itemB string
	for _, item := range draft.Items {
		if item.SKUID == skuB {
			itemB = item.ID
		}
	}
	const adminID = "51000000-0000-4000-8000-000000000030"
	if err := env.db.Exec("INSERT INTO carts (id, user_id) VALUES ('71000000-0000-4000-8000-000000000001', ?)", adminID).Error; err != nil {
		t.Fatalf("insert cart: %v", err)
	}
	err := env.db.Exec(`INSERT INTO cart_items (cart_id, sku_id, quantity, intended_price_source, flash_sale_item_id)
		VALUES ('71000000-0000-4000-8000-000000000001', ?, 1, 'flash_sale', ?)`, skuB, itemB).Error
	if err != nil {
		t.Fatalf("insert cart item: %v", err)
	}

	res := env.call(env.admin, http.MethodPatch, flashSaleBasePath+"/"+draft.ID, map[string]any{
		"name":  "Renamed",
		"items": []map[string]any{flashSaleItemPayload(skuA, 1500)},
	})
	body := requireProblem(t, res, http.StatusConflict, problem.CodeStateTransitionConflict)
	if body.Detail == "" {
		t.Error("conflict has no detail")
	}
	after := decodeJSON[flashsale.FlashSaleResponse](t, env.call(env.admin, http.MethodGet, flashSaleBasePath+"/"+draft.ID, nil))
	if after.Name != "Campaign" || len(after.Items) != 2 {
		t.Errorf("rejected patch was partially applied: %+v", after)
	}
}
