package auth

import (
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
)

// Login attempts are counted per normalized email in PostgreSQL so the limit holds across API instances.
const (
	maxLoginAttempts   = 5
	loginAttemptWindow = 15 * time.Minute
)

type TooManyAttemptsError struct {
	RetryAfter time.Duration
}

func (err *TooManyAttemptsError) Error() string {
	return fmt.Sprintf("too many failed login attempts; retry after %s", err.RetryAfter)
}

// reserveLoginAttempt counts the attempt before the password is checked, so parallel guesses are limited
// exactly like sequential ones. The row lock of the upsert orders concurrent attempts; a successful login
// deletes the row, so failure_count only ever holds unsuccessful attempts of the current window.
func reserveLoginAttempt(database *gorm.DB, normalizedEmail string) error {
	var state struct {
		AttemptCount      int     `gorm:"column:failure_count"`
		RetryAfterSeconds float64 `gorm:"column:retry_after_seconds"`
	}
	window := loginAttemptWindow.Seconds()
	err := database.Raw(
		`INSERT INTO login_throttles AS throttle (email_normalized, failure_count, window_started_at)
		VALUES (?, 1, now())
		ON CONFLICT (email_normalized) DO UPDATE SET
			failure_count = CASE
				WHEN throttle.window_started_at > now() - make_interval(secs => ?) THEN throttle.failure_count + 1
				ELSE 1
			END,
			window_started_at = CASE
				WHEN throttle.window_started_at > now() - make_interval(secs => ?) THEN throttle.window_started_at
				ELSE now()
			END
		RETURNING failure_count,
			EXTRACT(EPOCH FROM window_started_at + make_interval(secs => ?) - now()) AS retry_after_seconds`,
		normalizedEmail, window, window, window,
	).Scan(&state).Error
	if err != nil {
		return fmt.Errorf("reserve login attempt: %w", err)
	}
	if state.AttemptCount <= maxLoginAttempts {
		return nil
	}
	retryAfter := time.Duration(math.Max(1, math.Ceil(state.RetryAfterSeconds))) * time.Second
	return &TooManyAttemptsError{RetryAfter: retryAfter}
}

func clearLoginAttempts(database *gorm.DB, normalizedEmail string) error {
	if err := database.Exec("DELETE FROM login_throttles WHERE email_normalized = ?", normalizedEmail).Error; err != nil {
		return fmt.Errorf("clear login attempts: %w", err)
	}
	return nil
}
