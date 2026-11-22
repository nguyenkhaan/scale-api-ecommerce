package order

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// ExpiryStats summarizes one ExpireDue pass.
type ExpiryStats struct {
	Scanned int
	Expired int
	// LostRace counts due orders another worker held, or another path (payment, cancel) already ended.
	LostRace int
	Errors   int
	// Remaining is the number of still-due pending orders after the pass.
	Remaining int64
	// MaxLatenessSeconds is the largest now() - expires_at, by database time, among orders this pass expired.
	MaxLatenessSeconds float64
}

// dueCandidate is a due pending order and its keyset position in (expires_at, id) order.
type dueCandidate struct {
	ID        string    `gorm:"column:id"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
}

// ExpireDue handles up to batchSize pending orders whose deadline has passed by database time.
// Every order runs in its own short transaction and locks its row with SKIP LOCKED, so concurrent
// workers never wait on each other and a failing order cannot block or roll back the rest of the batch.
// Candidates are read page by page with a keyset cursor and failures do not count toward the batch, so
// permanently failing orders at the head of the queue cannot starve the healthy ones behind them.
func ExpireDue(ctx context.Context, db *gorm.DB, batchSize int) (ExpiryStats, error) {
	var stats ExpiryStats
	if batchSize < 1 {
		return stats, errors.New("expiry batch size must be positive")
	}
	handled := 0
	var last *dueCandidate
	for handled < batchSize {
		want := batchSize - handled
		query := `SELECT id, expires_at FROM orders WHERE status = 'pending_payment' AND expires_at <= now()`
		args := []any{}
		if last != nil {
			query += ` AND (expires_at, id) > (?, ?)`
			args = append(args, last.ExpiresAt, last.ID)
		}
		query += ` ORDER BY expires_at, id LIMIT ?`
		args = append(args, want)
		var candidates []dueCandidate
		if err := db.WithContext(ctx).Raw(query, args...).Scan(&candidates).Error; err != nil {
			return stats, fmt.Errorf("select due orders: %w", err)
		}
		stats.Scanned += len(candidates)
		for _, candidate := range candidates {
			if err := ctx.Err(); err != nil {
				return stats, err
			}
			expired, lateness, err := expireOne(ctx, db, candidate.ID)
			switch {
			case err != nil:
				stats.Errors++
				log.Printf("expiry worker: order %s skipped: %v", candidate.ID, err)
			case !expired:
				stats.LostRace++
				handled++
			default:
				stats.Expired++
				handled++
				stats.MaxLatenessSeconds = max(stats.MaxLatenessSeconds, lateness)
			}
		}
		if len(candidates) < want {
			break
		}
		last = &candidates[len(candidates)-1]
	}
	if err := db.WithContext(ctx).Raw(
		`SELECT count(*) FROM orders WHERE status = 'pending_payment' AND expires_at <= now()`,
	).Scan(&stats.Remaining).Error; err != nil {
		return stats, fmt.Errorf("count expiry backlog: %w", err)
	}
	return stats, nil
}

// expireOne expires a single order; expired is false when another worker or path got there first.
func expireOne(ctx context.Context, db *gorm.DB, orderID string) (expired bool, lateness float64, err error) {
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked []struct {
			Lateness float64 `gorm:"column:lateness"`
		}
		if err := tx.Raw(
			`SELECT EXTRACT(EPOCH FROM now() - expires_at)::float8 AS lateness FROM orders
			 WHERE id = ? AND status = 'pending_payment' FOR UPDATE SKIP LOCKED`, orderID,
		).Scan(&locked).Error; err != nil {
			return fmt.Errorf("lock order: %w", err)
		}
		if len(locked) == 0 {
			return nil
		}
		transition, err := Finish(tx, orderID, OutcomeExpired)
		if err != nil {
			return err
		}
		expired = transition.Applied && transition.Status == string(OutcomeExpired)
		lateness = locked[0].Lateness
		return nil
	})
	if err != nil {
		return false, 0, err
	}
	return expired, lateness, nil
}
