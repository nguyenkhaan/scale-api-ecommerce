package order

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
	"uuid"
)

// ErrInvalidCursor means the pagination cursor is malformed or was issued for another status filter.
var ErrInvalidCursor = errors.New("invalid order cursor")

// orderCursor is the keyset position (created_at DESC, id DESC) plus the filter it belongs to.
type orderCursor struct {
	CreatedAt time.Time `json:"c"`
	ID        string    `json:"i"`
	Status    string    `json:"s"`
}

func encodeOrderCursor(cursor orderCursor) (string, error) {
	cursor.CreatedAt = cursor.CreatedAt.UTC()
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeOrderCursor(raw, expectedStatus string) (orderCursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return orderCursor{}, ErrInvalidCursor
	}
	var cursor orderCursor
	if err := json.Unmarshal(payload, &cursor); err != nil {
		return orderCursor{}, ErrInvalidCursor
	}
	if cursor.CreatedAt.IsZero() || cursor.Status != expectedStatus {
		return orderCursor{}, ErrInvalidCursor
	}
	parsed, err := uuid.Parse(cursor.ID)
	if err != nil {
		return orderCursor{}, ErrInvalidCursor
	}
	cursor.ID = parsed.String()
	return cursor, nil
}
