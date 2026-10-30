package inventory

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"uuid"
)

var ErrInvalidCursor = errors.New("invalid inventory cursor")

type inventoryCursor struct {
	SKUCode string `json:"sku_code"`
	SKUID   string `json:"sku_id"`
}

func encodeInventoryCursor(skuCode, skuID string) (string, error) {
	payload, err := json.Marshal(inventoryCursor{SKUCode: skuCode, SKUID: skuID})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeInventoryCursor(raw string) (inventoryCursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return inventoryCursor{}, ErrInvalidCursor
	}
	var cursor inventoryCursor
	if err := json.Unmarshal(payload, &cursor); err != nil || cursor.SKUCode == "" {
		return inventoryCursor{}, ErrInvalidCursor
	}
	if _, err := uuid.Parse(cursor.SKUID); err != nil {
		return inventoryCursor{}, ErrInvalidCursor
	}
	return cursor, nil
}

type adjustmentCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
	SKUID     string    `json:"sku_id"`
}

func encodeAdjustmentCursor(createdAt time.Time, id, skuID string) (string, error) {
	payload, err := json.Marshal(adjustmentCursor{CreatedAt: createdAt.UTC(), ID: id, SKUID: skuID})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeAdjustmentCursor(raw, expectedSKUID string) (adjustmentCursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return adjustmentCursor{}, ErrInvalidCursor
	}
	var cursor adjustmentCursor
	if err := json.Unmarshal(payload, &cursor); err != nil || cursor.CreatedAt.IsZero() || cursor.SKUID != expectedSKUID {
		return adjustmentCursor{}, ErrInvalidCursor
	}
	if _, err := uuid.Parse(cursor.ID); err != nil {
		return adjustmentCursor{}, ErrInvalidCursor
	}
	return cursor, nil
}
