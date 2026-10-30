package catalog

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"uuid"
)

var ErrInvalidCursor = errors.New("invalid catalog cursor")

type productCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
	Filter    string    `json:"filter"`
}

func encodeProductCursor(createdAt time.Time, id, filter string) (string, error) {
	payload, err := json.Marshal(productCursor{CreatedAt: createdAt.UTC(), ID: id, Filter: filter})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeProductCursor(raw, expectedFilter string) (productCursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return productCursor{}, ErrInvalidCursor
	}
	var cursor productCursor
	if err := json.Unmarshal(payload, &cursor); err != nil {
		return productCursor{}, ErrInvalidCursor
	}
	if cursor.CreatedAt.IsZero() || cursor.Filter != expectedFilter {
		return productCursor{}, ErrInvalidCursor
	}
	if _, err := uuid.Parse(cursor.ID); err != nil {
		return productCursor{}, ErrInvalidCursor
	}
	return cursor, nil
}

func publicFilterFingerprint(query string, categoryID *string) string {
	category := ""
	if categoryID != nil {
		category = *categoryID
	}
	return cursorFilterFingerprint("public", query, category)
}

func adminProductFilterFingerprint(status *string) string {
	value := ""
	if status != nil {
		value = *status
	}
	return cursorFilterFingerprint("admin-products", value)
}

func cursorFilterFingerprint(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:])
}
