package backend_test

import (
	"net/http"
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/module/flashsale"
)

func TestFlashSaleStopOnlyStopsRunningCampaign(t *testing.T) {
	env := newFlashSaleEnv(t)
	sku := env.sku(100_000)

	running := env.draft(time.Hour, 2*time.Hour, sku)
	env.mustSchedule(running.ID)
	env.setWindow(running.ID, "now() - interval '1 minute'", "now() + interval '1 hour'")
	if got := decodeJSON[flashsale.FlashSaleResponse](t, env.call(env.admin, http.MethodGet, flashSaleBasePath+"/"+running.ID, nil)); got.Status != "active" {
		t.Fatalf("running campaign status = %q, want active", got.Status)
	}
	stopRes := env.action(running.ID, "stop")
	if stopRes.Code != http.StatusOK {
		t.Fatalf("stop running status = %d; body = %s", stopRes.Code, stopRes.Body.String())
	}
	if got := decodeJSON[flashsale.FlashSaleResponse](t, stopRes); got.Status != "stopped" {
		t.Errorf("stopped response status = %q, want stopped", got.Status)
	}
	requireProblem(t, env.action(running.ID, "stop"), http.StatusConflict, problem.CodeStateTransitionConflict)

	upcomingSKU := env.sku(100_000)
	upcoming := env.draft(time.Hour, 2*time.Hour, upcomingSKU)
	env.mustSchedule(upcoming.ID)
	requireProblem(t, env.action(upcoming.ID, "stop"), http.StatusConflict, problem.CodeStateTransitionConflict)
	if got := env.storedStatus(upcoming.ID); got != "scheduled" {
		t.Errorf("upcoming campaign status after rejected stop = %q, want scheduled", got)
	}

	endedSKU := env.sku(100_000)
	ended := env.draft(time.Hour, 2*time.Hour, endedSKU)
	env.mustSchedule(ended.ID)
	env.setWindow(ended.ID, "now() - interval '2 hours'", "now() - interval '1 hour'")
	requireProblem(t, env.action(ended.ID, "stop"), http.StatusConflict, problem.CodeStateTransitionConflict)
	if got := env.storedStatus(ended.ID); got != "scheduled" {
		t.Errorf("ended campaign stored status after rejected stop = %q, want scheduled", got)
	}
	endedView := decodeJSON[flashsale.FlashSaleResponse](t, env.call(env.admin, http.MethodGet, flashSaleBasePath+"/"+ended.ID, nil))
	if endedView.Status != "ended" {
		t.Errorf("ended campaign status = %q, want ended", endedView.Status)
	}
}

func TestFlashSaleCancelScheduleUsesDatabaseTime(t *testing.T) {
	env := newFlashSaleEnv(t)
	sku := env.sku(100_000)

	upcoming := env.draft(time.Hour, 2*time.Hour, sku)
	env.mustSchedule(upcoming.ID)
	cancelRes := env.action(upcoming.ID, "cancel-schedule")
	if cancelRes.Code != http.StatusOK {
		t.Fatalf("cancel upcoming status = %d; body = %s", cancelRes.Code, cancelRes.Body.String())
	}
	if got := decodeJSON[flashsale.FlashSaleResponse](t, cancelRes); got.Status != "draft" {
		t.Errorf("cancelled status = %q, want draft", got.Status)
	}
	requireProblem(t, env.action(upcoming.ID, "cancel-schedule"), http.StatusConflict, problem.CodeStateTransitionConflict)

	env.mustSchedule(upcoming.ID)
	env.setWindow(upcoming.ID, "now() - interval '1 minute'", "now() + interval '1 hour'")
	requireProblem(t, env.action(upcoming.ID, "cancel-schedule"), http.StatusConflict, problem.CodeStateTransitionConflict)
	if got := env.storedStatus(upcoming.ID); got != "scheduled" {
		t.Errorf("started campaign status after rejected cancel = %q, want scheduled", got)
	}
}

