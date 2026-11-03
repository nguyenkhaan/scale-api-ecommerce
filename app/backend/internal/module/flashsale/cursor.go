package flashsale

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"
	"uuid"
)

var ErrInvalidCursor = errors.New("invalid flash sale cursor")

type flashSaleCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
	Filter    string    `json:"filter"`
}

func encodeCursor(createdAt time.Time, id, filter string) (string, error) {
	payload, err := json.Marshal(flashSaleCursor{
		CreatedAt: createdAt.UTC(),
		ID:        id,
		Filter:    filter,
	})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeCursor(raw, expectedFilter string) (flashSaleCursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return flashSaleCursor{}, ErrInvalidCursor
	}
	var cursor flashSaleCursor
	if err := json.Unmarshal(payload, &cursor); err != nil {
		return flashSaleCursor{}, ErrInvalidCursor
	}
	if cursor.CreatedAt.IsZero() || cursor.Filter != expectedFilter {
		return flashSaleCursor{}, ErrInvalidCursor
	}
	if _, err := uuid.Parse(cursor.ID); err != nil {
		return flashSaleCursor{}, ErrInvalidCursor
	}
	return cursor, nil
}

func filterFingerprint(scope string, status *string) string {
	st := ""
	if status != nil {
		st = *status
	}
	sum := sha256.Sum256([]byte(scope + "\x00" + st))
	return hex.EncodeToString(sum[:])
}
