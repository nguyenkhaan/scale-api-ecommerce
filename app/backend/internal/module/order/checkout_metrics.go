package order

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

type noOpCheckoutRecorder struct{}

func (noOpCheckoutRecorder) RecordCheckout(CheckoutOutcome, string) {}