func TestFlashSaleScheduleRejectsEndedWindowAndUnknownCampaign(t *testing.T) {
	env := newFlashSaleEnv(t)
	sku := env.sku(100_000)

	draft := env.draft(time.Hour, 2*time.Hour, sku)
	env.setWindow(draft.ID, "now() - interval '2 hours'", "now() - interval '1 hour'")
	requireProblem(t, env.action(draft.ID, "schedule"), http.StatusConflict, problem.CodeStateTransitionConflict)
	if got := env.storedStatus(draft.ID); got != "draft" {
		t.Errorf("draft status after rejected schedule = %q, want draft", got)
	}

	missing := "00000000-0000-4000-8000-000000000999"
	for _, action := range []string{"schedule", "cancel-schedule", "stop"} {
		requireProblem(t, env.action(missing, action), http.StatusNotFound, problem.CodeResourceNotFound)
	}
	requireProblem(t, env.call(env.admin, http.MethodPatch, flashSaleBasePath+"/"+missing, map[string]any{"name": "x"}),
		http.StatusNotFound, problem.CodeResourceNotFound)

	scheduled := env.draft(time.Hour, 2*time.Hour, env.sku(100_000))
	env.mustSchedule(scheduled.ID)
	requireProblem(t, env.action(scheduled.ID, "schedule"), http.StatusConflict, problem.CodeStateTransitionConflict)
}

func TestFlashSaleStoppedCampaignStillBlocksOverlappingSchedule(t *testing.T) {
	env := newFlashSaleEnv(t)
	sku := env.sku(100_000)

	first := env.draft(time.Hour, 3*time.Hour, sku)
	env.mustSchedule(first.ID)
	env.setWindow(first.ID, "now() - interval '1 minute'", "now() + interval '3 hours'")
	if res := env.action(first.ID, "stop"); res.Code != http.StatusOK {
		t.Fatalf("stop status = %d; body = %s", res.Code, res.Body.String())
	}

	second := env.draft(time.Hour, 2*time.Hour, sku)
	requireProblem(t, env.action(second.ID, "schedule"), http.StatusConflict, problem.CodeSaleScheduleOverlap)
	if got := env.storedStatus(second.ID); got != "draft" {
		t.Errorf("overlapping draft status = %q, want draft", got)
	}

	later := env.draft(4*time.Hour, 5*time.Hour, sku)
	env.mustSchedule(later.ID)
}

func TestFlashSaleScheduleRechecksSKUAndPrice(t *testing.T) {
	env := newFlashSaleEnv(t)

	cheaper := env.sku(100_000)
	draft := env.draft(time.Hour, 2*time.Hour, cheaper)
	if err := env.db.Exec("UPDATE skus SET regular_price = 500 WHERE id = ?", cheaper).Error; err != nil {
		t.Fatalf("lower regular price: %v", err)
	}
	body := requireProblem(t, env.action(draft.ID, "schedule"), http.StatusUnprocessableEntity, problem.CodeValidation)
	if !flashSaleHasViolation(body, "items[0].sale_price") {
		t.Errorf("violations = %v, want items[0].sale_price", flashSaleViolationFields(body))
	}

	mutations := map[string]string{
		"inactive sku":     "UPDATE skus SET status = 'inactive' WHERE id = ?",
		"inactive product": "UPDATE products SET status = 'inactive' WHERE id = (SELECT product_id FROM skus WHERE id = ?)",
		"hidden category":  "UPDATE categories SET is_visible = false WHERE id = (SELECT category_id FROM products WHERE id = (SELECT product_id FROM skus WHERE id = ?))",
	}
	for name, mutation := range mutations {
		t.Run(name, func(t *testing.T) {
			sku := env.sku(100_000)
			draft := env.draft(time.Hour, 2*time.Hour, sku)
			if err := env.db.Exec(mutation, sku).Error; err != nil {
				t.Fatalf("mutate: %v", err)
			}
			defer env.db.Exec("UPDATE skus SET status = 'active' WHERE id = ?", sku)
			defer env.db.Exec("UPDATE products SET status = 'active'")
			defer env.db.Exec("UPDATE categories SET is_visible = true")
			body := requireProblem(t, env.action(draft.ID, "schedule"), http.StatusUnprocessableEntity, problem.CodeValidation)
			if !flashSaleHasViolation(body, "items[0].sku_id") {
				t.Errorf("violations = %v, want items[0].sku_id", flashSaleViolationFields(body))
			}
			if got := env.storedStatus(draft.ID); got != "draft" {
				t.Errorf("status after rejected schedule = %q, want draft", got)
			}
		})
	}
}

