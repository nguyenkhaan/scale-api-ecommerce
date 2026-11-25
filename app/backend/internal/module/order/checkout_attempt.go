package order

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var ErrAttemptNotFound = errors.New("checkout attempt not found")

func (s *CheckoutService) LookupAttempt(ctx context.Context, userID, key string) (IdempotencyResult, error) {
	var row struct {
		IdempotencyKey string       `gorm:"column:idempotency_key"`
		ResponseStatus *int         `gorm:"column:response_status"`
		ResponseBody   *string      `gorm:"column:response_body"`
		CompletedAt    sql.NullTime `gorm:"column:completed_at"`
	}
	query := s.database.Session(&gorm.Session{Logger: logger.Discard}).WithContext(ctx).Raw(`SELECT idempotency_key, response_status, response_body::text AS response_body, completed_at
		FROM idempotency_records WHERE user_id = ? AND operation = 'place_order' AND idempotency_key = ?`, userID, key).
		Scan(&row)
	if query.Error != nil {
		return IdempotencyResult{}, fmt.Errorf("lookup checkout attempt: %w", query.Error)
	}
	if query.RowsAffected != 1 || row.ResponseStatus == nil || row.ResponseBody == nil || !row.CompletedAt.Valid {
		return IdempotencyResult{}, ErrAttemptNotFound
	}
	return IdempotencyResult{
		IdempotencyKey: row.IdempotencyKey,
		ResponseStatus: *row.ResponseStatus,
		ResponseBody:   []byte(*row.ResponseBody),
		CompletedAt:    row.CompletedAt.Time,
	}, nil
}
