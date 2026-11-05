package backend_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"gorm.io/gorm"
)

// runConcurrently starts every function behind one barrier and returns when all have finished.
func runConcurrently(calls []func() *httptest.ResponseRecorder) []*httptest.ResponseRecorder {
	results := make([]*httptest.ResponseRecorder, len(calls))
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i, call := range calls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			results[i] = call()
		}()
	}
	close(start)
	wg.Wait()
	return results
}

const flashSaleOverlapPairsSQL = `
	SELECT count(*) FROM flash_sale_items a
	JOIN flash_sale_campaigns ca ON ca.id = a.campaign_id
	JOIN flash_sale_items b ON b.sku_id = a.sku_id AND b.campaign_id > a.campaign_id
	JOIN flash_sale_campaigns cb ON cb.id = b.campaign_id
	WHERE ca.status IN ('scheduled', 'stopped') AND cb.status IN ('scheduled', 'stopped')
	  AND ca.starts_at < cb.ends_at AND cb.starts_at < ca.ends_at`

func TestFlashSaleConcurrentScheduleOfOverlappingDraftsSchedulesExactlyOne(t *testing.T) {
	env := newFlashSaleEnv(t)
	const rounds, campaigns = 6, 8
	for round := 0; round < rounds; round++ {
		sku := env.sku(100_000)
		ids := make([]string, campaigns)
		calls := make([]func() *httptest.ResponseRecorder, campaigns)
		for i := range calls {
			draft := env.draft(time.Hour, 2*time.Hour, sku)
			ids[i] = draft.ID
			client := env.concurrentClient()
			calls[i] = func() *httptest.ResponseRecorder {
				return env.call(client, http.MethodPost, flashSaleBasePath+"/"+draft.ID+"/schedule", nil)
			}
		}

		ok, overlap := 0, 0
		for _, res := range runConcurrently(calls) {
			switch res.Code {
			case http.StatusOK:
				ok++
			case http.StatusConflict:
				requireProblem(t, res, http.StatusConflict, problem.CodeSaleScheduleOverlap)
				overlap++
			default:
				t.Errorf("round %d: unexpected status %d: %s", round, res.Code, res.Body.String())
			}
		}
		if ok != 1 || overlap != campaigns-1 {
			t.Errorf("round %d: %d scheduled, %d overlap conflicts, want 1 and %d", round, ok, overlap, campaigns-1)
		}
		if got := env.count("SELECT count(*) FROM flash_sale_campaigns WHERE status = 'scheduled' AND id IN ?", ids); got != 1 {
			t.Errorf("round %d: scheduled campaigns in database = %d, want 1", round, got)
		}
	}
	if got := env.count(flashSaleOverlapPairsSQL); got != 0 {
		t.Errorf("overlapping scheduled pairs = %d, want 0", got)
	}
}

func TestFlashSaleConcurrentPatchAndScheduleNeverOverlapOnSKU(t *testing.T) {
	env := newFlashSaleEnv(t)
	shared := env.sku(100_000)
	blocker := env.draft(time.Hour, 2*time.Hour, shared)
	env.mustSchedule(blocker.ID)

	const rounds = 25
	for round := 0; round < rounds; round++ {
		own := env.sku(100_000)
		draft := env.draft(time.Hour, 2*time.Hour, own)
		patchClient, scheduleClient := env.concurrentClient(), env.concurrentClient()
		results := runConcurrently([]func() *httptest.ResponseRecorder{
			func() *httptest.ResponseRecorder {
				return env.call(patchClient, http.MethodPatch, flashSaleBasePath+"/"+draft.ID, map[string]any{
					"items": []map[string]any{flashSaleItemPayload(shared, 1000)},
				})
			},
			func() *httptest.ResponseRecorder {
				return env.call(scheduleClient, http.MethodPost, flashSaleBasePath+"/"+draft.ID+"/schedule", nil)
			},
		})
		for _, res := range results {
			if res.Code >= http.StatusInternalServerError {
				t.Fatalf("round %d: server error %d: %s", round, res.Code, res.Body.String())
			}
		}
		if got := env.count(flashSaleOverlapPairsSQL); got != 0 {
			t.Fatalf("round %d: overlapping scheduled pairs = %d, want 0 (patch=%d schedule=%d)", round, got, results[0].Code, results[1].Code)
		}
	}
}

// holdLock runs hold inside an open transaction, runs call concurrently, proves call is blocked, then
// releases the lock and returns the call's response.
func flashSaleBlockedUntilRelease(t *testing.T, db *gorm.DB, hold func(tx *gorm.DB), call func() *httptest.ResponseRecorder) *httptest.ResponseRecorder {
	t.Helper()
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("begin: %v", tx.Error)
	}
	released := false
	defer func() {
		if !released {
			tx.Rollback()
		}
	}()
	hold(tx)

	done := make(chan *httptest.ResponseRecorder, 1)
	go func() { done <- call() }()
	select {
	case res := <-done:
		t.Fatalf("request finished while the lock was held (status %d): %s", res.Code, res.Body.String())
	case <-time.After(400 * time.Millisecond):
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("release lock: %v", err)
	}
	released = true
	select {
	case res := <-done:
		return res
	case <-time.After(10 * time.Second):
		t.Fatal("request did not finish after the lock was released")
		return nil
	}
}

func TestFlashSaleScheduleWaitsForSKUAdvisoryLock(t *testing.T) {
	env := newFlashSaleEnv(t)
	sku := env.sku(100_000)
	draft := env.draft(time.Hour, 2*time.Hour, sku)
	client := env.concurrentClient()

	res := flashSaleBlockedUntilRelease(t, env.db,
		func(tx *gorm.DB) {
			if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtextextended(?::text, 0))", sku).Error; err != nil {
				t.Fatalf("take advisory lock: %v", err)
			}
		},
		func() *httptest.ResponseRecorder {
			return env.call(client, http.MethodPost, flashSaleBasePath+"/"+draft.ID+"/schedule", nil)
		})
	if res.Code != http.StatusOK {
		t.Errorf("schedule after lock release status = %d; body = %s", res.Code, res.Body.String())
	}
}

func TestFlashSaleDraftEditAndScheduleWaitForCampaignRowLock(t *testing.T) {
	env := newFlashSaleEnv(t)
	sku := env.sku(100_000)
	other := env.sku(100_000)
	draft := env.draft(time.Hour, 2*time.Hour, sku)

	lockRow := func(tx *gorm.DB) {
		if err := tx.Exec("SELECT id FROM flash_sale_campaigns WHERE id = ? FOR UPDATE", draft.ID).Error; err != nil {
			t.Fatalf("lock campaign row: %v", err)
		}
	}
	patchClient := env.concurrentClient()
	res := flashSaleBlockedUntilRelease(t, env.db, lockRow, func() *httptest.ResponseRecorder {
		return env.call(patchClient, http.MethodPatch, flashSaleBasePath+"/"+draft.ID, map[string]any{
			"items": []map[string]any{flashSaleItemPayload(other, 1000)},
		})
	})
	if res.Code != http.StatusOK {
		t.Errorf("patch after lock release status = %d; body = %s", res.Code, res.Body.String())
	}

	scheduleClient := env.concurrentClient()
	res = flashSaleBlockedUntilRelease(t, env.db, lockRow, func() *httptest.ResponseRecorder {
		return env.call(scheduleClient, http.MethodPost, flashSaleBasePath+"/"+draft.ID+"/schedule", nil)
	})
	if res.Code != http.StatusOK {
		t.Errorf("schedule after lock release status = %d; body = %s", res.Code, res.Body.String())
	}
}
