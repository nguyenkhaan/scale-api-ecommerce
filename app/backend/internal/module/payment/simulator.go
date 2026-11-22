package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"uuid"
)

const (
	SignatureValid   = "valid"
	SignatureInvalid = "invalid"

	maxDelayMillis    = 600_000
	maxRepeatCount    = 10
	repeatStagger     = 20 * time.Millisecond
	deliveryTimeout   = 10 * time.Second
	invalidSecretSalt = "-invalid"
)

// ScheduleRequest is the SchedulePaymentEventRequest contract object. Pointers tell "absent" from zero.
type ScheduleRequest struct {
	SessionID        string  `json:"payment_session_id"`
	EventType        string  `json:"event_type"`
	DelayMillis      *int    `json:"delay_ms"`
	RepeatCount      *int    `json:"repeat_count"`
	DropDelivery     bool    `json:"drop_delivery"`
	EventID          *string `json:"event_id"`
	AmountOverride   *int64  `json:"amount_override"`
	CurrencyOverride *string `json:"currency_override"`
	SignatureMode    *string `json:"signature_mode"`
}

// ScheduledEvent is the ScheduledPaymentEvent contract object.
type ScheduledEvent struct {
	EventID     string    `json:"event_id"`
	ScheduledAt time.Time `json:"scheduled_at"`
}

type schedulePlan struct {
	request ScheduleRequest
	delay   time.Duration
	repeats int
	eventID string
}

// validateSchedule checks the request and resolves defaults.
func validateSchedule(request ScheduleRequest) (schedulePlan, []problem.Violation) {
	plan := schedulePlan{request: request, repeats: 1}
	var violations []problem.Violation
	reject := func(field, reason string) {
		violations = append(violations, problem.Violation{Field: field, Reason: reason})
	}

	if id, ok := canonicalUUID(request.SessionID); ok {
		plan.request.SessionID = id
	} else {
		reject("payment_session_id", "must be a valid UUID")
	}
	if request.EventType != EventSucceeded && request.EventType != EventFailed {
		reject("event_type", "must be succeeded or failed")
	}
	if request.DelayMillis != nil {
		if *request.DelayMillis < 0 || *request.DelayMillis > maxDelayMillis {
			reject("delay_ms", "must be between 0 and 600000")
		}
		plan.delay = time.Duration(*request.DelayMillis) * time.Millisecond
	}
	if request.RepeatCount != nil {
		if *request.RepeatCount < 1 || *request.RepeatCount > maxRepeatCount {
			reject("repeat_count", "must be between 1 and 10")
		}
		plan.repeats = *request.RepeatCount
	}
	if request.EventID != nil {
		if id, ok := canonicalUUID(*request.EventID); ok {
			plan.eventID = id
		} else {
			reject("event_id", "must be a valid UUID")
		}
	}
	if request.AmountOverride != nil && *request.AmountOverride < 1 {
		reject("amount_override", "must be a positive integer")
	}
	if request.CurrencyOverride != nil && !isCurrencyCode(*request.CurrencyOverride) {
		reject("currency_override", "must be a 3-letter currency code")
	}
	if mode := request.SignatureMode; mode != nil && *mode != SignatureValid && *mode != SignatureInvalid {
		reject("signature_mode", "must be valid or invalid")
	}
	if plan.eventID == "" {
		plan.eventID = uuid.New().String()
	}
	return plan, violations
}

// ScheduleEvent queues deliveries of one payment event. It returns ErrNotFound when the session is unknown.
// Signatures are computed when each delivery fires, so a long delay never produces a stale timestamp.
// Scheduled deliveries live in memory only; losing them on restart is the same as a lost callback.
func (service *Service) ScheduleEvent(ctx context.Context, plan schedulePlan) (ScheduledEvent, error) {
	var session Session
	read := service.db.WithContext(ctx).Raw(`SELECT `+sessionColumns+` FROM payment_sessions WHERE id = ?`,
		plan.request.SessionID).Scan(&session)
	if read.Error != nil {
		return ScheduledEvent{}, fmt.Errorf("read payment session: %w", read.Error)
	}
	if read.RowsAffected == 0 {
		return ScheduledEvent{}, ErrNotFound
	}

	now := time.Now()
	wire := callbackWire{
		EventID: plan.eventID, SessionID: session.ID, EventType: plan.request.EventType,
		Amount: &session.Amount, Currency: session.Currency, OccurredAt: &now,
	}
	if plan.request.AmountOverride != nil {
		wire.Amount = plan.request.AmountOverride
	}
	if plan.request.CurrencyOverride != nil {
		wire.Currency = *plan.request.CurrencyOverride
	}
	body, err := json.Marshal(wire)
	if err != nil {
		return ScheduledEvent{}, fmt.Errorf("encode callback: %w", err)
	}

	secret := service.cfg.SimulatorSecret
	if plan.request.SignatureMode != nil && *plan.request.SignatureMode == SignatureInvalid {
		secret += invalidSecretSalt
	}
	if !plan.request.DropDelivery {
		for index := 0; index < plan.repeats; index++ {
			time.AfterFunc(plan.delay+time.Duration(index)*repeatStagger, func() {
				service.deliver(Delivery{Body: body, Signature: Sign(secret, time.Now(), body)}, plan.eventID)
			})
		}
	}
	return ScheduledEvent{EventID: plan.eventID, ScheduledAt: now.Add(plan.delay)}, nil
}

// deliver never logs the signature or the payload; only the event id and the failure cause.
func (service *Service) deliver(delivery Delivery, eventID string) {
	ctx, cancel := context.WithTimeout(context.Background(), deliveryTimeout)
	defer cancel()
	if err := service.sender.Send(ctx, delivery); err != nil {
		log.Printf("payment simulator: delivery of event %s failed: %v", eventID, err)
	}
}
