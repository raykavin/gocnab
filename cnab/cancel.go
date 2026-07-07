package cnab

import (
	"time"

	"github.com/raykavin/gocnab/cnab/layout"
)

// CancelPayment cancels a payment previously sent in an earlier
// remittance file. It re-sends the original payment's data with the
// FEBRABAN movement type set to "9" (exclusion) and the movement
// instruction code set to "99", exactly as the standard requires.
type CancelPayment struct {
	// Original is the payment being cancelled. It must be the same
	// concrete payment value (or an equivalent reconstruction of it) that
	// was originally sent; the bank matches a cancellation to the
	// original instruction by its data, not by an internal SDK id.
	Original Payment
}

func (c CancelPayment) validate(now time.Time, opts validateOptions) error {
	if c.Original == nil {
		return &ValidationError{Context: "CancelPayment", Reason: "Original is required"}
	}
	// A cancellation legitimately references a payment date that has
	// already passed, so the past-date rule is intentionally not
	// enforced here.
	return c.Original.validate(now, validateOptions{AllowPastDate: true})
}

func (c CancelPayment) toSegments(l Layout) ([]DetailSegment, error) {
	segments, err := c.Original.toSegments(l)
	if err != nil {
		return nil, err
	}
	out := make([]DetailSegment, len(segments))
	for i, seg := range segments {
		values := cloneSegmentValues(seg.Values)
		values[layout.KeyMovementType] = "9"
		values[layout.KeyInstructionCode] = "99"
		out[i] = DetailSegment{Key: seg.Key, Values: values}
	}
	return out, nil
}
