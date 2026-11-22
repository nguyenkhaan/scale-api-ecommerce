package payment

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"cloudian/cloudian-scale/internal/httpserver/problem"
)

const (
	EventSucceeded = "succeeded"
	EventFailed    = "failed"
)

// callback is a validated PaymentCallback payload.
type callback struct {
	EventID    string
	SessionID  string
	EventType  string
	Amount     int64
	Currency   string
	OccurredAt time.Time
}

// callbackWire is the signed JSON body exchanged between the simulator and the callback endpoint.
type callbackWire struct {
	EventID    string     `json:"event_id"`
	SessionID  string     `json:"payment_session_id"`
	EventType  string     `json:"event_type"`
	Amount     *int64     `json:"amount"`
	Currency   string     `json:"currency"`
	OccurredAt *time.Time `json:"occurred_at"`
}

// EventReceipt is the PaymentEventReceipt contract object.
type EventReceipt struct {
	EventID                string `json:"event_id"`
	Result                 string `json:"result"`
	OrderStatus            string `json:"order_status,omitempty"`
	ReconciliationRequired bool   `json:"reconciliation_required"`
}

// decodeStrict rejects unknown fields and trailing data so the body means exactly what was signed.
func decodeStrict(body []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if decoder.More() {
		return errTrailingData
	}
	return nil
}

func isCurrencyCode(value string) bool {
	if len(value) != 3 {
		return false
	}
	for _, char := range value {
		if char < 'A' || char > 'Z' {
			return false
		}
	}
	return true
}

// parseCallback validates a callback body. Currency only needs to be a 3-letter code: a wrong currency is a
// reconciliation matter, not a malformed request.
func parseCallback(body []byte) (callback, []problem.Violation) {
	var wire callbackWire
	if err := decodeStrict(body, &wire); err != nil {
		return callback{}, []problem.Violation{{Field: "body", Reason: "must be a valid PaymentCallback JSON object"}}
	}
	var violations []problem.Violation
	reject := func(field, reason string) {
		violations = append(violations, problem.Violation{Field: field, Reason: reason})
	}
	eventID, ok := canonicalUUID(wire.EventID)
	if !ok {
		reject("event_id", "must be a valid UUID")
	}
	sessionID, ok := canonicalUUID(wire.SessionID)
	if !ok {
		reject("payment_session_id", "must be a valid UUID")
	}
	if wire.EventType != EventSucceeded && wire.EventType != EventFailed {
		reject("event_type", "must be succeeded or failed")
	}
	if wire.Amount == nil || *wire.Amount < 1 {
		reject("amount", "must be a positive integer")
	}
	if !isCurrencyCode(wire.Currency) {
		reject("currency", "must be a 3-letter currency code")
	}
	if wire.OccurredAt == nil {
		reject("occurred_at", "is required")
	}
	if len(violations) > 0 {
		return callback{}, violations
	}
	return callback{
		EventID: eventID, SessionID: sessionID, EventType: wire.EventType,
		Amount: *wire.Amount, Currency: wire.Currency, OccurredAt: *wire.OccurredAt,
	}, nil
}

// contentHash identifies what an event says, independent of how it was serialized or when it was sent, so
// a simulator redelivery of the same event is a duplicate while a different claim under one id is a conflict.
func (cb callback) contentHash() []byte {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s|%d|%s", cb.EventID, cb.SessionID, cb.EventType, cb.Amount, cb.Currency)))
	return sum[:]
}
