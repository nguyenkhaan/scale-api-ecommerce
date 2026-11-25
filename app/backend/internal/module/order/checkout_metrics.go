package order

import (
	"log"
	"sync/atomic"
)

type CheckoutOutcome string

const (
	OutcomeCreated    CheckoutOutcome = "created"
	OutcomeReplayed   CheckoutOutcome = "replayed"
	OutcomeRejected   CheckoutOutcome = "rejected"
	OutcomeConflict   CheckoutOutcome = "conflict"
	OutcomeInProgress CheckoutOutcome = "in_progress"
	OutcomeError      CheckoutOutcome = "error"
)

type CheckoutRecorder interface {
	RecordCheckout(outcome CheckoutOutcome, code string)
}

type CheckoutCounters struct {
	created    atomic.Int64
	replayed   atomic.Int64
	rejected   atomic.Int64
	conflict   atomic.Int64
	inProgress atomic.Int64
	error      atomic.Int64
}

func NewCheckoutCounters() *CheckoutCounters {
	return &CheckoutCounters{}
}

func (c *CheckoutCounters) RecordCheckout(outcome CheckoutOutcome, code string) {
	switch outcome {
	case OutcomeCreated:
		c.created.Add(1)
	case OutcomeReplayed:
		c.replayed.Add(1)
	case OutcomeRejected:
		c.rejected.Add(1)
	case OutcomeConflict:
		c.conflict.Add(1)
	case OutcomeInProgress:
		c.inProgress.Add(1)
	case OutcomeError:
		c.error.Add(1)
	}
	log.Printf("checkout outcome=%s code=%s", outcome, code)
}

func (c *CheckoutCounters) Snapshot() map[CheckoutOutcome]int64 {
	return map[CheckoutOutcome]int64{
		OutcomeCreated: c.created.Load(), OutcomeReplayed: c.replayed.Load(),
		OutcomeRejected: c.rejected.Load(), OutcomeConflict: c.conflict.Load(),
		OutcomeInProgress: c.inProgress.Load(), OutcomeError: c.error.Load(),
	}
}

type noOpCheckoutRecorder struct{}

func (noOpCheckoutRecorder) RecordCheckout(CheckoutOutcome, string) {}