func TestFlashSaleResponsesIncludeSKUAndProductDetails(t *testing.T) {
	env := newFlashSaleEnv(t)
	sku := env.sku(100_000)
	draft := env.draft(time.Hour, 2*time.Hour, sku)
	env.mustSchedule(draft.ID)

	check := func(label string, item flashsale.FlashSaleItemResponse) {
		t.Helper()
		if item.ID == "" || item.SKUID != sku || item.SKUCode != "FS-001" || item.ProductName != "Product" ||
			item.RegularPrice != 100_000 || item.SalePrice != 1000 || item.QuotaTotal != 10 || item.PurchaseLimit != 2 {
			t.Errorf("%s item = %+v", label, item)
		}
	}
	if len(draft.Items) != 1 {
		t.Fatalf("created items = %d, want 1", len(draft.Items))
	}
	check("create", draft.Items[0])

	admin := decodeJSON[flashsale.FlashSaleResponse](t, env.call(env.admin, http.MethodGet, flashSaleBasePath+"/"+draft.ID, nil))
	check("admin get", admin.Items[0])
	adminList := decodeJSON[flashsale.FlashSalePageResponse](t, env.call(env.admin, http.MethodGet, flashSaleBasePath, nil))
	if len(adminList.Items) != 1 || len(adminList.Items[0].Items) != 1 {
		t.Fatalf("admin list = %+v", adminList)
	}
	check("admin list", adminList.Items[0].Items[0])

	public := newTestClient(env.router)
	publicGet := decodeJSON[flashsale.FlashSaleResponse](t, public.send(t, http.MethodGet, "/api/v1/flash-sales/"+draft.ID, nil))
	check("public get", publicGet.Items[0])
	publicList := decodeJSON[flashsale.FlashSalePageResponse](t, public.send(t, http.MethodGet, "/api/v1/flash-sales", nil))
	if len(publicList.Items) != 1 {
		t.Fatalf("public list = %+v", publicList)
	}
	check("public list", publicList.Items[0].Items[0])
}

func TestFlashSaleListStatusFiltersUseDatabaseTime(t *testing.T) {
	env := newFlashSaleEnv(t)
	upcoming := env.draft(time.Hour, 2*time.Hour, env.sku(100_000))
	running := env.draft(time.Hour, 2*time.Hour, env.sku(100_000))
	ended := env.draft(time.Hour, 2*time.Hour, env.sku(100_000))
	for _, id := range []string{upcoming.ID, running.ID, ended.ID} {
		env.mustSchedule(id)
	}
	env.setWindow(running.ID, "now() - interval '1 minute'", "now() + interval '1 hour'")
	env.setWindow(ended.ID, "now() - interval '2 hours'", "now() - interval '1 hour'")

	public := newTestClient(env.router)
	for status, want := range map[string]string{"scheduled": upcoming.ID, "active": running.ID, "ended": ended.ID} {
		page := decodeJSON[flashsale.FlashSalePageResponse](t, public.send(t, http.MethodGet, "/api/v1/flash-sales?status="+status, nil))
		if len(page.Items) != 1 || page.Items[0].ID != want || page.Items[0].Status != status {
			t.Errorf("public status=%s page = %+v, want only %s", status, page.Items, want)
		}
	}
	requireProblem(t, public.send(t, http.MethodGet, "/api/v1/flash-sales?status=draft", nil), http.StatusUnprocessableEntity, problem.CodeValidation)
}
